package preparer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/lawrencegripper/ion/internal/app/sidecar/constants"
	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane"
	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane/metadata"
	"github.com/lawrencegripper/ion/internal/app/sidecar/helpers"
	"github.com/lawrencegripper/ion/internal/app/sidecar/logger"
	"github.com/lawrencegripper/ion/internal/app/sidecar/module"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	log "github.com/sirupsen/logrus"
)

// cSpell:ignore logrus, GUID, nolint

// Preparer holds the data and methods needed to prepare
// the module's environment.
type Preparer struct {
	dataPlane   *dataplane.DataPlane
	context     *common.Context
	environment *module.ModuleEnvironment

	baseDir         string
	developmentFlag bool
	logger          *log.Logger
}

// NewPreparer constructs a new preprarer
func NewPreparer(baseDir string, developmentFlag bool, logger *log.Logger) *Preparer {
	if baseDir == "" {
		baseDir = "/ion/"
	}

	preparer := &Preparer{
		baseDir:         baseDir,
		developmentFlag: developmentFlag,
		logger:          logger,
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
	if err := helpers.ErrorIfNil(dataPlane.BlobProvider, dataPlane.MetadataProvider, dataPlane.EventPublisher); err != nil {
		return err
	}
	if err := helpers.ErrorIfEmpty(context.EventID); err != nil {
		return err
	}

	p.context = context
	p.dataPlane = dataPlane

	p.environment = module.GetModuleEnvironment(p.baseDir)

	p.doPrepare()
	return nil
}

// Close cleans up the preparer
func (p *Preparer) Close() {
	logger.Info(p.logger, p.context, "Closing Preparer")

	defer p.dataPlane.Close()
}

// Prepare initializes the environment in which the module will run.
// This includes; creating the required directory structure and
// populating it with any input data.
func (p *Preparer) doPrepare() error {
	logger.Info(p.logger, p.context, "Preparing module environment")

	if err := p.prepareEnv(); err != nil {
		return err
	}
	if err := p.prepareData(); err != nil {
		return err
	}

	// If developmentFlag enabled, dump out an empty
	// file to indicate environment prepared.
	if p.developmentFlag {
		var empty struct{}
		_ = helpers.WriteDevFile("prepared", p.context.EventID, empty)
	}

	logger.Info(p.logger, p.context, "Successfully prepared module environment")
	return nil
}

// prepareEnv initializes the required directories
func (p *Preparer) prepareEnv() error {

	if err := p.environment.Build(); err != nil {
		return err
	}

	// If in developmentFlag mode, create the developmentFlag directories
	if p.developmentFlag {
		if _, err := os.Stat(constants.DevBaseDir); os.IsNotExist(err) {
			_ = os.Mkdir(constants.DevBaseDir, 0777)
		}
		devPath := path.Join(constants.DevBaseDir, p.context.EventID)
		_ = os.Mkdir(devPath, 0777)
	}
	return nil
}

func (p *Preparer) prepareData() error {

	context, err := p.getContext()
	if err != nil {
		return fmt.Errorf("Error fetching module's context %+v", err)
	}

	// Only get files for events with an existing context.
	// Assume those that don't have a context are the
	// first event in the graph or orphaned.
	if context != nil {
		err = p.dataPlane.GetBlobs(p.environment.InputBlobDirPath, context.Files)
		if err != nil {
			return err
		}

		if len(context.Data) > 0 {
			b, err := json.Marshal(context.Data)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(p.environment.InputMetaFilePath, b, 0777)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Preparer) getContext() (*metadata.EventContext, error) {
	context, _ := p.dataPlane.GetEventContextByID(p.context.EventID)
	//TODO: Fail on error conditions other than not found
	return context, nil
}
