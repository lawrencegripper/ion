package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
)

var a App
var sharedSecret string
var blobKey string

func TestMain(m *testing.M) {
	blobKey = "BLOBACCESSKEY123"
	sharedSecret = "SECRET"

	db := NewMockDB()
	pub := NewMockPublisher()
	blob := NewMockBlobStorage()
	logger := log.New()
	logger.Out = os.Stdout
	logger.Level = log.FatalLevel

	a = App{}
	a.Setup(
		sharedSecret,
		blobKey,
		db,
		pub,
		blob,
		logger,
	)

	code := m.Run()
	os.Exit(code)
}

func TestSecretAuth(t *testing.T) {
	testCases := []struct {
		secret     string
		statusCode int
	}{
		{
			secret:     "abcabc",
			statusCode: 401,
		},
		{
			secret:     "SECRET",
			statusCode: 200,
		},
		{
			secret:     "",
			statusCode: 401,
		},
	}
	for i, test := range testCases {
		req, _ := http.NewRequest("GET", "/meta/0", nil)
		if test.secret != "" {
			req.Header.Add("Secret", test.secret)
		}
		res := executeRequest(req)

		checkError(t, i, checkResponseCode(test.statusCode, res.Code))
	}
}

func TestGetMetadataByID(t *testing.T) {
	testCases := []struct {
		id         string
		statusCode int
		expected   Document
	}{
		{
			id:         "0",
			statusCode: 200,
			expected: Document{
				ID: "0",
				Entries: []Entry{
					{
						ID: "0",
						Metadata: map[string]string{
							"TEST123": "123TEST",
						},
					},
				},
			},
		},
		{
			id:         "5",
			statusCode: 404,
			expected:   Document{},
		},
	}
	for i, test := range testCases {
		req, _ := http.NewRequest("GET", "/meta/"+test.id, nil)
		req.Header.Add("Secret", sharedSecret)
		res := executeRequest(req)

		checkError(t, i, checkResponseCode(test.statusCode, res.Code))

		if res.Code == http.StatusOK {
			checkError(t, i, checkMetadata(test.expected, res.Body))
		}
	}
}

func TestUpdateMetadata(t *testing.T) {
	testCases := []struct {
		id               string
		updateStatusCode int
		getStatusCode    int
		patch            Entry
		expected         Document
	}{
		{
			id:               "0",
			updateStatusCode: 200,
			patch: Entry{
				ID: "1",
				Metadata: map[string]string{
					"BOB": "ALICE",
					"1":   "2",
					"big": "small",
				},
			},
			getStatusCode: 200,
			expected: Document{
				ID: "0",
				Entries: []Entry{
					{
						ID: "0",
						Metadata: map[string]string{
							"TEST123": "123TEST",
						},
					},
					{
						ID: "1",
						Metadata: map[string]string{
							"BOB": "ALICE",
							"1":   "2",
							"big": "small",
						},
					},
				},
			},
		},
		{
			id:               "5",
			updateStatusCode: 404,
			patch: Entry{
				ID: "1",
				Metadata: map[string]string{
					"BOB": "ALICE",
					"1":   "2",
					"big": "small",
				},
			},
			getStatusCode: 404,
			expected:      Document{},
		},
	}
	for i, test := range testCases {
		// perform update
		b, _ := json.Marshal(test.patch)
		updateReq, _ := http.NewRequest("POST", "/meta/"+test.id, bytes.NewReader(b))
		updateReq.Header.Add("Secret", sharedSecret)

		updateRes := executeRequest(updateReq)

		checkError(t, i, checkResponseCode(test.updateStatusCode, updateRes.Code))

		// check update
		getReq, _ := http.NewRequest("GET", "/meta/"+test.id, nil)
		getReq.Header.Add("Secret", sharedSecret)
		getRes := executeRequest(getReq)

		checkError(t, i, checkResponseCode(test.getStatusCode, getRes.Code))

		if getRes.Code == http.StatusOK {
			checkError(t, i, checkMetadata(test.expected, getRes.Body))
		}
		a.Close()
	}
}

func TestPublishEvent(t *testing.T) {
	//TODO
}

func TestGetBlobAccessKey(t *testing.T) {
	testCases := []struct {
		statusCode int
		expected   string
	}{
		{
			statusCode: 200,
			expected:   blobKey,
		},
	}
	for i, test := range testCases {
		req, _ := http.NewRequest("GET", "/blob", nil)
		req.Header.Add("Secret", sharedSecret)
		res := executeRequest(req)

		checkError(t, i, checkResponseCode(test.statusCode, res.Code))

		if res.Code == http.StatusOK {
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Errorf("error thrown in test case %d: %+v", i, err)
			}
			bodyStr := string(body)
			if bodyStr != blobKey {
				err := fmt.Errorf("expected blobkey %s. Got %s\n", blobKey, bodyStr)
				t.Errorf("error thrown in test case %d: %+v", i, err)
			}
		}
	}
}

func TestGetBlobsInContainerByID(t *testing.T) {
	testCases := []struct {
		containerName string
		statusCode    int
		expected      []BlobInfo
	}{
		{
			containerName: "test",
			statusCode:    200,
			expected: []BlobInfo{
				{
					Name: "blob1",
					URI:  "https://blob.com/container1/blob1?sas=1234",
				},
				{
					Name: "blob2",
					URI:  "https://blob.com/container2/blob2?sas=1234",
				},
				{
					Name: "blob3",
					URI:  "https://blob.com/container3/blob3?sas=1234",
				},
			},
		},
		{
			containerName: "test123",
			statusCode:    404,
			expected:      nil,
		},
	}
	for i, test := range testCases {
		req, _ := http.NewRequest("GET", "/blob/container/"+test.containerName, nil)
		req.Header.Add("Secret", sharedSecret)
		res := executeRequest(req)

		checkError(t, i, checkResponseCode(test.statusCode, res.Code))

		if res.Code == http.StatusOK {
			checkError(t, i, checkBlobs(test.expected, res.Body))
		}
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(expected, actual int) error {
	if expected != actual {
		return fmt.Errorf("expected response code %d. Got %d\n", expected, actual)
	}
	return nil
}

func checkMetadata(expectedDoc Document, reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	var doc Document
	err := decoder.Decode(&doc)
	if err != nil {
		return fmt.Errorf("could not decode metadata body: %+v", err)
	}
	if !reflect.DeepEqual(expectedDoc, doc) {
		return fmt.Errorf("expected document %+v. Got %+v\n", expectedDoc, doc)
	}
	return nil
}

func checkBlobs(expectedBlobs []BlobInfo, reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	var blobs []BlobInfo
	err := decoder.Decode(&blobs)
	if err != nil {
		return fmt.Errorf("could not decode metadata body: %+v", err)
	}
	if !reflect.DeepEqual(expectedBlobs, blobs) {
		return fmt.Errorf("expected document %+v. Got %+v\n", expectedBlobs, blobs)
	}
	return nil
}

func checkError(t *testing.T, i int, err error) {
	if err != nil {
		t.Errorf("error thrown in test case %d: %+v", i, err)
	}
}
