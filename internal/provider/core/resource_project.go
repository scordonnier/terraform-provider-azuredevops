package core

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
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

type ProjectResource struct {
	client *core.Client
}

type ProjectResourceModel struct {
	Description       *string      `tfsdk:"description"`
	Id                types.String `tfsdk:"id"`
	Name              string       `tfsdk:"name"`
	ProcessTemplateId string       `tfsdk:"process_template_id"`
	VersionControl    string       `tfsdk:"version_control"`
	Visibility        string       `tfsdk:"visibility"`
}

func (r *ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage a project within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the project.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the project.",
				Required:            true,
			},
			"process_template_id": schema.StringAttribute{
				MarkdownDescription: "The process template ID of the project.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
			"version_control": schema.StringAttribute{
				MarkdownDescription: "Specifies the visibility of the project. Must be `Git` or `Tfvc`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("Git", "Tfvc"),
				},
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "Specifies the visibility of the project. Must be `private` or `public`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("private", "public"),
				},
			},
		},
	}
}

func (r *ProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).CoreClient
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	operation, err := r.client.CreateProject(ctx, model.Name, model.Description, model.Visibility, model.ProcessTemplateId, model.VersionControl)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create project", err.Error())
		return
	}

	stateConf := r.client.OperationStateChangeConf(ctx, r.client, operation)
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		resp.Diagnostics.AddError("Waiting for project ready", err.Error())
		return
	}

	project, err := r.client.GetProject(ctx, model.Name)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve created project", err.Error())
		return
	}

	model.Id = types.StringValue(project.Id.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.GetProject(ctx, model.Id.ValueString())
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Project with Id '%s' not found", model.Id.ValueString()), err.Error())
		return
	}

	model.Description = project.Description
	model.Name = *project.Name
	model.ProcessTemplateId = (*project.Capabilities)[core.CapabilitiesProcessTemplate][core.CapabilitiesProcessTemplateTypeId]
	model.VersionControl = (*project.Capabilities)[core.CapabilitiesVersionControl][core.CapabilitiesVersionControlType]
	model.Visibility = *project.Visibility

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	operation, err := r.client.UpdateProject(ctx, model.Id.ValueString(), model.Name, model.Description)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Project with Id '%s' failed to update", model.Id.ValueString()), err.Error())
		return
	}

	stateConf := r.client.OperationStateChangeConf(ctx, r.client, operation)
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		resp.Diagnostics.AddError("Waiting for project update", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	operation, err := r.client.DeleteProject(ctx, model.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Project with Id '%s' failed to delete", model.Id.ValueString()), err.Error())
		return
	}

	stateConf := r.client.OperationStateChangeConf(ctx, r.client, operation)
	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		resp.Diagnostics.AddError("Waiting for project delete", err.Error())
		return
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
