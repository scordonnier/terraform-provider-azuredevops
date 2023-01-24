package graph

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/graph"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/validators"
)

var _ datasource.DataSource = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

type UsersDataSource struct {
	client *graph.Client
}

type UsersDataSourceModel struct {
	ProjectId string                `tfsdk:"project_id"`
	Users     []UserDataSourceModel `tfsdk:"users"`
}

func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about existing users within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					validators.UUID(),
				},
			},
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "The list of users within the project.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
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
							Computed:            true,
						},
						"origin": schema.StringAttribute{
							MarkdownDescription: "The type of source provider for the user (eg. AD, AAD, MSA).",
							Computed:            true,
						},
						"origin_id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier from the system of origin.",
							Computed:            true,
						},
						"project_id": schema.StringAttribute{
							MarkdownDescription: "The project ID of the user.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*clients.AzureDevOpsClient).GraphClient
}

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model UsersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	groups, err := d.client.GetUsers(ctx, model.ProjectId, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve users", err.Error())
		return
	}

	var userModels []UserDataSourceModel
	for _, user := range *groups {
		userModels = append(userModels, UserDataSourceModel{
			Descriptor:  user.Descriptor,
			DisplayName: user.DisplayName,
			MailAddress: *user.MailAddress,
			Origin:      user.Origin,
			OriginId:    user.OriginId,
			ProjectId:   model.ProjectId,
		})
	}
	model.Users = userModels

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
