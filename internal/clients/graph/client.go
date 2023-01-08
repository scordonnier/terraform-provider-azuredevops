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
	resp, err := c.vsspsClient.PostJSON(ctx, pathSegments, queryParams, body, networking.ApiVersion70Preview1)
	if err != nil {
		return nil, err
	}

	var group *GraphGroup
	err = c.vsspsClient.ParseJSON(ctx, resp, &group)
	return group, err
}

func (c *Client) CreateGroupMemberships(ctx context.Context, projectId string, groupName string, members []string) (*[]GraphMembership, error) {
	groupDescriptor, err := c.getGroupDescriptor(ctx, projectId, groupName)
	if err != nil {
		return nil, err
	}

	memberDescriptors, err := c.getMemberDescriptors(ctx, projectId, groupName, members)
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
	_, err := c.vsspsClient.DeleteJSON(ctx, pathSegments, nil, networking.ApiVersion70Preview1)
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
	resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70Preview1)
	if err != nil {
		return nil, err
	}

	var group *GraphGroup
	err = c.vsspsClient.ParseJSON(ctx, resp, &group)
	return group, err
}

func (c *Client) GetGroupMemberships(ctx context.Context, projectId string, name string) (*[]GraphMembership, error) {
	groupDescriptor, err := c.getGroupDescriptor(ctx, projectId, name)
	if err != nil {
		return nil, err
	}

	pathSegments := []string{pathApis, pathGraph, pathMemberships, *groupDescriptor}
	queryParams := url.Values{"direction": []string{"down"}}
	resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, queryParams, networking.ApiVersion70Preview1)
	if err != nil {
		return nil, err
	}

	var memberships *GraphMembershipCollection
	err = c.vsspsClient.ParseJSON(ctx, resp, &memberships)
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
	identity, err := c.getIdentityPickerIdentity(ctx, query, []string{"group"})
	if err != nil {
		return nil, err
	}

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
	identity, err := c.getIdentityPickerIdentity(ctx, query, []string{"user"})
	if err != nil {
		return nil, err
	}

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

	membersDescriptors, err := c.getMemberDescriptors(ctx, projectId, groupName, members)
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

func (c *Client) createGroupByOriginId(ctx context.Context, projectId string, groupName string, originId string) (*GraphGroup, error) {
	groupDescriptor, err := c.getGroupDescriptor(ctx, projectId, groupName)
	if err != nil {
		return nil, err
	}

	body := GraphGroupOriginIdCreationContext{
		OriginId: &originId,
	}
	pathSegments := []string{pathApis, pathGraph, pathGroups}
	queryParams := url.Values{"groupDescriptors": []string{*groupDescriptor}}
	resp, err := c.vsspsClient.PostJSON(ctx, pathSegments, queryParams, body, networking.ApiVersion70Preview1)
	if err != nil {
		return nil, err
	}

	var group *GraphGroup
	err = c.vsspsClient.ParseJSON(ctx, resp, &group)
	return group, err
}

func (c *Client) createGroupMembership(ctx context.Context, memberDescriptor string, containerDescriptor string) (*GraphMembership, error) {
	pathSegments := []string{pathApis, pathGraph, pathMemberships, memberDescriptor, containerDescriptor}
	resp, err := c.vsspsClient.PutJSON(ctx, pathSegments, nil, nil, networking.ApiVersion70Preview1)
	if err != nil {
		return nil, err
	}

	var membership *GraphMembership
	err = c.vsspsClient.ParseJSON(ctx, resp, &membership)
	return membership, err
}

func (c *Client) deleteGroupMembership(ctx context.Context, memberDescriptor string, containerDescriptor string) error {
	pathSegments := []string{pathApis, pathGraph, pathMemberships, memberDescriptor, containerDescriptor}
	_, err := c.vsspsClient.DeleteJSON(ctx, pathSegments, nil, networking.ApiVersion70Preview1)
	return err
}

func getCacheKey(params ...string) string {
	return strings.Join(params, "***")
}

func (c *Client) getGroupDescriptor(ctx context.Context, projectId string, name string) (*string, error) {
	cacheKey := getCacheKey("Group", projectId, name)
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

func (c *Client) getIdentityPickerIdentity(ctx context.Context, query string, identityTypes []string) (*IdentityPickerIdentity, error) {
	cacheKey := getCacheKey("IdentityPicker", query)
	if p, ok := c.cache.Get(cacheKey); ok {
		return p.(*IdentityPickerIdentity), nil
	}

	pathSegments := []string{pathApis, pathIdentityPicker, pathIdentities}
	body := IdentityPickerRequest{
		IdentityTypes:   &identityTypes,
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

	identity := (*result.Identities)[0]
	c.cache.Set(cacheKey, identity, cache.NoExpiration)
	return &identity, nil
}

func (c *Client) getMemberDescriptors(ctx context.Context, projectId string, groupName string, members []string) (*[]string, error) {
	var memberDescriptors []string
	for _, member := range members {
		identity, err := c.getIdentityPickerIdentity(ctx, member, []string{"user", "group"})
		if err != nil {
			return nil, err
		}

		memberDescriptor := identity.SubjectDescriptor
		if memberDescriptor == nil {
			group, err := c.createGroupByOriginId(ctx, projectId, groupName, *identity.OriginId)
			if err != nil {
				return nil, err
			}

			if group.Descriptor == nil {
				return nil, errors.New("unable to determine identity descriptor")
			}

			memberDescriptor = group.Descriptor
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
	cacheKey := getCacheKey("Project", projectId)
	if p, ok := c.cache.Get(cacheKey); ok {
		return p.(*string), nil
	}

	pathSegments := []string{pathApis, pathGraph, pathDescriptors, projectId}
	resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var result *GraphDescriptorResult
	err = c.vsspsClient.ParseJSON(ctx, resp, &result)
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
