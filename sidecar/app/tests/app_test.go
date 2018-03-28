package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"testing"

	a "github.com/lawrencegripper/mlops/sidecar/app"
	fs "github.com/lawrencegripper/mlops/sidecar/blob/filesystem"
	"github.com/lawrencegripper/mlops/sidecar/events/mock"
	"github.com/lawrencegripper/mlops/sidecar/meta/inmemory"
	"github.com/lawrencegripper/mlops/sidecar/types"
	"github.com/sirupsen/logrus"
)

const secret string = "secret"
const eventID string = "2"
const parentEventID string = "1"
const correlationID string = "0"
const tempDir string = "temp"

var app a.App

func TestMain(m *testing.M) {

	initialMeta := map[string][]types.MetaDoc{
		correlationID: []types.MetaDoc{
			types.MetaDoc{
				ID:            "0",
				ParentEventID: "0",
				CorrelationID: "0",
				Metadata: map[string]string{
					"name":       "alice",
					"age":        "40",
					"occupation": "professor",
				},
			},
			types.MetaDoc{
				ID:            "1",
				ParentEventID: "0",
				CorrelationID: "0",
				Metadata: map[string]string{
					"name":       "frank",
					"age":        "51",
					"occupation": "teacher",
				},
			},
			types.MetaDoc{
				ID:            "2",
				ParentEventID: "1",
				CorrelationID: "0",
				Metadata: map[string]string{
					"name":       "muhammad",
					"age":        "23",
					"occupation": "athlete",
				},
			},
		},
		"0987654321": []types.MetaDoc{
			types.MetaDoc{
				ID:            "9",
				ParentEventID: "0",
				CorrelationID: "5",
				Metadata: map[string]string{
					"colour": "blue",
					"size":   "large",
					"weight": "heavy",
				},
			},
		},
	}
	metaProvider := inmemory.NewInMemoryMetaProvider(initialMeta)

	os.MkdirAll(tempDir, 0644)
	parentDir := path.Join(tempDir, parentEventID)
	os.MkdirAll(parentDir, 0644)
	tempFile := path.Join(parentDir, "test.txt")
	os.MkdirAll(tempDir, 0644)
	err := ioutil.WriteFile(tempFile, []byte("hello world"), 0644)
	if err != nil {
		panic(err)
	}
	blobProvider := fs.NewFileSystemBlobProvider(tempDir)
	publisher := mock.NewMockEventPublisher()

	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Level = logrus.FatalLevel

	app = a.App{}
	app.Setup(
		secret,
		eventID,
		correlationID,
		parentEventID,
		metaProvider,
		publisher,
		blobProvider,
		logger,
	)

	code := m.Run()

	_ = os.RemoveAll(tempDir)

	os.Exit(code)
}

func TestSecretAuth(t *testing.T) {
	testCases := []struct {
		secret string
		code   int
	}{
		{secret: "secret", code: 200},
		{secret: "abcabc", code: 401},
		{secret: "", code: 401},
	}
	for i, test := range testCases {
		req, _ := http.NewRequest(http.MethodGet, "/meta", nil)
		if test.secret != "" {
			req.Header.Add("secret", test.secret)
		}
		res := executeRequest(req)
		checkError(t, i, checkResponseCode(test.code, res.Code))
	}
}

func TestGetMetaDocByID(t *testing.T) {
	testCases := []struct {
		route      string
		statusCode int
		expected   map[string]string
	}{
		{
			route:      "/self/meta",
			statusCode: 200,
			expected: map[string]string{
				"name":       "muhammad",
				"age":        "23",
				"occupation": "athlete",
			},
		},
		{
			route:      "/parent/meta",
			statusCode: 200,
			expected: map[string]string{
				"name":       "frank",
				"age":        "51",
				"occupation": "teacher",
			},
		},
	}
	for i, test := range testCases {
		req, _ := http.NewRequest("GET", test.route, nil)
		req.Header.Add("secret", secret)
		res := executeRequest(req)

		checkError(t, i, checkResponseCode(test.statusCode, res.Code))

		if res.Code == http.StatusOK {
			checkError(t, i, checkMetadata(test.expected, res.Body))
		}
	}
}

