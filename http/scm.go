package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/xumak-grid/bedrock"
	"github.com/xumak-grid/bedrock/commerce/ep"
	"github.com/xumak-grid/bedrock/k8s"
	"github.com/xumak-grid/bedrock/stack/gogs"
	"k8s.io/client-go/kubernetes"
)

// listSCM writes to w the list of the Source Control Managers available in the system
func listSCM(w http.ResponseWriter, r *http.Request) {
	list := scmVendors()
	encode(w, list)
}
func scmVendors() []bedrock.Vendor {
	return []bedrock.Vendor{
		gogs.Vendor(),
	}
}

func createSCMHandler(w http.ResponseWriter, r *http.Request) {
	scm := bedrock.SCM{}
	decode(r, &scm)
	if scm.Image == "" {
		jsonError(w, "image is required", http.StatusBadRequest)
		return
	}
	ns := chi.URLParam(r, "clientId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}
	if !validVendor(scm.SCMID, scmVendors()) {
		jsonError(w, "unknown scmId", http.StatusBadRequest)
		return
	}

	if scm.CustomConfig {
		if scm.Configuration == nil {
			jsonError(w, "configuration not provided when the request requires customConfig", http.StatusBadRequest)
			return
		}
		if scm.Configuration.InitData == nil {
			jsonError(w, "init_data not provided when the request requires customConfig", http.StatusBadRequest)
			return
		}
		d := scm.Configuration.InitData
		if d.AdminEmail == "" || d.AdminConfirmPass == "" || d.AdminName == "" || d.AdminPass == "" {
			jsonError(w, "admin account setting is invalid: one or more empty values", http.StatusBadRequest)
			return
		}
		if d.AdminName == "admin" {
			jsonError(w, "admin account setting is invalid: admin name is reserved", http.StatusBadRequest)
			return
		}
		if len(scm.Configuration.Repositories) > 0 {
			for _, repo := range scm.Configuration.Repositories {
				if repo.ContentSetupType == "ep-commerce" {
					if repo.EPObjectType == nil {
						jsonError(w, "ep_commerce not provided when the content_setup_type is set to ep-commerce", http.StatusBadRequest)
						return
					}
					initPack, err := ep.FindInitPackage(repo.EPObjectType.Version)
					if err != nil {
						jsonError(w, err.Error(), http.StatusBadRequest)
						return
					}
					if repo.EPObjectType.ExtensionVersion == "" {
						repo.EPObjectType.ExtensionVersion = ep.DefaultExtensionVersion
					}
					repo.EPObjectType.PlatformVersion = initPack.PlatformVersion

					// pre-signed url for the ep init package
					url, err := initPack.PreSignedURL()
					if err != nil {
						jsonError(w, err.Error(), http.StatusBadRequest)
						return
					}
					repo.EPObjectType.SourceCodeURL = url
				}
				if repo.ContentSetupType == "bloomreach-archetype" {
					if repo.BRObjectType == nil {
						jsonError(w, "bloomreach_archetype not provided when the content_setup_type is set to bloomreach-archetype", http.StatusBadRequest)
						return
					}
					b := repo.BRObjectType
					if b.ArchetypeVersion == "" || b.ArtifactID == "" || b.GroupID == "" || b.Package == "" || b.ProjectName == "" || b.Version == "" {
						jsonError(w, "required fields for bloomreach_archetype: archetype_version, group_id, artifact_id, version,package, project_name", http.StatusBadRequest)
						return
					}
				}
			}
		}
	}

	k8scli := getK8Client(r)
	err = createSCM(k8scli, ns, &scm)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	encode(w, scm)
}

// createSCM create a new SCM server and populates scm pointer with more data
// also creates k8s resources that are part of the SCM server
func createSCM(kubeCli kubernetes.Interface, ns string, scm *bedrock.SCM) error {
	k8Service, err := k8s.CreateService(kubeCli, ns, gogs.Service(ns))
	if err != nil {
		return err
	}
	k8Ingress, err := k8s.CreateIngress(kubeCli, ns, gogs.Ingress(ns))
	if err != nil {
		return err
	}
	k8Statefulset, err := k8s.CreateStatefulSet(kubeCli, ns, gogs.StatefulSet(scm.Image, ns))
	if err != nil {
		return err
	}

	// populating SCM
	scm.ServerName = k8Statefulset.Name
	scm.ServiceName = k8Service.Name
	scm.IngressName = k8Ingress.Name
	ingress := ""
	if len(k8Ingress.Spec.Rules) > 0 {
		ingress = k8Ingress.Spec.Rules[0].Host
	}
	scm.Host = "https://" + ingress

	// the request requires custom configuration
	if scm.CustomConfig {
		//default values for the init config
		scm.Configuration.InitData.Domain = ingress
		scm.Configuration.InitData.APPURL = scm.Host
		scm.Configuration.InitData.HTTPPort = "3000"
		scm.Configuration.InitData.RepoRootPath = "/data/git/gogs-repositories"
		scm.Configuration.InitData.LogRootPath = "/app/gogs/log"

		secretData, err := json.Marshal(scm.Configuration)
		if err != nil {
			return err
		}
		_, err = k8s.CreateSecret(kubeCli, ns, gogs.Secret(ns, secretData))
		if err != nil {
			return err
		}
		_, err = k8s.CreateJob(kubeCli, ns, gogs.InitJob(scm.Host, ns))
		if err != nil {
			return err
		}
	}
	return nil
}

func getSCM(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "clientId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}
	scm := bedrock.SCM{
		SCMID: chi.URLParam(r, "scmId"),
	}
	if !validVendor(scm.SCMID, scmVendors()) {
		jsonError(w, "unknown scmId", http.StatusBadRequest)
		return
	}
	k8sclient := getK8Client(r)
	k8StatfulSet, err := k8s.GetStatefulSet(k8sclient, ns, gogs.ServerName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	k8Service, err := k8s.GetService(k8sclient, ns, gogs.ServiceName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	k8Ingress, err := k8s.GetIngress(k8sclient, ns, gogs.IngressName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	scm.ServerName = k8StatfulSet.Name
	scm.ServiceName = k8Service.Name
	scm.IngressName = k8Ingress.Name

	if len(k8Ingress.Spec.Rules) > 0 {
		scm.Host = "https://" + k8Ingress.Spec.Rules[0].Host
	}

	encode(w, &scm)

}
func updateSCM(w http.ResponseWriter, r *http.Request) {
	jsonError(w, "operation not available", http.StatusInternalServerError)
}
func deleteSCM(w http.ResponseWriter, r *http.Request) {
	scm := bedrock.SCM{}
	ns := chi.URLParam(r, "clientId")
	scm.SCMID = chi.URLParam(r, "scmId")
	err := checkClient(r, ns)
	if err != nil {
		jsonError(w, "invalid clientId", http.StatusBadRequest)
		return
	}
	k8sclient := getK8Client(r)

	err = k8s.DeleteStatefulSet(k8sclient, ns, gogs.ServerName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = k8s.DeleteService(k8sclient, ns, gogs.ServiceName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = k8s.DeleteIngress(k8sclient, ns, gogs.IngressName)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	// delete if exist, ignoring errors
	err = k8s.DeleteJob(k8sclient, ns, gogs.InitJobName)
	if err != nil {
		log.Println("job not deleted:", err.Error())
	}
	err = k8s.DeleteSecret(k8sclient, ns, gogs.InitSecretName)
	if err != nil {
		log.Println("secret not deleted:", err.Error())
	}
	encode(w, scm)
}
