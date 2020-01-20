package drone

import (
	"fmt"
	"os"

	"github.com/xumak-grid/bedrock"
	"github.com/xumak-grid/bedrock/ingress"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
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

// Constants
const (
	EnvDroneOpen        = "DRONE_OPEN"
	EnvDroneGogs        = "DRONE_GOGS"
	EnvDroneGogsURL     = "DRONE_GOGS_URL"
	EnvDroneDebug       = "DRONE_DEBUG"
	EnvDroneSecret      = "DRONE_SECRET"
	EnvDroneServer      = "DRONE_SERVER"
	EnvDockerAPIVersion = "DOCKER_API_VERSION"
	// DronePort running port
	DronePort = 8000
	// DroneServerPort internal agent-server communication
	DroneServerPort = 9000
	PodInfoVolName  = "podinfo"
	PodVolumeDind   = "dind-socket"
	PodVolumeDB     = "drone-server-sqlite-db"
	// ServerName the server name for drone deployment
	ServerName = "drone-server"
	// AgentName the agent name for drone deployment
	AgentName = "drone-agent"
	// ServiceName the service name for drone deployment
	ServiceName = "drone-srvc"
	// IngressName the ingress name for drone
	IngressName = "drone-ingress"
)

// Vendor represents the vendor for drone and contains the images available to deploy
func Vendor() bedrock.Vendor {
	return bedrock.Vendor{
		Name: "drone",
		Images: []bedrock.Image{
			bedrock.Image{
				Name:      "grid/drone:0.8-alpine",
				Secondary: "grid/drone-agent:0.8",
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
	dindVol := v1.Volume{
		Name: PodVolumeDind,
	}
	dindVol.HostPath = &v1.HostPathVolumeSource{
		Path: "/var/run/docker.sock",
	}

	dbVol := v1.Volume{
		Name: PodVolumeDB,
	}
	dbVol.EmptyDir = &v1.EmptyDirVolumeSource{}

	vols := []v1.Volume{
		podInfoVol,
		dindVol,
		dbVol,
	}
	return vols

}
func server(secret, gogsURL, droneHost, serverImage string) v1.Container {
	return v1.Container{
		Name:  ServerName,
		Image: fmt.Sprintf("%s/%s", bedrock.GridDockerRepository(), serverImage),
		Env: []v1.EnvVar{
			v1.EnvVar{
				Name:  EnvDroneOpen,
				Value: "true",
			},
			v1.EnvVar{
				Name:  EnvDroneDebug,
				Value: "true",
			},
			v1.EnvVar{
				Name:  EnvDroneSecret,
				Value: secret,
			},
			v1.EnvVar{
				Name:  EnvDroneGogs,
				Value: "true",
			},
			v1.EnvVar{
				Name:  EnvDroneGogsURL,
				Value: gogsURL,
			},
			v1.EnvVar{
				Name:  EnvDockerAPIVersion,
				Value: "1.23",
			},
			v1.EnvVar{
				Name:  "DRONE_HOST",
				Value: droneHost,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      PodInfoVolName,
				MountPath: "/meta",
				ReadOnly:  false,
			},
			v1.VolumeMount{
				Name:      PodVolumeDB,
				MountPath: "/var/lib/drone",
			},
			v1.VolumeMount{
				Name:      PodVolumeDind,
				MountPath: "/var/run/docker.sock",
			},
		},
		Ports: []v1.ContainerPort{
			v1.ContainerPort{
				ContainerPort: DronePort,
				Protocol:      v1.ProtocolTCP,
			},
			v1.ContainerPort{
				ContainerPort: DroneServerPort,
				Protocol:      v1.ProtocolTCP,
			},
		},
	}
}

func agent(secret, agentImage string) v1.Container {
	return v1.Container{
		Name:  AgentName,
		Image: fmt.Sprintf("%s/%s", bedrock.GridDockerRepository(), agentImage),
		Env: []v1.EnvVar{
			v1.EnvVar{
				Name:  EnvDroneServer,
				Value: "localhost:9000",
			},
			v1.EnvVar{
				Name:  EnvDroneSecret,
				Value: secret,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name:      PodVolumeDind,
				MountPath: "/var/run/docker.sock",
			},
		},
	}
}

// Deployment returns a deployment configuration for Nexus.
func Deployment(gogsURL, droneHost, namespace, serverImage, agentImage string) *appsv1beta1.Deployment {
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
						server("somesecret", gogsURL, droneHost, serverImage),
						agent("somesecret", agentImage),
					},
					Volumes: volumes(),
					AutomountServiceAccountToken: &automountServiceAccount,
				},
			},
		},
	}
	return deployment
}

// StatefulSet to manage drone server
func StatefulSet(gogsURL, droneHost, namespace, serverImage, agentImage string) *appsv1beta2.StatefulSet {
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
						server("somesecret", gogsURL, droneHost, serverImage),
						agent("somesecret", agentImage),
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

// Service returns the service configuration for nexus.
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
					TargetPort: intstr.FromInt(DronePort),
				},
			},
			Type:     v1.ServiceTypeClusterIP,
			Selector: selector,
		},
	}
	return service
}

// Ingress returns a k8s ingress for drone
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
