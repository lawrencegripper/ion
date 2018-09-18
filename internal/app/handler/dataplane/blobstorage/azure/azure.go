package azure

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/azure/azure-storage-blob-go/2016-05-31/azblob"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage"
	"github.com/lawrencegripper/ion/internal/app/handler/helpers"
	"github.com/lawrencegripper/ion/internal/app/handler/module"
	log "github.com/sirupsen/logrus"
)

// cSpell:ignore nolint, golint, sasuris, sasuri

const (
	// ContainerAlreadyExistsErr returned when creating a container that already exists
	ContainerAlreadyExistsErr = "ContainerAlreadyExists"
)

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
	containerName    string
	outputBlobPrefix string
	inputBlobPrefix  string
	eventMeta        *documentstorage.EventMeta
	accountKey       string
	accountName      string
	env              *module.Environment
}

//NewBlobStorage creates a new Azure Blob Storage object
func NewBlobStorage(config *Config, inputBlobPrefix, outputBlobPrefix string, eventMeta *documentstorage.EventMeta, env *module.Environment) (*BlobStorage, error) {
	asb := &BlobStorage{
		containerName:    config.ContainerName,
		outputBlobPrefix: outputBlobPrefix,
		inputBlobPrefix:  inputBlobPrefix,
		eventMeta:        eventMeta,
		accountName:      config.BlobAccountName,
		accountKey:       config.BlobAccountKey,
		env:              env,
	}
	return asb, nil
}

//PutBlobs puts a file into Azure Blob Storage
func (a *BlobStorage) PutBlobs(filePaths []string) (map[string]string, error) {
	blobSASURIs := make(map[string]string)

	c := azblob.NewSharedKeyCredential(a.accountName, a.accountKey)
	p := azblob.NewPipeline(c, azblob.PipelineOptions{
		Retry: azblob.RetryOptions{
			Policy:   azblob.RetryPolicyExponential,
			MaxTries: 3,
		},
	})
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", a.accountName, a.containerName))
	containerURL := azblob.NewContainerURL(*URL, p)
	ctx := context.Background()
	_, err := containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	if err != nil {
		if serr, ok := err.(azblob.StorageError); !ok {
			return nil, err
		} else { // nolint: golint
			if serr.ServiceCode() != ContainerAlreadyExistsErr {
				return nil, err
			}
		}
	}

	for _, filePath := range filePaths {
		filePathOutOfEnv := strings.Replace(filePath, a.env.OutputBlobDirPath, "", 1)
		if filePathOutOfEnv[0] == '/' {
			filePathOutOfEnv = filePathOutOfEnv[1:]
		}
		filePathOutOfEnv = filepath.Clean(filePathOutOfEnv)
		blobPath := helpers.JoinBlobPath(a.outputBlobPrefix, filePathOutOfEnv)
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read data from file '%s', error: '%+v'", filePath, err)
		}
		defer file.Close() // nolint: errcheck

		stat, err := file.Stat()
		if err != nil {
			return nil, err
		}

		b := stat.Size()
		kb := float64(b) / 1024
		mb := float64(kb / 1024)
		var timeout time.Duration
		if mb > 5 {
			timeout = time.Duration(mb) * 60 * time.Second
		}
		p = azblob.NewPipeline(c, azblob.PipelineOptions{
			Retry: azblob.RetryOptions{
				Policy:     azblob.RetryPolicyExponential,
				MaxTries:   3,
				TryTimeout: timeout,
			},
		})
		blobURL := containerURL.WithPipeline(p).NewBlockBlobURL(blobPath)
		parallelism := uint16(runtime.NumCPU())
		_, err = azblob.UploadFileToBlockBlob(ctx, file, blobURL, azblob.UploadToBlockBlobOptions{
			BlockSize:   1 * 1024 * 1024,
			Parallelism: parallelism})
		if err != nil {
			return nil, err
		}

		sasQueryParams := azblob.BlobSASSignatureValues{
			Protocol:      azblob.SASProtocolHTTPS,
			StartTime:     time.Now().UTC().Add(-1 * time.Hour),
			ExpiryTime:    time.Now().UTC().Add(24 * time.Hour),
			Permissions:   azblob.BlobSASPermissions{Read: true}.String(),
			ContainerName: a.containerName,
			BlobName:      blobPath,
		}.NewSASQueryParameters(c)

		queryParams := sasQueryParams.Encode()
		_, filename := filepath.Split(filePathOutOfEnv)
		blobSASURIs[filename] = fmt.Sprintf("%s?%s", blobURL, queryParams)
	}
	return blobSASURIs, nil
}

//GetBlobs gets each of the provided blobs from Azure Blob Storage
func (a *BlobStorage) GetBlobs(outputDir string, filePaths []string) error {
	if a.eventMeta == nil {
		log.Info("skipping getblob as eventmeta is nil meaning this is an orphaned event or the first in a workflow")
		return nil
	}
	dataAsMap := a.eventMeta.Data.AsMap()
	for _, filePath := range filePaths {
		_, filename := filepath.Split(filePath)
		fileSASURL, ok := dataAsMap[filename]
		if !ok {
			log.WithField("filepath", filePath).WithField("eventMeta", a.eventMeta).Error("couldn't find SAS url for azure blob data")
			return fmt.Errorf("failed to find sas url for azure blob data in event meta: %+v", a.eventMeta)
		}

		resp, err := http.Get(fileSASURL)
		if err != nil {
			log.WithField("filepath", filePath).WithField("eventMeta", a.eventMeta).Error("couldn't download data from SAS url for azure blob data")
			return fmt.Errorf("Couldn't download data from SAS url for azure blob url: %+v data: %+v", fileSASURL, a.eventMeta)
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close() //nolint: errcheck

		if err != nil {
			return fmt.Errorf("failed to read blob '%s' with error '%+v'", fileSASURL, err)
		}
		dirPath := filepath.Dir(filePath)
		dirPathInEnv := path.Join(outputDir, dirPath)
		_ = os.MkdirAll(dirPathInEnv, os.ModePerm)
		filePathInEnv := filepath.Join(outputDir, filePath)
		err = ioutil.WriteFile(filePathInEnv, bytes, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to write file '%s' with error '%+v'", filePathInEnv, err)
		}
	}
	return nil
}

//Close cleans up any external resources
func (a *BlobStorage) Close() {
}
