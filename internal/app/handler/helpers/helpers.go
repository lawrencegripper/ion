package helpers

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/twinj/uuid"
	"os"
)

// cSpell:ignore twinj, uuid, nolint, strs, objs, GUID

//ErrorIfEmpty errors if any of the strings provided are empty
func ErrorIfEmpty(strs ...string) error {
	for _, s := range strs {
		if s == "" {
			return fmt.Errorf("required string is empty")
		}
	}
	return nil
}

//ErrorIfNil errors if any of the objects provided are nil
func ErrorIfNil(objs ...interface{}) error {
	for _, o := range objs {
		if o == nil {
			return fmt.Errorf("required obj is nil")
		}
	}
	return nil
}

//ClearDir removes all the content from a directory
func ClearDir(dirPath string) error {
	err := os.RemoveAll(dirPath)
	if err != nil {
		return fmt.Errorf("failed removing directory path '%s' with error: '%+v'", dirPath, err)
	}
	err = os.MkdirAll(dirPath, 0777)
	if err != nil {
		return fmt.Errorf("failed creating directory path '%s' with error: '%+v'", dirPath, err)
	}
	return nil
}

//RemoveFile removes a file from the file system
func RemoveFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // If file doesn't exist - return as if succeeded
	}
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to remove file at path '%s' with error: '%+v'", filePath, err)
	}
	return nil
}

//NewGUID generates a new guid as a string
func NewGUID() string {
	guid := fmt.Sprintf("%v", uuid.NewV4())
	return guid
}

//JoinBlobPath returns a formatted blob path
func JoinBlobPath(strs ...string) string {
	var allStrs []string
	for _, s := range strs {
		allStrs = append(allStrs, s)
	}
	return strings.Join(allStrs, `/`)
}

//ContainsString checks whether a string is in a slice of strings
func ContainsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

//CreateDirClean creates a directory - deleting any existing directory
func CreateDirClean(dirPath string) error {
	_ = os.RemoveAll(dirPath)
	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return fmt.Errorf("error creating directory '%s': '%+v'", dirPath, err)
	}
	return nil
}

//CreateFileClean creates a file - deleting any existing file
func CreateFileClean(filePath string) error {
	_ = os.Remove(filePath)

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file '%s': '%+v'", filePath, err)
	}
	defer f.Close() // nolint: errcheck
	return nil
}

// GetPath creates an OS specific filepath including the base directory
func GetPath(basePath, relativePath string) string {
	return filepath.FromSlash(path.Join(basePath, relativePath))
}
