package types

import "net/http"

//MetaProvider is a document storage DB for holding metadata
type MetaProvider interface {
	GetMetaDocByID(docID string) (*MetaDoc, error)
	GetMetaDocAll(correlationID string) ([]MetaDoc, error)
	AddOrUpdateMetaDoc(doc *MetaDoc) error
	Close()
}

//Resolver is responsible for resolving a HTTP request to be proxied
type Resolver func(resID string, r *http.Request) (*http.Request, error)

//BlobProxy is responsible for proxying a HTTP request
type BlobProxy interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

//BlobProvider is responsible for getting information about blobs stored externally
type BlobProvider interface {
	ResolveGet(resourcePath string, r *http.Request) (*http.Request, error)
	ResolveCreate(resourcePath string, r *http.Request) (*http.Request, error)
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
