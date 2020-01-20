package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/xumak-grid/bedrock"
	"github.com/xumak-grid/bedrock/k8s"
	"github.com/xumak-grid/bedrock/stack/drone"
	"k8s.io/client-go/kubernetes"
)

// listCI writes to w the list of the Continuous Integration managers available in the system
func listCI(w http.ResponseWriter, r *http.Request) {
	list := ciVendors()
	encode(w, list)
}
func ciVendors() []bedrock.Vendor {
	return []bedrock.Vendor{
		drone.Vendor(),
	}
}

// createCIHandler the handler to create CI server
func createCIHandler(w http.ResponseWriter, r *http.Request) {
	ci := bedrock.CI{}
	decode(r, &ci)
	if ci.Image == "" || ci.SecondImage == "" {
		jsonError(w, "image and secondImage are required", http.StatusBadRequest)
		return
	}
	if ci.ScmURL == "" {
		jsonError(w, "scmURL (Source Control manager URL) is required", http.StatusBadRequest)
		return
	}
	ns := chi.URLParam(r, "clientId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}
	if !validVendor(ci.CIID, ciVendors()) {
		jsonError(w, "unknown ciId", http.StatusBadRequest)
		return
	}

	k8sclient := getK8Client(r)
	err = createCI(k8sclient, ns, &ci)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	encode(w, ci)
}

// createCI create a new CI server and populates ci pointer with more data
// also creates k8s resources that are part of the CI server
func createCI(kubeCli kubernetes.Interface, ns string, ci *bedrock.CI) error {
	k8Service, err := k8s.CreateService(kubeCli, ns, drone.Service(ns))
	if err != nil {
		return err
	}
	k8Ingress, err := k8s.CreateIngress(kubeCli, ns, drone.Ingress(ns))
	if err != nil {
		return err
	}
	if len(k8Ingress.Spec.Rules) > 0 {
		ci.Host = "https://" + k8Ingress.Spec.Rules[0].Host
	}
	k8Statefulset, err := k8s.CreateStatefulSet(kubeCli, ns, drone.StatefulSet(ci.ScmURL, ci.Host, ns, ci.Image, ci.SecondImage))
	if err != nil {
		return err
	}

	// populating CI
	ci.ServerName = k8Statefulset.Name
	ci.ServiceName = k8Service.Name
	ci.IngressName = k8Ingress.Name
	ci.Image = drone.ServerName
	return nil
}

func getCI(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "clientId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}
	ci := bedrock.CI{
		CIID: chi.URLParam(r, "ciId"),
	}
	if !validVendor(ci.CIID, ciVendors()) {
		jsonError(w, "unknown ciId", http.StatusBadRequest)
		return
	}
	k8sclient := getK8Client(r)
	k8StatfulSet, err := k8s.GetStatefulSet(k8sclient, ns, drone.ServerName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	k8Service, err := k8s.GetService(k8sclient, ns, drone.ServiceName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	k8Ingress, err := k8s.GetIngress(k8sclient, ns, drone.IngressName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	ci.ServerName = k8StatfulSet.Name
	ci.ServiceName = k8Service.Name
	ci.IngressName = k8Ingress.Name
	ci.Image = drone.ServerName

	if len(k8Ingress.Spec.Rules) > 0 {
		ci.Host = "https://" + k8Ingress.Spec.Rules[0].Host
	}
	encode(w, &ci)
}
func updateCI(w http.ResponseWriter, r *http.Request) {
	jsonError(w, "operation not available", http.StatusInternalServerError)
}
func deleteCI(w http.ResponseWriter, r *http.Request) {
	ci := bedrock.CI{}
	ns := chi.URLParam(r, "clientId")
	ci.CIID = chi.URLParam(r, "ciId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}
	k8sclient := getK8Client(r)

	err = k8s.DeleteStatefulSet(k8sclient, ns, drone.ServerName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = k8s.DeleteService(k8sclient, ns, drone.ServiceName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = k8s.DeleteIngress(k8sclient, ns, drone.IngressName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	encode(w, ci)
}
