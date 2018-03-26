package azure

import (
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

//TODO: Cache auth token for reuse

//AzureBlobStorage is responsible for handling the connections to Azure Blob Storage
// nolint: golint
type AzureBlobStorage struct {
	client storage.BlobStorageClient
}

//NewAzureBlobStorage creates a new Azure Blob Storage object
func NewAzureBlobStorage(accountName, accountKey string) (*AzureBlobStorage, error) {
	client, err := storage.NewBasicClient(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("error creating storage client: %+v", err)
	}
	blob := client.GetBlobService()
	asb := &AzureBlobStorage{
		client: blob,
	}
	return asb, nil
}

//Resolve expands a given resource path into a valid Azure Blob Storage URI with SAS
func (a *AzureBlobStorage) Resolve(resourcePath string) (string, error) {
	containerName, blobName, err := parseResourcePath(resourcePath)
	if err != nil {
		return "", err
	}
	container := a.client.GetContainerReference(containerName)
	blob := container.GetBlobReference(blobName)
	url, err := getAuthToken(24, true, false, false, false, false, blob)
	if err != nil {
		return "", err
	}
	return url, nil
}

//Create creates a new resource location if it doesn't exist and returns its URI
func (a *AzureBlobStorage) Create(resourcePath string) (string, error) {
	containerName, blobName, err := parseResourcePath(resourcePath)
	if err != nil {
		return "", err
	}
	container := a.client.GetContainerReference(containerName)
	_, err = container.CreateIfNotExists(&storage.CreateContainerOptions{
		Access: storage.ContainerAccessTypePrivate,
	})
	if err != nil {
		return "", fmt.Errorf("error thrown creating container %s: %+v", containerName, err)
	}
	blob := container.GetBlobReference(blobName)
	url, err := getAuthToken(24, true, true, true, true, true, blob)
	if err != nil {
		return "", err
	}
	return url, nil
}

//Delete expands a resource path into a valid Azure Blob Storage URI then deletes it
func (a *AzureBlobStorage) Delete(resourcePath string) error {
	containerName, blobName, err := parseResourcePath(resourcePath)
	if err != nil {
		return err
	}
	container := a.client.GetContainerReference(containerName)
	blob := container.GetBlobReference(blobName)
	_, err = blob.DeleteIfExists(&storage.DeleteBlobOptions{})
	return err
}

//List lists all blobs under a given container name
func (a *AzureBlobStorage) List(resourcePath string) ([]string, error) {
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
