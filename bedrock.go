package bedrock

import "os"

const (
	// DevDockerRepository default aws repository.
	DevDockerRepository = "your.registry"

	// ProdDockerRepository default aws repository.
	ProdDockerRepository = "your.image"

	gridExternalDomainEnvVar = "GRID_EXTERNAL_DOMAIN"
)

// Client represents an abstraction of a client
type Client struct {
	// ClientID represents a namespace where the client resources will live
	ClientID string `json:"clientId"`
	// MetaData represents additional information of the client, this will be stored in annotations
	MetaData     map[string]string `json:"meta,omitempty"`
	CustomConfig bool              `json:"customConfig"`
	// Configuration is required when CustomConfig is set to true
	Configuration *ClientCustomConfig `json:"configuration,omitempty"`
	// DryRun allows to return the FullDeploy without create any resource
	DryRun bool `json:"dryRun,omitempty"`
}

// ClientCustomConfig represents basic information to create the fullDeploy for the client
type ClientCustomConfig struct {
	FullCompanyName            string   `json:"fullCompanyName"`
	AdminEmail                 string   `json:"adminEmail"`
	Environments               []string `json:"environments"`
	AEMInstancesVersion        string   `json:"aemInstancesVersion"`
	AEMInstancesType           string   `json:"aemInstancesType"`
	DispatcherInstancesVersion string   `json:"dispatcherInstancesVersion"`
	DispatcherInstancesType    string   `json:"dispatcherInstancesType"`
	// InitialRepositoryType defines the initial project in the repository
	// see options available in SCMRepository.ContentSetupType field
	InitialRepositoryType string `json:"initialRepositoryType"`
}

// FullDeploy represent a full deployment of a client, describes the resources that will be created
type FullDeploy struct {
	Client         Client
	AEMDeployments []AEMDeployment
	Artifactory    Artifactory
	SCM            SCM
	CI             CI
	Toolbelt       Toolbelt
}

// Toolbelt represent a toolbelt box to donwload
type Toolbelt struct {
	ClientID string `json:"clientId"`
	URL      string `json:"url"`
	Message  string `json:"message"`
}

// Deployment represents a bedrock stack deploy.
type Deployment struct {
	Version string `json:"version,omitempty"`
}

// AEMDeployment represents an AEM deployment
type AEMDeployment struct {
	// ClientID represents the namespace where the deployment is hosted
	// this value is taken from the URL
	ClientID string `json:"clientId"`
	// EnvironmentID represents the environment where the deployment deployed
	// also is the name of the aem deployment and is taken from the URL
	EnvironmentID string `json:"environmentId"`
	// Spec This is the configuration for the deployment, these are the value from the
	// body request
	Spec AEMDeploymentSpec `json:"spec,omitempty"`
	// Status shows the status of the AEM deployment
	Status string `json:"status,omitempty"`
}

// AEMDeploymentSpec represents the spec key in the AEM deployment
type AEMDeploymentSpec struct {
	Authors     Config `json:"authors,omitempty"`
	Publishers  Config `json:"publishers,omitempty"`
	Dispatchers Config `json:"dispatchers,omitempty"`
	// Version represents the version of the deployment for example "6.3"
	Version string `json:"version"`
	// DispatcherVersion represents the version of the dispatcher
	DispatcherVersion string `json:"dispatcher_version"`
}

// Config the instance config for AEM deployment app
type Config struct {
	// Type represents the type of the instances deployed for example "small"
	Type string `json:"type"`
	// Replicas is the number of replicas in the deployment
	Replicas int `json:"replicas"`
}

type SCMDeployment struct {
	Type string `json:"type,omitempty"`
	Deployment
}

type Deploy struct {
	Namespace   string         `json:"namespace"`
	ProjectName string         `json:"projectName"`
	SCM         *SCMDeployment `json:"scm,omitempty"`
}

type Environment struct {
	EnvironmentID string `json:"environmentId"`
}

