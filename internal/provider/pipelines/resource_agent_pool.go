package pipelines

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

var _ resource.Resource = &AgentPoolResource{}

func NewAgentPoolResource() resource.Resource {
	return &AgentPoolResource{}
}

type AgentPoolResource struct {
	client *pipelines.Client
}

type AgentPoolResourceModel struct {
	AutoProvision bool        `tfsdk:"auto_provision"`
	AutoUpdate    bool        `tfsdk:"auto_update"`
	Id            types.Int64 `tfsdk:"id"`
	Name          string      `tfsdk:"name"`
}

func (r *AgentPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_pool"
}

func (r *AgentPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage agent pools in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"auto_provision": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Auto-provision this agent pool in new projects.",
			},
			"auto_update": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Allow agents in this pool to automatically update.",
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the agent pool.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name which should be used for this agent pool.",
				Required:            true,
			},
		},
	}
}

func (r *AgentPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).PipelinesClient
}

func (r *AgentPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *AgentPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := r.client.CreateAgentPool(ctx, model.Name, model.AutoProvision, model.AutoUpdate)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create agent pool", err.Error())
		return
	}

	model.Id = types.Int64Value(int64(*pool.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AgentPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *AgentPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := r.client.GetAgentPool(ctx, int(model.Id.ValueInt64()))
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Unable to retrieve agent pool", err.Error())
		return
	}

	model.AutoProvision = *pool.AutoProvision
	model.AutoUpdate = *pool.AutoUpdate
	model.Name = *pool.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AgentPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *AgentPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateAgentPool(ctx, int(model.Id.ValueInt64()), model.Name, model.AutoProvision, model.AutoUpdate)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update agent pool", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AgentPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *AgentPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAgentPool(ctx, int(model.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete agent pool", err.Error())
	}
}
