package git

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
	permissionNameAdminister              = "Administer"
	permissionNameCreateBranch            = "CreateBranch"
	permissionNameCreateRepository        = "CreateRepository"
	permissionNameCreateTag               = "CreateTag"
	permissionNameContribute              = "GenericContribute"
	permissionNameDeleteRepository        = "DeleteRepository"
	permissionNameEditPolicies            = "EditPolicies"
	permissionNameForcePush               = "ForcePush"
	permissionNameManageNote              = "ManageNote"
	permissionNameManagePermissions       = "ManagePermissions"
	permissionNamePolicyExempt            = "PolicyExempt"
	permissionNamePullRequestBypassPolicy = "PullRequestBypassPolicy"
	permissionNamePullRequestContribute   = "PullRequestContribute"
	permissionNameRead                    = "GenericRead"
	permissionNameRemoveOthersLocks       = "RemoveOthersLocks"
	permissionNameRenameRepository        = "RenameRepository"
)

var _ resource.Resource = &GitPermissionsResource{}

func NewGitPermissionsResource() resource.Resource {
	return &GitPermissionsResource{}
}

type GitPermissionsResource struct {
	graphClient    *graph.Client
	securityClient *clientSecurity.Client
}

type GitPermissionsResourceModel struct {
	Id                  types.String   `tfsdk:"id"`
	Permissions         GitPermissions `tfsdk:"permissions"`
	PrincipalDescriptor types.String   `tfsdk:"principal_descriptor"`
	PrincipalName       string         `tfsdk:"principal_name"`
	ProjectId           string         `tfsdk:"project_id"`
}

type GitPermissions struct {
	Administer              string `tfsdk:"administer"`
	CreateBranch            string `tfsdk:"create_branch"`
	CreateRepository        string `tfsdk:"create_repository"`
	CreateTag               string `tfsdk:"create_tag"`
	Contribute              string `tfsdk:"contribute"`
	DeleteRepository        string `tfsdk:"delete_repository"`
	EditPolicies            string `tfsdk:"edit_policies"`
	ForcePush               string `tfsdk:"force_push"`
	ManageNote              string `tfsdk:"manage_note"`
	ManagePermissions       string `tfsdk:"manage_permissions"`
	PolicyExempt            string `tfsdk:"policy_exempt"`
	PullRequestBypassPolicy string `tfsdk:"pullrequest_bypass_policy"`
	PullRequestContribute   string `tfsdk:"pullrequest_contribute"`
	Read                    string `tfsdk:"read"`
	RemoveOthersLocks       string `tfsdk:"remove_others_locks"`
	RenameRepository        string `tfsdk:"rename_repository"`
}

func (r *GitPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_permissions"
}

