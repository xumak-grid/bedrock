package k8s

import (
	"errors"

	v1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateJob create a k8s job
func CreateJob(kubecli kubernetes.Interface, namespace string, job *v1.Job) (*v1.Job, error) {
	jb, err := kubecli.BatchV1().Jobs(namespace).Create(job)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return nil, errors.New("job already exists")
		}
		return nil, err
	}
	return jb, nil
}

// DeleteJob delete a k8s job
func DeleteJob(kubecli kubernetes.Interface, namespace, jobName string) error {
	policy := metav1.DeletePropagationBackground
	ops := &metav1.DeleteOptions{
		PropagationPolicy: &policy,
	}
	return kubecli.BatchV1().Jobs(namespace).Delete(jobName, ops)
}
