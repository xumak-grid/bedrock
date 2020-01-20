package gogs

import (
	"fmt"
	"os"

	"github.com/xumak-grid/bedrock"
	"github.com/xumak-grid/bedrock/ingress"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var labels = map[string]string{
	"app":   ServerName,
	"stack": "bedrock",
}

var automountServiceAccount = false

const (
	// EnvSocatLink socat link config for gogs.
	EnvSocatLink = "SOCAT_LINK"
	// PodInfoVolName is the name of the volume for injecting metadata into the pod.
	PodInfoVolName = "podinfo"
	//GogsPort running port
	GogsPort = 3000
	// ServerName the server name for gogs deployment
	ServerName = "gogs-server"
	// ServiceName the service name for gogs deployment
	ServiceName = "gogs-srvc"
	// IngressName the ingress name for gogs
	IngressName = "gogs-ingress"
	// InitJobName the name of the init k8s job
	InitJobName = "gogs-init-job"
	// initJobImage the k8s job image
	initJobImage = "grid/init-gogs:1.0.0"
	// InitSecretName the name of the secret
	InitSecretName = "gogs-init-config"
)

// Vendor represents the vendor for gogs and contains the images available to deploy
func Vendor() bedrock.Vendor {
	return bedrock.Vendor{
		Name: "gogs",
		Images: []bedrock.Image{
			bedrock.Image{
				Name: "grid/gogs:0.11.34",
			},
		},
	}
}

func volumes() []v1.Volume {
	podInfoVol := v1.Volume{
		Name: PodInfoVolName,
	}
	podInfoVol.DownwardAPI = &v1.DownwardAPIVolumeSource{
		Items: []v1.DownwardAPIVolumeFile{
			v1.DownwardAPIVolumeFile{
				Path: "labels.properties",
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "metadata.labels",
				},
			},
		},
	}
	vols := []v1.Volume{
		podInfoVol,
	}
	return vols

}

func server(gogsImage, namespace string) v1.Container {
	return v1.Container{
		Name:  "gogs",
		Image: fmt.Sprintf("%s/%s", bedrock.GridDockerRepository(), gogsImage),
		Env: []v1.EnvVar{
			v1.EnvVar{
				Name:  EnvSocatLink,
				Value: "false",
			},
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      PodInfoVolName,
				MountPath: "/meta",
				ReadOnly:  false,
			},
		},
		Ports: []v1.ContainerPort{
			v1.ContainerPort{
				ContainerPort: GogsPort,
				Protocol:      v1.ProtocolTCP,
			},
		},
	}
}

// Deployment returns a deployment configuration for Gogs.
func Deployment(gogsImage, namespace string) *appsv1beta1.Deployment {
	deployment := &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "gogs-server",
			Labels: labels,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "gogs-server",
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						server(gogsImage, namespace),
					},
					Volumes: volumes(),
					AutomountServiceAccountToken: &automountServiceAccount,
				},
			},
		},
	}
	return deployment
}

// StatefulSet to manage gogs server
func StatefulSet(gogsImage, namespace string) *appsv1beta2.StatefulSet {
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
						server(gogsImage, namespace),
					},
					Volumes: volumes(),
					AutomountServiceAccountToken: &automountServiceAccount,
				},
			},
		},
	}
	sfs.Spec.Template.Labels = labels
	return &sfs
}

//Service returns the service configuration for gogs.
func Service(namespace string) *v1.Service {
	selector := map[string]string{
		"app":   ServerName,
		"stack": "bedrock",
	}
	service := &v1.Service{
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
					TargetPort: intstr.FromInt(GogsPort),
				},
			},
			Type:     v1.ServiceTypeClusterIP,
			Selector: selector,
		},
	}
	return service
}

// Ingress returns a k8s ingress for gogs
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
		ingress.Annotations(class, class),
		80,
	)
}

// Secret is a k8s secret that includes the configuration to initial setup of gogs server
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

// InitJob is an k8s job to make a custom setup to gogs server
// the job uses a secret to obtain the init configuration of the gogs
func InitJob(gogsHost, namespace string) *batchv1.Job {

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
									Name:  "GOGS_HOST",
									Value: gogsHost,
								},
								v1.EnvVar{
									Name:  "GOGS_CONFIG_FILE",
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
