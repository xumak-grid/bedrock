package http

import (
	"net/http"

	"github.com/go-chi/chi"
)

func clientsRouter(r chi.Router) {
	r.Get("/", ListClients)
	r.Post("/", createClientHandler)
	r.Get("/{clientId}", GetClient)
	r.Delete("/{clientId}", DeleteClient)
	r.Route("/{clientId}/environments", environmentsRouter)
	r.Route("/{clientId}/tools", toolsRouter)
	r.Route("/{clientId}/artifactory", artifactoryRouter)
	r.Route("/{clientId}/scm", scmRouter)
	r.Route("/{clientId}/ci", ciRouter)
}

func artifactoryRouter(r chi.Router) {
	r.Post("/", createArtifactoryHandler)
	r.Get("/{artifactoryId}", getArtifactory)
	r.Patch("/{artifactoryId}", updateArtifactory)
	r.Delete("/{artifactoryId}", deleteArtifactory)
}

func scmRouter(r chi.Router) {
	r.Post("/", createSCMHandler)
	r.Get("/{scmId}", getSCM)
	r.Patch("/{scmId}", updateSCM)
	r.Delete("/{scmId}", deleteSCM)
}

func ciRouter(r chi.Router) {
	r.Post("/", createCIHandler)
	r.Get("/{ciId}", getCI)
	r.Patch("/{ciId}", updateCI)
	r.Delete("/{ciId}", deleteCI)
}

func toolsRouter(r chi.Router) {
	r.Get("/toolbelt", getToolbeltHandler)
	r.Post("/toolbelt", createToolbeltHandler)
	r.Delete("/toolbelt", deleteToolbeltHandler)
}

func dispatcherRouter(r chi.Router) {
	r.Get("/", getDispatcherConfigHandler)
	r.Patch("/", updateDispatcherConfigHandler)
}

func environmentsRouter(r chi.Router) {
	r.Get("/", listEnvironments)
	r.Get("/{environmentId}/aem", getAEMDeployment)
	r.Post("/{environmentId}/aem", createAEMDeploymentHandler)
	r.Patch("/{environmentId}/aem", updateAEMDeployment)
	r.Delete("/{environmentId}/aem", deleteAEMDeployment)
	r.Get("/{environmentId}/aem/instances", ListAEMPods)
	r.Route("/{environmentId}/aem/dispatcherconfig", dispatcherRouter)
}

func vendorsRouter(r chi.Router) {
	r.Get("/artifactory/list", listArtifactory)
	r.Get("/scm/list", listSCM)
	r.Get("/ci/list", listCI)
}

func imagesRouter(r chi.Router) {
	r.Get("/aem/list", aemImageList)
	r.Get("/dispatcher/list", dispatcherImageList)
}

func instancesRouter(r chi.Router) {
	r.Get("/type/list", instanceTypeList)
}

func globalToolsRouter(r chi.Router) {

}

func (s *Server) getAPIRouter() http.Handler {
	r := chi.NewRouter()
	r.Route("/clients", clientsRouter)
	r.Route("/tools", globalToolsRouter)
	r.Route("/vendors", vendorsRouter)
	r.Route("/images", imagesRouter)
	r.Route("/instances", instancesRouter)
	return r
}
