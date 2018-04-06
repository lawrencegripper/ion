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

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/lawrencegripper/ion/dispatcher/messaging"
	"github.com/lawrencegripper/ion/sidecar/types"
	log "github.com/sirupsen/logrus"
)

const (
	baseDir         string = "/ion"
	inputBlobDir    string = "/ion/in/data"
	inputMetaFile   string = "/ion/in/meta.json"
	outputBlobDir   string = "/ion/out/data"
	outputMetaFile  string = "/ion/out/meta.json"
	outputEventsDir string = "/ion/out/events"

	stateNew   = iota
	stateReady = iota
	stateDone  = iota
)

//App is the sidecar application
type App struct {
	Router    *mux.Router
	Meta      types.MetadataProvider
	Publisher types.EventPublisher
	Blob      types.BlobProvider
	Logger    *log.Logger

	moduleName      string
	secretHash      string
	correlationID   string
	eventID         string
	executionID     string
	validEventTypes []string
	state           int
}

//Setup initializes application
func (a *App) Setup(
	secret,
	eventID,
	correlationID,
	moduleName string,
	validEventTypes []string,
	meta types.MetadataProvider,
	publisher types.EventPublisher,
	blob types.BlobProvider,
	logger *log.Logger) {

	MustNotBeEmpty(secret, eventID)
	MustNotBeNil(meta, publisher, blob, logger)

	// Create output directories
	err := os.MkdirAll(inputBlobDir, 0777)
	if err != nil {
		panic(fmt.Errorf("error creating input blob directory '%s', error: '%+v'", inputBlobDir, err))
	}
	err = os.MkdirAll(outputBlobDir, 0777)
	if err != nil {
		panic(fmt.Errorf("error creating output blob directory '%s', error: '%+v'", outputBlobDir, err))
	}
	f, err := os.Create(outputMetaFile)
	if err != nil {
		panic(fmt.Errorf("error creating output meta file '%s', error: '%+v'", outputMetaFile, err))
	}
	f.Close() // nolint: errcheck
	err = os.MkdirAll(outputEventsDir, 0777)
	if err != nil {
		panic(fmt.Errorf("error creating output event directory '%s', error: '%+v'", outputEventsDir, err))
	}

	a.state = stateNew
	a.secretHash = Hash(secret)
	a.moduleName = moduleName
	a.eventID = eventID
	a.correlationID = correlationID
	a.validEventTypes = validEventTypes

	a.executionID = NewGUID()

	a.Meta = meta
	a.Publisher = publisher
	a.Blob = blob
	a.Logger = logger

	a.Router = mux.NewRouter()
	a.setupRoutes()

	a.Logger.Info("Sidecar configured")
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
	a.Logger.Fatal(http.ListenAndServe(addr, a.Router))
}

//Close cleans up any external resources
func (a *App) Close() {
	a.Logger.Info("Shutting down sidecar")

	// Clear directories
	_ = os.RemoveAll(baseDir)

	defer a.Meta.Close()
	defer a.Publisher.Close()
	defer a.Blob.Close()
}

//OnReady is called to initiate the modules environment (i.e. download any required blobs, etc.)
func (a *App) OnReady(w http.ResponseWriter, r *http.Request) {
	if a.state != stateNew {
		errStr := "/ready called whilst Sidecar is not in the 'stateNew' state."
		respondWithError(fmt.Errorf(errStr), http.StatusBadRequest, w)
		a.Logger.Error(errStr)
		return
	}
	a.Logger.WithFields(
		logrus.Fields{
			"executionID": a.executionID,
			"eventID":     a.eventID,
			"timestamp":   time.Now(),
		}).Info("OnReady() called")

	// Get the context of this execution
	context, err := a.getContext()
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	if context == nil {
		a.Logger.WithFields(
			logrus.Fields{
				"executionID": a.executionID,
				"eventID":     a.eventID,
				"timestamp":   time.Now(),
			}).Debug("No context passed, assuming first or orphan")
	} else {
		// Download the necessary files for the module
		err = a.Blob.GetBlobs(inputBlobDir, context.Files)
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
			err = ioutil.WriteFile(inputMetaFile, b, 0777)
			if err != nil {
				respondWithError(err, http.StatusInternalServerError, w)
				return
			}
		}
	}

	a.Logger.WithFields(
		logrus.Fields{
			"correlationID": a.correlationID,
			"executionID":   a.executionID,
			"eventID":       a.eventID,
			"timestamp":     time.Now(),
		}).Info("OnReady() complete")

	a.state = stateReady
	// Return
	w.WriteHeader(http.StatusOK)
}

