package management

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lawrencegripper/ion/internal/app/management/module"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type server struct{}

// Kubernetes client
type Kubernetes struct {
	client                    *kubernetes.Clientset
	namespace                 string
	AzureSPSecretRef          string
	AzureBlobStorageSecretRef string
	AzureServiceBusSecretRef  string
	MongoDBSecretRef          string
}

// Management holds metadata used by the management server
type Management struct {
	DispatcherImageName string
	DispatcherImageTag  string
	ID                  string
}

var k Kubernetes
var mgmt *Management
var sharedServicesSecretName string
var sharedImagePullSecretName string
var logLevel string

func genID() string {
	id := xid.New()
	return id.String()
}

// Run the GRPC server
func Run(config *Configuration) {

	logLevel = config.LogLevel

	mgmt = &Management{
		ID:                  genID(),
		DispatcherImageName: config.DispatcherImage,
		DispatcherImageTag:  config.DispatcherImageTag,
	}

	var err error
	k.client, err = getClientSet()
	if err != nil {
		panic(fmt.Errorf("error connecting to Kubernetes %+v", err))
	}
	k.namespace = config.Namespace

	err = createSharedServicesSecret(config)
	if err != nil {
		panic(err)
	}

	err = createSharedImagePullSecret(config)
	if err != nil {
		panic(err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		panic(fmt.Errorf("failed to listen: %v", err))
	}
	s := grpc.NewServer()
	module.RegisterModuleServiceServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		panic(fmt.Errorf("failed to serve: %v", err))
	}
}

func createSharedServicesSecret(config *Configuration) error {
	// create a shared secret that stores all the config
	// needed by the dispatcher to operate i.e. dataplane
	// provider connection
	sharedServicesSecretName = fmt.Sprintf("services-%s", mgmt.ID)

	secret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: sharedServicesSecretName,
			Labels: map[string]string{
				"createdBy": mgmt.ID,
			},
		},
		StringData: map[string]string{
			"CLIENTID":                                  config.AzureClientID,
			"CLIENTSECRET":                              config.AzureClientSecret,
			"SUBSCRIPTIONID":                            config.AzureSubscriptionID,
			"TENANTID":                                  config.AzureTenantID,
			"SERVICEBUSNAMESPACE":                       config.AzureServiceBusNamespace,
			"RESOURCEGROUP":                             config.AzureResourceGroup,
			"AZUREBATCH_POOLID":                         config.AzureBatchPoolID,
			"AZUREBATCH_BATCHACCOUNTLOCATION":           config.AzureBatchAccountLocation,
			"AZUREBATCH_BATCHACCOUNTNAME":               config.AzureBatchAccountName,
			"HANDLER_MONGODBDOCPROVIDER_PORT":           strconv.Itoa(config.MongoDBPort),
			"HANDLER_MONGODBDOCPROVIDER_NAME":           config.MongoDBName,
			"HANDLER_MONGODBDOCPROVIDER_PASSWORD":       config.MongoDBPassword,
			"HANDLER_MONGODBDOCPROVIDER_COLLECTION":     config.MongoDBCollection,
			"HANDLER_AZUREBLOBPROVIDER_BLOBACCOUNTNAME": config.AzureStorageAccountName,
			"HANDLER_AZUREBLOBPROVIDER_BLOBACCOUNTKEY":  config.AzureStorageAccountKey,
		},
	}

	secretsClient := k.client.CoreV1().Secrets(k.namespace)
	_, err := secretsClient.Create(secret)
	if err != nil {
		return fmt.Errorf("error creating dispatcher secret %+v", err)
	}
	return nil
}

func createSharedImagePullSecret(config *Configuration) error {
	// If container registry details are provided, create a secret
	// to store them. These will then be used when fetching the module's
	// container image.
	sharedImagePullSecretName = fmt.Sprintf("imagepull-%s", mgmt.ID)
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

		moduleImagePullSecret := &apiv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: sharedImagePullSecretName,
				Labels: map[string]string{
					"createdBy": mgmt.ID,
				},
			},
			Data: map[string][]byte{
				".dockerconfigjson": []byte(dockerAuthConfig),
			},
			Type: apiv1.SecretTypeDockerConfigJson,
		}

		secretsClient := k.client.CoreV1().Secrets(k.namespace)
		_, err := secretsClient.Create(moduleImagePullSecret)
		if err != nil {
			return fmt.Errorf("error creating dispatcher secret %+v", err)
		}
	}
	return nil
}

