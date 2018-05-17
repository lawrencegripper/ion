package providers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	dispatcherNameLabel = "dispatchername"
	messageIDLabel      = "messageid"
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
	sidecarArgs      []string
	workerEnvVars    map[string]interface{}
}

// NewKubernetesProvider Creates an instance and does basic setup
func NewKubernetesProvider(config *types.Configuration, sharedSidecarArgs []string) (*Kubernetes, error) {
	if config == nil {
		return nil, fmt.Errorf("invalid config. Cannot be nil")
	}
	if config.Job == nil {
		return nil, fmt.Errorf("invalid JobConfig. Cannot be nil")
	}

	k := Kubernetes{}
	k.sidecarArgs = sharedSidecarArgs
	k.workerEnvVars = map[string]interface{}{
		"SIDECAR_PORT": config.Sidecar.ServerPort,
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
		if !ok {
			log.WithField("job", j).Error("job seen without messageid present in labels... skipping")
			continue
		}

		sourceMessage, ok := k.inflightJobStore[messageID]
		// If we don't have a message in flight for this job check some error cases
		if !ok {
			dipatcherName, ok := j.Labels[dispatcherNameLabel]
			// Is it malformed?
			if !ok {
				log.WithField("job", j).Error("job seen without dispatcher present in labels... skipping")
				continue
			}
			// Is it someone elses?
			if dipatcherName != k.dispatcherName {
				log.WithField("job", j).Debug("job seen with different dispatcher name present in labels... skipping")
				continue
			}
			// Is it ours and we've forgotten
			if dipatcherName == k.dispatcherName {
				//log.WithField("job", j).Info("job seen which dispatcher stared but doesn't have source message... likely following a dispatcher restart")
				// Todo: Should we clean these up at some point. Maybe after a wait time?
				// We want to leave them for a bit as they may be the result of a crash
				// in which case we can use them to recover
				continue
			}

			//Unknown case?!
			log.WithField("job", j).Info("unknown case when reconciling job")
			continue
		}

		for _, condition := range j.Status.Conditions {
			// Job failed - reject the message so it goes back on the queue to be retried
			if condition.Type == batchv1.JobFailed {
				//Remove the job from k8s
				err = k.removeJob(&j)
				if err != nil {
					log.WithError(err).WithField("job", j).WithField("messageID", messageID).Error("Failed to remove FAILED job from k8s")
				}

				err := sourceMessage.Reject()

				if err != nil {
					log.WithFields(log.Fields{
						"message": sourceMessage,
						"job":     j,
					}).Error("failed to reject message")
					return err
				}

				//Remove the message from the inflight message store
				delete(k.inflightJobStore, messageID)
			}

			// Job succeeded - accept the message so it is removed from the queue
			if condition.Type == batchv1.JobComplete {
				//Remove the job from k8s
				err = k.removeJob(&j)
				if err != nil {
					log.WithError(err).WithField("job", j).WithField("messageID", messageID).Error("Failed to remove COMPLETED job from k8s")
				}

				err := sourceMessage.Accept()

				if err != nil {
					log.WithFields(log.Fields{
						"message": sourceMessage,
						"job":     j,
					}).Error("failed to accept message")
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

	perJobArgs, err := getMessageSidecarArgs(message)
	if err != nil {
		return fmt.Errorf("failed generating sidecar args from message: %v", err)
	}
	fullSidecarArgs := append(k.sidecarArgs, perJobArgs...)

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

	sidecarPrepareAgs := append(fullSidecarArgs, "--action=prepare")
	sidecarCommitAgs := append(fullSidecarArgs, "--action=commit")
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
							Image:           k.jobConfig.SidecarImage,
							Args:            sidecarPrepareAgs,
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
							Image:           k.jobConfig.SidecarImage,
							Args:            sidecarCommitAgs,
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

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
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
