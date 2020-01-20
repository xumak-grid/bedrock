package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/xumak-grid/bedrock"
	"github.com/xumak-grid/bedrock/k8s"
)

// getDispatcherConfigHandler returns the dispatcher configuration from an aem deployment
// the configuration is stored in a k8s configMap
func getDispatcherConfigHandler(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "clientId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}
	env := chi.URLParam(r, "environmentId")

	kubecli := getK8Client(r)
	k8scm, err := k8s.GetConfigMap(kubecli, ns, env+"-dispatcher")
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cmap := bedrock.ConfigMap{
		ClientID:      ns,
		EnvironmentID: env,
		Name:          k8scm.Name,
		Data:          k8scm.Data,
	}
	encode(w, cmap)
}

// updateDispatcherConfigHandler updates dispatcher configuration
// the configuration is updated in a k8s configMap
func updateDispatcherConfigHandler(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "clientId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}
	cmap := bedrock.ConfigMap{}
	decode(r, &cmap)
	if cmap.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	kubecli := getK8Client(r)
	err = k8s.UpdateConfigMap(kubecli, ns, cmap.Name, cmap.Data)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	cmap.ClientID = ns
	cmap.EnvironmentID = chi.URLParam(r, "environmentId")
	encode(w, cmap)
}
