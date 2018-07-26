package providers

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/Azure/go-autorest/autorest/azure"

	"github.com/Azure/azure-sdk-for-go/services/batch/2017-09-01.6.0/batch"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/lawrencegripper/ion/internal/app/dispatcher/helpers"
	"github.com/lawrencegripper/ion/internal/pkg/messaging"
	"github.com/lawrencegripper/ion/internal/pkg/types"
	"github.com/lawrencegripper/pod2docker"
	log "github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	v1resource "k8s.io/apimachinery/pkg/api/resource"
)

//Check providers match interface at compile time
var _ Provider = &AzureBatch{}

// AzureBatch schedules jobs onto k8s from the queue and monitors their progress
type AzureBatch struct {
	inprogressJobStore map[string]messaging.Message
	jobID              string
	handlerArgs        []string
	workerEnvVars      map[string]interface{}
	ctx                context.Context
	cancelOps          context.CancelFunc
	logStore           *LogStore

	jobConfig   *types.JobConfig
	batchConfig *types.AzureBatchConfig
	poolClient  *batch.PoolClient
	jobClient   *batch.JobClient
	taskClient  *batch.TaskClient
	fileClient  *batch.FileClient

	// Used to allow mocking of the batch api for testing
	createTask func(taskDetails batch.TaskAddParameter) (autorest.Response, error)
	listTasks  func() (*[]batch.CloudTask, error)
	removeTask func(*batch.CloudTask) (autorest.Response, error)
	getLogs    func(*batch.CloudTask) string
}

// NewAzureBatchProvider creates a provider for azure batch.
func NewAzureBatchProvider(config *types.Configuration, sharedHandlerArgs []string) (*AzureBatch, error) {
	if config == nil || config.AzureBatch == nil || config.Job == nil {
		return nil, fmt.Errorf("Cannot create a provider - invalid configuration, require config, AzureBatch and Job")
	}
	b := AzureBatch{}
	b.handlerArgs = sharedHandlerArgs
	b.inprogressJobStore = make(map[string]messaging.Message)
	b.batchConfig = config.AzureBatch
	b.jobConfig = config.Job
	b.jobID = config.Hostname + "-" + config.ModuleName
	b.workerEnvVars = make(map[string]interface{})
	ctx, cancel := context.WithCancel(context.Background())
	b.ctx = ctx
	b.cancelOps = cancel

	auth := helpers.GetAzureADAuthorizer(config, azure.PublicCloud.BatchManagementEndpoint)

	// Add module specific config
	envs, err := getModuleEnvironmentVars(config.ModuleConfigPath)
	if err != nil {
		log.WithField("filepath", config.ModuleConfigPath).Error("failed to load addition module config from file")
	} else {
		for key, value := range envs {
			b.workerEnvVars[key] = value
		}
	}

	batchBaseURL := getBatchBaseURL(config.AzureBatch.BatchAccountName, config.AzureBatch.BatchAccountLocation)
	poolClient, err := getPool(ctx, batchBaseURL, config.AzureBatch.PoolID, auth)
	if err != nil {
		return nil, err
	}
	b.poolClient = poolClient

	jobClient, err := createOrGetJob(ctx, batchBaseURL, b.jobID, config.AzureBatch.PoolID, auth)
	if err != nil {
		return nil, err
	}
	b.jobClient = jobClient

	taskclient := batch.NewTaskClientWithBaseURI(batchBaseURL)
	taskclient.Authorizer = auth
	b.taskClient = &taskclient

	fileClient := batch.NewFileClientWithBaseURI(batchBaseURL)
	fileClient.Authorizer = auth
	b.fileClient = &fileClient

	b.createTask = func(taskDetails batch.TaskAddParameter) (autorest.Response, error) {
		return b.taskClient.Add(b.ctx, b.jobID, taskDetails, nil, nil, nil, nil)
	}
	b.listTasks = func() (*[]batch.CloudTask, error) {
		res, err := b.taskClient.List(b.ctx, b.jobID, "", "", "", nil, nil, nil, nil, nil)
		if err != nil {
			return &[]batch.CloudTask{}, err
		}
		currentTasks := res.Values()
		for res.NotDone() {
			err = res.Next()
			if err != nil {
				return &[]batch.CloudTask{}, err
			}
			pageTasks := res.Values()
			if pageTasks != nil || len(pageTasks) != 0 {
				currentTasks = append(currentTasks, pageTasks...)
			}
		}

		return &currentTasks, nil
	}
	b.removeTask = func(t *batch.CloudTask) (autorest.Response, error) {
		//remove the task
		return b.taskClient.Delete(b.ctx, b.jobID, *t.ID, nil, nil, nil, nil, "", "", nil, nil)
	}
	b.getLogs = func(t *batch.CloudTask) string { return getLogsForTask(b.ctx, b.fileClient, t, b.jobID) }

	if config.Handler != nil &&
		config.Handler.MongoDBDocumentStorageProvider != nil &&
		config.Handler.AzureBlobStorageProvider != nil {
		b.logStore, err = NewLogStore(config.Handler.MongoDBDocumentStorageProvider, config.Handler.AzureBlobStorageProvider, config.ModuleName)
		if err != nil {
			log.WithError(err).Error("failed to create log store")
			return nil, err
		}
	} else {
		log.Info("Skipping logstore config as not provided")
		b.logStore = &LogStore{}
	}
	return &b, nil
}

