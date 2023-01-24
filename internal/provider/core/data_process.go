package core

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/validators"
)

var _ datasource.DataSource = &ProcessDataSource{}

func NewProcessDataSource() datasource.DataSource {
	return &ProcessDataSource{}
}

type ProcessDataSource struct {
	client *core.Client
}

type ProcessDataSourceModel struct {
	Description types.String `tfsdk:"description"`
	Name        string       `tfsdk:"name"`
	Id          types.String `tfsdk:"id"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
}

func (d *ProcessDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_process"
}

func (d *ProcessDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about an existing process within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the process.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the process.",
				Required:            true,
				Validators: []validator.String{
					validators.StringNotEmpty(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the process.",
				Computed:            true,
			},
			"is_default": schema.BoolAttribute{
				MarkdownDescription: "`true` if the process is the default process within the organization. Otherwise `false`.",
				Computed:            true,
			},
		},
	}
}

func (d *ProcessDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*clients.AzureDevOpsClient).CoreClient
}

func (d *ProcessDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model ProcessDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	process, err := d.client.GetProcess(ctx, model.Name)
	if err != nil {
		resp.Diagnostics.AddError("Process not found", err.Error())
		return
	}

	model.Description = types.StringValue(*process.Description)
	model.Id = types.StringValue(process.Id.String())
	model.IsDefault = types.BoolValue(*process.IsDefault)
	model.Name = *process.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
