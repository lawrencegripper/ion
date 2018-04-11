package app_test

import (
	"encoding/json"
	"fmt"
	"github.com/lawrencegripper/ion/common"
	"github.com/lawrencegripper/ion/sidecar/app"
	"github.com/lawrencegripper/ion/sidecar/blob/filesystem"
	"github.com/lawrencegripper/ion/sidecar/events/mock"
	"github.com/lawrencegripper/ion/sidecar/meta/inmemory"
	"github.com/lawrencegripper/ion/sidecar/types"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

var outputBlobDir string
var inputBlobDir string
var outputEventsDir string
var outputMetaFilePath string
var inputMetaFilePath string

//TODO: Use individual directories per test to enable parallelism

var persistentBlobDir string
var persistentEventsDir string

var a app.App
var db types.MetadataProvider

func TestMain(m *testing.M) {
	outputBlobDir = path.Join("testdata", "out", "data")
	inputBlobDir = path.Join("testdata", "in", "data")
	outputEventsDir = path.Join("testdata", "out", "events")
	outputMetaFilePath = path.Join("testdata", "out", "meta.json")
	inputMetaFilePath = path.Join("testdata", "out", "meta.json")
	persistentBlobDir = path.Join("testdata", "blob")
	persistentEventsDir = path.Join("testdata", "events")

	var err error
	db, err = inmemory.NewInMemoryDB()
	if err != nil {
		panic(fmt.Sprintf("failed to create in memory DB with error '%+v'", err))
	}
	blob, err := filesystem.NewBlobStorage(&filesystem.Config{
		BaseDir: persistentBlobDir,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create file system storage with error '%+v'", err))
	}
	sb := mock.NewEventPublisher(persistentEventsDir)

	logger := logrus.New()
	logger.Out = os.Stdout

	context := &types.Context{
		Name:          "testModule",
		EventID:       "01010101",
		CorrelationID: "frank",
		ParentEventID: "10101010",
	}

	eventTypes := []string{"test_events"}

	a = app.App{}
	a.Setup(
		"test",
		"testdata",
		context,
		eventTypes,
		db,
		sb,
		blob,
		logger,
		false,
	)
	go a.Run(fmt.Sprintf(":%d", 8080))
	defer a.Close()

	defer func() {
		_ = os.RemoveAll("testdata")
	}()
	os.Exit(m.Run())
}

func TestCommitBlob(t *testing.T) {
	testCases := []struct {
		files []string
	}{
		{
			files: []string{
				"file1.txt",
				"file2.txt",
			},
		},
		{
			files: []string{
				"file1.txt",
				"file2.txt",
				"file3.txt",
				"file4.txt",
			},
		},
	}
	for _, test := range testCases {
		for _, file := range test.files {
			path := path.Join(outputBlobDir, file)
			f, err := os.Create(path)
			f.Close()
			if err != nil {
				t.Errorf("error creating test file '%s'", file)
				continue
			}
		}
		_, err := a.CommitBlob(outputBlobDir)
		if err != nil {
			t.Errorf("error commiting test blobs '%+v'", err)
			continue
		}
		files, err := ioutil.ReadDir(persistentBlobDir)
		if err != nil {
			t.Errorf("error reading blob directory '%+v'", err)
			continue
		}
		outLen := len(files)
		blobLen := len(test.files)
		if outLen <= 0 {
			t.Error("expected the blob directory to not be empty but was")
			continue
		}
		if outLen != blobLen {
			t.Error("expected the blob directory to be the same size as the output directory but wasn't")
			continue
		}

		// Refresh blob directory between tests
		_ = os.RemoveAll(persistentBlobDir)
		_ = os.Mkdir(persistentBlobDir, 0777)
	}
	// Clear blob directory
	_ = os.RemoveAll(persistentBlobDir)
}

func TestCommitMeta(t *testing.T) {
	testCases := []struct {
		kvps common.KeyValuePairs
	}{
		{
			kvps: common.KeyValuePairs{
				common.KeyValuePair{
					Key:   "testKey",
					Value: "testValue",
				},
			},
		},
	}
	for _, test := range testCases {
		b, err := json.Marshal(&test.kvps)
		if err != nil {
			t.Errorf("error encoding insights: '%+v'", err)
			continue
		}
		if err := ioutil.WriteFile(outputMetaFilePath, b, 0777); err != nil {
			t.Errorf("error writing insight file: '%+v'", err)
			continue
		}
		if err := a.CommitMeta(outputMetaFilePath); err != nil {
			t.Errorf("error commiting insights: '%+v'", err)
			continue
		}
		meta := a.Meta.(*inmemory.InMemoryDB)
		for _, insight := range meta.Insights {
			if !reflect.DeepEqual(insight.Data, test.kvps) {
				t.Errorf("expected '%+v' but got '%+v'", test.kvps, insight.Data)
				continue
			}
		}
		_ = os.Remove(outputMetaFilePath)
	}
}

func TestCommitEvents(t *testing.T) {
	testCases := []struct {
		events []common.KeyValuePairs
	}{
		{
			events: []common.KeyValuePairs{
				{
					common.KeyValuePair{
						Key:   "eventType",
						Value: "test_events",
					},
					common.KeyValuePair{
						Key:   "files",
						Value: "",
					},
				},
			},
		},
		{
			events: []common.KeyValuePairs{
				{
					common.KeyValuePair{
						Key:   "eventType",
						Value: "test_events",
					},
					common.KeyValuePair{
						Key:   "files",
						Value: "file1.png",
					},
				},
			},
		},
		{
			events: []common.KeyValuePairs{
				{
					common.KeyValuePair{
						Key:   "eventType",
						Value: "test_events",
					},
				},
			},
		},
		{
			events: []common.KeyValuePairs{
				{
					common.KeyValuePair{
						Key:   "files",
						Value: "",
					},
					common.KeyValuePair{
						Key:   "eventType",
						Value: "test_events",
					},
				},
			},
		},
	}
	for _, test := range testCases {
		blobURIs := make(map[string]string)
		for i, event := range test.events {
			b, err := json.Marshal(&event)
			if err != nil {
				t.Errorf("error encoding event: '%+v'", err)
				continue
			}
			outputEventFilePath := path.Join(outputEventsDir, fmt.Sprintf("event%d.json", i))
			if err := ioutil.WriteFile(outputEventFilePath, b, 0777); err != nil {
				t.Errorf("error writing event file: '%+v'", err)
				continue
			}
			blobURIs[outputEventFilePath] = "fake.blob.uri"
		}
		if err := a.CommitEvents(outputEventsDir, blobURIs); err != nil {
			t.Errorf("error commiting events: '%+v'", err)
			continue
		}
		files, err := ioutil.ReadDir(persistentEventsDir)
		if err != nil {
			t.Errorf("error reading events directory '%s': '%+v'", persistentEventsDir, err)
			continue
		}

		inEvents := []common.Event{}
		for _, f := range files {
			b, err := ioutil.ReadFile(path.Join(persistentEventsDir, f.Name()))
			if err != nil {
				t.Errorf("error reading event '%s' from disk '%+v'", f.Name(), err)
			}
			var inEvent common.Event
			err = json.Unmarshal(b, &inEvent)
			if err != nil {
				t.Errorf("error unmarshalling event '%+v'", err)
			}
			inEvents = append(inEvents, inEvent)
		}

		inLen := len(inEvents)
		outLen := len(test.events)
		if inLen <= 0 {
			t.Errorf("expected events directory to be not empty but was")
			continue
		}
		if inLen != outLen {
			t.Error("expected the events directory to be the same size as the output directory but wasn't")
			continue
		}
		// Refresh blob directory between tests
		_ = os.RemoveAll(persistentEventsDir)
		_ = os.Mkdir(persistentEventsDir, 0777)

	}
	_ = os.RemoveAll(persistentEventsDir)
}
