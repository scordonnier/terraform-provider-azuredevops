package graph

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/graph"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/validators"
)

var _ resource.Resource = &GroupResource{}

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

type GroupResource struct {
	client *graph.Client
}

type GroupResourceModel struct {
	Description types.String `tfsdk:"description"`
	Descriptor  types.String `tfsdk:"descriptor"`
	DisplayName string       `tfsdk:"display_name"`
	Name        types.String `tfsdk:"name"`
	Origin      types.String `tfsdk:"origin"`
	OriginId    types.String `tfsdk:"origin_id"`
	ProjectId   string       `tfsdk:"project_id"`
}

func (r *GroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage a group within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the group.",
				Optional:            true,
			},
			"descriptor": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The descriptor of the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name which should be used for this group.",
				Required:            true,
				Validators: []validator.String{
					validators.StringNotEmpty(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the group.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "The type of source provider for the group (eg. AD, AAD, MSA).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"origin_id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier from the system of origin.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project. Changing this forces a new group to be created.",
				Required:            true,
				Validators: []validator.String{
					validators.UUID(),
				},
			},
		},
	}
}

func (r *GroupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).GraphClient
}

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.CreateGroup(ctx, model.ProjectId, model.DisplayName, model.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create group", err.Error())
		return
	}

	model.Descriptor = types.StringValue(*group.Descriptor)
	model.Name = types.StringValue(*group.PrincipalName)
	model.Origin = types.StringValue(*group.Origin)
	model.OriginId = types.StringValue(*group.OriginId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetGroup(ctx, model.Descriptor.ValueString())
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Group with descriptor '%s' not found", model.Descriptor.ValueString()), err.Error())
		return
	}

	model.Description = types.StringValue(*group.Description)
	model.DisplayName = *group.DisplayName
	model.Name = types.StringValue(*group.PrincipalName)
	model.Origin = types.StringValue(*group.Origin)
	model.OriginId = types.StringValue(*group.OriginId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateGroup(ctx, model.Descriptor.ValueString(), model.DisplayName, model.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Group with descriptor '%s' failed to update", model.Descriptor.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGroup(ctx, model.Descriptor.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Group with descriptor '%s' failed to delete", model.Descriptor.ValueString()), err.Error())
	}
}