type Instance struct {
	Name        string `json:"name"`
	Account     string `json:"account"`
	Environment string `json:"environment"`
	Runmode     string `json:"runmode"`
	Running     bool   `json:"running"`
	Ready       bool   `json:"ready"`
	Password    string `json:"password"`
}

// Artifactory represents an Artifactory manager for example nexus
type Artifactory struct {
	ArtifactoryID string `json:"artifactoryId"`
	ServerName    string `json:"serverName"`
	IngressName   string `json:"ingressName"`
	Image         string `json:"image,omitempty"`
	ServiceName   string `json:"serviceName"`
	Host          string `json:"host"`
	CustomConfig  bool   `json:"customConfig"`
	// Configuration is required when CustomConfig is set to true
	// includes definition on how the artifactory will be configured by a k8s job
	Configuration *ArtifactoryConfig `json:"configuration,omitempty"`
}

// ArtifactoryConfig represents the global configuration to apply in nexus
type ArtifactoryConfig struct {
	Users   []ArtifactoryUser   `json:"users,omitempty"`
	Groups  []ArtifactoryGroup  `json:"groups,omitempty"`
	Hosteds []ArtifactoryHosted `json:"hosteds,omitempty"`
	Proxies []ArtifactoryProxy  `json:"proxies,omitempty"`
}

// ArtifactoryUser represents a user in the server
type ArtifactoryUser struct {
	// Action the action for this user: CHANGE, CREATE
	Action      string `json:"action"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	NewPassword string `json:"newpassword"`
}

// ArtifactoryGroup represents a group repository
type ArtifactoryGroup struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// ArtifactoryHosted represents a hosted repository
type ArtifactoryHosted struct {
	Name string `json:"name"`
	// VersionPolicy the options are: RELEASE SNAPSHOT MIXED
	VersionPolicy string `json:"versionPolicy"`
	// LayoutPolicy the options are: STRICT PERMISSIVE
	LayoutPolicy string `json:"layoutPolicy"`
}

// ArtifactoryProxy represents a proxy repository
type ArtifactoryProxy struct {
	Name string `json:"name"`
	// VersionPolicy the options are: RELEASE SNAPSHOT MIXED
	VersionPolicy string `json:"versionPolicy"`
	// LayoutPolicy the options are: STRICT PERMISSIVE
	LayoutPolicy string `json:"layoutPolicy"`
	// RemoteURL is remote url to be proxied
	RemoteURL string `json:"remoteUrl"`
	// RequiredAuth set to true if the proxy requires authentication
	RequiredAuth bool `json:"requiredAuth"`
	// Authentication is required if RequiredAuth is set to true
	Authentication *ArtifactoryAuth `json:"authentication"`
}

// ArtifactoryAuth is the auth for artifactory proxy repository
type ArtifactoryAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SCM represents Source Control Manager for example gogs
type SCM struct {
	SCMID        string `json:"scmId"`
	ServerName   string `json:"serverName"`
	IngressName  string `json:"ingressName"`
	Image        string `json:"image,omitempty"`
	ServiceName  string `json:"serviceName"`
	Host         string `json:"host"`
	CustomConfig bool   `json:"customConfig"`
	// Configuration is required when CustomConfig is set to true
	// includes definition on how the ci will be configured by a k8s job
	Configuration *SCMConfig `json:"configuration,omitempty"`
}

// SCMConfig custom configuration for a scm
type SCMConfig struct {
	InitData      *SCMInitData      `json:"init_data"`
	Organizations []SCMOrganization `json:"organizations"`
	Repositories  []SCMRepository   `json:"repositories"`
}

// SCMInitData is the initial configuration for gogs server
type SCMInitData struct {
	Domain           string `json:"domain"`
	HTTPPort         string `json:"http_port"`
	APPURL           string `json:"app_url"`
	AdminName        string `json:"admin_name"`
	AdminPass        string `json:"admin_passwd"`
	AdminConfirmPass string `json:"admin_confirm_passwd"`
	AdminEmail       string `json:"admin_email"`
	RepoRootPath     string `json:"repo_root_path"`
	LogRootPath      string `json:"log_root_path"`
}

// SCMOrganization represents an organization in a scm
type SCMOrganization struct {
	UserName    string `json:"username"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Website     string `json:"website"`
	Location    string `json:"location"`
}

