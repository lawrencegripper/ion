package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lawrencegripper/ion/common"
	"github.com/lawrencegripper/ion/sidecar/types"
	log "github.com/sirupsen/logrus"
)

// cSpell:ignore logrus, GUID, nolint

const (
	stateNew    = iota
	stateReady  = iota
	stateDone   = iota
	stateClosed = iota
)

//App is the sidecar application
type App struct {
	Router    *mux.Router
	Meta      types.MetadataProvider
	Publisher types.EventPublisher
	Blob      types.BlobProvider
	Logger    *log.Logger

	server          *http.Server
	secretHash      string
	baseDir         string
	context         *types.Context
	executionID     string
	validEventTypes []string
	state           int
}

//Setup initializes application
func (a *App) Setup(
	secret, baseDir string,
	context *types.Context,
	validEventTypes []string,
	meta types.MetadataProvider,
	publisher types.EventPublisher,
	blob types.BlobProvider,
	logger *log.Logger) {

	types.MustNotBeNil(meta, publisher, blob, logger, context)
	types.MustNotBeEmpty(secret, context.EventID)

	a.baseDir = baseDir
	if baseDir == "" {
		a.baseDir = "/ion/"
	}

	a.state = stateNew
	a.secretHash = types.Hash(secret)
	a.context = context
	a.validEventTypes = validEventTypes

	a.executionID = types.NewGUID()

	a.Meta = meta
	a.Publisher = publisher
	a.Blob = blob
	a.Logger = logger

	a.Router = mux.NewRouter()
	a.setupRoutes()
	a.setupDirs()

	a.Logger.Info("Sidecar configured")
}

//setupDirs initializes the required directories
func (a *App) setupDirs() {
	inBlobs := path.Join(a.baseDir, inputBlobDir())
	outBlobs := path.Join(a.baseDir, outputBlobDir())
	outMeta := path.Join(a.baseDir, outputMetaFile())
	outEvents := path.Join(a.baseDir, outputEventsDir())

	err := os.MkdirAll(inBlobs, 0777)
	if err != nil {
		panic(fmt.Errorf("error creating input blob directory '%s', error: '%+v'", inBlobs, err))
	}
	err = os.MkdirAll(outBlobs, 0777)
	if err != nil {
		panic(fmt.Errorf("error creating output blob directory '%s', error: '%+v'", outBlobs, err))
	}
	f, err := os.Create(outMeta)
	if err != nil {
		panic(fmt.Errorf("error creating output meta file '%s', error: '%+v'", outMeta, err))
	}
	f.Close() // nolint: errcheck
	err = os.MkdirAll(outEvents, 0777)
	if err != nil {
		panic(fmt.Errorf("error creating output event directory '%s', error: '%+v'", outEvents, err))
	}
}

//setupRoutes initializes the API routing
func (a *App) setupRoutes() {
	// Adds a simple shared secret check
	auth := AddAuth(a.secretHash)
	// Adds logging
	log := AddLog(a.Logger)

	// GET /ready
	// Gets any parent blob data and ensures the environment is ready
	onReadyHandler := http.HandlerFunc(a.OnReady)
	a.Router.Handle("/ready", log(auth(onReadyHandler))).Methods(http.MethodGet)

	// GET /done
	// Commits state (blobs, documents, events) to external providers
	onDoneHandler := http.HandlerFunc(a.OnDone)
	a.Router.Handle("/done", log(auth(onDoneHandler))).Methods(http.MethodGet)
}

//Run and block on web server
func (a *App) Run(addr string) {
	a.server = &http.Server{Addr: addr,
		Handler: a.Router}
	a.Logger.Warning(a.server.ListenAndServe())
}

//Close cleans up any external resources
func (a *App) Close() {
	a.Logger.Info("Shutting down sidecar")

	if a.state == stateClosed {
		return // Sidecar already closed
	}

	// Clear directories
	_ = os.RemoveAll(a.baseDir)

	defer a.Meta.Close()
	defer a.Publisher.Close()
	defer a.Blob.Close()
	if err := a.server.Shutdown(nil); err != nil {
		panic(err)
	}
	defer func() { a.state = stateClosed }()
}

