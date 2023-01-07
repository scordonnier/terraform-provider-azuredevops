package graph

import (
	"context"
	"errors"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"net/url"
)

const (
	queryContinuationToken = "continuationToken"
	queryScopeDescriptor   = "scopeDescriptor"

	pathApis           = "_apis"
	pathDescriptors    = "descriptors"
	pathGraph          = "graph"
	pathGroups         = "groups"
	pathIdentityPicker = "identitypicker"
	pathIdentities     = "identities"
	pathUsers          = "users"
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
	descriptor, err := c.getDescriptor(ctx, projectId)
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
	descriptor, err := c.getDescriptor(ctx, projectId)
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
	descriptor, err := c.getDescriptor(ctx, projectId)
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

func (c *Client) SearchGroup(ctx context.Context, query string) (*GraphGroup, error) {
	response, err := c.getIdentityPickerResponse(ctx, query, "group")
	if err != nil {
		return nil, err
	}

	identity := (*(*response.Results)[0].Identities)[0]
	group := &GraphGroup{
		Descriptor:    identity.SubjectDescriptor,
		DisplayName:   identity.DisplayName,
		MailAddress:   identity.Mail,
		Origin:        identity.OriginDirectory,
		OriginId:      identity.OriginId,
		PrincipalName: identity.SamAccountName,
	}
	return group, nil
}

func (c *Client) SearchUser(ctx context.Context, query string) (*GraphUser, error) {
	response, err := c.getIdentityPickerResponse(ctx, query, "user")
	if err != nil {
		return nil, err
	}

	identity := (*(*response.Results)[0].Identities)[0]
	user := &GraphUser{
		Descriptor:    identity.SubjectDescriptor,
		DisplayName:   identity.DisplayName,
		MailAddress:   identity.Mail,
		Origin:        identity.OriginDirectory,
		OriginId:      identity.OriginId,
		PrincipalName: identity.SamAccountName,
	}
	return user, nil
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

// Private Methods

func (c *Client) getDescriptor(ctx context.Context, storageKey string) (*string, error) {
	pathSegments := []string{pathApis, pathGraph, pathDescriptors, storageKey}
	resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var descriptor *GraphDescriptorResult
	err = c.vsspsClient.ParseJSON(ctx, resp, &descriptor)
	return descriptor.Value, err
}

func (c *Client) getIdentityPickerResponse(ctx context.Context, query string, identityType string) (*IdentityPickerResponse, error) {
	pathSegments := []string{pathApis, pathIdentityPicker, pathIdentities}
	body := IdentityPickerRequest{
		IdentityTypes:   &[]string{identityType},
		OperationScopes: &[]string{"ims", "source"},
		Options: &IdentityPickerOptions{
			MaxResults: 1,
			MinResults: 1,
		},
		Properties: &[]string{
			"DisplayName",
			"Mail",
			"SamAccountName",
			"SubjectDescriptor",
		},
		Query: &query,
	}
	resp, err := c.vsspsClient.PostJSON(ctx, pathSegments, nil, body, networking.ApiVersion70Preview1)
	if err != nil {
		return nil, err
	}

	var response *IdentityPickerResponse
	err = c.vsspsClient.ParseJSON(ctx, resp, &response)
	if err != nil {
		return nil, err
	}

	result := (*response.Results)[0]
	if len(*result.Identities) == 0 || len(*result.Identities) > 1 {
		return nil, errors.New("identity not found or more than one identity found")
	}

	return response, nil
}
