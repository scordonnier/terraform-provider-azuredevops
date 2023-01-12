package pipelines

import (
	"context"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
)

const (
	PipelinePermissionsResourceTypeEndpoint = "endpoint"

	pathApis                = "_apis"
	pathPipelinePermissions = "pipelinepermissions"
	pathPipelines           = "pipelines"
)

type Client struct {
	restClient *networking.RestClient
}

func NewClient(restClient *networking.RestClient) *Client {
	return &Client{
		restClient: restClient,
	}
}

func (c *Client) GetPipelinePermissions(ctx context.Context, projectId string, resourceType string, resourceId string) (*ResourcePipelinePermissions, error) {
	pathSegments := []string{projectId, pathApis, pathPipelines, pathPipelinePermissions, resourceType, resourceId}
	permissions, _, err := networking.GetJSON[ResourcePipelinePermissions](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70Preview1)
	return permissions, err
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
