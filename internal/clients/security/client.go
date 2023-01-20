package security

import (
	"context"
	"errors"
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"net/url"
	"strconv"
	"strings"
)

const (
	pathAccessControlEntries = "accesscontrolentries"
	pathAccessControlLists   = "accesscontrollists"
	pathApis                 = "_apis"
	pathIdentities           = "identities"
	pathSecurityNamespaces   = "securitynamespaces"
)

type Client struct {
	azdoClient  *networking.RestClient
	cache       *cache.Cache
	vsspsClient *networking.RestClient
}

func NewClient(restClient *networking.RestClient, vsspsClient *networking.RestClient) *Client {
	return &Client{
		azdoClient:  restClient,
		cache:       cache.New(cache.NoExpiration, 0),
		vsspsClient: vsspsClient,
	}
}

func (c *Client) GetAccessControlLists(ctx context.Context, namespaceId string, token string) (*AccessControlListCollection, error) {
	pathSegments := []string{pathApis, pathAccessControlLists, namespaceId}
	queryParams := url.Values{"token": []string{token}, "includeExtendedInfo": []string{"true"}}
	acls, _, err := networking.GetJSON[AccessControlListCollection](c.azdoClient, ctx, pathSegments, queryParams, networking.ApiVersion70)
	return acls, err
}

func (c *Client) GetClassificationNodeToken(iterationId string) string {
	return fmt.Sprintf("vstfs:///Classification/Node/%s", iterationId)
}

func (c *Client) GetEnvironmentToken(projectId string, environmentId int) string {
	token := "Environments/" + projectId
	if environmentId > 0 {
		token += "/" + strconv.Itoa(environmentId)
	}
	return token
}

func (c *Client) GetIdentityByDescriptor(ctx context.Context, descriptor string) (*Identity, error) {
	queryParams := url.Values{"descriptors": []string{descriptor}, "queryMembership": []string{"none"}}
	return c.getIdentity(ctx, queryParams, descriptor)
}

func (c *Client) GetIdentityBySubjectDescriptor(ctx context.Context, subjectDescriptor string) (*Identity, error) {
	queryParams := url.Values{"subjectDescriptors": []string{subjectDescriptor}, "queryMembership": []string{"none"}}
	return c.getIdentity(ctx, queryParams, subjectDescriptor)
}

func (c *Client) GetPipelineToken(projectId string, pipelineId int) string {
	token := projectId
	if pipelineId > 0 {
		token += "/" + strconv.Itoa(pipelineId)
	}
	return token
}

func (c *Client) GetProjectToken(projectId string) string {
	return fmt.Sprintf("$PROJECT:vstfs:///Classification/TeamProject/%s", projectId)
}

func (c *Client) GetSecurityNamespaces(ctx context.Context) (*SecurityNamespacesCollection, error) {
	cacheKey := utils.GetCacheKey("Namespaces")
	if n, ok := c.cache.Get(cacheKey); ok {
		return n.(*SecurityNamespacesCollection), nil
	}

	pathSegments := []string{pathApis, pathSecurityNamespaces}
	namespaces, _, err := networking.GetJSON[SecurityNamespacesCollection](c.azdoClient, ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, namespaces, cache.NoExpiration)
	return namespaces, err
}

func (c *Client) GetServiceEndpointToken(projectId string, serviceEndpointId string) string {
	token := "endpoints/" + projectId
	if serviceEndpointId != "" {
		token += "/" + serviceEndpointId
	}
	return token
}

func (c *Client) RemoveAccessControlEntries(ctx context.Context, namespaceId string, token string, descriptors []string) error {
	pathSegments := []string{pathApis, pathAccessControlEntries, namespaceId}
	queryParams := url.Values{"token": []string{token}, "descriptors": []string{strings.Join(descriptors, ",")}}
	_, _, err := networking.DeleteJSON[networking.NoJSON](c.azdoClient, ctx, pathSegments, queryParams, networking.ApiVersion70)
	return err
}

func (c *Client) RemoveAccessControlLists(ctx context.Context, namespaceId string, token string) error {
	pathSegments := []string{pathApis, pathAccessControlLists, namespaceId}
	queryParams := url.Values{"tokens": []string{token}}
	_, _, err := networking.DeleteJSON[networking.NoJSON](c.azdoClient, ctx, pathSegments, queryParams, networking.ApiVersion70)
	return err
}

func (c *Client) SetAccessControlEntries(ctx context.Context, namespaceId string, token string, accessControlEntries *[]AccessControlEntry) error {
	pathSegments := []string{pathApis, pathAccessControlEntries, namespaceId}
	body := SetAccessControlEntriesArgs{
		AccessControlEntries: accessControlEntries,
		Merge:                utils.Bool(false),
		Token:                &token,
	}
	_, _, err := networking.PostJSON[AccessControlEntryCollection](c.azdoClient, ctx, pathSegments, nil, body, networking.ApiVersion70)
	return err
}

func (c *Client) SetAccessControlLists(ctx context.Context, namespaceId string, accessControlList *[]AccessControlList) error {
	pathSegments := []string{pathApis, pathAccessControlLists, namespaceId}
	body := AccessControlListCollection{
		Value: accessControlList,
	}
	_, _, err := networking.PostJSON[networking.NoJSON](c.azdoClient, ctx, pathSegments, nil, body, networking.ApiVersion70)
	return err
}

// Private Methods

func (c *Client) getIdentity(ctx context.Context, queryParams url.Values, descriptor string) (*Identity, error) {
	cacheKey := utils.GetCacheKey("Identity", descriptor)
	if i, ok := c.cache.Get(cacheKey); ok {
		return i.(*Identity), nil
	}

	pathSegments := []string{pathApis, pathIdentities}
	identityResult, _, err := networking.GetJSON[IdentityCollection](c.vsspsClient, ctx, pathSegments, queryParams, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	if len(*identityResult.Value) == 0 || len(*identityResult.Value) > 1 {
		return nil, errors.New(fmt.Sprintf("Identity not found or too many identities '%s'", descriptor))
	}

	identity := (*identityResult.Value)[0]
	c.cache.Set(cacheKey, &identity, cache.NoExpiration)
	return &identity, err
}
