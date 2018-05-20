package module

import (
	"fmt"
	"github.com/lawrencegripper/ion/internal/app/sidecar/constants"
	"github.com/lawrencegripper/ion/internal/app/sidecar/helpers"
)

// cSpell:ignore bson

// ModuleEnvironment represents the directory structure in
// which the module operates
type ModuleEnvironment struct {
	InputBlobDirPath  string
	InputMetaFilePath string

	OutputBlobDirPath   string
	OutputMetaFilePath  string
	OutputEventsDirPath string
}

// GetModuleEnvironment returns a struct that represents
// the require directory structure for the module
// environment.
func GetModuleEnvironment(baseDir string) *ModuleEnvironment {
	return &ModuleEnvironment{
		InputBlobDirPath:  helpers.GetPath(baseDir, constants.InputBlobDir),
		InputMetaFilePath: helpers.GetPath(baseDir, constants.InputEventMetaFile),

		OutputBlobDirPath:   helpers.GetPath(baseDir, constants.OutputBlobDir),
		OutputMetaFilePath:  helpers.GetPath(baseDir, constants.OutputInsightsFile),
		OutputEventsDirPath: helpers.GetPath(baseDir, constants.OutputEventsDir),
	}
}

// Build creates a clean directory structure for the module
func (m *ModuleEnvironment) Build() error {
	if err := helpers.CreateDirClean(m.InputBlobDirPath); err != nil {
		return fmt.Errorf("could not create input blob directory, %+v", err)
	}
	if err := helpers.CreateDirClean(m.OutputBlobDirPath); err != nil {
		return fmt.Errorf("could not create output blob directory, %+v", err)
	}
	if err := helpers.CreateFileClean(m.OutputMetaFilePath); err != nil {
		return fmt.Errorf("could not create output meta file, %+v", err)
	}
	if err := helpers.CreateDirClean(m.OutputEventsDirPath); err != nil {
		return fmt.Errorf("could not create output events directory, %+v", err)
	}
	return nil
}

// Clear will clean down the module's directory structure
func (m *ModuleEnvironment) Clear() error {
	if err := helpers.ClearDir(m.InputBlobDirPath); err != nil {
		return fmt.Errorf("could not create input blob directory, %+v", err)
	}
	if err := helpers.ClearDir(m.OutputBlobDirPath); err != nil {
		return fmt.Errorf("could not create output blob directory, %+v", err)
	}
	if err := helpers.RemoveFile(m.OutputMetaFilePath); err != nil {
		return fmt.Errorf("could not create output meta file, %+v", err)
	}
	if err := helpers.ClearDir(m.OutputEventsDirPath); err != nil {
		return fmt.Errorf("could not create output events directory, %+v", err)
	}
	return nil
}
