package providers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lawrencegripper/ion/internal/pkg/messaging"

	"github.com/Azure/go-autorest/autorest/to"

	"github.com/lawrencegripper/ion/internal/pkg/types"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	dispatcherNameLabel = "ion/createdBy"
	messageIDLabel      = "messageid"
	correlationIDLabel  = "correlationid"
	deliverycountlabel  = "deliverycount"
)

//Check providers match interface at compile time
var _ Provider = &Kubernetes{}

// Kubernetes schedules jobs onto k8s from the queue and monitors their progress
type Kubernetes struct {
	createJob        func(*batchv1.Job) (*batchv1.Job, error)
	listAllJobs      func() (*batchv1.JobList, error)
	removeJob        func(*batchv1.Job) error
	client           *kubernetes.Clientset
	jobConfig        *types.JobConfig
	inflightJobStore map[string]messaging.Message
	dispatcherName   string
	Namespace        string
	pullSecret       string
	handlerArgs      []string
	workerEnvVars    map[string]interface{}
}

// NewKubernetesProvider Creates an instance and does basic setup
func NewKubernetesProvider(config *types.Configuration, sharedHandlerArgs []string) (*Kubernetes, error) {
	if config == nil {
		return nil, fmt.Errorf("invalid config. Cannot be nil")
	}
	if config.Job == nil {
		return nil, fmt.Errorf("invalid JobConfig. Cannot be nil")
	}

	k := Kubernetes{}
	k.handlerArgs = sharedHandlerArgs
	k.workerEnvVars = map[string]interface{}{
		"HANDLER_PORT": config.Handler.ServerPort,
	}

	// Add module specific config
	envs, err := getModuleEnvironmentVars(config.ModuleConfigPath)
	if err != nil {
		log.WithField("filepath", config.ModuleConfigPath).Error("failed to load addition module config from file")
	} else {
		for key, value := range envs {
			k.workerEnvVars[key] = value
		}
	}

	client, err := getClientSet()
	if err != nil {
		return nil, err
	}
	k.client = client

	k.Namespace = config.Kubernetes.Namespace
	k.pullSecret = config.Kubernetes.ImagePullSecretName
	k.jobConfig = config.Job
	k.dispatcherName = config.Hostname
	k.inflightJobStore = map[string]messaging.Message{}
	k.createJob = func(b *batchv1.Job) (*batchv1.Job, error) {
		return k.client.BatchV1().Jobs(k.Namespace).Create(b)
	}
	k.listAllJobs = func() (*batchv1.JobList, error) {
		return k.client.BatchV1().Jobs(k.Namespace).List(metav1.ListOptions{})
	}
	k.removeJob = func(j *batchv1.Job) error {
		return k.client.BatchV1().Jobs(k.Namespace).Delete(j.Name, &metav1.DeleteOptions{})
	}

	return &k, nil
}

// InProgressCount provides a count of the currently running jobs
func (k *Kubernetes) InProgressCount() int {
	return len(k.inflightJobStore)
}

// Reconcile will review the state of running jobs and accept or reject messages accordingly
func (k *Kubernetes) Reconcile() error {
	if k == nil {
		return fmt.Errorf("invalid properties. Provider cannot be nil")
	}
	// Todo: investigate using the field selector to limit the returned data to only
	// completed or failed jobs
	jobs, err := k.listAllJobs()
	if err != nil {
		return err
	}

	for _, j := range jobs.Items {
		messageID, ok := j.ObjectMeta.Labels[messageIDLabel]
		contextualLogger := getLoggerForJob(&j)
		if !ok {
			contextualLogger.Error("job seen without messageid present in labels... skipping")
			continue
		}

		sourceMessage, ok := k.inflightJobStore[messageID]
		contextualLogger = GetLoggerForMessage(sourceMessage, contextualLogger)
		// If we don't have a message in flight for this job check some error cases
		if !ok {
			dipatcherName, ok := j.Labels[dispatcherNameLabel]
			// Is it malformed?
			if !ok {
				contextualLogger.Error("job seen without dispatcher present in labels... skipping")
				continue
			}
			// Is it someone elses?
			if dipatcherName != k.dispatcherName {
				contextualLogger.Debug("job seen with different dispatcher name present in labels... skipping")
				continue
			}
			// Is it ours and we've forgotten
			if dipatcherName == k.dispatcherName {
				for _, condition := range j.Status.Conditions {
					if jobIsFinishedAndOlderThanAnHour(condition) {
						//Cleanup stuff that's been around a while
						err = k.removeJob(&j)
						if err != nil {
							contextualLogger.Error("cleanup: failed to remove old job job from k8s")
						}
					}
				}

				continue
			}

			//Unknown case?!
			contextualLogger.Info("unknown case when reconciling job")
			continue
		}

		for _, condition := range j.Status.Conditions {
			// Job failed - reject the message so it goes back on the queue to be retried
			if condition.Type == batchv1.JobFailed {
				contextualLogger.Warning("job failed to execute in k8s")

				err := sourceMessage.Reject()
				if err != nil {
					contextualLogger.Error("failed to reject message")
					return err
				}

				//Remove the message from the inflight message store
				delete(k.inflightJobStore, messageID)
			}

			// Job succeeded - accept the message so it is removed from the queue
			if condition.Type == batchv1.JobComplete {
				contextualLogger.Info("job successfully to execute in k8s")

				err := sourceMessage.Accept()

				if err != nil {
					contextualLogger.Error("failed to accept message")
					return err
				}

				//Remove the message from the inflight message store
				delete(k.inflightJobStore, messageID)
			}
		}
	}

	return nil
}

