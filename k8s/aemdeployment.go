package k8s

import (
	aemv1beta1 "github.com/xumak-grid/aem-operator/pkg/apis/aem/v1beta1"
	aemclientset "github.com/xumak-grid/aem-operator/pkg/generated/clientset/versioned"
	"github.com/xumak-grid/bedrock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// CreateAEMDeployment creates an aem deployment.
func CreateAEMDeployment(cli aemclientset.Interface, aemDep *bedrock.AEMDeployment) (*aemv1beta1.AEMDeployment, error) {
	k8sdep := &aemv1beta1.AEMDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        aemDep.EnvironmentID,
			Namespace:   aemDep.ClientID,
			Annotations: gridLabels,
		},
		Spec: aemv1beta1.AEMDeploymentSpec{
			Version:           aemDep.Spec.Version,
			DispatcherVersion: aemDep.Spec.DispatcherVersion,
			Authors: aemv1beta1.InstanceSpec{
				Replicas: aemDep.Spec.Authors.Replicas,
				Type:     aemDep.Spec.Authors.Type,
			},
			Publishers: aemv1beta1.InstanceSpec{
				Replicas: aemDep.Spec.Publishers.Replicas,
				Type:     aemDep.Spec.Publishers.Type,
			},
			Dispatchers: aemv1beta1.InstanceSpec{
				Replicas: aemDep.Spec.Dispatchers.Replicas,
				Type:     aemDep.Spec.Dispatchers.Type,
			},
		},
	}
	return cli.AemV1beta1().AEMDeployments(aemDep.ClientID).Create(k8sdep)
}

// DeleteAEMDeployment deletes the aem deployment base on the clientID and the environment
func DeleteAEMDeployment(cli aemclientset.Interface, aemDep *bedrock.AEMDeployment) error {
	return cli.AemV1beta1().AEMDeployments(aemDep.ClientID).Delete(aemDep.EnvironmentID, &metav1.DeleteOptions{})
}

// GetAEMDeployment returns and AEM deployment base on the bedrock AEM deployment
func GetAEMDeployment(cli aemclientset.Interface, aemDep *bedrock.AEMDeployment) (*aemv1beta1.AEMDeployment, error) {
	return cli.AemV1beta1().AEMDeployments(aemDep.ClientID).Get(aemDep.EnvironmentID, metav1.GetOptions{})
}

// UpdateAEMDeployment updates the aem deployment
func UpdateAEMDeployment(cli aemclientset.Interface, aemDep *bedrock.AEMDeployment) error {
	k8sDep, err := GetAEMDeployment(cli, aemDep)
	if err != nil {
		return err
	}
	k8sDep.Spec = aemv1beta1.AEMDeploymentSpec{
		Version:           aemDep.Spec.Version,
		DispatcherVersion: aemDep.Spec.DispatcherVersion,
		Authors: aemv1beta1.InstanceSpec{
			Replicas: aemDep.Spec.Authors.Replicas,
			Type:     aemDep.Spec.Authors.Type,
		},
		Publishers: aemv1beta1.InstanceSpec{
			Replicas: aemDep.Spec.Publishers.Replicas,
			Type:     aemDep.Spec.Publishers.Type,
		},
		Dispatchers: aemv1beta1.InstanceSpec{
			Replicas: aemDep.Spec.Dispatchers.Replicas,
			Type:     aemDep.Spec.Dispatchers.Type,
		},
	}

	_, err = cli.AemV1beta1().AEMDeployments(aemDep.ClientID).Update(k8sDep)
	if err != nil {
		return err
	}
	return nil
}

func ListAEMDeployments(cli aemclientset.Interface, ns string) ([]aemv1beta1.AEMDeployment, error) {
	deployments, err := cli.AemV1beta1().AEMDeployments(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return deployments.Items, nil
}

// ListAEMDeploymentPods lists the PODs from the AEM deployment
func ListAEMDeploymentPods(cli kubernetes.Interface, aemDep *bedrock.AEMDeployment) ([]v1.Pod, error) {
	podLabels := map[string]string{
		"app":        "aem",
		"deployment": aemDep.EnvironmentID,
	}
	selector := labels.SelectorFromSet(podLabels)
	pods, err := cli.CoreV1().Pods(aemDep.ClientID).List(metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil

}
