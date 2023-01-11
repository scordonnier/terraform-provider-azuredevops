package serviceendpoint

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoint"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

var _ resource.Resource = &ServiceEndpointGitHubResource{}
var _ resource.ResourceWithImportState = &ServiceEndpointGitHubResource{}

func NewServiceEndpointGitHubResource() resource.Resource {
	return &ServiceEndpointGitHubResource{}
}

type ServiceEndpointGitHubResource struct {
	pipelineClient        *pipelines.Client
	serviceEndpointClient *serviceendpoint.Client
}

type ServiceEndpointGitHubResourceModel struct {
	AccessToken       string       `tfsdk:"access_token"`
	Description       string       `tfsdk:"description"`
	GrantAllPipelines bool         `tfsdk:"grant_all_pipelines"`
	Id                types.String `tfsdk:"id"`
	Name              string       `tfsdk:"name"`
	ProjectId         string       `tfsdk:"project_id"`
}

func (r *ServiceEndpointGitHubResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_github"
}

func (r *ServiceEndpointGitHubResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceShema := GetServiceEndpointResourceSchemaBase("Manages a GitHub service endpoint within an Azure DevOps project.")
	resourceShema.Attributes["access_token"] = schema.StringAttribute{
		MarkdownDescription: "GitHub personal access token.",
		Required:            true,
		Sensitive:           true,
		Validators: []validator.String{
			utils.StringNotEmptyValidator(),
		},
	}
	resp.Schema = resourceShema
}

func (r *ServiceEndpointGitHubResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.pipelineClient = req.ProviderData.(*clients.AzureDevOpsClient).PipelineClient
	r.serviceEndpointClient = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointClient
}

func (r *ServiceEndpointGitHubResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ServiceEndpointGitHubResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, err := CreateResourceServiceEndpoint(ctx, model.ProjectId, r.getCreateOrUpdateServiceEndpointArgs(model), r.serviceEndpointClient, r.pipelineClient, resp)
	if err != nil {
		return
	}

	model.Id = types.StringValue(serviceEndpoint.Id.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointGitHubResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ServiceEndpointGitHubResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, granted, err := ReadResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointClient, r.pipelineClient, resp)
	if err != nil {
		return
	}

	model.Description = *serviceEndpoint.Description
	model.GrantAllPipelines = granted
	model.Name = *serviceEndpoint.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointGitHubResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ServiceEndpointGitHubResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := UpdateResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.getCreateOrUpdateServiceEndpointArgs(model), r.serviceEndpointClient, r.pipelineClient, resp)
	if err != nil {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointGitHubResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ServiceEndpointGitHubResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointClient, resp)
}

func (r *ServiceEndpointGitHubResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Private Methods

func (r *ServiceEndpointGitHubResource) getCreateOrUpdateServiceEndpointArgs(model *ServiceEndpointGitHubResourceModel) *serviceendpoint.CreateOrUpdateServiceEndpointArgs {
	return &serviceendpoint.CreateOrUpdateServiceEndpointArgs{
		Description:       model.Description,
		GrantAllPipelines: model.GrantAllPipelines,
		Name:              model.Name,
		Type:              serviceendpoint.ServiceEndpointTypeGitHub,
		Token:             model.AccessToken,
	}
}
