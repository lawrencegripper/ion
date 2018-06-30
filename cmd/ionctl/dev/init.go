package dev

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/twinj/uuid"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type initOptions struct {
	files []string
}

var initOpts initOptions

// initCmd represents the fire command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialise an origin module",
	RunE:  Init,
}

// Init a new origin module used to bootstrap a development flow
func Init(cmd *cobra.Command, args []string) error {

	moduleID := fmt.Sprintf("%v", uuid.NewV4())

	moduleDir := filepath.FromSlash(filepath.Join(ionModulesDir, moduleID))
	_, err := os.Stat(moduleDir)
	if err != nil {
		if err := os.MkdirAll(moduleDir, 0777); err != nil {
			return fmt.Errorf("error creating module directory %s: %+v", moduleDir, err)
		}
	}
	inputBlobDir := filepath.FromSlash(filepath.Join(moduleDir, "_blobs"))
	_, err = os.Stat(inputBlobDir)
	if err != nil {
		if err := os.MkdirAll(inputBlobDir, 0777); err != nil {
			return fmt.Errorf("error creating module blob directory %s: %+v", inputBlobDir, err)
		}
	}
	inputEventDir := filepath.FromSlash(filepath.Join(moduleDir, "_events"))
	_, err = os.Stat(inputEventDir)
	if err != nil {
		if err := os.MkdirAll(inputEventDir, 0777); err != nil {
			return fmt.Errorf("error creating module event directory %s: %+v", inputEventDir, err)
		}
	}

	for _, file := range initOpts.files {
		filename := filepath.Base(file)
		dest := filepath.FromSlash(filepath.Join(inputBlobDir, filename))
		if err := copyFile(file, dest); err != nil {
			return fmt.Errorf("error copying file %s to destination %s: %+v", file, dest, err)
		}
	}

	fmt.Printf("new parent event id: %s\n", moduleID)

	return nil
}

func init() {

	// Local flags for the init command
	initCmd.Flags().StringSliceVar(&initOpts.files, "files", []string{}, "module input files")

	// Mark required flags
	initCmd.MarkPersistentFlagRequired("files") //nolint: errcheck
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
