package main

import (
	"errors"
	"os"
	"strings"

	"github.com/lawrencegripper/ion/internal/pkg/tools"
	"github.com/lawrencegripper/ion/internal/pkg/types"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var errConfigurationMissing = errors.New("missing configuration values, use '--printconfig' to show current config on start")

var cfg = types.Configuration{
	Kubernetes: &types.KubernetesConfig{},
	Job:        &types.JobConfig{},
	Handler: &types.HandlerConfig{
		AzureBlobStorageProvider:       &types.AzureBlobConfig{},
		MongoDBDocumentStorageProvider: &types.MongoDBConfig{},
	},
	AzureBatch: &types.AzureBatchConfig{},
}
var cfgFile string

// NewDispatcherCommand return cobra.Command to run ion-disptacher commands
func NewDispatcherCommand() *cobra.Command {
	dispatcherCmd := &cobra.Command{
		Use:   "dispatcher",
		Short: "dispatcher: ...",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Read config file
			viper.SetConfigFile(cfgFile)
			if err := viper.ReadInConfig(); err != nil {
				log.WithError(err).Warningln("Can't read config")
			}
			viper.AutomaticEnv()

			// Fill config with global settings
			cfg.LogLevel = viper.GetString("loglevel")
			cfg.ModuleName = viper.GetString("modulename")
			cfg.SubscribesToEvent = viper.GetString("subscribestoevent")
			cfg.EventsPublished = viper.GetString("eventspublished")
			cfg.ServiceBusNamespace = viper.GetString("servicebusnamespace")
			cfg.ResourceGroup = viper.GetString("resourcegroup")
			cfg.PrintConfig = viper.GetBool("printconfig")
			cfg.ModuleConfigPath = viper.GetString("moduleconfigpath")
			cfg.LogSensitiveConfig = viper.GetBool("logsensitiveconfig")
			// kubernetes.*
			cfg.Kubernetes.Namespace = viper.GetString("kubernetes.namespace")
			cfg.Kubernetes.ImagePullSecretName = viper.GetString("kubernetes.imagepullsecretname")
			// job.*
			cfg.Job.MaxRunningTimeMins = viper.GetInt("job.maxrunningtimemins")
			cfg.Job.RetryCount = viper.GetInt("job.retrycount")
			cfg.Job.WorkerImage = viper.GetString("job.workerimage")
			cfg.Job.HandlerImage = viper.GetString("job.handlerimage")
			cfg.Job.PullAlways = viper.GetBool("job.pullalways")
			// handler.*
			cfg.Handler.ServerPort = viper.GetInt("handler.serverport")
			cfg.Handler.PrintConfig = viper.GetBool("handler.printconfig")
			// handler.azureblobprovider.*
			cfg.Handler.AzureBlobStorageProvider.BlobAccountName = viper.GetString("handler.azureblobprovider.blobaccountname")
			cfg.Handler.AzureBlobStorageProvider.BlobAccountKey = viper.GetString("handler.azureblobprovider.blobaccountkey")
			cfg.Handler.AzureBlobStorageProvider.UseProxy = viper.GetBool("handler.azureblobprovider.useproxy")
			// handler.mongodbdocprovider.*
			cfg.Handler.MongoDBDocumentStorageProvider.Name = viper.GetString("handler.mongodbdocprovider.name")
			cfg.Handler.MongoDBDocumentStorageProvider.Password = viper.GetString("handler.mongodbdocprovider.password")
			cfg.Handler.MongoDBDocumentStorageProvider.Collection = viper.GetString("handler.mongodbdocprovider.collection")
			cfg.Handler.MongoDBDocumentStorageProvider.Port = viper.GetInt("handler.mongodbdocprovider.port")
			// azurebatch.*
			cfg.AzureBatch.RequiresGPU = viper.GetBool("azurebatch.requiresgpu")
			cfg.AzureBatch.ResourceGroup = viper.GetString("azurebatch.resourcegroup")
			cfg.AzureBatch.PoolID = viper.GetString("azurebatch.poolid")
			cfg.AzureBatch.JobID = viper.GetString("azurebatch.jobid")
			cfg.AzureBatch.BatchAccountName = viper.GetString("azurebatch.batchaccountname")
			cfg.AzureBatch.BatchAccountLocation = viper.GetString("azurebatch.batchaccountlocation")
			cfg.AzureBatch.ImageRepositoryServer = viper.GetString("azurebatch.imagerepositoryserver")
			cfg.AzureBatch.ImageRepositoryUsername = viper.GetString("azurebatch.imagerepositoryusername")
			cfg.AzureBatch.ImageRepositoryPassword = viper.GetString("azurebatch.imagerepositorypassword")

			// Globally set configuration level
			switch strings.ToLower(cfg.LogLevel) {
			case "debug":
				log.SetLevel(log.DebugLevel)
			case "info":
				log.SetLevel(log.InfoLevel)
			case "warn":
				log.SetLevel(log.WarnLevel)
			case "error":
				log.SetLevel(log.ErrorLevel)
			default:
				log.SetLevel(log.WarnLevel)
			}

			hostName, err := os.Hostname()
			if err != nil {
				return errors.New("Unable to automatically set instanceid to hostname")
			}
			cfg.Hostname = hostName

			return nil
		},
	}

	// Add 'dispatcher' flags
	dispatcherCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "../../configs/dispatcher.yaml", "Config file path")
	dispatcherCmd.PersistentFlags().StringP("loglevel", "l", "warn", "Log level (debug|info|warn|error)")
	dispatcherCmd.PersistentFlags().String("modulename", "", "Name of the module")
	dispatcherCmd.PersistentFlags().String("subscribestoevent", "", "Event this modules subscribes to")
	dispatcherCmd.PersistentFlags().String("eventspublished", "", "Events this modules can publish")
	dispatcherCmd.PersistentFlags().String("servicebusnamespace", "", "Namespace to use for ServiceBus")
	dispatcherCmd.PersistentFlags().String("resourcegroup", "", "Azure ResourceGroup to use")
	dispatcherCmd.PersistentFlags().Bool("logsensitiveconfig", false, "Print out sensitive config when logging")
	dispatcherCmd.PersistentFlags().String("moduleconfigpath", "", "Path to environment variables file for module")
	dispatcherCmd.PersistentFlags().BoolP("printconfig", "P", false, "Print out config when starting")
	// kubernetes.*
	dispatcherCmd.PersistentFlags().String("kubernetes.namespace", "default", "The Kubernetes namespace in which jobs will be created")
	dispatcherCmd.PersistentFlags().String("kubernetes.imagepullsecretname", "", "")
	// job.*
	dispatcherCmd.PersistentFlags().Int("job.maxrunningtimemins", 10, "Max time a job can run for in mins")
	dispatcherCmd.PersistentFlags().Int("job.retrycount", 0, "Max number of times a job can be retried")
	dispatcherCmd.PersistentFlags().String("job.workerimage", "", "Image to use for the worker")
	dispatcherCmd.PersistentFlags().String("job.handlerimage", "", "Image to use for the handler")
	dispatcherCmd.PersistentFlags().Bool("job.pullalways", true, "Should docker images always be pulled")
	// handler.*
	dispatcherCmd.PersistentFlags().Int("handler.serverport", 8080, "")
	dispatcherCmd.PersistentFlags().Bool("handler.printconfig", false, "Print out config when starting")
	// handler.azureblobprovider.*
	dispatcherCmd.PersistentFlags().String("handler.azureblobprovider.blobaccountname", "", "Azure Blob Storage account name")
	dispatcherCmd.PersistentFlags().String("handler.azureblobprovider.blobaccountkey", "", "Azure Blob Storage account key")
	dispatcherCmd.PersistentFlags().Bool("handler.azureblobprovider.useproxy", false, "Enable proxy")
	// handler.mongodbdocprovider.*
	dispatcherCmd.PersistentFlags().String("handler.mongodbdocprovider.name", "", "MongoDB database name")
	dispatcherCmd.PersistentFlags().String("handler.mongodbdocprovider.password", "", "MongoDB database password")
	dispatcherCmd.PersistentFlags().String("handler.mongodbdocprovider.collection", "", "MongoDB database collection to use")
	dispatcherCmd.PersistentFlags().Int("handler.mongodbdocprovider.port", 27017, "MongoDB server port")
	// azurebatch.*
	dispatcherCmd.PersistentFlags().Bool("azurebatch.requiresgpu", false, "Module requries gpu")
	dispatcherCmd.PersistentFlags().String("azurebatch.resourcegroup", "", "")
	dispatcherCmd.PersistentFlags().String("azurebatch.poolid", "", "")
	dispatcherCmd.PersistentFlags().String("azurebatch.jobid", "", "")
	dispatcherCmd.PersistentFlags().String("azurebatch.batchaccountname", "", "")
	dispatcherCmd.PersistentFlags().String("azurebatch.batchaccountlocation", "", "")
	dispatcherCmd.PersistentFlags().String("azurebatch.imagerepositoryserver", "", "")
	dispatcherCmd.PersistentFlags().String("azurebatch.imagerepositoryusername", "", "")
	dispatcherCmd.PersistentFlags().String("azurebatch.imagerepositorypassword", "", "")

	// Mark required flags (won't mark required setting, onyl CLI flag presence will be checked)
	//dispatcherCmd.MarkPersistentFlagRequired("")

	// Bind flags and config file values
	viper.BindPFlag("loglevel", dispatcherCmd.PersistentFlags().Lookup("loglevel"))
	viper.BindPFlag("modulename", dispatcherCmd.PersistentFlags().Lookup("modulename"))
	viper.BindPFlag("subscribestoevent", dispatcherCmd.PersistentFlags().Lookup("subscribestoevent"))
	viper.BindPFlag("eventspublished", dispatcherCmd.PersistentFlags().Lookup("eventspublished"))
	viper.BindPFlag("servicebusnamespace", dispatcherCmd.PersistentFlags().Lookup("servicebusnamespace"))
	viper.BindPFlag("resourcegroup", dispatcherCmd.PersistentFlags().Lookup("resourcegroup"))
	viper.BindPFlag("logsensitiveconfig", dispatcherCmd.PersistentFlags().Lookup("logsensitiveconfig"))
	viper.BindPFlag("moduleconfigpath", dispatcherCmd.PersistentFlags().Lookup("moduleconfigpath"))
	viper.BindPFlag("printconfig", dispatcherCmd.PersistentFlags().Lookup("printconfig"))
	// kubernetes.*
	viper.BindPFlag("kubernetes.namespace", dispatcherCmd.PersistentFlags().Lookup("kubernetes.namespace"))
	viper.BindPFlag("kubernetes.imagepullsecretname", dispatcherCmd.PersistentFlags().Lookup("kubernetes.imagepullsecretname"))
	// job.*
	viper.BindPFlag("job.maxrunningtimemins", dispatcherCmd.PersistentFlags().Lookup("job.maxrunningtimemins"))
	viper.BindPFlag("job.retrycount", dispatcherCmd.PersistentFlags().Lookup("job.retrycount"))
	viper.BindPFlag("job.workerimage", dispatcherCmd.PersistentFlags().Lookup("job.workerimage"))
	viper.BindPFlag("job.handlerimage", dispatcherCmd.PersistentFlags().Lookup("job.handlerimage"))
	viper.BindPFlag("job.pullalways", dispatcherCmd.PersistentFlags().Lookup("job.pullalways"))
	// handler.*
	viper.BindPFlag("handler.serverport", dispatcherCmd.PersistentFlags().Lookup("handler.serverport"))
	viper.BindPFlag("handler.printconfig", dispatcherCmd.PersistentFlags().Lookup("handler.printconfig"))
	// handler.azureblobprovider.*
	viper.BindPFlag("handler.azureblobprovider.blobaccountname", dispatcherCmd.PersistentFlags().Lookup("handler.azureblobprovider.blobaccountname"))
	viper.BindPFlag("handler.azureblobprovider.blobaccountkey", dispatcherCmd.PersistentFlags().Lookup("handler.azureblobprovider.blobaccountkey"))
	viper.BindPFlag("handler.azureblobprovider.useproxy", dispatcherCmd.PersistentFlags().Lookup("handler.azureblobprovider.useproxy"))
	// handler.mongodbdocprovider.*
	viper.BindPFlag("handler.mongodbdocprovider.name", dispatcherCmd.PersistentFlags().Lookup("handler.mongodbdocprovider.name"))
	viper.BindPFlag("handler.mongodbdocprovider.password", dispatcherCmd.PersistentFlags().Lookup("handler.mongodbdocprovider.password"))
	viper.BindPFlag("handler.mongodbdocprovider.collection", dispatcherCmd.PersistentFlags().Lookup("handler.mongodbdocprovider.collection"))
	viper.BindPFlag("handler.mongodbdocprovider.port", dispatcherCmd.PersistentFlags().Lookup("handler.mongodbdocprovider.port"))
	// azurebatch.*
	viper.BindPFlag("azurebatch.requiresgpu", dispatcherCmd.PersistentFlags().Lookup("azurebatch.requiresgpu"))
	viper.BindPFlag("azurebatch.resourcegroup", dispatcherCmd.PersistentFlags().Lookup("azurebatch.resourcegroup"))
	viper.BindPFlag("azurebatch.poolid", dispatcherCmd.PersistentFlags().Lookup("azurebatch.poolid"))
	viper.BindPFlag("azurebatch.jobid", dispatcherCmd.PersistentFlags().Lookup("azurebatch.jobid"))
	viper.BindPFlag("azurebatch.batchaccountname", dispatcherCmd.PersistentFlags().Lookup("azurebatch.batchaccountname"))
	viper.BindPFlag("azurebatch.batchaccountlocation", dispatcherCmd.PersistentFlags().Lookup("azurebatch.batchaccountlocation"))
	viper.BindPFlag("azurebatch.imagerepositoryserver", dispatcherCmd.PersistentFlags().Lookup("azurebatch.imagerepositoryserver"))
	viper.BindPFlag("azurebatch.imagerepositoryusername", dispatcherCmd.PersistentFlags().Lookup("azurebatch.imagerepositoryusername"))
	viper.BindPFlag("azurebatch.imagerepositorypassword", dispatcherCmd.PersistentFlags().Lookup("azurebatch.imagerepositorypassword"))

	// Add sub-commands
	dispatcherCmd.AddCommand(NewCmdStart())

	return dispatcherCmd
}

func printConfig() {
	if cfg.PrintConfig {
		if cfg.LogSensitiveConfig {
			log.Infoln(tools.PrettyPrintStruct(cfg))
		} else {
			log.Infoln(tools.PrettyPrintStruct(types.RedactConfigSecrets(&cfg)))
		}
	}
}
