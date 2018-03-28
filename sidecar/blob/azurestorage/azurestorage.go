package azurestorage

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/lawrencegripper/mlops/sidecar/types"
	"github.com/vulcand/oxy/forward"
)

//TODO: Cache auth token for reuse

//Config to setup a BlobStorage blob provider
type Config struct {
	BlobAccountName string `description:"Azure Blob Storage account name"`
	BlobAccountKey  string `description:"Azure Blob Storage account key"`
	UseProxy        bool   `description:"Enable proxy"`
}

//BlobStorage is responsible for handling the connections to Azure Blob Storage
// nolint: golint
type BlobStorage struct {
	blobClient storage.BlobStorageClient
	proxy      types.BlobProxy
}

//NewBlobStorage creates a new Azure Blob Storage object
func NewBlobStorage(config *Config) (*BlobStorage, error) {
	blobClient, err := storage.NewBasicClient(config.BlobAccountName, config.BlobAccountKey)
	if err != nil {
		return nil, fmt.Errorf("error creating storage blobClient: %+v", err)
	}
	blob := blobClient.GetBlobService()
	asb := &BlobStorage{
		blobClient: blob,
	}
	if config.UseProxy {
		proxy, _ := forward.New(
			forward.Stream(true),
		)
		asb.proxy = NewAzureBlobProxy(proxy, asb)
	}
	return asb, nil
}

//Proxy is used to inform clients that this provider has a proxy
func (a *BlobStorage) Proxy() types.BlobProxy {
	return a.proxy
}

//Create creates a new blob and returns a url pointing to the new blob
func (a *BlobStorage) Create(resourcePath string, blobData io.ReadCloser) (string, error) {
	blob, err := createContainerForBlobIfNotExist(&a.blobClient, resourcePath)
	if err != nil {
		return "", err
	}
	err = blob.CreateBlockBlobFromReader(blobData, &storage.PutBlobOptions{})
	if err != nil {
		return "", err
	}
	url := blob.GetURL()
	segs := strings.Split(url, "?") // remove SAS token if present
	if len(segs) < 2 {
		return url, nil
	}
	return segs[0], nil
}

//Get gets a blob resource
func (a *BlobStorage) Get(resourcePath string) (io.ReadCloser, error) {
	blob, err := getBlobFromResourcePath(&a.blobClient, resourcePath)
	if err != nil {
		return nil, err
	}
	reader, err := blob.Get(&storage.GetBlobOptions{})
	if err != nil {
		return nil, err
	}
	return reader, nil
}

//Delete expands a blob resource path into a valid Azure Blob Storage URI then deletes it
func (a *BlobStorage) Delete(resourcePath string) (bool, error) {
	blob, err := getBlobFromResourcePath(&a.blobClient, resourcePath)
	if err != nil {
		return false, err
	}
	deleted, err := blob.DeleteIfExists(&storage.DeleteBlobOptions{})
	return deleted, err
}

//List expands a blob resource path into a container name and then lists all blobs inside it
func (a *BlobStorage) List(resourcePath string) ([]string, error) {
	container := a.blobClient.GetContainerReference(resourcePath)
	blobs, err := container.ListBlobs(storage.ListBlobsParameters{})
	if err != nil {
		return nil, err
	}
	var blobList []string
	for _, blob := range blobs.Blobs {
		blobList = append(blobList, blob.Name)
	}
	return blobList, nil
}

//Close cleans up any external resources
func (a *BlobStorage) Close() {
}

//parseResourcePath extracts a container name and blob name from a combined resource string
func parseResourcePath(resourcePath string) (string, string, error) {
	if resourcePath[0] == '/' {
		resourcePath = resourcePath[1:]
	}
	pathSegments := strings.Split(resourcePath, "/")
	if len(pathSegments) <= 1 {
		return "", "", fmt.Errorf("resource path '%s' is not a valid path", resourcePath)
	}
	containerName := pathSegments[0]
	blobName := strings.Join(pathSegments[1:], "/")
	return containerName, blobName, nil
}

//getAuthToken returns a SAS authenticated URI for a given blob resource
func getAuthToken(durationInHours int, canRead, canAdd, canCreate, canWrite, canDelete bool,
	blob *storage.Blob) (string, error) {
	sasURL, err := blob.GetSASURI(storage.BlobSASOptions{
		BlobServiceSASPermissions: storage.BlobServiceSASPermissions{
			Read:   canRead,
			Add:    canAdd,
			Create: canCreate,
			Delete: canDelete,
		},
		SASOptions: storage.SASOptions{
			Start:    time.Now().Add(time.Hour * time.Duration(-6)),
			Expiry:   time.Now().Add(time.Hour * time.Duration(durationInHours)),
			UseHTTPS: false,
		},
	})
	if err != nil {
		return "", fmt.Errorf("error thrown getting SAS uri for blob %s in container %s: %+v", blob.Name, blob.Container.Name, err)
	}
	return sasURL, nil
}

func createContainerForBlobIfNotExist(client *storage.BlobStorageClient, resourcePath string) (*storage.Blob, error) {
	containerName, blobName, err := parseResourcePath(resourcePath)
	if err != nil {
		return nil, err
	}
	container := client.GetContainerReference(containerName)
	_, err = container.CreateIfNotExists(&storage.CreateContainerOptions{
		Access: storage.ContainerAccessTypePrivate,
	})
	if err != nil {
		return nil, fmt.Errorf("error thrown creating container %s: %+v", containerName, err)
	}
	blob := container.GetBlobReference(blobName)
	return blob, nil
}

func getBlobFromResourcePath(client *storage.BlobStorageClient, resourcePath string) (*storage.Blob, error) {
	containerName, blobName, err := parseResourcePath(resourcePath)
	if err != nil {
		return nil, err
	}
	container := client.GetContainerReference(containerName)
	blob := container.GetBlobReference(blobName)
	return blob, nil
}
