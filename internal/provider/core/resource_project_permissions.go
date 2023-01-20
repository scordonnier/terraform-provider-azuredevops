package core

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
	permissionNameBypassRules               = "BYPASS_RULES"
	permissionNameChangeProcess             = "CHANGE_PROCESS"
	permissionNameDelete                    = "DELETE"
	permissionNameDeleteTestResults         = "DELETE_TEST_RESULTS"
	permissionNameManageProperties          = "MANAGE_PROPERTIES"
	permissionNameManageTestConfigurations  = "MANAGE_TEST_CONFIGURATIONS"
	permissionNameManageTestEnvironments    = "MANAGE_TEST_ENVIRONMENTS"
	permissionNamePublishTestResults        = "PUBLISH_TEST_RESULTS"
	permissionNameRead                      = "GENERIC_READ"
	permissionNameRename                    = "RENAME"
	permissionNameSuppressNotifications     = "SUPPRESS_NOTIFICATIONS"
	permissionNameUpdateVisibility          = "UPDATE_VISIBILITY"
	permissionNameViewTestResults           = "VIEW_TEST_RESULTS"
	permissionNameWorkItemDelete            = "WORK_ITEM_DELETE"
	permissionNameWorkItemMove              = "WORK_ITEM_MOVE"
	permissionNameWorkItemPermanentlyDelete = "WORK_ITEM_PERMANENTLY_DELETE"
	permissionNameWrite                     = "GENERIC_WRITE"
)

var _ resource.Resource = &ProjectPermissionsResource{}

func NewProjectPermissionsResource() resource.Resource {
	return &ProjectPermissionsResource{}
}

type ProjectPermissionsResource struct {
	graphClient    *graph.Client
	securityClient *clientSecurity.Client
}

type ProjectPermissionsResourceModel struct {
	Permissions []ProjectPermissions `tfsdk:"permissions"`
	ProjectId   string               `tfsdk:"project_id"`
}

type ProjectPermissions struct {
	Boards             ProjectBoardsPermissions    `tfsdk:"boards"`
	General            ProjectGeneralPermissions   `tfsdk:"general"`
	IdentityDescriptor types.String                `tfsdk:"identity_descriptor"`
	IdentityName       string                      `tfsdk:"identity_name"`
	TestPlans          ProjectTestPlansPermissions `tfsdk:"test_plans"`
}

type ProjectBoardsPermissions struct {
	BypassRules               string `tfsdk:"bypass_rules"`
	ChangeProcess             string `tfsdk:"change_process"`
	WorkItemDelete            string `tfsdk:"workitem_delete"`
	WorkItemMove              string `tfsdk:"workitem_move"`
	WorkItemPermanentlyDelete string `tfsdk:"workitem_permanently_delete"`
}

type ProjectGeneralPermissions struct {
	Delete                string `tfsdk:"delete"`
	ManageProperties      string `tfsdk:"manage_properties"`
	Rename                string `tfsdk:"rename"`
	Read                  string `tfsdk:"read"`
	SuppressNotifications string `tfsdk:"suppress_notifications"`
	UpdateVisibility      string `tfsdk:"update_visibility"`
	Write                 string `tfsdk:"write"`
}

type ProjectTestPlansPermissions struct {
	DeleteTestResults        string `tfsdk:"delete_test_results"`
	ManageTestConfigurations string `tfsdk:"manage_test_configurations"`
	ManageTestEnvironments   string `tfsdk:"manage_test_environments"`
	PublishTestResults       string `tfsdk:"publish_test_results"`
	ViewTestResults          string `tfsdk:"view_test_results"`
}

func (r *ProjectPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_permissions"
}

