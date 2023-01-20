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
	permissionNameAdministerBuildPermissions     = "AdministerBuildPermissions"
	permissionNameDeleteBuildDefinition          = "DeleteBuildDefinition"
	permissionNameDeleteBuilds                   = "DeleteBuilds"
	permissionNameDestroyBuilds                  = "DestroyBuilds"
	permissionNameEditBuildDefinition            = "EditBuildDefinition"
	permissionNameEditBuildQuality               = "EditBuildQuality"
	permissionNameManageBuildQualities           = "ManageBuildQualities"
	permissionNameManageBuildQueue               = "ManageBuildQueue"
	permissionNameOverrideBuildCheckInValidation = "OverrideBuildCheckInValidation"
	permissionNameQueueBuilds                    = "QueueBuilds"
	permissionNameRetainIndefinitely             = "RetainIndefinitely"
	permissionNameStopBuilds                     = "StopBuilds"
	permissionNameUpdateBuildInformation         = "UpdateBuildInformation"
	permissionNameViewBuildDefinition            = "ViewBuildDefinition"
	permissionNameViewBuilds                     = "ViewBuilds"
)

var _ resource.Resource = &PipelinePermissionsResource{}

func NewPipelinePermissionsResource() resource.Resource {
	return &PipelinePermissionsResource{}
}

type PipelinePermissionsResource struct {
	graphClient    *graph.Client
	securityClient *clientSecurity.Client
}

type PipelinePermissionsResourceModel struct {
	Id          types.Int64           `tfsdk:"id"`
	Permissions []PipelinePermissions `tfsdk:"permissions"`
	ProjectId   string                `tfsdk:"project_id"`
}

type PipelinePermissions struct {
	AdministerBuildPermissions     string       `tfsdk:"administer_build_permissions"`
	DeleteBuildDefinition          string       `tfsdk:"delete_build_definition"`
	DeleteBuilds                   string       `tfsdk:"delete_builds"`
	DestroyBuilds                  string       `tfsdk:"destroy_builds"`
	EditBuildDefinition            string       `tfsdk:"edit_build_definition"`
	EditBuildQuality               string       `tfsdk:"edit_build_quality"`
	IdentityDescriptor             types.String `tfsdk:"identity_descriptor"`
	IdentityName                   string       `tfsdk:"identity_name"`
	ManageBuildQualities           string       `tfsdk:"manage_build_qualities"`
	ManageBuildQueue               string       `tfsdk:"manage_build_queue"`
	OverrideBuildCheckInValidation string       `tfsdk:"override_build_checkin_validation"`
	QueueBuilds                    string       `tfsdk:"queue_builds"`
	RetainIndefinitely             string       `tfsdk:"retain_indefinitely"`
	StopBuilds                     string       `tfsdk:"stop_builds"`
	UpdateBuildInformation         string       `tfsdk:"update_build_information"`
	ViewBuildDefinition            string       `tfsdk:"view_build_definition"`
	ViewBuilds                     string       `tfsdk:"view_builds"`
}

func (r *PipelinePermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_permissions"
}