func TestGetMetaDocAll(t *testing.T) {
	testCases := []struct {
		statusCode int
		expected   []types.MetaDoc
	}{
		{
			statusCode: 200,
			expected: []types.MetaDoc{
				types.MetaDoc{
					ID:            "0",
					ParentEventID: "0",
					CorrelationID: "0",
					Metadata: map[string]string{
						"name":       "alice",
						"age":        "40",
						"occupation": "professor",
					},
				},
				types.MetaDoc{
					ID:            "1",
					ParentEventID: "0",
					CorrelationID: "0",
					Metadata: map[string]string{
						"name":       "frank",
						"age":        "51",
						"occupation": "teacher",
					},
				},
				types.MetaDoc{
					ID:            "2",
					ParentEventID: "1",
					CorrelationID: "0",
					Metadata: map[string]string{
						"name":       "muhammad",
						"age":        "23",
						"occupation": "athlete",
					},
				},
			},
		},
	}
	for i, test := range testCases {
		req, _ := http.NewRequest("GET", "/meta", nil)
		req.Header.Add("secret", secret)
		res := executeRequest(req)

		checkError(t, i, checkResponseCode(test.statusCode, res.Code))

		if res.Code == http.StatusOK {
			checkError(t, i, checkMetadataAll(test.expected, res.Body))
		}
	}
}

func TestUpdateMetaDoc(t *testing.T) {
	testCases := []struct {
		updateStatusCode int
		getStatusCode    int
		route            string
		eventID          string
		patch            map[string]string
	}{
		{
			updateStatusCode: 200,
			route:            "/self/meta",
			patch: map[string]string{
				"Title":  "Effective C++",
				"Author": "Scott Meyers",
			},
			getStatusCode: 200,
		},
	}
	for i, test := range testCases {
		b, _ := json.Marshal(test.patch)
		updateReq, _ := http.NewRequest("PUT", test.route, bytes.NewReader(b))
		updateReq.Header.Add("secret", secret)

		doc, _ := app.Meta.GetMetaDocByID(eventID)
		doc.Metadata["Title"] = "Effective C++"
		doc.Metadata["Author"] = "Scott Meyers"

		updateRes := executeRequest(updateReq)
		checkError(t, i, checkResponseCode(test.updateStatusCode, updateRes.Code))

		docAfter, _ := app.Meta.GetMetaDocByID(eventID)

		for k, v := range docAfter.Metadata {
			if v != doc.Metadata[k] {
				checkError(t, i, fmt.Errorf("values for key '%s' do not match, expected '%+v'. Got '%+v'", k, v, doc.Metadata[k]))
			}
		}
	}
}

func TestGetBlob(t *testing.T) {
	testCases := []struct {
		route   string
		path    string
		code    int
		content string
	}{
		{
			route:   "/parent/blob",
			path:    "test.txt",
			code:    200,
			content: "hello world",
		},
		{
			route:   "/self/blob",
			path:    "fake.txt",
			code:    404,
			content: "",
		},
	}
	for i, test := range testCases {
		req, _ := http.NewRequest("GET", test.route+"?res="+test.path, nil)
		req.Header.Add("secret", secret)
		res := executeRequest(req)

		checkError(t, i, checkResponseCode(test.code, res.Code))

		if res.Code == http.StatusOK {
			content, _ := ioutil.ReadAll(res.Body)
			checkError(t, i, checkContent(test.content, string(content)))
		}
	}
}

func TestCreateBlob(t *testing.T) {
	testCases := []struct {
		path    string
		code    int
		content string
	}{
		{
			path:    "test2.txt",
			code:    200,
			content: "the cake is a lie!",
		},
		{
			path:    "images/myimage.png",
			code:    200,
			content: "13780787113102610",
		},
	}
	createdFiles := make([]string, 0)
	for i, test := range testCases {
		b := []byte(test.content)
		req, _ := http.NewRequest("PUT", "/self/blob?res="+test.path, bytes.NewReader(b))
		req.Header.Add("secret", secret)
		res := executeRequest(req)

		checkError(t, i, checkResponseCode(test.code, res.Code))
		createdFiles = append(createdFiles, test.path)
	}
	for _, file := range createdFiles {
		dir, _ := path.Split(file)
		path := path.Join(tempDir, eventID, dir)
		_ = os.RemoveAll(path)
	}
}

