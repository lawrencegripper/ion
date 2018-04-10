package common

import (
	"fmt"
)

//KeyValuePair is a key value pair
type KeyValuePair struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value" json:"value"`
}

//KeyValuePairs is a named type for a slice of key value pairs
type KeyValuePairs []KeyValuePair

//Append adds a new key value pair to the end of the slice
func (kvps KeyValuePairs) Append(kvp KeyValuePair) {
	kvps = append(kvps, kvp)
}

//Remove a key value pair at an index by shifting the slice
func (kvps KeyValuePairs) Remove(index int) error {
	if (index > len(kvps)+1) || (index < 0) {
		return fmt.Errorf("Invalid index provided")
	}
	kvps = append(kvps[:index], kvps[index+1:]...)
	return nil
}

//Event the basic event data format
type Event struct {
	EventID        string        `json:"eventID"`
	Type           string        `json:"type"`
	PreviousStages []string      `json:"previousStages"`
	CorrelationID  string        `json:"correlationID"`
	Data           KeyValuePairs `json:"data"`
}
