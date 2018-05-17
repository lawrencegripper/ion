package inmemory //nolint:golint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane/metadata"
)

const onDiskName = ".memdb"

//nolint:golint
//InMemoryDB is an in memory DB
type InMemoryDB struct {
	Insights map[string]metadata.Insight      `json:"insights"`
	Contexts map[string]metadata.EventContext `json:"contexts"`
}

//NewInMemoryDB creates a new InMemoryDB object
func NewInMemoryDB() (*InMemoryDB, error) {

	if _, err := os.Stat(onDiskName); os.IsNotExist(err) {
		// Create new
		insights := make(map[string]metadata.Insight)
		contexts := make(map[string]metadata.EventContext)
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

//GetEventContextByID returns a single document matching a given document ID
func (db *InMemoryDB) GetEventContextByID(id string) (*metadata.EventContext, error) {
	context, exist := db.Contexts[id]
	if !exist {
		return nil, fmt.Errorf("no record found for id '%s'", id)
	}
	return &context, nil
}

//CreateEventContext creates a new event context document
func (db *InMemoryDB) CreateEventContext(eventContext *metadata.EventContext) error {
	db.Contexts[eventContext.EventID] = *eventContext
	return nil
}

//CreateInsight creates an insights document
func (db *InMemoryDB) CreateInsight(insight *metadata.Insight) error {
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
