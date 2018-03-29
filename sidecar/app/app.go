package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lawrencegripper/mlops/sidecar/types"
	log "github.com/sirupsen/logrus"
)

const requestID string = "request-id"

//TODO:
// - API versioning
// - Stop eventID being globally unique in metastore

//App is the sidecar application
type App struct {
	Router    *mux.Router
	Meta      types.MetaProvider
	Publisher types.EventPublisher
	Blob      types.BlobProvider
	Logger    *log.Logger

	secretHash    string
	parentEventID string
	correlationID string
	eventID       string
}

//Setup initializes application
func (a *App) Setup(
	secret,
	eventID,
	correlationID,
	parentEventID string,
	Meta types.MetaProvider,
	publisher types.EventPublisher,
	blob types.BlobProvider,
	logger *log.Logger) {

	MustNotBeEmpty(secret, eventID, correlationID, parentEventID)
	MustNotBeNil(Meta, publisher, blob, logger)

	a.secretHash = Hash(secret)
	a.eventID = eventID
	a.correlationID = correlationID
	a.parentEventID = parentEventID

	a.Meta = Meta
	a.Publisher = publisher
	a.Blob = blob
	a.Logger = logger

	a.Router = mux.NewRouter()
	a.setupRoutes()

	a.Logger.Info("Sidecar initialized!")
}

//setupRoutes initializes the API routing
func (a *App) setupRoutes() {
	// Adds a simple shared secret check
	auth := AddAuth(a.secretHash)
	// Adds logging
	log := AddLog(a.Logger)
	// Adds self identity header
	self := AddIdentity(a.eventID)
	// Adds parent identity header
	parent := AddIdentity(a.parentEventID)

	// GET /meta
	// Returns all metadata currently stored as part of this chain
	getMeta := http.HandlerFunc(a.GetAllMeta)
	a.Router.Handle("/meta", log(auth(getMeta))).Methods(http.MethodGet)

	// GET /parent/meta
	// Returns only metadata stored by the parent of this module
	getParentMeta := http.HandlerFunc(a.GetMetaByID)
	a.Router.Handle("/parent/meta", log(auth(parent(getParentMeta)))).Methods(http.MethodGet)

	// POST /self/meta
	// Stores metadata against this modules meta store
	updateSelfMeta := http.HandlerFunc(a.UpdateMeta)
	a.Router.Handle("/self/meta", log(auth(updateSelfMeta))).Methods(http.MethodPut)

	// GET /self/meta
	// Returns the metadata currently in this modules meta store
	getSelfMeta := http.HandlerFunc(a.GetMetaByID)
	a.Router.Handle("/self/meta", log(auth(self(getSelfMeta)))).Methods(http.MethodGet)

	// GET /parent/blob
	// Returns a named blob from the parent's blob store
	getParentBlob := http.HandlerFunc(a.GetBlob)
	a.Router.Handle("/parent/blob", log(auth(parent(getParentBlob)))).Methods(http.MethodGet)

	// GET /self/blob
	// Returns a named blob from this modules blob store
	getSelfBlob := http.HandlerFunc(a.GetBlob)
	a.Router.Handle("/self/blob", log(auth(self(getSelfBlob)))).Methods(http.MethodGet)

	// PUT /self/blob
	// Stores a blob in this modules blob store
	addSelfBlob := http.HandlerFunc(a.CreateBlob)
	a.Router.Handle("/self/blob", log(auth(self(addSelfBlob)))).Methods(http.MethodPut)

	// DELETE /self/blob
	// Deletes a named blob from this modules blob store
	deleteSelfBlobs := http.HandlerFunc(a.DeleteBlobs)
	a.Router.Handle("/self/blob", log(auth(self(deleteSelfBlobs)))).Methods(http.MethodDelete)

	// GET /self/blobs
	// Returns a list of blobs currently stored in this modules blob store
	listSelfBlobs := http.HandlerFunc(a.ListBlobs)
	a.Router.Handle("/self/blobs", log(auth(self(listSelfBlobs)))).Methods(http.MethodGet)

	// POST /events
	// Publishes a new event to the messaging system
	publishEventHandler := http.HandlerFunc(a.Publish)
	a.Router.Handle("/events", log(auth(publishEventHandler))).Methods(http.MethodPost)
}

//Run and block on web server
func (a *App) Run(addr string) {
	a.Logger.Fatal(http.ListenAndServe(addr, a.Router))
}

//Close cleans up any external resources
func (a *App) Close() {
	a.Logger.Info("Shutting down sidecar")

	defer a.Meta.Close()
	defer a.Publisher.Close()
	defer a.Blob.Close()
}

