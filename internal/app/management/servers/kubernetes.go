//nolint: golint
package servers

import (
	"encoding/base64"
	"fmt"
	"github.com/lawrencegripper/ion/internal/app/management/types"
	"github.com/lawrencegripper/ion/internal/pkg/management/module"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	context "golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/errors"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

//Check at compile time if we implement the interface
var _ module.ModuleServiceServer = (*Kubernetes)(nil)

// Kubernetes management server
type Kubernetes struct {
	client                    *kubernetes.Clientset
	namespace                 string
	AzureSPSecretRef          string
	AzureBlobStorageSecretRef string
	AzureServiceBusSecretRef  string
	MongoDBSecretRef          string
	DispatcherImageName       string
	ID                        string
}

const createdByLabel = "ion/createdBy"
const moduleNameLabel = "ion/moduleName"
const idLabel = "ion/id"

var sharedServicesSecretName string
var sharedImagePullSecretName string
var logLevel string

func genID() string {
	id := xid.New()
	return id.String()[0:5]
}

// NewKubernetesManagementServer creates and initializes a new Kubernetes management server
func NewKubernetesManagementServer(config *types.Configuration) (*Kubernetes, error) {
	k := Kubernetes{}

	logLevel = config.LogLevel

	k.ID = "management-api"
	k.DispatcherImageName = config.DispatcherImage

	var err error
	k.client, err = getClientSet()
	if err != nil {
		return nil, fmt.Errorf("error connecting to Kubernetes %+v", err)
	}
	k.namespace = config.Namespace

	err = k.createSharedServicesSecret(config)
	if err != nil {
		return nil, err
	}

	err = k.createSharedImagePullSecret(config)
	if err != nil {
		return nil, err
	}

	return &k, nil
}

// createSharedServicesSecret creates a shared secret
// that stores all the config needed by the dispatcher
// to operate i.e. dataplane provider connection
func (k *Kubernetes) createSharedServicesSecret(config *types.Configuration) error {
	sharedServicesSecretName = fmt.Sprintf("services-%s", k.ID)

	secret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: sharedServicesSecretName,
			Labels: map[string]string{
				createdByLabel: k.ID,
			},
		},
		StringData: map[string]string{
			"CLIENTID":                                  config.AzureClientID,
			"CLIENTSECRET":                              config.AzureClientSecret,
			"SUBSCRIPTIONID":                            config.AzureSubscriptionID,
			"TENANTID":                                  config.AzureTenantID,
			"SERVICEBUSNAMESPACE":                       config.AzureServiceBusNamespace,
			"RESOURCEGROUP":                             config.AzureResourceGroup,
			"AZUREBATCH_JOBID":                          config.AzureBatchJobID,
			"AZUREBATCH_POOLID":                         config.AzureBatchPoolID,
			"AZUREBATCH_BATCHACCOUNTLOCATION":           config.AzureBatchAccountLocation,
			"AZUREBATCH_BATCHACCOUNTNAME":               config.AzureBatchAccountName,
			"AZUREBATCH_REQUIRESGPU":                    strconv.FormatBool(config.AzureBatchRequiresGPU),
			"AZUREBATCH_RESOURCEGROUP":                  config.AzureBatchResourceGroup,
			"AZUREBATCH_IMAGEREPOSITORYSERVER":          config.AzureBatchImageRepositoryServer,
			"AZUREBATCH_IMAGEREPOSITORYPASSWORD":        config.AzureBatchImageRepositoryPassword,
			"AZUREBATCH_IMAGEREPOSITORYUSERNAME":        config.AzureBatchImageRepositoryUsername,
			"HANDLER_MONGODBDOCPROVIDER_PORT":           strconv.Itoa(config.MongoDBPort),
			"HANDLER_MONGODBDOCPROVIDER_NAME":           config.MongoDBName,
			"HANDLER_MONGODBDOCPROVIDER_PASSWORD":       config.MongoDBPassword,
			"HANDLER_MONGODBDOCPROVIDER_COLLECTION":     config.MongoDBCollection,
			"HANDLER_AZUREBLOBPROVIDER_BLOBACCOUNTNAME": config.AzureStorageAccountName,
			"HANDLER_AZUREBLOBPROVIDER_BLOBACCOUNTKEY":  config.AzureStorageAccountKey,
			"LOGGING_APPINSIGHTS":                       config.AppInsightsKey,
		},
	}

	if err := k.createSecretIfNotExist(secret); err != nil {
		return err
	}
	return nil
}

