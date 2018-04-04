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
	"github.com/lawrencegripper/ion/sidecar/types"
	log "github.com/sirupsen/logrus"
)

const (
	requestID       string = "request-id"
	inputBlobDir    string = "in/data"
	outputBlobDir   string = "out/data"
	outputMetaFile  string = "out/meta.json"
	outputEventsDir string = "out/events"
)

//TODO:
// - API versioning
// - Stop eventID being globally unique in metastore

const (
	STATE_NEW   = iota
	STATE_READY = iota
	STATE_DONE  = iota
)

//App is the sidecar application
type App struct {
	Router    *mux.Router
	Meta      types.MetadataProvider
	Publisher types.EventPublisher
	Blob      types.BlobProvider
	Logger    *log.Logger

	moduleName    string
	secretHash    string
	parentEventID string
	correlationID string
	eventID       string
	executionID   string
	state         int
	debug         bool
}

//Setup initializes application
func (a *App) Setup(
	secret,
	eventID,
	executionID,
	moduleName string,
	meta types.MetadataProvider,
	publisher types.EventPublisher,
	blob types.BlobProvider,
	debug bool,
	logger *log.Logger) {

	MustNotBeEmpty(secret, eventID, executionID)
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
	f.Close()
	err = os.MkdirAll(outputEventsDir, 0777)
	if err != nil {
		panic(fmt.Errorf("error creating output event directory '%s', error: '%+v'", outputEventsDir, err))
	}

	a.debug = debug
	a.state = STATE_NEW
	a.secretHash = Hash(secret)
	a.moduleName = moduleName
	a.eventID = eventID
	a.executionID = executionID

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

	// Clear output directories
	_ = os.RemoveAll(inputBlobDir)
	_ = os.RemoveAll(outputBlobDir)
	_ = os.RemoveAll(outputEventsDir)
	_ = os.Remove(outputMetaFile)

	defer a.Meta.Close()
	defer a.Publisher.Close()
	defer a.Blob.Close()
}

//OnReady is called to initiate the modules environment (i.e. download any required blobs, etc.)
func (a *App) OnReady(w http.ResponseWriter, r *http.Request) {
	if a.state != STATE_NEW && !a.debug {
		errStr := "/ready called whilst Sidecar is not in the 'STATE_NEW' state."
		respondWithError(fmt.Errorf(errStr), http.StatusBadRequest, w)
		a.Logger.Error(errStr)
		return
	}
	a.Logger.WithFields(
		logrus.Fields{
			"executionID": a.executionID,
			"eventID":     a.eventID,
			"timestamp":   time.Now(),
		}).Info("Initializing execution environment")

	// Get the context of this execution
	context, err := a.getContext(a.executionID)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	a.correlationID = context.CorrelationID
	a.parentEventID = context.ParentEventID

	// Download the necessary files for the module
	files, err := getFiles(context)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	err = a.Blob.GetBlobs("in/data", files)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}

	a.Logger.WithFields(
		logrus.Fields{
			"correlationID": a.correlationID,
			"executionID":   a.executionID,
			"parentEventID": a.parentEventID,
			"eventID":       a.eventID,
			"timestamp":     time.Now(),
		}).Info("Completed initializing execution environment")

	a.state = STATE_READY
	// Return
	w.WriteHeader(http.StatusOK)
}

//OnDone is called when the module is finished and wishes to commit their state to an external provider
func (a *App) OnDone(w http.ResponseWriter, r *http.Request) {
	if a.state != STATE_READY && !a.debug {
		errStr := "/done called whilst Sidecar is not in the 'STATE_READY' state."
		respondWithError(fmt.Errorf(errStr), http.StatusBadRequest, w)
		a.Logger.Error(errStr)
	}

	a.Logger.WithFields(
		logrus.Fields{
			"correlationID": a.correlationID,
			"executionID":   a.executionID,
			"parentEventID": a.parentEventID,
			"eventID":       a.eventID,
			"timestamp":     time.Now(),
		}).Info("Committing state")

	// Synchronize blob data with external blob store
	blobsPath := path.Join("out", "data")
	err := a.commitBlob(blobsPath)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	// Clear local blob directory
	err = ClearDir(blobsPath)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}

	// Synchronize metadata with external document store
	metadataPath := path.Join("out", "meta.json")
	err = a.commitMeta(metadataPath)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	// Clear local metadata document
	err = RemoveFile(metadataPath)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}

	// Synchronize events with external event system
	eventsPath := path.Join("out", "events")
	err = a.commitEvents(eventsPath)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	// Clear local events directory
	err = ClearDir(eventsPath)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}

	a.state = STATE_DONE
}

