package filesystem

import (
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/lawrencegripper/ion/sidecar/types"
)

//Config to setup a FileSystem blob provider
type Config struct {
	BaseDir string `description:"Base directory to use store blob data"`
}

//FileSystem is a local file system based implementation of a blob provider
type FileSystem struct {
	baseDir string
}

//NewFileSystemBlobProvider creates a new FileSystem blob provider object
func NewFileSystemBlobProvider(config *Config) *FileSystem {
	baseDir := config.BaseDir
	_ = os.Mkdir(baseDir, 0777)
	return &FileSystem{
		baseDir: baseDir,
	}
}

//Proxy used by clients to check whether a proxy is available
func (f *FileSystem) Proxy() types.BlobProxy {
	return nil
}

//Create will create the necessary storage location and file
func (f *FileSystem) Create(resourcePath string, blob io.ReadCloser) (string, error) {
	bytes, err := ioutil.ReadAll(blob)
	if err != nil {
		return "", err
	}
	dir, _ := path.Split(resourcePath)
	dirPath := path.Join(f.baseDir, dir)
	err = createDirIfNotExist(dirPath)
	if err != nil {
		return "", err
	}
	path := path.Join(f.baseDir, resourcePath)
	err = ioutil.WriteFile(path, bytes, 0777)
	if err != nil {
		return "", err
	}
	return path, nil
}

//Get will attempt to retrieve a blob given a resource path
func (f *FileSystem) Get(resourcePath string) (io.ReadCloser, error) {
	path := path.Join(f.baseDir, resourcePath)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

//List will list the contents of a directory
func (f *FileSystem) List(resourcePath string) ([]string, error) {
	path := path.Join(f.baseDir, resourcePath)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var fileNames []string
	for _, f := range files {
		fileNames = append(fileNames, f.Name())
	}
	return fileNames, nil
}

//Delete will delete a blob resource given a resource path
func (f *FileSystem) Delete(resourcePath string) (bool, error) {
	path := path.Join(f.baseDir, resourcePath)
	err := os.Remove(path)
	if err != nil {
		return false, err
	}
	return true, nil
}

//Close cleans up any external resources
func (f *FileSystem) Close() {
}

func createDirIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(path, 0777)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
