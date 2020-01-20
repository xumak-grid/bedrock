package http

import (
	"log"
	"net/http"

	"github.com/xumak-grid/bedrock"

	"github.com/go-chi/chi"
	aemclientset "github.com/xumak-grid/aem-operator/pkg/generated/clientset/versioned"
	"github.com/xumak-grid/bedrock/k8s"
	"github.com/xumak-grid/bedrock/secrets/vault"
)

// createAEMDeploymentHandler to create AEMDeployments
func createAEMDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	aemDeploy := bedrock.AEMDeployment{}
	err := decode(r, &aemDeploy)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if aemDeploy.Spec.DispatcherVersion == "" || aemDeploy.Spec.Version == "" {
		jsonError(w, "dispatcher_version and version are required", http.StatusInternalServerError)
		return
	}
	aemDeploy.ClientID = chi.URLParam(r, "clientId")
	aemDeploy.EnvironmentID = chi.URLParam(r, "environmentId")
	err = checkClient(r, aemDeploy.ClientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	aemClient := getAEMClient(r)
	err = createAEMDeployment(aemClient, aemDeploy)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	encode(w, aemDeploy)
}

// createAEMDeployment creates an k8s AEM deployment
func createAEMDeployment(aemClient aemclientset.Interface, deploy bedrock.AEMDeployment) error {
	_, err := k8s.CreateAEMDeployment(aemClient, &deploy)
	if err != nil {
		return err
	}
	return nil
}

