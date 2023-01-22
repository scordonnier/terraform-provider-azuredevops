package serviceendpoints

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoints"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/validators"
)

var _ resource.Resource = &ServiceEndpointJFrogResource{}

func NewServiceEndpointJFrogResource() resource.Resource {
	return &ServiceEndpointJFrogResource{}
}

type ServiceEndpointJFrogResource struct {
	pipelinesClient        *pipelines.Client
	serviceEndpointsClient *serviceendpoints.Client
}

type ServiceEndpointJFrogResourceModel struct {
	AccessToken       types.String `tfsdk:"access_token"`
	Description       *string      `tfsdk:"description"`
	Id                types.String `tfsdk:"id"`
	GrantAllPipelines bool         `tfsdk:"grant_all_pipelines"`
	Name              string       `tfsdk:"name"`
	Password          types.String `tfsdk:"password"`
	ProjectId         string       `tfsdk:"project_id"`
	Service           string       `tfsdk:"service"`
	Username          types.String `tfsdk:"username"`
	URL               string       `tfsdk:"url"`
}

func (r *ServiceEndpointJFrogResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_jfrog"
}

func (r *ServiceEndpointJFrogResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceShema := GetServiceEndpointResourceSchemaBase("Manages a JFrog service endpoint within an Azure DevOps project. You need to install [JFrog Azure DevOps Extension](https://marketplace.visualstudio.com/items?itemName=JFrog.jfrog-azure-devops-extension) from the Marketplace.")
	resourceShema.Attributes["access_token"] = schema.StringAttribute{
		MarkdownDescription: "Access Token with deploy permissions.",
		Optional:            true,
		Sensitive:           true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRoot("password"), path.MatchRoot("username")),
			validators.StringNotEmptyValidator(),
		},
	}
	resourceShema.Attributes["password"] = schema.StringAttribute{
		MarkdownDescription: "Password or API key of an JFrog user with deploy permissions.",
		Optional:            true,
		Sensitive:           true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRoot("access_token")),
			validators.StringNotEmptyValidator(),
		},
	}
	resourceShema.Attributes["service"] = schema.StringAttribute{
		MarkdownDescription: "JFrog service type. Must be `artifactory`, `distribution`, `platform` or `xray`.",
		Required:            true,
		Validators: []validator.String{
			stringvalidator.OneOfCaseInsensitive("artifactory", "distribution", "platform", "xray"),
		},
	}
	resourceShema.Attributes["url"] = schema.StringAttribute{
		MarkdownDescription: "Specify the root URL of your JFrog platform (eg. https://my.jfrog.io/).",
		Required:            true,
		Validators: []validator.String{
			validators.StringNotEmptyValidator(),
		},
	}
	resourceShema.Attributes["username"] = schema.StringAttribute{
		MarkdownDescription: "JFrog username with deploy permissions.",
		Optional:            true,
		Sensitive:           true,
		Validators: []validator.String{
			stringvalidator.ConflictsWith(path.MatchRoot("access_token")),
			validators.StringNotEmptyValidator(),
		},
	}
	resp.Schema = resourceShema
}

func (r *ServiceEndpointJFrogResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.pipelinesClient = req.ProviderData.(*clients.AzureDevOpsClient).PipelinesClient
	r.serviceEndpointsClient = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointsClient
}

func (r *ServiceEndpointJFrogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ServiceEndpointJFrogResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, err := CreateResourceServiceEndpoint(ctx, model.ProjectId, r.getCreateOrUpdateServiceEndpointArgs(model), r.serviceEndpointsClient, r.pipelinesClient, resp)
	if err != nil {
		return
	}

	model.Id = types.StringValue(serviceEndpoint.Id.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointJFrogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ServiceEndpointJFrogResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, granted, err := ReadResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointsClient, r.pipelinesClient, resp)
	if err != nil {
		return
	}

	model.Description = utils.IfThenElse[*string](serviceEndpoint.Description != nil, model.Description, utils.EmptyString)
	model.GrantAllPipelines = granted
	model.Name = *serviceEndpoint.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointJFrogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ServiceEndpointJFrogResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := UpdateResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.getCreateOrUpdateServiceEndpointArgs(model), r.serviceEndpointsClient, r.pipelinesClient, resp)
	if err != nil {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointJFrogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ServiceEndpointJFrogResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointsClient, resp)
}

// Private Methods

func (r *ServiceEndpointJFrogResource) getCreateOrUpdateServiceEndpointArgs(model *ServiceEndpointJFrogResourceModel) *serviceendpoints.CreateOrUpdateServiceEndpointArgs {
	description := utils.IfThenElse[*string](model.Description != nil, model.Description, utils.EmptyString)

	var service string
	switch model.Service {
	case "artifactory":
		service = serviceendpoints.ServiceEndpointTypeJFrogArtifactory
	case "distribution":
		service = serviceendpoints.ServiceEndpointTypeJFrogDistribution
	case "platform":
		service = serviceendpoints.ServiceEndpointTypeJFrogPlatform
	case "xray":
		service = serviceendpoints.ServiceEndpointTypeJFrogXray
	}

	return &serviceendpoints.CreateOrUpdateServiceEndpointArgs{
		Description:       *description,
		GrantAllPipelines: model.GrantAllPipelines,
		Name:              model.Name,
		Password:          model.Password.ValueString(),
		Token:             model.AccessToken.ValueString(),
		Type:              service,
		Username:          model.Username.ValueString(),
		Url:               model.URL,
	}
}
