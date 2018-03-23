package azure

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

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

//GetBlobAuthURL returns a authenticated URL to a blob resource
func (a *AzureBlobStorage) GetBlobAuthURL(url string) (string, error) {
	rx, err := regexp.Compile(`^(?:https?:\/\/)?(?:[^@\/\n]+@)?(?:www\.)?([^:\/\n]+)`)
	if err != nil {
		return "", fmt.Errorf("error thrown compiling regex: %+v", err)
	}
	match := rx.FindString(url)
	if match == "" {
		return "", fmt.Errorf("not domain match in url %s", url)
	}
	path := strings.Replace(url, match, "", 1)
	if path[0] == '/' {
		path = path[1:]
	}
	pathSegments := strings.Split(path, "/")
	if len(pathSegments) <= 1 {
		return "", fmt.Errorf("url does not contain a valid path %s", url)
	}
	containerName := pathSegments[0]
	container := a.client.GetContainerReference(containerName)
	blobName := strings.Join(pathSegments[1:], "/")
	blob := container.GetBlobReference(blobName)
	sasURL, err := blob.GetSASURI(storage.BlobSASOptions{
		BlobServiceSASPermissions: storage.BlobServiceSASPermissions{
			Read: true,
		},
		SASOptions: storage.SASOptions{
			Start:    time.Now().Add(time.Hour * time.Duration(-6)),
			Expiry:   time.Now().Add(time.Hour * time.Duration(24)),
			UseHTTPS: true,
		},
	})
	if err != nil {
		return "", fmt.Errorf("error thrown getting SAS uri for blob %s in container %s: %+v", blobName, containerName, err)
	}
	return sasURL, nil
}

//CreateBlobContainer creates a new container if it doesn't exist and return an authenticated URL
func (a *AzureBlobStorage) CreateBlobContainer(locationName string) (string, error) {
	container := a.client.GetContainerReference(locationName)
	_, err := container.CreateIfNotExists(&storage.CreateContainerOptions{
		Access: storage.ContainerAccessTypePrivate,
	})
	if err != nil {
		return "", fmt.Errorf("error thrown creating container %s: %+v", locationName, err)
	}
	sasURL, err := container.GetSASURI(storage.ContainerSASOptions{
		ContainerSASPermissions: storage.ContainerSASPermissions{
			List: true,
			BlobServiceSASPermissions: storage.BlobServiceSASPermissions{
				Read:  true,
				Add:   true,
				Write: true,
			},
		},
		SASOptions: storage.SASOptions{
			Start:    time.Now().Add(time.Hour * time.Duration(-6)),
			Expiry:   time.Now().Add(time.Hour * time.Duration(24)),
			UseHTTPS: true,
		},
	})
	if err != nil {
		return "", fmt.Errorf("error thrown getting SAS uri for container %s: %+v", locationName, err)
	}
	return sasURL, nil
}
