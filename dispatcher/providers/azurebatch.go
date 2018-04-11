package providers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Azure/go-autorest/autorest/azure"

	"github.com/Azure/azure-sdk-for-go/services/batch/2017-09-01.6.0/batch"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/lawrencegripper/ion/dispatcher/helpers"
	"github.com/lawrencegripper/ion/dispatcher/messaging"
	"github.com/lawrencegripper/ion/dispatcher/types"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	apiv1 "k8s.io/api/core/v1"
)

//Check providers match interface at compile time
var _ Provider = &AzureBatch{}

// AzureBatch schedules jobs onto k8s from the queue and monitors their progress
type AzureBatch struct {
	inprogressJobStore map[string]messaging.Message
	dispatcherName     string
	sidecarArgs        []string
	workerEnvVars      map[string]interface{}
	ctx                context.Context
	cancelOps          context.CancelFunc

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
}

// NewAzureBatchProvider creates a provider for azure batch.
func NewAzureBatchProvider(config *types.Configuration, sharedSidecarArgs []string) (*AzureBatch, error) {
	if config == nil || config.AzureBatch == nil || config.Job == nil {
		return nil, fmt.Errorf("Cannot create a provider - invalid configuration, require config, AzureBatch and Job")
	}
	b := AzureBatch{}
	b.inprogressJobStore = make(map[string]messaging.Message)
	b.batchConfig = config.AzureBatch
	b.jobConfig = config.Job
	b.dispatcherName = config.Hostname + "-" + config.ModuleName
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

	// Todo: Allow users to pass in/choose a different machine type and init script
	createOrGetPool(&b, auth)
	createOrGetJob(&b, auth)

	taskclient := batch.NewTaskClientWithBaseURI(getBatchBaseURL(b.batchConfig))
	taskclient.Authorizer = auth
	b.taskClient = &taskclient

	fileClient := batch.NewFileClientWithBaseURI(getBatchBaseURL(b.batchConfig))
	fileClient.Authorizer = auth
	b.fileClient = &fileClient

	b.createTask = func(taskDetails batch.TaskAddParameter) (autorest.Response, error) {
		return b.taskClient.Add(b.ctx, b.dispatcherName, taskDetails, nil, nil, nil, nil)
	}
	b.listTasks = func() (*[]batch.CloudTask, error) {
		res, err := b.taskClient.List(b.ctx, b.dispatcherName, "", "", "", nil, nil, nil, nil, nil)
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
		return b.taskClient.Delete(b.ctx, b.dispatcherName, *t.ID, nil, nil, nil, nil, "", "", nil, nil)
	}

	return &b, nil
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

	perJobArgs, err := getMessageSidecarArgs(message)
	if err != nil {
		return fmt.Errorf("failed generating sidecar args from message: %v", err)
	}
	fullSidecarArgs := append(b.sidecarArgs, perJobArgs...)

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

	containers := []apiv1.Container{
		{
			Name:            "sidecar",
			Image:           b.jobConfig.SidecarImage,
			Args:            fullSidecarArgs,
			ImagePullPolicy: pullPolicy,
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ionvolume",
					MountPath: "/ion",
				},
			},
		},
		{
			Name:            "worker",
			Image:           b.jobConfig.WorkerImage,
			Env:             workerEnvVars,
			ImagePullPolicy: pullPolicy,
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ionvolume",
					MountPath: "/ion",
				},
			},
		},
	}

	// Todo: Pull this out into a standalone package once stabilized
	podCommand, err := getPodCommand(batchPodComponents{
		Containers: containers,
		PodName:    message.ID(),
		TaskID:     message.ID() + "-v" + strconv.Itoa(message.DeliveryCount()),
		Volumes: []apiv1.Volume{
			{
				Name: "ionvolume",
				VolumeSource: apiv1.VolumeSource{
					EmptyDir: &apiv1.EmptyDirVolumeSource{},
				},
			},
		},
	})

	log.WithField("commandtoexec", podCommand).Info("Created command for Batch")

	if err != nil {
		return err
	}

	task := batch.TaskAddParameter{
		DisplayName: to.StringPtr(fmt.Sprintf("%s:%s", b.dispatcherName, message.ID())),
		ID:          to.StringPtr(message.ID()),
		CommandLine: to.StringPtr(fmt.Sprintf(`/bin/bash -c "%s"`, podCommand)),
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
		sourceMessage, ok := b.inprogressJobStore[messageID]
		if !ok {
			log.WithField("task", t).Info("job seen which dispatcher stared but doesn't have source message... likely following a dispatcher restart")
			continue
		}

		// Job succeeded - accept the message so it is removed from the queue
		if t.State == batch.TaskStateCompleted && *t.ExecutionInfo.ExitCode == 0 {
			//Remove the task from batch
			_, err := b.removeTask(&t)
			if err != nil {
				log.WithError(err).WithField("task", t).WithField("messageID", messageID).Error("Failed to remove COMPLETED task from batch")
			}

			err = sourceMessage.Accept()

			if err != nil {
				log.WithFields(log.Fields{
					"message": sourceMessage,
					"task":    t,
				}).Error("failed to accept message")
				return err
			}

			log.WithFields(log.Fields{
				"message": sourceMessage,
				"task":    t,
			}).Info("Task completed with success exit code")

			//Remove the message from the inflight message store
			delete(b.inprogressJobStore, messageID)
			continue
		}

		if t.State == batch.TaskStateCompleted && *t.ExecutionInfo.ExitCode != 0 {
			//Remove the task from batch
			_, err := b.removeTask(&t)
			if err != nil {
				log.WithError(err).WithField("task", t).WithField("messageID", messageID).Error("Failed to remove FAILED task from batch")
			}

			err = sourceMessage.Reject()

			if err != nil {
				log.WithFields(log.Fields{
					"message": sourceMessage,
					"task":    t,
				}).Error("failed to accept message")
				return err
			}

			log.WithFields(log.Fields{
				"message": sourceMessage,
				"task":    t,
			}).Info("Task completed with failed exit code")

			//Remove the message from the inflight message store
			delete(b.inprogressJobStore, messageID)
			continue
		}

		if t.State != batch.TaskStateCompleted {
			// Todo: Handle max execution time here
			log.WithFields(log.Fields{
				"message": sourceMessage,
				"task":    t,
			}).Info("Skipping task not currently in completed state")
		}
	}

	return nil
}

type imageRegistryCredential struct {
	Server   string `json:"server,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// BatchPodComponents provides details for the batch task wrapper
// to run a pod
type batchPodComponents struct {
	PullCredentials []imageRegistryCredential
	Containers      []v1.Container
	Volumes         []v1.Volume
	PodName         string
	TaskID          string
}
