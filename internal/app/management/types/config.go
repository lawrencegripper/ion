package types

// Configuration reprsents all the config values
// needed to run the management server
type Configuration struct {
	Provider                          string
	Port                              int
	Namespace                         string
	DispatcherImage                   string
	DispatcherImageTag                string
	ContainerImageRegistryURL         string
	ContainerImageRegistryUsername    string
	ContainerImageRegistryPassword    string
	ContainerImageRegistryEmail       string
	AzureClientID                     string
	AzureClientSecret                 string
	AzureSubscriptionID               string
	AzureTenantID                     string
	AzureServiceBusNamespace          string
	AzureResourceGroup                string
	AzureBatchJobID                   string
	AzureBatchPoolID                  string
	AzureBatchAccountLocation         string
	AzureBatchAccountName             string
	AzureBatchRequiresGPU             bool
	AzureBatchResourceGroup           string
	AzureBatchImageRepositoryServer   string
	AzureBatchImageRepositoryUsername string
	AzureBatchImageRepositoryPassword string
	MongoDBPort                       int
	MongoDBName                       string
	MongoDBPassword                   string
	MongoDBCollection                 string
	AzureStorageAccountName           string
	AzureStorageAccountKey            string
	LogLevel                          string
	PrintConfig                       bool
}

// NewConfiguration create an empty config
func NewConfiguration() Configuration {
	cfg := Configuration{}
	return cfg
}
