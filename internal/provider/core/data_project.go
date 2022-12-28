package core

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

var _ datasource.DataSource = &TeamProjectDataSource{}

func NewTeamProjectDataSource() datasource.DataSource {
	return &TeamProjectDataSource{}
}

type TeamProjectDataSource struct {
	client *core.Client
}

type TeamProjectDataSourceModel struct {
	Name string       `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

func (d *TeamProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *TeamProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "", // TODO: Documentation
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "", // TODO: Documentation
				Computed:            true,
			},
		},
	}
}

func (d *TeamProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*clients.AzureDevOpsClient).CoreClient
}

func (d *TeamProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model TeamProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := model.Name
	if name == "" {
		resp.Diagnostics.AddError("Project name must not be empty", "")
		return
	}

	project, err := d.client.GetProject(ctx, name)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("Project with name '%s' does not exist", name), "")
			return
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Error looking up project with name '%s', %+v ", name, err), "")
		return
	}

	model.Id = types.StringValue(project.Id.String())
	model.Name = *project.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
