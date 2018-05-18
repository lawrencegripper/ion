package dataplane

import (
	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane/metadata"
	"github.com/lawrencegripper/ion/internal/pkg/common"
)

//MetadataProvider is a document storage DB for storing document data
type MetadataProvider interface {
	GetEventContextByID(id string) (*metadata.EventContext, error)
	CreateEventContext(metadata *metadata.EventContext) error
	CreateInsight(insight *metadata.Insight) error
	Close()
}

//BlobProvider is responsible for getting information about blobs stored externally
type BlobProvider interface {
	GetBlobs(outputDir string, filePaths []string) error
	PutBlobs(filePaths []string) (map[string]string, error)
	Close()
}

//EventPublisher is responsible for publishing events to a remote system
type EventPublisher interface {
	Publish(e common.Event) error
	Close()
}

// DataPlane is the module's API to
// external providers
type DataPlane struct {
	BlobProvider
	MetadataProvider
	EventPublisher
}

// Close cleans up the data plane providers
func (d *DataPlane) Close() {
	defer d.BlobProvider.Close()
	defer d.MetadataProvider.Close()
	defer d.EventPublisher.Close()
}
