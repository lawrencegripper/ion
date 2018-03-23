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
	"reflect"
	"testing"

	"github.com/lawrencegripper/mlops/sidecar/app"
	"github.com/lawrencegripper/mlops/sidecar/common"
	log "github.com/sirupsen/logrus"
)

var a app.App
var sharedsecret, blobKey, correlationID, parentEventID, eventID string

func TestMain(m *testing.M) {
	blobKey = "BLOBACCESSKEY123"
	sharedsecret = "secret"
	correlationID = "0"
	parentEventID = "1"
	eventID = "3"

	db := NewMockDB()
	pub := NewMockPublisher()
	blob := NewMockBlobStorage()
	logger := log.New()
	logger.Out = os.Stdout
	logger.Level = log.FatalLevel

	a = app.App{}
	a.Setup(
		sharedsecret,
		blobKey,
		eventID,
		correlationID,
		parentEventID,
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
			secret:     "secret",
			statusCode: 200,
		},
		{
			secret:     "",
			statusCode: 401,
		},
	}
	for i, test := range testCases {
		req, _ := http.NewRequest("GET", "/meta", nil)
		if test.secret != "" {
			req.Header.Add("secret", test.secret)
		}
		res := executeRequest(req)

		checkError(t, i, checkResponseCode(test.statusCode, res.Code))
	}
}

func TestGetMetaDocByID(t *testing.T) {
	testCases := []struct {
		statusCode int

		expected map[string]string
	}{
		{
			statusCode: 200,
			expected: map[string]string{
				"ALICE": "BOB",
			},
		},
	}
	for i, test := range testCases {
		req, _ := http.NewRequest("GET", "/meta/inputs", nil)
		req.Header.Add("secret", sharedsecret)
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
		expected   []common.MetaDoc
	}{
		{
			statusCode: 200,
			expected: []common.MetaDoc{
				{
					ID:            "0",
					CorrelationID: "0",
					ParentEventID: "0",
					Metadata: map[string]string{
						"TEST123": "123TEST",
					},
				},
				{
					ID:            "1",
					CorrelationID: "0",
					ParentEventID: "0",
					Metadata: map[string]string{
						"ALICE": "BOB",
					},
				},
				{
					ID:            "2",
					CorrelationID: "0",
					ParentEventID: "1",
					Metadata: map[string]string{
						"BLUE": "GREEN",
						"JACK": "JILL",
					},
				},
			},
		},
	}
	for i, test := range testCases {
		req, _ := http.NewRequest("GET", "/meta", nil)
		req.Header.Add("secret", sharedsecret)
		res := executeRequest(req)

		checkError(t, i, checkResponseCode(test.statusCode, res.Code))

		if res.Code == http.StatusOK {
			checkError(t, i, checkMetaDocs(test.expected, res.Body))
		}
	}
}

func TestUpdateMetaDoc(t *testing.T) {
	testCases := []struct {
		updateStatusCode int
		getStatusCode    int
		patch            map[string]string
		expectedDiff     int
	}{
		{
			updateStatusCode: 200,
			patch: map[string]string{
				"BOB": "ALICE",
				"1":   "2",
				"big": "small",
			},
			getStatusCode: 200,
			expectedDiff:  1,
		},
	}
	for i, test := range testCases {
		b, _ := json.Marshal(test.patch)
		updateReq, _ := http.NewRequest("POST", "/meta/outputs", bytes.NewReader(b))
		updateReq.Header.Add("secret", sharedsecret)

		docs, _ := a.MetaDB.GetMetaDocAll(correlationID)
		initial := len(docs)

		updateRes := executeRequest(updateReq)
		checkError(t, i, checkResponseCode(test.updateStatusCode, updateRes.Code))

		docs, _ = a.MetaDB.GetMetaDocAll(correlationID)
		if docs == nil {
			checkError(t, i, fmt.Errorf("update failed to initialize new metadata document"))
		}
		after := len(docs)
		if test.expectedDiff != (after - initial) {
			checkError(t, i, fmt.Errorf("expected a difference of %d but got %d", test.expectedDiff, after))
		}

		a.Close()
	}
}

func TestPublishEvent(t *testing.T) {
	//TODO
}

func TestGetBlobInputsURL(t *testing.T) {
	testCases := []struct {
		baseURL     string
		querystring string
		statusCode  int
		expected    string
	}{
		{
			querystring: "url=https://blob.com/container/myfile.jpeg",
			statusCode:  200,
			expected:    "https://blob.com/container/myfile.jpeg?se=2018-03-10T19%3A33%3A48Z&sig=nm%2B53E%2FgqklbjmkcvG2bTKGaIOJSGNDS%3D&sp=rl&spr=https&sr=c&st=2018-03-10T19%3A33%3A48Z&sv=2016-05-31",
		},
		{
			querystring: "test=123",
			statusCode:  400,
			expected:    "",
		},
	}
	for i, test := range testCases {
		req, _ := http.NewRequest("GET", "/blob/inputs?"+test.querystring, nil)
		req.Header.Add("secret", sharedsecret)
		res := executeRequest(req)

		checkError(t, i, checkResponseCode(test.statusCode, res.Code))

		if res.Code == http.StatusOK {
			checkError(t, i, checkSAS(test.expected, res.Body))
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

func checkMetaDocs(expected []common.MetaDoc, reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	var actual []common.MetaDoc
	err := decoder.Decode(&actual)
	if err != nil {
		return fmt.Errorf("could not decode metadata body: %+v", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		return fmt.Errorf("expected document %+v. Got %+v", expected, actual)
	}
	return nil
}

func checkSAS(expected string, reader io.Reader) error {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read request body")
	}
	actual := string(bytes)
	if actual != expected {
		return fmt.Errorf("expected response code %s. Got %s", expected, actual)
	}
	return nil
}

func checkError(t *testing.T, i int, err error) {
	if err != nil {
		t.Errorf("error thrown in test case %d: %+v", i, err)
	}
}
