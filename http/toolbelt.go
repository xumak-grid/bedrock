package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/xumak-grid/bedrock"
	"github.com/xumak-grid/bedrock/awscli"
	"github.com/xumak-grid/bedrock/k8s"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	toolbletSecretName = "toolbelt"
)

// createToolbeltHandler manages the handler to create a new toolbelt
func createToolbeltHandler(w http.ResponseWriter, r *http.Request) {

	ns := chi.URLParam(r, "clientId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}

	tb := bedrock.Toolbelt{
		ClientID: ns,
	}
	kubecli := getK8Client(r)
	err = createToolbelt(kubecli, &tb)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	encode(w, tb)

}

// createToolbelt creates a presigned URL from the demo box and creates a k8s secret to persiste the data
func createToolbelt(kubeCli kubernetes.Interface, tb *bedrock.Toolbelt) error {

	bucket := "xumak-grid-boxes"
	client := "demo"
	box := "boot2docker_virtualbox2.box"
	hours := 24

	sess, err := awscli.Session()
	if err != nil {
		return fmt.Errorf("error session %s", err.Error())
	}

	key := client + "/" + box
	s3obj := awscli.NewS3Object(bucket, key)

	tb.URL, err = s3obj.PreSignedURL(sess, hours)
	tb.Message = fmt.Sprintf("url expires in %dhrs, time created: %v", hours, time.Now())

	_, err = k8s.CreateSecret(kubeCli, tb.ClientID, toolbeltSecret(*tb))
	if err != nil {
		return fmt.Errorf("error creating secret. %s", err.Error())
	}
	return nil
}

// getToolbeltHandler returns a toolbelt data from k8s secret
func getToolbeltHandler(w http.ResponseWriter, r *http.Request) {

	ns := chi.URLParam(r, "clientId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}

	kubecli := getK8Client(r)
	k8secret, err := k8s.GetSecret(kubecli, ns, toolbletSecretName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	tb := bedrock.Toolbelt{
		ClientID: ns,
		URL:      string(k8secret.Data["url"]),
		Message:  string(k8secret.Data["message"]),
	}
	encode(w, tb)
}

// deleteToolbeltHandler deletes k8s service with toolbelt data
func deleteToolbeltHandler(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "clientId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}

	kubecli := getK8Client(r)
	err = k8s.DeleteSecret(kubecli, ns, toolbletSecretName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	tb := bedrock.Toolbelt{
		ClientID: ns,
		Message:  "toolbelt deleted",
	}
	encode(w, tb)
}

// toolbeltSecret returns a k8s secret with toolbelt data (url and message)
func toolbeltSecret(tb bedrock.Toolbelt) *v1.Secret {
	var labels = map[string]string{
		"app":   toolbletSecretName,
		"stack": "bedrock",
	}
	burl := []byte(tb.URL)
	bmsg := []byte(tb.Message)
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      toolbletSecretName,
			Namespace: tb.ClientID,
			Labels:    labels,
		},
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			"url":     burl,
			"message": bmsg,
		},
	}
}
