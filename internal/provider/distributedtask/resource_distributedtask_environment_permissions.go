package distributedtask

import (
	"context"
	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	clientSecurity "github.com/scordonnier/terraform-provider-azuredevops/internal/clients/security"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/provider/security"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

const (
	permissionNameAdminister    = "Administer"
	permissionNameCreate        = "Create"
	permissionNameManage        = "Manage"
	permissionNameManageHistory = "ManageHistory"
	permissionNameUse           = "Use"
	permissionNameView          = "View"
)

var _ resource.Resource = &ResourceEnvironmentPermissions{}
var _ resource.ResourceWithImportState = &ResourceEnvironmentPermissions{}

func NewResourceEnvironmentPermissions() resource.Resource {
	return &ResourceEnvironmentPermissions{}
}

type ResourceEnvironmentPermissions struct {
	client *clientSecurity.Client
}

type ResourceEnvironmentPermissionsModel struct {
	Id          types.Int64              `tfsdk:"id"`
	ProjectId   string                   `tfsdk:"project_id"`
	Permissions []EnvironmentPermissions `tfsdk:"permissions"`
}

type EnvironmentPermissions struct {
	IdentityDescriptor types.String `tfsdk:"identity_descriptor"`
	IdentityName       string       `tfsdk:"identity_name"`
	IdentityType       string       `tfsdk:"identity_type"`
	View               string       `tfsdk:"view"`
	Manage             string       `tfsdk:"manage"`
	ManageHistory      string       `tfsdk:"manage_history"`
	Administer         string       `tfsdk:"administer"`
	Use                string       `tfsdk:"use"`
	Create             string       `tfsdk:"create"`
}

func (r *ResourceEnvironmentPermissions) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_distributedtask_environment_permissions"
}

func (r *ResourceEnvironmentPermissions) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets permissions on environments in Azure Pipelines. All permissions that currently exists will be overwritten.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the environment. If you omit the value, the permissions are applied to the environments page and by default all environments inherit permissions from there.",
				Optional:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"permissions": schema.ListNestedBlock{
				MarkdownDescription: "The permissions to assign.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"administer": schema.StringAttribute{
							MarkdownDescription: "Sets the `Administer` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("notset", "allow", "deny"),
							},
						},
						"create": schema.StringAttribute{
							MarkdownDescription: "Sets the `Create` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("notset", "allow", "deny"),
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
						},
						"identity_type": schema.StringAttribute{
							MarkdownDescription: "The identity type to assign the permissions. Must be `group`  or `user`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("group", "user"),
							},
						},
						"manage": schema.StringAttribute{
							MarkdownDescription: "Sets the `Manage` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("notset", "allow", "deny"),
							},
						},
						"manage_history": schema.StringAttribute{
							MarkdownDescription: "Sets the `ManageHistory` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("notset", "allow", "deny"),
							},
						},
						"use": schema.StringAttribute{
							MarkdownDescription: "Sets the `Use` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("notset", "allow", "deny"),
							},
						},
						"view": schema.StringAttribute{
							MarkdownDescription: "Sets the `View` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOfCaseInsensitive("notset", "allow", "deny"),
							},
						},
					},
				},
			},
		},
	}
}

func (r *ResourceEnvironmentPermissions) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).SecurityClient
}

func (r *ResourceEnvironmentPermissions) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ResourceEnvironmentPermissionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.client.GetEnvironmentToken(int(model.Id.ValueInt64()), model.ProjectId)
	permissions := r.getIdentityPermissions(&model.Permissions)
	err := security.CreateOrUpdateIdentityPermissions(ctx, clientSecurity.NamespaceIdEnvironment, token, permissions, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.updateEnvironmentPermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceEnvironmentPermissions) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ResourceEnvironmentPermissionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.client.GetEnvironmentToken(int(model.Id.ValueInt64()), model.ProjectId)
	permissions, err := security.ReadIdentityPermissions(ctx, clientSecurity.NamespaceIdEnvironment, token, r.client)
	if err != nil {
		if permissions != nil && len(permissions) == 0 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Unable to retrieve access control lists", err.Error())
		return
	}

	r.updateEnvironmentPermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceEnvironmentPermissions) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ResourceEnvironmentPermissionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.client.GetEnvironmentToken(int(model.Id.ValueInt64()), model.ProjectId)
	permissions := r.getIdentityPermissions(&model.Permissions)
	err := security.CreateOrUpdateIdentityPermissions(ctx, clientSecurity.NamespaceIdEnvironment, token, permissions, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update permissions", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ResourceEnvironmentPermissions) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ResourceEnvironmentPermissionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.client.GetEnvironmentToken(int(model.Id.ValueInt64()), model.ProjectId)
	err := r.client.RemoveAccessControlLists(ctx, clientSecurity.NamespaceIdEnvironment, token)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete environment permissions", err.Error())
	}
}

func (r *ResourceEnvironmentPermissions) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ResourceEnvironmentPermissions) getIdentityPermissions(p *[]EnvironmentPermissions) []*security.IdentityPermissions {
	var permissions []*security.IdentityPermissions
	for _, permission := range *p {
		permissions = append(permissions, &security.IdentityPermissions{
			IdentityDescriptor: permission.IdentityDescriptor.ValueString(),
			IdentityName:       permission.IdentityName,
			IdentityType:       permission.IdentityType,
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

func (r *ResourceEnvironmentPermissions) updateEnvironmentPermissions(p1 *[]EnvironmentPermissions, p2 []*security.IdentityPermissions) {
	for index := range *p1 {
		permission := &(*p1)[index]
		identityPermissions := linq.From(p2).FirstWith(func(p interface{}) bool {
			return p.(*security.IdentityPermissions).IdentityName == permission.IdentityName
		}).(*security.IdentityPermissions)
		permission.IdentityDescriptor = types.StringValue(identityPermissions.IdentityDescriptor)
		permission.IdentityName = identityPermissions.IdentityName
		permission.IdentityType = identityPermissions.IdentityType
		permission.Administer = identityPermissions.Permissions[permissionNameAdminister]
		permission.Create = identityPermissions.Permissions[permissionNameCreate]
		permission.Manage = identityPermissions.Permissions[permissionNameManage]
		permission.ManageHistory = identityPermissions.Permissions[permissionNameManageHistory]
		permission.Use = identityPermissions.Permissions[permissionNameUse]
		permission.View = identityPermissions.Permissions[permissionNameView]
	}
}