func (s *server) Create(ctx context.Context, r *module.ModuleCreateRequest) (*module.ModuleCreateResponse, error) {
	// a unique ID for this creation
	id := fmt.Sprintf("%s-%s", r.Modulename, genID())

	// Create a configmap to store the configuration details
	// needed by the module. These will be mounted into the
	// dispatcher as a volume and then passed on when it
	// dispatches the module.
	moduleConfigMapName := id

	var buffer strings.Builder
	for k, v := range r.Configmap {
		_, _ = buffer.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	configMapStr := strings.TrimSuffix(buffer.String(), "\n")

	moduleConfigMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: moduleConfigMapName,
			Labels: map[string]string{
				"createdBy":  mgmt.ID,
				"id":         id,
				"moduleName": r.Modulename,
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

	useAzureBatchProvider := false
	if r.Provider == "azurebatch" {
		useAzureBatchProvider = true
	}

	// Create an argument list to provide the the dispatcher binary
	dispatcherArgs := []string{
		"start",
		"--modulename=" + r.Modulename,
		"--moduleconfigpath=" + fmt.Sprintf("%s/module", configMapFilePath),
		"--subscribestoevent=" + r.Eventsubscriptions,
		"--eventspublished=" + r.Eventpublications,
		"--azurebatch.enabled=" + strconv.FormatBool(useAzureBatchProvider),
		"--job.workerimage=" + r.Moduleimage + ":" + r.Moduleimagetag,
		"--job.handlerimage=" + r.Handlerimage + ":" + r.Handlerimagetag,
		"--job.retrycount=" + fmt.Sprintf("%d", r.Retrycount),
		"--job.pullalways=false",
		"--kubernetes.namespace=" + k.namespace,
		"--kubernetes.imagepullsecretname=" + sharedImagePullSecretName,
		"--loglevel=" + logLevel,
	}

	dispatcherDeploymentName := id

	// Create a deployment that runs a dispatcher
	// pod, passing in environment variables from
	// a secret and mounting a volume from a configmap.
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: dispatcherDeploymentName,
			Labels: map[string]string{
				"createdBy":  mgmt.ID,
				"id":         dispatcherDeploymentName,
				"moduleName": r.Modulename,
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
							Image: fmt.Sprintf("%s:%s", mgmt.DispatcherImageName, mgmt.DispatcherImageTag),
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

func (s *server) Delete(ctx context.Context, r *module.ModuleDeleteRequest) (*module.ModuleDeleteResponse, error) {
	// Find deployments with matching label and delete them
	deploymentsClient := k.client.AppsV1().Deployments(k.namespace)
	deployments, err := deploymentsClient.List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("id=%s", r.Name),
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
		LabelSelector: fmt.Sprintf("id=%s", r.Name),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing configmaps with name %s", r.Name)
	}
	for _, configmap := range configmaps.Items {
		if err := configMapClient.Delete(configmap.Name, nil); err != nil {
			return nil, fmt.Errorf("error deleting configmap %s", configmap.Name)
		}
	}

	// Find the secrets with the matching label and delete them
	secretsClient := k.client.CoreV1().Secrets(k.namespace)
	secrets, err := secretsClient.List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("id=%s", r.Name),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing configmaps with name %s", r.Name)
	}
	for _, secret := range secrets.Items {
		if err := secretsClient.Delete(secret.Name, nil); err != nil {
			return nil, fmt.Errorf("error deleting secret %s", secret.Name)
		}
	}

	var deleteResponse = &module.ModuleDeleteResponse{
		Name: r.Name,
	}

	return deleteResponse, nil
}

func (s *server) List(ctx context.Context, r *module.ModuleListRequest) (*module.ModuleListResponse, error) {
	// Find deployments with matching label
	deploymentsClient := k.client.AppsV1().Deployments(k.namespace)
	deployments, err := deploymentsClient.List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("createdBy=%s", mgmt.ID),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing deployments with label %s", mgmt.ID)
	}
	var list = &module.ModuleListResponse{}
	for _, deployment := range deployments.Items {
		list.Names = append(list.Names, deployment.Name)
	}
	return list, nil
}

func (s *server) Get(ctx context.Context, r *module.ModuleGetRequest) (*module.ModuleGetResponse, error) {
	// Get the deployment with the given name
	deploymentsClient := k.client.AppsV1().Deployments(k.namespace)
	deployments, err := deploymentsClient.Get(r.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting deployments with name %s", r.Name)
	}
	var getResponse = &module.ModuleGetResponse{
		Name: deployments.Name,
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
