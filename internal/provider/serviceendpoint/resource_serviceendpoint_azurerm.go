package serviceendpoint

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoint"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

var _ resource.Resource = &ResourceServiceEndpointAzureRm{}
var _ resource.ResourceWithImportState = &ResourceServiceEndpointAzureRm{}

func NewResourceServiceEndpointAzureRm() resource.Resource {
	return &ResourceServiceEndpointAzureRm{}
}

type ResourceServiceEndpointAzureRm struct {
	client *serviceendpoint.Client
}

type ResourceServiceEndpointAzureRmModel struct {
	Description         string       `tfsdk:"description"`
	Id                  types.String `tfsdk:"id"`
	Name                string       `tfsdk:"name"`
	ProjectId           string       `tfsdk:"project_id"`
	ServicePrincipalId  string       `tfsdk:"service_principal_id"`
	ServicePrincipalKey string       `tfsdk:"service_principal_key"`
	SubscriptionId      string       `tfsdk:"subscription_id"`
	SubscriptionName    string       `tfsdk:"subscription_name"`
	TenantId            string       `tfsdk:"tenant_id"`
}

func (r *ResourceServiceEndpointAzureRm) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_azurerm"
}

func (r *ResourceServiceEndpointAzureRm) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "", // TODO: Documentation
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
			"service_principal_id": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Required:            true,
				Sensitive:           true,
			},
			"service_principal_key": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Required:            true,
				Sensitive:           true,
			},
			"subscription_id": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Required:            true,
			},
			"subscription_name": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Required:            true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Required:            true,
			},
		},
	}
}

func (r *ResourceServiceEndpointAzureRm) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointClient
}

func (r *ResourceServiceEndpointAzureRm) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ResourceServiceEndpointAzureRmModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, err := CreateResourceServiceEndpoint(ctx, r.getCreateOrUpdateServiceEndpointArgs(model), model.ProjectId, r.client, resp)
	if err != nil {
		return
	}

	model.Id = types.StringValue(serviceEndpoint.Id.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceServiceEndpointAzureRm) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ResourceServiceEndpointAzureRmModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, err := ReadResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.client, resp)
	if err != nil {
		return
	}

	model.Description = *serviceEndpoint.Description
	model.Name = *serviceEndpoint.Name
	model.ServicePrincipalId = (*serviceEndpoint.Authorization.Parameters)[serviceendpoint.ServiceEndpointAuthorizationParamsServicePrincipalId]
	model.SubscriptionId = (*serviceEndpoint.Data)[serviceendpoint.ServiceEndpointDataSubscriptionId]
	model.SubscriptionName = (*serviceEndpoint.Data)[serviceendpoint.ServiceEndpointDataSubscriptionName]
	model.TenantId = (*serviceEndpoint.Authorization.Parameters)[serviceendpoint.ServiceEndpointAuthorizationParamsServiceTenantId]

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceServiceEndpointAzureRm) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ResourceServiceEndpointAzureRmModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := UpdateResourceServiceEndpoint(ctx, model.Id.ValueString(), r.getCreateOrUpdateServiceEndpointArgs(model), model.ProjectId, r.client, resp)
	if err != nil {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceServiceEndpointAzureRm) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ResourceServiceEndpointAzureRmModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.client, resp)
}

func (r *ResourceServiceEndpointAzureRm) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ResourceServiceEndpointAzureRm) getCreateOrUpdateServiceEndpointArgs(model *ResourceServiceEndpointAzureRmModel) *serviceendpoint.CreateOrUpdateServiceEndpointArgs {
	return &serviceendpoint.CreateOrUpdateServiceEndpointArgs{
		Description:         model.Description,
		Name:                model.Name,
		ServicePrincipalId:  model.ServicePrincipalId,
		ServicePrincipalKey: model.ServicePrincipalKey,
		SubscriptionId:      model.SubscriptionId,
		SubscriptionName:    model.SubscriptionName,
		TenantId:            model.TenantId,
		Type:                serviceendpoint.ServiceEndpointTypeAzureRm,
	}
}
