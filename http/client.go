package http

import (
	"net/http"

	"github.com/go-chi/chi"
	certclient "github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	"github.com/xumak-grid/bedrock"
	"github.com/xumak-grid/bedrock/k8s"
	"k8s.io/client-go/kubernetes"
)

// createClientHandler creates a client that is represented by a namespace
// the client can include a customConfig to create a full deployment
// dryRun option helps to return the fullDeploy object without creating any resources
func createClientHandler(w http.ResponseWriter, r *http.Request) {
	c := bedrock.Client{}
	err := decode(r, &c)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if c.ClientID == "" {
		jsonError(w, "clienId is required", http.StatusBadRequest)
		return
	}
	if c.DryRun && !c.CustomConfig {
		jsonError(w, "dryRun only available with customConfig", http.StatusBadRequest)
		return
	}
	if c.CustomConfig {
		if c.Configuration == nil {
			jsonError(w, "configuration not provided when the request requires cusotomConfig", http.StatusBadRequest)
			return
		}
		if c.Configuration.AdminEmail == "" {
			jsonError(w, "adminEmail is required", http.StatusBadRequest)
			return
		}
		if c.Configuration.FullCompanyName == "" {
			jsonError(w, "fullCompanyName is required", http.StatusBadRequest)
			return
		}
		if len(c.Configuration.Environments) < 1 {
			jsonError(w, "environments are required, minimun 1", http.StatusBadRequest)
			return
		}
		if c.Configuration.AEMInstancesType == "" || c.Configuration.AEMInstancesVersion == "" {
			jsonError(w, "aemInstancesVersion or aemInstancesType are empty", http.StatusBadRequest)
			return
		}
		if c.Configuration.DispatcherInstancesType == "" || c.Configuration.DispatcherInstancesVersion == "" {
			jsonError(w, "dispatcherInstancesVersion or dispatcherInstancesType are empty", http.StatusBadRequest)
			return
		}
	}

	kubecli := getK8Client(r)
	certMClient := getCertManagerClient(r)
	if !c.DryRun {
		err = createClient(kubecli, certMClient, c)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// creating fullDeploy with customConfig
	if c.CustomConfig {
		fullDeploy := prepareFullDeploy(c)
		if c.DryRun {
			encode(w, fullDeploy)
			return
		}
		aemcli := getAEMClient(r)
		err := createFullDeploy(&fullDeploy, kubecli, aemcli)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(201)
		encode(w, fullDeploy)
		return
	}

	w.WriteHeader(201)
	encode(w, c)
}

// createClient creates a new client represented by a namespace in k8s
// also a certManager Certificate is created to allow tls endpoints with the ingresses
func createClient(kubecli kubernetes.Interface, certMClient certclient.Interface, c bedrock.Client) error {
	ns, err := k8s.CreateNamespace(kubecli, c.ClientID, c.MetaData)
	if err != nil {
		return err
	}

	// create a new certManager certificate
	_, err = k8s.CreateCertficate(certMClient, ns.Name)
	if err != nil {
		return err
	}
	return nil
}

// ListClients list the clients that are represented by the namespaces
func ListClients(w http.ResponseWriter, r *http.Request) {
	kubecli := getK8Client(r)
	namespaces, err := k8s.GetNamespaces(kubecli)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	clients := []bedrock.Client{}
	for _, ns := range namespaces {
		clients = append(clients, bedrock.Client{ClientID: ns.Name, MetaData: ns.Annotations})
	}
	encode(w, clients)
}

// GetClient returns a client getting information from the namespace
func GetClient(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientId")
	kubecli := getK8Client(r)
	ns, err := k8s.GetNamespace(kubecli, clientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	client := &bedrock.Client{}
	client.ClientID = ns.Name
	client.MetaData = ns.Annotations
	encode(w, client)
}

// DeleteClient deletes a client that is represented by a namespace
func DeleteClient(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientId")
	kubecli := getK8Client(r)
	err := k8s.DeleteNamespace(kubecli, clientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	c := bedrock.Client{ClientID: clientID}
	encode(w, c)
}

func checkClient(r *http.Request, clientID string) error {
	kubecli := getK8Client(r)
	// check if namespace exists
	_, err := k8s.GetNamespace(kubecli, clientID)
	if err != nil {
		return err
	}
	return nil
}
