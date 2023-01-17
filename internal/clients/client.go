package clients

import (
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/graph"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/security"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoints"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/workitems"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"path"
	"strings"
)

type AzureDevOpsClient struct {
	CoreClient             *core.Client
	GraphClient            *graph.Client
	PipelinesClient        *pipelines.Client
	SecurityClient         *security.Client
	ServiceEndpointsClient *serviceendpoints.Client
	WorkItemsClient        *workitems.Client
}

func NewAzureDevOpsClient(organizationUrl string, authorization string, providerVersion string) *AzureDevOpsClient {
	azdoClient := networking.NewRestClient(organizationUrl, authorization, providerVersion)
	organizationName := path.Base(strings.TrimSuffix(organizationUrl, "/"))
	vsspsClient := networking.NewRestClient("https://vssps.dev.azure.com/"+organizationName, authorization, providerVersion)
	return &AzureDevOpsClient{
		CoreClient:             core.NewClient(azdoClient),
		GraphClient:            graph.NewClient(vsspsClient),
		PipelinesClient:        pipelines.NewClient(azdoClient),
		SecurityClient:         security.NewClient(azdoClient, vsspsClient),
		ServiceEndpointsClient: serviceendpoints.NewClient(azdoClient),
		WorkItemsClient:        workitems.NewClient(azdoClient),
	}
}
