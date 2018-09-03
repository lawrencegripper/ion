package committer_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/lawrencegripper/ion/internal/app/handler/committer"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/blobstorage/filesystem"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage/inmemory"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/events/mock"
	"github.com/lawrencegripper/ion/internal/app/handler/module"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	log "github.com/sirupsen/logrus"
)

const testdata = "testdata"

var c *committer.Committer
var environment *module.Environment
var dataPlane *dataplane.DataPlane
var eventTypes []string
var context *common.Context

var persistentInBlobDir string
var persistentOutBlobDir string
var persistentEventsDir string

func TestMain(m *testing.M) {

	eventID := "01010101"
	parentEventID := "10101010"
	persistentInBlobDir = filepath.FromSlash(fmt.Sprintf("%s/%s/blob", testdata, parentEventID))
	persistentOutBlobDir = filepath.FromSlash(fmt.Sprintf("%s/%s/blob", testdata, eventID))
	persistentEventsDir = filepath.FromSlash(fmt.Sprintf("%s/events", testdata))
	eventTypes = append(eventTypes, "test_events")

	// Create mock dataplane...

	// Metadata store
	meta, err := inmemory.NewInMemoryDB()
	if err != nil {
		panic(fmt.Sprintf("failed to create in memory DB with error '%+v'", err))
	}

	// Blob storage
	blob, err := filesystem.NewBlobStorage(&filesystem.Config{
		InputDir:  persistentInBlobDir,
		OutputDir: persistentOutBlobDir,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create file system storage with error '%+v'", err))
	}

	// Messaging system
	events := mock.NewEventPublisher(persistentEventsDir)

	dataPlane = &dataplane.DataPlane{
		BlobStorageProvider:     blob,
		DocumentStorageProvider: meta,
		EventPublisher:          events,
	}

	// Create mock context
	context = &common.Context{
		Name:          "testModule",
		EventID:       "eventid",
		CorrelationID: "frank",
		ParentEventID: "parentid",
	}

	environment = module.GetModuleEnvironment(testdata)
	environment.Build()

	// Create committer
	log.SetOutput(os.Stdout)
	c = committer.NewCommitter(testdata, nil)

	exitCode := m.Run()

	c.Close()
	_ = os.RemoveAll(testdata)
	_ = os.Remove(".memdb")
	os.Exit(exitCode)
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
			path := filepath.FromSlash(path.Join(environment.OutputBlobDirPath, file))
			f, err := os.Create(path)
			f.Close()
			if err != nil {
				t.Fatalf("error creating test file '%s'", file)
				continue
			}
		}
		if err := c.Commit(context, dataPlane, eventTypes); err != nil {
			t.Fatal(err)
		}
		files, err := ioutil.ReadDir(persistentOutBlobDir)
		if err != nil {
			t.Fatalf("error reading blob directory '%+v'", err)
			continue
		}
		outLen := len(files)
		blobLen := len(test.files)
		if outLen <= 0 {
			t.Fatal("expected the blob directory to be populated but it was empty")
			continue
		}
		if outLen != blobLen {
			t.Fatal("expected the blob directory to be the same size as the output directory but wasn't")
			continue
		}

		reset()
	}
}

func TestCommitInsights(t *testing.T) {
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
		if err := ioutil.WriteFile(environment.OutputMetaFilePath, b, 0777); err != nil {
			t.Errorf("error writing insight file: '%+v'", err)
			continue
		}
		if err := c.Commit(context, dataPlane, eventTypes); err != nil {
			t.Errorf("error commiting insights: '%+v'", err)
			continue
		}
		meta := dataPlane.DocumentStorageProvider.(*inmemory.InMemoryDB)
		for _, insight := range meta.Insights {
			if !reflect.DeepEqual(insight.Data, test.kvps) {
				t.Errorf("expected '%+v' but got '%+v'", test.kvps, insight.Data)
				continue
			}
		}
		reset()
	}
}

func TestCommitEvents(t *testing.T) {
	testCases := []struct {
		events []common.KeyValuePairs
		files  []string
		err    bool
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
			files: []string{},
			err:   false,
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
			files: []string{
				"file1.png",
			},
			err: false,
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
			files: []string{},
			err:   false,
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
			files: []string{},
			err:   false,
		},
		{
			events: []common.KeyValuePairs{
				{
					common.KeyValuePair{
						Key:   "files",
						Value: "nonexistentfile.jpg,existentfile.png",
					},
					common.KeyValuePair{
						Key:   "eventType",
						Value: "test_events",
					},
				},
			},
			files: []string{
				"existentfile.png",
			},
			err: true,
		},
		{
			events: []common.KeyValuePairs{
				{
					common.KeyValuePair{
						Key:   "files",
						Value: "myfile.jpg",
					},
					common.KeyValuePair{
						Key:   "eventType",
						Value: "invalid",
					},
				},
			},
			files: []string{
				"myfile.jpg",
			},
			err: true,
		},
		{
			events: []common.KeyValuePairs{
				{
					common.KeyValuePair{
						Key:   "files",
						Value: "",
					},
				},
			},
			files: []string{},
			err:   true,
		},
	}
	for _, test := range testCases {
		for _, file := range test.files {
			path := filepath.FromSlash(path.Join(environment.OutputBlobDirPath, file))
			f, err := os.Create(path)
			f.Close()
			if err != nil {
				t.Fatalf("error creating test file '%s'", file)
				continue
			}
		}
		blobURIs := make(map[string]string)
		for i, event := range test.events {
			b, err := json.Marshal(&event)
			if err != nil {
				t.Errorf("error encoding event: '%+v'", err)
				continue
			}
			outputEventFilePath := filepath.FromSlash(path.Join(environment.OutputEventsDirPath, fmt.Sprintf("event%d.json", i)))
			if err := ioutil.WriteFile(outputEventFilePath, b, 0777); err != nil {
				t.Errorf("error writing event file: '%+v'", err)
				continue
			}
			blobURIs[outputEventFilePath] = "fake.blob.uri"
		}
		if err := c.Commit(context, dataPlane, eventTypes); err != nil {
			if !test.err {
				t.Errorf("error commiting events: '%+v'", err)
			}
			continue
		}
		files, err := ioutil.ReadDir(persistentEventsDir)
		if err != nil {
			t.Errorf("error reading events directory '%s': '%+v'", persistentEventsDir, err)
			continue
		}

		inEvents := []common.Event{}
		for _, f := range files {
			b, err := ioutil.ReadFile(filepath.FromSlash(path.Join(persistentEventsDir, f.Name())))
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
		reset()
	}
}

func reset() {
	refreshDataplane()
	refreshEnv()
}

func refreshDataplane() {
	refreshDir(persistentEventsDir)
	refreshDir(persistentInBlobDir)
	refreshDir(persistentOutBlobDir)
}

func refreshDir(dirPath string) {
	_ = os.RemoveAll(dirPath)
	_ = os.MkdirAll(dirPath, os.ModePerm)
}

func refreshEnv() {
	_ = environment.Clear()
}
