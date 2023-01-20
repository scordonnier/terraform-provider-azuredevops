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
	Id          types.Int64              `tfsdk:"id"`
	ProjectId   string                   `tfsdk:"project_id"`
	Permissions []EnvironmentPermissions `tfsdk:"permissions"`
}

type EnvironmentPermissions struct {
	IdentityDescriptor types.String `tfsdk:"identity_descriptor"`
	IdentityName       string       `tfsdk:"identity_name"`
	View               string       `tfsdk:"view"`
	Manage             string       `tfsdk:"manage"`
	ManageHistory      string       `tfsdk:"manage_history"`
	Administer         string       `tfsdk:"administer"`
	Use                string       `tfsdk:"use"`
	Create             string       `tfsdk:"create"`
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
			"permissions": schema.ListNestedAttribute{
				MarkdownDescription: "The permissions to assign.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"administer": schema.StringAttribute{
							MarkdownDescription: "Sets the `Administer` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"create": schema.StringAttribute{
							MarkdownDescription: "Sets the `Create` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"identity_descriptor": schema.StringAttribute{
							MarkdownDescription: "The identity descriptor to assign the permissions.",
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"identity_name": schema.StringAttribute{
							MarkdownDescription: "The identity name to assign the permissions.",
							Required:            true,
							Validators: []validator.String{
								validators.StringNotEmptyValidator(),
							},
						},
						"manage": schema.StringAttribute{
							MarkdownDescription: "Sets the `Manage` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"manage_history": schema.StringAttribute{
							MarkdownDescription: "Sets the `ManageHistory` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"use": schema.StringAttribute{
							MarkdownDescription: "Sets the `Use` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"view": schema.StringAttribute{
							MarkdownDescription: "Sets the `View` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
					},
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					validators.UUIDStringValidator(),
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
	permissions := r.getPermissions(&model.Permissions)
	err := security.CreateOrUpdateAccessControlList(ctx, clientSecurity.NamespaceIdEnvironment, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *EnvironmentPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetEnvironmentToken(model.ProjectId, int(model.Id.ValueInt64()))
	permissions, err := security.ReadIdentityPermissions(ctx, clientSecurity.NamespaceIdEnvironment, token, r.securityClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve access control lists", err.Error())
		return
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *EnvironmentPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetEnvironmentToken(model.ProjectId, int(model.Id.ValueInt64()))
	permissions := r.getPermissions(&model.Permissions)
	err := security.CreateOrUpdateAccessControlList(ctx, clientSecurity.NamespaceIdEnvironment, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update permissions", err.Error())
		return
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *EnvironmentPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetEnvironmentToken(model.ProjectId, int(model.Id.ValueInt64()))
	err := r.securityClient.RemoveAccessControlLists(ctx, clientSecurity.NamespaceIdEnvironment, token)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete environment permissions", err.Error())
	}
}

// Private Methods

func (r *EnvironmentPermissionsResource) getPermissions(p *[]EnvironmentPermissions) []*security.IdentityPermissions {
	var permissions []*security.IdentityPermissions
	for _, permission := range *p {
		permissions = append(permissions, &security.IdentityPermissions{
			IdentityDescriptor: permission.IdentityDescriptor.ValueString(),
			IdentityName:       permission.IdentityName,
			Permissions: map[string]string{
				permissionNameAdminister:    permission.Administer,
				permissionNameCreate:        permission.Create,
				permissionNameManage:        permission.Manage,
				permissionNameManageHistory: permission.ManageHistory,
				permissionNameUse:           permission.Use,
				permissionNameView:          permission.View,
			},
		})
	}
	return permissions
}

func (r *EnvironmentPermissionsResource) updatePermissions(p1 *[]EnvironmentPermissions, p2 []*security.IdentityPermissions) {
	if len(p2) == 0 {
		return
	}

	for index := range *p1 {
		permission := &(*p1)[index]
		identityPermissions := linq.From(p2).FirstWith(func(p interface{}) bool {
			return p.(*security.IdentityPermissions).IdentityName == permission.IdentityName
		}).(*security.IdentityPermissions)
		permission.IdentityDescriptor = types.StringValue(identityPermissions.IdentityDescriptor)
		permission.IdentityName = identityPermissions.IdentityName
		permission.Administer = identityPermissions.Permissions[permissionNameAdminister]
		permission.Create = identityPermissions.Permissions[permissionNameCreate]
		permission.Manage = identityPermissions.Permissions[permissionNameManage]
		permission.ManageHistory = identityPermissions.Permissions[permissionNameManageHistory]
		permission.Use = identityPermissions.Permissions[permissionNameUse]
		permission.View = identityPermissions.Permissions[permissionNameView]
	}
}
