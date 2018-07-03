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

//AsMap returns the key value pairs as a map
func (kvps KeyValuePairs) AsMap() map[string]string {
	result := make(map[string]string, len(kvps))
	for _, k := range kvps {
		result[k.Key] = k.Value
	}
	return result
}

//Append adds a new key value pair to the end of the slice
func (kvps KeyValuePairs) Append(kvp KeyValuePair) KeyValuePairs {
	return append(kvps, kvp)
}

//Remove a key value pair at an index by shifting the slice
func (kvps KeyValuePairs) Remove(index int) (KeyValuePairs, error) {
	if (index > len(kvps)+1) || (index < 0) {
		return KeyValuePairs{}, fmt.Errorf("Invalid index provided")
	}
	kvps = append(kvps[:index], kvps[index+1:]...)
	return kvps, nil
}

//Event the basic event data format
type Event struct {
	Context        *Context      `json:"context"`
	Type           string        `json:"type"`
	PreviousStages []string      `json:"previousStages"`
	Data           KeyValuePairs `json:"data"`
}

//Context carries the data for configuring the module
type Context struct {
	Name          string `description:"module name" bson:"name" json:"name"`
	EventID       string `description:"event identifier" bson:"eventId" json:"eventId"`
	CorrelationID string `description:"correlation identifier" bson:"correlationId" json:"correlationId"`
	ParentEventID string `description:"parent event identifier" bson:"parentEventId" json:"parentEventId"`
}