//GetActiveMessages gets the currently active messages
func (b *AzureBatch) GetActiveMessages() []messaging.Message {
	activeMessages := make([]messaging.Message, 0, len(b.inprogressJobStore))
	for _, m := range b.inprogressJobStore {
		activeMessages = append(activeMessages, m)
	}
	return activeMessages
}

// InProgressCount will show how many tasks are currently in progress
func (b *AzureBatch) InProgressCount() int {
	return len(b.inprogressJobStore)
}

// Dispatch will dispatch a job onto Azure Batch
func (b *AzureBatch) Dispatch(message messaging.Message) error {
	if message == nil {
		return fmt.Errorf("invalid input. Message cannot be nil")
	}
	if b == nil {
		return fmt.Errorf("invalid properties. Provider cannot be nil")
	}

	perJobArgs, err := getMessageHandlerArgs(message)
	if err != nil {
		return fmt.Errorf("failed generating handler args from message: %v", err)
	}
	fullHandlerArgs := append(b.handlerArgs, perJobArgs...)
	//Prevent later append calls overwriting original backing array: https://stackoverflow.com/a/40036950/3437018
	fullHandlerArgs = fullHandlerArgs[:len(fullHandlerArgs):len(fullHandlerArgs)]

	workerEnvVars := []apiv1.EnvVar{
		{
			Name:  "SHARED_SECRET",
			Value: message.ID(), //Todo: source from common place with args
		},
	}
	for k, v := range b.workerEnvVars {
		envVar := apiv1.EnvVar{
			Name:  k,
			Value: fmt.Sprintf("%v", v),
		}
		workerEnvVars = append(workerEnvVars, envVar)
	}

	pullPolicy := apiv1.PullIfNotPresent
	if b.jobConfig.PullAlways {
		pullPolicy = apiv1.PullAlways
	}

	//Todo: Refactor this bit as got a bit messy now.
	moduleContainer := apiv1.Container{
		Name:            "modulecontainer",
		Image:           b.jobConfig.WorkerImage,
		Env:             workerEnvVars,
		ImagePullPolicy: pullPolicy,
		VolumeMounts: []apiv1.VolumeMount{
			{
				Name:      "ionvolume",
				MountPath: "/ion",
			},
		},
	}

	if b.batchConfig != nil && b.batchConfig.RequiresGPU {
		moduleContainer.Resources.Limits = apiv1.ResourceList{}
		moduleContainer.Resources.Limits["nvidia.com/gpu"] = *v1resource.NewQuantity(1, v1resource.DecimalSI)
	}
	initContainers := []apiv1.Container{
		{
			Name:            "prepare",
			Image:           b.jobConfig.HandlerImage,
			Args:            append(fullHandlerArgs, "--action=prepare"),
			ImagePullPolicy: pullPolicy,
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ionvolume",
					MountPath: "/ion",
				},
			},
		},
		moduleContainer,
	}
	containers := []apiv1.Container{
		{
			Name:            "commit",
			Image:           b.jobConfig.HandlerImage,
			Args:            append(fullHandlerArgs, "--action=commit"),
			ImagePullPolicy: pullPolicy,
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ionvolume",
					MountPath: "/ion",
				},
			},
		},
	}

	podComponent := pod2docker.PodComponents{
		InitContainers: initContainers,
		Containers:     containers,
		PodName:        message.ID() + "-v" + strconv.Itoa(message.DeliveryCount()),
		Volumes: []apiv1.Volume{
			{
				Name: "ionvolume",
				VolumeSource: apiv1.VolumeSource{
					EmptyDir: &apiv1.EmptyDirVolumeSource{},
				},
			},
		},
	}

	if b.batchConfig != nil && b.batchConfig.ImageRepositoryServer != "" {
		podComponent.PullCredentials = []pod2docker.ImageRegistryCredential{
			{
				Server:   b.batchConfig.ImageRepositoryServer,
				Username: b.batchConfig.ImageRepositoryUsername,
				Password: b.batchConfig.ImageRepositoryPassword,
			},
		}
	}

	podCommand, err := pod2docker.GetBashCommand(podComponent)

	log.WithField("commandtoexec", podCommand).Info("Created command for Batch")

	if err != nil {
		return err
	}

	task := batch.TaskAddParameter{
		DisplayName: to.StringPtr(fmt.Sprintf("%s:%s", b.jobID, message.ID())),
		ID:          to.StringPtr(message.ID()),
		CommandLine: to.StringPtr(fmt.Sprintf(`/bin/bash -c "%s"`, podCommand)),
		Constraints: &batch.TaskConstraints{
			MaxWallClockTime: to.StringPtr(fmt.Sprintf("PT%dM", b.jobConfig.MaxRunningTimeMins)),
		},
		UserIdentity: &batch.UserIdentity{
			AutoUser: &batch.AutoUserSpecification{
				ElevationLevel: batch.Admin,
				Scope:          batch.Pool,
			},
		},
	}
	_, err = b.createTask(task)
	if err != nil {
		log.WithError(err).Error("failed scheduling azurebatch task")
		mErr := message.Reject()
		if mErr != nil {
			log.WithError(mErr).Error("failed rejecting message after failing to schedule azurebatch task")
		}
		return err
	}

	b.inprogressJobStore[message.ID()] = message

	return nil
}

