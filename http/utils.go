package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	certclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	aemclientset "github.com/xumak-grid/aem-operator/pkg/generated/clientset/versioned"
	"github.com/xumak-grid/bedrock"
	"k8s.io/client-go/kubernetes"
)

func getK8Client(r *http.Request) kubernetes.Interface {
	kubecli, ok := r.Context().Value(K8sClient).(kubernetes.Interface)
	if ok {
		return kubecli
	}
	return nil
}
func getAEMClient(r *http.Request) aemclientset.Interface {
	kubecli, ok := r.Context().Value(K8sAEMClient).(aemclientset.Interface)
	if ok {
		return kubecli
	}
	return nil
}

// getCertManagerClient returns the cerManagerClient from the context in the request
func getCertManagerClient(r *http.Request) certclient.Interface {
	kubecli, ok := r.Context().Value(CertManagerClientKey).(certclient.Interface)
	if ok {
		return kubecli
	}
	return nil
}

func decode(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func encode(w io.Writer, v interface{}) error {
	rw, ok := w.(http.ResponseWriter)
	if ok {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.Header().Set("X-Content-Type-Options", "nosniff")
	}
	return json.NewEncoder(w).Encode(v)
}

// JSONError represents an error in JSON format.
type JSONError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	err := &JSONError{Code: code}
	err.Msg = message
	_ = encode(w, err)
	return
}
func validVendor(name string, vendors []bedrock.Vendor) bool {
	for _, vendor := range vendors {
		if name == vendor.Name {
			return true
		}
	}
	return false
}

// getPodSecretKey returns the string key to be used when save secret for the given pod
func getPodSecretKey(nsName, deploymentName, podName string) string {
	return fmt.Sprintf("%s/%s", getSecretBasePath(nsName, deploymentName), podName)
}

// getSecretBasePath returns the base path to be used when save secrets for the given namespace and deployment
func getSecretBasePath(ns, deployment string) string {
	return fmt.Sprintf("secret/%v/%v", ns, deployment)
}
