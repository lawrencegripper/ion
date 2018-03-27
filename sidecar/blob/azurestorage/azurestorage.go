package azurestorage

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

//TODO: Cache auth token for reuse

//BlobStorage is responsible for handling the connections to Azure Blob Storage
// nolint: golint
type BlobStorage struct {
	client storage.BlobStorageClient
}

//NewBlobStorage creates a new Azure Blob Storage object
func NewBlobStorage(accountName, accountKey string) (*BlobStorage, error) {
	client, err := storage.NewBasicClient(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("error creating storage client: %+v", err)
	}
	blob := client.GetBlobService()
	asb := &BlobStorage{
		client: blob,
	}
	return asb, nil
}

//ResolveGet constructs a valid HTTP request for proxying to get a blob resource
func (a *BlobStorage) ResolveGet(resourcePath string, r *http.Request) (*http.Request, error) {
	containerName, blobName, err := parseResourcePath(resourcePath)
	if err != nil {
		return nil, err
	}
	container := a.client.GetContainerReference(containerName)
	blob := container.GetBlobReference(blobName)
	sasURL, err := getAuthToken(24, true, false, false, false, false, blob)
	if err != nil {
		return nil, err
	}
	r.URL, err = url.Parse(sasURL)
	if err != nil {
		return nil, err
	}
	return r, nil
}

//ResolveCreate creates a new blob resource location if needed and then constructs
//a valid  HTTP request for proxying to put a blob resource at this location
func (a *BlobStorage) ResolveCreate(resourcePath string, r *http.Request) (*http.Request, error) {
	containerName, blobName, err := parseResourcePath(resourcePath)
	if err != nil {
		return nil, err
	}
	container := a.client.GetContainerReference(containerName)
	_, err = container.CreateIfNotExists(&storage.CreateContainerOptions{
		Access: storage.ContainerAccessTypePrivate,
	})
	if err != nil {
		return nil, fmt.Errorf("error thrown creating container %s: %+v", containerName, err)
	}
	blob := container.GetBlobReference(blobName)
	sasURL, err := getAuthToken(24, true, true, true, true, true, blob)
	if err != nil {
		return nil, err
	}
	r.URL, err = url.Parse(sasURL)
	if err != nil {
		return nil, err
	}
	r.Header.Set("x-ms-blob-type", "BlockBlob")
	return r, nil
}

//Delete expands a blob resource path into a valid Azure Blob Storage URI then deletes it
func (a *BlobStorage) Delete(resourcePath string) (bool, error) {
	containerName, blobName, err := parseResourcePath(resourcePath)
	if err != nil {
		return false, err
	}
	container := a.client.GetContainerReference(containerName)
	blob := container.GetBlobReference(blobName)
	deleted, err := blob.DeleteIfExists(&storage.DeleteBlobOptions{})
	return deleted, err
}

//List expands a blob resource path into a container name and then lists all blobs inside it
func (a *BlobStorage) List(resourcePath string) ([]string, error) {
	container := a.client.GetContainerReference(resourcePath)
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