// Dispatch creates a job on kubernetes for the message
func (k *Kubernetes) Dispatch(message messaging.Message) error {
	if message == nil {
		return fmt.Errorf("invalid input. Message cannot be nil")
	}
	if k == nil {
		return fmt.Errorf("invalid properties. Provider cannot be nil")
	}

	perJobArgs, err := getMessageHandlerArgs(message)
	if err != nil {
		return fmt.Errorf("failed generating handler args from message: %v", err)
	}
	fullHandlerArgs := append(k.handlerArgs, perJobArgs...)
	//Prevent later append calls overwriting original backing array: https://stackoverflow.com/a/40036950/3437018
	fullHandlerArgs = fullHandlerArgs[:len(fullHandlerArgs):len(fullHandlerArgs)]

	labels := map[string]string{
		dispatcherNameLabel: k.dispatcherName,
		messageIDLabel:      message.ID(),
		deliverycountlabel:  strconv.Itoa(message.DeliveryCount()),
	}

	workerEnvVars := []apiv1.EnvVar{
		{
			Name:  "SHARED_SECRET",
			Value: message.ID(), //Todo: source from common place with args
		},
	}
	for key, value := range k.workerEnvVars {
		envVar := apiv1.EnvVar{
			Name:  key,
			Value: fmt.Sprintf("%v", value),
		}
		workerEnvVars = append(workerEnvVars, envVar)
	}

	pullPolicy := apiv1.PullIfNotPresent
	if k.jobConfig.PullAlways {
		pullPolicy = apiv1.PullAlways
	}

	handlerPrepareAgs := append(fullHandlerArgs, "--action=prepare")
	handlerCommitAgs := append(fullHandlerArgs, "--action=commit")
	deadlineSeconds := k.jobConfig.MaxRunningTimeMins * 60
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   getJobName(message),
			Labels: labels,
		},
		Spec: batchv1.JobSpec{
			Completions:           to.Int32Ptr(1),
			BackoffLimit:          to.Int32Ptr(1),
			ActiveDeadlineSeconds: to.Int64Ptr(int64(deadlineSeconds)), // Use k8s to enforce the job maxRunningTime param
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					InitContainers: []apiv1.Container{
						{
							Name:            "prepare",
							Image:           k.jobConfig.HandlerImage,
							Args:            handlerPrepareAgs,
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
							Image:           k.jobConfig.WorkerImage,
							Env:             workerEnvVars,
							ImagePullPolicy: pullPolicy,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "ionvolume",
									MountPath: "/ion",
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:            "commit",
							Image:           k.jobConfig.HandlerImage,
							Args:            handlerCommitAgs,
							ImagePullPolicy: pullPolicy,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "ionvolume",
									MountPath: "/ion",
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "ionvolume",
							VolumeSource: apiv1.VolumeSource{
								EmptyDir: &apiv1.EmptyDirVolumeSource{},
							},
						},
					},
					RestartPolicy: apiv1.RestartPolicyNever,
				},
			},
		},
	}

	// Set pull secrete if specified
	if k.pullSecret != "" {
		job.Spec.Template.Spec.ImagePullSecrets = []apiv1.LocalObjectReference{
			{
				Name: k.pullSecret,
			},
		}
	}

	kjob, err := k.createJob(job)

	if err != nil {
		log.WithError(err).Error("failed scheduling k8s job")
		mErr := message.Reject()
		if mErr != nil {
			log.WithError(mErr).Error("failed rejecting message after failing to schedule k8s job")
		}
		return err
	}

	log.WithField("messageid", message.ID()).Infof("pod created for message %s", kjob.GetSelfLink())
	k.inflightJobStore[message.ID()] = message

	return nil
}

func jobIsFinishedAndOlderThanAnHour(condition batchv1.JobCondition) bool {
	return condition.Type == batchv1.JobFailed ||
		condition.Type == batchv1.JobComplete &&
			condition.LastTransitionTime.Time.Before(time.Now().Add(-time.Hour))
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func getLoggerForJob(job *batchv1.Job) *log.Entry {
	if job == nil {
		return log.WithField("niljob", true)
	}

	entity := log.WithField("job", job)
	if job.Status.CompletionTime != nil && job.Status.StartTime != nil {
		log.WithField("taskDurationSec", job.Status.CompletionTime.Sub(job.Status.StartTime.Time).Seconds)
	}

	if c, ok := job.Annotations[correlationIDLabel]; ok {
		entity = entity.WithField(correlationIDLabel, c)
	}
	if m, ok := job.Annotations[messageIDLabel]; ok {
		entity = entity.WithField(messageIDLabel, m)
	}
	return entity
}

func getClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.WithError(err).Warn("failed getting in-cluster config attempting to use kubeconfig from homedir")
		var kubeconfig string
		if home := homeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}

		if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
			log.WithError(err).Panic("kubeconfig not found in homedir")
		}

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.WithError(err).Panic("getting kubeconf from current context")
			return nil, err
		}
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithError(err).Error("Getting clientset from config")
		return nil, err
	}

	return clientset, nil
}

func getJobName(m messaging.Message) string {
	return strings.ToLower(m.ID()) + "-v" + strconv.Itoa(m.DeliveryCount())
}
