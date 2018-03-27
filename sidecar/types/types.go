package types

import (
	"encoding/json"
	"io"
	"net/http"
)

//MetaProvider is a document storage DB for holding metadata
type MetaProvider interface {
	GetMetaDocByID(docID string) (*MetaDoc, error)
	GetMetaDocAll(correlationID string) ([]MetaDoc, error)
	AddOrUpdateMetaDoc(doc *MetaDoc) error
	Close()
}

//Proxy represents a proxy capable of serving a HTTP request
type Proxy interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

//BlobProxy is responsible for proxying HTTP requests against the Azure storage REST API
type BlobProxy interface {
	Create(resourcePath string, w http.ResponseWriter, r *http.Request)
	Get(resourcePath string, w http.ResponseWriter, r *http.Request)
}

//BlobProvider is responsible for getting information about blobs stored externally
type BlobProvider interface {
	Proxy() BlobProxy
	Create(resourcePath string, blob io.ReadCloser) (string, error)
	Get(resourcePath string) (io.ReadCloser, error)
	List(resourcePath string) ([]string, error)
	Delete(resourcePath string) (bool, error)
	Close()
}

//EventPublisher is responsible for publishing events to a remote system
type EventPublisher interface {
	Publish(e Event) error
	Close()
}

//MetaDoc is a single entry in a document
type MetaDoc struct {
	ID            string            `bson:"id" json:"id"`
	CorrelationID string            `bson:"correlationId" json:"correlationId"`
	ParentEventID string            `bson:"parentId" json:"parentId"`
	Metadata      map[string]string `bson:"metadata" json:"metadata"`
}

//Event is a message for downstream services
type Event struct {
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	PreviousStages []string          `json:"previousStages"`
	ParentEventID  string            `json:"parentId"`
	Data           map[string]string `json:"data"`
}

//ErrorResponse is a struct intended as JSON HTTP response
type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

//Send returns a structured error object
func (e *ErrorResponse) Send(w http.ResponseWriter) {
	w.Header().Set(ContentType, ContentTypeApplicationJSON)
	w.WriteHeader(e.StatusCode)
	_ = json.NewEncoder(w).Encode(e.Message)
}

//StatusCodeResponseWriter is used to expose the HTTP status code for a ResponseWriter
type StatusCodeResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

//NewStatusCodeResponseWriter creates new StatusCodeResponseWriter
func NewStatusCodeResponseWriter(w http.ResponseWriter) *StatusCodeResponseWriter {
	return &StatusCodeResponseWriter{w, http.StatusOK}
}

//WriteHeader hijacks a ResponseWriter.WriteHeader call and stores the status code
func (w *StatusCodeResponseWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}
