package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	c "github.com/lawrencegripper/mlops/sidecar/common"
	log "github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/forward"
)

//TODO: API versioning

const secretHeaderKey = "secret"
const requestIDHeaderKey = "request-id"

//App is the sidecar application
type App struct {
	Router    *mux.Router
	MetaDB    c.MetaDB
	Publisher c.EventPublisher
	Blob      c.BlobProvider
	Logger    *log.Logger
	Proxy     *forward.Forwarder

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
	metaDB c.MetaDB,
	publisher c.EventPublisher,
	blob c.BlobProvider,
	logger *log.Logger) {

	a.Proxy, _ = forward.New(
		forward.Stream(true),
	)

	c.MustNotBeEmpty(secret, eventID, correlationID, parentEventID)
	c.MustNotBeNil(metaDB, publisher, blob, logger, a.Proxy)

	a.secretHash = c.Hash(secret)
	a.eventID = eventID
	a.correlationID = correlationID
	a.parentEventID = parentEventID

	a.MetaDB = metaDB
	a.Publisher = publisher
	a.Blob = blob
	a.Logger = logger

	a.Router = mux.NewRouter()
	a.setupRoutes()

	a.Logger.Info("Sidecar initialized!")
}

//setupRoutes initializes the API routing
func (a *App) setupRoutes() {
	// Adds a simple shared secret check to each request
	auth := AddAuth(a.secretHash)
	// Adds logging to each request
	log := AddLog(a.Logger)
	// Adds a identity header to each request
	self := AddIdentity(a.eventID)
	parent := AddIdentity(a.parentEventID)

	// GET /meta
	// Returns all metadata currently stored as part of this chain
	getMeta := http.HandlerFunc(a.GetAllMeta)
	a.Router.Handle("/meta", log(auth(getMeta))).Methods("GET")

	// GET /parent/meta
	// Returns only metadata stored by the parent of this module
	getParentMeta := http.HandlerFunc(a.GetMetaByID)
	a.Router.Handle("/parent/meta", log(auth(parent(getParentMeta)))).Methods("GET")

	// POST /self/meta
	// Stores metadata against this modules meta store
	updateSelfMeta := http.HandlerFunc(a.UpdateMeta)
	a.Router.Handle("/self/meta", log(auth(updateSelfMeta))).Methods("PUT")

	// GET /self/meta
	// Returns the metadata currently in this modules meta store
	getSelfMeta := http.HandlerFunc(a.GetMetaByID)
	a.Router.Handle("/self/meta", log(auth(self(getSelfMeta)))).Methods("GET")

	// GET /parent/blob
	// Returns a named blob from the parent's blob store
	getParentBlob := http.HandlerFunc(AddProxy(a, nil,
		a.Blob.Resolve,
		respondWithNothing))
	a.Router.Handle("/parent/blob", log(auth(parent(getParentBlob)))).Methods("GET")

	// GET /self/blob
	// Returns a named blob from this modules blob store
	getSelfBlob := http.HandlerFunc(AddProxy(a, nil,
		a.Blob.Resolve,
		respondWithNothing))
	a.Router.Handle("/self/blob", log(auth(self(getSelfBlob)))).Methods("GET")

	// DELETE /self/blob
	// Returns a named blob from this modules blob store
	deleteSelfBlobs := http.HandlerFunc(a.DeleteBlobs)
	a.Router.Handle("/self/blob", log(auth(self(deleteSelfBlobs)))).Methods("DELETE")

	// PUT /self/blob
	// Stores a blob in this modules blob store
	addSelfBlob := http.HandlerFunc(AddProxy(a, map[string]string{
		"x-ms-blob-type": "BlockBlob",
	},
		a.Blob.Create,
		respondWithCreatedURI))
	a.Router.Handle("/self/blob", log(auth(self(addSelfBlob)))).Methods("PUT")

	// GET /self/blobs
	// Returns a list of blobs currently stored in this modules blob store
	listSelfBlobs := http.HandlerFunc(a.ListBlobs)
	a.Router.Handle("/self/blobs", log(auth(self(listSelfBlobs)))).Methods("GET")

	// POST /events
	// Publishes a new event to the messaging system
	publishEventHandler := http.HandlerFunc(a.Publish)
	a.Router.Handle("/events", log(auth(publishEventHandler))).Methods("POST")
}

//Run and block on web server
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

//Close cleans up any external resources
func (a *App) Close() {
	a.Logger.Info("Closing sidecar")

	defer a.MetaDB.Close()
	defer a.Publisher.Close()
}

//GetMetaByID gets the metadata document with the associated ID
// nolint: errcheck
func (a *App) GetMetaByID(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get(requestIDHeaderKey)
	if id == "" {
		id = a.eventID
	}

	w.Header().Set("Content-Type", "application/json")
	doc, err := a.MetaDB.GetMetaDocByID(id)
	if err != nil {
		respondWithError(err, http.StatusNotFound, w)
		return
	}
	docs, _ := c.StripBlobStore([]c.MetaDoc{*doc})
	err = json.NewEncoder(w).Encode(docs[0].Metadata)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
}

//GetAllMeta get all the metadata documents with the associated correlation ID
// nolint: errcheck
func (a *App) GetAllMeta(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	docs, err := a.MetaDB.GetMetaDocAll(a.correlationID)
	if err != nil {
		respondWithError(err, http.StatusNotFound, w)
		return
	}
	docs, _ = c.StripBlobStore(docs)
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

	doc := &c.MetaDoc{
		ID:            a.eventID,
		CorrelationID: a.correlationID,
		ParentEventID: a.parentEventID,
		Metadata:      metadata,
	}
	err = a.MetaDB.AddOrUpdateMetaDoc(doc)
	if err != nil {
		respondWithError(err, http.StatusNotFound, w)
		return
	}
	w.WriteHeader(http.StatusOK)
}

//ListBlobs returns a list of blobs stored by the current module
func (a *App) ListBlobs(w http.ResponseWriter, r *http.Request) {
	resource := a.eventID
	blobs, err := a.Blob.List(resource)
	if err != nil {
		respondWithError(err, http.StatusNotFound, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(blobs)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
}

//DeleteBlobs deletes a named blob resource
func (a *App) DeleteBlobs(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("res")
	id := r.Header.Get(requestIDHeaderKey)
	if id == "" || resource == "" {
		respondWithError(fmt.Errorf("resource or id could not be found in request"), http.StatusBadRequest, w)
		return
	}
	resourceID := id + "/" + resource
	err := a.Blob.Delete(resourceID)
	if err != nil {
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
	//TODO: validate event before publishing
	event := c.Event{
		PreviousStages: nil,
		ParentEventID:  a.eventID,
		Data:           eventData,
		Type:           "EVENTTYPE",
	}
	err = a.Publisher.Publish(event)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

//respondWithError returns a JSON formatted HTTP error
func respondWithError(err error, code int, w http.ResponseWriter) {
	errRes := &c.ErrorResponse{
		StatusCode: code,
		Message:    err.Error(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errRes)
}

//respondWithCreatedURI returns the URL for a newly created resource
func respondWithCreatedURI(resourceURI string, w StatusCodeResponseWriter) {
	if w.statusCode == http.StatusCreated {
		uri := strings.Split(resourceURI, "?")
		w.Write([]byte(uri[0]))
	}
}

//responds with nothing
func respondWithNothing(resource string, w StatusCodeResponseWriter) {
	return
}
