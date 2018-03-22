package main

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

//AzureBlobStorage is responsible for handling the connections to Azure Blob Storage
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

//GetBlobsInContainerByID returns a list of blobs in an given container
func (a *AzureBlobStorage) GetBlobsInContainerByID(id string) ([]BlobInfo, error) {
	container := a.client.GetContainerReference(id)
	if container == nil {
		return nil, fmt.Errorf("no container found for id %s", id)
	}
	//TODO: handle pagination
	blobList, err := container.ListBlobs(storage.ListBlobsParameters{})
	if err != nil {
		return nil, fmt.Errorf("error listing blobs in container %s: %+v", id, err)
	}
	blobInfoList := make([]BlobInfo, 0)
	for _, blob := range blobList.Blobs {
		name := blob.Name
		uri, err := blob.GetSASURI(storage.BlobSASOptions{
			BlobServiceSASPermissions: storage.BlobServiceSASPermissions{
				Read: true,
			},
			SASOptions: storage.SASOptions{
				Start:    time.Now(),
				Expiry:   time.Now().Add(time.Hour * time.Duration(24)),
				UseHTTPS: true,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error getting blob '%s' SAS uri in container %s: %+v", name, id, err)
		}
		b := BlobInfo{
			URI:  uri,
			Name: name,
		}
		blobInfoList = append(blobInfoList, b)
	}
	return blobInfoList, nil
}
