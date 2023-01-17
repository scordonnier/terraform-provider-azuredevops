package pipelines

import (
	"context"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"strconv"
)

const (
	PipelinePermissionsResourceTypeEndpoint = "endpoint"

	pathApis                = "_apis"
	pathBuild               = "build"
	pathDistributedTask     = "distributedtask"
	pathEnvironments        = "environments"
	pathGeneralSettings     = "generalsettings"
	pathKubernetes          = "kubernetes"
	pathPipelinePermissions = "pipelinepermissions"
	pathPipelines           = "pipelines"
	pathProviders           = "providers"
	pathRetention           = "retention"
)

type Client struct {
	restClient *networking.RestClient
}

func NewClient(restClient *networking.RestClient) *Client {
	return &Client{
		restClient: restClient,
	}
}

func (c *Client) CreateEnvironment(ctx context.Context, projectId string, name string, description string) (*EnvironmentInstance, error) {
	pathSegments := []string{projectId, pathApis, pathDistributedTask, pathEnvironments}
	body := &CreateOrUpdateEnvironmentArgs{
		Description: description,
		Name:        name,
	}
	environment, _, err := networking.PostJSON[EnvironmentInstance](c.restClient, ctx, pathSegments, nil, body, networking.ApiVersion70)
	return environment, err
}

func (c *Client) CreateEnvironmentResourceKubernetes(ctx context.Context, projectId string, environmentId int, resource *EnvironmentResourceKubernetes) (*EnvironmentResourceKubernetes, error) {
	pathSegments := []string{projectId, pathApis, pathDistributedTask, pathEnvironments, strconv.Itoa(environmentId), pathProviders, pathKubernetes}
	environmentResource, _, err := networking.PostJSON[EnvironmentResourceKubernetes](c.restClient, ctx, pathSegments, nil, resource, networking.ApiVersion70)
	return environmentResource, err
}

func (c *Client) DeleteEnvironment(ctx context.Context, projectId string, id int) error {
	pathSegments := []string{projectId, pathApis, pathDistributedTask, pathEnvironments, strconv.Itoa(id)}
	_, _, err := networking.DeleteJSON[networking.NoJSON](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return err
}

func (c *Client) DeleteEnvironmentResourceKubernetes(ctx context.Context, projectId string, environmentId int, resourceId int) error {
	pathSegments := []string{projectId, pathApis, pathDistributedTask, pathEnvironments, strconv.Itoa(environmentId), pathProviders, pathKubernetes, strconv.Itoa(resourceId)}
	_, _, err := networking.DeleteJSON[networking.NoJSON](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return err
}

func (c *Client) GetEnvironment(ctx context.Context, projectId string, id int) (*EnvironmentInstance, error) {
	pathSegments := []string{projectId, pathApis, pathDistributedTask, pathEnvironments, strconv.Itoa(id)}
	environment, _, err := networking.GetJSON[EnvironmentInstance](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return environment, err
}

func (c *Client) GetEnvironmentResourceKubernetes(ctx context.Context, projectId string, environmentId int, resourceId int) (*EnvironmentResourceKubernetes, error) {
	pathSegments := []string{projectId, pathApis, pathDistributedTask, pathEnvironments, strconv.Itoa(environmentId), pathProviders, pathKubernetes, strconv.Itoa(resourceId)}
	environmentResource, _, err := networking.GetJSON[EnvironmentResourceKubernetes](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return environmentResource, err
}

func (c *Client) GetPipelinePermissions(ctx context.Context, projectId string, resourceType string, resourceId string) (*ResourcePipelinePermissions, error) {
	pathSegments := []string{projectId, pathApis, pathPipelines, pathPipelinePermissions, resourceType, resourceId}
	permissions, _, err := networking.GetJSON[ResourcePipelinePermissions](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70Preview1)
	return permissions, err
}

func (c *Client) GetPipelineRetentionSettings(ctx context.Context, projectId string) (*PipelineRetentionSettings, error) {
	pathSegments := []string{projectId, pathApis, pathBuild, pathRetention}
	settings, _, err := networking.GetJSON[PipelineRetentionSettings](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return settings, err
}

func (c *Client) GetPipelineSettings(ctx context.Context, projectId string) (*PipelineGeneralSettings, error) {
	pathSegments := []string{projectId, pathApis, pathBuild, pathGeneralSettings}
	settings, _, err := networking.GetJSON[PipelineGeneralSettings](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return settings, err
}

func (c *Client) GrantAllPipelines(ctx context.Context, projectId string, resourceType string, resourceId string, granted bool) (*ResourcePipelinePermissions, error) {
	pathSegments := []string{projectId, pathApis, pathPipelines, pathPipelinePermissions, resourceType, resourceId}
	body := &ResourcePipelinePermissions{
		AllPipelines: &Permission{
			Authorized: &granted,
		},
		Resource: &Resource{
			Id:   &resourceId,
			Type: &resourceType,
		},
	}
	permissions, _, err := networking.PatchJSON[ResourcePipelinePermissions](c.restClient, ctx, pathSegments, nil, body, networking.ApiVersion70Preview1)
	return permissions, err
}

func (c *Client) UpdateEnvironment(ctx context.Context, projectId string, id int, name string, description string) (*EnvironmentInstance, error) {
	pathSegments := []string{projectId, pathApis, pathDistributedTask, pathEnvironments, strconv.Itoa(id)}
	body := &CreateOrUpdateEnvironmentArgs{
		Description: description,
		Name:        name,
	}
	environment, _, err := networking.PatchJSON[EnvironmentInstance](c.restClient, ctx, pathSegments, nil, body, networking.ApiVersion70)
	return environment, err
}

func (c *Client) UpdatePipelineRetentionSettings(ctx context.Context, projectId string, settings *UpdatePipelineRetentionSettings) (*PipelineRetentionSettings, error) {
	pathSegments := []string{projectId, pathApis, pathBuild, pathRetention}
	retentionSettings, _, err := networking.PatchJSON[PipelineRetentionSettings](c.restClient, ctx, pathSegments, nil, settings, networking.ApiVersion70)
	return retentionSettings, err
}

func (c *Client) UpdatePipelineSettings(ctx context.Context, projectId string, settings *PipelineGeneralSettings) (*PipelineGeneralSettings, error) {
	pathSegments := []string{projectId, pathApis, pathBuild, pathGeneralSettings}
	generalSettings, _, err := networking.PatchJSON[PipelineGeneralSettings](c.restClient, ctx, pathSegments, nil, settings, networking.ApiVersion71Preview1)
	return generalSettings, err
}
