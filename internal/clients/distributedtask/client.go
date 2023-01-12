package distributedtask

import (
	"context"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"strconv"
)

const (
	pathApis            = "_apis"
	pathDistributedTask = "distributedtask"
	pathEnvironments    = "environments"
	pathKubernetes      = "kubernetes"
	pathProviders       = "providers"
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

func (c *Client) UpdateEnvironment(ctx context.Context, projectId string, id int, name string, description string) (*EnvironmentInstance, error) {
	pathSegments := []string{projectId, pathApis, pathDistributedTask, pathEnvironments, strconv.Itoa(id)}
	body := &CreateOrUpdateEnvironmentArgs{
		Description: description,
		Name:        name,
	}
	environment, _, err := networking.PatchJSON[EnvironmentInstance](c.restClient, ctx, pathSegments, nil, body, networking.ApiVersion70)
	return environment, err
}
