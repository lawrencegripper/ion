package management

// Configuration reprsents all the config values
// needed to run the management server
type Configuration struct {
	Port                      int
	Namespace                 string
	DispatcherImage           string
	DispatcherImageTag        string
	AzureClientID             string
	AzureClientSecret         string
	AzureSubscriptionID       string
	AzureTenantID             string
	AzureServiceBusNamespace  string
	AzureResourceGroup        string
	AzureBatchPoolID          string
	AzureADResource           string
	AzureBatchAccountLocation string
	AzureBatchAccountName     string
	MongoDBPort               int
	MongoDBName               string
	MongoDBPassword           string
	MongoDBCollection         string
	AzureStorageAccountName   string
	AzureStorageAccountKey    string
	LogLevel                  string
	PrintConfig               bool
}

// NewConfiguration create an empty config
func NewConfiguration() Configuration {
	cfg := Configuration{}
	return cfg
}
