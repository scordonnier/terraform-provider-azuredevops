package pipelines

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"strconv"
)

var _ resource.Resource = &AgentQueueResource{}

func NewAgentQueueResource() resource.Resource {
	return &AgentQueueResource{}
}

type AgentQueueResource struct {
	client *pipelines.Client
}

type AgentQueueResourceModel struct {
	AgentPoolId       int         `tfsdk:"agent_pool_id"`
	GrantAllPipelines bool        `tfsdk:"grant_all_pipelines"`
	Id                types.Int64 `tfsdk:"id"`
	ProjectId         string      `tfsdk:"project_id"`
}

func (r *AgentQueueResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_queue"
}

func (r *AgentQueueResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage agent queues within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"agent_pool_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The ID of the agent pool.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"grant_all_pipelines": schema.BoolAttribute{
				MarkdownDescription: "Set to true to grant access to all pipelines in the project.",
				Required:            true,
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the queue.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
		},
	}
}

func (r *AgentQueueResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).PipelinesClient
}

func (r *AgentQueueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *AgentQueueResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	queue, err := r.client.CreateAgentQueue(ctx, model.ProjectId, model.AgentPoolId, model.GrantAllPipelines)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create queue", err.Error())
		return
	}

	model.Id = types.Int64Value(int64(*queue.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AgentQueueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *AgentQueueResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.GetAgentQueue(ctx, model.ProjectId, int(model.Id.ValueInt64()))
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Unable to retrieve queue", err.Error())
		return
	}

	granted, err := r.client.GetPipelinePermissions(ctx, model.ProjectId, pipelines.PipelinePermissionsResourceTypeQueue, strconv.Itoa(int(model.Id.ValueInt64())))
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve grant status", err.Error())
		return
	}

	if granted.AllPipelines != nil {
		model.GrantAllPipelines = *granted.AllPipelines.Authorized
	} else {
		model.GrantAllPipelines = false
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AgentQueueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *AgentQueueResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.GrantAllPipelines(ctx, model.ProjectId, pipelines.PipelinePermissionsResourceTypeQueue, strconv.Itoa(int(model.Id.ValueInt64())), model.GrantAllPipelines)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update queue", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AgentQueueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *AgentQueueResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAgentQueue(ctx, model.ProjectId, int(model.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete queue", err.Error())
	}
}
