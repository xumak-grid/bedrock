// Package ingress provides k8s ingress related
package ingress

import (
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// ClassEnvVar represents the ingress.class of the ingress controller (nginx or contour)
	ClassEnvVar = "INGRESS_CLASS"
	// SSLRedirectEnvVar if the controller will force to ssl redirect (true or false)
	SSLRedirectEnvVar = "INGRESS_FORCE_SSL_REDIRECT"
)

// Annotations returns the annotations of an ingress base on the parameters
func Annotations(forceSSLRedirect string, ingressClass string) map[string]string {
	return map[string]string{
		"ingress.kubernetes.io/force-ssl-redirect": forceSSLRedirect,
		"kubernetes.io/ingress.class":              ingressClass,
	}
}

// New returns a k8s ingress base on the parameters
func New(name, ns, host, serviceName, secretName string, labels, annotations map[string]string, servicePort int) *v1beta1.Ingress {
	return &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				v1beta1.IngressRule{
					Host: host,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								v1beta1.HTTPIngressPath{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: serviceName,
										ServicePort: intstr.FromInt(servicePort),
									},
								},
							},
						},
					},
				},
			},
			TLS: []v1beta1.IngressTLS{
				v1beta1.IngressTLS{
					Hosts: []string{
						host,
					},
					SecretName: secretName,
				},
			},
		},
	}
}

// DefaultIngressClass returns the default ingress class
func DefaultIngressClass() string {
	return "nginx"
}

// DefaultSSLRedirect returns the default ingress ssl redirect option
func DefaultSSLRedirect() string {
	return "true"
}
