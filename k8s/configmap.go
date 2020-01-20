package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetConfigMap get a k8s configMap
func GetConfigMap(kubecli kubernetes.Interface, ns, cmapName string) (*v1.ConfigMap, error) {
	return kubecli.CoreV1().ConfigMaps(ns).Get(cmapName, metav1.GetOptions{})
}

// UpdateConfigMap update the data of a k8s configMap
func UpdateConfigMap(kubecli kubernetes.Interface, ns, cmapName string, data map[string]string) error {
	cmap, err := GetConfigMap(kubecli, ns, cmapName)
	if err != nil {
		return err
	}

	cmap.Data = data
	_, err = kubecli.CoreV1().ConfigMaps(ns).Update(cmap)
	if err != nil {
		return err
	}
	return nil
}
