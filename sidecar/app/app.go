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
	. "github.com/lawrencegripper/ion/sidecar/types"
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
	Meta      MetadataProvider
	Publisher EventPublisher
	Blob      BlobProvider
	Logger    *log.Logger

	server          *http.Server
	secretHash      string
	baseDir         string
	context         *Context
	executionID     string
	validEventTypes []string
	state           int
	development     bool
}

//Setup initializes application
func (a *App) Setup(
	secret, baseDir string,
	context *Context,
	validEventTypes []string,
	meta MetadataProvider,
	publisher EventPublisher,
	blob BlobProvider,
	logger *log.Logger,
	developmentMode bool) {

	MustNotBeNil(meta, publisher, blob, logger, context)
	MustNotBeEmpty(secret, context.EventID)

	a.baseDir = baseDir
	if baseDir == "" {
		a.baseDir = "/ion/"
	}

	a.state = stateNew
	a.secretHash = Hash(secret)
	a.context = context
	a.validEventTypes = validEventTypes

	a.executionID = NewGUID()

	a.Meta = meta
	a.Publisher = publisher
	a.Blob = blob
	a.Logger = logger

	a.development = developmentMode

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

	if err := CreateDirClean(inBlobs); err != nil {
		panic(fmt.Sprintf("could not create input blob directory, %+v", err))
	}
	if err := CreateDirClean(outBlobs); err != nil {
		panic(fmt.Sprintf("could not create output blob directory, %+v", err))
	}
	if err := CreateFileClean(outMeta); err != nil {
		panic(fmt.Sprintf("could not create output meta file, %+v", err))
	}
	if err := CreateDirClean(outEvents); err != nil {
		panic(fmt.Sprintf("could not create output events directory, %+v", err))
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
	blobURIs, err := a.CommitBlob(outBlobs)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	// Clear local blob directory
	err = ClearDir(outBlobs)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}

	// Synchronize metadata with external document store
	err = a.CommitMeta(outMeta)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	// Clear local metadata document
	err = RemoveFile(outMeta)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}

	// Synchronize events with external event system
	err = a.CommitEvents(outEvents, blobURIs)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	// Clear local events directory
	err = ClearDir(outEvents)
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

//CommitBlob commits the blob directory to an external blob provider
func (a *App) CommitBlob(blobsPath string) (map[string]string, error) {
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

//CommitMeta commits the metadata document to an external provider
func (a *App) CommitMeta(metadataPath string) error {
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return fmt.Errorf("metadata file '%s' does not exists '%+v'", metadataPath, err)
	}
	bytes, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata document '%s' with error '%+v'", metadataPath, err)
	}
	if len(bytes) == 0 {
		return nil // Handle no metadata
	}
	var m common.KeyValuePairs
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata '%s' with error: '%+v'", metadataPath, err)
	}
	insight := Insight{
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
	if a.development {
		_ = writeDevelopmentFile("meta.json", insight)
	}
	return nil
}

