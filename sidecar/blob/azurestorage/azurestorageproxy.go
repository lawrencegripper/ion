package azurestorage

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/lawrencegripper/ion/sidecar/types"
)

//AzureBlobProxy represents an Azure blob proxy object
type AzureBlobProxy struct {
	types.Proxy
	parent *BlobStorage
}

//NewAzureBlobProxy creates a new blob proxy instance
func NewAzureBlobProxy(proxy types.Proxy, parent *BlobStorage) types.BlobProxy {
	return &AzureBlobProxy{
		Proxy:  proxy,
		parent: parent,
	}
}

//Get fulfils a request to get a blob resource by proxying directly to the Azure Storage REST API
func (p *AzureBlobProxy) Get(resourcePath string, w http.ResponseWriter, r *http.Request) {
	blob, err := getBlobFromResourcePath(&p.parent.blobClient, resourcePath)
	if err != nil {
		resErr := types.ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
		resErr.Send(w)
	}
	sasURL, err := getAuthToken(24, true, false, false, false, false, blob)
	if err != nil {
		resErr := types.ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
		resErr.Send(w)
	}
	r.URL, err = url.Parse(sasURL)
	if err != nil {
		resErr := types.ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
		resErr.Send(w)
	}
	p.Proxy.ServeHTTP(w, r)
}

//Create fulfils a request to create a blob resource by proxying directly to the Azure Storage REST API
func (p *AzureBlobProxy) Create(resourcePath string, w http.ResponseWriter, r *http.Request) {
	blob, err := createContainerForBlobIfNotExist(&p.parent.blobClient, resourcePath)
	if err != nil {
		resErr := types.ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
		resErr.Send(w)
	}
	sasURL, err := getAuthToken(24, true, true, true, true, true, blob)
	if err != nil {
		resErr := types.ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
		resErr.Send(w)
	}
	r.URL, err = url.Parse(sasURL)
	if err != nil {
		resErr := types.ErrorResponse{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
		resErr.Send(w)
	}
	r.Header.Set("x-ms-blob-type", "BlockBlob")
	rw := types.NewStatusCodeResponseWriter(w)

	// Forward the request to Azure
	p.Proxy.ServeHTTP(rw, r)

	// Return the newly creates resource URI
	defer func(url string, rw *types.StatusCodeResponseWriter) {
		if rw.StatusCode == http.StatusCreated {
			segs := strings.Split(url, "?") // remove SAS token if present
			if len(segs) < 2 {
				_, _ = w.Write([]byte(url))
			}
			_, _ = w.Write([]byte(segs[0]))
		}
	}(sasURL, rw)
}
