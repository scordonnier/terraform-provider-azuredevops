package graph

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/graph"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"strings"
)

var _ datasource.DataSource = &GroupDataSource{}

func NewGroupDataSource() datasource.DataSource {
	return &GroupDataSource{}
}

type GroupDataSource struct {
	client *graph.Client
}

type GroupDataSourceModel struct {
	Description *string `tfsdk:"description"`
	Descriptor  *string `tfsdk:"descriptor"`
	DisplayName string  `tfsdk:"display_name"`
	Name        *string `tfsdk:"name"`
	Origin      *string `tfsdk:"origin"`
	OriginId    *string `tfsdk:"origin_id"`
	ProjectId   string  `tfsdk:"project_id"`
}

func (d *GroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *GroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about an existing group within an Azure DevOps project.",
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
				Required:            true,
				Validators: []validator.String{
					utils.StringNotEmptyValidator(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the group.",
				Computed:            true,
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "The type of source provider for the group (eg. AD, AAD, MSA).",
				Computed:            true,
			},
			"origin_id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier from the system of origin.",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The project ID of the group.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
		},
	}
}

func (d *GroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*clients.AzureDevOpsClient).GraphClient
}

func (d *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model GroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	groups, err := d.client.GetGroups(ctx, model.ProjectId, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve groups", err.Error())
		return
	}

	var group *graph.GraphGroup
	for _, g := range *groups {
		if strings.EqualFold(*g.DisplayName, model.DisplayName) {
			group = &g
			break
		}
	}

	if group == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Group with name '%s' not found", model.DisplayName), "")
		return
	}

	model.Description = group.Description
	model.Descriptor = group.Descriptor
	model.Name = group.PrincipalName
	model.Origin = group.Origin
	model.OriginId = group.OriginId

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
