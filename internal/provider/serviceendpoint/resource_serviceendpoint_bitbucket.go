package serviceendpoint

import (
	"context"
	"fmt"
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
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"regexp"
)

var _ resource.Resource = &ResourceServiceEndpointBitbucket{}
var _ resource.ResourceWithImportState = &ResourceServiceEndpointBitbucket{}

func NewResourceServiceEndpointBitbucket() resource.Resource {
	return &ResourceServiceEndpointBitbucket{}
}

type ResourceServiceEndpointBitbucket struct {
	client *serviceendpoint.Client
}

type ResourceServiceEndpointBitbucketModel struct {
	Description types.String `tfsdk:"description"`
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Password    types.String `tfsdk:"password"`
	ProjectId   types.String `tfsdk:"project_id"`
	UserName    types.String `tfsdk:"username"`
}

func (r *ResourceServiceEndpointBitbucket) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_bitbucket"
}

func (r *ResourceServiceEndpointBitbucket) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"password": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Required:            true,
				Sensitive:           true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"), "must be a valid UUID"),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *ResourceServiceEndpointBitbucket) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointClient
}

func (r *ResourceServiceEndpointBitbucket) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ResourceServiceEndpointBitbucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	args := getBitbucketCreateOrUpdateServiceEndpointArgs(model)
	serviceEndpoint, err := r.client.CreateServiceEndpoint(ctx, args, model.ProjectId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create service endpoint", err.Error())
		return
	}

	model.Id = types.StringValue(serviceEndpoint.Id.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceServiceEndpointBitbucket) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ResourceServiceEndpointBitbucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := model.Id.ValueString()
	serviceEndpoint, err := r.client.GetServiceEndpoint(ctx, id, model.ProjectId.ValueString())
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Error looking up service endpoint with Id '%s', %+v", id, err), "")
		return
	}

	if serviceEndpoint == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	model.Description = types.StringValue(*serviceEndpoint.Description)
	model.Name = types.StringValue(*serviceEndpoint.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceServiceEndpointBitbucket) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ResourceServiceEndpointBitbucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	args := getBitbucketCreateOrUpdateServiceEndpointArgs(model)
	id := model.Id.ValueString()
	_, err := r.client.UpdateServiceEndpoint(ctx, id, args, model.ProjectId.ValueString())
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("Service connection with Id '%s' does not exist", id), "")
			return
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Error looking up service endpoint with Id '%s', %+v", id, err), "")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceServiceEndpointBitbucket) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ResourceServiceEndpointBitbucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := model.Id.ValueString()
	err := r.client.DeleteServiceEndpoint(ctx, id, model.ProjectId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Service connection with Id '%s' failed to delete", id), err.Error())
	}
}

func (r *ResourceServiceEndpointBitbucket) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func getBitbucketCreateOrUpdateServiceEndpointArgs(model *ResourceServiceEndpointBitbucketModel) *serviceendpoint.CreateOrUpdateServiceEndpointArgs {
	return &serviceendpoint.CreateOrUpdateServiceEndpointArgs{
		Description: model.Description.ValueString(),
		Name:        model.Name.ValueString(),
		Password:    model.Password.ValueString(),
		Type:        serviceendpoint.ServiceEndpointTypeBitbucket,
		UserName:    model.UserName.ValueString(),
	}
}
