package filesystem

import (
	"fmt"
	"io"
	"os"
	"path"
)

//Config to setup a FileSystem storage provider
type Config struct {
	BaseDir string `description:"Base directory used to persist files"`
}

//FileSystemStorage stores blobs on local disk
type FileSystemStorage struct {
	baseDir string
}

//NewFileSystemStorage creates a new file system blob provider
func NewFileSystemStorage(config *Config) (*FileSystemStorage, error) {
	err := os.MkdirAll(config.BaseDir, 0777)
	if err != nil {
		return nil, fmt.Errorf("error creating directory for filesystem blob provider '%+v'", err)
	}
	fs := &FileSystemStorage{
		baseDir: config.BaseDir,
	}
	return fs, nil
}

//PutBlobs puts in the filesystem directory
func (a *FileSystemStorage) PutBlobs(filePaths []string) (map[string]string, error) {
	uris := make(map[string]string)
	for _, filePath := range filePaths {
		_, nakedFilePath := path.Split(filePath)
		destPath := path.Join(a.baseDir, nakedFilePath)
		if err := copy(filePath, destPath); err != nil {
			return nil, fmt.Errorf("error copying file to blob storage '%+v'", err)
		}
		uris[filePath] = destPath
	}
	return uris, nil
}

//GetBlobs gets each of the referenced blobs from the file system
func (a *FileSystemStorage) GetBlobs(outputDir string, filePaths []string) error {
	for _, file := range filePaths {
		srcPath := path.Join(a.baseDir, file)
		_, err := os.Stat(srcPath)
		if err != nil {
			return fmt.Errorf("error getting blob '%s': '%+v'", file, err)
		}
		destPath := path.Join(outputDir, file)
		if err := copy(srcPath, destPath); err != nil {
			return fmt.Errorf("error copying from blob '%s': '%+v'", file, err)
		}
	}
	return nil
}

//Close cleans up any external resources
func (a *FileSystemStorage) Close() {
}

//copy a file from a source path to a destination path
func copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
