package serviceendpoints

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoints"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/validators"
	"golang.org/x/exp/slices"
)

var _ resource.Resource = &ServiceEndpointShareResource{}

func NewServiceEndpointShareResource() resource.Resource {
	return &ServiceEndpointShareResource{}
}

type ServiceEndpointShareResource struct {
	client *serviceendpoints.Client
}

type ServiceEndpointShareResourceModel struct {
	Description string       `tfsdk:"description"`
	Id          types.String `tfsdk:"id"`
	Name        string       `tfsdk:"name"`
	ProjectId   string       `tfsdk:"project_id"`
	ProjectIds  []string     `tfsdk:"project_ids"`
}

func (r *ServiceEndpointShareResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_share"
}

func (r *ServiceEndpointShareResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Shares a service endpoint with multiple Azure DevOps projects.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the service endpoint.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the service endpoint.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service endpoint.",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project hosting the service endpoint.",
				Required:            true,
				Validators: []validator.String{
					validators.UUID(),
				},
			},
			"project_ids": schema.SetAttribute{
				MarkdownDescription: "The IDs of the projects to share the service endpoint.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validators.UUID()),
				},
			},
		},
	}
}

func (r *ServiceEndpointShareResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointsClient
}

func (r *ServiceEndpointShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ServiceEndpointShareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.createOrUpdateServiceEndpointShare(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create service endpoint share", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ServiceEndpointShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, err := r.client.GetServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId)
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	var projectIds []string
	for _, reference := range *serviceEndpoint.ServiceEndpointProjectReferences {
		projectIds = append(projectIds, reference.ProjectReference.Id.String())
	}

	model.ProjectIds = projectIds

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ServiceEndpointShareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.createOrUpdateServiceEndpointShare(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update service endpoint share", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ServiceEndpointShareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var projectIds []string
	for _, projectId := range model.ProjectIds {
		if projectId != model.ProjectId {
			projectIds = append(projectIds, projectId)
		}
	}

	err := r.client.DeleteServiceEndpoint(ctx, model.Id.ValueString(), projectIds)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete service endpoint share", err.Error())
	}
}

// Private Methods

func (r *ServiceEndpointShareResource) createOrUpdateServiceEndpointShare(ctx context.Context, model *ServiceEndpointShareResourceModel) error {
	id := model.Id.ValueString()
	projectId := model.ProjectId

	serviceEndpoint, err := r.client.GetServiceEndpoint(ctx, id, projectId)
	if err != nil {
		return err
	}

	var currentProjectIds []string
	for _, projectReference := range *serviceEndpoint.ServiceEndpointProjectReferences {
		currentProjectIds = append(currentProjectIds, projectReference.ProjectReference.Id.String())
	}

	var deleteProjectIds []string
	for _, deleteProjectId := range currentProjectIds {
		if !slices.Contains(model.ProjectIds, deleteProjectId) {
			deleteProjectIds = append(deleteProjectIds, deleteProjectId)
		}
	}

	if len(deleteProjectIds) > 0 {
		deleteErr := r.client.DeleteServiceEndpoint(ctx, id, deleteProjectIds)
		if deleteErr != nil {
			return deleteErr
		}
	}

	var addProjectIds []string
	for _, addProjectId := range model.ProjectIds {
		if !slices.Contains(currentProjectIds, addProjectId) {
			addProjectIds = append(addProjectIds, addProjectId)
		}
	}

	if len(addProjectIds) > 0 {
		err := r.client.ShareServiceEndpoint(ctx, id, model.Name, model.Description, projectId, addProjectIds)
		if err != nil {
			return err
		}
	}

	return nil
}
