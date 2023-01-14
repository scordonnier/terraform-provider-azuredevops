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

func (c *Client) CreateArea(ctx context.Context, projectId string, path string, name string) (*WorkItemClassificationNode, error) {
	return c.createWorkItemClassificationNode(ctx, pathClassificationNodeAreas, projectId, path, name, nil)
}

func (c *Client) CreateIteration(ctx context.Context, projectId string, path string, name string, startDate *string, finishDate *string) (*WorkItemClassificationNode, error) {
	return c.createWorkItemClassificationNode(ctx, pathClassificationNodeIterations, projectId, path, name, c.buildIterationAttributes(startDate, finishDate))
}

func (c *Client) DeleteArea(ctx context.Context, projectId string, path string) error {
	return c.deleteWorkItemClassificationNode(ctx, pathClassificationNodeAreas, projectId, path)
}

func (c *Client) DeleteIteration(ctx context.Context, projectId string, path string) error {
	return c.deleteWorkItemClassificationNode(ctx, pathClassificationNodeIterations, projectId, path)
}

func (c *Client) GetArea(ctx context.Context, projectId string, path string) (*WorkItemClassificationNode, error) {
	return c.getWorkItemClassificationNode(ctx, pathClassificationNodeAreas, projectId, path)
}

func (c *Client) GetIteration(ctx context.Context, projectId string, path string) (*WorkItemClassificationNode, error) {
	return c.getWorkItemClassificationNode(ctx, pathClassificationNodeIterations, projectId, path)
}

func (c *Client) MoveArea(ctx context.Context, projectId string, path string, nodeId int, name string) (*WorkItemClassificationNode, error) {
	return c.moveWorkItemClassificationNode(ctx, pathClassificationNodeAreas, projectId, path, nodeId, name, nil)
}

func (c *Client) MoveIteration(ctx context.Context, projectId string, path string, nodeId int, name string, startDate *string, finishDate *string) (*WorkItemClassificationNode, error) {
	return c.moveWorkItemClassificationNode(ctx, pathClassificationNodeIterations, projectId, path, nodeId, name, c.buildIterationAttributes(startDate, finishDate))
}

func (c *Client) UpdateArea(ctx context.Context, projectId string, path string, name string) (*WorkItemClassificationNode, error) {
	return c.updateWorkItemClassificationNode(ctx, pathClassificationNodeAreas, projectId, path, name, nil)
}

func (c *Client) UpdateIteration(ctx context.Context, projectId string, path string, name string, startDate *string, finishDate *string) (*WorkItemClassificationNode, error) {
	return c.updateWorkItemClassificationNode(ctx, pathClassificationNodeIterations, projectId, path, name, c.buildIterationAttributes(startDate, finishDate))
}

// Private Methods

func (c *Client) buildIterationAttributes(startDate *string, finishDate *string) *map[string]any {
	attributes := map[string]any{}
	if startDate != nil && finishDate != nil {
		attributes["startDate"] = startDate
		attributes["finishDate"] = finishDate
	}
	return &attributes
}

func (c *Client) createWorkItemClassificationNode(ctx context.Context, nodeType string, projectId string, path string, name string, attributes *map[string]any) (*WorkItemClassificationNode, error) {
	pathSegments := []string{projectId, pathApis, pathWit, pathClassificationNodes, nodeType, path}
	body := WorkItemClassificationNode{
		Attributes: attributes,
		Name:       &name,
	}
	area, _, err := networking.PostJSON[WorkItemClassificationNode](c.restClient, ctx, pathSegments, nil, body, networking.ApiVersion70)
	return area, err
}

func (c *Client) deleteWorkItemClassificationNode(ctx context.Context, nodeType string, projectId string, path string) error {
	pathSegments := []string{projectId, pathApis, pathWit, pathClassificationNodes, nodeType, path}
	_, _, err := networking.DeleteJSON[networking.NoJSON](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return err
}

func (c *Client) getWorkItemClassificationNode(ctx context.Context, nodeType string, projectId string, path string) (*WorkItemClassificationNode, error) {
	pathSegments := []string{projectId, pathApis, pathWit, pathClassificationNodes, nodeType, path}
	area, _, err := networking.GetJSON[WorkItemClassificationNode](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return area, err
}

func (c *Client) moveWorkItemClassificationNode(ctx context.Context, nodeType string, projectId string, path string, nodeId int, name string, attributes *map[string]any) (*WorkItemClassificationNode, error) {
	pathSegments := []string{projectId, pathApis, pathWit, pathClassificationNodes, nodeType, path}
	body := WorkItemClassificationNode{
		Attributes: attributes,
		Id:         &nodeId,
		Name:       &name,
	}
	area, _, err := networking.PostJSON[WorkItemClassificationNode](c.restClient, ctx, pathSegments, nil, body, networking.ApiVersion70)
	return area, err
}

func (c *Client) updateWorkItemClassificationNode(ctx context.Context, nodeType string, projectId string, path string, name string, attributes *map[string]any) (*WorkItemClassificationNode, error) {
	pathSegments := []string{projectId, pathApis, pathWit, pathClassificationNodes, nodeType, path}
	body := WorkItemClassificationNode{
		Attributes: attributes,
		Name:       &name,
	}
	area, _, err := networking.PatchJSON[WorkItemClassificationNode](c.restClient, ctx, pathSegments, nil, body, networking.ApiVersion70)
	return area, err
}
