package k8s

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	// gridLabels are labels that identified grid resources
	gridLabels = map[string]string{
		"grid": "true",
	}
)

// CreateNamespace creates a namespaces with gridLabels
func CreateNamespace(kubecli kubernetes.Interface, name string, annotations map[string]string) (*v1.Namespace, error) {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      gridLabels,
			Annotations: annotations,
		},
	}
	return kubecli.CoreV1().Namespaces().Create(ns)
}

// GetNamespaces lists all namespaces with gridLabels
func GetNamespaces(kubecli kubernetes.Interface) ([]v1.Namespace, error) {
	ops := metav1.ListOptions{
		LabelSelector: joinLabels(gridLabels),
	}
	l, err := kubecli.CoreV1().Namespaces().List(ops)
	namespaces := []v1.Namespace{}
	if err != nil {
		return []v1.Namespace{}, err
	}
	namespaces = l.Items
	return namespaces, nil
}

// GetNamespace returns a namespace with gridLabels
func GetNamespace(kubecli kubernetes.Interface, name string) (*v1.Namespace, error) {
	namespaces, err := GetNamespaces(kubecli)
	if err != nil {
		return nil, err
	}
	for _, ns := range namespaces {
		if ns.Name == name {
			return &ns, nil
		}
	}
	return nil, fmt.Errorf("namespace \"%v\" not found", name)
}

// DeleteNamespace deletes a namespace with gridLabels
func DeleteNamespace(kubecli kubernetes.Interface, name string) error {
	namespace, err := GetNamespace(kubecli, name)
	if err != nil {
		return err
	}
	ops := &metav1.DeleteOptions{}
	return kubecli.CoreV1().Namespaces().Delete(namespace.Name, ops)
}

// joinLabels joins key/value from m in format "key1=value1,key2=valu2"
func joinLabels(m map[string]string) string {
	str := ""
	for k, v := range m {
		str += fmt.Sprintf("%v=%v,", k, v)
	}
	str = strings.Trim(str, ",")
	return str
}
