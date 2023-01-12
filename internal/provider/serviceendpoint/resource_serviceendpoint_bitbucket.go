package serviceendpoint

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoint"
)

var _ resource.Resource = &ServiceEndpointBitbucketResource{}
var _ resource.ResourceWithImportState = &ServiceEndpointBitbucketResource{}

func NewServiceEndpointBitbucketResource() resource.Resource {
	return &ServiceEndpointBitbucketResource{}
}

type ServiceEndpointBitbucketResource struct {
	pipelineClient        *pipelines.Client
	serviceEndpointClient *serviceendpoint.Client
}

type ServiceEndpointBitbucketResourceModel struct {
	Description       string       `tfsdk:"description"`
	Id                types.String `tfsdk:"id"`
	GrantAllPipelines bool         `tfsdk:"grant_all_pipelines"`
	Name              string       `tfsdk:"name"`
	Password          string       `tfsdk:"password"`
	ProjectId         string       `tfsdk:"project_id"`
	UserName          string       `tfsdk:"username"`
}

func (r *ServiceEndpointBitbucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_bitbucket"
}

func (r *ServiceEndpointBitbucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceShema := GetServiceEndpointResourceSchemaBase("Manages a Bitbucket service endpoint within an Azure DevOps project.")
	resourceShema.Attributes["password"] = schema.StringAttribute{
		MarkdownDescription: "Bitbucket account password.",
		Required:            true,
		Sensitive:           true,
	}
	resourceShema.Attributes["username"] = schema.StringAttribute{
		MarkdownDescription: "Bitbucket account username.",
		Required:            true,
		Sensitive:           true,
	}
	resp.Schema = resourceShema
}

func (r *ServiceEndpointBitbucketResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.pipelineClient = req.ProviderData.(*clients.AzureDevOpsClient).PipelineClient
	r.serviceEndpointClient = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointClient
}

func (r *ServiceEndpointBitbucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ServiceEndpointBitbucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, err := CreateResourceServiceEndpoint(ctx, model.ProjectId, r.getCreateOrUpdateServiceEndpointArgs(model), r.serviceEndpointClient, r.pipelineClient, resp)
	if err != nil {
		return
	}

	model.Id = types.StringValue(serviceEndpoint.Id.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointBitbucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ServiceEndpointBitbucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, granted, err := ReadResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointClient, r.pipelineClient, resp)
	if err != nil {
		return
	}

	model.Description = *serviceEndpoint.Description
	model.GrantAllPipelines = granted
	model.Name = *serviceEndpoint.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointBitbucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ServiceEndpointBitbucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := UpdateResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.getCreateOrUpdateServiceEndpointArgs(model), r.serviceEndpointClient, r.pipelineClient, resp)
	if err != nil {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointBitbucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ServiceEndpointBitbucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointClient, resp)
}

func (r *ServiceEndpointBitbucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Private Methods

func (r *ServiceEndpointBitbucketResource) getCreateOrUpdateServiceEndpointArgs(model *ServiceEndpointBitbucketResourceModel) *serviceendpoint.CreateOrUpdateServiceEndpointArgs {
	return &serviceendpoint.CreateOrUpdateServiceEndpointArgs{
		Description:       model.Description,
		GrantAllPipelines: model.GrantAllPipelines,
		Name:              model.Name,
		Password:          model.Password,
		Type:              serviceendpoint.ServiceEndpointTypeBitbucket,
		UserName:          model.UserName,
	}
}