//CommitEvents commits the events directory to an external provider
func (a *App) CommitEvents(eventsPath string, blobURIs map[string]string) error {
	if _, err := os.Stat(eventsPath); os.IsNotExist(err) {
		return fmt.Errorf("events output directory '%s' does not exists '%+v'", eventsPath, err)
	}
	files, err := ioutil.ReadDir(eventsPath)
	if err != nil {
		return err
	}
	// Read each of the event files stored in the
	// output events directory. Events will be
	// deserialized into an expected structure.
	// Enriched, validated and then split into
	// an event to send via the messaging system
	// and a context document for the event to
	// reference.
	for _, file := range files {
		fileName := file.Name()
		eventFilePath := path.Join(eventsPath, fileName)
		f, err := os.Open(eventFilePath)
		defer f.Close() // nolint: errcheck
		if err != nil {
			return fmt.Errorf("failed to read file '%s' with error: '%+v'", fileName, err)
		}
		// Decode event into map
		var keyValuePairs common.KeyValuePairs
		decoder := json.NewDecoder(f)
		err = decoder.Decode(&keyValuePairs)
		if err != nil {
			return fmt.Errorf("failed to unmarshal map '%s' with error: '%+v'", fileName, err)
		}

		var eventType string
		var includedFilesCSV string
		var eventTypeIndex, filesIndex int

		// For each key/value in event data array.
		for i, kvp := range keyValuePairs {
			// Check the key against required keys
			switch kvp.Key {
			case EventType:
				// Check whether the event type is valid for this module
				if ContainsString(a.validEventTypes, kvp.Value) == false {
					return fmt.Errorf("this module is unable to publish event's of type '%s'", eventType)
				}
				eventType = kvp.Value
				eventTypeIndex = i
				break
			case FilesToInclude:
				includedFilesCSV = kvp.Value
				filesIndex = i
				break
			default:
				// Ignore non required keys
				break
			}
		}
		itemsRemoved := 0

		// [Required] Check that the key 'eventType' was found in the data
		// if it wasn't return an error. If it was, remove it
		// from the key value pairs as it is no longer needed
		if eventType == "" {
			return fmt.Errorf("all events must contain an 'eventType' field, error: '%+v'", err)
		}
		if err := keyValuePairs.Remove(eventTypeIndex); err != nil {
			return fmt.Errorf("error removing event type from metadata: '%+v'", err)
		}
		itemsRemoved++

		// [Optional] Check whether the key 'files' was supplied in order
		// to pass file references to event context. If it wasn't, log it
		// and ignore it. If it was, remove it from the key value pairs
		// as it is no longer needed and then add the file list and their
		// blob uri for each of the files to the event context.
		var fileSlice []string
		if len(includedFilesCSV) == 0 {
			a.Logger.WithFields(log.Fields{
				"executionID":   a.executionID,
				"eventID":       a.context.EventID,
				"correlationID": a.context.CorrelationID,
				"name":          a.context.Name,
				"timestamp":     time.Now(),
			}).Debug("Event contains no file references")
		} else {
			if err := keyValuePairs.Remove(filesIndex - itemsRemoved); err != nil {
				return fmt.Errorf("error removing event type from metadata: '%+v'", err)
			}
			itemsRemoved++
			fileSlice = strings.Split(includedFilesCSV, ",")
			for _, f := range fileSlice {
				blobInfo := common.KeyValuePair{
					Key:   f,
					Value: blobURIs[f],
				}
				keyValuePairs.Append(blobInfo)
			}
		}

		// Create new event to publish
		// via the messaging system
		eventID := NewGUID()
		event := common.Event{
			PreviousStages: []string{},
			EventID:        eventID,
			Type:           eventType,
		}

		// Create a new context for the event
		// to reference. We can only build a
		// partial context as we don't know which
		// modules will process the message.
		// The context will be completed later.
		context := &Context{
			CorrelationID: a.context.CorrelationID,
			ParentEventID: a.context.EventID,
			EventID:       eventID,
		}
		eventContext := EventContext{
			Context: context,
			Files:   fileSlice,
			Data:    keyValuePairs,
		}
		err = a.Meta.CreateEventContext(&eventContext)
		if err != nil {
			return fmt.Errorf("failed to add context '%+v' with error '%+v'", eventContext, err)
		}
		err = a.Publisher.Publish(event)
		if err != nil {
			return fmt.Errorf("failed to publish event '%+v' with error '%+v'", event, err)
		}
		if a.development {
			_ = writeDevelopmentFile("context_"+fileName, eventContext)
			_ = writeDevelopmentFile("event_"+fileName, eventContext)
		}
	}
	return nil
}

func writeDevelopmentFile(fileName string, obj interface{}) error {
	if _, err := os.Stat(outputDevDir()); os.IsNotExist(err) {
		_ = os.Mkdir(outputDevDir(), 0777)
	}
	// TODO: Handle errors here?
	path := path.Join(outputDevDir(), "dev."+fileName)
	b, err := json.Marshal(&obj)
	if err != nil {
		return fmt.Errorf("error generating development logs, '%+v'", err)
	}
	err = ioutil.WriteFile(path, b, 0777)
	if err != nil {
		return fmt.Errorf("error writing development logs, '%+v'", err)
	}
	return nil
}

//getContext get context metadata document
func (a *App) getContext() (*EventContext, error) {
	context, _ := a.Meta.GetEventContextByID(a.context.EventID)
	//TODO: Fail on error conditions other than not found
	return context, nil
}

//respondWithError returns a JSON formatted HTTP error
func respondWithError(err error, code int, w http.ResponseWriter) {
	errRes := &ErrorResponse{
		StatusCode: code,
		Message:    err.Error(),
	}
	w.Header().Set(ContentType, ContentTypeApplicationJSON)
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
func outputDevDir() string {
	return path.Join("out", "dev")
}
