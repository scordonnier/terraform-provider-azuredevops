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

var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	client *graph.Client
}

type UserDataSourceModel struct {
	Descriptor  *string `tfsdk:"descriptor"`
	DisplayName *string `tfsdk:"display_name"`
	MailAddress string  `tfsdk:"mail_address"`
	Origin      *string `tfsdk:"origin"`
	ProjectId   string  `tfsdk:"project_id"`
}

func (d *UserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about an existing user within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"descriptor": schema.StringAttribute{
				MarkdownDescription: "The descriptor of the user.",
				Computed:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the user.",
				Computed:            true,
			},
			"mail_address": schema.StringAttribute{
				MarkdownDescription: "The mail address of the user.",
				Required:            true,
				Validators: []validator.String{
					utils.StringNotEmptyValidator(),
				},
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "The type of source provider for the user (eg. AD, AAD, MSA).",
				Computed:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The project ID of the user.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
		},
	}
}

func (d *UserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*clients.AzureDevOpsClient).GraphClient
}

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model UserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	users, err := d.client.GetUsers(ctx, model.ProjectId, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve users", err.Error())
		return
	}

	var user *graph.GraphUser
	for _, u := range *users {
		if strings.EqualFold(*u.MailAddress, model.MailAddress) {
			user = &u
			break
		}
	}

	if user == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("User with mail address '%s' not found", model.MailAddress), "")
		return
	}

	model.Descriptor = user.Descriptor
	model.DisplayName = user.DisplayName
	model.Origin = user.Origin

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