//OnReady is called to initiate the modules environment (i.e. download any required blobs, etc.)
func (a *App) OnReady(w http.ResponseWriter, r *http.Request) {
	if a.state != stateNew {
		errStr := "/ready called whilst Sidecar is not in the 'stateNew' state."
		respondWithError(fmt.Errorf(errStr), http.StatusBadRequest, w)
		a.Logger.Error(errStr)
		return
	}
	a.Logger.WithFields(log.Fields{
		"executionID":   a.executionID,
		"eventID":       a.context.EventID,
		"correlationID": a.context.CorrelationID,
		"name":          a.context.Name,
		"timestamp":     time.Now(),
	}).Info("Ready called. Preparing module's environment")

	// Get the context of this execution
	context, err := a.getContext()
	if err != nil {
		respondWithError(fmt.Errorf("failed to get context with error '%+v'", err), http.StatusInternalServerError, w)
		return
	}
	// Only get files for events with an existing context.
	// Assume those that don't have a context are the first
	// event in the graph or orphaned.
	if context != nil {
		inBlobs := path.Join(a.baseDir, inputBlobDir())
		err = a.Blob.GetBlobs(inBlobs, context.Files)
		if err != nil {
			respondWithError(err, http.StatusInternalServerError, w)
			return
		}
		if len(context.Data) > 0 {
			b, err := json.Marshal(context.Data)
			if err != nil {
				respondWithError(err, http.StatusInternalServerError, w)
				return
			}
			inMeta := path.Join(a.baseDir, inputMetaFile())
			err = ioutil.WriteFile(inMeta, b, 0777)
			if err != nil {
				respondWithError(err, http.StatusInternalServerError, w)
				return
			}
		}
	}
	a.state = stateReady

	a.Logger.WithFields(log.Fields{
		"executionID":   a.executionID,
		"eventID":       a.context.EventID,
		"correlationID": a.context.CorrelationID,
		"name":          a.context.Name,
		"timestamp":     time.Now(),
	}).Info("Ready complete. Module's environment prepared.")

	w.WriteHeader(http.StatusOK)
}

//OnDone is called when the module is finished and wishes to commit their state to an external provider
func (a *App) OnDone(w http.ResponseWriter, r *http.Request) {
	if a.state != stateReady {
		errStr := "/done called whilst Sidecar is not in the 'stateReady' state."
		respondWithError(fmt.Errorf(errStr), http.StatusBadRequest, w)
		a.Logger.Error(errStr)
	}

	a.Logger.WithFields(log.Fields{
		"executionID":   a.executionID,
		"eventID":       a.context.EventID,
		"correlationID": a.context.CorrelationID,
		"name":          a.context.Name,
		"timestamp":     time.Now(),
	}).Info("Done called. Committing module's state.")

	outBlobs := path.Join(a.baseDir, outputBlobDir())
	outMeta := path.Join(a.baseDir, outputMetaFile())
	outEvents := path.Join(a.baseDir, outputEventsDir())

	// Synchronize blob data with external blob store
	blobURIs, err := a.commitBlob(outBlobs)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	// Clear local blob directory
	err = types.ClearDir(outBlobs)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}

	// Synchronize metadata with external document store
	err = a.commitMeta(outMeta)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	// Clear local metadata document
	err = types.RemoveFile(outMeta)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}

	// Synchronize events with external event system
	err = a.commitEvents(outEvents, blobURIs)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	// Clear local events directory
	err = types.ClearDir(outEvents)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}

	a.state = stateDone
	a.Logger.WithFields(log.Fields{
		"executionID":   a.executionID,
		"eventID":       a.context.EventID,
		"correlationID": a.context.CorrelationID,
		"name":          a.context.Name,
		"timestamp":     time.Now(),
	}).Info("Done complete. Module's state committed.")

	w.WriteHeader(http.StatusOK)
}

//commitBlob commits the blob directory to an external blob provider
func (a *App) commitBlob(blobsPath string) (map[string]string, error) {
	if _, err := os.Stat(blobsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("blob output directory '%s' does not exists '%+v'", blobsPath, err)
	}
	files, err := ioutil.ReadDir(blobsPath)
	if err != nil {
		return nil, err
	}
	// Get each of the file names in the blob's directory
	// TODO: Search recursively to support sub folders.
	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, path.Join(blobsPath, file.Name()))
	}
	blobURIs, err := a.Blob.PutBlobs(fileNames)
	if err != nil {
		return nil, fmt.Errorf("failed to commit blob: %+v", err)
	}
	a.Logger.WithFields(log.Fields{
		"executionID":   a.executionID,
		"eventID":       a.context.EventID,
		"correlationID": a.context.CorrelationID,
		"name":          a.context.Name,
		"timestamp":     time.Now(),
	}).Info("Committed blobs")
	return blobURIs, nil
}

//commitMeta commits the metadata document to an external provider
func (a *App) commitMeta(metadataPath string) error {
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return fmt.Errorf("metadata file '%s' does not exists '%+v'", metadataPath, err)
	}
	bytes, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata document '%s' with error '%+v'", metadataPath, err)
	}
	var m []common.KeyValuePair
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata '%s' with error: '%+v'", metadataPath, err)
	}
	insight := types.Insight{
		Context:     a.context,
		ExecutionID: a.executionID,
		Data:        m,
	}
	err = a.Meta.CreateInsight(&insight)
	if err != nil {
		return fmt.Errorf("failed to add metadata document '%+v' with error: '%+v'", m, err)
	}
	a.Logger.WithFields(log.Fields{
		"executionID":   a.executionID,
		"eventID":       a.context.EventID,
		"correlationID": a.context.CorrelationID,
		"name":          a.context.Name,
		"timestamp":     time.Now(),
	}).Info("Committed metadata")
	return nil
}

