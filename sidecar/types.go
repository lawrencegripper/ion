package main

//MetadataDB is a document storage DB for holding metadata
type MetadataDB interface {
	GetByID(id string) (*Document, error)
	Update(id string, entry Entry) error
	Close()
}

//Publisher is responsible for publishing events to a remote system
type Publisher interface {
	PublishEvent(e Event) (int, error)
	Close()
}

//BlobStorage is responsible for getting information about blobs stored externally
type BlobStorage interface {
	GetBlobsInContainerByID(id string) ([]BlobInfo, error)
}

//BlobInfo is a simplified view of a Blob object
type BlobInfo struct {
	URI  string
	Name string
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
