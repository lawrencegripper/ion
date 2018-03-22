package kubernetes

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/Azure/go-autorest/autorest/to"

	"pack.ag/amqp"

	"github.com/lawrencegripper/mlops/dispatcher/types"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// Dispatch creates a job on kubernetes for the message
func Dispatch(message *amqp.Message, rootConfig *types.Configuration) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.WithError(err).Panic("Getting kubeconf from current context")
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.WithError(err).Panic("Getting clientset from config")
	}

	// create a namespace for the module
	// Todo: add regex validation to ensure namespace is valid in k8 before submitting
	// a DNS-1123 label must consist of lower case alphanumeric characters or '-', and
	// must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?'
	namespace := "mlops-" + strings.ToLower(rootConfig.ModuleName)
	n, err := clientset.CoreV1().Namespaces().Create(&apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	})

	if err != nil && !errors.IsAlreadyExists(err) {
		log.WithError(err).Panic()
	}

	log.Warn(n)

	job, err := clientset.BatchV1().Jobs(namespace).Create(&batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-deployment4",
		},
		Spec: batchv1.JobSpec{
			Completions:  to.Int32Ptr(1),
			BackoffLimit: to.Int32Ptr(1),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:    "test",
							Image:   "busybox",
							Command: []string{"echo 'hello'"},
							// Command: []string{"/bin/echo", "hello"},
						},
					},
					RestartPolicy: apiv1.RestartPolicyNever,
				},
			},
		},
	})

	if err != nil {
		log.WithError(err).Panic()
	}

	log.WithField("job", types.PrettyPrintStruct(job)).Warn("Created job")

	// pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	// if err != nil {
	// 	log.WithError(err).Panic("Getting pods from cluster")
	// }
	// log.Printf("There are %d pods in the cluster\n", len(pods.Items))

	// // Examples for error handling:
	// // - Use helper functions like e.g. errors.IsNotFound()
	// // - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	// namespace := "default"
	// pod := "example-xxxxx"
	// _, err = clientset.CoreV1().Pods(namespace).Get(pod, metav1.GetOptions{})
	// if errors.IsNotFound(err) {
	// 	fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
	// } else if statusError, isStatus := err.(*errors.StatusError); isStatus {
	// 	fmt.Printf("Error getting pod %s in namespace %s: %v\n",
	// 		pod, namespace, statusError.ErrStatus.Message)
	// } else if err != nil {
	// 	panic(err.Error())
	// } else {
	// 	fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
	// }
}
