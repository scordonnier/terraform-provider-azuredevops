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
	Id                  types.String               `tfsdk:"id"`
	Permissions         ServiceEndpointPermissions `tfsdk:"permissions"`
	PrincipalDescriptor types.String               `tfsdk:"principal_descriptor"`
	PrincipalName       string                     `tfsdk:"principal_name"`
	ProjectId           string                     `tfsdk:"project_id"`
}

type ServiceEndpointPermissions struct {
	Administer        string `tfsdk:"administer"`
	Create            string `tfsdk:"create"`
	Use               string `tfsdk:"use"`
	ViewAuthorization string `tfsdk:"view_authorization"`
	ViewEndpoint      string `tfsdk:"view_endpoint"`
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
			"permissions": schema.SingleNestedAttribute{
				MarkdownDescription: "The permissions to assign.",
				Required:            true,
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
	permissions := r.getPermissions(model)
	err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdServiceEndpoints, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.setPermissions(model, []*security.PrincipalPermissions{permissions})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ServiceEndpointPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetServiceEndpointToken(model.ProjectId, model.Id.ValueString())
	permissions, err := security.ReadPrincipalPermissions(ctx, clientSecurity.NamespaceIdServiceEndpoints, token, r.securityClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve access control lists", err.Error())
		return
	}

	r.setPermissions(model, permissions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ServiceEndpointPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetServiceEndpointToken(model.ProjectId, model.Id.ValueString())
	permissions := r.getPermissions(model)
	err := security.CreateOrUpdateAccessControlEntry(ctx, clientSecurity.NamespaceIdServiceEndpoints, token, permissions, r.securityClient, r.graphClient)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create permissions", err.Error())
		return
	}

	r.setPermissions(model, []*security.PrincipalPermissions{permissions})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ServiceEndpointPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := r.securityClient.GetServiceEndpointToken(model.ProjectId, model.Id.ValueString())
	err := r.securityClient.RemoveAccessControlEntries(ctx, clientSecurity.NamespaceIdServiceEndpoints, token, []string{model.PrincipalDescriptor.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete permissions", err.Error())
		return
	}
}

// Private Methods

func (r *ServiceEndpointPermissionsResource) getPermissions(model *ServiceEndpointPermissionsResourceModel) *security.PrincipalPermissions {
	return &security.PrincipalPermissions{
		PrincipalDescriptor: model.PrincipalDescriptor.ValueString(),
		PrincipalName:       model.PrincipalName,
		Permissions: map[string]string{
			permissionNameAdminister:        model.Permissions.Administer,
			permissionNameCreate:            model.Permissions.Create,
			permissionNameUse:               model.Permissions.Use,
			permissionNameViewAuthorization: model.Permissions.ViewAuthorization,
			permissionNameViewEndpoint:      model.Permissions.ViewEndpoint,
		},
	}
}

func (r *ServiceEndpointPermissionsResource) setPermissions(model *ServiceEndpointPermissionsResourceModel, p []*security.PrincipalPermissions) {
	if len(p) == 0 {
		return
	}

	principalPermissions := linq.From(p).FirstWith(func(p interface{}) bool {
		return p.(*security.PrincipalPermissions).PrincipalName == model.PrincipalName
	}).(*security.PrincipalPermissions)
	model.Permissions.Administer = principalPermissions.Permissions[permissionNameAdminister]
	model.Permissions.Create = principalPermissions.Permissions[permissionNameCreate]
	model.Permissions.Use = principalPermissions.Permissions[permissionNameUse]
	model.Permissions.ViewAuthorization = principalPermissions.Permissions[permissionNameViewAuthorization]
	model.Permissions.ViewEndpoint = principalPermissions.Permissions[permissionNameViewEndpoint]
	model.PrincipalDescriptor = types.StringValue(principalPermissions.PrincipalDescriptor)
	model.PrincipalName = principalPermissions.PrincipalName
}