// SCMRepository represents a repository in a scm
type SCMRepository struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	Owner       string `json:"owner"`
	// ContentSetupType is the initial files for this repository
	// the options available are: "danta-aem-demo", "ep-commerce", "bloomreach-archetype" and ""
	ContentSetupType string `json:"content_setup_type"`
	// EPObjectType is required when the ContentSetupType is set to "ep-commerce"
	EPObjectType *EPObjectType `json:"ep_commerce,omitempty"`
	// BRObjectType is required when the ContentSetupType is set to "bloomreach-archetype"
	BRObjectType *BRObjectType `json:"bloomreach_archetype,omitempty"`
}

// EPObjectType is required when the ContentSetupType is set to "ep-commerce"
type EPObjectType struct {
	// ep version. Required (in format 7.1)
	Version string `json:"version"`
	// zip package url from S3. This is generated with a pre-signed url
	SourceCodeURL string `json:"source_code_url"`
	// Nexus ep-repository-group URL. Default ""
	MavenRepoURL string `json:"maven_rep_url"`
	// Project EP version that comes in the init package. This is generated
	PlatformVersion string `json:"platform_version"`
	// New EP version for the proyect. Default 0.0.0-SNAPSHOT
	ExtensionVersion string `json:"extension_version"`
}

// BRObjectType represents a bloomreach configuration and is required when
// ContentSetupType is set to "bloomreach-archetype"
type BRObjectType struct {
	// Bloomreach archetype version. Required (e.g. 12.2.0)
	ArchetypeVersion string `json:"archetype_version"`
	// Required (e.g. org.example)
	GroupID string `json:"group_id"`
	// Required (e.g. myCompany)
	ArtifactID string `json:"artifact_id"`
	// Version of the project. Required (e.g. 0.1.0-SNAPSHOT)
	Version string `json:"version"`
	// Project package. Required (e.g. com.example)
	Package string `json:"package"`
	// Required. (e.g. myProject)
	ProjectName string `json:"project_name"`
}

// CI represents Continuous Integration Manager, for example drone
type CI struct {
	CIID        string `json:"ciId"`
	ServerName  string `json:"serverName"`
	IngressName string `json:"ingressName"`
	Image       string `json:"image,omitempty"`
	SecondImage string `json:"secondImage,omitempty"`
	ServiceName string `json:"serviceName"`
	Host        string `json:"host"`
	ScmURL      string `json:"scmURL"`
}

// Vendor represents a Vendor for the services
type Vendor struct {
	Name string `json:"name"`
	// Images the list of images available for this Vendor
	Images []Image `json:"images"`
}

// Image represents a simple image
type Image struct {
	// Name the name of the image to use in the deployment, this must include the image and tag,
	// in the format grid/drone:0.8-alpine if the tag is not present 'latest' will be used
	Name string `json:"name"`
	// Secondary is an aditional image for the deployment, for example drone requires 2
	// images to deploy the drone and agent images
	Secondary string `json:"secondary,omitempty"`
}

// InstanceType represents the instance type e.g. small, medium, large
type InstanceType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ConfigMap represent a k8s configMap
type ConfigMap struct {
	ClientID      string            `json:"clientId"`
	EnvironmentID string            `json:"environmentId"`
	Name          string            `json:"name"`
	Data          map[string]string `json:"data"`
}

func GridDockerRepository() string {
	if os.Getenv("DEVELOPMENT") == "true" {
		return DevDockerRepository
	}
	return ProdDockerRepository
}

// GridExternalDomain is the grid DNS to all the ingresses
func GridExternalDomain() string {
	return os.Getenv(gridExternalDomainEnvVar)
}
