package graph

import (
	"context"
	"errors"
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"net/url"
	"strings"
	"time"
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
	pathMemberships    = "memberships"
	pathUsers          = "users"
)

type Client struct {
	cache       *cache.Cache
	vsspsClient *networking.RestClient
}

func NewClient(vsspsClient *networking.RestClient) *Client {
	return &Client{
		cache:       cache.New(cache.NoExpiration, 0),
		vsspsClient: vsspsClient,
	}
}

func (c *Client) CreateGroup(ctx context.Context, projectId string, name string, description string) (*GraphGroup, error) {
	descriptor, err := c.getProjectDescriptor(ctx, projectId)
	if err != nil {
		return nil, err
	}

	body := GraphGroupVstsCreationContext{
		DisplayName: &name,
		Description: &description,
	}
	pathSegments := []string{pathApis, pathGraph, pathGroups}
	queryParams := url.Values{queryScopeDescriptor: []string{*descriptor}}
	group, _, err := networking.PostJSON[GraphGroup](c.vsspsClient, ctx, pathSegments, queryParams, body, networking.ApiVersion70Preview1)
	return group, err
}

func (c *Client) CreateGroupByOriginId(ctx context.Context, originId string) (*GraphGroup, error) {
	body := GraphGroupOriginIdCreationContext{
		OriginId: &originId,
	}
	pathSegments := []string{pathApis, pathGraph, pathGroups}
	group, _, err := networking.PostJSON[GraphGroup](c.vsspsClient, ctx, pathSegments, nil, body, networking.ApiVersion70Preview1)
	return group, err
}

func (c *Client) CreateGroupMemberships(ctx context.Context, projectId string, groupName string, members []string) (*[]GraphMembership, error) {
	groupDescriptor, err := c.getGroupDescriptor(ctx, projectId, groupName)
	if err != nil {
		return nil, err
	}

	memberDescriptors, err := c.getMemberDescriptors(ctx, members)
	if err != nil {
		return nil, err
	}

	for _, memberDescriptor := range *memberDescriptors {
		_, err := c.createGroupMembership(ctx, memberDescriptor, *groupDescriptor)
		if err != nil {
			return nil, err
		}
	}

	stateConf := c.membershipsStateChangeConf(ctx, projectId, groupName, memberDescriptors)
	memberships, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}

	return memberships.(*[]GraphMembership), nil
}

func (c *Client) DeleteGroup(ctx context.Context, descriptor string) error {
	pathSegments := []string{pathApis, pathGraph, pathGroups, descriptor}
	_, _, err := networking.DeleteJSON[networking.NoJSON](c.vsspsClient, ctx, pathSegments, nil, networking.ApiVersion70Preview1)
	return err
}

