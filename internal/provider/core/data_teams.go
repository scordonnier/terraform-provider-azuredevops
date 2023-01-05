package core

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

var _ datasource.DataSource = &TeamsDataSource{}

func NewTeamsDataSource() datasource.DataSource {
	return &TeamsDataSource{}
}

type TeamsDataSource struct {
	client *core.Client
}

type TeamsDataSourceModel struct {
	ProjectId string                `tfsdk:"project_id"`
	Teams     []TeamDataSourceModel `tfsdk:"teams"`
}

func (d *TeamsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

func (d *TeamsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about existing teams within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
			"teams": schema.ListNestedAttribute{
				MarkdownDescription: "The list of teams within the project.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							MarkdownDescription: "The description of the team.",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "The ID of the team.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the team.",
							Computed:            true,
						},
						"project_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the project.",
							Computed:            true,
						},
						"project_name": schema.StringAttribute{
							MarkdownDescription: "The name of the project hosting the team.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *TeamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*clients.AzureDevOpsClient).CoreClient
}

func (d *TeamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model TeamsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	teams, err := d.client.GetTeams(ctx, model.ProjectId)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve teams", err.Error())
		return
	}

	var teamModels []TeamDataSourceModel
	for _, team := range *teams {
		teamModels = append(teamModels, TeamDataSourceModel{
			Description: types.StringValue(*team.Description),
			Id:          types.StringValue(team.Id.String()),
			Name:        *team.Name,
			ProjectId:   team.ProjectId.String(),
			ProjectName: types.StringValue(*team.ProjectName),
		})
	}
	model.Teams = teamModels

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
