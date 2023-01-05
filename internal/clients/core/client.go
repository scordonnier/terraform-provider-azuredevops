package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/google/uuid"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"net/url"
	"strings"
	"time"
)

const (
	pathApis       = "_apis"
	pathOperations = "operations"
	pathProcess    = "process"
	pathProcesses  = "processes"
	pathProjects   = "projects"
	pathTeams      = "teams"
)

type Client struct {
	restClient *networking.RestClient
}

func NewClient(restClient *networking.RestClient) *Client {
	return &Client{
		restClient: restClient,
	}
}

func (c *Client) CreateProject(ctx context.Context, name string, description *string, visibility string, processTemplateId string, versionControl string) (*OperationReference, error) {
	pathSegments := []string{pathApis, pathProjects}
	body := TeamProject{
		Capabilities: &map[string]map[string]string{
			CapabilitiesProcessTemplate: {
				CapabilitiesProcessTemplateTypeId: processTemplateId,
			},
			CapabilitiesVersionControl: {
				CapabilitiesVersionControlType: versionControl,
			},
		},
		Description: description,
		Name:        &name,
		Visibility:  &visibility,
	}
	resp, err := c.restClient.PostJSON(ctx, pathSegments, nil, body, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var operation *OperationReference
	err = c.restClient.ParseJSON(ctx, resp, &operation)
	return operation, err
}

func (c *Client) CreateTeam(ctx context.Context, projectId string, name string, description string) (*WebApiTeam, error) {
	pathSegments := []string{pathApis, pathProjects, projectId, pathTeams}
	body := WebApiTeam{
		Description: &description,
		Name:        &name,
	}
	resp, err := c.restClient.PostJSON(ctx, pathSegments, nil, body, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var team *WebApiTeam
	err = c.restClient.ParseJSON(ctx, resp, &team)
	return team, err
}

func (c *Client) DeleteProject(ctx context.Context, projectId string) (*OperationReference, error) {
	pathSegments := []string{pathApis, pathProjects, projectId}
	resp, err := c.restClient.DeleteJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var operation *OperationReference
	err = c.restClient.ParseJSON(ctx, resp, &operation)
	return operation, err
}

func (c *Client) DeleteTeam(ctx context.Context, projectId string, teamId string) error {
	pathSegments := []string{pathApis, pathProjects, projectId, pathTeams, teamId}
	_, err := c.restClient.DeleteJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	return err
}

func (c *Client) GetOperation(ctx context.Context, id string, pluginId *uuid.UUID) (*Operation, error) {
	pathSegments := []string{pathApis, pathOperations, id}
	queryParams := url.Values{}
	if pluginId != nil {
		queryParams.Add("pluginId", pluginId.String())
	}

	resp, err := c.restClient.GetJSON(ctx, pathSegments, queryParams, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var operation *Operation
	err = c.restClient.ParseJSON(ctx, resp, &operation)
	return operation, err
}

func (c *Client) GetProcess(ctx context.Context, name string) (*Process, error) {
	pathSegments := []string{pathApis, pathProcess, pathProcesses}
	resp, err := c.restClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var processes *ProcessCollection
	err = c.restClient.ParseJSON(ctx, resp, &processes)
	if err != nil {
		return nil, err
	}

	process := linq.From(*processes.Value).FirstWith(func(p interface{}) bool {
		return strings.EqualFold(*p.(Process).Name, name)
	})
	if process == nil {
		return nil, errors.New(fmt.Sprintf("Unable to find process with name '%s'", name))
	}

	p := process.(Process)
	return &p, nil
}

func (c *Client) GetProject(ctx context.Context, id string) (*TeamProject, error) {
	pathSegments := []string{pathApis, pathProjects, id}
	queryParams := url.Values{"includeCapabilities": []string{"true"}}
	resp, err := c.restClient.GetJSON(ctx, pathSegments, queryParams, networking.ApiVersion70)
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

func (c *Client) OperationStateChangeConf(ctx context.Context, client *Client, operation *OperationReference) *utils.StateChangeConf {
	return &utils.StateChangeConf{
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
		Pending:    []string{"inProgress", "queued", "notSet"},
		Target:     []string{"failed", "succeeded", "cancelled"},
		Refresh:    client.operationStatusRefreshFunc(ctx, client, operation),
		Timeout:    2 * time.Minute,
	}
}

func (c *Client) UpdateProject(ctx context.Context, projectId string, name string, description *string) (*OperationReference, error) {
	pathSegments := []string{pathApis, pathProjects, projectId}
	body := TeamProject{
		Description: description,
		Name:        &name,
	}
	resp, err := c.restClient.PatchJSON(ctx, pathSegments, nil, body, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var operation *OperationReference
	err = c.restClient.ParseJSON(ctx, resp, &operation)
	return operation, err
}

func (c *Client) UpdateTeam(ctx context.Context, projectId string, teamId string, name string, description string) (*WebApiTeam, error) {
	pathSegments := []string{pathApis, pathProjects, projectId, pathTeams, teamId}
	body := WebApiTeam{
		Description: &description,
		Name:        &name,
	}
	resp, err := c.restClient.PatchJSON(ctx, pathSegments, nil, body, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var team *WebApiTeam
	err = c.restClient.ParseJSON(ctx, resp, &team)
	return team, err
}

func (c *Client) operationStatusRefreshFunc(ctx context.Context, client *Client, operation *OperationReference) utils.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pendingOperation, err := client.GetOperation(ctx, operation.Id.String(), operation.PluginId)
		if err != nil {
			return nil, "failed", err
		}

		return pendingOperation, *pendingOperation.Status, err
	}
}
