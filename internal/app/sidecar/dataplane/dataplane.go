package dataplane

import (
	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane/documentstorage"
	"github.com/lawrencegripper/ion/internal/pkg/common"
)

//DocumentStorageProvider is a document storage DB for storing document data
type DocumentStorageProvider interface {
	GetEventMetaByID(id string) (*documentstorage.EventMeta, error)
	CreateEventMeta(metadata *documentstorage.EventMeta) error
	CreateInsight(insight *documentstorage.Insight) error
	Close()
}

//BlobStorageProvider is responsible for getting information about blobs stored externally
type BlobStorageProvider interface {
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
	BlobStorageProvider
	DocumentStorageProvider
	EventPublisher
}

// Close cleans up the data plane providers
func (d *DataPlane) Close() {
	defer d.BlobStorageProvider.Close()
	defer d.DocumentStorageProvider.Close()
	defer d.EventPublisher.Close()
}
