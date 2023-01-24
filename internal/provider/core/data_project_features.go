package core

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/validators"
)

var _ datasource.DataSource = &ProjectFeaturesDataSource{}

func NewProjectFeaturesDataSource() datasource.DataSource {
	return &ProjectFeaturesDataSource{}
}

type ProjectFeaturesDataSource struct {
	client *core.Client
}

type ProjectFeaturesDataSourceModel struct {
	Artifacts    *string `tfsdk:"artifacts"`
	Boards       *string `tfsdk:"boards"`
	Pipelines    *string `tfsdk:"pipelines"`
	ProjectId    string  `tfsdk:"project_id"`
	Repositories *string `tfsdk:"repositories"`
	TestPlans    *string `tfsdk:"testplans"`
}

func (d *ProjectFeaturesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_features"
}

func (d *ProjectFeaturesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about features of an existing project within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"artifacts": schema.StringAttribute{
				MarkdownDescription: "If enabled, gives access to Azure Artifacts.",
				Computed:            true,
			},
			"boards": schema.StringAttribute{
				MarkdownDescription: "If enabled, gives access to Azure Boards.",
				Computed:            true,
			},
			"pipelines": schema.StringAttribute{
				MarkdownDescription: "If enabled, gives access to Azure Pipelines.",
				Computed:            true,
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
				Computed:            true,
			},
			"testplans": schema.StringAttribute{
				MarkdownDescription: "If enabled, gives access to Azure Test Plans.",
				Computed:            true,
			},
		},
	}
}

func (d *ProjectFeaturesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*clients.AzureDevOpsClient).CoreClient
}

func (d *ProjectFeaturesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model ProjectFeaturesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	features, err := d.client.GetProjectFeatures(ctx, model.ProjectId)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError("Project does not exist", err.Error())
			return
		}

		resp.Diagnostics.AddError("Unable to retrieve project features", err.Error())
		return
	}

	featureStates := *features.FeatureStates
	model.Artifacts = featureStates[core.ProjectFeatureArtifacts].State
	model.Boards = featureStates[core.ProjectFeatureBoards].State
	model.Pipelines = featureStates[core.ProjectFeaturePipelines].State
	model.Repositories = featureStates[core.ProjectFeatureRepositories].State
	model.TestPlans = featureStates[core.ProjectFeatureTestPlans].State

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