//GetMetaByID gets the metadata document with the associated ID
// nolint: errcheck
func (a *App) GetMetaByID(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get(requestID)
	if id == "" {
		id = a.eventID
	}
	w.Header().Set(types.ContentType, types.ContentTypeApplicationJSON)
	doc, err := a.Meta.GetMetaDocByID(id)
	if err != nil {
		respondWithError(err, http.StatusNotFound, w)
		return
	}
	docs, _ := StripBlobStore([]types.MetaDoc{*doc})
	err = json.NewEncoder(w).Encode(docs[0].Metadata)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
}

//GetAllMeta get all the metadata documents with the associated correlation ID
// nolint: errcheck
func (a *App) GetAllMeta(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(types.ContentType, types.ContentTypeApplicationJSON)
	docs, err := a.Meta.GetMetaDocAll(a.correlationID)
	if err != nil {
		respondWithError(err, http.StatusNotFound, w)
		return
	}
	docs, _ = StripBlobStore(docs)
	err = json.NewEncoder(w).Encode(docs)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
}

//UpdateMeta creates or updates a metadata document
// nolint: errcheck
func (a *App) UpdateMeta(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var metadata map[string]string
	err := decoder.Decode(&metadata)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	doc := &types.MetaDoc{
		ID:            a.eventID,
		CorrelationID: a.correlationID,
		ParentEventID: a.parentEventID,
		Metadata:      metadata,
	}
	err = a.Meta.AddOrUpdateMetaDoc(doc)
	if err != nil {
		respondWithError(err, http.StatusNotFound, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

//GetBlob gets a blob object
func (a *App) GetBlob(w http.ResponseWriter, r *http.Request) {
	resID, err := getResourceID(r)
	if err != nil {
		respondWithError(err, http.StatusBadRequest, w)
		return
	}
	if a.Blob.Proxy() != nil {
		r.RequestURI = "" // This is a hack to bypass this issue: https://github.com/vulcand/oxy/issues/57
		a.Blob.Proxy().Get(resID, w, r)
		return
	}
	blob, err := a.Blob.Get(resID)
	if err != nil {
		respondWithError(err, http.StatusNotFound, w)
		return
	}
	defer blob.Close()
	bytes, err := ioutil.ReadAll(blob)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	w.Write(bytes)
}

//CreateBlob creates a new blob and returns a path to it
func (a *App) CreateBlob(w http.ResponseWriter, r *http.Request) {
	resID, err := getResourceID(r)
	if err != nil {
		respondWithError(err, http.StatusBadRequest, w)
		return
	}
	if a.Blob.Proxy() != nil {
		r.RequestURI = "" // This is a hack to bypass this issue: https://github.com/vulcand/oxy/issues/57
		a.Blob.Proxy().Create(resID, w, r)
		return
	}
	path, err := a.Blob.Create(resID, r.Body)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	w.Write([]byte(path))
}

//ListBlobs returns a list of blobs stored by the current module
func (a *App) ListBlobs(w http.ResponseWriter, r *http.Request) {
	resource := a.eventID
	blobs, err := a.Blob.List(resource)
	if err != nil {
		respondWithError(err, http.StatusNotFound, w)
		return
	}
	w.Header().Set(types.ContentType, types.ContentTypeApplicationJSON)
	err = json.NewEncoder(w).Encode(blobs)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
}

//DeleteBlobs deletes a named blob resource
func (a *App) DeleteBlobs(w http.ResponseWriter, r *http.Request) {
	resID, err := getResourceID(r)
	if err != nil {
		respondWithError(err, http.StatusBadRequest, w)
		return
	}
	deleted, err := a.Blob.Delete(resID)
	if err != nil || !deleted {
		respondWithError(err, http.StatusNotFound, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

//Publish posts an event to the event topic
// nolint: errcheck
func (a *App) Publish(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var eventData map[string]string
	err := decoder.Decode(&eventData)
	if err != nil {
		respondWithError(err, http.StatusBadRequest, w)
		return
	}
	defer r.Body.Close()
	eventType := eventData[types.EventType]
	if eventType == "" {
		respondWithError(fmt.Errorf("metadata must contain an event type with the key '%s'", types.EventType), http.StatusBadRequest, w)
		return
	}
	delete(eventData, types.EventType)
	//TODO: validate event before publishing
	event := types.Event{
		PreviousStages: []string{},
		CorrelationID:  a.correlationID,
		ParentEventID:  a.eventID,
		Data:           eventData,
		Type:           eventType,
	}
	err = a.Publisher.Publish(event)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w) //TODO: Return proper error codes
		return
	}
	w.WriteHeader(http.StatusCreated)
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
