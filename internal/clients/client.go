package clients

import (
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
)

type AzureDevOpsClient struct {
	CoreClient *core.Client
}

func NewAzureDevOpsClient(organizationUrl string, authorization string, providerVersion string) *AzureDevOpsClient {
	restClient := networking.NewRestClient(organizationUrl, authorization, providerVersion)
	return &AzureDevOpsClient{
		CoreClient: core.NewClient(restClient),
	}
}
