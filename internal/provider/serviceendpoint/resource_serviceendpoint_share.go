package serviceendpoint

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoint"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"golang.org/x/exp/slices"
)

var _ resource.Resource = &ResourceServiceEndpointShare{}
var _ resource.ResourceWithImportState = &ResourceServiceEndpointShare{}

func NewResourceServiceEndpointShare() resource.Resource {
	return &ResourceServiceEndpointShare{}
}

type ResourceServiceEndpointShare struct {
	client *serviceendpoint.Client
}

type ResourceServiceEndpointShareModel struct {
	Description string       `tfsdk:"description"`
	Id          types.String `tfsdk:"id"`
	Name        string       `tfsdk:"name"`
	ProjectId   string       `tfsdk:"project_id"`
	ProjectIds  []string     `tfsdk:"project_ids"`
}

func (r *ResourceServiceEndpointShare) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_share"
}

func (r *ResourceServiceEndpointShare) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Shares a service endpoint with multiple Azure DevOps projects.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the service endpoint.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the service endpoint.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service endpoint.",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project hosting the service endpoint.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
			"project_ids": schema.ListAttribute{
				MarkdownDescription: "The IDs of the projects to share the service endpoint.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(utils.UUIDStringValidator()),
				},
			},
		},
	}
}

func (r *ResourceServiceEndpointShare) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointClient
}

func (r *ResourceServiceEndpointShare) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ResourceServiceEndpointShareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.createOrUpdateServiceEndpointShare(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create service endpoint share", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceServiceEndpointShare) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ResourceServiceEndpointShareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, err := r.client.GetServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var projectIds []string
	for _, reference := range *serviceEndpoint.ServiceEndpointProjectReferences {
		projectIds = append(projectIds, reference.ProjectReference.Id.String())
	}

	model.ProjectIds = projectIds

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceServiceEndpointShare) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ResourceServiceEndpointShareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.createOrUpdateServiceEndpointShare(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update service endpoint share", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceServiceEndpointShare) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ResourceServiceEndpointShareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var projectIds []string
	for _, projectId := range model.ProjectIds {
		if projectId != model.ProjectId {
			projectIds = append(projectIds, projectId)
		}
	}

	err := r.client.DeleteServiceEndpoint(ctx, model.Id.ValueString(), projectIds)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete service endpoint share", err.Error())
	}
}

func (r *ResourceServiceEndpointShare) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ResourceServiceEndpointShare) createOrUpdateServiceEndpointShare(ctx context.Context, model *ResourceServiceEndpointShareModel) error {
	id := model.Id.ValueString()
	projectId := model.ProjectId

	serviceEndpoint, err := r.client.GetServiceEndpoint(ctx, id, projectId)
	if err != nil {
		return err
	}

	var currentProjectIds []string
	for _, projectReference := range *serviceEndpoint.ServiceEndpointProjectReferences {
		currentProjectIds = append(currentProjectIds, projectReference.ProjectReference.Id.String())
	}

	var deleteProjectIds []string
	for _, deleteProjectId := range currentProjectIds {
		if !slices.Contains(model.ProjectIds, deleteProjectId) {
			deleteProjectIds = append(deleteProjectIds, deleteProjectId)
		}
	}

	if len(deleteProjectIds) > 0 {
		deleteErr := r.client.DeleteServiceEndpoint(ctx, id, deleteProjectIds)
		if deleteErr != nil {
			return deleteErr
		}
	}

	var addProjectIds []string
	for _, addProjectId := range model.ProjectIds {
		if !slices.Contains(currentProjectIds, addProjectId) {
			addProjectIds = append(addProjectIds, addProjectId)
		}
	}

	if len(addProjectIds) > 0 {
		err := r.client.ShareServiceEndpoint(ctx, id, model.Name, model.Description, projectId, addProjectIds)
		if err != nil {
			return err
		}
	}

	return nil
}
