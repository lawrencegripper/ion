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

// DispatcherMeta holds metadata used by the dispatcher
type DispatcherMeta struct {
	ImageName string
	ImageTag  string
	ID        string
}

var k Kubernetes
var dispatcher *DispatcherMeta
var dispatcherSecretName string
var logLevel string

func genID() string {
	id := xid.New()
	return id.String()
}

// Run the GRPC server
func Run(config *Configuration) error {

	logLevel = config.LogLevel

	dispatcher = &DispatcherMeta{
		ID:        genID(),
		ImageName: config.DispatcherImage,
		ImageTag:  config.DispatcherImageTag,
	}

	var err error
	k.client, err = getClientSet()
	if err != nil {
		return fmt.Errorf("error connecting to Kubernetes %+v", err)
	}
	k.namespace = config.Namespace

	err = createDispatcherSecret(config)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	module.RegisterModuleServiceServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func createDispatcherSecret(config *Configuration) error {
	secretsClient := k.client.CoreV1().Secrets(k.namespace)

	dispatcherSecretName = fmt.Sprintf("dispatcher-%s", dispatcher.ID)

	dispatcherSecret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: dispatcherSecretName,
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

	_, err := secretsClient.Create(dispatcherSecret)
	if err != nil {
		return fmt.Errorf("error creating dispatcher secret %+v", err)
	}
	return nil
}

func (s *server) Create(ctx context.Context, r *module.ModuleCreateRequest) (*module.ModuleCreateRequest, error) {

	// If container registry details are provided, create a secret
	// to store them. These will then be used when fetching the module's
	// container image.
	moduleImagePullSecretName := ""
	if r.Containerregistrypassword != "" &&
		r.Containerregistryusername != "" &&
		r.Containerregistryurl != "" &&
		r.Containerregistryemail != "" {

		moduleImagePullSecretName = fmt.Sprintf("module-image-secret-%s", dispatcher.ID)

		auth := encodeBase64(fmt.Sprintf("%s:%s", r.Containerregistryusername, r.Containerregistrypassword))
		dockerAuthConfig := fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"%s","email":"%s","auth":"%s"}}}`,
			r.Containerregistryurl,
			r.Containerregistryusername,
			r.Containerregistrypassword,
			r.Containerregistryemail,
			auth)

		moduleImagePullSecret := &apiv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: moduleImagePullSecretName,
			},
			Data: map[string][]byte{
				".dockerconfigjson": []byte(dockerAuthConfig),
			},
			Type: apiv1.SecretTypeDockerConfigJson,
		}

		secretsClient := k.client.CoreV1().Secrets(k.namespace)
		_, err := secretsClient.Create(moduleImagePullSecret)
		if err != nil {
			return nil, fmt.Errorf("error creating dispatcher secret %+v", err)
		}
	}

	// Create a configmap to store the configuration details
	// needed by the module. These will be mounted into the
	// dispatcher as a volume and then passed on when i
	// dispatches the module.
	moduleConfigMapName := fmt.Sprintf("module-config-%s", dispatcher.ID)

	var buffer strings.Builder
	for k, v := range r.Configmap {
		buffer.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	configMapStr := strings.TrimSuffix(buffer.String(), "\n")

	moduleConfigMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: moduleConfigMapName,
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

	// Create an argument list to provide the the dispatcher
	dispatcherArgs := []string{
		"start",
		"--modulename=" + r.Name,
		"--moduleconfigpath=" + fmt.Sprintf("%s/module", configMapFilePath),
		"--subscribestoevent=" + r.Eventsubscriptions,
		"--eventspublished=" + r.Eventpublications,
		"--job.workerimage=" + r.Moduleimage + ":" + r.Moduleimagetag,
		"--job.handlerimage=" + r.Handlerimage + ":" + r.Handlerimagetag,
		"--job.retrycount=" + fmt.Sprintf("%d", r.Retrycount),
		"--job.pullalways=false",
		"--kubernetes.namespace=" + k.namespace,
		"--kubernetes.imagepullsecretname=" + moduleImagePullSecretName,
		"--loglevel=" + logLevel,
	}

	dispatcherDeploymentName := fmt.Sprintf("dispatcher-%s-%s", r.Name, dispatcher.ID)

	// Create a deployment that runs a dispatcher
	// pod, passing in environment variables from
	// a secret and mounting a volume from a configmap.
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: dispatcherDeploymentName,
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
							Image: fmt.Sprintf("%s:%s", dispatcher.ImageName, dispatcher.ImageTag),
							Args:  dispatcherArgs,
							EnvFrom: []apiv1.EnvFromSource{
								{
									SecretRef: &apiv1.SecretEnvSource{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: dispatcherSecretName,
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

	return r, nil
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
