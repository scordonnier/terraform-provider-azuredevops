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
	pathApis               = "_apis"
	pathFeatureManagement  = "FeatureManagement"
	pathFeatureStates      = "FeatureStates"
	pathFeatureStatesQuery = "FeatureStatesQuery"
	pathHost               = "host"
	pathOperations         = "operations"
	pathProcess            = "process"
	pathProcesses          = "processes"
	pathProject            = "project"
	pathProjects           = "projects"
	pathTeams              = "teams"
)

type Client struct {
	restClient *networking.RestClient
}

func NewClient(restClient *networking.RestClient) *Client {
	return &Client{
		restClient: restClient,
	}
}

func (c *Client) CreateProject(ctx context.Context, name string, description string, visibility string, processTemplateId string, versionControl string) (*OperationReference, error) {
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
		Description: &description,
		Name:        &name,
		Visibility:  &visibility,
	}
	operation, _, err := networking.PostJSON[OperationReference](c.restClient, ctx, pathSegments, nil, body, networking.ApiVersion70)
	return operation, err
}

func (c *Client) CreateTeam(ctx context.Context, projectId string, name string, description string) (*WebApiTeam, error) {
	pathSegments := []string{pathApis, pathProjects, projectId, pathTeams}
	body := WebApiTeam{
		Description: &description,
		Name:        &name,
	}
	team, _, err := networking.PostJSON[WebApiTeam](c.restClient, ctx, pathSegments, nil, body, networking.ApiVersion70)
	return team, err
}

func (c *Client) DeleteProject(ctx context.Context, projectId string) (*OperationReference, error) {
	pathSegments := []string{pathApis, pathProjects, projectId}
	operation, _, err := networking.DeleteJSON[OperationReference](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return operation, err
}

func (c *Client) DeleteTeam(ctx context.Context, projectId string, teamId string) error {
	pathSegments := []string{pathApis, pathProjects, projectId, pathTeams, teamId}
	_, _, err := networking.DeleteJSON[networking.NoJSON](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return err
}

func (c *Client) GetOperation(ctx context.Context, id string, pluginId *uuid.UUID) (*Operation, error) {
	pathSegments := []string{pathApis, pathOperations, id}
	queryParams := url.Values{}
	if pluginId != nil {
		queryParams.Add("pluginId", pluginId.String())
	}
	operation, _, err := networking.GetJSON[Operation](c.restClient, ctx, pathSegments, queryParams, networking.ApiVersion70)
	return operation, err
}

func (c *Client) GetProcess(ctx context.Context, name string) (*Process, error) {
	pathSegments := []string{pathApis, pathProcess, pathProcesses}
	processes, _, err := networking.GetJSON[ProcessCollection](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
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
	project, _, err := networking.GetJSON[TeamProject](c.restClient, ctx, pathSegments, queryParams, networking.ApiVersion70)
	return project, err
}

func (c *Client) GetProjectFeatures(ctx context.Context, projectId string) (*ContributedFeatureStateQuery, error) {
	pathSegments := []string{pathApis, pathFeatureManagement, pathFeatureStatesQuery, pathHost, pathProject, projectId}
	body := &ContributedFeatureStateQuery{
		FeatureIds: &[]string{
			ProjectFeatureBoards,
			ProjectFeatureRepositories,
			ProjectFeaturePipelines,
			ProjectFeatureTestPlans,
			ProjectFeatureArtifacts,
		},
		ScopeValues: &map[string]string{
			"project": projectId,
		},
	}
	featureStates, _, err := networking.PostJSON[ContributedFeatureStateQuery](c.restClient, ctx, pathSegments, nil, body, networking.ApiVersion70Preview1)
	return featureStates, err
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
	projects, resp, err := networking.GetJSON[TeamProjectReferenceCollection](c.restClient, ctx, pathSegments, queryParams, networking.ApiVersion70)
	if err != nil {
		return nil, "", err
	}

	continuationToken = resp.Header.Get(networking.HeaderKeyContinuationToken)
	return projects.Value, continuationToken, err
}

func (c *Client) GetTeam(ctx context.Context, projectId string, id string) (*WebApiTeam, error) {
	pathSegments := []string{pathApis, pathProjects, projectId, pathTeams, id}
	team, _, err := networking.GetJSON[WebApiTeam](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	return team, err
}

func (c *Client) GetTeams(ctx context.Context, projectId string) (*[]WebApiTeam, error) {
	pathSegments := []string{pathApis, pathProjects, projectId, pathTeams}
	teams, _, err := networking.GetJSON[WebApiTeamCollection](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

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

func (c *Client) UpdateProject(ctx context.Context, projectId string, name string, description string) (*OperationReference, error) {
	pathSegments := []string{pathApis, pathProjects, projectId}
	project := TeamProject{Description: &description}
	if name != "" {
		project.Name = &name
	}
	operation, _, err := networking.PatchJSON[OperationReference](c.restClient, ctx, pathSegments, nil, project, networking.ApiVersion70)
	return operation, err
}

func (c *Client) UpdateProjectFeature(ctx context.Context, projectId string, featureId string, state string) (*ContributedFeatureState, error) {
	pathSegments := []string{pathApis, pathFeatureManagement, pathFeatureStates, pathHost, pathProject, projectId, featureId}
	body := &ContributedFeatureState{
		FeatureId: &featureId,
		Scope: &ContributedFeatureSettingScope{
			SettingScope: utils.String("project"),
			UserScoped:   utils.Bool(false),
		},
		State: &state,
	}
	featureState, _, err := networking.PatchJSON[ContributedFeatureState](c.restClient, ctx, pathSegments, nil, body, networking.ApiVersion70Preview1)
	return featureState, err
}

func (c *Client) UpdateTeam(ctx context.Context, projectId string, teamId string, name string, description string) (*WebApiTeam, error) {
	pathSegments := []string{pathApis, pathProjects, projectId, pathTeams, teamId}
	body := WebApiTeam{
		Description: &description,
		Name:        &name,
	}
	team, _, err := networking.PatchJSON[WebApiTeam](c.restClient, ctx, pathSegments, nil, body, networking.ApiVersion70)
	return team, err
}

// Private Methods

func (c *Client) operationStatusRefreshFunc(ctx context.Context, client *Client, operation *OperationReference) utils.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pendingOperation, err := client.GetOperation(ctx, operation.Id.String(), operation.PluginId)
		if err != nil {
			return nil, "failed", err
		}

		return pendingOperation, *pendingOperation.Status, err
	}
}
