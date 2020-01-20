package k8s

import (
	"errors"

	appsv1beta2 "k8s.io/api/apps/v1beta2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateStatefulSet create the statefulSet
func CreateStatefulSet(kubecli kubernetes.Interface, namespace string, statefulSet *appsv1beta2.StatefulSet) (*appsv1beta2.StatefulSet, error) {
	sfs, err := kubecli.AppsV1beta2().StatefulSets(namespace).Create(statefulSet)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil, errors.New("statefulSet already exists")
		}
		return nil, err
	}
	return sfs, nil
}

// DeleteStatefulSet delete a statefulSet
func DeleteStatefulSet(kubecli kubernetes.Interface, namespace, statefulSetName string) error {
	policy := metav1.DeletePropagationBackground
	ops := &metav1.DeleteOptions{
		PropagationPolicy: &policy,
	}
	return kubecli.AppsV1beta2().StatefulSets(namespace).Delete(statefulSetName, ops)
}

// GetStatefulSet get a statefulSet
func GetStatefulSet(kubecli kubernetes.Interface, namespace, statefulSetName string) (*appsv1beta2.StatefulSet, error) {
	return kubecli.AppsV1beta2().StatefulSets(namespace).Get(statefulSetName, metav1.GetOptions{})
}

// UpdateStatefulSet update a statefulSet
func UpdateStatefulSet(kubecli kubernetes.Interface, namespace string, statefulSet *appsv1beta2.StatefulSet) (*appsv1beta2.StatefulSet, error) {
	return kubecli.AppsV1beta2().StatefulSets(namespace).Update(statefulSet)
}
