package graph

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/graph"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

var _ datasource.DataSource = &GroupsDataSource{}

func NewGroupsDataSource() datasource.DataSource {
	return &GroupsDataSource{}
}

type GroupsDataSource struct {
	client *graph.Client
}

type GroupsDataSourceModel struct {
	ProjectId string                 `tfsdk:"project_id"`
	Groups    []GroupDataSourceModel `tfsdk:"groups"`
}

func (d *GroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *GroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about existing groups within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
			"groups": schema.ListNestedAttribute{
				MarkdownDescription: "The list of groups within the project.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							MarkdownDescription: "The description of the group.",
							Computed:            true,
						},
						"descriptor": schema.StringAttribute{
							MarkdownDescription: "The descriptor of the group.",
							Computed:            true,
						},
						"display_name": schema.StringAttribute{
							MarkdownDescription: "The display name of the group.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the group.",
							Computed:            true,
						},
						"origin": schema.StringAttribute{
							MarkdownDescription: "The type of source provider for the group (eg. AD, AAD, MSA).",
							Computed:            true,
						},
						"project_id": schema.StringAttribute{
							MarkdownDescription: "The project ID of the group.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *GroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*clients.AzureDevOpsClient).GraphClient
}

func (d *GroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model GroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	groups, err := d.client.GetGroups(ctx, model.ProjectId, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve groups", err.Error())
		return
	}

	var groupModels []GroupDataSourceModel
	for _, group := range *groups {
		groupModels = append(groupModels, GroupDataSourceModel{
			Description: group.Description,
			Descriptor:  group.Descriptor,
			DisplayName: *group.DisplayName,
			Name:        group.PrincipalName,
			Origin:      group.Origin,
			ProjectId:   model.ProjectId,
		})
	}
	model.Groups = groupModels

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
