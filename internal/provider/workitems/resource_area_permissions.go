package workitems

import (
	"context"
	"fmt"
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
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/workitems"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/provider/security"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/validators"
	"strings"
)

var _ resource.Resource = &AreaPermissionsResource{}

func NewAreaPermissionsResource() resource.Resource {
	return &AreaPermissionsResource{}
}

type AreaPermissionsResource struct {
	graphClient     *graph.Client
	securityClient  *clientSecurity.Client
	workItemsClient *workitems.Client
}

type AreaPermissionsResourceModel struct {
	Path        string            `tfsdk:"path"`
	Permissions []AreaPermissions `tfsdk:"permissions"`
	ProjectId   string            `tfsdk:"project_id"`
}

type AreaPermissions struct {
	Create             string       `tfsdk:"create"`
	Delete             string       `tfsdk:"delete"`
	IdentityDescriptor types.String `tfsdk:"identity_descriptor"`
	IdentityName       string       `tfsdk:"identity_name"`
	ManageTestPlans    string       `tfsdk:"manage_test_plans"`
	ManageTestSuites   string       `tfsdk:"manage_test_suites"`
	Read               string       `tfsdk:"read"`
	WorkItemsRead      string       `tfsdk:"workitems_read"`
	WorkItemsWrite     string       `tfsdk:"workitems_write"`
	Write              string       `tfsdk:"write"`
}

func (r *AreaPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_area_permissions"
}

func (r *AreaPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets permissions on areas of an existing project within Azure DevOps. All permissions that currently exists will be overwritten.",
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				MarkdownDescription: "The path of the area.",
				Required:            true,
			},
			"permissions": schema.ListNestedAttribute{
				MarkdownDescription: "The permissions to assign.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"create": schema.StringAttribute{
							MarkdownDescription: "Sets the `CREATE_CHILDREN` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"delete": schema.StringAttribute{
							MarkdownDescription: "Sets the `DELETE` permission for the identity. Must be `notset`, `allow` or `deny`.",
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
						"manage_test_plans": schema.StringAttribute{
							MarkdownDescription: "Sets the `MANAGE_TEST_PLANS` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"manage_test_suites": schema.StringAttribute{
							MarkdownDescription: "Sets the `MANAGE_TEST_SUITES` permission for the identity. Must be `notset`, `allow` or `deny`.",
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
						"workitems_read": schema.StringAttribute{
							MarkdownDescription: "Sets the `WORK_ITEM_READ` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"workitems_write": schema.StringAttribute{
							MarkdownDescription: "Sets the `WORK_ITEM_WRITE` permission for the identity. Must be `notset`, `allow` or `deny`.",
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

func (r *AreaPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.graphClient = req.ProviderData.(*clients.AzureDevOpsClient).GraphClient
	r.securityClient = req.ProviderData.(*clients.AzureDevOpsClient).SecurityClient
	r.workItemsClient = req.ProviderData.(*clients.AzureDevOpsClient).WorkItemsClient
}

func (r *AreaPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *AreaPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token, err := r.getToken(ctx, model.ProjectId, model.Path)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve token", err.Error())
		return
	}

	permissions := r.getPermissions(&model.Permissions)
	err = security.CreateOrUpdateIdentityPermissions(ctx, clientSecurity.NamespaceIdCSS, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AreaPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *AreaPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token, err := r.getToken(ctx, model.ProjectId, model.Path)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve token", err.Error())
		return
	}

	permissions, err := security.ReadIdentityPermissions(ctx, clientSecurity.NamespaceIdCSS, token, r.securityClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve access control lists", err.Error())
		return
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AreaPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *AreaPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token, err := r.getToken(ctx, model.ProjectId, model.Path)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve token", err.Error())
		return
	}

	permissions := r.getPermissions(&model.Permissions)
	err = security.CreateOrUpdateIdentityPermissions(ctx, clientSecurity.NamespaceIdCSS, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update permissions", err.Error())
		return
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AreaPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *AreaPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token, err := r.getToken(ctx, model.ProjectId, model.Path)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve token", err.Error())
		return
	}

	err = r.securityClient.RemoveAccessControlLists(ctx, clientSecurity.NamespaceIdCSS, token)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete environment permissions", err.Error())
	}
}

// Private Methods

func (r *AreaPermissionsResource) getPermissions(p *[]AreaPermissions) []*security.IdentityPermissions {
	var permissions []*security.IdentityPermissions
	for _, permission := range *p {
		permissions = append(permissions, &security.IdentityPermissions{
			IdentityDescriptor: permission.IdentityDescriptor.ValueString(),
			IdentityName:       permission.IdentityName,
			Permissions: map[string]string{
				permissionNameCreate:           permission.Create,
				permissionNameDelete:           permission.Delete,
				permissionNameManageTestPlans:  permission.ManageTestPlans,
				permissionNameManageTestSuites: permission.ManageTestSuites,
				permissionNameRead:             permission.Read,
				permissionNameWorkItemsRead:    permission.WorkItemsRead,
				permissionNameWorkItemsWrite:   permission.WorkItemsWrite,
				permissionNameWrite:            permission.Write,
			},
		})
	}
	return permissions
}

func (r *AreaPermissionsResource) getToken(ctx context.Context, projectId string, path string) (string, error) {
	var tokens []string
	if path != "" {
		area, err := r.workItemsClient.GetArea(ctx, projectId, "")
		if err != nil {
			return "", err
		}

		tokens = append(tokens, r.securityClient.GetClassificationNodeToken(area.Identifier.String()))
	}

	components := strings.Split(path, "/")
	for i, component := range components {
		tokenPath := component
		if i > 0 {
			tokenPath = fmt.Sprintf("%s/%s", strings.Join(components[:i], "/"), component)
		}
		iteration, err := r.workItemsClient.GetArea(ctx, projectId, tokenPath)
		if err != nil {
			return "", err
		}

		tokens = append(tokens, r.securityClient.GetClassificationNodeToken(iteration.Identifier.String()))
	}

	return strings.Join(tokens, ":"), nil
}

func (r *AreaPermissionsResource) updatePermissions(p1 *[]AreaPermissions, p2 []*security.IdentityPermissions) {
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
		permission.Create = identityPermissions.Permissions[permissionNameCreate]
		permission.Delete = identityPermissions.Permissions[permissionNameDelete]
		permission.ManageTestPlans = identityPermissions.Permissions[permissionNameManageTestPlans]
		permission.ManageTestSuites = identityPermissions.Permissions[permissionNameManageTestSuites]
		permission.Read = identityPermissions.Permissions[permissionNameRead]
		permission.WorkItemsRead = identityPermissions.Permissions[permissionNameWorkItemsRead]
		permission.WorkItemsWrite = identityPermissions.Permissions[permissionNameWorkItemsWrite]
		permission.Write = identityPermissions.Permissions[permissionNameWrite]
	}
}
