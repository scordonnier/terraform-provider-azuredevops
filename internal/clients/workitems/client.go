package workitems

import (
	"context"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
)

const (
	pathApis                         = "_apis"
	pathClassificationNodes          = "classificationnodes"
	pathClassificationNodeAreas      = "areas"
	pathClassificationNodeIterations = "iterations"
	pathWit                          = "wit"
)

type Client struct {
	restClient *networking.RestClient
}

func NewClient(restClient *networking.RestClient) *Client {
	return &Client{
		restClient: restClient,
	}
}

func (c *Client) GetArea(ctx context.Context, projectId string, path string) (*WorkItemClassificationNode, error) {
	pathSegments := []string{projectId, pathApis, pathWit, pathClassificationNodes, pathClassificationNodeAreas, path}
	area, _, err := networking.GetJSON[WorkItemClassificationNode](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return area, err
}

func (c *Client) GetIteration(ctx context.Context, projectId string, path string) (*WorkItemClassificationNode, error) {
	pathSegments := []string{projectId, pathApis, pathWit, pathClassificationNodes, pathClassificationNodeIterations, path}
	iteration, _, err := networking.GetJSON[WorkItemClassificationNode](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return iteration, err
}
