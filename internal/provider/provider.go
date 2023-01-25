package provider

import (
	"context"
	"encoding/base64"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/provider/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/provider/git"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/provider/graph"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/provider/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/provider/serviceendpoints"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/provider/workitems"
)

var _ provider.Provider = &AzureDevOpsProvider{}

type AzureDevOpsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type AzureDevOpsProviderModel struct {
	OrganizationUrl     string `tfsdk:"organization_url"`
	PersonalAccessToken string `tfsdk:"personal_access_token"`
}

func (p *AzureDevOpsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "azuredevops"
	resp.Version = p.version
}

func (p *AzureDevOpsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization_url": schema.StringAttribute{
				MarkdownDescription: "The url of the Azure DevOps instance which should be used.",
				Required:            true,
			},
			"personal_access_token": schema.StringAttribute{
				MarkdownDescription: "The personal access token which should be used.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *AzureDevOpsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AzureDevOpsProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client := clients.NewAzureDevOpsClient(data.OrganizationUrl, "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+data.PersonalAccessToken)), p.version)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *AzureDevOpsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		core.NewProcessDataSource,
		core.NewProjectDataSource,
		core.NewProjectFeaturesDataSource,
		core.NewTeamDataSource,
		core.NewTeamsDataSource,
		graph.NewGroupDataSource,
		graph.NewGroupsDataSource,
		graph.NewUserDataSource,
		graph.NewUsersDataSource,
		pipelines.NewPipelineSettingsDataSource,
		workitems.NewAreaDataSource,
		workitems.NewIterationDataSource,
	}
}

func (p *AzureDevOpsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		core.NewProjectResource,
		core.NewProjectFeaturesResource,
		core.NewProjectPermissionsResource,
		core.NewTeamResource,
		git.NewGitPermissionsResource,
		graph.NewGroupResource,
		graph.NewGroupMembershipResource,
		pipelines.NewAgentPoolResource,
		pipelines.NewAgentQueueResource,
		pipelines.NewEnvironmentResource,
		pipelines.NewEnvironmentKubernetesResource,
		pipelines.NewEnvironmentPermissionsResource,
		pipelines.NewPipelinePermissionsResource,
		pipelines.NewPipelineSettingsResource,
		serviceendpoints.NewServiceEndpointAzureRmResource,
		serviceendpoints.NewServiceEndpointBitbucketResource,
		serviceendpoints.NewServiceEndpointDockerRegistryResource,
		serviceendpoints.NewServiceEndpointGenericResource,
		serviceendpoints.NewServiceEndpointGitHubResource,
		serviceendpoints.NewServiceEndpointJFrogResource,
		serviceendpoints.NewServiceEndpointKubernetesResource,
		serviceendpoints.NewServiceEndpointNuGetResource,
		serviceendpoints.NewServiceEndpointNpmResource,
		serviceendpoints.NewServiceEndpointPermissionsResource,
		serviceendpoints.NewServiceEndpointShareResource,
		serviceendpoints.NewServiceEndpointVsAppCenterResource,
		workitems.NewAreaPermissionsResource,
		workitems.NewAreaResource,
		workitems.NewIterationPermissionsResource,
		workitems.NewIterationResource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AzureDevOpsProvider{
			version: version,
		}
	}
}
