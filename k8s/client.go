package k8s

import (
	"errors"
	"net"
	"os"
	"path/filepath"

	certclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	aemclientset "github.com/xumak-grid/aem-operator/pkg/generated/clientset/versioned"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// MustNewKubeClient ...
func MustNewKubeClient() kubernetes.Interface {
	cfg, err := OutClusterConfig()
	if err != nil {
		panic(err)
	}
	return kubernetes.NewForConfigOrDie(cfg)
}

func NewKubeClient(cfg *rest.Config) kubernetes.Interface {
	return kubernetes.NewForConfigOrDie(cfg)
}

// InClusterConfig gets a rest config based on the service account token injected by k8s
func InClusterConfig() (*rest.Config, error) {
	// Work around https://github.com/kubernetes/kubernetes/issues/40973
	// See https://github.com/coreos/etcd-operator/issues/731#issuecomment-283804819
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		addrs, err := net.LookupHost("kubernetes.default.svc")
		if err != nil {
			panic(err)
		}
		os.Setenv("KUBERNETES_SERVICE_HOST", addrs[0])
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	}
	return rest.InClusterConfig()
}

func OutClusterConfig() (*rest.Config, error) {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	return config, err
}

func CreateDeployment(kubecli kubernetes.Interface, namespace string, deployment *appsv1beta1.Deployment) (*appsv1beta1.Deployment, error) {
	deploymentsClient := kubecli.AppsV1beta1().Deployments(namespace)
	d, e := deploymentsClient.Create(deployment)
	if k8serrors.IsAlreadyExists(e) {
		return nil, errors.New("Deployment already exists")
	}
	return d, nil
}

// AEMClient creates an AEM k8s client.
func AEMClient(cfg *rest.Config) (aemclientset.Interface, error) {
	return aemclientset.NewForConfig(cfg)
}

// CertManagerClient creates a certManager k8s client.
func CertManagerClient(cfg *rest.Config) (certclient.Interface, error) {
	return certclient.NewForConfig(cfg)
}

// isRunningAndReady returns true if pod is in the PodRunning Phase, if it has a condition of PodReady.
func isRunningAndReady(pod *v1.Pod) bool {
	return IsPodRunning(pod) && IsPodReady(pod)
}

// IsPodRunning returns true if pod is in the PodRunning Phase
func IsPodRunning(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodRunning
}

// isCreated returns true if pod has been created and is maintained by the API server
func isCreated(pod *v1.Pod) bool {
	return pod.Status.Phase != ""
}

// isFailed returns true if pod has a Phase of PodFailed
func isFailed(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodFailed
}

// isTerminating returns true if pod's DeletionTimestamp has been set
func isTerminating(pod *v1.Pod) bool {
	return pod.DeletionTimestamp != nil
}

// isHealthy returns true if pod is running and ready and has not been terminated
func isHealthy(pod *v1.Pod) bool {
	return isRunningAndReady(pod) && !isTerminating(pod)
}

// IsPodReady returns true if a pod is ready; false otherwise.
func IsPodReady(pod *v1.Pod) bool {
	return IsPodReadyConditionTrue(pod.Status)
}

// IsPodReadyConditionTrue retruns true if a pod is ready; false otherwise.
func IsPodReadyConditionTrue(status v1.PodStatus) bool {
	condition := GetPodReadyCondition(status)
	return condition != nil && condition.Status == v1.ConditionTrue
}

// GetPodReadyCondition extracts the pod ready condition from the given status and returns that.
// Returns nil if the condition is not present.
func GetPodReadyCondition(status v1.PodStatus) *v1.PodCondition {
	_, condition := GetPodCondition(&status, v1.PodReady)
	return condition
}

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetPodCondition(status *v1.PodStatus, conditionType v1.PodConditionType) (int, *v1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return i, &status.Conditions[i]
		}
	}
	return -1, nil
}

// BuildKubeConfig allows to build a k8s kubernetes from a path or in cluster config
// it checks if the KUBECONFIG envVar is set, for example ~/.kube/config
func BuildKubeConfig() (*rest.Config, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return InClusterConfig()
}
