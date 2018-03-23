package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	c "github.com/lawrencegripper/mlops/sidecar/common"
	log "github.com/sirupsen/logrus"
)

const secretHeaderKey = "secret"

//App is the sidecar application
type App struct {
	Router    *mux.Router
	MetaDB    c.MetaDB
	Publisher c.Publisher
	Blob      c.BlobStorage
	Logger    *log.Logger

	secretHash    string
	blobAccessKey string
	parentEventID string
	correlationID string
	eventID       string
}

//Setup initializes application
func (a *App) Setup(
	secret,
	blobAccessKey,
	eventID,
	correlationID,
	parentEventID string,
	metaDB c.MetaDB,
	publisher c.Publisher,
	blob c.BlobStorage,
	logger *log.Logger) {

	c.MustNotBeEmpty(secret, blobAccessKey, eventID, correlationID, parentEventID)
	c.MustNotBeNil(metaDB, publisher, blob, logger)

	a.secretHash = c.Hash(secret)
	a.blobAccessKey = blobAccessKey
	a.eventID = eventID
	a.correlationID = correlationID
	a.parentEventID = parentEventID

	a.MetaDB = metaDB
	a.Publisher = publisher
	a.Blob = blob
	a.Logger = logger

	a.Router = mux.NewRouter()
	a.setupRoutes()
}

//setupRoutes initializes the API routing
func (a *App) setupRoutes() {
	auth := AddAuth(a.secretHash)
	log := AddLog(a.Logger)
	self := AddIdentity(a.eventID)
	parent := AddIdentity(a.parentEventID)

	// Gets all metadata associated with correlation ID
	getAllMeta := http.HandlerFunc(a.GetAllMeta)
	a.Router.Handle("/meta", log(auth(getAllMeta))).Methods("GET")

	// Gets specific metadata associated with source ID
	getMetaInputs := http.HandlerFunc(a.GetMetaByID)
	a.Router.Handle("/meta/inputs", parent(log(auth(getMetaInputs)))).Methods("GET")

	// Update metadata associated with this event ID
	updateMeta := http.HandlerFunc(a.UpdateMeta)
	a.Router.Handle("/meta/outputs", log(auth(updateMeta))).Methods("POST")

	// Get existing metadata data associated with this event ID
	getMetaOutputs := http.HandlerFunc(a.GetMetaByID)
	a.Router.Handle("/meta/outputs", self(log(auth(getMetaOutputs)))).Methods("GET")

	// Get an authenticated URL for the given blob resource
	getBlobInputs := http.HandlerFunc(a.GetBlobInputs)
	a.Router.Handle("/blob/inputs", log(auth(getBlobInputs))).Methods("GET")

	// Create new blob storage location if one doesn't exist and return authenticated URL
	getBlobOutputs := http.HandlerFunc(a.GetBlobOutputs)
	a.Router.Handle("/blob/outputs", log(auth(getBlobOutputs))).Methods("GET")

	// Publish a new event
	publishEventHandler := http.HandlerFunc(a.PublishEvent)
	a.Router.Handle("/events", log(auth(publishEventHandler))).Methods("POST")
}

//Run and block on web server
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

//GetMetaByID gets the metadata document with the associated ID
// nolint: errcheck
func (a *App) GetMetaByID(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get("request-id")
	if id == "" {
		id = a.eventID
	}

	w.Header().Set("Content-Type", "application/json")
	doc, err := a.MetaDB.GetMetaDocByID(id)
	if err != nil {
		respondWithError(err, http.StatusNotFound, w)
		return
	}
	err = json.NewEncoder(w).Encode(doc.Metadata)
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

//PublishEvent posts an event to the event topic
// nolint: errcheck
func (a *App) PublishEvent(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var eventData map[string]string
	err := decoder.Decode(&eventData)
	if err != nil {
		respondWithError(err, http.StatusBadRequest, w)
		return
	}
	defer r.Body.Close()

	//EVENT HEADERCHECK AGAINST VALID
	//TODO: validate event before publishing
	event := c.Event{
		PreviousStages: nil,
		ParentEventID:  a.eventID,
		Data:           eventData,
		Type:           "EVENTTYPE",
	}
	err = a.Publisher.PublishEvent(event)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

//GetBlobInputs gets a storage access key for blob storage
// nolint: errcheck
func (a *App) GetBlobInputs(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		respondWithError(fmt.Errorf("querystring parameter 'url' is required"), http.StatusBadRequest, w)
		return
	}
	authURL, err := a.Blob.GetBlobAuthURL(url)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	w.Write([]byte(authURL))
}

//GetBlobOutputs returns an authenticated URL to an output blob container
// nolint: errcheck
func (a *App) GetBlobOutputs(w http.ResponseWriter, r *http.Request) {
	url, err := a.Blob.CreateBlobContainer(a.eventID)
	if err != nil {
		respondWithError(err, http.StatusInternalServerError, w)
		return
	}
	w.Write([]byte(url))
}

//Close cleans up any external resources
func (a *App) Close() {
	defer a.MetaDB.Close()
	defer a.Publisher.Close()
}

//AddAuth enforces a shared secret between client and server for authentication
func AddAuth(secretHash string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secret := r.Header.Get(secretHeaderKey)
			err := c.CompareHash(secret, secretHash)
			if err != nil {
				respondWithError(err, http.StatusUnauthorized, w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

//AddIdentity adds an identity header to each requests
func AddIdentity(id string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("request-id", id)
			next.ServeHTTP(w, r)
		})
	}
}

//AddLog logs each request (Warning - performance impact)
func AddLog(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debugf("request received: %+v", r)
			next.ServeHTTP(w, r)
		})
	}
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
