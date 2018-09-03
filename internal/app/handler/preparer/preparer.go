package preparer

import (
	"encoding/json"
	"fmt"
	"github.com/lawrencegripper/ion/internal/app/handler/development"
	"io/ioutil"
	"os"

	"github.com/lawrencegripper/ion/internal/app/handler/dataplane"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage"
	"github.com/lawrencegripper/ion/internal/app/handler/helpers"
	"github.com/lawrencegripper/ion/internal/app/handler/logger"
	"github.com/lawrencegripper/ion/internal/app/handler/module"
	"github.com/lawrencegripper/ion/internal/pkg/common"
)

// cSpell:ignore logrus, GUID, nolint

// Preparer holds the data and methods needed to prepare
// the module's environment.
type Preparer struct {
	dataPlane   *dataplane.DataPlane
	context     *common.Context
	environment *module.Environment

	baseDir   string
	devConfig *development.Configuration
}

// NewPreparer constructs a new preprarer
func NewPreparer(baseDir string, devCfg *development.Configuration) *Preparer {
	if baseDir == "" {
		baseDir = "/ion/"
	}

	if devCfg == nil {
		devCfg = &development.Configuration{}
	}

	preparer := &Preparer{
		baseDir:   baseDir,
		devConfig: devCfg,
	}

	return preparer
}

// Prepare is the entry point for the Preparer.
func (p *Preparer) Prepare(
	context *common.Context,
	dataPlane *dataplane.DataPlane) error {

	if err := helpers.ErrorIfNil(dataPlane, context); err != nil {
		return err
	}
	if err := helpers.ErrorIfNil(dataPlane.BlobStorageProvider, dataPlane.DocumentStorageProvider, dataPlane.EventPublisher); err != nil {
		return err
	}
	if err := helpers.ErrorIfEmpty(context.EventID); err != nil {
		return err
	}

	p.context = context
	p.dataPlane = dataPlane

	p.environment = module.GetModuleEnvironment(p.baseDir)

	err := p.doPrepare()
	if err != nil {
		return err
	}
	return nil
}

// Close cleans up the preparer
func (p *Preparer) Close() {
	logger.Info(p.context, "closing Preparer")

	defer p.dataPlane.Close()
}

// Prepare initializes the environment in which the module will run.
// This includes; creating the required directory structure and
// populating it with any input data.
func (p *Preparer) doPrepare() error {
	logger.Info(p.context, "preparing module environment")

	if err := p.prepareEnv(); err != nil {
		return err
	}
	if err := p.prepareData(); err != nil {
		return err
	}

	// If development enabled, dump out an empty
	// file to indicate environment prepared.
	if p.devConfig.Enabled {
		var empty struct{}
		_ = p.devConfig.WriteOutput("prepared", empty)
	}

	logger.Info(p.context, "successfully prepared module environment")
	return nil
}

// prepareEnv initializes the required directories
func (p *Preparer) prepareEnv() error {

	if err := p.environment.Build(); err != nil {
		return err
	}

	// If in development enabled, make sure the development directories exist
	if p.devConfig.Enabled {
		if _, err := os.Stat(p.devConfig.ModuleDir); os.IsNotExist(err) {
			_ = os.MkdirAll(p.devConfig.ModuleDir, os.ModePerm)
		}
	}
	return nil
}

func (p *Preparer) prepareData() error {

	eventMeta, err := p.getEventMeta()
	if err != nil {
		return fmt.Errorf("Error fetching module's context %+v", err)
	}

	// Only get files for events with an existing context.
	// Assume those that don't have a context are the
	// first event in the graph or orphaned.
	if eventMeta != nil {
		logger.InfoWithFields(p.context, "getting blobs for files", map[string]interface{}{
			"files": eventMeta.Files,
			"data":  eventMeta.Data,
		})
		err = p.dataPlane.GetBlobs(p.environment.InputBlobDirPath, eventMeta.Files)
		if err != nil {
			return err
		}

		if len(eventMeta.Data) > 0 {
			b, err := json.Marshal(eventMeta.Data)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(p.environment.InputMetaFilePath, b, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Preparer) getEventMeta() (*documentstorage.EventMeta, error) {
	context, _ := p.dataPlane.GetEventMetaByID(p.context.EventID)
	//TODO: Fail on error conditions other than not found
	return context, nil
}
