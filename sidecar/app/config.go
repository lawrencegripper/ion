package app

//Configuration represents the input configuration schema
type Configuration struct {
	SharedSecret            string `description:"A shared secret to authenticate client requests with"`
	BlobStorageAccessKey    string `description:"A access token for an external blob storage provider"`
	BlobStorageAccountName  string `description:"Blob storage account name"`
	DBName                  string `description:"The name of database to store metadata"`
	DBPassword              string `description:"The password to access the metadata database"`
	DBCollection            string `description:"The document collection name on the metadata database"`
	DBPort                  int    `description:"The database port"`
	PublisherName           string `description:"The name or namespace for the publisher"`
	PublisherTopic          string `description:"The topic to publish events on"`
	PublisherAccessKey      string `description:"An access key for the publisher"`
	PublisherAccessRuleName string `description:"The rule name associated with the given access key"`
	LogFile                 string `description:"File to log output to"`
	LogLevel                string `description:"Logging level, possible values {debug, info, warn, error}"`
	EventID                 string `description:"The unique ID for this module"`
	ParentEventID           string `description:"Previous event ID"`
	CorrelationID           string `description:"CorrelationID used to correlate this module with others"`
	ServerPort              int    `description:"The port for the web server to listen on"`
}