func TestListBlob(t *testing.T) {
	testCases := []struct {
		code  int
		paths []string
		list  []string
	}{
		{
			code: 200,
			paths: []string{
				"file1.txt",
				"file2.txt",
				"file3.txt",
			},
			list: []string{
				"file1.txt",
				"file2.txt",
				"file3.txt",
			},
		},
	}
	for i, test := range testCases {

		for _, p := range test.paths {
			b := []byte(p)
			req, _ := http.NewRequest("PUT", "/self/blob?res="+p, bytes.NewReader(b))
			req.Header.Add("secret", secret)
			_ = executeRequest(req)
		}

		listReq, _ := http.NewRequest("GET", "/self/blobs", nil)
		listReq.Header.Add("secret", secret)
		res := executeRequest(listReq)

		checkError(t, i, checkResponseCode(test.code, res.Code))

		checkError(t, i, checkList(test.list, res.Body))
	}
}

func TestDeleteBlob(t *testing.T) {
	testCases := []struct {
		code int
		path string
	}{
		{
			code: 200,
			path: "deleteme.avi",
		},
	}
	for i, test := range testCases {
		b := []byte(test.path)
		createReq, _ := http.NewRequest("PUT", "/self/blob?res="+test.path, bytes.NewReader(b))
		createReq.Header.Add("secret", secret)
		_ = executeRequest(createReq)

		getBeforeReq, _ := http.NewRequest("GET", "/self/blob?res="+test.path, nil)
		getBeforeReq.Header.Add("secret", secret)
		getBeforeRes := executeRequest(getBeforeReq)

		checkError(t, i, checkResponseCode(http.StatusOK, getBeforeRes.Code))

		deleteReq, _ := http.NewRequest("DELETE", "/self/blob?res="+test.path, nil)
		deleteReq.Header.Add("secret", secret)
		res := executeRequest(deleteReq)

		checkError(t, i, checkResponseCode(test.code, res.Code))

		getAfterReq, _ := http.NewRequest("GET", "/self/blob?res="+test.path, nil)
		getAfterReq.Header.Add("secret", secret)
		getAfterRes := executeRequest(getAfterReq)

		checkError(t, i, checkResponseCode(http.StatusNotFound, getAfterRes.Code))
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)
	return rr
}

func checkError(t *testing.T, i int, err error) {
	if err != nil {
		t.Errorf("error thrown in test case %d: %+v", i, err)
	}
}

func checkResponseCode(expected, actual int) error {
	if expected != actual {
		return fmt.Errorf("expected response code %d. Got %d", expected, actual)
	}
	return nil
}

func checkMetadata(expected map[string]string, reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	var actual map[string]string
	err := decoder.Decode(&actual)
	if err != nil {
		return fmt.Errorf("could not decode metadata body: %+v", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		return fmt.Errorf("expected document %+v. Got %+v", expected, actual)
	}
	return nil
}

func checkMetadataAll(expected []types.MetaDoc, reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	var actual []types.MetaDoc
	err := decoder.Decode(&actual)
	if err != nil {
		return fmt.Errorf("could not decode metadata body: %+v", err)
	}
	if len(expected) != len(actual) {
		return fmt.Errorf("expected document %+v. Got %+v", expected, actual)
	}
	return nil
}

func checkContent(expected, actual string) error {
	if expected != actual {
		return fmt.Errorf("expected document %+v. Got %+v", expected, actual)
	}
	return nil
}

func checkList(expected []string, reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	var actual []string
	err := decoder.Decode(&actual)
	if err != nil {
		return fmt.Errorf("could not decode metadata body: %+v", err)
	}
	if len(expected) != len(actual) {
		return fmt.Errorf("expected document %+v. Got %+v", expected, actual)
	}
	return nil
}
