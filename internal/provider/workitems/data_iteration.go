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

var _ datasource.DataSource = &IterationDataSource{}

func NewIterationDataSource() datasource.DataSource {
	return &IterationDataSource{}
}

type IterationDataSource struct {
	client *workitems.Client
}

type IterationDataSourceModel struct {
	HasChildren *bool   `tfsdk:"has_children"`
	Name        *string `tfsdk:"name"`
	Path        string  `tfsdk:"path"`
	ProjectId   string  `tfsdk:"project_id"`
}

func (d *IterationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iteration"
}

func (d *IterationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about an existing iteration within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"has_children": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the iteration has any child iterations.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the iteration.",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The path of the iteration.",
				Required:            true,
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

func (d *IterationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*clients.AzureDevOpsClient).WorkItemClient
}

func (d *IterationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model IterationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	iteration, err := d.client.GetIteration(ctx, model.ProjectId, model.Path)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("Iteration at path '%s' does not exist", model.Path), err.Error())
			return
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Unable to find iteration ath path '%s'", model.Path), err.Error())
		return
	}

	model.HasChildren = iteration.HasChildren
	model.Name = iteration.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
