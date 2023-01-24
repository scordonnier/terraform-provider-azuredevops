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
	Path                string          `tfsdk:"path"`
	Permissions         AreaPermissions `tfsdk:"permissions"`
	PrincipalDescriptor types.String    `tfsdk:"principal_descriptor"`
	PrincipalName       string          `tfsdk:"principal_name"`
	ProjectId           string          `tfsdk:"project_id"`
}

type AreaPermissions struct {
	Create           string `tfsdk:"create"`
	Delete           string `tfsdk:"delete"`
	ManageTestPlans  string `tfsdk:"manage_test_plans"`
	ManageTestSuites string `tfsdk:"manage_test_suites"`
	Read             string `tfsdk:"read"`
	WorkItemsRead    string `tfsdk:"workitems_read"`
	WorkItemsWrite   string `tfsdk:"workitems_write"`
	Write            string `tfsdk:"write"`
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
			"permissions": schema.SingleNestedAttribute{
				MarkdownDescription: "The permissions to assign.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"create": schema.StringAttribute{
						MarkdownDescription: "Sets the `CREATE_CHILDREN` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"delete": schema.StringAttribute{
						MarkdownDescription: "Sets the `DELETE` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"manage_test_plans": schema.StringAttribute{
						MarkdownDescription: "Sets the `MANAGE_TEST_PLANS` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"manage_test_suites": schema.StringAttribute{
						MarkdownDescription: "Sets the `MANAGE_TEST_SUITES` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"read": schema.StringAttribute{
						MarkdownDescription: "Sets the `GENERIC_READ` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"workitems_read": schema.StringAttribute{
						MarkdownDescription: "Sets the `WORK_ITEM_READ` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"workitems_write": schema.StringAttribute{
						MarkdownDescription: "Sets the `WORK_ITEM_WRITE` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"write": schema.StringAttribute{
						MarkdownDescription: "Sets the `GENERIC_WRITE` permission for the identity. Must be `notset`, `allow` or `deny`.",
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

	permissions := r.getPermissions(model)
	err = security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdCSS, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.setPermissions(model, []*security.PrincipalPermissions{permissions})

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

	permissions, err := security.ReadPrincipalPermissions(ctx, clientSecurity.NamespaceIdCSS, token, r.securityClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve access control lists", err.Error())
		return
	}

	r.setPermissions(model, permissions)

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

	permissions := r.getPermissions(model)
	err = security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdCSS, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.setPermissions(model, []*security.PrincipalPermissions{permissions})

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

	err = r.securityClient.RemoveAccessControlEntries(ctx, clientSecurity.NamespaceIdCSS, token, []string{model.PrincipalDescriptor.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete permissions", err.Error())
		return
	}
}

// Private Methods

func (r *AreaPermissionsResource) getPermissions(model *AreaPermissionsResourceModel) *security.PrincipalPermissions {
	return &security.PrincipalPermissions{
		PrincipalDescriptor: model.PrincipalDescriptor.ValueString(),
		PrincipalName:       model.PrincipalName,
		Permissions: map[string]string{
			permissionNameCreate:           model.Permissions.Create,
			permissionNameDelete:           model.Permissions.Delete,
			permissionNameManageTestPlans:  model.Permissions.ManageTestPlans,
			permissionNameManageTestSuites: model.Permissions.ManageTestSuites,
			permissionNameRead:             model.Permissions.Read,
			permissionNameWorkItemsRead:    model.Permissions.WorkItemsRead,
			permissionNameWorkItemsWrite:   model.Permissions.WorkItemsWrite,
			permissionNameWrite:            model.Permissions.Write,
		},
	}
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

func (r *AreaPermissionsResource) setPermissions(model *AreaPermissionsResourceModel, p []*security.PrincipalPermissions) {
	if len(p) == 0 {
		return
	}

	principalPermissions := linq.From(p).FirstWith(func(p interface{}) bool {
		return p.(*security.PrincipalPermissions).PrincipalName == model.PrincipalName
	}).(*security.PrincipalPermissions)
	model.Permissions.Create = principalPermissions.Permissions[permissionNameCreate]
	model.Permissions.Delete = principalPermissions.Permissions[permissionNameDelete]
	model.Permissions.ManageTestPlans = principalPermissions.Permissions[permissionNameManageTestPlans]
	model.Permissions.ManageTestSuites = principalPermissions.Permissions[permissionNameManageTestSuites]
	model.Permissions.Read = principalPermissions.Permissions[permissionNameRead]
	model.Permissions.WorkItemsRead = principalPermissions.Permissions[permissionNameWorkItemsRead]
	model.Permissions.WorkItemsWrite = principalPermissions.Permissions[permissionNameWorkItemsWrite]
	model.Permissions.Write = principalPermissions.Permissions[permissionNameWrite]
	model.PrincipalDescriptor = types.StringValue(principalPermissions.PrincipalDescriptor)
	model.PrincipalName = principalPermissions.PrincipalName
}
