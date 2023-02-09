package pipelines

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/validators"
	"strconv"
)

var _ resource.Resource = &EnvironmentResource{}

func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

type EnvironmentResource struct {
	client *pipelines.Client
}

type EnvironmentResourceModel struct {
	Description       *string     `tfsdk:"description"`
	GrantAllPipelines types.Bool  `tfsdk:"grant_all_pipelines"`
	Id                types.Int64 `tfsdk:"id"`
	Name              string      `tfsdk:"name"`
	ProjectId         string      `tfsdk:"project_id"`
}

func (r *EnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *EnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage environments in Azure Pipelines.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the environment.",
				Optional:            true,
			},
			"grant_all_pipelines": schema.BoolAttribute{
				MarkdownDescription: "Set to true to grant access to all pipelines in the project.",
				Required:            true,
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the environment.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name which should be used for this environment.",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project. Changing this forces a new environment to be created.",
				Required:            true,
				Validators: []validator.String{
					validators.UUID(),
				},
			},
		},
	}
}

func (r *EnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).PipelinesClient
}

func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *EnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	description := utils.IfThenElse[*string](model.Description != nil, model.Description, utils.EmptyString)
	environment, err := r.client.CreateEnvironment(ctx, model.ProjectId, model.Name, *description)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create environment", err.Error())
		return
	}

	_, err = r.client.GrantAllPipelines(ctx, model.ProjectId, pipelines.PipelinePermissionsResourceTypeEnvironment, strconv.Itoa(*environment.Id), model.GrantAllPipelines.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Unable to grant environment access to all pipelines", err.Error())
		return
	}

	model.Id = types.Int64Value(int64(*environment.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *EnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	environment, err := r.client.GetEnvironment(ctx, model.ProjectId, int(model.Id.ValueInt64()))
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Error looking up environment with Id '%d'", model.Id.ValueInt64()), err.Error())
		return
	}

	permissions, err := r.client.GetPipelinePermissions(ctx, model.ProjectId, pipelines.PipelinePermissionsResourceTypeEnvironment, model.Id.String())
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve grant access", err.Error())
		return
	}

	model.Description = utils.IfThenElse[*string](environment.Description != nil, model.Description, utils.EmptyString)
	model.Name = *environment.Name
	if permissions.AllPipelines != nil {
		model.GrantAllPipelines = types.BoolValue(*permissions.AllPipelines.Authorized)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *EnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	description := utils.IfThenElse[*string](model.Description != nil, model.Description, utils.EmptyString)
	_, err := r.client.UpdateEnvironment(ctx, model.ProjectId, int(model.Id.ValueInt64()), model.Name, *description)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Environment with Id '%d' failed to update", model.Id.ValueInt64()), err.Error())
		return
	}

	_, err = r.client.GrantAllPipelines(ctx, model.ProjectId, pipelines.PipelinePermissionsResourceTypeEnvironment, model.Id.String(), model.GrantAllPipelines.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Unable to grant environment access to all pipelines", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *EnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEnvironment(ctx, model.ProjectId, int(model.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Environment with Id '%d' failed to delete", model.Id.ValueInt64()), err.Error())
	}
}