//commitEvents commits the events directory to an external provider
func (a *App) commitEvents(eventsPath string, blobURIs map[string]string) error {
	if _, err := os.Stat(eventsPath); os.IsNotExist(err) {
		return fmt.Errorf("events output directory '%s' does not exists '%+v'", eventsPath, err)
	}
	files, err := ioutil.ReadDir(eventsPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		fileName := file.Name()
		eventFilePath := path.Join(eventsPath, fileName)
		f, err := os.Open(eventFilePath)
		defer f.Close() // nolint: errcheck
		if err != nil {
			return fmt.Errorf("failed to read file '%s' with error: '%+v'", fileName, err)
		}
		// Decode event into map
		var eventKeyValuePairs []common.KeyValuePair
		decoder := json.NewDecoder(f)
		err = decoder.Decode(&eventKeyValuePairs)
		if err != nil {
			return fmt.Errorf("failed to unmarshal map '%s' with error: '%+v'", fileName, err)
		}

		// Check required fields
		var eventType string
		var includedFilesCSV string
		var eventTypeIndex, filesIndex int

		// For each key/value in event data array.
		for i, kvp := range eventKeyValuePairs {
			// Check the key against required keys
			switch kvp.Key {
			case types.EventType:
				// Check whether the event type is valid for this module
				if a.isValidEvent(kvp.Value) == false {
					return fmt.Errorf("this module is unable to publish event's of type '%s'", eventType)
				}
				eventType = kvp.Value
				eventTypeIndex = i
				break
			case types.FilesToInclude:
				includedFilesCSV = kvp.Value
				filesIndex = i
				break
			default:
				// Ignore non required keys
				break
			}
		}
		// Check required types are fulfilled
		if eventType == "" {
			return fmt.Errorf("all events must contain an 'eventType' field, error: '%+v'", err)
		}
		if len(includedFilesCSV) == 0 {
			return fmt.Errorf("all events must contain a 'files' field, error: '%+v'", err)
		}

		// Remove extracted files from event data array as no longer needed
		eventKeyValuePairs = types.Remove(eventKeyValuePairs, eventTypeIndex)
		eventKeyValuePairs = types.Remove(eventKeyValuePairs, filesIndex-1) // -1 as array will be shifted above

		// Get the files to include in event as an array
		fileSlice := strings.Split(includedFilesCSV, ",")

		// Append each file's name + external URI to the event data
		for _, f := range fileSlice {
			blobInfo := common.KeyValuePair{
				Key:   f,
				Value: blobURIs[f],
			}
			eventKeyValuePairs = append(eventKeyValuePairs, blobInfo)
		}

		// Create new event
		eventID := types.NewGUID()
		event := common.Event{
			PreviousStages: []string{},
			EventID:        eventID,
			Type:           eventType,
		}

		// Create a new context for the event.
		// We can only build a partial context
		// as we don't know who will process the
		// message. The context will be completed
		// in the dispatcher.
		context := &types.Context{
			CorrelationID: a.context.CorrelationID,
			ParentEventID: a.context.EventID,
			EventID:       eventID,
		}
		eventContext := types.EventContext{
			Context: context,
			Files:   fileSlice,
			Data:    eventKeyValuePairs,
		}
		err = a.Meta.CreateEventContext(&eventContext)
		if err != nil {
			return fmt.Errorf("failed to add context '%+v' with error '%+v'", eventContext, err)
		}
		err = a.Publisher.Publish(event)
		if err != nil {
			return fmt.Errorf("failed to publish event '%+v' with error '%+v'", event, err)
		}
	}
	a.Logger.WithFields(log.Fields{
		"executionID":   a.executionID,
		"eventID":       a.context.EventID,
		"correlationID": a.context.CorrelationID,
		"name":          a.context.Name,
		"timestamp":     time.Now(),
	}).Info("Committed events")
	return nil
}

//getContext get context metadata document
func (a *App) getContext() (*types.EventContext, error) {
	context, _ := a.Meta.GetEventContextByID(a.context.EventID)
	//TODO: Fail on error conditions other than not found
	return context, nil
}

//isValidEvent checks the event type is in the list of valid event types
func (a *App) isValidEvent(eventType string) bool {
	for _, validEventType := range a.validEventTypes {
		if eventType == validEventType {
			return true
		}
	}
	return false
}

//respondWithError returns a JSON formatted HTTP error
func respondWithError(err error, code int, w http.ResponseWriter) {
	errRes := &types.ErrorResponse{
		StatusCode: code,
		Message:    err.Error(),
	}
	w.Header().Set(types.ContentType, types.ContentTypeApplicationJSON)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errRes)
}

func inputBlobDir() string {
	return path.Join("in", "data")
}
func outputBlobDir() string {
	return path.Join("out", "data")
}
func outputEventsDir() string {
	return path.Join("out", "events")
}
func inputMetaFile() string {
	return path.Join("in", "meta.json")
}
func outputMetaFile() string {
	return path.Join("out", "meta.json")
}
