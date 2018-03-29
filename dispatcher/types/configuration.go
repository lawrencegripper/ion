package types

const redacted = "****"

// Configuration for the application
type Configuration struct {
	Hostname            string
	LogLevel            string         `short:"l" description:"Log level"`
	ModuleName          string         `description:"Name of the module"`
	SubscribesToEvent   string         `description:"Event this modules subscribes to"`
	EventsPublished     string         `description:"Events this modules can publish"`
	ServiceBusNamespace string         `description:"Namespace to use for ServiceBus"`
	ResourceGroup       string         `description:"Azure ResourceGroup to use"`
	SubscriptionID      string         `description:"SubscriptionID for Azure"`
	ClientID            string         `description:"ClientID of Service Principal for Azure access"`
	ClientSecret        string         `description:"Client Secrete of Service Principal for Azure access"`
	TenantID            string         `description:"TentantID for Azure"`
	LogSensitiveConfig  bool           `description:"Print out sensitive config when logging"`
	Job                 *JobConfig     `description:"Configure settings for the jobs to be run"`
	Sidecar             *SidecarConfig `description:"Configure settings for the sidecar"`
}

// JobConfig configures the information about the jobs which will be run
type JobConfig struct {
	MaxRunningTimeMins int    `description:"Max time a job can run for in mins"`
	RetryCount         int    `description:"Max number of times a job can be retried"`
	WorkerImage        string `description:"Image to use for the worker"`
	SidecarImage       string `description:"Image to use for the sidecar"`
}

// SidecarConfig configures the information about the jobs which will be run
type SidecarConfig struct {
	ServerPort          int              `description:"~~Todo~~"`
	AzureBlobProvider   *AzureBlobConfig `description:"~~Todo~~"`
	MongoDBMetaProvider *MongoDBConfig   `description:"~~Todo~~"`
	PrintConfig         bool             `description:"~~Todo~~"`
}

// MongoDBConfig is configuration required to setup a MongoDB metadata store
type MongoDBConfig struct {
	Name       string `description:"MongoDB database name"`
	Password   string `description:"MongoDB database password"`
	Collection string `description:"MongoDB database collection to use"`
	Port       int    `description:"MongoDB server port"`
}

// AzureBlobConfig is configuration required to setup a Azure Blob Store
type AzureBlobConfig struct {
	BlobAccountName string `description:"Azure Blob Storage account name"`
	BlobAccountKey  string `description:"Azure Blob Storage account key"`
	UseProxy        bool   `description:"Enable proxy"`
}

// RedactConfigSecrets strips sensitive data from the config
func RedactConfigSecrets(config *Configuration) Configuration {
	c := *config
	if !c.LogSensitiveConfig {
		c.ClientID = redacted
		c.ClientSecret = redacted
		c.TenantID = redacted
		c.SubscriptionID = redacted
	}
	return c
}
