package pipelines

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/validators"
)

var _ resource.Resource = &EnvironmentKubernetesResource{}

func NewEnvironmentKubernetesResource() resource.Resource {
	return &EnvironmentKubernetesResource{}
}

type EnvironmentKubernetesResource struct {
	client *pipelines.Client
}

type EnvironmentKubernetesResourceModel struct {
	EnvironmentId     int         `tfsdk:"environment_id"`
	Id                types.Int64 `tfsdk:"id"`
	Name              string      `tfsdk:"name"`
	Namespace         string      `tfsdk:"namespace"`
	ProjectId         string      `tfsdk:"project_id"`
	ServiceEndpointId string      `tfsdk:"service_endpoint_id"`
}

func (r *EnvironmentKubernetesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_kubernetes"
}

func (r *EnvironmentKubernetesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage a Kubernetes resource on an environment in Azure Pipelines.",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the environment.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validators.StringNotEmpty(),
				},
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The namespace on the Kubernetes cluster.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validators.StringNotEmpty(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validators.UUID(),
				},
			},
			"service_endpoint_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the service endpoint.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validators.UUID(),
				},
			},
		},
	}
}

func (r *EnvironmentKubernetesResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).PipelinesClient
}

func (r *EnvironmentKubernetesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *EnvironmentKubernetesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	environmentResource := getEnvironmentResource(model)
	environmentResource, err := r.client.CreateEnvironmentResourceKubernetes(ctx, model.ProjectId, model.EnvironmentId, environmentResource)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create the Kubernetes resource", err.Error())
		return
	}

	model.Id = types.Int64Value(int64(*environmentResource.Id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentKubernetesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *EnvironmentKubernetesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.GetEnvironmentResourceKubernetes(ctx, model.ProjectId, model.EnvironmentId, int(model.Id.ValueInt64()))
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Unable to find Kubernetes resource with Id '%d'", model.Id.ValueInt64()), err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentKubernetesResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *EnvironmentKubernetesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *EnvironmentKubernetesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEnvironmentResourceKubernetes(ctx, model.ProjectId, model.EnvironmentId, int(model.Id.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete Kubernetes resource", err.Error())
	}
}

// Private Methods

func getEnvironmentResource(model *EnvironmentKubernetesResourceModel) *pipelines.EnvironmentResourceKubernetes {
	return &pipelines.EnvironmentResourceKubernetes{
		Name:              &model.Name,
		Namespace:         &model.Namespace,
		ServiceEndpointId: &model.ServiceEndpointId,
	}
}
