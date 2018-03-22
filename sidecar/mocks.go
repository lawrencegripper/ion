package main

import "fmt"

type MockDB struct {
	Document *Document
}

func (db *MockDB) GetByID(id string) (*Document, error) {
	if id != db.Document.ID {
		return nil, fmt.Errorf("the requested document does not exist")
	}
	return db.Document, nil
}

func (db *MockDB) Update(id string, entry Entry) error {
	if id != db.Document.ID {
		return fmt.Errorf("the document you are trying to update does not exists")
	}
	db.Document.Entries = append(db.Document.Entries, entry)
	return nil
}

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

func NewMockDB() *MockDB {
	db := &MockDB{}
	defer db.Close() // Resets mock DB to a known state
	return db
}

type MockPublisher struct {
}

func (p *MockPublisher) PublishEvent(e Event) (error, int) {
	return nil, 200
}

func (p *MockPublisher) Close() {
}

func NewMockPublisher() *MockPublisher {
	pub := &MockPublisher{}
	return pub
}

type MockBlobStorage struct {
	Blobs []BlobInfo
	ID    string
}

func (b *MockBlobStorage) GetBlobsInContainerByID(id string) ([]BlobInfo, error) {
	if id != b.ID {
		return nil, fmt.Errorf("container id could not be found")
	}
	return b.Blobs, nil
}

func (b *MockBlobStorage) Close() {
}

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
