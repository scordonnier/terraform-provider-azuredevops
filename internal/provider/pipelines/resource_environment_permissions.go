package pipelines

import (
	"context"
	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/graph"
	clientSecurity "github.com/scordonnier/terraform-provider-azuredevops/internal/clients/security"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/provider/security"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/validators"
)

const (
	permissionNameAdminister    = "Administer"
	permissionNameCreate        = "Create"
	permissionNameManage        = "Manage"
	permissionNameManageHistory = "ManageHistory"
	permissionNameUse           = "Use"
	permissionNameView          = "View"
)

var _ resource.Resource = &EnvironmentPermissionsResource{}

func NewEnvironmentPermissionsResource() resource.Resource {
	return &EnvironmentPermissionsResource{}
}

type EnvironmentPermissionsResource struct {
	graphClient    *graph.Client
	securityClient *clientSecurity.Client
}

type EnvironmentPermissionsResourceModel struct {
	Id                  types.Int64            `tfsdk:"id"`
	Permissions         EnvironmentPermissions `tfsdk:"permissions"`
	PrincipalDescriptor types.String           `tfsdk:"principal_descriptor"`
	PrincipalName       string                 `tfsdk:"principal_name"`
	ProjectId           string                 `tfsdk:"project_id"`
}

type EnvironmentPermissions struct {
	Administer    string `tfsdk:"administer"`
	Create        string `tfsdk:"create"`
	Manage        string `tfsdk:"manage"`
	ManageHistory string `tfsdk:"manage_history"`
	Use           string `tfsdk:"use"`
	View          string `tfsdk:"view"`
}

func (r *EnvironmentPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_permissions"
}

func (r *EnvironmentPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets permissions on environments in Azure Pipelines. All permissions that currently exists will be overwritten.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the environment. If you omit the value, the permissions are applied to the environments page and by default all environments inherit permissions from there.",
				Optional:            true,
			},
			"permissions": schema.SingleNestedAttribute{
				MarkdownDescription: "The permissions to assign.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"administer": schema.StringAttribute{
						MarkdownDescription: "Sets the `Administer` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"create": schema.StringAttribute{
						MarkdownDescription: "Sets the `Create` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"manage": schema.StringAttribute{
						MarkdownDescription: "Sets the `Manage` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"manage_history": schema.StringAttribute{
						MarkdownDescription: "Sets the `ManageHistory` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"use": schema.StringAttribute{
						MarkdownDescription: "Sets the `Use` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"view": schema.StringAttribute{
						MarkdownDescription: "Sets the `View` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
				},
			},
			"principal_descriptor": schema.StringAttribute{
				MarkdownDescription: "The principal descriptor to assign the permissions.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"principal_name": schema.StringAttribute{
				MarkdownDescription: "The principal name to assign the permissions.",
				Required:            true,
				Validators: []validator.String{
					validators.StringNotEmpty(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					validators.UUID(),
				},
			},
		},
	}
}

func (r *EnvironmentPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.graphClient = req.ProviderData.(*clients.AzureDevOpsClient).GraphClient
	r.securityClient = req.ProviderData.(*clients.AzureDevOpsClient).SecurityClient
}

func (r *EnvironmentPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *EnvironmentPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetEnvironmentToken(model.ProjectId, int(model.Id.ValueInt64()))
	permissions := r.getPermissions(model)
	err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdEnvironment, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.setPermissions(model, []*security.PrincipalPermissions{permissions})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *EnvironmentPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetEnvironmentToken(model.ProjectId, int(model.Id.ValueInt64()))
	permissions, err := security.ReadPrincipalPermissions(ctx, clientSecurity.NamespaceIdEnvironment, token, r.securityClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve access control lists", err.Error())
		return
	}

	r.setPermissions(model, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *EnvironmentPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetEnvironmentToken(model.ProjectId, int(model.Id.ValueInt64()))
	permissions := r.getPermissions(model)
	err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdEnvironment, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.setPermissions(model, []*security.PrincipalPermissions{permissions})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *EnvironmentPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetEnvironmentToken(model.ProjectId, int(model.Id.ValueInt64()))
	err := r.securityClient.RemoveAccessControlEntries(ctx, clientSecurity.NamespaceIdEnvironment, token, []string{model.PrincipalDescriptor.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete permissions", err.Error())
		return
	}
}

// Private Methods

func (r *EnvironmentPermissionsResource) getPermissions(model *EnvironmentPermissionsResourceModel) *security.PrincipalPermissions {
	return &security.PrincipalPermissions{
		PrincipalDescriptor: model.PrincipalDescriptor.ValueString(),
		PrincipalName:       model.PrincipalName,
		Permissions: map[string]string{
			permissionNameAdminister:    model.Permissions.Administer,
			permissionNameCreate:        model.Permissions.Create,
			permissionNameManage:        model.Permissions.Manage,
			permissionNameManageHistory: model.Permissions.ManageHistory,
			permissionNameUse:           model.Permissions.Use,
			permissionNameView:          model.Permissions.View,
		},
	}
}

func (r *EnvironmentPermissionsResource) setPermissions(model *EnvironmentPermissionsResourceModel, p []*security.PrincipalPermissions) {
	if len(p) == 0 {
		return
	}

	principalPermissions := linq.From(p).FirstWith(func(p interface{}) bool {
		return p.(*security.PrincipalPermissions).PrincipalName == model.PrincipalName
	}).(*security.PrincipalPermissions)
	model.Permissions.Administer = principalPermissions.Permissions[permissionNameAdminister]
	model.Permissions.Create = principalPermissions.Permissions[permissionNameCreate]
	model.Permissions.Manage = principalPermissions.Permissions[permissionNameManage]
	model.Permissions.ManageHistory = principalPermissions.Permissions[permissionNameManageHistory]
	model.Permissions.Use = principalPermissions.Permissions[permissionNameUse]
	model.Permissions.View = principalPermissions.Permissions[permissionNameView]
	model.PrincipalDescriptor = types.StringValue(principalPermissions.PrincipalDescriptor)
	model.PrincipalName = principalPermissions.PrincipalName
}
