package serviceendpoint

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoint"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"gopkg.in/yaml.v3"
)

var _ resource.Resource = &ServiceEndpointKubernetesResource{}
var _ resource.ResourceWithImportState = &ServiceEndpointKubernetesResource{}

func NewServiceEndpointKubernetesResource() resource.Resource {
	return &ServiceEndpointKubernetesResource{}
}

type ServiceEndpointKubernetesResource struct {
	client *serviceendpoint.Client
}

type ServiceEndpointKubernetesResourceModel struct {
	Description string                    `tfsdk:"description"`
	Id          types.String              `tfsdk:"id"`
	Kubeconfig  ServiceEndpointKubeconfig `tfsdk:"kubeconfig"`
	Name        string                    `tfsdk:"name"`
	ProjectId   string                    `tfsdk:"project_id"`
}

type ServiceEndpointKubeconfig struct {
	AcceptUntrustedCertificates bool   `tfsdk:"accept_untrusted_certs"`
	Yaml                        string `tfsdk:"yaml"`
}

func (r *ServiceEndpointKubernetesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_kubernetes"
}

func (r *ServiceEndpointKubernetesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kubernetes service endpoint within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the service endpoint.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service endpoint.",
				Required:            true,
				Validators: []validator.String{
					utils.StringNotEmptyValidator(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"kubeconfig": schema.SingleNestedBlock{
				MarkdownDescription: "The information required to connect a cluster with a kubeconfig file.",
				Attributes: map[string]schema.Attribute{
					"accept_untrusted_certs": schema.BoolAttribute{
						MarkdownDescription: "Set this option to allow clients to accept a self-signed certificate.",
						Required:            true,
					},
					"yaml": schema.StringAttribute{
						MarkdownDescription: "The content of the kubeconfig in YAML notation to be used to communicate with the API-Server of Kubernetes. The kubeconfig MUST contains only 1 cluster.",
						Required:            true,
						Sensitive:           true,
						Validators: []validator.String{
							utils.StringNotEmptyValidator(),
						},
					},
				},
			},
		},
	}
}

func (r *ServiceEndpointKubernetesResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointClient
}

func (r *ServiceEndpointKubernetesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ServiceEndpointKubernetesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	args, err := r.getCreateOrUpdateServiceEndpointArgs(model)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	serviceEndpoint, err := CreateResourceServiceEndpoint(ctx, args, model.ProjectId, r.client, resp)
	if err != nil {
		return
	}

	model.Id = types.StringValue(serviceEndpoint.Id.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointKubernetesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ServiceEndpointKubernetesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serviceEndpoint, err := ReadResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.client, resp)
	if err != nil {
		return
	}

	model.Description = *serviceEndpoint.Description
	model.Name = *serviceEndpoint.Name

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointKubernetesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model *ServiceEndpointKubernetesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	args, err := r.getCreateOrUpdateServiceEndpointArgs(model)
	if err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	_, err = UpdateResourceServiceEndpoint(ctx, model.Id.ValueString(), args, model.ProjectId, r.client, resp)
	if err != nil {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServiceEndpointKubernetesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ServiceEndpointKubernetesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.client, resp)
}

func (r *ServiceEndpointKubernetesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Private Methods

func (r *ServiceEndpointKubernetesResource) getCreateOrUpdateServiceEndpointArgs(model *ServiceEndpointKubernetesResourceModel) (*serviceendpoint.CreateOrUpdateServiceEndpointArgs, error) {
	var yamlKubeconfig map[string]interface{}
	err := yaml.Unmarshal([]byte(model.Kubeconfig.Yaml), &yamlKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("kubeconfig contains an invalid YAML : %s", err)
	}

	clusters := yamlKubeconfig["clusters"].([]interface{})
	contexts := yamlKubeconfig["contexts"].([]interface{})
	if len(clusters) == 0 || len(clusters) > 1 || len(contexts) == 0 || len(contexts) > 1 {
		return nil, errors.New("kubeconfig contains no clusters/contexts or more than one clusters/contexts")
	}

	server := clusters[0].(map[string]interface{})["cluster"].(map[string]interface{})["server"].(string)
	clusterContext := contexts[0].(map[string]interface{})["name"].(string)
	return &serviceendpoint.CreateOrUpdateServiceEndpointArgs{
		AcceptUntrustedCertificates: model.Kubeconfig.AcceptUntrustedCertificates,
		ClusterContext:              clusterContext,
		Description:                 model.Description,
		Kubeconfig:                  model.Kubeconfig.Yaml,
		Name:                        model.Name,
		Type:                        serviceendpoint.ServiceEndpointTypekubernetes,
		Url:                         server,
	}, nil
}
