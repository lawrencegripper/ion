package azure

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/lawrencegripper/ion/internal/app/handler/helpers"
)

// cSpell:ignore nolint, golint, sasuris, sasuri

//Config to setup a BlobStorage blob provider
type Config struct {
	Enabled         bool   `description:"Enable Azure Blob storage provider"`
	BlobAccountName string `description:"Azure Blob Storage account name"`
	BlobAccountKey  string `description:"Azure Blob Storage account key"`
	ContainerName   string `description:"Azure Blob Storage container name"`
}

//BlobStorage is responsible for handling the connections to Azure Blob Storage
// nolint: golint
type BlobStorage struct {
	blobClient       storage.BlobStorageClient
	containerName    string
	outputBlobPrefix string
	inputBlobPrefix  string
}

//NewBlobStorage creates a new Azure Blob Storage object
func NewBlobStorage(config *Config, inputBlobPrefix, outputBlobPrefix string) (*BlobStorage, error) {
	blobClient, err := storage.NewBasicClient(config.BlobAccountName, config.BlobAccountKey)
	if err != nil {
		return nil, fmt.Errorf("error creating storage blobClient: %+v", err)
	}
	blob := blobClient.GetBlobService()
	asb := &BlobStorage{
		blobClient:       blob,
		containerName:    config.ContainerName,
		outputBlobPrefix: outputBlobPrefix,
		inputBlobPrefix:  inputBlobPrefix,
	}
	return asb, nil
}

//PutBlobs puts a file into Azure Blob Storage
func (a *BlobStorage) PutBlobs(filePaths []string) (map[string]string, error) {
	container, err := a.createContainerIfNotExist()
	if err != nil {
		return nil, err
	}

	readStorageOptions := storage.BlobSASOptions{
		BlobServiceSASPermissions: storage.BlobServiceSASPermissions{
			Read: true,
		},
		SASOptions: storage.SASOptions{
			Start:  time.Now().Add(time.Duration(-1) * time.Hour),
			Expiry: time.Now().Add(time.Duration(24) * time.Hour),
		},
	}

	blobSASURIs := make(map[string]string)

	for _, filePath := range filePaths {
		_, nakedFilePath := path.Split(filePath)
		blobPath := helpers.JoinBlobPath(a.outputBlobPrefix, nakedFilePath)
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read data from file '%s', error: '%+v'", filePath, err)
		}
		defer file.Close() // nolint: errcheck
		blobRef := container.GetBlobReference(blobPath)
		_, err = blobRef.DeleteIfExists(&storage.DeleteBlobOptions{})
		if err != nil {
			return nil, err
		}
		err = blobRef.CreateBlockBlobFromReader(file, &storage.PutBlobOptions{})
		if err != nil {
			return nil, err
		}

		uri, err := blobRef.GetSASURI(readStorageOptions)
		if err != nil {
			return nil, err
		}

		blobSASURIs[nakedFilePath] = uri
	}
	return blobSASURIs, nil
}

//GetBlobs gets each of the provided blobs from Azure Blob Storage
func (a *BlobStorage) GetBlobs(outputDir string, filePaths []string) error {
	containerName := a.containerName
	container := a.blobClient.GetContainerReference(containerName)
	for _, filePath := range filePaths {
		blobPath := helpers.JoinBlobPath(a.inputBlobPrefix, filePath)
		blobRef := container.GetBlobReference(blobPath)
		blob, err := blobRef.Get(&storage.GetBlobOptions{})
		if err != nil {
			return fmt.Errorf("failed to get blob '%s' with error '%+v'", blobPath, err)
		}
		var bytes []byte
		_, err = blob.Read(bytes)
		if err != nil {
			return fmt.Errorf("failed to read blob '%s' with error '%+v'", blobPath, err)
		}
		defer blob.Close() // nolint: errcheck
		outputFilePath := path.Join(outputDir, filePath)
		err = ioutil.WriteFile(outputFilePath, bytes, 0777)
		if err != nil {
			return fmt.Errorf("failed to write file '%s' with error '%+v'", outputFilePath, err)
		}
	}
	return nil
}

//Close cleans up any external resources
func (a *BlobStorage) Close() {
}

//createContainerIfNotExist creates the container if it doesn't exist
func (a *BlobStorage) createContainerIfNotExist() (*storage.Container, error) {
	containerName := a.containerName
	container := a.blobClient.GetContainerReference(containerName)
	_, err := container.CreateIfNotExists(&storage.CreateContainerOptions{
		Access: storage.ContainerAccessTypePrivate,
	})
	if err != nil {
		return nil, fmt.Errorf("error thrown creating container %s: %+v", containerName, err)
	}
	return container, nil
}
