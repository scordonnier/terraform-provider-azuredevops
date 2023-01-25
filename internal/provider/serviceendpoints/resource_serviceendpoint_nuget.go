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

var _ resource.Resource = &ServiceEndpointNuGetResource{}

func NewServiceEndpointNuGetResource() resource.Resource {
	return &ServiceEndpointNuGetResource{}
}

type ServiceEndpointNuGetResource struct {
	pipelinesClient        *pipelines.Client
	serviceEndpointsClient *serviceendpoints.Client
}

type ServiceEndpointNuGetResourceModel struct {
	ApiKey            types.String `tfsdk:"api_key"`
	Description       *string      `tfsdk:"description"`
	GrantAllPipelines bool         `tfsdk:"grant_all_pipelines"`
	Id                types.String `tfsdk:"id"`
	Name              string       `tfsdk:"name"`
	Password          types.String `tfsdk:"password"`
	ProjectId         string       `tfsdk:"project_id"`
	URL               string       `tfsdk:"url"`
	Username          types.String `tfsdk:"username"`
}

func (r *ServiceEndpointNuGetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_nuget"
}

func (r *ServiceEndpointNuGetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceShema := GetServiceEndpointResourceSchemaBase("Manages a NuGet service endpoint within an Azure DevOps project.")
	resourceShema.Attributes["api_key"] = schema.StringAttribute{
		MarkdownDescription: "API Key for NuGet feed.",
		Optional:            true,
		Sensitive:           true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRoot("password"), path.MatchRoot("username")),
			validators.StringNotEmpty(),
		},
	}
	resourceShema.Attributes["password"] = schema.StringAttribute{
		MarkdownDescription: "The password for NuGet feed.",
		Optional:            true,
		Sensitive:           true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRoot("api_key")),
			validators.StringNotEmpty(),
		},
	}
	resourceShema.Attributes["url"] = schema.StringAttribute{
		MarkdownDescription: "URL of the NuGet feed.",
		Required:            true,
		Validators: []validator.String{
			validators.StringNotEmpty(),
		},
	}
	resourceShema.Attributes["username"] = schema.StringAttribute{
		MarkdownDescription: "The username for NuGet feed.",
		Optional:            true,
		Sensitive:           true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRoot("api_key")),
			validators.StringNotEmpty(),
		},
	}
	resp.Schema = resourceShema
}

func (r *ServiceEndpointNuGetResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.pipelinesClient = req.ProviderData.(*clients.AzureDevOpsClient).PipelinesClient
	r.serviceEndpointsClient = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointsClient
}

func (r *ServiceEndpointNuGetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ServiceEndpointNuGetResourceModel
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

func (r *ServiceEndpointNuGetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ServiceEndpointNuGetResourceModel
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

func (r *ServiceEndpointNuGetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ServiceEndpointNuGetResourceModel
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

func (r *ServiceEndpointNuGetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ServiceEndpointNuGetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointsClient, resp)
}

// Private Methods

func (r *ServiceEndpointNuGetResource) getCreateOrUpdateServiceEndpointArgs(model *ServiceEndpointNuGetResourceModel) *serviceendpoints.CreateOrUpdateServiceEndpointArgs {
	description := utils.IfThenElse[*string](model.Description != nil, model.Description, utils.EmptyString)
	return &serviceendpoints.CreateOrUpdateServiceEndpointArgs{
		ApiKey:            model.ApiKey.ValueString(),
		Description:       *description,
		GrantAllPipelines: model.GrantAllPipelines,
		Name:              model.Name,
		Password:          model.Password.ValueString(),
		Type:              serviceendpoints.ServiceEndpointTypeNuGet,
		Url:               model.URL,
		Username:          model.Username.ValueString(),
	}
}
