package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

var a App
var sharedSecret string

func TestMain(m *testing.M) {
	blobAccessToken := "MYACCESSTOKEN"
	sharedSecret = "SECRET"

	mockDB := NewMockDB()

	a = App{}
	a.Setup(
		sharedSecret,
		blobAccessToken,
		mockDB,
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
	for _, test := range testCases {
		req, _ := http.NewRequest("GET", "/meta/0", nil)
		if test.secret != "" {
			req.Header.Add("Secret", test.secret)
		}
		res := executeRequest(req)

		checkResponseCode(t, test.statusCode, res.Code)
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
	for _, test := range testCases {
		req, _ := http.NewRequest("GET", "/meta/"+test.id, nil)
		req.Header.Add("Secret", sharedSecret)
		res := executeRequest(req)

		checkResponseCode(t, test.statusCode, res.Code)

		checkMetadata(t, test.expected, res.Body)
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
	for _, test := range testCases {
		// perform update
		b, _ := json.Marshal(test.patch)
		updateReq, _ := http.NewRequest("POST", "/meta/"+test.id, bytes.NewReader(b))
		updateReq.Header.Add("Secret", sharedSecret)

		updateRes := executeRequest(updateReq)

		checkResponseCode(t, test.updateStatusCode, updateRes.Code)

		// check update
		getReq, _ := http.NewRequest("GET", "/meta/"+test.id, nil)
		getReq.Header.Add("Secret", sharedSecret)
		getRes := executeRequest(getReq)

		checkResponseCode(t, test.getStatusCode, getRes.Code)

		if getRes.Code == http.StatusOK {
			checkMetadata(t, test.expected, getRes.Body)
		}
		a.Close()
	}
}

func TestPublishEvent(t *testing.T) {
	req, _ := http.NewRequest("GET", "/events", nil)
	res := executeRequest(req)

	checkResponseCode(t, http.StatusOK, res.Code)

	if body := res.Body.String(); body != "[]" {
		t.Errorf("expected an empty array. Got %s", body)
	}
}

func TestGetBlobAccessKey(t *testing.T) {
	req, _ := http.NewRequest("GET", "/blob", nil)
	res := executeRequest(req)

	checkResponseCode(t, http.StatusOK, res.Code)

	if body := res.Body.String(); body != "[]" {
		t.Errorf("expected an empty array. Got %s", body)
	}
}

type MockDB struct {
	Document *Document
}

func (db *MockDB) GetByID(id string) (*Document, error) {
	if id != db.Document.ID {
		return nil, fmt.Errorf("the requested document does not exist")
	}
	return db.Document, nil
}

func (db *MockDB) Update(id string, entry Entry) error {
	if id != db.Document.ID {
		return fmt.Errorf("the document you are trying to update does not exists")
	}
	db.Document.Entries = append(db.Document.Entries, entry)
	return nil
}

func (db *MockDB) Close() {
	db.Document = &Document{
		ID: "0",
		Entries: []Entry{
			{
				ID: "0",
				Metadata: map[string]string{
					"TEST123": "123TEST",
				},
			},
		},
	}
}

func NewMockDB() *MockDB {
	db := &MockDB{}
	defer db.Close() // Resets mock DB to a known state
	return db
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("expected response code %d. Got %d\n", expected, actual)
	}
}

func checkMetadata(t *testing.T, expectedDoc Document, reader io.Reader) {
	decoder := json.NewDecoder(reader)
	var doc Document
	err := decoder.Decode(&doc)
	if err != nil {
		t.Errorf("could not decode metadata body: %+v", err)
	}
	if !reflect.DeepEqual(expectedDoc, doc) {
		t.Errorf("expected document %+v. Got %+v\n", expectedDoc, doc)
	}
}
