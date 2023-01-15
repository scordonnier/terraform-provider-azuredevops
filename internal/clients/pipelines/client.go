package pipelines

import (
	"context"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
)

const (
	PipelinePermissionsResourceTypeEndpoint = "endpoint"

	pathApis                = "_apis"
	pathBuild               = "build"
	pathGeneralSettings     = "generalsettings"
	pathPipelinePermissions = "pipelinepermissions"
	pathPipelines           = "pipelines"
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
