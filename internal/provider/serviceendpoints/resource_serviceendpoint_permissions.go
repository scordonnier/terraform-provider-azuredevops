package serviceendpoints

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
	permissionNameAdminister        = "Administer"
	permissionNameCreate            = "Create"
	permissionNameUse               = "Use"
	permissionNameViewAuthorization = "ViewAuthorization"
	permissionNameViewEndpoint      = "ViewEndpoint"
)

var _ resource.Resource = &ServiceEndpointPermissionsResource{}

func NewServiceEndpointPermissionsResource() resource.Resource {
	return &ServiceEndpointPermissionsResource{}
}

type ServiceEndpointPermissionsResource struct {
	graphClient    *graph.Client
	securityClient *clientSecurity.Client
}

type ServiceEndpointPermissionsResourceModel struct {
	Id          types.String                 `tfsdk:"id"`
	ProjectId   string                       `tfsdk:"project_id"`
	Permissions []ServiceEndpointPermissions `tfsdk:"permissions"`
}

type ServiceEndpointPermissions struct {
	Administer         string       `tfsdk:"administer"`
	Create             string       `tfsdk:"create"`
	IdentityDescriptor types.String `tfsdk:"identity_descriptor"`
	IdentityName       string       `tfsdk:"identity_name"`
	Use                string       `tfsdk:"use"`
	ViewAuthorization  string       `tfsdk:"view_authorization"`
	ViewEndpoint       string       `tfsdk:"view_endpoint"`
}

func (r *ServiceEndpointPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_permissions"
}

func (r *ServiceEndpointPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets permissions on service endpoints of an existing project within Azure DevOps. All permissions that currently exists will be overwritten.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the service endpoint. If you omit the value, the permissions are applied to the service connections page and by default all service connections inherit permissions from there.",
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
						"use": schema.StringAttribute{
							MarkdownDescription: "Sets the `Use` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"view_authorization": schema.StringAttribute{
							MarkdownDescription: "Sets the `ViewAuthorization` permission for the identity. Must be `notset`, `allow` or `deny`.",
							Required:            true,
							Validators: []validator.String{
								validators.PermissionsValidator(),
							},
						},
						"view_endpoint": schema.StringAttribute{
							MarkdownDescription: "Sets the `ViewEndpoint` permission for the identity. Must be `notset`, `allow` or `deny`.",
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

func (r *ServiceEndpointPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.graphClient = req.ProviderData.(*clients.AzureDevOpsClient).GraphClient
	r.securityClient = req.ProviderData.(*clients.AzureDevOpsClient).SecurityClient
}

func (r *ServiceEndpointPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ServiceEndpointPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetServiceEndpointToken(model.ProjectId, model.Id.ValueString())
	permissions := r.getPermissions(&model.Permissions)
	err := security.CreateOrUpdateIdentityPermissions(ctx, clientSecurity.NamespaceIdServiceEndpoints, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ServiceEndpointPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetServiceEndpointToken(model.ProjectId, model.Id.ValueString())
	permissions, err := security.ReadIdentityPermissions(ctx, clientSecurity.NamespaceIdServiceEndpoints, token, r.securityClient)
	if err != nil {
		if permissions != nil && len(permissions) == 0 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Unable to retrieve access control lists", err.Error())
		return
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ServiceEndpointPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetServiceEndpointToken(model.ProjectId, model.Id.ValueString())
	permissions := r.getPermissions(&model.Permissions)
	err := security.CreateOrUpdateIdentityPermissions(ctx, clientSecurity.NamespaceIdServiceEndpoints, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update permissions", err.Error())
		return
	}

	r.updatePermissions(&model.Permissions, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ServiceEndpointPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetServiceEndpointToken(model.ProjectId, model.Id.ValueString())
	err := r.securityClient.RemoveAccessControlLists(ctx, clientSecurity.NamespaceIdEnvironment, token)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete environment permissions", err.Error())
	}
}

// Private Methods

func (r *ServiceEndpointPermissionsResource) getPermissions(p *[]ServiceEndpointPermissions) []*security.IdentityPermissions {
	var permissions []*security.IdentityPermissions
	for _, permission := range *p {
		permissions = append(permissions, &security.IdentityPermissions{
			IdentityDescriptor: permission.IdentityDescriptor.ValueString(),
			IdentityName:       permission.IdentityName,
			Permissions: map[string]string{
				permissionNameAdminister:        permission.Administer,
				permissionNameCreate:            permission.Create,
				permissionNameUse:               permission.Use,
				permissionNameViewAuthorization: permission.ViewAuthorization,
				permissionNameViewEndpoint:      permission.ViewEndpoint,
			},
		})
	}
	return permissions
}

func (r *ServiceEndpointPermissionsResource) updatePermissions(p1 *[]ServiceEndpointPermissions, p2 []*security.IdentityPermissions) {
	for index := range *p1 {
		permission := &(*p1)[index]
		identityPermissions := linq.From(p2).FirstWith(func(p interface{}) bool {
			return p.(*security.IdentityPermissions).IdentityName == permission.IdentityName
		}).(*security.IdentityPermissions)
		permission.IdentityDescriptor = types.StringValue(identityPermissions.IdentityDescriptor)
		permission.IdentityName = identityPermissions.IdentityName
		permission.Administer = identityPermissions.Permissions[permissionNameAdminister]
		permission.Create = identityPermissions.Permissions[permissionNameCreate]
		permission.Use = identityPermissions.Permissions[permissionNameUse]
		permission.ViewAuthorization = identityPermissions.Permissions[permissionNameViewAuthorization]
		permission.ViewEndpoint = identityPermissions.Permissions[permissionNameViewEndpoint]
	}
}