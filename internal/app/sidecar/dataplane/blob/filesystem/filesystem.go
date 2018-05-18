package filesystem

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

//Config to setup a FileSystem storage provider
type Config struct {
	InputDir  string `description:"Input directory used to persist files"`
	OutputDir string `description:"Output directory used to persist files"`
}

//BlobStorage stores blobs on local disk
type BlobStorage struct {
	inDir  string
	outDir string
}

//NewBlobStorage creates a new file system blob provider
func NewBlobStorage(config *Config) (*BlobStorage, error) {
	err := os.MkdirAll(config.OutputDir, 0777)
	if err != nil {
		return nil, fmt.Errorf("error creating directory for filesystem blob provider '%+v'", err)
	}
	fs := &BlobStorage{
		inDir:  config.InputDir,
		outDir: config.OutputDir,
	}
	return fs, nil
}

//PutBlobs puts in the filesystem directory
func (a *BlobStorage) PutBlobs(filePaths []string) (map[string]string, error) {
	uris := make(map[string]string)
	for _, filePath := range filePaths {
		_, nakedFilePath := filepath.Split(filePath)
		destPath := filepath.FromSlash(path.Join(a.outDir, nakedFilePath))
		if err := copy(filePath, destPath); err != nil {
			return nil, fmt.Errorf("error copying file to blob storage '%+v'", err)
		}
		uris[filePath] = destPath
	}
	return uris, nil
}

//GetBlobs gets each of the referenced blobs from the file system
func (a *BlobStorage) GetBlobs(outputDir string, filePaths []string) error {
	for _, file := range filePaths {
		srcPath := filepath.FromSlash(path.Join(a.inDir, file))
		_, err := os.Stat(srcPath)
		if err != nil {
			return fmt.Errorf("error getting blob '%s': '%+v'", file, err)
		}
		destPath := filepath.FromSlash(path.Join(outputDir, file))
		if err := copy(srcPath, destPath); err != nil {
			return fmt.Errorf("error copying from blob '%s': '%+v'", file, err)
		}
	}
	return nil
}

//Close cleans up any external resources
func (a *BlobStorage) Close() {
}

//copy a file from a source path to a destination path
func copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close() //nolint:errcheck

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close() //nolint:errcheck

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close() //nolint:errcheck
}