func (r *GitPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets permissions on repositories within an Azure DevOps project. All permissions that currently exists will be overwritten.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the repository. If you omit the value, the permissions are applied to the repositories page and by default all repositories inherit permissions from there.",
				Optional:            true,
				Validators: []validator.String{
					validators.UUID(),
				},
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
					"create_branch": schema.StringAttribute{
						MarkdownDescription: "Sets the `CreateBranch` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"create_repository": schema.StringAttribute{
						MarkdownDescription: "Sets the `CreateRepository` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"create_tag": schema.StringAttribute{
						MarkdownDescription: "Sets the `CreateTag` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"contribute": schema.StringAttribute{
						MarkdownDescription: "Sets the `GenericContribute` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"delete_repository": schema.StringAttribute{
						MarkdownDescription: "Sets the `DeleteRepository` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"edit_policies": schema.StringAttribute{
						MarkdownDescription: "Sets the `EditPolicies` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"force_push": schema.StringAttribute{
						MarkdownDescription: "Sets the `ForcePush` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"manage_note": schema.StringAttribute{
						MarkdownDescription: "Sets the `ManageNote` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"manage_permissions": schema.StringAttribute{
						MarkdownDescription: "Sets the `ManagePermissions` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"policy_exempt": schema.StringAttribute{
						MarkdownDescription: "Sets the `PolicyExempt` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"pullrequest_bypass_policy": schema.StringAttribute{
						MarkdownDescription: "Sets the `PullRequestBypassPolicy` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"pullrequest_contribute": schema.StringAttribute{
						MarkdownDescription: "Sets the `PullRequestContribute` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"read": schema.StringAttribute{
						MarkdownDescription: "Sets the `GenericRead` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"remove_others_locks": schema.StringAttribute{
						MarkdownDescription: "Sets the `RemoveOthersLocks` permission for the identity. Must be `notset`, `allow` or `deny`.",
						Required:            true,
						Validators: []validator.String{
							validators.AllowDenyNotset(),
						},
					},
					"rename_repository": schema.StringAttribute{
						MarkdownDescription: "Sets the `RenameRepository` permission for the identity. Must be `notset`, `allow` or `deny`.",
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

func (r *GitPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.graphClient = req.ProviderData.(*clients.AzureDevOpsClient).GraphClient
	r.securityClient = req.ProviderData.(*clients.AzureDevOpsClient).SecurityClient
}

func (r *GitPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *GitPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetRepositoryToken(model.ProjectId, model.Id.ValueString())
	permissions := r.getPermissions(model)
	err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdGitRepositories, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.setPermissions(model, []*security.PrincipalPermissions{permissions})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *GitPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *GitPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetRepositoryToken(model.ProjectId, model.Id.ValueString())
	permissions, err := security.ReadPrincipalPermissions(ctx, clientSecurity.NamespaceIdGitRepositories, token, r.securityClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve access control lists", err.Error())
		return
	}

	r.setPermissions(model, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *GitPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *GitPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetRepositoryToken(model.ProjectId, model.Id.ValueString())
	permissions := r.getPermissions(model)
	err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdGitRepositories, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update permissions", err.Error())
		return
	}

	r.setPermissions(model, []*security.PrincipalPermissions{permissions})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *GitPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *GitPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetRepositoryToken(model.ProjectId, model.Id.ValueString())
	err := r.securityClient.RemoveAccessControlEntries(ctx, clientSecurity.NamespaceIdGitRepositories, token, []string{model.PrincipalDescriptor.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete permissions", err.Error())
		return
	}
}

// Private Methods

func (r *GitPermissionsResource) getPermissions(model *GitPermissionsResourceModel) *security.PrincipalPermissions {
	return &security.PrincipalPermissions{
		PrincipalDescriptor: model.PrincipalDescriptor.ValueString(),
		PrincipalName:       model.PrincipalName,
		Permissions: map[string]string{
			permissionNameAdminister:              model.Permissions.Administer,
			permissionNameCreateBranch:            model.Permissions.CreateBranch,
			permissionNameCreateRepository:        model.Permissions.CreateRepository,
			permissionNameCreateTag:               model.Permissions.CreateTag,
			permissionNameContribute:              model.Permissions.Contribute,
			permissionNameDeleteRepository:        model.Permissions.DeleteRepository,
			permissionNameEditPolicies:            model.Permissions.EditPolicies,
			permissionNameForcePush:               model.Permissions.ForcePush,
			permissionNameManageNote:              model.Permissions.ManageNote,
			permissionNameManagePermissions:       model.Permissions.ManagePermissions,
			permissionNamePolicyExempt:            model.Permissions.PolicyExempt,
			permissionNamePullRequestBypassPolicy: model.Permissions.PullRequestBypassPolicy,
			permissionNamePullRequestContribute:   model.Permissions.PullRequestContribute,
			permissionNameRead:                    model.Permissions.Read,
			permissionNameRemoveOthersLocks:       model.Permissions.RemoveOthersLocks,
			permissionNameRenameRepository:        model.Permissions.RenameRepository,
		},
	}
}

func (r *GitPermissionsResource) setPermissions(model *GitPermissionsResourceModel, p []*security.PrincipalPermissions) {
	if len(p) == 0 {
		return
	}

	principalPermissions := linq.From(p).FirstWith(func(p interface{}) bool {
		return p.(*security.PrincipalPermissions).PrincipalName == model.PrincipalName
	}).(*security.PrincipalPermissions)
	model.Permissions.Administer = principalPermissions.Permissions[permissionNameAdminister]
	model.Permissions.CreateBranch = principalPermissions.Permissions[permissionNameCreateBranch]
	model.Permissions.CreateRepository = principalPermissions.Permissions[permissionNameCreateRepository]
	model.Permissions.CreateTag = principalPermissions.Permissions[permissionNameCreateTag]
	model.Permissions.Contribute = principalPermissions.Permissions[permissionNameContribute]
	model.Permissions.DeleteRepository = principalPermissions.Permissions[permissionNameDeleteRepository]
	model.Permissions.EditPolicies = principalPermissions.Permissions[permissionNameEditPolicies]
	model.Permissions.ForcePush = principalPermissions.Permissions[permissionNameForcePush]
	model.Permissions.ManageNote = principalPermissions.Permissions[permissionNameManageNote]
	model.Permissions.ManagePermissions = principalPermissions.Permissions[permissionNameManagePermissions]
	model.Permissions.PolicyExempt = principalPermissions.Permissions[permissionNamePolicyExempt]
	model.Permissions.PullRequestBypassPolicy = principalPermissions.Permissions[permissionNamePullRequestBypassPolicy]
	model.Permissions.PullRequestContribute = principalPermissions.Permissions[permissionNamePullRequestContribute]
	model.Permissions.Read = principalPermissions.Permissions[permissionNameRead]
	model.Permissions.RemoveOthersLocks = principalPermissions.Permissions[permissionNameRemoveOthersLocks]
	model.Permissions.RenameRepository = principalPermissions.Permissions[permissionNameRenameRepository]
	model.PrincipalDescriptor = types.StringValue(principalPermissions.PrincipalDescriptor)
	model.PrincipalName = principalPermissions.PrincipalName
}