func (r *ProjectPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets permissions on projects within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"permissions": schema.ListNestedAttribute{
				MarkdownDescription: "The permissions to assign.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"boards": schema.SingleNestedAttribute{
							Required: true,
							Attributes: map[string]schema.Attribute{
								"bypass_rules": schema.StringAttribute{
									MarkdownDescription: "Sets the `BYPASS_RULES` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"change_process": schema.StringAttribute{
									MarkdownDescription: "Sets the `CHANGE_PROCESS` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"workitem_delete": schema.StringAttribute{
									MarkdownDescription: "Sets the `WORK_ITEM_DELETE` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"workitem_move": schema.StringAttribute{
									MarkdownDescription: "Sets the `WORK_ITEM_MOVE` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"workitem_permanently_delete": schema.StringAttribute{
									MarkdownDescription: "Sets the `WORK_ITEM_PERMANENTLY_DELETE` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
							},
						},
						"general": schema.SingleNestedAttribute{
							Required: true,
							Attributes: map[string]schema.Attribute{
								"delete": schema.StringAttribute{
									MarkdownDescription: "Sets the `DELETE` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"manage_properties": schema.StringAttribute{
									MarkdownDescription: "Sets the `MANAGE_PROPERTIES` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"rename": schema.StringAttribute{
									MarkdownDescription: "Sets the `RENAME` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"read": schema.StringAttribute{
									MarkdownDescription: "Sets the `GENERIC_READ` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"suppress_notifications": schema.StringAttribute{
									MarkdownDescription: "Sets the `SUPPRESS_NOTIFICATIONS` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"update_visibility": schema.StringAttribute{
									MarkdownDescription: "Sets the `UPDATE_VISIBILITY` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"write": schema.StringAttribute{
									MarkdownDescription: "Sets the `GENERIC_WRITE` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
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
						"test_plans": schema.SingleNestedAttribute{
							Required: true,
							Attributes: map[string]schema.Attribute{
								"delete_test_results": schema.StringAttribute{
									MarkdownDescription: "Sets the `DELETE_TEST_RESULTS` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"manage_test_configurations": schema.StringAttribute{
									MarkdownDescription: "Sets the `MANAGE_TEST_CONFIGURATIONS` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"manage_test_environments": schema.StringAttribute{
									MarkdownDescription: "Sets the `MANAGE_TEST_ENVIRONMENTS` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"publish_test_results": schema.StringAttribute{
									MarkdownDescription: "Sets the `PUBLISH_TEST_RESULTS` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
								"view_test_results": schema.StringAttribute{
									MarkdownDescription: "Sets the `VIEW_TEST_RESULTS` permission for the identity. Must be `notset`, `allow` or `deny`.",
									Required:            true,
									Validators: []validator.String{
										validators.PermissionsValidator(),
									},
								},
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

func (r *ProjectPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.graphClient = req.ProviderData.(*clients.AzureDevOpsClient).GraphClient
	r.securityClient = req.ProviderData.(*clients.AzureDevOpsClient).SecurityClient
}

func (r *ProjectPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ProjectPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetProjectToken(model.ProjectId)
	permissions := r.getPermissions(&model.Permissions)
	for _, permission := range permissions {
		err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdProject, token, permission, r.securityClient, r.graphClient)
		if err != nil {
			resp.Diagnostics.AddError("Unable to create permissions", err.Error())
			return
		}
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ProjectPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetProjectToken(model.ProjectId)
	permissions, err := security.ReadIdentityPermissions(ctx, clientSecurity.NamespaceIdProject, token, r.securityClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve access control lists", err.Error())
		return
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ProjectPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetProjectToken(model.ProjectId)
	permissions := r.getPermissions(&model.Permissions)
	for _, permission := range permissions {
		err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdProject, token, permission, r.securityClient, r.graphClient)
		if err != nil {
			resp.Diagnostics.AddError("Unable to update permissions", err.Error())
			return
		}
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ProjectPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetProjectToken(model.ProjectId)
	var descriptors []string
	for _, permission := range model.Permissions {
		descriptors = append(descriptors, permission.IdentityDescriptor.ValueString())
	}
	err := r.securityClient.RemoveAccessControlEntries(ctx, clientSecurity.NamespaceIdProject, token, descriptors)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete permissions", err.Error())
		return
	}
}

// Private Methods

func (r *ProjectPermissionsResource) getPermissions(p *[]ProjectPermissions) []*security.IdentityPermissions {
	var permissions []*security.IdentityPermissions
	for _, permission := range *p {
		permissions = append(permissions, &security.IdentityPermissions{
			IdentityDescriptor: permission.IdentityDescriptor.ValueString(),
			IdentityName:       permission.IdentityName,
			Permissions: map[string]string{
				// Boards
				permissionNameBypassRules:               permission.Boards.BypassRules,
				permissionNameChangeProcess:             permission.Boards.ChangeProcess,
				permissionNameWorkItemDelete:            permission.Boards.WorkItemDelete,
				permissionNameWorkItemMove:              permission.Boards.WorkItemMove,
				permissionNameWorkItemPermanentlyDelete: permission.Boards.WorkItemPermanentlyDelete,
				// General
				permissionNameDelete:                permission.General.Delete,
				permissionNameManageProperties:      permission.General.ManageProperties,
				permissionNameRead:                  permission.General.Read,
				permissionNameRename:                permission.General.Rename,
				permissionNameSuppressNotifications: permission.General.SuppressNotifications,
				permissionNameUpdateVisibility:      permission.General.UpdateVisibility,
				permissionNameWrite:                 permission.General.Write,
				// Test Plans
				permissionNameDeleteTestResults:        permission.TestPlans.DeleteTestResults,
				permissionNameManageTestConfigurations: permission.TestPlans.ManageTestConfigurations,
				permissionNameManageTestEnvironments:   permission.TestPlans.ManageTestEnvironments,
				permissionNamePublishTestResults:       permission.TestPlans.PublishTestResults,
				permissionNameViewTestResults:          permission.TestPlans.ViewTestResults,
			},
		})
	}
	return permissions
}

func (r *ProjectPermissionsResource) updatePermissions(p1 *[]ProjectPermissions, p2 []*security.IdentityPermissions) {
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
		permission.Boards.BypassRules = identityPermissions.Permissions[permissionNameBypassRules]
		permission.Boards.ChangeProcess = identityPermissions.Permissions[permissionNameChangeProcess]
		permission.Boards.WorkItemDelete = identityPermissions.Permissions[permissionNameWorkItemDelete]
		permission.Boards.WorkItemMove = identityPermissions.Permissions[permissionNameWorkItemMove]
		permission.Boards.WorkItemPermanentlyDelete = identityPermissions.Permissions[permissionNameWorkItemPermanentlyDelete]
		permission.General.Delete = identityPermissions.Permissions[permissionNameDelete]
		permission.General.ManageProperties = identityPermissions.Permissions[permissionNameManageProperties]
		permission.General.Read = identityPermissions.Permissions[permissionNameRead]
		permission.General.Rename = identityPermissions.Permissions[permissionNameRename]
		permission.General.SuppressNotifications = identityPermissions.Permissions[permissionNameSuppressNotifications]
		permission.General.UpdateVisibility = identityPermissions.Permissions[permissionNameUpdateVisibility]
		permission.General.Write = identityPermissions.Permissions[permissionNameWrite]
		permission.TestPlans.DeleteTestResults = identityPermissions.Permissions[permissionNameDeleteTestResults]
		permission.TestPlans.ManageTestConfigurations = identityPermissions.Permissions[permissionNameManageTestConfigurations]
		permission.TestPlans.ManageTestEnvironments = identityPermissions.Permissions[permissionNameManageTestEnvironments]
		permission.TestPlans.PublishTestResults = identityPermissions.Permissions[permissionNamePublishTestResults]
		permission.TestPlans.ViewTestResults = identityPermissions.Permissions[permissionNameViewTestResults]
	}
}