// createSharedImagePullSecret creates a shared secrect to store
// the module container registry connection details if they
// are provided. These will be used by the dispatcher to pull
// the module image.
func (k *Kubernetes) createSharedImagePullSecret(config *types.Configuration) error {

	sharedImagePullSecretName = fmt.Sprintf("imagepull-%s", k.ID)
	if config.ContainerImageRegistryUsername != "" &&
		config.ContainerImageRegistryPassword != "" &&
		config.ContainerImageRegistryURL != "" {

		auth := encodeBase64(fmt.Sprintf("%s:%s", config.ContainerImageRegistryUsername, config.ContainerImageRegistryPassword))
		dockerAuthConfig := fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"%s","email":"%s","auth":"%s"}}}`,
			config.ContainerImageRegistryURL,
			config.ContainerImageRegistryUsername,
			config.ContainerImageRegistryPassword,
			config.ContainerImageRegistryEmail,
			auth)

		secret := &apiv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: sharedImagePullSecretName,
				Labels: map[string]string{
					createdByLabel: k.ID,
				},
			},
			Data: map[string][]byte{
				".dockerconfigjson": []byte(dockerAuthConfig),
			},
			Type: apiv1.SecretTypeDockerConfigJson,
		}

		if err := k.createSecretIfNotExist(secret); err != nil {
			return err
		}
	}
	return nil
}

func (k *Kubernetes) createSecretIfNotExist(secret *apiv1.Secret) error {
	secretsClient := k.client.CoreV1().Secrets(k.namespace)
	_, err := secretsClient.Get(secret.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("secret %s not found, creating it\n", secret.Name)
		_, err := secretsClient.Create(secret)
		if err != nil {
			return fmt.Errorf("error creating dispatcher secret %+v", err)
		}
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		return fmt.Errorf("error getting secret %s: %v\n", secret.Name, statusError.ErrStatus.Message)
	} else if err != nil {
		return err
	}

	return nil
}

// Create will create the necessary services to support the execution of
// a module. This includes a configmap to hold the module's configuration
// and a deployment that runs a disptcher pod. The dispatcher pod will
// orchestrate the execution of the module itself.
func (k *Kubernetes) Create(ctx context.Context, r *module.ModuleCreateRequest) (*module.ModuleCreateResponse, error) {
	// a unique ID for this creation
	id := fmt.Sprintf("%s-%s", r.Modulename, genID())

	// Validate provider
	useAzureBatchProvider := false
	switch strings.ToLower(r.Provider) {
	case "azurebatch":
		useAzureBatchProvider = true
	case "kubernetes":
		// noop
	default:
		return nil, fmt.Errorf("unrecognized provider %s", r.Provider)
	}

	// Create a configmap to store the configuration details
	// needed by the module. These will be mounted into the
	// dispatcher as a volume and then passed on when it
	// dispatches the module.
	moduleConfigMapName := id

	var stringBuilder strings.Builder
	for k, v := range r.Configmap {
		_, _ = stringBuilder.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	configMapStr := strings.TrimSuffix(stringBuilder.String(), "\n")

	moduleConfigMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: moduleConfigMapName,
			Labels: map[string]string{
				createdByLabel:  k.ID,
				idLabel:         id,
				moduleNameLabel: r.Modulename,
			},
		},
		Data: map[string]string{
			"module": configMapStr,
		},
	}

	configMapClient := k.client.CoreV1().ConfigMaps(k.namespace)
	_, err := configMapClient.Create(moduleConfigMap)
	if err != nil {
		return nil, fmt.Errorf("error creating module config map %+v", err)
	}

	configMapFilePath := "/etc/config"

	// Create an argument list to provide the the dispatcher binary
	dispatcherArgs := []string{
		"start",
		"--modulename=" + r.Modulename,
		"--moduleconfigpath=" + fmt.Sprintf("%s/module", configMapFilePath),
		"--subscribestoevent=" + r.Eventsubscriptions,
		"--eventspublished=" + r.Eventpublications,
		"--azurebatch.enabled=" + strconv.FormatBool(useAzureBatchProvider),
		"--job.workerimage=" + r.Moduleimage,
		"--job.handlerimage=" + r.Handlerimage,
		"--job.retrycount=" + fmt.Sprintf("%d", r.Retrycount),
		"--job.pullalways=false",
		"--job.maxrunningtimemins=" + fmt.Sprintf("%d", r.Maxexecutiontimemins),
		"--kubernetes.namespace=" + k.namespace,
		"--kubernetes.imagepullsecretname=" + sharedImagePullSecretName,
		"--loglevel=" + logLevel,
		"--printconfig=true",
	}

	dispatcherDeploymentName := id

	// Create a deployment that runs a dispatcher
	// pod, passing in environment variables from
	// a secret and mounting a volume from a configmap.
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: dispatcherDeploymentName,
			Labels: map[string]string{
				createdByLabel:  k.ID,
				idLabel:         dispatcherDeploymentName,
				moduleNameLabel: r.Modulename,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(r.Instancecount),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "ion-dispatcher",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: dispatcherDeploymentName,
					Labels: map[string]string{
						"app": "ion-dispatcher",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "ion-dispatcher",
							Image: k.DispatcherImageName,
							Args:  dispatcherArgs,
							EnvFrom: []apiv1.EnvFromSource{
								{
									SecretRef: &apiv1.SecretEnvSource{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: sharedServicesSecretName,
										},
									},
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "module-config",
									MountPath: configMapFilePath,
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: "module-config",
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: moduleConfigMapName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	deploymentsClient := k.client.AppsV1().Deployments(k.namespace)
	_, err = deploymentsClient.Create(deployment)
	if err != nil {
		return nil, fmt.Errorf("error creating dispatcher deployment %+v", err)
	}

	var createResponse = &module.ModuleCreateResponse{
		Name: id,
	}

	return createResponse, nil
}

// Delete will delete all the components associated with a module deployment.
// This includes deleting the configmap that holds the module's configuration
// and the deployment of the module's dispatcher.
func (k *Kubernetes) Delete(ctx context.Context, r *module.ModuleDeleteRequest) (*module.ModuleDeleteResponse, error) {
	// Find deployments with matching label and delete them
	deploymentsClient := k.client.AppsV1().Deployments(k.namespace)
	deployments, err := deploymentsClient.List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", idLabel, r.Name),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing deployments with name %s", r.Name)
	}
	for _, deployment := range deployments.Items {
		if err := deploymentsClient.Delete(deployment.Name, nil); err != nil {
			return nil, fmt.Errorf("error deleting deployment %s", deployment.Name)
		}
	}

	// Find configmaps with matching label and delete them
	configMapClient := k.client.CoreV1().ConfigMaps(k.namespace)
	configmaps, err := configMapClient.List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", idLabel, r.Name),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing configmaps with name %s", r.Name)
	}
	for _, configmap := range configmaps.Items {
		if err := configMapClient.Delete(configmap.Name, nil); err != nil {
			return nil, fmt.Errorf("error deleting configmap %s", configmap.Name)
		}
	}

	var deleteResponse = &module.ModuleDeleteResponse{
		Name: r.Name,
	}

	return deleteResponse, nil
}

// List will list all the deployments that have been created by this
// management server. It will simply list the deployment modules name.
func (k *Kubernetes) List(ctx context.Context, r *module.ModuleListRequest) (*module.ModuleListResponse, error) {
	// Find deployments with matching label
	deploymentsClient := k.client.AppsV1().Deployments(k.namespace)
	deployments, err := deploymentsClient.List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", createdByLabel, k.ID),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing deployments with label %s", k.ID)
	}
	var list = &module.ModuleListResponse{}
	for _, deployment := range deployments.Items {
		list.Names = append(list.Names, deployment.Name)
	}
	return list, nil
}

// Get will get information about a deployed module
func (k *Kubernetes) Get(ctx context.Context, r *module.ModuleGetRequest) (*module.ModuleGetResponse, error) {
	// Get the deployment with the given name
	deploymentsClient := k.client.AppsV1().Deployments(k.namespace)
	deployment, err := deploymentsClient.Get(r.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting deployments with name %s", r.Name)
	}
	var getResponse = &module.ModuleGetResponse{
		Name:          deployment.Name,
		Status:        string(deployment.Status.Conditions[0].Type),
		StatusMessage: deployment.Status.Conditions[0].Message,
	}
	return getResponse, nil
}

func encodeBase64(s string) string {
	const lineLen = 70
	encLen := base64.StdEncoding.EncodedLen(len(s))
	lines := encLen/lineLen + 1
	buf := make([]byte, encLen*2+lines)
	in := buf[0:encLen]
	out := buf[encLen:]
	base64.StdEncoding.Encode(in, []byte(s))
	k := 0
	for i := 0; i < len(in); i += lineLen {
		j := i + lineLen
		if j > len(in) {
			j = len(in)
		}
		k += copy(out[k:], in[i:j])
		if lines > 1 {
			out[k] = '\n'
			k++
		}
	}
	return string(out[:k])
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

func int32Ptr(i int32) *int32 { return &i }
