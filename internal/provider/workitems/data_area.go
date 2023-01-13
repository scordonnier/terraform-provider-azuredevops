package workitems

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/workitems"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

var _ datasource.DataSource = &AreaDataSource{}

func NewAreaDataSource() datasource.DataSource {
	return &AreaDataSource{}
}

type AreaDataSource struct {
	client *workitems.Client
}

type AreaDataSourceModel struct {
	HasChildren *bool   `tfsdk:"has_children"`
	Name        *string `tfsdk:"name"`
	Path        string  `tfsdk:"path"`
	ProjectId   string  `tfsdk:"project_id"`
}

func (d *AreaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_area"
}

func (d *AreaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about an existing area within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"has_children": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the area has any child areas.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the area.",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The path of the area.",
				Required:            true,
				Validators: []validator.String{
					utils.StringNotEmptyValidator(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
		},
	}
}

func (d *AreaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*clients.AzureDevOpsClient).WorkItemClient
}

func (d *AreaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model AreaDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	area, err := d.client.GetArea(ctx, model.ProjectId, model.Path)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("Area at path '%s' does not exist", model.Path), err.Error())
			return
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Unable to find area ath path '%s'", model.Path), err.Error())
		return
	}

	model.HasChildren = area.HasChildren
	model.Name = area.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
