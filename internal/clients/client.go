package clients

import (
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/distributedtask"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/graph"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/security"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoint"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/workitems"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"path"
	"strings"
)

type AzureDevOpsClient struct {
	CoreClient            *core.Client
	DistributedTaskClient *distributedtask.Client
	GraphClient           *graph.Client
	PipelineClient        *pipelines.Client
	SecurityClient        *security.Client
	ServiceEndpointClient *serviceendpoint.Client
	WorkItemClient        *workitems.Client
}

func NewAzureDevOpsClient(organizationUrl string, authorization string, providerVersion string) *AzureDevOpsClient {
	azdoClient := networking.NewRestClient(organizationUrl, authorization, providerVersion)
	organizationName := path.Base(strings.TrimSuffix(organizationUrl, "/"))
	vsspsClient := networking.NewRestClient("https://vssps.dev.azure.com/"+organizationName, authorization, providerVersion)
	return &AzureDevOpsClient{
		CoreClient:            core.NewClient(azdoClient),
		DistributedTaskClient: distributedtask.NewClient(azdoClient),
		GraphClient:           graph.NewClient(vsspsClient),
		PipelineClient:        pipelines.NewClient(azdoClient),
		SecurityClient:        security.NewClient(azdoClient, vsspsClient),
		ServiceEndpointClient: serviceendpoint.NewClient(azdoClient),
		WorkItemClient:        workitems.NewClient(azdoClient),
	}
}
