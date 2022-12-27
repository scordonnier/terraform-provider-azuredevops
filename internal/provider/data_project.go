package provider

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

var _ datasource.DataSource = &TeamProjectDataSourceImpl{}

func TeamProjectDataSource() datasource.DataSource {
	return &TeamProjectDataSourceImpl{}
}

type TeamProjectDataSourceImpl struct {
	client *core.Client
}

type TeamProjectDataSourceModel struct {
	Name types.String `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

func (d *TeamProjectDataSourceImpl) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *TeamProjectDataSourceImpl) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *TeamProjectDataSourceImpl) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*clients.AzureDevOpsClient).CoreClient
}

func (d *TeamProjectDataSourceImpl) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TeamProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
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

	data.Id = types.StringValue(project.Id.String())
	data.Name = types.StringValue(*project.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