// Reconcile will check inprogress tasks against and accept/reject messages were the job has completed/failed
func (b *AzureBatch) Reconcile() error {
	if b == nil {
		return fmt.Errorf("invalid properties. Provider cannot be nil")
	}
	tasks, err := b.listTasks()
	if err != nil {
		return err
	}
	if tasks == nil {
		return fmt.Errorf("task list returned nil")
	}

	for _, t := range *tasks {
		messageIDPtr := t.ID
		if messageIDPtr == nil {
			log.WithField("task", t).Error("task seen with nil messageid... skipping")
			continue
		}

		messageID := *messageIDPtr
		contextualLogger := log.WithFields(logFieldsForTask(&t)).WithField("messageID", messageID)

		sourceMessage, ok := b.inprogressJobStore[messageID]
		if !ok {
			contextualLogger.Info("job seen which dispatcher stared but doesn't have source message... likely following a dispatcher restart")
			continue
		}

		// Job succeeded - accept the message so it is removed from the queue
		if t.State == batch.TaskStateCompleted {

			if *t.ExecutionInfo.ExitCode == 0 {
				// Task has completed successfully
				contextualLogger.Info("Task completed with success exit code")

				logs := b.getLogs(&t)
				if err != nil {
					contextualLogger.WithError(err).Error("failed to get logs for job: getLogsFailed")
				}
				err := b.logStore.StoreLogs(contextualLogger, sourceMessage, logs, true)
				if err != nil {
					contextualLogger.WithError(err).Error("failed to log to logstore")
				}

				//Remove the task from batch
				_, err = b.removeTask(&t)
				if err != nil {
					contextualLogger.Error("Failed to remove COMPLETED task from batch")
				}

				//ACK the message to remove from queue
				err = sourceMessage.Accept()

				if err != nil {
					contextualLogger.Error("failed to accept message")
					return err
				}

				//Remove the message from the inflight message store
				delete(b.inprogressJobStore, messageID)
				continue
			} else {
				//Task has failed!
				contextualLogger.Info("Task completed with failed exit code")

				logs := b.getLogs(&t)
				if err != nil {
					contextualLogger.WithError(err).Error("failed to get logs for job: getLogsFailed")
				}
				err := b.logStore.StoreLogs(contextualLogger, sourceMessage, logs, false)
				if err != nil {
					contextualLogger.WithError(err).Error("failed to log to logstore")
				}

				// Remove the task from batch
				_, err = b.removeTask(&t)
				if err != nil {
					log.WithError(err).WithField("task", t).WithField("messageID", messageID).Error("Failed to remove FAILED task from batch")
				}

				//ACK the message to requeue failure
				err = sourceMessage.Reject()

				if err != nil {
					log.WithFields(log.Fields{
						"message": sourceMessage,
						"task":    t,
					}).Error("failed to accept message")
					return err
				}

				//Remove the message from the inflight message store
				delete(b.inprogressJobStore, messageID)
				continue
			}
		}

		if t.State != batch.TaskStateCompleted {
			log.WithFields(log.Fields{
				"message": sourceMessage,
				"task":    t,
			}).Info("Skipping task not currently in completed state")
		}
	}

	return nil
}