// getAEMDeployment router to get an AEM deployment
func getAEMDeployment(w http.ResponseWriter, r *http.Request) {
	aemDeploy := bedrock.AEMDeployment{}
	aemDeploy.ClientID = chi.URLParam(r, "clientId")
	aemDeploy.EnvironmentID = chi.URLParam(r, "environmentId")
	err := checkClient(r, aemDeploy.ClientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	aemClient := getAEMClient(r)
	k8sDep, err := k8s.GetAEMDeployment(aemClient, &aemDeploy)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	aemDeploy.Spec.Authors.Type = k8sDep.Spec.Authors.Type
	aemDeploy.Spec.Authors.Replicas = k8sDep.Spec.Authors.Replicas
	aemDeploy.Spec.Publishers.Type = k8sDep.Spec.Publishers.Type
	aemDeploy.Spec.Publishers.Replicas = k8sDep.Spec.Publishers.Replicas
	aemDeploy.Spec.Dispatchers.Type = k8sDep.Spec.Dispatchers.Type
	aemDeploy.Spec.Dispatchers.Replicas = k8sDep.Spec.Dispatchers.Replicas
	aemDeploy.Spec.Version = k8sDep.Spec.Version
	aemDeploy.Spec.DispatcherVersion = k8sDep.Spec.DispatcherVersion
	aemDeploy.Status = string(k8sDep.Status.Phase)
	encode(w, aemDeploy)
}

// deleteAEMDeployment router to delete an AEM deployment
func deleteAEMDeployment(w http.ResponseWriter, r *http.Request) {
	aemDeploy := bedrock.AEMDeployment{}
	aemDeploy.ClientID = chi.URLParam(r, "clientId")
	aemDeploy.EnvironmentID = chi.URLParam(r, "environmentId")
	err := checkClient(r, aemDeploy.ClientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	aemClient := getAEMClient(r)
	err = k8s.DeleteAEMDeployment(aemClient, &aemDeploy)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	encode(w, aemDeploy)

}

// updateAEMDeployment router to update an AEM deployment
func updateAEMDeployment(w http.ResponseWriter, r *http.Request) {
	aemDeploy := bedrock.AEMDeployment{}
	err := decode(r, &aemDeploy)
	aemDeploy.ClientID = chi.URLParam(r, "clientId")
	aemDeploy.EnvironmentID = chi.URLParam(r, "environmentId")
	err = checkClient(r, aemDeploy.ClientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	if aemDeploy.Spec.DispatcherVersion == "" || aemDeploy.Spec.Version == "" {
		jsonError(w, "dispatcher_version and version are required", http.StatusInternalServerError)
		return
	}

	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	aemClient := getAEMClient(r)
	err = k8s.UpdateAEMDeployment(aemClient, &aemDeploy)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	encode(w, aemDeploy)
}

// ListAEMPods list all pods for a given deployment in k8s.
func ListAEMPods(w http.ResponseWriter, r *http.Request) {
	aemDeploy := bedrock.AEMDeployment{}
	aemDeploy.ClientID = chi.URLParam(r, "clientId")
	aemDeploy.EnvironmentID = chi.URLParam(r, "environmentId")

	k8sclient := getK8Client(r)
	l, err := k8s.ListAEMDeploymentPods(k8sclient, &aemDeploy)
	if err != nil {
		log.Println(err)
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	instances := make([]bedrock.Instance, 0)
	for _, i := range l {

		password, err := getPodPassword(i.Namespace, aemDeploy.EnvironmentID, i.Name)
		if err != nil {
			log.Println(err)
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		instances = append(instances, bedrock.Instance{
			Name:        i.Name,
			Account:     i.Namespace,
			Environment: aemDeploy.EnvironmentID,
			Runmode:     i.Labels["runmode"],
			Running:     k8s.IsPodRunning(&i),
			Ready:       k8s.IsPodReady(&i),
			Password:    password,
		})
	}
	encode(w, instances)
}

// getPodPassword obtains the instance passwored stored in Vault
func getPodPassword(namespace string, deployment string, podName string) (string, error) {
	// Getting the instance passwords
	podSecretsKey := getPodSecretKey(namespace, deployment, podName)
	// podSecrets, err := secrets.SecretService.Get(podSecretsKey)
	secrets, err := vault.NewSecretService()
	if err != nil {
		return "", err
	}
	podSecrets, err := secrets.Get(podSecretsKey)
	if err != nil {
		return "", err
	}
	pwd, ok := podSecrets["password"]
	if !ok || pwd == nil {
		return "", nil
	}
	password, ok := pwd.(string)
	if !ok {
		return "", nil
	}
	return password, nil
}

// listEnvironments handler obtains the deployments in the namespace clientID
// the name of the deployment is the name of the environment
func listEnvironments(w http.ResponseWriter, r *http.Request) {
	aemDeploy := bedrock.AEMDeployment{}
	aemDeploy.ClientID = chi.URLParam(r, "clientId")
	aemDeploy.EnvironmentID = chi.URLParam(r, "environmentId")
	err := checkClient(r, aemDeploy.ClientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}

	aemClient := getAEMClient(r)
	k8sDeps, err := k8s.ListAEMDeployments(aemClient, aemDeploy.ClientID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	environments := []bedrock.Environment{}
	for _, k8sDep := range k8sDeps {
		environments = append(environments, bedrock.Environment{EnvironmentID: k8sDep.Name})
	}
	encode(w, environments)
}

func aemImageList(w http.ResponseWriter, r *http.Request) {
	list := []bedrock.Image{
		bedrock.Image{
			Name: "grid/aem-danta:6.3-1.0.5-jdk8",
		},
	}
	encode(w, list)
}
func dispatcherImageList(w http.ResponseWriter, r *http.Request) {
	list := []bedrock.Image{
		bedrock.Image{
			Name: "grid/dispatcher:4.2.2",
		},
	}
	encode(w, list)
}

func instanceTypeList(w http.ResponseWriter, r *http.Request) {
	list := []bedrock.InstanceType{
		bedrock.InstanceType{
			Name:        "small",
			Description: "instance type small",
		},
		bedrock.InstanceType{
			Name:        "medium",
			Description: "instance type medium",
		},
		bedrock.InstanceType{
			Name:        "large",
			Description: "instance type large",
		},
	}
	encode(w, list)
}
