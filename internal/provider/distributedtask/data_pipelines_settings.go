package distributedtask

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

var _ datasource.DataSource = &PipelineSettingsDataSource{}

func NewPipelineSettingsDataSource() datasource.DataSource {
	return &PipelineSettingsDataSource{}
}

type PipelineSettingsDataSource struct {
	client *pipelines.Client
}

type PipelineSettingsDataSourceModel struct {
	General   *PipelineGeneralSettings   `tfsdk:"general"`
	ProjectId string                     `tfsdk:"project_id"`
	Retention *PipelineRetentionSettings `tfsdk:"retention"`
}

func (d *PipelineSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipelines_settings"
}

func (d *PipelineSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about pipeline settings of an existing project within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"general": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"disable_classic_pipeline_creation": schema.BoolAttribute{
						MarkdownDescription: "When this is enabled, users will not be able to create / import classic pipelines, classic release pipelines, task groups, and deployment groups. Existing classic (release) pipelines, task groups, and deployment groups will continue to work.",
						Computed:            true,
					},
					"enforce_job_auth_scope": schema.BoolAttribute{
						MarkdownDescription: "If enabled, scope of access for all non-release pipelines reduces to the current project.",
						Computed:            true,
					},
					"enforce_job_auth_scope_for_releases": schema.BoolAttribute{
						MarkdownDescription: "If enabled, scope of access for all release pipelines reduces to the current project.",
						Computed:            true,
					},
					"enforce_referenced_repo_scoped_token": schema.BoolAttribute{
						MarkdownDescription: "Restricts the scope of access for all pipelines to only repositories explicitly referenced by the pipeline.",
						Computed:            true,
					},
					"enforce_settable_var": schema.BoolAttribute{
						MarkdownDescription: "If enabled, only those variables that are explicitly marked as \"Settable at queue time\" can be set at queue time.",
						Computed:            true,
					},
					"publish_pipeline_metadata": schema.BoolAttribute{
						MarkdownDescription: "Allows pipelines to record metadata.",
						Computed:            true,
					},
					"status_badges_are_private": schema.BoolAttribute{
						MarkdownDescription: "Anonymous users can access the status badge API for all pipelines unless this option is enabled.",
						Computed:            true,
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
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"days_to_keep_artifacts": schema.Int64Attribute{
						MarkdownDescription: "Number of days to keep artifacts, symbols and attachments.",
						Computed:            true,
					},
					"days_to_keep_pullrequest_runs": schema.Int64Attribute{
						MarkdownDescription: "Number of days to keep pull request runs.",
						Computed:            true,
					},
					"days_to_keep_runs": schema.Int64Attribute{
						MarkdownDescription: "Number of days to keep runs.",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (d *PipelineSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*clients.AzureDevOpsClient).PipelineClient
}

func (d *PipelineSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model PipelineSettingsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	pipelineSettings, err := d.client.GetPipelineSettings(ctx, model.ProjectId)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve pipeline settings", err.Error())
		return
	}

	retentionSettings, err := d.client.GetPipelineRetentionSettings(ctx, model.ProjectId)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve project retention settings", err.Error())
		return
	}

	model.General = &PipelineGeneralSettings{
		DisableClassicPipelineCreation:   pipelineSettings.DisableClassicPipelineCreation,
		EnforceJobAuthScope:              pipelineSettings.EnforceJobAuthScope,
		EnforceJobAuthScopeForReleases:   pipelineSettings.EnforceJobAuthScopeForReleases,
		EnforceReferencedRepoScopedToken: pipelineSettings.EnforceReferencedRepoScopedToken,
		EnforceSettableVar:               pipelineSettings.EnforceSettableVar,
		PublishPipelineMetadata:          pipelineSettings.PublishPipelineMetadata,
		StatusBadgesArePrivate:           pipelineSettings.StatusBadgesArePrivate,
	}
	model.Retention = &PipelineRetentionSettings{
		DaysToKeepArtifacts:       retentionSettings.PurgeArtifacts.Value,
		DaysToKeepPullRequestRuns: retentionSettings.PurgePullRequestRuns.Value,
		DaysToKeepRuns:            retentionSettings.PurgeRuns.Value,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
