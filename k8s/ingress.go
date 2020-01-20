package k8s

import (
	"errors"

	v1beta1 "k8s.io/api/extensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateIngress create an ingress
func CreateIngress(kubecli kubernetes.Interface, namespace string, ingress *v1beta1.Ingress) (*v1beta1.Ingress, error) {
	ingress, err := kubecli.ExtensionsV1beta1().Ingresses(namespace).Create(ingress)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil, errors.New("ingress already exists")
		}
		return nil, err
	}
	return ingress, nil
}

// DeleteIngress delete an ingress
func DeleteIngress(kubecli kubernetes.Interface, namespace, ingressName string) error {
	policy := metav1.DeletePropagationBackground
	ops := &metav1.DeleteOptions{
		PropagationPolicy: &policy,
	}
	return kubecli.ExtensionsV1beta1().Ingresses(namespace).Delete(ingressName, ops)
}

// GetIngress get an ingress
func GetIngress(kubecli kubernetes.Interface, namespace, ingressName string) (*v1beta1.Ingress, error) {
	return kubecli.ExtensionsV1beta1().Ingresses(namespace).Get(ingressName, metav1.GetOptions{})
}

// UpdateIngress update an ingress
func UpdateIngress(kubecli kubernetes.Interface, namespace string, ingress *v1beta1.Ingress) (*v1beta1.Ingress, error) {
	return kubecli.ExtensionsV1beta1().Ingresses(namespace).Update(ingress)
}
