package serviceendpoint

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoint"
	"regexp"
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
	Description         types.String `tfsdk:"description"`
	Id                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	ProjectId           types.String `tfsdk:"project_id"`
	ServicePrincipalId  types.String `tfsdk:"service_principal_id"`
	ServicePrincipalKey types.String `tfsdk:"service_principal_key"`
	SubscriptionId      types.String `tfsdk:"subscription_id"`
	SubscriptionName    types.String `tfsdk:"subscription_name"`
	TenantId            types.String `tfsdk:"tenant_id"`
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
					stringvalidator.RegexMatches(regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"), "must be a valid UUID"),
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

	serviceEndpoint, err := CreateResourceServiceEndpoint(ctx, r.getCreateOrUpdateServiceEndpointArgs(model), model.ProjectId.ValueString(), r.client, resp)
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

	serviceEndpoint, err := ReadResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId.ValueString(), r.client, resp)
	if err != nil {
		return
	}

	model.Description = types.StringValue(*serviceEndpoint.Description)
	model.Name = types.StringValue(*serviceEndpoint.Name)
	model.ServicePrincipalId = types.StringValue((*serviceEndpoint.Authorization.Parameters)[serviceendpoint.ServiceEndpointAuthorizationParamsServicePrincipalId])
	model.SubscriptionId = types.StringValue((*serviceEndpoint.Data)[serviceendpoint.ServiceEndpointDataSubscriptionId])
	model.SubscriptionName = types.StringValue((*serviceEndpoint.Data)[serviceendpoint.ServiceEndpointDataSubscriptionName])
	model.TenantId = types.StringValue((*serviceEndpoint.Authorization.Parameters)[serviceendpoint.ServiceEndpointAuthorizationParamsServiceTenantId])

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceServiceEndpointAzureRm) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ResourceServiceEndpointAzureRmModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := UpdateResourceServiceEndpoint(ctx, model.Id.ValueString(), r.getCreateOrUpdateServiceEndpointArgs(model), model.ProjectId.ValueString(), r.client, resp)
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

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId.ValueString(), r.client, resp)
}

func (r *ResourceServiceEndpointAzureRm) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ResourceServiceEndpointAzureRm) getCreateOrUpdateServiceEndpointArgs(model *ResourceServiceEndpointAzureRmModel) *serviceendpoint.CreateOrUpdateServiceEndpointArgs {
	return &serviceendpoint.CreateOrUpdateServiceEndpointArgs{
		Description:         model.Description.ValueString(),
		Name:                model.Name.ValueString(),
		ServicePrincipalId:  model.ServicePrincipalId.ValueString(),
		ServicePrincipalKey: model.ServicePrincipalKey.ValueString(),
		SubscriptionId:      model.SubscriptionId.ValueString(),
		SubscriptionName:    model.SubscriptionName.ValueString(),
		TenantId:            model.TenantId.ValueString(),
		Type:                serviceendpoint.ServiceEndpointTypeAzureRm,
	}
}
