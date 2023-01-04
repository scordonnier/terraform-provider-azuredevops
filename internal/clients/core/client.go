package core

import (
	"context"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"net/url"
)

const (
	pathApis     = "_apis"
	pathProjects = "projects"
	pathTeams    = "teams"
)

type Client struct {
	restClient *networking.RestClient
}

func NewClient(restClient *networking.RestClient) *Client {
	return &Client{
		restClient: restClient,
	}
}

func (c *Client) GetProject(ctx context.Context, id string) (*TeamProject, error) {
	pathSegments := []string{pathApis, pathProjects, id}
	resp, err := c.restClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var project *TeamProject
	err = c.restClient.ParseJSON(ctx, resp, &project)
	return project, err
}

func (c *Client) GetProjects(ctx context.Context, state string, continuationToken string) (*[]TeamProjectReference, string, error) {
	pathSegments := []string{pathApis, pathProjects}
	queryParams := url.Values{"$top": []string{"100"}}
	if state != "" {
		queryParams.Add("stateFilter", state)
	}
	if continuationToken != "" {
		queryParams.Add("continuationToken", continuationToken)
	}

	resp, err := c.restClient.GetJSON(ctx, pathSegments, queryParams, networking.ApiVersion70)
	if err != nil {
		return nil, "", err
	}

	var project *TeamProjectReferenceCollection
	err = c.restClient.ParseJSON(ctx, resp, &project)
	continuationToken = resp.Header.Get(networking.HeaderKeyContinuationToken)
	return project.Value, continuationToken, err
}

func (c *Client) GetTeam(ctx context.Context, projectId string, id string) (*WebApiTeam, error) {
	pathSegments := []string{pathApis, pathProjects, projectId, pathTeams, id}
	resp, err := c.restClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var teams *WebApiTeam
	err = c.restClient.ParseJSON(ctx, resp, &teams)
	return teams, err
}

func (c *Client) GetTeams(ctx context.Context, projectId string) (*[]WebApiTeam, error) {
	pathSegments := []string{pathApis, pathProjects, projectId, pathTeams}
	resp, err := c.restClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var teams *WebApiTeamCollection
	err = c.restClient.ParseJSON(ctx, resp, &teams)
	return teams.Value, err
}
