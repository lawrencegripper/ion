package common

//KeyValuePair is a key value pair
type KeyValuePair struct {
	Key   string `bson:"key" json:"key"`
	Value string `bson:"value" json:"value"`
}

//Event the basic event data format
type Event struct {
	EventID        string         `json:"eventID"`
	Type           string         `json:"type"`
	PreviousStages []string       `json:"previousStages"`
	CorrelationID  string         `json:"correlationID"`
	Data           []KeyValuePair `json:"data"`
}
