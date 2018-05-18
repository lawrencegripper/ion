package metadata

import (
	"github.com/lawrencegripper/ion/internal/pkg/common"
)

//Insight is used to export structure data
type Insight struct {
	*common.Context
	ExecutionID string               `bson:"id" json:"id"`
	Data        common.KeyValuePairs `bson:"data" json:"data"`
}

//EventContext is a single entry in a document
type EventContext struct {
	*common.Context
	ParentEventID string               `bson:"parentEventId" json:"parentEventId"`
	Files         []string             `bson:"files" json:"files"`
	Data          common.KeyValuePairs `bson:"data" json:"data"`
}
