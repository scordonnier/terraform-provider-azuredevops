package workitems

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
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/workitems"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"strings"
)

var _ resource.Resource = &IterationResource{}
var _ resource.ResourceWithModifyPlan = &IterationResource{}

func NewIterationResource() resource.Resource {
	return &IterationResource{}
}

type IterationResource struct {
	client *workitems.Client
}

type IterationResourceModel struct {
	FinishDate *string      `tfsdk:"finish_date"`
	Id         types.Int64  `tfsdk:"id"`
	Name       string       `tfsdk:"name"`
	ParentPath string       `tfsdk:"parent_path"`
	Path       types.String `tfsdk:"path"`
	ProjectId  string       `tfsdk:"project_id"`
	StartDate  *string      `tfsdk:"start_date"`
}

func (r *IterationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iteration"
}

func (r *IterationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage an iteration within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"finish_date": schema.StringAttribute{
				MarkdownDescription: "The finish date of the iteration.",
				Optional:            true,
				Validators: []validator.String{
					utils.DateTimeValidator(),
				},
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the iteration.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the iteration.",
				Required:            true,
				Validators: []validator.String{
					utils.StringNotEmptyValidator(),
				},
			},
			"parent_path": schema.StringAttribute{
				MarkdownDescription: "The parent path of the iteration.",
				Required:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The path of the iteration.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
			"start_date": schema.StringAttribute{
				MarkdownDescription: "The start date of the iteration.",
				Optional:            true,
				Validators: []validator.String{
					utils.DateTimeValidator(),
				},
			},
		},
	}
}

func (r *IterationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).WorkItemsClient
}

func (r *IterationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *IterationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	area, err := r.client.CreateIteration(ctx, model.ProjectId, model.ParentPath, model.Name, model.StartDate, model.FinishDate)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create iteration", err.Error())
		return
	}

	model.Id = types.Int64Value(int64(*area.Id))
	model.Path = types.StringValue(GetAreaOrIterationPath(area))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *IterationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *IterationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	area, err := r.client.GetIteration(ctx, model.ProjectId, model.Path.ValueString())
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Iteration not found", err.Error())
		return
	}

	model.Name = *area.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *IterationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var currentModel *IterationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentModel)...)

	var newModel *IterationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &newModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var area *workitems.WorkItemClassificationNode
	var err error
	if strings.EqualFold(currentModel.ParentPath, newModel.ParentPath) {
		area, err = r.client.UpdateIteration(ctx, currentModel.ProjectId, currentModel.Path.ValueString(), newModel.Name, newModel.StartDate, newModel.FinishDate)
	} else {
		area, err = r.client.MoveIteration(ctx, currentModel.ProjectId, newModel.ParentPath, int(currentModel.Id.ValueInt64()), newModel.Name, newModel.StartDate, newModel.FinishDate)
	}

	if err != nil {
		resp.Diagnostics.AddError("Failed to update iteration", err.Error())
		return
	}

	newModel.Path = types.StringValue(GetAreaOrIterationPath(area))

	resp.Diagnostics.Append(resp.State.Set(ctx, &newModel)...)
}

func (r *IterationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *IterationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteIteration(ctx, model.ProjectId, model.Path.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete iteration", err.Error())
	}
}

func (r *IterationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var currentModel *IterationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentModel)...)

	var newModel *IterationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &newModel)...)

	// Do change the plan when creating or deleting a resource
	if currentModel == nil || newModel == nil {
		return
	}

	if strings.EqualFold(currentModel.ParentPath, newModel.ParentPath) {
		newModel.Path = types.StringValue(PlanAreaOrIterationPath(currentModel.Path.ValueString(), newModel.Name, false))
	} else {
		newModel.Path = types.StringValue(PlanAreaOrIterationPath(newModel.ParentPath, newModel.Name, true))
	}
	resp.Plan.Set(ctx, *newModel)
}
