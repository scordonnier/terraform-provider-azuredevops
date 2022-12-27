package core

import (
	"context"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
)

const (
	pathProjects = "projects"
)

type Client struct {
	restClient *networking.RestClient
}

func NewClient(restClient *networking.RestClient) *Client {
	return &Client{
		restClient: restClient,
	}
}

func (client *Client) GetProject(ctx context.Context, id string) (*TeamProject, error) {
	pathSegments := []string{pathProjects, id}
	resp, err := client.restClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var project *TeamProject
	err = client.restClient.ParseJSON(resp, &project)
	return project, err
}
