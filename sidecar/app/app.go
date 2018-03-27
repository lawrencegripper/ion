package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lawrencegripper/mlops/sidecar/types"
	log "github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/forward"
)

//TODO: API versioning

const requestID string = "request-id"
const contentType string = "Content-Type"
const applicationJSON string = "application/json"

//App is the sidecar application
type App struct {
	Router    *mux.Router
	Meta      types.MetaProvider
	Publisher types.EventPublisher
	Blob      types.BlobProvider
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
	Meta types.MetaProvider,
	publisher types.EventPublisher,
	blob types.BlobProvider,
	logger *log.Logger) {

	a.Proxy, _ = forward.New(
		forward.Stream(true),
	)

	MustNotBeEmpty(secret, eventID, correlationID, parentEventID)
	MustNotBeNil(Meta, publisher, blob, logger, a.Proxy)

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
	// Adds get blob resolver
	get := AddResolver(a.Blob.ResolveGet)
	// Adds create blob resolver
	create := AddResolver(a.Blob.ResolveCreate)

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
	getParentBlobProxy := http.HandlerFunc(AddProxy(a.Proxy, nil))
	a.Router.Handle("/parent/blob", log(auth(parent(get(getParentBlobProxy))))).Methods("GET")

	// GET /self/blob
	// Returns a named blob from this modules blob store
	getSelfBlobProxy := http.HandlerFunc(AddProxy(a.Proxy, nil))
	a.Router.Handle("/self/blob", log(auth(self(get(getSelfBlobProxy))))).Methods("GET")

	// PUT /self/blob
	// Stores a blob in this modules blob store
	addSelfBlobProxy := http.HandlerFunc(AddProxy(a.Proxy, respondWithCreatedURI))
	a.Router.Handle("/self/blob", log(auth(self(create(addSelfBlobProxy))))).Methods("PUT")

	// DELETE /self/blob
	// Deletes a named blob from this modules blob store
	deleteSelfBlobs := http.HandlerFunc(a.DeleteBlobs)
	a.Router.Handle("/self/blob", log(auth(self(deleteSelfBlobs)))).Methods("DELETE")

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
	w.Header().Set(contentType, applicationJSON)
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
	w.Header().Set(contentType, applicationJSON)
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

//ListBlobs returns a list of blobs stored by the current module
func (a *App) ListBlobs(w http.ResponseWriter, r *http.Request) {
	resource := a.eventID
	blobs, err := a.Blob.List(resource)
	if err != nil {
		respondWithError(err, http.StatusNotFound, w)
		return
	}
	w.Header().Set(contentType, applicationJSON)
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
	//TODO: validate event before publishing
	event := types.Event{
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
	errRes := &types.ErrorResponse{
		StatusCode: code,
		Message:    err.Error(),
	}
	w.Header().Set(contentType, applicationJSON)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errRes)
}

//respondWithCreatedURI returns the URL for a newly created resource
func respondWithCreatedURI(w *StatusCodeResponseWriter) {
	if w.statusCode == http.StatusCreated {
		resID := w.ResponseWriter.Header().Get("resource-id")
		if resID != "" {
			uri := strings.Split(resID, "?") // remove SAS token
			if len(uri) < 2 {
				respondWithError(fmt.Errorf("invalid URL '%s' provided by blob provider", uri), http.StatusInternalServerError, w)
				return
			}
			_, err := w.Write([]byte(uri[0]))
			if err != nil {
				panic(err)
			}
		}
	}
}