func (c *Client) DeleteGroupMemberships(ctx context.Context, projectId string, groupName string) error {
	memberships, err := c.GetGroupMemberships(ctx, projectId, groupName)
	if err != nil {
		return err
	}

	for _, membership := range *memberships {
		err := c.deleteGroupMembership(ctx, *membership.MemberDescriptor, *membership.ContainerDescriptor)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetGroup(ctx context.Context, descriptor string) (*GraphGroup, error) {
	pathSegments := []string{pathApis, pathGraph, pathGroups, descriptor}
	group, _, err := networking.GetJSON[GraphGroup](c.vsspsClient, ctx, pathSegments, nil, networking.ApiVersion70Preview1)
	return group, err
}

func (c *Client) GetGroupMemberships(ctx context.Context, projectId string, name string) (*[]GraphMembership, error) {
	groupDescriptor, err := c.getGroupDescriptor(ctx, projectId, name)
	if err != nil {
		return nil, err
	}

	pathSegments := []string{pathApis, pathGraph, pathMemberships, *groupDescriptor}
	queryParams := url.Values{"direction": []string{"down"}}
	memberships, _, err := networking.GetJSON[GraphMembershipCollection](c.vsspsClient, ctx, pathSegments, queryParams, networking.ApiVersion70Preview1)
	if err != nil {
		return nil, err
	}

	return memberships.Value, nil
}

func (c *Client) GetGroups(ctx context.Context, projectId string, continuationToken string) (*[]GraphGroup, error) {
	projectDescriptor, err := c.getProjectDescriptor(ctx, projectId)
	if err != nil {
		return nil, err
	}

	pathSegments := []string{pathApis, pathGraph, pathGroups}
	queryParams := url.Values{queryScopeDescriptor: []string{*projectDescriptor}}
	if continuationToken != "" {
		queryParams.Add(queryContinuationToken, continuationToken)
	}

	var groups []GraphGroup
	hasMore := true
	for hasMore {
		collection, resp, err := networking.GetJSON[GraphGroupCollection](c.vsspsClient, ctx, pathSegments, queryParams, networking.ApiVersion70Preview1)
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

func (c *Client) GetIdentityPickerIdentity(ctx context.Context, query string) (*IdentityPickerIdentity, error) {
	cacheKey := utils.GetCacheKey("IdentityPicker", query)
	if i, ok := c.cache.Get(cacheKey); ok {
		return i.(*IdentityPickerIdentity), nil
	}

	pathSegments := []string{pathApis, pathIdentityPicker, pathIdentities}
	body := IdentityPickerRequest{
		IdentityTypes:   &[]string{"user", "group"},
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
	response, _, err := networking.PostJSON[IdentityPickerResponse](c.vsspsClient, ctx, pathSegments, nil, body, networking.ApiVersion70Preview1)
	if err != nil {
		return nil, err
	}

	result := (*response.Results)[0]
	if len(*result.Identities) == 0 || len(*result.Identities) > 1 {
		return nil, errors.New(fmt.Sprintf("Identity not found or more than one identity found : '%s'", query))
	}

	identity := (*result.Identities)[0]
	c.cache.Set(cacheKey, &identity, cache.NoExpiration)
	return &identity, nil
}

func (c *Client) GetUser(ctx context.Context, descriptor string) (*GraphUser, error) {
	pathSegments := []string{pathApis, pathGraph, pathUsers, descriptor}
	user, _, err := networking.GetJSON[GraphUser](c.vsspsClient, ctx, pathSegments, nil, networking.ApiVersion70Preview1)
	return user, err
}

func (c *Client) CreateUserByOriginId(ctx context.Context, originId string) (*GraphUser, error) {
	body := GraphUserOriginIdCreationContext{
		OriginId: &originId,
	}
	pathSegments := []string{pathApis, pathGraph, pathUsers}
	user, _, err := networking.PostJSON[GraphUser](c.vsspsClient, ctx, pathSegments, nil, body, networking.ApiVersion70Preview1)
	return user, err
}

func (c *Client) GetUsers(ctx context.Context, projectId string, continuationToken string) (*[]GraphUser, error) {
	projectDescriptor, err := c.getProjectDescriptor(ctx, projectId)
	if err != nil {
		return nil, err
	}

	pathSegments := []string{pathApis, pathGraph, pathUsers}
	queryParams := url.Values{queryScopeDescriptor: []string{*projectDescriptor}}
	if continuationToken != "" {
		queryParams.Add(queryContinuationToken, continuationToken)
	}

	var users []GraphUser
	hasMore := true
	for hasMore {
		collection, resp, err := networking.GetJSON[GraphUserCollection](c.vsspsClient, ctx, pathSegments, queryParams, networking.ApiVersion70Preview1)
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
	group, _, err := networking.PatchJSONSpecialContentType[GraphGroup](c.vsspsClient, ctx, pathSegments, nil, body, networking.ApiVersion70Preview1)
	return group, err
}

func (c *Client) UpdateGroupMemberships(ctx context.Context, projectId string, groupName string, members []string) (*[]GraphMembership, error) {
	groupDescriptor, err := c.getGroupDescriptor(ctx, projectId, groupName)
	if err != nil {
		return nil, err
	}

	membershipsDescriptors, err := c.GetGroupMemberships(ctx, projectId, groupName)
	if err != nil {
		return nil, err
	}

	var currentDescriptors []string
	for _, membership := range *membershipsDescriptors {
		currentDescriptors = append(currentDescriptors, *membership.MemberDescriptor)
	}

	membersDescriptors, err := c.getMemberDescriptors(ctx, members)
	if err != nil {
		return nil, err
	}

	membersToDelete := utils.Difference(&currentDescriptors, membersDescriptors)
	for _, m := range *membersToDelete {
		err := c.deleteGroupMembership(ctx, m, *groupDescriptor)
		if err != nil {
			return nil, err
		}
	}

	membersToAdd := utils.Difference(membersDescriptors, &currentDescriptors)
	for _, m := range *membersToAdd {
		_, err := c.createGroupMembership(ctx, m, *groupDescriptor)
		if err != nil {
			return nil, err
		}
	}

	stateConf := c.membershipsStateChangeConf(ctx, projectId, groupName, membersDescriptors)
	memberships, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}

	return memberships.(*[]GraphMembership), nil
}

// Private Methods

func (c *Client) createGroupMembership(ctx context.Context, memberDescriptor string, containerDescriptor string) (*GraphMembership, error) {
	pathSegments := []string{pathApis, pathGraph, pathMemberships, memberDescriptor, containerDescriptor}
	membership, _, err := networking.PutJSON[GraphMembership](c.vsspsClient, ctx, pathSegments, nil, nil, networking.ApiVersion70Preview1)
	return membership, err
}

func (c *Client) deleteGroupMembership(ctx context.Context, memberDescriptor string, containerDescriptor string) error {
	pathSegments := []string{pathApis, pathGraph, pathMemberships, memberDescriptor, containerDescriptor}
	_, _, err := networking.DeleteJSON[networking.NoJSON](c.vsspsClient, ctx, pathSegments, nil, networking.ApiVersion70Preview1)
	return err
}

func (c *Client) getGroupDescriptor(ctx context.Context, projectId string, name string) (*string, error) {
	cacheKey := utils.GetCacheKey("Group", projectId, name)
	if p, ok := c.cache.Get(cacheKey); ok {
		return p.(*string), nil
	}

	groups, err := c.GetGroups(ctx, projectId, "")
	if err != nil {
		return nil, err
	}

	var descriptor *string
	for _, group := range *groups {
		if strings.EqualFold(*group.DisplayName, name) || strings.EqualFold(*group.PrincipalName, name) {
			descriptor = group.Descriptor
			break
		}
	}

	if descriptor == nil {
		return nil, errors.New(fmt.Sprintf("Group with name '%s' in project '%s' not found", name, projectId))
	}

	c.cache.Set(cacheKey, descriptor, cache.NoExpiration)
	return descriptor, nil
}

func (c *Client) getMemberDescriptors(ctx context.Context, members []string) (*[]string, error) {
	var memberDescriptors []string
	for _, member := range members {
		identity, err := c.GetIdentityPickerIdentity(ctx, member)
		if err != nil {
			return nil, err
		}

		memberDescriptor := identity.SubjectDescriptor
		if memberDescriptor == nil {
			switch strings.ToLower(*identity.EntityType) {
			case "group":
				group, err := c.CreateGroupByOriginId(ctx, *identity.OriginId)
				if err != nil {
					return nil, err
				}

				if group.Descriptor == nil {
					return nil, errors.New(fmt.Sprintf("Unable to find identity descriptor for '%s'", member))
				}

				memberDescriptor = group.Descriptor
			case "user":
				user, err := c.CreateUserByOriginId(ctx, *identity.OriginId)
				if err != nil {
					return nil, err
				}

				if user.Descriptor == nil {
					return nil, errors.New(fmt.Sprintf("Unable to find identity descriptor for '%s'", member))
				}

				memberDescriptor = user.Descriptor
			default:
				return nil, errors.New(fmt.Sprintf("Unknown entity type '%s'", *identity.EntityType))
			}
		}

		memberDescriptors = append(memberDescriptors, *memberDescriptor)
	}

	return &memberDescriptors, nil
}

func (c *Client) getMembershipDescriptors(memberships *[]GraphMembership) *[]string {
	var descriptors []string
	for _, membership := range *memberships {
		descriptors = append(descriptors, *membership.MemberDescriptor)
	}
	return &descriptors
}

func (c *Client) getProjectDescriptor(ctx context.Context, projectId string) (*string, error) {
	cacheKey := utils.GetCacheKey("Project", projectId)
	if p, ok := c.cache.Get(cacheKey); ok {
		return p.(*string), nil
	}

	pathSegments := []string{pathApis, pathGraph, pathDescriptors, projectId}
	result, _, err := networking.GetJSON[GraphDescriptorResult](c.vsspsClient, ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, result.Value, cache.NoExpiration)
	return result.Value, nil
}

func (c *Client) membershipsStateChangeConf(ctx context.Context, projectId string, groupName string, members *[]string) *utils.StateChangeConf {
	return &utils.StateChangeConf{
		Delay:      2 * time.Second,
		MinTimeout: 5 * time.Second,
		Timeout:    2 * time.Minute,
		Pending:    []string{"Waiting"},
		Target:     []string{"Synced"},
		Refresh: func() (interface{}, string, error) {
			state := "Waiting"
			memberships, err := c.GetGroupMemberships(ctx, projectId, groupName)
			if err != nil {
				return nil, "", err
			}

			descriptors := c.getMembershipDescriptors(memberships)
			toAdd := utils.Difference(descriptors, members)
			toDelete := utils.Difference(members, descriptors)

			if len(*toAdd) == 0 && len(*toDelete) == 0 {
				state = "Synced"
			}

			return memberships, state, nil
		},
	}
}
