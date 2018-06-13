package management

import (
	"encoding/base64"
	"fmt"
	pb "github.com/lawrencegripper/ion/internal/app/management/protobuf"
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
	"net"
	"os"
	"path/filepath"
	"strconv"
)

type server struct{}

type Kubernetes struct {
	client                    *kubernetes.Clientset
	namespace                 string
	AzureSPSecretRef          string
	AzureBlobStorageSecretRef string
	AzureServiceBusSecretRef  string
	MongoDBSecretRef          string
}

type DispatcherMeta struct {
	ImageName string
	ImageTag  string
	ID        string
}

var k Kubernetes
var dispatcher *DispatcherMeta
var empty pb.Empty

func genID() string {
	id := xid.New()
	return id.String()
}

// Run the GRPC server
func Run(config *configuration) error {

	dispatcher := &DispatcherMeta{
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

	secretsClient := k.client.CoreV1().Secrets(k.namespace)

	dispatcherSecret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("dispatcher-%s", dispatcher.ID),
		},
		StringData: map[string]string{
			"AZURE_CLIENT_ID":              config.AzureClientID,
			"AZURE_CLIENT_SECRET":          config.AzureClientSecret,
			"AZURE_SUBSCRIPTION_ID":        config.AzureSubscriptionID,
			"AZURE_TENANT_ID":              config.AzureTenantID,
			"AZURE_SERVICEBUS_NAMESPACE":   config.AzureServiceBusNamespace,
			"AZURE_RESOURCE_GROUP":         config.AzureResourceGroup,
			"AZURE_BATCH_POOLID":           config.AzureBatchPoolID,
			"AZURE_AD_RESOURCE":            config.AzureADResource,
			"AZURE_BATCH_ACCOUNT_LOCATION": config.AzureBatchAccountLocation,
			"AZURE_BATCH_ACCOUNT_NAME":     config.AzureBatchAccountName,
			"MONGODB_PORT":                 strconv.Itoa(config.MongoDBPort),
			"MONGODB_NAME":                 config.MongoDBName,
			"MONGODB_PASSWORD":             config.MongoDBPassword,
			"MONGODB_COLLECTION":           config.MongoDBCollection,
			"AZURE_STORAGE_ACCOUNT_NAME":   config.AzureStorageAccountName,
			"AZURE_STORAGE_ACCOUNT_KEY":    config.AzureStorageAccountKey,
		},
	}

	_, err = secretsClient.Create(dispatcherSecret)
	if err != nil {
		return fmt.Errorf("error creating dispatcher secret %+v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterModuleServiceServer(s, &server{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func (s *server) Create(ctx context.Context, r *pb.ModuleCreateRequest) (*pb.Empty, error) {

	// Create Module pull secret
	auth := encodeBase64(fmt.Sprintf("%s:%s", r.Containerregistryusername, r.Containerregistrypassword))
	dockerAuthConfig := fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"%s","email":"%s","auth":"%s"}}}`,
		r.Containerregistryurl,
		r.Containerregistryusername,
		r.Containerregistrypassword,
		r.Containerregistryemail,
		auth)

	moduleImagePullSecret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("module-secret-%s", dispatcher.ID),
		},
		Data: map[string][]byte{
			".dockerconfigjson": []byte(dockerAuthConfig),
		},
		Type: apiv1.SecretTypeDockerConfigJson,
	}

	secretsClient := k.client.CoreV1().Secrets(k.namespace)
	_, err := secretsClient.Create(moduleImagePullSecret)
	if err != nil {
		return &empty, fmt.Errorf("error creating dispatcher secret %+v", err)
	}

	// Create Module config map
	moduleConfigMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("module-config-%s", dispatcher.ID),
		},
		Data: r.Configmap,
	}

	configMapClient := k.client.CoreV1().ConfigMaps(k.namespace)
	_, err = configMapClient.Create(moduleConfigMap)
	if err != nil {
		return &empty, fmt.Errorf("error creating module config map %+v", err)
	}

	configMapFilePath := "/etc/config/"

	// Create Module arguments
	dispatcherArgs := []string{
		"--modulename=" + r.Name,
		"--subscribestoevent=" + r.Subsribestoevent,
		"--eventspublished=" + r.Eventspublished,
		"--job.workerimage=" + r.Moduleimage + ":" + r.Moduleimagetag,
		"--job.handlerimage=" + r.Handlerimage + ":" + r.Handlerimagetag,
		"--job.retrycount=" + fmt.Sprintf("%s", r.Retrycount),
		"--job.pullalways=false",
		"--kubernetes.namespace=" + k.namespace,
		"--Kubernetes.imagepullsecretname=" + fmt.Sprintf("module-secret-%s", dispatcher.ID),
		"--moduleconfigpath=" + configMapFilePath,
		"--loglevel=warn",
	}

	// Create Dispatcher Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("module-%s-%s", r.Name, dispatcher.ID),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(r.InstanceCount),
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "ion-dispatcher",
							Image: fmt.Sprintf("%s:%s", dispatcher.ImageName, dispatcher.ImageTag),
							Args:  dispatcherArgs,
							EnvFrom: []apiv1.EnvFromSource{
								{
									ConfigMapRef: &apiv1.ConfigMapEnvSource{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: fmt.Sprintf("dispatcher-%s", dispatcher.ID),
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
										Name: fmt.Sprintf("module-config-%s", dispatcher.ID),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Dispatcher Deployment
	deploymentsClient := k.client.AppsV1().Deployments(k.namespace)
	_, err = deploymentsClient.Create(deployment)
	if err != nil {
		return &empty, fmt.Errorf("error creating dispatcher deployment %+v", err)
	}

	return &empty, nil
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
