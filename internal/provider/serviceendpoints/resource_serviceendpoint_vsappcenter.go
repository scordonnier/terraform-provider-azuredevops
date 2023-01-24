package serviceendpoints

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoints"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/validators"
)

var _ resource.Resource = &ServiceEndpointVsAppCenterResource{}

func NewServiceEndpointVsAppCenterResource() resource.Resource {
	return &ServiceEndpointVsAppCenterResource{}
}

type ServiceEndpointVsAppCenterResource struct {
	pipelinesClient        *pipelines.Client
	serviceEndpointsClient *serviceendpoints.Client
}

type ServiceEndpointVsAppCenterResourceModel struct {
	ApiToken          string       `tfsdk:"api_token"`
	Description       *string      `tfsdk:"description"`
	GrantAllPipelines bool         `tfsdk:"grant_all_pipelines"`
	Id                types.String `tfsdk:"id"`
	Name              string       `tfsdk:"name"`
	ProjectId         string       `tfsdk:"project_id"`
}

func (r *ServiceEndpointVsAppCenterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_vsappcenter"
}

func (r *ServiceEndpointVsAppCenterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceShema := GetServiceEndpointResourceSchemaBase("Manages a Visual Studio App Center service endpoint within an Azure DevOps project.")
	resourceShema.Attributes["api_token"] = schema.StringAttribute{
		MarkdownDescription: "Visual Studio App Center API token.",
		Required:            true,
		Sensitive:           true,
		Validators: []validator.String{
			validators.StringNotEmpty(),
		},
	}
	resp.Schema = resourceShema
}

func (r *ServiceEndpointVsAppCenterResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.pipelinesClient = req.ProviderData.(*clients.AzureDevOpsClient).PipelinesClient
	r.serviceEndpointsClient = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointsClient
}

func (r *ServiceEndpointVsAppCenterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ServiceEndpointVsAppCenterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, err := CreateResourceServiceEndpoint(ctx, model.ProjectId, r.getCreateOrUpdateServiceEndpointArgs(model), r.serviceEndpointsClient, r.pipelinesClient, resp)
	if err != nil {
		return
	}

	model.Id = types.StringValue(serviceEndpoint.Id.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointVsAppCenterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ServiceEndpointVsAppCenterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, granted, err := ReadResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointsClient, r.pipelinesClient, resp)
	if err != nil {
		return
	}

	model.Description = utils.IfThenElse[*string](serviceEndpoint.Description != nil, model.Description, utils.EmptyString)
	model.GrantAllPipelines = granted
	model.Name = *serviceEndpoint.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointVsAppCenterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ServiceEndpointVsAppCenterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := UpdateResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.getCreateOrUpdateServiceEndpointArgs(model), r.serviceEndpointsClient, r.pipelinesClient, resp)
	if err != nil {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointVsAppCenterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ServiceEndpointVsAppCenterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointsClient, resp)
}

// Private Methods

func (r *ServiceEndpointVsAppCenterResource) getCreateOrUpdateServiceEndpointArgs(model *ServiceEndpointVsAppCenterResourceModel) *serviceendpoints.CreateOrUpdateServiceEndpointArgs {
	description := utils.IfThenElse[*string](model.Description != nil, model.Description, utils.EmptyString)
	return &serviceendpoints.CreateOrUpdateServiceEndpointArgs{
		Description:       *description,
		GrantAllPipelines: model.GrantAllPipelines,
		Name:              model.Name,
		Type:              serviceendpoints.ServiceEndpointTypeVsAppCenter,
		Token:             model.ApiToken,
	}
}
