package k8s

import (
	"errors"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateSecret create a k8s secret
func CreateSecret(kubecli kubernetes.Interface, namespace string, secret *v1.Secret) (*v1.Secret, error) {
	scrt, err := kubecli.CoreV1().Secrets(namespace).Create(secret)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil, errors.New("secret already exists")
		}
		return nil, err
	}
	return scrt, nil
}

// DeleteSecret delete a k8s secret
func DeleteSecret(kubecli kubernetes.Interface, namespace, secretName string) error {
	return kubecli.CoreV1().Secrets(namespace).Delete(secretName, &metav1.DeleteOptions{})
}

// GetSecret get a k8s secret
func GetSecret(kubecli kubernetes.Interface, namespace, secretName string) (*v1.Secret, error) {
	return kubecli.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
}
