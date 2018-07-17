package handler

import (
	"encoding/json"
	"fmt"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	"github.com/lawrencegripper/ion/modules/helpers/Go/env"
	"github.com/lawrencegripper/ion/modules/helpers/Go/log"

	"io/ioutil"
)

//ReadEventMetaData return the event metadata for the event which triggered the module
// this will commonly contain useful information from the parent event
func ReadEventMetaData() (*common.KeyValuePairs, error) {
	dat, err := ioutil.ReadFile(env.InputEventMetaFile())
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to read eventmeta input file from %s", env.InputEventMetaFile()))
		return nil, err
	}

	eventMeta := common.KeyValuePairs{}
	err = json.Unmarshal(dat, &eventMeta)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to deserialize eventmeta input file form %s with body: %s with error: %v", env.InputEventMetaFile(), string(dat), err))
		return nil, err
	}

	return &eventMeta, nil
}
