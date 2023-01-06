package graph

import (
	"context"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"net/url"
)

const (
	queryContinuationToken = "continuationToken"
	queryScopeDescriptor   = "scopeDescriptor"

	pathApis        = "_apis"
	pathDescriptors = "descriptors"
	pathGraph       = "graph"
	pathGroups      = "groups"
	pathUsers       = "users"
)

type Client struct {
	vsspsClient *networking.RestClient
}

func NewClient(vsspsClient *networking.RestClient) *Client {
	return &Client{
		vsspsClient: vsspsClient,
	}
}

func (c *Client) CreateGroup(ctx context.Context, projectId string, name string, description string) (*GraphGroup, error) {
	descriptor, err := c.GetDescriptor(ctx, projectId)
	if err != nil {
		return nil, err
	}

	body := GraphGroupVstsCreationContext{
		DisplayName: &name,
		Description: &description,
	}
	pathSegments := []string{pathApis, pathGraph, pathGroups}
	queryParams := url.Values{"scopeDescriptor": []string{*descriptor}}
	resp, err := c.vsspsClient.PostJSON(ctx, pathSegments, queryParams, body, networking.ApiVersion70Preview1)
	if err != nil {
		return nil, err
	}

	var group *GraphGroup
	err = c.vsspsClient.ParseJSON(ctx, resp, &group)
	return group, err
}

func (c *Client) DeleteGroup(ctx context.Context, descriptor string) error {
	pathSegments := []string{pathApis, pathGraph, pathGroups, descriptor}
	_, err := c.vsspsClient.DeleteJSON(ctx, pathSegments, nil, networking.ApiVersion70Preview1)
	return err
}

func (c *Client) GetDescriptor(ctx context.Context, storageKey string) (*string, error) {
	pathSegments := []string{pathApis, pathGraph, pathDescriptors, storageKey}
	resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var descriptor *GraphDescriptorResult
	err = c.vsspsClient.ParseJSON(ctx, resp, &descriptor)
	return descriptor.Value, err
}

func (c *Client) GetGroup(ctx context.Context, descriptor string) (*GraphGroup, error) {
	pathSegments := []string{pathApis, pathGraph, pathGroups, descriptor}
	resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70Preview1)
	if err != nil {
		return nil, err
	}

	var group *GraphGroup
	err = c.vsspsClient.ParseJSON(ctx, resp, &group)
	return group, err
}

func (c *Client) GetGroups(ctx context.Context, projectId string, continuationToken string) (*[]GraphGroup, error) {
	descriptor, err := c.GetDescriptor(ctx, projectId)
	if err != nil {
		return nil, err
	}

	pathSegments := []string{pathApis, pathGraph, pathGroups}
	queryParams := url.Values{queryScopeDescriptor: []string{*descriptor}}
	if continuationToken != "" {
		queryParams.Add(queryContinuationToken, continuationToken)
	}

	var groups []GraphGroup
	hasMore := true
	for hasMore {
		resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, queryParams, networking.ApiVersion70Preview1)
		if err != nil {
			return nil, err
		}

		var collection *GraphGroupCollection
		err = c.vsspsClient.ParseJSON(ctx, resp, &collection)
		if err != nil {
			return nil, err
		}

		for _, group := range *collection.Value {
			groups = append(groups, group)
		}

		continuationToken = resp.Header.Get(networking.HeaderKeyContinuationToken)
		hasMore = continuationToken != ""
	}

	return &groups, nil
}

func (c *Client) GetUser(ctx context.Context, descriptor string) (*GraphUser, error) {
	pathSegments := []string{pathApis, pathGraph, pathUsers, descriptor}
	resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70Preview1)
	if err != nil {
		return nil, err
	}

	var user *GraphUser
	err = c.vsspsClient.ParseJSON(ctx, resp, &user)
	return user, err
}

func (c *Client) GetUsers(ctx context.Context, projectId string, continuationToken string) (*[]GraphUser, error) {
	descriptor, err := c.GetDescriptor(ctx, projectId)
	if err != nil {
		return nil, err
	}

	pathSegments := []string{pathApis, pathGraph, pathUsers}
	queryParams := url.Values{queryScopeDescriptor: []string{*descriptor}}
	if continuationToken != "" {
		queryParams.Add(queryContinuationToken, continuationToken)
	}

	var users []GraphUser
	hasMore := true
	for hasMore {
		resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, queryParams, networking.ApiVersion70Preview1)
		if err != nil {
			return nil, err
		}

		var collection *GraphUserCollection
		err = c.vsspsClient.ParseJSON(ctx, resp, &collection)
		if err != nil {
			return nil, err
		}

		for _, user := range *collection.Value {
			users = append(users, user)
		}

		continuationToken = resp.Header.Get(networking.HeaderKeyContinuationToken)
		hasMore = continuationToken != ""
	}

	return &users, nil
}

func (c *Client) UpdateGroup(ctx context.Context, descriptor string, displayName string, description string) (*GraphGroup, error) {
	pathSegments := []string{pathApis, pathGraph, pathGroups, descriptor}
	body := []core.JsonPatchOperation{
		{Op: "replace", Path: "/description", Value: description},
		{Op: "replace", Path: "/displayName", Value: displayName},
	}
	resp, err := c.vsspsClient.PatchJSONSpecialContentType(ctx, pathSegments, nil, body, networking.ApiVersion70Preview1)
	if err != nil {
		return nil, err
	}

	var group *GraphGroup
	err = c.vsspsClient.ParseJSON(ctx, resp, &group)
	return group, err
}
