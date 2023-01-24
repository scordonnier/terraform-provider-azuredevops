package core

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/validators"
)

const (
	stateDisabled = "disabled"
	stateEnabled  = "enabled"
)

var _ resource.Resource = &ProjectFeaturesResource{}

func NewProjectFeaturesResource() resource.Resource {
	return &ProjectFeaturesResource{}
}

type ProjectFeaturesResource struct {
	client *core.Client
}

type ProjectFeaturesResourceModel struct {
	Artifacts    string `tfsdk:"artifacts"`
	Boards       string `tfsdk:"boards"`
	Pipelines    string `tfsdk:"pipelines"`
	ProjectId    string `tfsdk:"project_id"`
	Repositories string `tfsdk:"repositories"`
	TestPlans    string `tfsdk:"testplans"`
}

func (r *ProjectFeaturesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_features"
}

func (r *ProjectFeaturesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage features of an existing project within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"artifacts": schema.StringAttribute{
				MarkdownDescription: "If enabled, gives access to Azure Artifacts.",
				Required:            true,
				Validators: []validator.String{
					validators.EnabledDisabled(),
				},
			},
			"boards": schema.StringAttribute{
				MarkdownDescription: "If enabled, gives access to Azure Boards.",
				Required:            true,
				Validators: []validator.String{
					featuresDependenciesValidator{},
					validators.EnabledDisabled(),
				},
			},
			"pipelines": schema.StringAttribute{
				MarkdownDescription: "If enabled, gives access to Azure Pipelines.",
				Required:            true,
				Validators: []validator.String{
					validators.EnabledDisabled(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					validators.UUID(),
				},
			},
			"repositories": schema.StringAttribute{
				MarkdownDescription: "If enabled, gives access to Azure Repos.",
				Required:            true,
				Validators: []validator.String{
					validators.EnabledDisabled(),
				},
			},
			"testplans": schema.StringAttribute{
				MarkdownDescription: "If enabled, gives access to Azure Test Plans.",
				Required:            true,
				Validators: []validator.String{
					validators.EnabledDisabled(),
				},
			},
		},
	}
}

func (r *ProjectFeaturesResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).CoreClient
}

func (r *ProjectFeaturesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ProjectFeaturesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateFeatures(ctx, nil, model)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update project features", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectFeaturesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ProjectFeaturesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	features, err := r.client.GetProjectFeatures(ctx, model.ProjectId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to retrieve project features", err.Error())
		return
	}

	featureStates := *features.FeatureStates
	model.Artifacts = *featureStates[core.ProjectFeatureArtifacts].State
	model.Boards = *featureStates[core.ProjectFeatureBoards].State
	model.Pipelines = *featureStates[core.ProjectFeaturePipelines].State
	model.Repositories = *featureStates[core.ProjectFeatureRepositories].State
	model.TestPlans = *featureStates[core.ProjectFeatureTestPlans].State

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectFeaturesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var currentModel *ProjectFeaturesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentModel)...)

	var newModel *ProjectFeaturesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &newModel)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateFeatures(ctx, currentModel, newModel)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update project features", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newModel)...)
}

func (r *ProjectFeaturesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ProjectFeaturesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	model.Artifacts = stateEnabled
	model.Boards = stateEnabled
	model.Pipelines = stateEnabled
	model.Repositories = stateEnabled
	model.TestPlans = stateEnabled

	err := r.updateFeatures(ctx, nil, model)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete project features", err.Error())
	}
}

// Private Methods

func (r *ProjectFeaturesResource) updateFeatures(ctx context.Context, currentModel *ProjectFeaturesResourceModel, newModel *ProjectFeaturesResourceModel) error {
	if currentModel == nil || currentModel.Artifacts != newModel.Artifacts {
		_, err := r.client.UpdateProjectFeature(ctx, newModel.ProjectId, core.ProjectFeatureArtifacts, newModel.Artifacts)
		if err != nil {
			return err
		}
	}

	if currentModel == nil || currentModel.Boards != newModel.Boards {
		_, err := r.client.UpdateProjectFeature(ctx, newModel.ProjectId, core.ProjectFeatureBoards, newModel.Boards)
		if err != nil {
			return err
		}
	}

	if currentModel == nil || currentModel.Pipelines != newModel.Pipelines {
		_, err := r.client.UpdateProjectFeature(ctx, newModel.ProjectId, core.ProjectFeaturePipelines, newModel.Pipelines)
		if err != nil {
			return err
		}
	}

	if currentModel == nil || currentModel.Repositories != newModel.Repositories {
		_, err := r.client.UpdateProjectFeature(ctx, newModel.ProjectId, core.ProjectFeatureRepositories, newModel.Repositories)
		if err != nil {
			return err
		}
	}

	if currentModel == nil || currentModel.TestPlans != newModel.TestPlans {
		_, err := r.client.UpdateProjectFeature(ctx, newModel.ProjectId, core.ProjectFeatureTestPlans, newModel.TestPlans)
		if err != nil {
			return err
		}
	}

	return nil
}

// Validators

type featuresDependenciesValidator struct{}

func (v featuresDependenciesValidator) Description(_ context.Context) string {
	return "Features are not configured correctly"
}

func (v featuresDependenciesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v featuresDependenciesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	switch req.Path.String() {
	case "boards":
		var testplans types.String
		diags := req.Config.GetAttribute(ctx, path.Root("testplans"), &testplans)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		if req.ConfigValue.ValueString() == stateDisabled && testplans.ValueString() == stateEnabled {
			resp.Diagnostics.AddError(v.Description(ctx), "`Azure Test Plans` can't be enabled when `Azure Boards` is disabled")
		}
	default:
		return
	}
}