func (r *PipelinePermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets permissions on pipelines within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the pipeline. If you omit the value, the permissions are applied to the pipelines page and by default all pipelines inherit permissions from there.",
				Optional:            true,
			},
			"permissions": schema.ListNestedAttribute{
				MarkdownDescription: "The permissions to assign.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"administer_build_permissions": schema.StringAttribute{
							MarkdownDescription: "Sets the `AdministerBuildPermissions` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"delete_build_definition": schema.StringAttribute{
							MarkdownDescription: "Sets the `DeleteBuildDefinition` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"delete_builds": schema.StringAttribute{
							MarkdownDescription: "Sets the `DeleteBuilds` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"destroy_builds": schema.StringAttribute{
							MarkdownDescription: "Sets the `DestroyBuilds` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"edit_build_definition": schema.StringAttribute{
							MarkdownDescription: "Sets the `EditBuildDefinition` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"edit_build_quality": schema.StringAttribute{
							MarkdownDescription: "Sets the `EditBuildQuality` permission for the identity. Must be `notset`, `allow` or `deny`.",
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
						"manage_build_qualities": schema.StringAttribute{
							MarkdownDescription: "Sets the `ManageBuildQualities` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"manage_build_queue": schema.StringAttribute{
							MarkdownDescription: "Sets the `ManageBuildQueue` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"override_build_checkin_validation": schema.StringAttribute{
							MarkdownDescription: "Sets the `OverrideBuildCheckInValidation` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"queue_builds": schema.StringAttribute{
							MarkdownDescription: "Sets the `QueueBuilds` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"retain_indefinitely": schema.StringAttribute{
							MarkdownDescription: "Sets the `RetainIndefinitely` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"stop_builds": schema.StringAttribute{
							MarkdownDescription: "Sets the `StopBuilds` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"update_build_information": schema.StringAttribute{
							MarkdownDescription: "Sets the `UpdateBuildInformation` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"view_build_definition": schema.StringAttribute{
							MarkdownDescription: "Sets the `ViewBuildDefinition` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"view_builds": schema.StringAttribute{
							MarkdownDescription: "Sets the `ViewBuilds` permission for the identity. Must be `notset`, `allow` or `deny`.",
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

func (r *PipelinePermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.graphClient = req.ProviderData.(*clients.AzureDevOpsClient).GraphClient
	r.securityClient = req.ProviderData.(*clients.AzureDevOpsClient).SecurityClient
}

func (r *PipelinePermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *PipelinePermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetPipelineToken(model.ProjectId, int(model.Id.ValueInt64()))
	permissions := r.getPermissions(&model.Permissions)
	for _, permission := range permissions {
		err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdBuild, token, permission, r.securityClient, r.graphClient)
		if err != nil {
			resp.Diagnostics.AddError("Unable to create permissions", err.Error())
			return
		}
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PipelinePermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *PipelinePermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetPipelineToken(model.ProjectId, int(model.Id.ValueInt64()))
	permissions, err := security.ReadIdentityPermissions(ctx, clientSecurity.NamespaceIdBuild, token, r.securityClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve access control lists", err.Error())
		return
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PipelinePermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *PipelinePermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetPipelineToken(model.ProjectId, int(model.Id.ValueInt64()))
	permissions := r.getPermissions(&model.Permissions)
	for _, permission := range permissions {
		err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdBuild, token, permission, r.securityClient, r.graphClient)
		if err != nil {
			resp.Diagnostics.AddError("Unable to update permissions", err.Error())
			return
		}
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PipelinePermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *PipelinePermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetPipelineToken(model.ProjectId, int(model.Id.ValueInt64()))
	var descriptors []string
	for _, permission := range model.Permissions {
		descriptors = append(descriptors, permission.IdentityDescriptor.ValueString())
	}
	err := r.securityClient.RemoveAccessControlEntries(ctx, clientSecurity.NamespaceIdBuild, token, descriptors)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete permissions", err.Error())
		return
	}
}

// Private Methods

func (r *PipelinePermissionsResource) getPermissions(p *[]PipelinePermissions) []*security.IdentityPermissions {
	var permissions []*security.IdentityPermissions
	for _, permission := range *p {
		permissions = append(permissions, &security.IdentityPermissions{
			IdentityDescriptor: permission.IdentityDescriptor.ValueString(),
			IdentityName:       permission.IdentityName,
			Permissions: map[string]string{
				permissionNameAdministerBuildPermissions:     permission.AdministerBuildPermissions,
				permissionNameDeleteBuildDefinition:          permission.DeleteBuildDefinition,
				permissionNameDeleteBuilds:                   permission.DeleteBuilds,
				permissionNameDestroyBuilds:                  permission.DestroyBuilds,
				permissionNameEditBuildDefinition:            permission.EditBuildDefinition,
				permissionNameEditBuildQuality:               permission.EditBuildQuality,
				permissionNameManageBuildQualities:           permission.ManageBuildQualities,
				permissionNameManageBuildQueue:               permission.ManageBuildQueue,
				permissionNameOverrideBuildCheckInValidation: permission.OverrideBuildCheckInValidation,
				permissionNameQueueBuilds:                    permission.QueueBuilds,
				permissionNameRetainIndefinitely:             permission.RetainIndefinitely,
				permissionNameStopBuilds:                     permission.StopBuilds,
				permissionNameUpdateBuildInformation:         permission.UpdateBuildInformation,
				permissionNameViewBuildDefinition:            permission.ViewBuildDefinition,
				permissionNameViewBuilds:                     permission.ViewBuilds,
			},
		})
	}
	return permissions
}

func (r *PipelinePermissionsResource) updatePermissions(p1 *[]PipelinePermissions, p2 []*security.IdentityPermissions) {
	if len(p2) == 0 {
		return
	}

	for index := range *p1 {
		permission := &(*p1)[index]
		identityPermissions := linq.From(p2).FirstWith(func(p interface{}) bool {
			return p.(*security.IdentityPermissions).IdentityName == permission.IdentityName
		}).(*security.IdentityPermissions)
		permission.AdministerBuildPermissions = identityPermissions.Permissions[permissionNameAdministerBuildPermissions]
		permission.DeleteBuildDefinition = identityPermissions.Permissions[permissionNameDeleteBuildDefinition]
		permission.DeleteBuilds = identityPermissions.Permissions[permissionNameDeleteBuilds]
		permission.DestroyBuilds = identityPermissions.Permissions[permissionNameDestroyBuilds]
		permission.EditBuildDefinition = identityPermissions.Permissions[permissionNameEditBuildDefinition]
		permission.EditBuildQuality = identityPermissions.Permissions[permissionNameEditBuildQuality]
		permission.IdentityDescriptor = types.StringValue(identityPermissions.IdentityDescriptor)
		permission.IdentityName = identityPermissions.IdentityName
		permission.ManageBuildQualities = identityPermissions.Permissions[permissionNameManageBuildQualities]
		permission.ManageBuildQueue = identityPermissions.Permissions[permissionNameManageBuildQueue]
		permission.OverrideBuildCheckInValidation = identityPermissions.Permissions[permissionNameOverrideBuildCheckInValidation]
		permission.QueueBuilds = identityPermissions.Permissions[permissionNameQueueBuilds]
		permission.RetainIndefinitely = identityPermissions.Permissions[permissionNameRetainIndefinitely]
		permission.StopBuilds = identityPermissions.Permissions[permissionNameStopBuilds]
		permission.UpdateBuildInformation = identityPermissions.Permissions[permissionNameUpdateBuildInformation]
		permission.ViewBuildDefinition = identityPermissions.Permissions[permissionNameViewBuildDefinition]
		permission.ViewBuilds = identityPermissions.Permissions[permissionNameViewBuilds]
	}
}