//commitBlob commits the blob directory to an external blob provider
func (a *App) commitBlob(blobsPath string) error {
	//TODO: Validate blobsPath

	files, err := ioutil.ReadDir(blobsPath)
	if err != nil {
		return err
	}
	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, path.Join(outputBlobDir, file.Name()))
	}
	err = a.Blob.CreateBlobs(fileNames)
	if err != nil {
		return fmt.Errorf("failed to commit blob: %+v", err)
	}
	return nil
}

//commitMeta commits the metadata document to an external provider
func (a *App) commitMeta(metadataPath string) error {
	//TODO: Validate metadata file path

	bytes, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata document '%s' with error '%+v'", metadataPath, err)
	}
	var m map[string]string
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return fmt.Errorf("failed to unmarshal map '%s' with error: '%+v'", metadataPath, err)
	}
	if m[types.IncludedFiles] == "" {
		return fmt.Errorf("metadata requires a '%s' field", types.IncludedFiles)
	}
	metadata := types.Metadata{
		ExecutionID:   a.executionID,
		CorrelationID: a.correlationID,
		ParentEventID: a.parentEventID,
		Data:          m,
	}
	err = a.Meta.UpsertMetadataDocument(&metadata)
	if err != nil {
		return fmt.Errorf("failed to add metadata document '%+v' with error: '%+v'", metadata, err)
	}
	return nil
}

//commitEvents commits the events directory to an external provider
func (a *App) commitEvents(eventsPath string) error {
	//TODO: Validate events directory path

	files, err := ioutil.ReadDir(eventsPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		fileName := file.Name()
		f, err := os.Open(fileName)
		defer f.Close()
		if err != nil {
			return fmt.Errorf("failed to read file '%s' with error: '%+v'", fileName, err)
		}
		// Deserialize event into map
		var m map[string]string
		decoder := json.NewDecoder(f)
		err = decoder.Decode(&m)
		if err != nil {
			return fmt.Errorf("failed to unmarshal map '%s' with error: '%+v'", fileName, err)
		}
		// Check required fields
		eventType := m[types.EventType]
		if eventType == "" {
			return fmt.Errorf("all events must contain an 'eventType' field, error: '%+v'", err)
		}
		delete(m, types.EventType)

		// Create new event
		//TODO: Generate execution ID for each event...
		executionID := NewExecutionID(a.moduleName)
		event := types.Event{
			PreviousStages: []string{},
			ExecutionID:    executionID,
			Type:           eventType,
		}

		// Create new context document
		contextMetadata := types.Metadata{
			ExecutionID:   executionID,
			CorrelationID: a.correlationID,
			ParentEventID: a.eventID,
			Data:          m,
		}
		err = a.Meta.UpsertMetadataDocument(&contextMetadata)
		if err != nil {
			return fmt.Errorf("failed to add context '%+v' with error '%+v'", contextMetadata, err)
		}
		err = a.Publisher.Publish(event)
		if err != nil {
			return fmt.Errorf("failed to publish event '%+v' with error '%+v'", event, err)
		}
	}
	return nil
}

//getContext get context metadata document
func (a *App) getContext(executionID string) (*types.Metadata, error) {
	context, err := a.Meta.GetMetadataDocumentByID(executionID)
	if err != nil {
		return nil, fmt.Errorf("failed getting context document using ID '%s' with error: '%+v'", executionID, err)
	}
	return context, nil
}

//getFiles from context metadata document
func getFiles(metadata *types.Metadata) ([]string, error) {
	fileCSV, exist := metadata.Data[types.IncludedFiles]
	if !exist {
		return nil, fmt.Errorf("required key '%s' not found in metadata", types.IncludedFiles)
	}
	files := strings.Split(fileCSV, ",")
	return files, nil
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
