package k8s

import (
	"errors"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateService create the service
func CreateService(kubecli kubernetes.Interface, namespace string, service *v1.Service) (*v1.Service, error) {
	srvc, err := kubecli.CoreV1().Services(namespace).Create(service)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil, errors.New("Service already exists")
		}
		return nil, err
	}
	return srvc, nil
}

// DeleteService delete a service
func DeleteService(kubecli kubernetes.Interface, namespace, serviceName string) error {
	return kubecli.CoreV1().Services(namespace).Delete(serviceName, &metav1.DeleteOptions{})
}

// GetService get a service
func GetService(kubecli kubernetes.Interface, namespace, serviceName string) (*v1.Service, error) {
	return kubecli.CoreV1().Services(namespace).Get(serviceName, metav1.GetOptions{})
}
