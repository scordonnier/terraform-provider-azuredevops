package serviceendpoints

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

var _ resource.Resource = &ServiceEndpointNpmResource{}

func NewServiceEndpointNpmResource() resource.Resource {
	return &ServiceEndpointNpmResource{}
}

type ServiceEndpointNpmResource struct {
	pipelinesClient        *pipelines.Client
	serviceEndpointsClient *serviceendpoints.Client
}

type ServiceEndpointNpmResourceModel struct {
	AccessToken       types.String `tfsdk:"access_token"`
	Description       *string      `tfsdk:"description"`
	GrantAllPipelines bool         `tfsdk:"grant_all_pipelines"`
	Id                types.String `tfsdk:"id"`
	Name              string       `tfsdk:"name"`
	Password          types.String `tfsdk:"password"`
	ProjectId         string       `tfsdk:"project_id"`
	URL               string       `tfsdk:"url"`
	Username          types.String `tfsdk:"username"`
}

func (r *ServiceEndpointNpmResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_npm"
}

func (r *ServiceEndpointNpmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceShema := GetServiceEndpointResourceSchemaBase("Manages a npm registry service endpoint within an Azure DevOps project.")
	resourceShema.Attributes["access_token"] = schema.StringAttribute{
		MarkdownDescription: "Access Token for npm registry.",
		Optional:            true,
		Sensitive:           true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRoot("password"), path.MatchRoot("username")),
			validators.StringNotEmpty(),
		},
	}
	resourceShema.Attributes["password"] = schema.StringAttribute{
		MarkdownDescription: "The password for npm registry.",
		Optional:            true,
		Sensitive:           true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRoot("access_token")),
			validators.StringNotEmpty(),
		},
	}
	resourceShema.Attributes["url"] = schema.StringAttribute{
		MarkdownDescription: "URL of the npm registry.",
		Required:            true,
		Validators: []validator.String{
			validators.StringNotEmpty(),
		},
	}
	resourceShema.Attributes["username"] = schema.StringAttribute{
		MarkdownDescription: "The username for npm registry.",
		Optional:            true,
		Sensitive:           true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRoot("access_token")),
			validators.StringNotEmpty(),
		},
	}
	resp.Schema = resourceShema
}

func (r *ServiceEndpointNpmResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.pipelinesClient = req.ProviderData.(*clients.AzureDevOpsClient).PipelinesClient
	r.serviceEndpointsClient = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointsClient
}

func (r *ServiceEndpointNpmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ServiceEndpointNpmResourceModel
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

func (r *ServiceEndpointNpmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ServiceEndpointNpmResourceModel
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

func (r *ServiceEndpointNpmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ServiceEndpointNpmResourceModel
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

func (r *ServiceEndpointNpmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ServiceEndpointNpmResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointsClient, resp)
}

// Private Methods

func (r *ServiceEndpointNpmResource) getCreateOrUpdateServiceEndpointArgs(model *ServiceEndpointNpmResourceModel) *serviceendpoints.CreateOrUpdateServiceEndpointArgs {
	description := utils.IfThenElse[*string](model.Description != nil, model.Description, utils.EmptyString)
	return &serviceendpoints.CreateOrUpdateServiceEndpointArgs{
		Description:       *description,
		GrantAllPipelines: model.GrantAllPipelines,
		Name:              model.Name,
		Password:          model.Password.ValueString(),
		Token:             model.AccessToken.ValueString(),
		Type:              serviceendpoints.ServiceEndpointTypeNpm,
		Url:               model.URL,
		Username:          model.Username.ValueString(),
	}
}
