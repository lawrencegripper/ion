package inmemory //nolint:golint

import (
	"fmt"

	"github.com/lawrencegripper/ion/internal/app/sidecar/types"
)

//nolint:golint
//InMemoryDB is an in memory DB
type InMemoryDB struct {
	Insights map[string]types.Insight
	Contexts map[string]types.EventContext
}

//NewInMemoryDB creates a new InMemoryDB object
func NewInMemoryDB() (*InMemoryDB, error) {
	insights := make(map[string]types.Insight)
	contexts := make(map[string]types.EventContext)
	return &InMemoryDB{
		Insights: insights,
		Contexts: contexts,
	}, nil
}

//GetEventContextByID returns a single document matching a given document ID
func (db *InMemoryDB) GetEventContextByID(id string) (*types.EventContext, error) {
	context, exist := db.Contexts[id]
	if !exist {
		return nil, fmt.Errorf("no record found for id '%s'", id)
	}
	return &context, nil
}

//CreateEventContext creates a new event context document
func (db *InMemoryDB) CreateEventContext(eventContext *types.EventContext) error {
	db.Contexts[eventContext.EventID] = *eventContext
	return nil
}

//CreateInsight creates an insights document
func (db *InMemoryDB) CreateInsight(insight *types.Insight) error {
	db.Insights[insight.ExecutionID] = *insight
	return nil
}

//Close cleans up external resources
func (db *InMemoryDB) Close() {
}
