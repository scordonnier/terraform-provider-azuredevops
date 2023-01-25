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

var _ resource.Resource = &ServiceEndpointGenericResource{}

func NewServiceEndpointGenericResource() resource.Resource {
	return &ServiceEndpointGenericResource{}
}

type ServiceEndpointGenericResource struct {
	pipelinesClient        *pipelines.Client
	serviceEndpointsClient *serviceendpoints.Client
}

type ServiceEndpointGenericResourceModel struct {
	Description       *string      `tfsdk:"description"`
	GrantAllPipelines bool         `tfsdk:"grant_all_pipelines"`
	Id                types.String `tfsdk:"id"`
	Name              string       `tfsdk:"name"`
	Password          types.String `tfsdk:"password"`
	ProjectId         string       `tfsdk:"project_id"`
	URL               string       `tfsdk:"url"`
	Username          types.String `tfsdk:"username"`
}

func (r *ServiceEndpointGenericResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_generic"
}

func (r *ServiceEndpointGenericResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceShema := GetServiceEndpointResourceSchemaBase("Manages a generic service endpoint within an Azure DevOps project, which can be used to authenticate to any external server using basic authentication via a username and password.")
	resourceShema.Attributes["password"] = schema.StringAttribute{
		MarkdownDescription: "The password or token key used to authenticate to the server url using basic authentication.",
		Optional:            true,
		Sensitive:           true,
		Validators: []validator.String{
			validators.StringNotEmpty(),
		},
	}
	resourceShema.Attributes["url"] = schema.StringAttribute{
		MarkdownDescription: "The URL of the server associated with the service endpoint.",
		Required:            true,
		Validators: []validator.String{
			validators.StringNotEmpty(),
		},
	}
	resourceShema.Attributes["username"] = schema.StringAttribute{
		MarkdownDescription: "The username used to authenticate to the server url using basic authentication.",
		Optional:            true,
		Sensitive:           true,
		Validators: []validator.String{
			validators.StringNotEmpty(),
		},
	}
	resp.Schema = resourceShema
}

func (r *ServiceEndpointGenericResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.pipelinesClient = req.ProviderData.(*clients.AzureDevOpsClient).PipelinesClient
	r.serviceEndpointsClient = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointsClient
}

func (r *ServiceEndpointGenericResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ServiceEndpointGenericResourceModel
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

func (r *ServiceEndpointGenericResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ServiceEndpointGenericResourceModel
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

func (r *ServiceEndpointGenericResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ServiceEndpointGenericResourceModel
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

func (r *ServiceEndpointGenericResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ServiceEndpointGenericResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointsClient, resp)
}

// Private Methods

func (r *ServiceEndpointGenericResource) getCreateOrUpdateServiceEndpointArgs(model *ServiceEndpointGenericResourceModel) *serviceendpoints.CreateOrUpdateServiceEndpointArgs {
	description := utils.IfThenElse[*string](model.Description != nil, model.Description, utils.EmptyString)
	return &serviceendpoints.CreateOrUpdateServiceEndpointArgs{
		Description:       *description,
		GrantAllPipelines: model.GrantAllPipelines,
		Name:              model.Name,
		Password:          model.Password.ValueString(),
		Type:              serviceendpoints.ServiceEndpointTypeGeneric,
		Url:               model.URL,
		Username:          model.Username.ValueString(),
	}
}
