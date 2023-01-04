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

var _ resource.Resource = &ResourceServiceEndpointBitbucket{}
var _ resource.ResourceWithImportState = &ResourceServiceEndpointBitbucket{}

func NewResourceServiceEndpointBitbucket() resource.Resource {
	return &ResourceServiceEndpointBitbucket{}
}

type ResourceServiceEndpointBitbucket struct {
	client *serviceendpoint.Client
}

type ResourceServiceEndpointBitbucketModel struct {
	Description string       `tfsdk:"description"`
	Id          types.String `tfsdk:"id"`
	Name        string       `tfsdk:"name"`
	Password    string       `tfsdk:"password"`
	ProjectId   string       `tfsdk:"project_id"`
	UserName    string       `tfsdk:"username"`
}

func (r *ResourceServiceEndpointBitbucket) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_bitbucket"
}

func (r *ResourceServiceEndpointBitbucket) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Bitbucket service endpoint within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the service endpoint.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service endpoint.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Bitbucket account password.",
				Required:            true,
				Sensitive:           true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Bitbucket account username.",
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

	serviceEndpoint, err := CreateResourceServiceEndpoint(ctx, r.getCreateOrUpdateServiceEndpointArgs(model), model.ProjectId, r.client, resp)
	if err != nil {
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

	serviceEndpoint, err := ReadResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.client, resp)
	if err != nil {
		return
	}

	model.Description = *serviceEndpoint.Description
	model.Name = *serviceEndpoint.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceServiceEndpointBitbucket) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ResourceServiceEndpointBitbucketModel
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

func (r *ResourceServiceEndpointBitbucket) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ResourceServiceEndpointBitbucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.client, resp)
}

func (r *ResourceServiceEndpointBitbucket) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ResourceServiceEndpointBitbucket) getCreateOrUpdateServiceEndpointArgs(model *ResourceServiceEndpointBitbucketModel) *serviceendpoint.CreateOrUpdateServiceEndpointArgs {
	return &serviceendpoint.CreateOrUpdateServiceEndpointArgs{
		Description: model.Description,
		Name:        model.Name,
		Password:    model.Password,
		Type:        serviceendpoint.ServiceEndpointTypeBitbucket,
		UserName:    model.UserName,
	}
}
