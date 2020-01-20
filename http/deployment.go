package http

import (
	"log"

	aemclientset "github.com/xumak-grid/aem-operator/pkg/generated/clientset/versioned"
	"github.com/xumak-grid/bedrock"
	"k8s.io/client-go/kubernetes"
)

// prepareFullDeploy returns a FullDeploy with the client configuration and some defaults
func prepareFullDeploy(c bedrock.Client) bedrock.FullDeploy {
	fullDeployment := bedrock.FullDeploy{}
	fullDeployment.Client = c

	// preparing the aemDeployments
	aemDeployments := []bedrock.AEMDeployment{}
	for _, i := range c.Configuration.Environments {
		newDeploy := bedrock.AEMDeployment{
			ClientID:      c.ClientID,
			EnvironmentID: i,
			Spec: bedrock.AEMDeploymentSpec{
				Authors: bedrock.Config{
					Type:     c.Configuration.AEMInstancesType,
					Replicas: 1,
				},
				Publishers: bedrock.Config{
					Type:     c.Configuration.AEMInstancesType,
					Replicas: 1,
				},
				Dispatchers: bedrock.Config{
					Type:     c.Configuration.DispatcherInstancesType,
					Replicas: 1,
				},
				Version:           c.Configuration.AEMInstancesVersion,
				DispatcherVersion: c.Configuration.DispatcherInstancesVersion,
			},
		}
		aemDeployments = append(aemDeployments, newDeploy)
	}
	fullDeployment.AEMDeployments = aemDeployments

	// preparing the artifactory
	hostedGroup := c.ClientID + "-group"
	hostedReleases := c.ClientID + "-releases"
	hostedSnapshots := c.ClientID + "-snapshots"
	proxyDanta := "xumak-danta"

	artifactory := bedrock.Artifactory{
		ArtifactoryID: "nexus",
		Image:         "grid/nexus:3.8.0",
		CustomConfig:  true,
		Configuration: &bedrock.ArtifactoryConfig{
			Hosteds: []bedrock.ArtifactoryHosted{
				bedrock.ArtifactoryHosted{
					Name:          hostedReleases,
					VersionPolicy: "RELEASE",
					LayoutPolicy:  "STRICT",
				},
				bedrock.ArtifactoryHosted{
					Name:          hostedSnapshots,
					VersionPolicy: "SNAPSHOT",
					LayoutPolicy:  "PERMISSIVE",
				},
			},
			Proxies: []bedrock.ArtifactoryProxy{
				bedrock.ArtifactoryProxy{
					Name:          proxyDanta,
					VersionPolicy: "RELEASE",
					LayoutPolicy:  "STRICT",
					RemoteURL:     "http://repo.tikaltechnologies.io/repository/danta-group",
					RequiredAuth:  false,
				},
			},
			Groups: []bedrock.ArtifactoryGroup{
				bedrock.ArtifactoryGroup{
					Name: hostedGroup,
					Members: []string{
						proxyDanta, hostedReleases, hostedSnapshots,
					},
				},
			},
		},
	}
	fullDeployment.Artifactory = artifactory

	// preparing the scm
	orgName := c.ClientID
	repoName := c.ClientID + "-app"
	scm := bedrock.SCM{
		SCMID:        "gogs",
		Image:        "grid/gogs:0.11.34",
		CustomConfig: true,
		Configuration: &bedrock.SCMConfig{
			InitData: &bedrock.SCMInitData{
				AdminName:        "xumak",
				AdminEmail:       c.Configuration.AdminEmail,
				AdminPass:        "xumakgt",
				AdminConfirmPass: "xumakgt",
			},
			Organizations: []bedrock.SCMOrganization{
				bedrock.SCMOrganization{
					UserName: orgName,
					FullName: c.Configuration.FullCompanyName,
				},
			},
			Repositories: []bedrock.SCMRepository{
				bedrock.SCMRepository{
					Name:             repoName,
					Owner:            orgName,
					ContentSetupType: c.Configuration.InitialRepositoryType,
				},
			},
		},
	}
	fullDeployment.SCM = scm

	// preparing the CI
	ci := bedrock.CI{
		CIID:        "drone",
		Image:       "grid/drone:0.8-alpine",
		SecondImage: "grid/drone-agent:0.8",
	}
	fullDeployment.CI = ci

	fullDeployment.Toolbelt.ClientID = c.ClientID

	return fullDeployment
}

// createFullDeploy creates a full deployment this action includes AEMDeployments, an Artifactory, a SCM and a CI resources
func createFullDeploy(fullDeploy *bedrock.FullDeploy, kubecli kubernetes.Interface, aemcli aemclientset.Interface) error {

	for _, i := range fullDeploy.AEMDeployments {
		err := createAEMDeployment(aemcli, i)
		if err != nil {
			return err
		}
	}

	err := createArtifactory(kubecli, fullDeploy.Client.ClientID, &fullDeploy.Artifactory)
	if err != nil {
		return err
	}

	err = createSCM(kubecli, fullDeploy.Client.ClientID, &fullDeploy.SCM)
	if err != nil {
		return err
	}

	fullDeploy.CI.ScmURL = fullDeploy.SCM.Host
	err = createCI(kubecli, fullDeploy.Client.ClientID, &fullDeploy.CI)
	if err != nil {
		return err
	}

	err = createToolbelt(kubecli, &fullDeploy.Toolbelt)
	if err != nil {
		log.Printf("error: toolbelt not created reason: %v", err.Error())
	}

	return nil
}
