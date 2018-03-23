package tests

import (
	"fmt"

	"github.com/lawrencegripper/mlops/sidecar/common"
)

//NewMockDB is a mock metadata database
// nolint: deadcode
func NewMockDB() *MockDB {
	db := &MockDB{}
	defer db.Close() // Resets mock DB to a known state
	return db
}

//NewMockBlobStorage is a mock blob store
// nolint: deadcode
func NewMockBlobStorage() *MockBlobStorage {
	blob := &MockBlobStorage{
		SAS:     "se=2018-03-10T19%3A33%3A48Z&sig=nm%2B53E%2FgqklbjmkcvG2bTKGaIOJSGNDS%3D&sp=rl&spr=https&sr=c&st=2018-03-10T19%3A33%3A48Z&sv=2016-05-31",
		baseURL: "https://blob.com",
	}
	return blob
}

//NewMockPublisher is a mock event publisher
// nolint: deadcode
func NewMockPublisher() *MockPublisher {
	pub := &MockPublisher{}
	return pub
}

//MockDB is a mock metadata database
type MockDB struct {
	MetaDocs []common.MetaDoc
}

//GetMetaDocByID mock function to get a metadata document by id
func (db *MockDB) GetMetaDocByID(docID string) (*common.MetaDoc, error) {
	for _, doc := range db.MetaDocs {
		if doc.ID == docID {
			return &doc, nil
		}
	}
	return nil, fmt.Errorf("no document found with matching id %s", docID)
}

//GetMetaDocAll mock function to get a metadata document by id
func (db *MockDB) GetMetaDocAll(correlationID string) ([]common.MetaDoc, error) {
	docs := []common.MetaDoc{}
	for _, doc := range db.MetaDocs {
		if doc.CorrelationID == correlationID {
			docs = append(docs, doc)
		}
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("no matching documents found for id %s", correlationID)
	}
	return docs, nil
}

//AddOrUpdateMetaDoc mock function to update a metadata document
func (db *MockDB) AddOrUpdateMetaDoc(doc *common.MetaDoc) error {
	for _, d := range db.MetaDocs {
		if d.ID == doc.ID {
			d = *doc
			return nil
		}
	}
	db.MetaDocs = append(db.MetaDocs, *doc)
	return nil
}

//Close mock function - used to reset the DB to a known state
func (db *MockDB) Close() {
	db.MetaDocs = []common.MetaDoc{
		{
			ID:            "0",
			CorrelationID: "0",
			ParentEventID: "0",
			Metadata: map[string]string{
				"TEST123": "123TEST",
			},
		},
		{
			ID:            "1",
			CorrelationID: "0",
			ParentEventID: "0",
			Metadata: map[string]string{
				"ALICE": "BOB",
			},
		},
		{
			ID:            "2",
			CorrelationID: "0",
			ParentEventID: "1",
			Metadata: map[string]string{
				"BLUE": "GREEN",
				"JACK": "JILL",
			},
		},
	}
}

//MockPublisher is a mock event publisher
type MockPublisher struct {
}

//PublishEvent is a mock function for publishing events
func (p *MockPublisher) PublishEvent(e common.Event) error {
	return nil
}

//Close is a mock clean up function
func (p *MockPublisher) Close() {
}

//MockBlobStorage is a mock blob storage
type MockBlobStorage struct {
	SAS     string
	baseURL string
}

//GetBlobAuthURL mock function used to get an authenticated URL to a blob resource
func (b *MockBlobStorage) GetBlobAuthURL(url string) (string, error) {
	return fmt.Sprintf("%s?%s", url, b.SAS), nil
}

//CreateBlobContainer mock function used to create a new addressable blob resource location
func (b *MockBlobStorage) CreateBlobContainer(id string) (string, error) {
	return fmt.Sprintf("%s/%s?%s", b.baseURL, id, b.SAS), nil
}

//Close mock clean up function
func (b *MockBlobStorage) Close() {
}
