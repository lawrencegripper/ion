package development

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// BlobsDirExt is the directory extension of this module development blobs
const BlobsDirExt = "_blobs"

// EventsDirExt is the directory extension of this module development events
const EventsDirExt = "_events"

// MetadataDirExt is the directory extension of this module development metadata
const MetadataDirExt = "_metadata"

// InsightsDirExt is the directory extension of this module development insights
const InsightsDirExt = "_insights"

// Configuration holds development config
type Configuration struct {
	ModuleDir       string
	ParentModuleDir string
	Enabled         bool
	BaseDir         string
}

// Init initializes a new development configuration
func (c *Configuration) Init(parentEventID, eventID string) error {
	if eventID == "" {
		return fmt.Errorf("cannot initialize development configuration for empty eventID")
	}

	if _, err := os.Stat(c.BaseDir); os.IsNotExist(err) {
		if err = os.MkdirAll(c.BaseDir, os.ModePerm); err != nil {
			return fmt.Errorf("error creating development base directory %+v", err)
		}
	}

	err := filepath.Walk(c.BaseDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil // ignore files
		}
		if filepath.Base(path) == parentEventID {
			c.ParentModuleDir, _ = filepath.Abs(path)
			return nil
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error searching for parent development dir %s in base dir %s", parentEventID, c.BaseDir)
	}

	prefix := ""
	if c.ParentModuleDir == "" {
		prefix = c.BaseDir
	}

	// If parentModuleDir is empty, assume module is root of tree
	moduleDir := filepath.Join(prefix, c.ParentModuleDir, eventID)
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		_ = os.Mkdir(moduleDir, os.ModePerm)
	}
	blobsDir := filepath.Join(moduleDir, BlobsDirExt)
	if _, err := os.Stat(blobsDir); os.IsNotExist(err) {
		_ = os.Mkdir(blobsDir, os.ModePerm)
	}
	eventsDir := filepath.Join(moduleDir, EventsDirExt)
	if _, err := os.Stat(eventsDir); os.IsNotExist(err) {
		_ = os.Mkdir(eventsDir, os.ModePerm)

	}
	metadataDir := filepath.Join(moduleDir, MetadataDirExt)
	if _, err := os.Stat(metadataDir); os.IsNotExist(err) {
		_ = os.Mkdir(metadataDir, os.ModePerm)
	}
	insightsDir := filepath.Join(moduleDir, InsightsDirExt)
	if _, err := os.Stat(insightsDir); os.IsNotExist(err) {
		_ = os.Mkdir(insightsDir, os.ModePerm)
	}

	if moduleDir == "" {
		return fmt.Errorf("unable to create development directory for module %s", eventID)
	}
	c.ModuleDir = moduleDir

	return nil
}

// WriteMetadata writes out a development file
func (c *Configuration) WriteMetadata(filename string, obj interface{}) error {
	err := c.writeOutput(filename, filepath.Join(c.ModuleDir, MetadataDirExt), obj)
	if err != nil {
		return fmt.Errorf("error writing development logs, '%+v'", err)
	}
	return nil
}

// WriteEvent writes out a development file
func (c *Configuration) WriteEvent(filename string, obj interface{}) error {
	err := c.writeOutput(filename, filepath.Join(c.ModuleDir, EventsDirExt), obj)
	if err != nil {
		return fmt.Errorf("error writing development logs, '%+v'", err)
	}
	return nil
}

// WriteBlob writes out a development file
func (c *Configuration) WriteBlob(filename string, obj interface{}) error {
	err := c.writeOutput(filename, filepath.Join(c.ModuleDir, BlobsDirExt), obj)
	if err != nil {
		return fmt.Errorf("error writing development logs, '%+v'", err)
	}
	return nil
}

// WriteInsight writes out a development file
func (c *Configuration) WriteInsight(filename string, obj interface{}) error {
	err := c.writeOutput(filename, filepath.Join(c.ModuleDir, InsightsDirExt), obj)
	if err != nil {
		return fmt.Errorf("error writing development logs, '%+v'", err)
	}
	return nil
}

// WriteOutput writes out a development file
func (c *Configuration) WriteOutput(filename string, obj interface{}) error {
	err := c.writeOutput(filename, c.ModuleDir, obj)
	if err != nil {
		return fmt.Errorf("error writing development logs, '%+v'", err)
	}
	return nil
}

func (c *Configuration) writeOutput(filename string, dir string, obj interface{}) error {
	// TODO: Handle errors here?
	path := filepath.Join(dir, filename)
	b, err := json.Marshal(&obj)
	if err != nil {
		return fmt.Errorf("error generating development logs, '%+v'", err)
	}
	err = ioutil.WriteFile(path, b, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing development logs, '%+v'", err)
	}
	return nil
}

// EventsDir gets the path to the module's events directory
func (c *Configuration) EventsDir() string {
	return filepath.FromSlash(filepath.Join(c.ModuleDir, EventsDirExt))
}

// BlobsDir gets the path to the module's blobs directory
func (c *Configuration) BlobsDir() string {
	return filepath.FromSlash(filepath.Join(c.ModuleDir, BlobsDirExt))
}

// MetadataDir gets the path to the module's metadata directory
func (c *Configuration) MetadataDir() string {
	return filepath.FromSlash(filepath.Join(c.ModuleDir, MetadataDirExt))
}

// InsightsDir gets the path to the module's insights directory
func (c *Configuration) InsightsDir() string {
	return filepath.FromSlash(filepath.Join(c.ModuleDir, InsightsDirExt))
}
