package serviceendpoints

import (
	"github.com/google/uuid"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
)

const (
	ServiceEndpointDataAcceptUntrustedCertificates = "acceptUntrustedCerts"
	ServiceEndpointDataAuthorizationType           = "authorizationType"
	ServiceEndpointDataCreationMode                = "creationMode"
	ServiceEndpointDataEnvironment                 = "environment"
	ServiceEndpointDataRegistryType                = "registrytype"
	ServiceEndpointDataScopeLevel                  = "scopeLevel"
	ServiceEndpointDataSubscriptionId              = "subscriptionId"
	ServiceEndpointDataSubscriptionName            = "subscriptionName"

	ServiceEndpointAuthorizationParamsAccessToken         = "AccessToken"
	ServiceEndpointAuthorizationParamsApiToken            = "apitoken"
	ServiceEndpointAuthorizationParamsAuthenticationType  = "authenticationType"
	ServiceEndpointAuthorizationParamsClusterContext      = "clusterContext"
	ServiceEndpointAuthorizationParamsKubeconfig          = "kubeconfig"
	ServiceEndpointAuthorizationParamsPassword            = "password"
	ServiceEndpointAuthorizationParamsRegistry            = "registry"
	ServiceEndpointAuthorizationParamsServicePrincipalId  = "serviceprincipalid"
	ServiceEndpointAuthorizationParamsServicePrincipalKey = "serviceprincipalkey"
	ServiceEndpointAuthorizationParamsServiceTenantId     = "tenantid"
	ServiceEndpointAuthorizationParamsUserName            = "username"

	ServiceEndpointAuthorizationSchemeJwt                    = "JWT"
	ServiceEndpointAuthorizationSchemeKubernetes             = "Kubernetes"
	ServiceEndpointAuthorizationSchemeManagedServiceIdentity = "ManagedServiceIdentity"
	ServiceEndpointAuthorizationSchemeNone                   = "None"
	ServiceEndpointAuthorizationSchemeOAuth                  = "OAuth"
	ServiceEndpointAuthorizationSchemeOAuth2                 = "OAuth2"
	ServiceEndpointAuthorizationSchemePersonalAccessToken    = "PersonalAccessToken"
	ServiceEndpointAuthorizationSchemeServicePrincipal       = "ServicePrincipal"
	ServiceEndpointAuthorizationSchemeToken                  = "Token"
	ServiceEndpointAuthorizationSchemeUsernamePassword       = "UsernamePassword"

	ServiceEndpointOwnerAgentCloud  = "agentcloud"
	ServiceEndpointOwnerBoards      = "boards"
	ServiceEndpointOwnerEnvironment = "environment"
	ServiceEndpointOwnerLibrary     = "library"

	ServiceEndpointTypeAzure             = "Azure"
	ServiceEndpointTypeAzureRm           = "AzureRM"
	ServiceEndpointTypeBitbucket         = "Bitbucket"
	ServiceEndpointTypeDockerRegistry    = "dockerregistry"
	ServiceEndpointTypeGeneric           = "Generic"
	ServiceEndpointTypeGit               = "Git"
	ServiceEndpointTypeGitHub            = "GitHub"
	ServiceEndpointTypeGitHubEnterprise  = "GitHubEnterprise"
	ServiceEndpointTypeJFrogArtifactory  = "jfrogArtifactoryService"
	ServiceEndpointTypeJFrogDistribution = "jfrogDistributionService"
	ServiceEndpointTypeJFrogPlatform     = "jfrogPlatformService"
	ServiceEndpointTypeJFrogXray         = "jfrogXrayService"
	ServiceEndpointTypeKubernetes        = "kubernetes"
	ServiceEndpointTypeNpm               = "externalnpmregistry"
	ServiceEndpointTypeSSH               = "SSH"
	ServiceEndpointTypeVsAppCenter       = "vsmobilecenter"
)

type CreateOrUpdateServiceEndpointArgs struct {
	AcceptUntrustedCertificates bool
	ClusterContext              string
	Description                 string
	GrantAllPipelines           bool
	Kubeconfig                  string
	Name                        string
	Password                    string
	ServicePrincipalId          string
	ServicePrincipalKey         string
	SubscriptionId              string
	SubscriptionName            string
	TenantId                    string
	Token                       string
	Type                        string
	Url                         string
	Username                    string
}

type EndpointAuthorization struct {
	Parameters *map[string]string `json:"parameters,omitempty"`
	Scheme     *string            `json:"scheme,omitempty"`
}

type ServiceEndpoint struct {
	AdministratorsGroup              *core.IdentityRef                  `json:"administratorsGroup,omitempty"`
	Authorization                    *EndpointAuthorization             `json:"authorization,omitempty"`
	CreatedBy                        *core.IdentityRef                  `json:"createdBy,omitempty"`
	Data                             *map[string]string                 `json:"data,omitempty"`
	Description                      *string                            `json:"description"`
	GroupScopeId                     *uuid.UUID                         `json:"groupScopeId,omitempty"`
	Id                               *uuid.UUID                         `json:"id,omitempty"`
	IsReady                          *bool                              `json:"isReady,omitempty"`
	IsShared                         *bool                              `json:"isShared,omitempty"`
	Name                             *string                            `json:"name,omitempty"`
	OperationStatus                  interface{}                        `json:"operationStatus,omitempty"`
	Owner                            *string                            `json:"owner,omitempty"`
	ReadersGroup                     *core.IdentityRef                  `json:"readersGroup,omitempty"`
	ServiceEndpointProjectReferences *[]ServiceEndpointProjectReference `json:"serviceEndpointProjectReferences,omitempty"`
	Type                             *string                            `json:"type,omitempty"`
	Url                              *string                            `json:"url,omitempty"`
}

type ServiceEndpointProjectReference struct {
	Description      *string                `json:"description,omitempty"`
	Name             *string                `json:"name,omitempty"`
	ProjectReference *core.ProjectReference `json:"projectReference,omitempty"`
}
