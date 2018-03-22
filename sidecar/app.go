package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const secretHeaderKey = "secret"

//App is the sidecar application
type App struct {
	Router     *mux.Router
	MetadataDB MetadataDB
	Publisher  Publisher
	Blob       BlobStorage
	Logger     *log.Logger

	secretHash    string
	blobAccessKey string
}

//Setup initializes application
func (a *App) Setup(
	secret, blobAccessKey string,
	metadataDB MetadataDB,
	publisher Publisher,
	blob BlobStorage,
	logger *log.Logger) {

	if secret == "" || blobAccessKey == "" || metadataDB == nil ||
		publisher == nil || blob == nil || logger == nil {
		panic("nil or empty argument(s) passed to App.Setup()")
	}

	a.secretHash = hash(secret)
	a.blobAccessKey = blobAccessKey

	a.MetadataDB = metadataDB
	a.Publisher = publisher
	a.Blob = blob
	a.Logger = logger

	a.Router = mux.NewRouter()
	a.setupRoutes()
}

//setupRoutes initializes the API routing
func (a *App) setupRoutes() {
	secretAuth := SecretAuth(a.secretHash)
	logRequest := LogRequest(a.Logger)

	getMetadataByIDHandler := http.HandlerFunc(a.GetMetadataByID)
	a.Router.Handle("/meta/{id}", logRequest(secretAuth(getMetadataByIDHandler))).Methods("GET")

	updateMetadataHandler := http.HandlerFunc(a.UpdateMetadata)
	a.Router.Handle("/meta/{id}", logRequest(secretAuth(updateMetadataHandler))).Methods("POST")

	getBlobAccessKeyHandler := http.HandlerFunc(a.GetBlobAccessKey)
	a.Router.Handle("/blob", logRequest(secretAuth(getBlobAccessKeyHandler))).Methods("GET")

	getBlobsInContainerByIDHandler := http.HandlerFunc(a.GetBlobsInContainerByID)
	a.Router.Handle("/blob/container/{id}", logRequest(secretAuth(getBlobsInContainerByIDHandler))).Methods("GET")

	publishEventHandler := http.HandlerFunc(a.PublishEvent)
	a.Router.Handle("/events", logRequest(secretAuth(publishEventHandler))).Methods("POST")
}

//Run and block on web server
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

//GetMetadataByID gets a metadata document by id from a metadata store
// nolint: errcheck
func (a *App) GetMetadataByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	w.Header().Set("Content-Type", "application/json")
	doc, err := a.MetadataDB.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(doc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

//UpdateMetadata updates an existing metadata document
// nolint: errcheck
func (a *App) UpdateMetadata(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	decoder := json.NewDecoder(r.Body)
	var entry Entry
	err := decoder.Decode(&entry)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	//TODO: validate entry before updating
	err = a.MetadataDB.Update(id, entry)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

//PublishEvent posts an event to the event topic
// nolint: errcheck
func (a *App) PublishEvent(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var event Event
	err := decoder.Decode(&event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	//TODO: validate event before publishing
	code, err := a.Publisher.PublishEvent(event)
	if err != nil {
		w.WriteHeader(code)
		return
	}
	w.WriteHeader(code)
}

//GetBlobAccessKey gets a storage access key for blob storage
// nolint: errcheck
func (a *App) GetBlobAccessKey(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(a.blobAccessKey))
}

//GetBlobsInContainerByID returns a list of blobs in a given container/bucket
// nolint: errcheck
func (a *App) GetBlobsInContainerByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	w.Header().Set("Content-Type", "application/json")
	blobs, err := a.Blob.GetBlobsInContainerByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(blobs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

//Close cleans up any external resources
func (a *App) Close() {
	defer a.MetadataDB.Close()
	defer a.Publisher.Close()
}

//SecretAuth enforces a shared secret between client and server for authentication
func SecretAuth(secretHash string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secret := r.Header.Get(secretHeaderKey)
			err := compareHash(secret, secretHash)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

//LogRequest logs the request (Warning - performance impact)
func LogRequest(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Debugf("request received: %+v", r)
			next.ServeHTTP(w, r)
		})
	}
}

//compareHash compares a secret string against a hash
func compareHash(secret, secretHash string) error {
	if secret == "" {
		return fmt.Errorf("secret header was not found")
	}
	if hash(secret) != secretHash {
		return fmt.Errorf("secret did not match")
	}
	return nil
}

//hash returns a MD5 hash of the provided string
// nolint: errcheck
func hash(s string) string {
	hasher := md5.New()
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}
