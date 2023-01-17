package serviceendpoints

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoints"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"gopkg.in/yaml.v3"
)

var _ resource.Resource = &ServiceEndpointKubernetesResource{}

func NewServiceEndpointKubernetesResource() resource.Resource {
	return &ServiceEndpointKubernetesResource{}
}

type ServiceEndpointKubernetesResource struct {
	pipelinesClient        *pipelines.Client
	serviceEndpointsClient *serviceendpoints.Client
}

type ServiceEndpointKubernetesResourceModel struct {
	Description       *string                   `tfsdk:"description"`
	GrantAllPipelines bool                      `tfsdk:"grant_all_pipelines"`
	Id                types.String              `tfsdk:"id"`
	Kubeconfig        ServiceEndpointKubeconfig `tfsdk:"kubeconfig"`
	Name              string                    `tfsdk:"name"`
	ProjectId         string                    `tfsdk:"project_id"`
}

type ServiceEndpointKubeconfig struct {
	AcceptUntrustedCertificates bool   `tfsdk:"accept_untrusted_certs"`
	YamlContent                 string `tfsdk:"yaml_content"`
}

func (r *ServiceEndpointKubernetesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_kubernetes"
}

func (r *ServiceEndpointKubernetesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceShema := GetServiceEndpointResourceSchemaBase("Manages a Kubernetes service endpoint within an Azure DevOps project.")
	resourceShema.Attributes["kubeconfig"] = schema.SingleNestedAttribute{
		MarkdownDescription: "The information required to connect a cluster with a kubeconfig.",
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"accept_untrusted_certs": schema.BoolAttribute{
				MarkdownDescription: "Set to true to allow clients to accept a self-signed certificate.",
				Required:            true,
			},
			"yaml_content": schema.StringAttribute{
				MarkdownDescription: "The content of the kubeconfig in YAML notation to be used to communicate with the API-Server of Kubernetes. The kubeconfig MUST contains only 1 cluster.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
	resp.Schema = resourceShema
}

func (r *ServiceEndpointKubernetesResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.pipelinesClient = req.ProviderData.(*clients.AzureDevOpsClient).PipelinesClient
	r.serviceEndpointsClient = req.ProviderData.(*clients.AzureDevOpsClient).ServiceEndpointsClient
}

func (r *ServiceEndpointKubernetesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ServiceEndpointKubernetesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	args, err := r.getCreateOrUpdateServiceEndpointArgs(model)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build service endpoint arguments.", err.Error())
		return
	}

	serviceEndpoint, err := CreateResourceServiceEndpoint(ctx, model.ProjectId, args, r.serviceEndpointsClient, r.pipelinesClient, resp)
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

	serviceEndpoint, granted, err := ReadResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointsClient, r.pipelinesClient, resp)
	if err != nil {
		return
	}

	model.Description = utils.IfThenElse[*string](serviceEndpoint.Description != nil, model.Description, utils.EmptyString)
	model.GrantAllPipelines = granted
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
		resp.Diagnostics.AddError("Unable to build service endpoint arguments.", err.Error())
		return
	}

	_, err = UpdateResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, args, r.serviceEndpointsClient, r.pipelinesClient, resp)
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

	DeleteResourceServiceEndpoint(ctx, model.Id.ValueString(), model.ProjectId, r.serviceEndpointsClient, resp)
}

// Private Methods

func (r *ServiceEndpointKubernetesResource) getCreateOrUpdateServiceEndpointArgs(model *ServiceEndpointKubernetesResourceModel) (*serviceendpoints.CreateOrUpdateServiceEndpointArgs, error) {
	var yamlKubeconfig map[string]interface{}
	err := yaml.Unmarshal([]byte(model.Kubeconfig.YamlContent), &yamlKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("kubeconfig contains an invalid YAML : %s", err)
	}

	clusters := yamlKubeconfig["clusters"].([]interface{})
	contexts := yamlKubeconfig["contexts"].([]interface{})
	if len(clusters) == 0 || len(clusters) > 1 || len(contexts) == 0 || len(contexts) > 1 {
		return nil, errors.New("kubeconfig contains no or more than one cluster/context")
	}

	server := clusters[0].(map[string]interface{})["cluster"].(map[string]interface{})["server"].(string)
	clusterContext := contexts[0].(map[string]interface{})["name"].(string)
	description := utils.IfThenElse[*string](model.Description != nil, model.Description, utils.EmptyString)
	return &serviceendpoints.CreateOrUpdateServiceEndpointArgs{
		AcceptUntrustedCertificates: model.Kubeconfig.AcceptUntrustedCertificates,
		ClusterContext:              clusterContext,
		GrantAllPipelines:           model.GrantAllPipelines,
		Description:                 *description,
		Kubeconfig:                  model.Kubeconfig.YamlContent,
		Name:                        model.Name,
		Type:                        serviceendpoints.ServiceEndpointTypekubernetes,
		Url:                         server,
	}, nil
}
