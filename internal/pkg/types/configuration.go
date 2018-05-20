package types

const redacted = "****"

// Configuration for the application
type Configuration struct {
	Hostname            string            `yaml:"hostname"`
	LogLevel            string            `yaml:"loglevel"`
	ModuleName          string            `yaml:"modulename"`
	SubscribesToEvent   string            `yaml:"subscribestoevent"`
	EventsPublished     string            `yaml:"eventspublished"`
	ServiceBusNamespace string            `yaml:"servicebusnamespace"`
	ResourceGroup       string            `yaml:"resourcegroup"`
	SubscriptionID      string            `yaml:"subscriptionid"`
	ClientID            string            `yaml:"clientid"`
	ClientSecret        string            `yaml:"clientsecret"`
	TenantID            string            `yaml:"tenantid"`
	LogSensitiveConfig  bool              `yaml:"logsensitiveconfig"`
	ModuleConfigPath    string            `yaml:"moduleconfigpath"`
	PrintConfig         bool              `yaml:"printconfig"`
	Kubernetes          *KubernetesConfig `yaml:"kubernetes"`
	Job                 *JobConfig        `yaml:"job"`
	Handler             *HandlerConfig    `yaml:"handler"`
	AzureBatch          *AzureBatchConfig `yaml:"azurebatch"`
}

// JobConfig configures the information about the jobs which will be run
type JobConfig struct {
	MaxRunningTimeMins int    `yaml:"maxrunningtimemins"`
	RetryCount         int    `yaml:"retrycount"`
	WorkerImage        string `yaml:"workerimage"`
	HandlerImage       string `yaml:"handlerimage"`
	PullAlways         bool   `yaml:"pullalways"`
}

// HandlerConfig configures the information about the jobs which will be run
type HandlerConfig struct {
	ServerPort                     int              `yaml:"serverport"`
	AzureBlobStorageProvider       *AzureBlobConfig `yaml:"azureblobprovider"`
	MongoDBDocumentStorageProvider *MongoDBConfig   `yaml:"mongodbdocprovider"`
	PrintConfig                    bool             `yaml:"printconfig"`
}

// MongoDBConfig is configuration required to setup a MongoDB metadata store
type MongoDBConfig struct {
	Name       string `yaml:"name"`
	Password   string `yaml:"password"`
	Collection string `yaml:"collection"`
	Port       int    `yaml:"port"`
}

// AzureBlobConfig is configuration required to setup a Azure Blob Store
type AzureBlobConfig struct {
	BlobAccountName string `yaml:"blobaccountname"`
	BlobAccountKey  string `yaml:"blobaccountkey"`
	UseProxy        bool   `yaml:"useproxy"`
}

// AzureBatchConfig - Basic azure config used to interact with ARM resources.
type AzureBatchConfig struct {
	ResourceGroup           string `yaml:"resourcegroup"`
	PoolID                  string `yaml:"poolid"`
	JobID                   string `yaml:"jobid"`
	BatchAccountName        string `yaml:"batchaccountname"`
	BatchAccountLocation    string `yaml:"batchaccountlocation"`
	ImageRepositoryServer   string `yaml:"imagerepositoryserver"`
	ImageRepositoryUsername string `yaml:"imagerepositoryusername"`
	ImageRepositoryPassword string `yaml:"imagerepositorypassword"`
}

// KubernetesConfig - k8s config used to schedule jobs.
type KubernetesConfig struct {
	Namespace           string `yaml:"namespace"`
	ImagePullSecretName string `yaml:"imagepullsecretname"`
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
