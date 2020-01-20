package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/xumak-grid/bedrock"

	"github.com/go-chi/chi"
	"github.com/xumak-grid/bedrock/k8s"
	"github.com/xumak-grid/bedrock/stack/nexus"
	"k8s.io/client-go/kubernetes"
)

// listArtifactory writes to w the list of the artifactory managers available in the system
func listArtifactory(w http.ResponseWriter, r *http.Request) {
	list := artifactoryVendors()
	encode(w, list)
}

// artifactoryVendors is the slice of vendors available
func artifactoryVendors() []bedrock.Vendor {
	return []bedrock.Vendor{
		nexus.Vendor(),
	}
}

// createArtifactoryHandler create the artifactory requested by the client
// this artifactory manager is created based on the list of artifactories available
func createArtifactoryHandler(w http.ResponseWriter, r *http.Request) {
	artifactory := bedrock.Artifactory{}
	decode(r, &artifactory)
	if artifactory.Image == "" {
		jsonError(w, "image is required", http.StatusBadRequest)
		return
	}
	ns := chi.URLParam(r, "clientId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}
	if !validVendor(artifactory.ArtifactoryID, artifactoryVendors()) {
		jsonError(w, "unknown artifactoryId", http.StatusBadRequest)
		return
	}
	if artifactory.CustomConfig {
		if artifactory.Configuration == nil {
			jsonError(w, "configuration not provided when the request requires cusotomConfig", http.StatusBadRequest)
			return
		}
		ops := len(artifactory.Configuration.Users) + len(artifactory.Configuration.Groups) + len(artifactory.Configuration.Hosteds) + len(artifactory.Configuration.Proxies)
		if ops == 0 {
			jsonError(w, "requires at least 1 member of users, groups, hosteds or proxies", http.StatusBadRequest)
			return
		}
	}

	k8scli := getK8Client(r)
	err = createArtifactory(k8scli, ns, &artifactory)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	encode(w, artifactory)
}

// createArtifactory creates a new artifactory and populates the artifactory pointer with more data
// also creates k8s resources that are part of the artifactory
func createArtifactory(kubeCli kubernetes.Interface, ns string, artifactory *bedrock.Artifactory) error {
	k8Service, err := k8s.CreateService(kubeCli, ns, nexus.Service(ns))
	if err != nil {
		return err
	}
	k8Ingress, err := k8s.CreateIngress(kubeCli, ns, nexus.Ingress(ns))
	if err != nil {
		return err
	}
	k8Statefulset, err := k8s.CreateStatefulSet(kubeCli, ns, nexus.StatefulSet(artifactory.Image, ns))
	if err != nil {
		return err
	}

	// populating artifactory
	artifactory.ServerName = k8Statefulset.Name
	artifactory.ServiceName = k8Service.Name
	artifactory.IngressName = k8Ingress.Name
	if len(k8Ingress.Spec.Rules) > 0 {
		artifactory.Host = "https://" + k8Ingress.Spec.Rules[0].Host
	}

	// the request requires custom configuration
	if artifactory.CustomConfig {
		secretData, err := json.Marshal(artifactory.Configuration)
		if err != nil {
			return err
		}
		_, err = k8s.CreateSecret(kubeCli, ns, nexus.Secret(ns, secretData))
		if err != nil {
			return err
		}
		_, err = k8s.CreateJob(kubeCli, ns, nexus.InitJob(artifactory.Host, ns))
		if err != nil {
			return err
		}
	}
	return nil
}

func getArtifactory(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "clientId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}
	artifactory := bedrock.Artifactory{
		ArtifactoryID: chi.URLParam(r, "artifactoryId"),
	}
	if !validVendor(artifactory.ArtifactoryID, artifactoryVendors()) {
		jsonError(w, "unknown artifactoryId", http.StatusBadRequest)
		return
	}
	k8sclient := getK8Client(r)
	k8StatfulSet, err := k8s.GetStatefulSet(k8sclient, ns, nexus.ServerName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	k8Service, err := k8s.GetService(k8sclient, ns, nexus.ServiceName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	k8Ingress, err := k8s.GetIngress(k8sclient, ns, nexus.IngressName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	artifactory.ServerName = k8StatfulSet.Name
	artifactory.ServiceName = k8Service.Name
	artifactory.IngressName = k8Ingress.Name

	if len(k8Ingress.Spec.Rules) > 0 {
		artifactory.Host = "https://" + k8Ingress.Spec.Rules[0].Host
	}

	encode(w, &artifactory)
}

func updateArtifactory(w http.ResponseWriter, r *http.Request) {
	jsonError(w, "operation not available", http.StatusInternalServerError)
}

func deleteArtifactory(w http.ResponseWriter, r *http.Request) {
	artifactory := bedrock.Artifactory{}
	ns := chi.URLParam(r, "clientId")
	artifactory.ArtifactoryID = chi.URLParam(r, "artifactoryId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}
	k8sclient := getK8Client(r)

	err = k8s.DeleteStatefulSet(k8sclient, ns, nexus.ServerName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = k8s.DeleteService(k8sclient, ns, nexus.ServiceName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = k8s.DeleteIngress(k8sclient, ns, nexus.IngressName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// delete if exist, ignoring errors
	err = k8s.DeleteJob(k8sclient, ns, nexus.InitJobName)
	if err != nil {
		log.Println("job not deleted:", err.Error())
	}
	err = k8s.DeleteSecret(k8sclient, ns, nexus.InitSecretName)
	if err != nil {
		log.Println("secret not deleted:", err.Error())
	}

	encode(w, artifactory)
}
