package common

//MetaDB is a document storage DB for holding metadata
type MetaDB interface {
	GetMetaDocByID(docID string) (*MetaDoc, error)
	GetMetaDocAll(correlationID string) ([]MetaDoc, error)
	AddOrUpdateMetaDoc(doc *MetaDoc) error
	Close()
}

//BlobProvider is responsible for getting information about blobs stored externally
type BlobProvider interface {
	Resolve(resourcePath string) (string, error)
	Create(resourcePath string) (string, error)
	List(resourcePath string) ([]string, error)
	Delete(resourcePath string) error
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
