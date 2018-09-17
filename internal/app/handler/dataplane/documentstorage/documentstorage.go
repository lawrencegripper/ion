package documentstorage

import (
	"github.com/lawrencegripper/ion/internal/pkg/common"
)

const (
	// NotFoundErr returned when a document is not found in document storage
	NotFoundErr = "document not found"
)

//Insight is used to export structure data
type Insight struct {
	*common.Context
	ExecutionID string               `bson:"id" json:"id"`
	Data        common.KeyValuePairs `bson:"data" json:"data"`
}

//EventMeta is a single entry in a document
type EventMeta struct {
	*common.Context
	Files []string             `bson:"files" json:"files"`
	Data  common.KeyValuePairs `bson:"data" json:"data"`
}

//ModuleLogs is a single entry in a document
type ModuleLogs struct {
	*common.Context
	Description string `bson:"desc" json:"desc"`
	Logs        string `bson:"logs" json:"logs"`
	Succeeded   bool   `bson:"succeeded" json:"succeeded"`
}
