package distributedtask

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

var _ resource.Resource = &PipelinesSettingsResource{}

func NewPipelinesSettingsResource() resource.Resource {
	return &PipelinesSettingsResource{}
}

type PipelinesSettingsResource struct {
	client *pipelines.Client
}

type PipelinesSettingsResourceModel struct {
	General   PipelineGeneralSettings   `tfsdk:"general"`
	ProjectId string                    `tfsdk:"project_id"`
	Retention PipelineRetentionSettings `tfsdk:"retention"`
}

func (r *PipelinesSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipelines_settings"
}

func (r *PipelinesSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage pipeline settings of an existing project within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"general": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"disable_classic_pipeline_creation": schema.BoolAttribute{
						MarkdownDescription: "When this is enabled, users will not be able to create / import classic pipelines, classic release pipelines, task groups, and deployment groups. Existing classic (release) pipelines, task groups, and deployment groups will continue to work.",
						Required:            true,
					},
					"enforce_job_auth_scope": schema.BoolAttribute{
						MarkdownDescription: "If enabled, scope of access for all non-release pipelines reduces to the current project.",
						Required:            true,
					},
					"enforce_job_auth_scope_for_releases": schema.BoolAttribute{
						MarkdownDescription: "If enabled, scope of access for all release pipelines reduces to the current project.",
						Required:            true,
					},
					"enforce_referenced_repo_scoped_token": schema.BoolAttribute{
						MarkdownDescription: "Restricts the scope of access for all pipelines to only repositories explicitly referenced by the pipeline.",
						Required:            true,
					},
					"enforce_settable_var": schema.BoolAttribute{
						MarkdownDescription: "If enabled, only those variables that are explicitly marked as \"Settable at queue time\" can be set at queue time.",
						Required:            true,
					},
					"publish_pipeline_metadata": schema.BoolAttribute{
						MarkdownDescription: "Allows pipelines to record metadata.",
						Required:            true,
					},
					"status_badges_are_private": schema.BoolAttribute{
						MarkdownDescription: "Anonymous users can access the status badge API for all pipelines unless this option is enabled.",
						Required:            true,
					},
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
			"retention": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"days_to_keep_artifacts": schema.Int64Attribute{
						MarkdownDescription: "Number of days to keep artifacts, symbols and attachments.",
						Required:            true,
						Validators: []validator.Int64{
							int64validator.Between(1, 60),
						},
					},
					"days_to_keep_pullrequest_runs": schema.Int64Attribute{
						MarkdownDescription: "Number of days to keep pull request runs.",
						Required:            true,
						Validators: []validator.Int64{
							int64validator.Between(1, 30),
						},
					},
					"days_to_keep_runs": schema.Int64Attribute{
						MarkdownDescription: "Number of days to keep runs.",
						Required:            true,
						Validators: []validator.Int64{
							int64validator.Between(30, 731),
						},
					},
				},
			},
		},
	}
}

func (r *PipelinesSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).PipelineClient
}

func (r *PipelinesSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *PipelinesSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updatePipelineSettings(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update pipeline settings", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PipelinesSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *PipelinesSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pipelineSettings, err := r.client.GetPipelineSettings(ctx, model.ProjectId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to retrieve pipelines settings", err.Error())
		return
	}

	retentionSettings, err := r.client.GetPipelineRetentionSettings(ctx, model.ProjectId)
	if err != nil {
		resp.Diagnostics.AddError("Failed to retrieve retention settings", err.Error())
		return
	}

	model.General.DisableClassicPipelineCreation = pipelineSettings.DisableClassicPipelineCreation
	model.General.EnforceJobAuthScope = pipelineSettings.EnforceJobAuthScope
	model.General.EnforceJobAuthScopeForReleases = pipelineSettings.EnforceJobAuthScopeForReleases
	model.General.EnforceReferencedRepoScopedToken = pipelineSettings.EnforceReferencedRepoScopedToken
	model.General.EnforceSettableVar = pipelineSettings.EnforceSettableVar
	model.General.PublishPipelineMetadata = pipelineSettings.PublishPipelineMetadata
	model.General.StatusBadgesArePrivate = pipelineSettings.StatusBadgesArePrivate

	model.Retention.DaysToKeepArtifacts = retentionSettings.PurgeArtifacts.Value
	model.Retention.DaysToKeepPullRequestRuns = retentionSettings.PurgePullRequestRuns.Value
	model.Retention.DaysToKeepRuns = retentionSettings.PurgeRuns.Value

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PipelinesSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *PipelinesSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updatePipelineSettings(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update pipeline settings", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PipelinesSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *PipelinesSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	model.General.DisableClassicPipelineCreation = utils.Bool(false)
	model.General.EnforceJobAuthScope = utils.Bool(true)
	model.General.EnforceJobAuthScopeForReleases = utils.Bool(true)
	model.General.EnforceReferencedRepoScopedToken = utils.Bool(true)
	model.General.EnforceSettableVar = utils.Bool(true)
	model.General.PublishPipelineMetadata = utils.Bool(false)
	model.General.StatusBadgesArePrivate = utils.Bool(true)

	model.Retention.DaysToKeepArtifacts = utils.Int(30)
	model.Retention.DaysToKeepPullRequestRuns = utils.Int(10)
	model.Retention.DaysToKeepRuns = utils.Int(30)

	err := r.updatePipelineSettings(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete pipeline settings", err.Error())
	}
}

// Private Methods

func (r *PipelinesSettingsResource) updatePipelineSettings(ctx context.Context, model *PipelinesSettingsResourceModel) error {
	pipelineSettings := &pipelines.PipelineGeneralSettings{
		DisableClassicPipelineCreation:   model.General.DisableClassicPipelineCreation,
		EnforceJobAuthScope:              model.General.EnforceJobAuthScope,
		EnforceJobAuthScopeForReleases:   model.General.EnforceJobAuthScopeForReleases,
		EnforceReferencedRepoScopedToken: model.General.EnforceReferencedRepoScopedToken,
		EnforceSettableVar:               model.General.EnforceSettableVar,
		PublishPipelineMetadata:          model.General.PublishPipelineMetadata,
		StatusBadgesArePrivate:           model.General.StatusBadgesArePrivate,
	}
	_, err := r.client.UpdatePipelineSettings(ctx, model.ProjectId, pipelineSettings)
	if err != nil {
		return err
	}

	retentionSettings := &pipelines.UpdatePipelineRetentionSettings{
		PurgeArtifacts:       &pipelines.RetentionSetting{Value: model.Retention.DaysToKeepArtifacts},
		PurgePullRequestRuns: &pipelines.RetentionSetting{Value: model.Retention.DaysToKeepPullRequestRuns},
		PurgeRuns:            &pipelines.RetentionSetting{Value: model.Retention.DaysToKeepRuns},
	}
	_, err = r.client.UpdatePipelineRetentionSettings(ctx, model.ProjectId, retentionSettings)
	if err != nil {
		return err
	}

	return nil
}
