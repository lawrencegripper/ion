package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

//App is the sidecar application
type App struct {
	Router     *mux.Router
	MetadataDB MetadataDB
	Publisher  Publisher

	secretHash    string
	blobAccessKey string
}

//MetadataDB is a document storage DB for holding metadata
type MetadataDB interface {
	GetByID(id string) (*Document, error)
	Update(id string, entry Entry) error
	Close()
}

//Publisher is responsible for publishing events to a remote system
type Publisher interface {
	PublishEvent(e Event) (error, int)
	Close()
}

//Event is a message for downstream services
type Event struct {
	ID        string `json:"id"`
	Desc      string `json:"desc"`
	CreatedAt string `json:"createdAt"`
}

//Document is a collection of entries associated with a workflow
type Document struct {
	ID      string  `bson:"id" json:"id"`
	Entries []Entry `bson:"entries" json:"entries"`
}

//Entry is a single entry in a document
type Entry struct {
	ID       string            `bson:"entryID" json:"entryID"`
	Metadata map[string]string `bson:"metadata" json:"metadata"`
}

//Setup initializes application
func (a *App) Setup(secret, blobAccessKey string, metadataDB MetadataDB, publisher Publisher) {
	a.secretHash = hash(secret)
	a.blobAccessKey = blobAccessKey

	a.MetadataDB = metadataDB
	a.Publisher = publisher

	a.Router = mux.NewRouter()
	a.setupRoutes()
}

//setupRoutes initializes the API routing
func (a *App) setupRoutes() {
	secretAuth := SecretAuth(a.secretHash)

	getMetadataByIDHandler := http.HandlerFunc(a.GetMetadataByID)
	a.Router.Handle("/meta/{id}", secretAuth(getMetadataByIDHandler)).Methods("GET")

	updateMetadataHandler := http.HandlerFunc(a.UpdateMetadata)
	a.Router.Handle("/meta/{id}", secretAuth(updateMetadataHandler)).Methods("POST")

	getBlobAccessKeyHandler := http.HandlerFunc(a.GetBlobAccessKey)
	a.Router.Handle("/blob", secretAuth(getBlobAccessKeyHandler)).Methods("GET")

	publishEventHandler := http.HandlerFunc(a.PublishEvent)
	a.Router.Handle("/events", secretAuth(publishEventHandler)).Methods("POST")
}

//Run and block on web server
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

//GetMetadataByID gets a metadata document by id from a metadata store
func (a *App) GetMetadataByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	doc, err := a.MetadataDB.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	b, err := json.Marshal(doc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

//UpdateMetadata updates an existing metadata document
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
	err, code := a.Publisher.PublishEvent(event)
	if err != nil {
		w.WriteHeader(code)
		return
	}
	w.WriteHeader(code)
}

//GetBlobAccessKey gets a storage access key for blob storage
func (a *App) GetBlobAccessKey(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(a.blobAccessKey))
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
			secret := r.Header.Get("Secret")
			err := compareHash(secret, secretHash)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
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
func hash(s string) string {
	hasher := md5.New()
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}
