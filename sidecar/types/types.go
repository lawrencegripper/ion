package types

import (
	"encoding/json"
	"io"
	"net/http"
)

//MetadataProvider is a document storage DB for storing document data
type MetadataProvider interface {
	GetMetadataDocumentByID(id string) (*Metadata, error)
	GetMetadataDocumentsByID(id string) ([]Metadata, error)
	UpsertMetadataDocument(metadata *Metadata) error
	Close()
}

//BlobProvider is responsible for getting information about blobs stored externally
type BlobProvider interface {
	GetBlobs(outputDir string, filePaths []string) error
	CreateBlobs(filePaths []string) error
	Close()
}

//EventPublisher is responsible for publishing events to a remote system
type EventPublisher interface {
	Publish(e Event) error
	Close()
}

type Blob struct {
	Name string
	Data io.ReadCloser
}

//Metadata is a single entry in a document
type Metadata struct {
	ExecutionID   string            `bson:"id" json:"id"`
	CorrelationID string            `bson:"correlationId" json:"correlationId"`
	ParentEventID string            `bson:"parentEventId" json:"parentEventId"`
	Data          map[string]string `bson:"data" json:"data"`
}

//Event the basic event data format
type Event struct {
	Type           string            `json:"type"`
	PreviousStages []string          `json:"previousStages"`
	ExecutionID    string            `json:"contextId"`
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