//OnDone is called when the module is finished and wishes to commit their state to an external provider
func (a *App) OnDone(w http.ResponseWriter, r *http.Request) {
	if a.state != stateReady {
		errStr := "/done called whilst Sidecar is not in the 'stateReady' state."
		respondWithError(fmt.Errorf(errStr), http.StatusBadRequest, w)
		a.Logger.Error(errStr)
	}

	a.Logger.WithFields(
		logrus.Fields{
			"correlationID": a.correlationID,
			"executionID":   a.executionID,
			"eventID":       a.eventID,
			"timestamp":     time.Now(),
		}).Info("OnDone() called")

	// Synchronize blob data with external blob store
	blobSASUris, err := a.commitBlob(outputBlobDir)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	// Clear local blob directory
	err = ClearDir(outputBlobDir)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}

	// Synchronize metadata with external document store
	err = a.commitMeta(outputMetaFile)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	// Clear local metadata document
	err = RemoveFile(outputMetaFile)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}

	// Synchronize events with external event system
	err = a.commitEvents(outputEventsDir, blobSASUris)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	// Clear local events directory
	err = ClearDir(outputEventsDir)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}

	a.Logger.WithFields(
		logrus.Fields{
			"correlationID": a.correlationID,
			"executionID":   a.executionID,
			"eventID":       a.eventID,
			"timestamp":     time.Now(),
		}).Info("OnDone() complete")

	a.state = stateDone
	// Return
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
	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, path.Join(outputBlobDir, file.Name()))
	}
	blobSASUris, err := a.Blob.PutBlobs(fileNames)
	if err != nil {
		return nil, fmt.Errorf("failed to commit blob: %+v", err)
	}
	return blobSASUris, nil
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
	var m []messaging.KeyValuePair
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata '%s' with error: '%+v'", metadataPath, err)
	}
	insight := types.Insight{
		ExecutionID:   a.executionID,
		CorrelationID: a.correlationID,
		EventID:       a.eventID,
		Data:          m,
	}
	err = a.Meta.CreateInsight(&insight)
	if err != nil {
		return fmt.Errorf("failed to add metadata document '%+v' with error: '%+v'", m, err)
	}
	return nil
}

//commitEvents commits the events directory to an external provider
func (a *App) commitEvents(eventsPath string, blobSASUris map[string]string) error {
	if _, err := os.Stat(eventsPath); os.IsNotExist(err) {
		return fmt.Errorf("events output directory '%s' does not exists '%+v'", eventsPath, err)
	}
	files, err := ioutil.ReadDir(eventsPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		fileName := file.Name()
		eventFilePath := path.Join(outputEventsDir, fileName)
		f, err := os.Open(eventFilePath)
		defer f.Close() // nolint: errcheck
		if err != nil {
			return fmt.Errorf("failed to read file '%s' with error: '%+v'", fileName, err)
		}
		// Deserialize event into map
		var kvps []messaging.KeyValuePair
		decoder := json.NewDecoder(f)
		err = decoder.Decode(&kvps)
		if err != nil {
			return fmt.Errorf("failed to unmarshal map '%s' with error: '%+v'", fileName, err)
		}
		// Check required fields
		var eventType string
		for i, kvp := range kvps {
			if kvp.Key == types.EventType {
				var ok bool
				eventType, ok = kvp.Value.(string)
				if !ok {
					continue
				}
				kvps = Remove(kvps, i)
				break
			}
		}
		if eventType == "" {
			return fmt.Errorf("all events must contain an 'eventType' field, error: '%+v'", err)
		}

		var files string
		for i, kvp := range kvps {
			if kvp.Key == types.IncludedFiles {
				fmt.Printf("kvp: %+v", kvp)
				var ok bool
				files, ok = kvp.Value.(string)
				if !ok {
					continue
				}
				kvps = Remove(kvps, i)
				break
			}
		}
		if len(files) == 0 {
			return fmt.Errorf("all events must contain a 'files' field, error: '%+v'", err)
		}

		var isValid bool
		for _, et := range a.validEventTypes {
			if et == eventType {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("this module is unable to publish event's of type '%s'", eventType)
		}
		fileSlice := strings.Split(files, ",")

		// Create new event
		eventID := NewGUID()
		event := messaging.Event{
			PreviousStages: []string{},
			EventID:        eventID,
			Type:           eventType,
		}

		// Create new context document
		eventContext := types.EventContext{
			EventID:       eventID,
			CorrelationID: a.correlationID,
			ParentEventID: a.eventID,
			Files:         fileSlice,
			Data:          kvps,
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
	return nil
}

//getContext get context metadata document
func (a *App) getContext() (*types.EventContext, error) {
	context, _ := a.Meta.GetEventContextByID(a.eventID)
	//TODO: Fail on error conditions other than not found
	return context, nil
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
