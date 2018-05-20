package inmemory //nolint:golint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane/documentstorage"
)

const onDiskName = ".memdb"

//nolint:golint
//InMemoryDB is an in memory DB
type InMemoryDB struct {
	Insights map[string]documentstorage.Insight   `json:"insights"`
	Contexts map[string]documentstorage.EventMeta `json:"contexts"`
}

//NewInMemoryDB creates a new InMemoryDB object
func NewInMemoryDB() (*InMemoryDB, error) {

	if _, err := os.Stat(onDiskName); os.IsNotExist(err) {
		// Create new
		insights := make(map[string]documentstorage.Insight)
		contexts := make(map[string]documentstorage.EventMeta)
		return &InMemoryDB{
			Insights: insights,
			Contexts: contexts,
		}, nil
	}
	// Load from disk
	var db InMemoryDB
	b, err := ioutil.ReadFile(onDiskName)
	if err != nil {
		panic("could not read state from disk")
	}
	err = json.Unmarshal(b, &db)
	if err != nil {
		panic("could not deserialize self from disk")
	}
	return &db, nil
}

//GetEventMetaByID returns a single document matching a given document ID
func (db *InMemoryDB) GetEventMetaByID(id string) (*documentstorage.EventMeta, error) {
	context, exist := db.Contexts[id]
	if !exist {
		return nil, fmt.Errorf("no record found for id '%s'", id)
	}
	return &context, nil
}

//CreateEventMeta creates a new event context document
func (db *InMemoryDB) CreateEventMeta(eventMeta *documentstorage.EventMeta) error {
	db.Contexts[eventMeta.EventID] = *eventMeta
	return nil
}

//CreateInsight creates an insights document
func (db *InMemoryDB) CreateInsight(insight *documentstorage.Insight) error {
	db.Insights[insight.ExecutionID] = *insight
	return nil
}

//Close cleans up external resources
func (db *InMemoryDB) Close() {
	b, err := json.Marshal(db)
	if err != nil {
		panic("Failed to serialize self on close")
	}
	err = ioutil.WriteFile(onDiskName, b, 0777)
	if err != nil {
		panic("Failed to write serialized self to file")
	}
}
