package types

const redacted = "****"

// Configuration for the application
// type Configuration struct {
// 	Hostname            string
// 	LogLevel            string            `short:"l" description:"Log level"`
// 	ModuleName          string            `description:"Name of the module"`
// 	SubscribesToEvent   string            `description:"Event this modules subscribes to"`
// 	EventsPublished     string            `description:"Events this modules can publish"`
// 	ServiceBusNamespace string            `description:"Namespace to use for ServiceBus"`
// 	ResourceGroup       string            `description:"Azure ResourceGroup to use"`
// 	SubscriptionID      string            `description:"SubscriptionID for Azure"`
// 	ClientID            string            `description:"ClientID of Service Principal for Azure access"`
// 	ClientSecret        string            `description:"Client Secrete of Service Principal for Azure access"`
// 	TenantID            string            `description:"TentantID for Azure"`
// 	LogSensitiveConfig  bool              `description:"Print out sensitive config when logging"`
// 	ModuleConfigPath    string            `description:"Path to environment variables file for module"`
// 	Kubernetes          *KubernetesConfig `description:"Configure k8s provider"`
// 	Job                 *JobConfig        `description:"Configure settings for the jobs to be run"`
// 	Sidecar             *SidecarConfig    `description:"Configure settings for the sidecar"`
// 	AzureBatch          *AzureBatchConfig `description:"Configure AzureBatch provider"`
// }
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
	Sidecar             *SidecarConfig    `yaml:"sidecar"`
	AzureBatch          *AzureBatchConfig `yaml:"azurebatch"`
}

// JobConfig configures the information about the jobs which will be run
// type JobConfig struct {
// 	MaxRunningTimeMins int    `description:"Max time a job can run for in mins"`
// 	RetryCount         int    `description:"Max number of times a job can be retried"`
// 	WorkerImage        string `description:"Image to use for the worker"`
// 	SidecarImage       string `description:"Image to use for the sidecar"`
// 	PullAlways         bool   `description:"Should docker images always be pulled"`
// }
type JobConfig struct {
	MaxRunningTimeMins int    `yaml:"maxrunningtimemins"`
	RetryCount         int    `yaml:"retrycount"`
	WorkerImage        string `yaml:"workerimage"`
	SidecarImage       string `yaml:"sidecarimage"`
	PullAlways         bool   `yaml:"pullalways"`
}

// SidecarConfig configures the information about the jobs which will be run
type SidecarConfig struct {
	ServerPort          int              `yaml:"serverport"`
	AzureBlobProvider   *AzureBlobConfig `yaml:"azureblobprovider"`
	MongoDBMetaProvider *MongoDBConfig   `yaml:"mongodbmetaprovider"`
	PrintConfig         bool             `yaml:"printconfig"`
}

// MongoDBConfig is configuration required to setup a MongoDB metadata store
// type MongoDBConfig struct {
// 	Name       string `description:"MongoDB database name"`
// 	Password   string `description:"MongoDB database password"`
// 	Collection string `description:"MongoDB database collection to use"`
// 	Port       int    `description:"MongoDB server port"`
// }
type MongoDBConfig struct {
	Name       string `yaml:"name"`
	Password   string `yaml:"password"`
	Collection string `yaml:"collection"`
	Port       int    `yaml:"port"`
}

// AzureBlobConfig is configuration required to setup a Azure Blob Store
// type AzureBlobConfig struct {
// 	BlobAccountName string `description:"Azure Blob Storage account name"`
// 	BlobAccountKey  string `description:"Azure Blob Storage account key"`
// 	UseProxy        bool   `description:"Enable proxy"`
// }
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
// type KubernetesConfig struct {
// 	Namespace           string `description:"The namespace in which jobs will be created"`
// 	ImagePullSecretName string `description:"~~Todo~~"`
// }
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
