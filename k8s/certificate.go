package k8s

import (
	"os"

	certmanager "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	certclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateCertficate creates a cerManager Certificate for all ingresses in the namespace
func CreateCertficate(kubecli certclient.Interface, ns string) (*certmanager.Certificate, error) {

	gridDomain := os.Getenv("GRID_EXTERNAL_DOMAIN")
	certManagerIssuer := os.Getenv("CERT_MANAGER_ISSUER")
	if certManagerIssuer == "" {
		certManagerIssuer = "letsencrypt-prod-dns"
	}

	cerManagerProvider := os.Getenv("CERT_MANAGER_DNS_PROVIDER")
	if cerManagerProvider == "" {
		cerManagerProvider = "prod-dns"
	}

	cert := &certmanager.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ns + "-account-certificate",
			Namespace: ns,
		},
		Spec: certmanager.CertificateSpec{
			SecretName: ns + "-public-tls",
			IssuerRef: certmanager.ObjectReference{
				Kind: "ClusterIssuer",
				Name: certManagerIssuer,
			},
			CommonName: "*." + gridDomain,
			DNSNames: []string{
				gridDomain,
			},
			ACME: &certmanager.ACMECertificateConfig{
				Config: []certmanager.ACMECertificateDomainConfig{
					certmanager.ACMECertificateDomainConfig{
						ACMESolverConfig: certmanager.ACMESolverConfig{
							DNS01: &certmanager.ACMECertificateDNS01Config{
								Provider: cerManagerProvider,
							},
						},
						Domains: []string{
							"*." + gridDomain,
							gridDomain,
						},
					},
				},
			},
		},
	}
	return kubecli.Certmanager().Certificates(ns).Create(cert)
}
