package main

import "fmt"

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
		ID: "test",
		Blobs: []BlobInfo{
			{
				Name: "blob1",
				URI:  "https://blob.com/container1/blob1?sas=1234",
			},
			{
				Name: "blob2",
				URI:  "https://blob.com/container2/blob2?sas=1234",
			},
			{
				Name: "blob3",
				URI:  "https://blob.com/container3/blob3?sas=1234",
			},
		},
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
	Document *Document
}

//GetByID mock function to get a metadata document by id
func (db *MockDB) GetByID(id string) (*Document, error) {
	if id != db.Document.ID {
		return nil, fmt.Errorf("the requested document does not exist")
	}
	return db.Document, nil
}

//Update mock function to update a metadata document
func (db *MockDB) Update(id string, entry Entry) error {
	if id != db.Document.ID {
		return fmt.Errorf("the document you are trying to update does not exists")
	}
	db.Document.Entries = append(db.Document.Entries, entry)
	return nil
}

//Close mock function - used to reset the DB to a known state
func (db *MockDB) Close() {
	db.Document = &Document{
		ID: "0",
		Entries: []Entry{
			{
				ID: "0",
				Metadata: map[string]string{
					"TEST123": "123TEST",
				},
			},
		},
	}
}

//MockPublisher is a mock event publisher
type MockPublisher struct {
}

//PublishEvent is a mock function for publishing events
func (p *MockPublisher) PublishEvent(e Event) (int, error) {
	return 200, nil
}

//Close is a mock clean up function
func (p *MockPublisher) Close() {
}

//MockBlobStorage is a mock blob storage
type MockBlobStorage struct {
	Blobs []BlobInfo
	ID    string
}

//GetBlobsInContainerByID mock function used to get a list of blobs
func (b *MockBlobStorage) GetBlobsInContainerByID(id string) ([]BlobInfo, error) {
	if id != b.ID {
		return nil, fmt.Errorf("container id could not be found")
	}
	return b.Blobs, nil
}

//Close mock clean up function
func (b *MockBlobStorage) Close() {
}
