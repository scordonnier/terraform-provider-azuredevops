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
	Id                  types.Int64         `tfsdk:"id"`
	Permissions         PipelinePermissions `tfsdk:"permissions"`
	PrincipalDescriptor types.String        `tfsdk:"principal_descriptor"`
	PrincipalName       string              `tfsdk:"principal_name"`
	ProjectId           string              `tfsdk:"project_id"`
}

type PipelinePermissions struct {
	AdministerBuildPermissions     string `tfsdk:"administer_build_permissions"`
	DeleteBuildDefinition          string `tfsdk:"delete_build_definition"`
	DeleteBuilds                   string `tfsdk:"delete_builds"`
	DestroyBuilds                  string `tfsdk:"destroy_builds"`
	EditBuildDefinition            string `tfsdk:"edit_build_definition"`
	EditBuildQuality               string `tfsdk:"edit_build_quality"`
	ManageBuildQualities           string `tfsdk:"manage_build_qualities"`
	ManageBuildQueue               string `tfsdk:"manage_build_queue"`
	OverrideBuildCheckInValidation string `tfsdk:"override_build_checkin_validation"`
	QueueBuilds                    string `tfsdk:"queue_builds"`
	RetainIndefinitely             string `tfsdk:"retain_indefinitely"`
	StopBuilds                     string `tfsdk:"stop_builds"`
	UpdateBuildInformation         string `tfsdk:"update_build_information"`
	ViewBuildDefinition            string `tfsdk:"view_build_definition"`
	ViewBuilds                     string `tfsdk:"view_builds"`
}

func (r *PipelinePermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_permissions"
}

func (r *PipelinePermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets permissions on pipelines within an Azure DevOps project. All permissions that currently exists will be overwritten.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the pipeline. If you omit the value, the permissions are applied to the pipelines page and by default all pipelines inherit permissions from there.",
				Optional:            true,
			},
			"permissions": schema.SingleNestedAttribute{
				MarkdownDescription: "The permissions to assign.",
				Required:            true,
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
	permissions := r.getPermissions(model)
	err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdBuild, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.setPermissions(model, []*security.PrincipalPermissions{permissions})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PipelinePermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *PipelinePermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetPipelineToken(model.ProjectId, int(model.Id.ValueInt64()))
	permissions, err := security.ReadPrincipalPermissions(ctx, clientSecurity.NamespaceIdBuild, token, r.securityClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve access control lists", err.Error())
		return
	}

	r.setPermissions(model, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PipelinePermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *PipelinePermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetPipelineToken(model.ProjectId, int(model.Id.ValueInt64()))
	permissions := r.getPermissions(model)
	err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdBuild, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update permissions", err.Error())
		return
	}

	r.setPermissions(model, []*security.PrincipalPermissions{permissions})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PipelinePermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *PipelinePermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetPipelineToken(model.ProjectId, int(model.Id.ValueInt64()))
	err := r.securityClient.RemoveAccessControlEntries(ctx, clientSecurity.NamespaceIdBuild, token, []string{model.PrincipalDescriptor.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete permissions", err.Error())
		return
	}
}

// Private Methods

func (r *PipelinePermissionsResource) getPermissions(model *PipelinePermissionsResourceModel) *security.PrincipalPermissions {
	return &security.PrincipalPermissions{
		PrincipalDescriptor: model.PrincipalDescriptor.ValueString(),
		PrincipalName:       model.PrincipalName,
		Permissions: map[string]string{
			permissionNameAdministerBuildPermissions:     model.Permissions.AdministerBuildPermissions,
			permissionNameDeleteBuildDefinition:          model.Permissions.DeleteBuildDefinition,
			permissionNameDeleteBuilds:                   model.Permissions.DeleteBuilds,
			permissionNameDestroyBuilds:                  model.Permissions.DestroyBuilds,
			permissionNameEditBuildDefinition:            model.Permissions.EditBuildDefinition,
			permissionNameEditBuildQuality:               model.Permissions.EditBuildQuality,
			permissionNameManageBuildQualities:           model.Permissions.ManageBuildQualities,
			permissionNameManageBuildQueue:               model.Permissions.ManageBuildQueue,
			permissionNameOverrideBuildCheckInValidation: model.Permissions.OverrideBuildCheckInValidation,
			permissionNameQueueBuilds:                    model.Permissions.QueueBuilds,
			permissionNameRetainIndefinitely:             model.Permissions.RetainIndefinitely,
			permissionNameStopBuilds:                     model.Permissions.StopBuilds,
			permissionNameUpdateBuildInformation:         model.Permissions.UpdateBuildInformation,
			permissionNameViewBuildDefinition:            model.Permissions.ViewBuildDefinition,
			permissionNameViewBuilds:                     model.Permissions.ViewBuilds,
		},
	}
}

func (r *PipelinePermissionsResource) setPermissions(model *PipelinePermissionsResourceModel, p []*security.PrincipalPermissions) {
	if len(p) == 0 {
		return
	}

	principalPermissions := linq.From(p).FirstWith(func(p interface{}) bool {
		return p.(*security.PrincipalPermissions).PrincipalName == model.PrincipalName
	}).(*security.PrincipalPermissions)
	model.Permissions.AdministerBuildPermissions = principalPermissions.Permissions[permissionNameAdministerBuildPermissions]
	model.Permissions.DeleteBuildDefinition = principalPermissions.Permissions[permissionNameDeleteBuildDefinition]
	model.Permissions.DeleteBuilds = principalPermissions.Permissions[permissionNameDeleteBuilds]
	model.Permissions.DestroyBuilds = principalPermissions.Permissions[permissionNameDestroyBuilds]
	model.Permissions.EditBuildDefinition = principalPermissions.Permissions[permissionNameEditBuildDefinition]
	model.Permissions.EditBuildQuality = principalPermissions.Permissions[permissionNameEditBuildQuality]
	model.Permissions.ManageBuildQualities = principalPermissions.Permissions[permissionNameManageBuildQualities]
	model.Permissions.ManageBuildQueue = principalPermissions.Permissions[permissionNameManageBuildQueue]
	model.Permissions.OverrideBuildCheckInValidation = principalPermissions.Permissions[permissionNameOverrideBuildCheckInValidation]
	model.Permissions.QueueBuilds = principalPermissions.Permissions[permissionNameQueueBuilds]
	model.Permissions.RetainIndefinitely = principalPermissions.Permissions[permissionNameRetainIndefinitely]
	model.Permissions.StopBuilds = principalPermissions.Permissions[permissionNameStopBuilds]
	model.Permissions.UpdateBuildInformation = principalPermissions.Permissions[permissionNameUpdateBuildInformation]
	model.Permissions.ViewBuildDefinition = principalPermissions.Permissions[permissionNameViewBuildDefinition]
	model.Permissions.ViewBuilds = principalPermissions.Permissions[permissionNameViewBuilds]
	model.PrincipalDescriptor = types.StringValue(principalPermissions.PrincipalDescriptor)
	model.PrincipalName = principalPermissions.PrincipalName
}
