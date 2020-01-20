package nexus

import (
	"fmt"
	"os"

	"github.com/xumak-grid/bedrock"
	"github.com/xumak-grid/bedrock/ingress"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// NexusPort running port
	NexusPort = 8081
	// ServerName the server name for nexus deployment
	ServerName = "nexus-server"
	// ServiceName the service name for nexus deployment
	ServiceName = "nexus-srvc"
	// IngressName the ingress name for nexus
	IngressName = "nexus-ingress"
	// InitJobName the name of the init k8s job
	InitJobName = "nexus-init-job"
	// initJobImage the k8s job image
	initJobImage = "grid/init-nexus:1.0.0"
	// InitSecretName the name of the secret
	InitSecretName = "nexus-init-config"
)

var labels = map[string]string{
	"app":   ServerName,
	"stack": "bedrock",
}

var automountServiceAccount = false

// Vendor represents the vendor for nexus and contains the images available to deploy
func Vendor() bedrock.Vendor {
	return bedrock.Vendor{
		Name: "nexus",
		Images: []bedrock.Image{
			bedrock.Image{
				Name: "grid/nexus:3.8.0",
			},
			bedrock.Image{
				Name: "grid/nexus:3.9.0",
			},
			bedrock.Image{
				Name: "grid/nexus:3.12.0",
			},
		},
	}
}
func server(nexusImage, namespace string) v1.Container {
	return v1.Container{
		Name:  ServerName,
		Image: fmt.Sprintf("%s/%s", bedrock.GridDockerRepository(), nexusImage),
		Ports: []v1.ContainerPort{
			v1.ContainerPort{
				ContainerPort: NexusPort,
				Protocol:      v1.ProtocolTCP,
			},
		},
	}
}

// Deployment returns a deployment configuration for Nexus.
func Deployment(projectName, nexusImage, namespace string) *appsv1beta1.Deployment {
	deployment := &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ServerName,
			Labels: labels,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   ServerName,
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						server(nexusImage, namespace),
					},
					AutomountServiceAccountToken: &automountServiceAccount,
				},
			},
		},
	}
	return deployment
}

// StatefulSet to manage nexus server
func StatefulSet(nexusImage, namespace string) *appsv1beta2.StatefulSet {
	replicas := int32(1)
	sfs := appsv1beta2.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ServerName,
			Labels: labels,
		},
		Spec: appsv1beta2.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas:    &replicas,
			ServiceName: ServiceName,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						server(nexusImage, namespace),
					},
					AutomountServiceAccountToken: &automountServiceAccount,
				},
			},
		},
	}
	sfs.Spec.Template.Labels = labels
	return &sfs
}

// Service returns the service configuration for nexus.
func Service(namespace string) *v1.Service {
	selector := map[string]string{
		"app":   ServerName,
		"stack": "bedrock",
	}
	src := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ServiceName,
			Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "http",
					Port:       80,
					Protocol:   v1.ProtocolTCP,
					TargetPort: intstr.FromInt(NexusPort),
				},
			},
			Type:     v1.ServiceTypeClusterIP,
			Selector: selector,
		},
	}
	return src
}

// Ingress returns a k8s ingress for nexus
func Ingress(namespace string) *v1beta1.Ingress {

	class := os.Getenv(ingress.ClassEnvVar)
	if class == "" {
		class = ingress.DefaultIngressClass()
	}
	redirect := os.Getenv(ingress.SSLRedirectEnvVar)
	if redirect == "" {
		redirect = ingress.DefaultSSLRedirect()
	}

	return ingress.New(
		IngressName,
		namespace,
		fmt.Sprintf("%s-%s.%s", ServerName, namespace, bedrock.GridExternalDomain()),
		ServiceName,
		namespace+"-public-tls",
		labels,
		ingress.Annotations(redirect, class),
		80,
	)
}

// Secret is a k8s secret that includes the configuration to initial setup of nexus server
func Secret(namespace string, secretData []byte) *v1.Secret {

	s := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      InitSecretName,
			Namespace: namespace,
			Labels:    labels,
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"configFile.json": secretData,
		},
	}
	return s
}

// InitJob is an k8s job to make a custom setup to nexus server
// the job uses a secret to obtain the init configuration of the nexus
func InitJob(nexusHost, namespace string) *batchv1.Job {

	backofflimit := int32(3)

	j := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      InitJobName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backofflimit,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					AutomountServiceAccountToken: &automountServiceAccount,
					RestartPolicy:                v1.RestartPolicyNever,
					Containers: []v1.Container{
						v1.Container{
							Name:            InitJobName,
							Image:           fmt.Sprintf("%s/%s", bedrock.GridDockerRepository(), initJobImage),
							ImagePullPolicy: v1.PullAlways,
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "NEXUS_USER",
									Value: "admin",
								},
								v1.EnvVar{
									Name:  "NEXUS_PASS",
									Value: "admin123",
								},
								v1.EnvVar{
									Name:  "NEXUS_HOST",
									Value: nexusHost,
								},
								v1.EnvVar{
									Name:  "NEXUS_CONFIG_FILE",
									Value: "/app/config/configFile.json",
								},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "init-config",
									MountPath: "/app/config",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "init-config",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: InitSecretName,
								},
							},
						},
					},
				},
			},
		},
	}
	return j
}
