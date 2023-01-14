package serviceendpoint

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoint"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

var _ resource.Resource = &ServiceEndpointAzureRmResource{}

func NewServiceEndpointAzureRmResource() resource.Resource {
	return &ServiceEndpointAzureRmResource{}
}

type ServiceEndpointAzureRmResource struct {
	pipelineClient        *pipelines.Client
	serviceEndpointClient *serviceendpoint.Client
}

type ServiceEndpointAzureRmResourceModel struct {
	Description         *string      `tfsdk:"description"`
	Id                  types.String `tfsdk:"id"`
	GrantAllPipelines   bool         `tfsdk:"grant_all_pipelines"`
	Name                string       `tfsdk:"name"`
	ProjectId           string       `tfsdk:"project_id"`
	ServicePrincipalId  string       `tfsdk:"service_principal_id"`
	ServicePrincipalKey string       `tfsdk:"service_principal_key"`
	SubscriptionId      string       `tfsdk:"subscription_id"`
	SubscriptionName    string       `tfsdk:"subscription_name"`
	TenantId            string       `tfsdk:"tenant_id"`
}

func (r *ServiceEndpointAzureRmResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_azurerm"
}

func (r *ServiceEndpointAzureRmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceShema := GetServiceEndpointResourceSchemaBase("Manages an AzureRM service endpoint within an Azure DevOps project.")
	resourceShema.Attributes["service_principal_id"] = schema.StringAttribute{
		MarkdownDescription: "The ID of the service principal.",
		Required:            true,
		Sensitive:           true,
	}
	resourceShema.Attributes["service_principal_key"] = schema.StringAttribute{
		MarkdownDescription: "The secret key of the service principal.",
		Required:            true,
		Sensitive:           true,
	}
	resourceShema.Attributes["subscription_id"] = schema.StringAttribute{
		MarkdownDescription: "The ID of the Azure subscription.",
		Required:            true,
	}
	resourceShema.Attributes["subscription_name"] = schema.StringAttribute{
		MarkdownDescription: "The name of the Azure subscription.",
		Required:            true,
	}
	resourceShema.Attributes["tenant_id"] = schema.StringAttribute{
		MarkdownDescription: "The tenant ID of the service principal.",
		Required:            true,
	}
	resp.Schema = resourceShema
}

func (r *ServiceEndpointAzureRmResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.pipelineClient = req.ProviderData.(*clients.AzureDevOpsClient).PipelineClient
	r.serviceEndpointClient = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointClient
}

func (r *ServiceEndpointAzureRmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ServiceEndpointAzureRmResourceModel
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

func (r *ServiceEndpointAzureRmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ServiceEndpointAzureRmResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, granted, err := ReadResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointClient, r.pipelineClient, resp)
	if err != nil {
		return
	}

	model.Description = utils.IfThenElse[*string](serviceEndpoint.Description != nil, model.Description, utils.EmptyString)
	model.GrantAllPipelines = granted
	model.Name = *serviceEndpoint.Name
	model.ServicePrincipalId = (*serviceEndpoint.Authorization.Parameters)[serviceendpoint.ServiceEndpointAuthorizationParamsServicePrincipalId]
	model.SubscriptionId = (*serviceEndpoint.Data)[serviceendpoint.ServiceEndpointDataSubscriptionId]
	model.SubscriptionName = (*serviceEndpoint.Data)[serviceendpoint.ServiceEndpointDataSubscriptionName]
	model.TenantId = (*serviceEndpoint.Authorization.Parameters)[serviceendpoint.ServiceEndpointAuthorizationParamsServiceTenantId]

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointAzureRmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ServiceEndpointAzureRmResourceModel
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

func (r *ServiceEndpointAzureRmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ServiceEndpointAzureRmResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointClient, resp)
}

// Private Methods

func (r *ServiceEndpointAzureRmResource) getCreateOrUpdateServiceEndpointArgs(model *ServiceEndpointAzureRmResourceModel) *serviceendpoint.CreateOrUpdateServiceEndpointArgs {
	description := utils.IfThenElse[*string](model.Description != nil, model.Description, utils.EmptyString)
	return &serviceendpoint.CreateOrUpdateServiceEndpointArgs{
		Description:         *description,
		GrantAllPipelines:   model.GrantAllPipelines,
		Name:                model.Name,
		ServicePrincipalId:  model.ServicePrincipalId,
		ServicePrincipalKey: model.ServicePrincipalKey,
		SubscriptionId:      model.SubscriptionId,
		SubscriptionName:    model.SubscriptionName,
		TenantId:            model.TenantId,
		Type:                serviceendpoint.ServiceEndpointTypeAzureRm,
	}
}
