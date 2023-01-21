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
	Permissions         ProjectPermissions `tfsdk:"permissions"`
	PrincipalDescriptor types.String       `tfsdk:"principal_descriptor"`
	PrincipalName       string             `tfsdk:"principal_name"`
	ProjectId           string             `tfsdk:"project_id"`
}

type ProjectPermissions struct {
	Boards    ProjectBoardsPermissions    `tfsdk:"boards"`
	General   ProjectGeneralPermissions   `tfsdk:"general"`
	TestPlans ProjectTestPlansPermissions `tfsdk:"test_plans"`
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
		MarkdownDescription: "Sets permissions on projects within Azure DevOps. All permissions that currently exists will be overwritten.",
		Attributes: map[string]schema.Attribute{
			"permissions": schema.SingleNestedAttribute{
				MarkdownDescription: "The permissions to assign.",
				Required:            true,
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
					validators.StringNotEmptyValidator(),
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
	permissions := r.getPermissions(model)
	err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdProject, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.setPermissions(model, []*security.PrincipalPermissions{permissions})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ProjectPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetProjectToken(model.ProjectId)
	permissions, err := security.ReadPrincipalPermissions(ctx, clientSecurity.NamespaceIdProject, token, r.securityClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve access control lists", err.Error())
		return
	}

	r.setPermissions(model, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ProjectPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetProjectToken(model.ProjectId)
	permissions := r.getPermissions(model)
	err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdProject, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update permissions", err.Error())
		return
	}

	r.setPermissions(model, []*security.PrincipalPermissions{permissions})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ProjectPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetProjectToken(model.ProjectId)
	err := r.securityClient.RemoveAccessControlEntries(ctx, clientSecurity.NamespaceIdProject, token, []string{model.PrincipalDescriptor.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete permissions", err.Error())
		return
	}
}

// Private Methods

func (r *ProjectPermissionsResource) getPermissions(model *ProjectPermissionsResourceModel) *security.PrincipalPermissions {
	return &security.PrincipalPermissions{
		PrincipalDescriptor: model.PrincipalDescriptor.ValueString(),
		PrincipalName:       model.PrincipalName,
		Permissions: map[string]string{
			// Boards
			permissionNameBypassRules:               model.Permissions.Boards.BypassRules,
			permissionNameChangeProcess:             model.Permissions.Boards.ChangeProcess,
			permissionNameWorkItemDelete:            model.Permissions.Boards.WorkItemDelete,
			permissionNameWorkItemMove:              model.Permissions.Boards.WorkItemMove,
			permissionNameWorkItemPermanentlyDelete: model.Permissions.Boards.WorkItemPermanentlyDelete,
			// General
			permissionNameDelete:                model.Permissions.General.Delete,
			permissionNameManageProperties:      model.Permissions.General.ManageProperties,
			permissionNameRead:                  model.Permissions.General.Read,
			permissionNameRename:                model.Permissions.General.Rename,
			permissionNameSuppressNotifications: model.Permissions.General.SuppressNotifications,
			permissionNameUpdateVisibility:      model.Permissions.General.UpdateVisibility,
			permissionNameWrite:                 model.Permissions.General.Write,
			// Test Plans
			permissionNameDeleteTestResults:        model.Permissions.TestPlans.DeleteTestResults,
			permissionNameManageTestConfigurations: model.Permissions.TestPlans.ManageTestConfigurations,
			permissionNameManageTestEnvironments:   model.Permissions.TestPlans.ManageTestEnvironments,
			permissionNamePublishTestResults:       model.Permissions.TestPlans.PublishTestResults,
			permissionNameViewTestResults:          model.Permissions.TestPlans.ViewTestResults,
		},
	}
}

func (r *ProjectPermissionsResource) setPermissions(model *ProjectPermissionsResourceModel, p []*security.PrincipalPermissions) {
	if len(p) == 0 {
		return
	}

	principalPermissions := linq.From(p).FirstWith(func(p interface{}) bool {
		return p.(*security.PrincipalPermissions).PrincipalName == model.PrincipalName
	}).(*security.PrincipalPermissions)
	// Boards
	model.Permissions.Boards.BypassRules = principalPermissions.Permissions[permissionNameBypassRules]
	model.Permissions.Boards.ChangeProcess = principalPermissions.Permissions[permissionNameChangeProcess]
	model.Permissions.Boards.WorkItemDelete = principalPermissions.Permissions[permissionNameWorkItemDelete]
	model.Permissions.Boards.WorkItemMove = principalPermissions.Permissions[permissionNameWorkItemMove]
	model.Permissions.Boards.WorkItemPermanentlyDelete = principalPermissions.Permissions[permissionNameWorkItemPermanentlyDelete]
	// General
	model.Permissions.General.Delete = principalPermissions.Permissions[permissionNameDelete]
	model.Permissions.General.ManageProperties = principalPermissions.Permissions[permissionNameManageProperties]
	model.Permissions.General.Read = principalPermissions.Permissions[permissionNameRead]
	model.Permissions.General.Rename = principalPermissions.Permissions[permissionNameRename]
	model.Permissions.General.SuppressNotifications = principalPermissions.Permissions[permissionNameSuppressNotifications]
	model.Permissions.General.UpdateVisibility = principalPermissions.Permissions[permissionNameUpdateVisibility]
	model.Permissions.General.Write = principalPermissions.Permissions[permissionNameWrite]
	// Test Plans
	model.Permissions.TestPlans.DeleteTestResults = principalPermissions.Permissions[permissionNameDeleteTestResults]
	model.Permissions.TestPlans.ManageTestConfigurations = principalPermissions.Permissions[permissionNameManageTestConfigurations]
	model.Permissions.TestPlans.ManageTestEnvironments = principalPermissions.Permissions[permissionNameManageTestEnvironments]
	model.Permissions.TestPlans.PublishTestResults = principalPermissions.Permissions[permissionNamePublishTestResults]
	model.Permissions.TestPlans.ViewTestResults = principalPermissions.Permissions[permissionNameViewTestResults]

	model.PrincipalDescriptor = types.StringValue(principalPermissions.PrincipalDescriptor)
	model.PrincipalName = principalPermissions.PrincipalName
}