func getLogFileContent(ctx context.Context, t *batch.CloudTask, fileClient *batch.FileClient, jobName, filepath string) (string, error) {
	stdout, err := fileClient.GetFromTask(ctx, jobName, *t.ID, filepath, nil, nil, nil, nil, "", nil, nil)
	if err != nil {
		return "", err
	}
	stdoutByptes, err := ioutil.ReadAll(*stdout.Value)
	if err != nil {
		return "", err
	}
	return string(stdoutByptes), nil
}

func getLogsForTask(ctx context.Context, fileClient *batch.FileClient, t *batch.CloudTask, jobID string) string {
	sb := strings.Builder{}
	//Log the details from the running task
	logs, err := getLogFileContent(ctx, t, fileClient, jobID, "stdout.txt")
	sb.WriteString("\n\n ------ [Debug] Batch logs: stdout ------ \n\n") //nolint: errcheck
	if err != nil {
		log.WithError(err).WithField("task", *t).Warningf("failed to get %v from task", "stout")
		sb.WriteString("failed to get %v from task: stdout") //nolint: errcheck
		return sb.String()
	}
	sb.WriteString(logs) //nolint: errcheck

	logs, err = getLogFileContent(ctx, t, fileClient, jobID, "stderr.txt")

	sb.WriteString("\n\n ------ [Debug] Batch logs: sterr ------ \n\n") //nolint: errcheck
	if err != nil {
		log.WithError(err).WithField("task", *t).Warningf("failed to get %v from task", "stout")
		sb.WriteString("failed to get %v from task: stdout") //nolint: errcheck
		return sb.String()
	}
	sb.WriteString(logs) //nolint: errcheck

	logs, err = getLogFileContent(ctx, t, fileClient, jobID, "wd/preparecontainer.log")
	sb.WriteString("\n\n ------ Preparer logs ------ \n\n") //nolint: errcheck
	if err != nil {
		log.WithError(err).WithField("task", *t).Warningf("failed to get %v from task", "stout")
		sb.WriteString("failed to get %v from task: stdout") //nolint: errcheck
		return sb.String()
	}
	sb.WriteString(logs) //nolint: errcheck

	logs, err = getLogFileContent(ctx, t, fileClient, jobID, "wd/modulecontainer.log")
	sb.WriteString("\n\n ------ Module logs ------ \n\n") //nolint: errcheck
	if err != nil {
		log.WithError(err).WithField("task", *t).Warningf("failed to get %v from task", "stout")
		sb.WriteString("failed to get %v from task: stdout") //nolint: errcheck
		return sb.String()
	}
	sb.WriteString(logs) //nolint: errcheck

	logs, err = getLogFileContent(ctx, t, fileClient, jobID, "wd/commitcontainer.log")
	sb.WriteString("\n\n ------ Commit logs ------ \n\n") //nolint: errcheck
	if err != nil {
		log.WithError(err).WithField("task", *t).Warningf("failed to get %v from task", "stout")
		sb.WriteString("failed to get %v from task: stdout") //nolint: errcheck
		return sb.String()
	}
	sb.WriteString(logs) //nolint: errcheck

	return sb.String()

}
