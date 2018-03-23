package messaging

// Event the basic event data format
type Event struct {
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	PreviousStages []string          `json:"previousStages"`
	ParentEventID  string            `json:"parentId"`
	CorrelationID  string            `json:"correlationId"`
	Data           map[string]string `json:"data"`
}
