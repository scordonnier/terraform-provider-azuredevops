package distributedtask

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/distributedtask"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

var _ resource.Resource = &ResourceEnvironment{}
var _ resource.ResourceWithImportState = &ResourceEnvironment{}

func NewResourceEnvironment() resource.Resource {
	return &ResourceEnvironment{}
}

type ResourceEnvironment struct {
	client *distributedtask.Client
}

type ResourceEnvironmentModel struct {
	Description string      `tfsdk:"description"`
	Id          types.Int64 `tfsdk:"id"`
	Name        string      `tfsdk:"name"`
	ProjectId   string      `tfsdk:"project_id"`
}

func (r *ResourceEnvironment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *ResourceEnvironment) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage environments in Azure Pipelines.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the environment.",
				Optional:            true,
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
					utils.UUIDStringValidator(),
				},
			},
		},
	}
}

func (r *ResourceEnvironment) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).DistributedTaskClient
}

func (r *ResourceEnvironment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ResourceEnvironmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	environment, err := r.client.CreateEnvironment(ctx, model.ProjectId, model.Name, model.Description)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create environment", err.Error())
		return
	}

	model.Id = types.Int64Value(int64(*environment.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceEnvironment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ResourceEnvironmentModel
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

	model.Description = *environment.Description
	model.Name = *environment.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceEnvironment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ResourceEnvironmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateEnvironment(ctx, model.ProjectId, int(model.Id.ValueInt64()), model.Name, model.Description)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Environment with Id '%d' failed to update", model.Id.ValueInt64()), err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceEnvironment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ResourceEnvironmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEnvironment(ctx, model.ProjectId, int(model.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Environment with Id '%d' failed to delete", model.Id.ValueInt64()), err.Error())
	}
}

func (r *ResourceEnvironment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
