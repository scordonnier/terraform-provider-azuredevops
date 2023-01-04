package security

import (
	"context"
	"errors"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"net/url"
	"strconv"
)

const (
	pathAccessControlLists = "accesscontrollists"
	pathApis               = "_apis"
	pathIdentities         = "identities"
	pathSecurityNamespaces = "securitynamespaces"
)

type Client struct {
	azdoClient  *networking.RestClient
	vsspsClient *networking.RestClient
}

func NewClient(restClient *networking.RestClient, vsspsClient *networking.RestClient) *Client {
	return &Client{
		azdoClient:  restClient,
		vsspsClient: vsspsClient,
	}
}

func (c *Client) GetAccessControlLists(ctx context.Context, namespaceId string, token string) (*AccessControlListCollection, error) {
	pathSegments := []string{pathApis, pathAccessControlLists, namespaceId}
	queryParams := url.Values{"token": []string{token}, "includeExtendedInfo": []string{"true"}}
	resp, err := c.azdoClient.GetJSON(ctx, pathSegments, queryParams, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var accessControlLists *AccessControlListCollection
	err = c.azdoClient.ParseJSON(ctx, resp, &accessControlLists)
	return accessControlLists, err
}

func (c *Client) GetIdentity(ctx context.Context, name string, identityType string) (*Identity, error) {
	pathSegments := []string{pathApis, pathIdentities}
	var searchFilter string
	switch identityType {
	case "group":
		searchFilter = "DisplayName"
	case "user":
		searchFilter = "MailAddress"
	}
	queryParams := url.Values{"searchFilter": []string{searchFilter}, "filterValue": []string{name}, "queryMembership": []string{"none"}}
	resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, queryParams, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var identityResult *IdentityReadResult
	err = c.azdoClient.ParseJSON(ctx, resp, &identityResult)
	if err != nil {
		return nil, err
	}

	if len(*identityResult.Value) == 0 || len(*identityResult.Value) > 1 {
		return nil, errors.New("identity not found or too many identities matching the criteria")
	}

	identity := (*identityResult.Value)[0]
	return &identity, err
}

func (c *Client) GetIdentityByDescriptor(ctx context.Context, descriptor string) (*Identity, error) {
	pathSegments := []string{pathApis, pathIdentities}
	queryParams := url.Values{"descriptors": []string{descriptor}, "queryMembership": []string{"none"}}
	resp, err := c.vsspsClient.GetJSON(ctx, pathSegments, queryParams, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var identityResult *IdentityReadResult
	err = c.azdoClient.ParseJSON(ctx, resp, &identityResult)
	if err != nil {
		return nil, err
	}

	if len(*identityResult.Value) == 0 || len(*identityResult.Value) > 1 {
		return nil, errors.New("identity not found or too many identities matching the criteria")
	}

	identity := (*identityResult.Value)[0]
	return &identity, err
}

func (c *Client) GetEnvironmentToken(environmentId int, projectId string) string {
	token := "Environments/" + projectId
	if environmentId > 0 {
		token += "/" + strconv.Itoa(environmentId)
	}
	return token
}

func (c *Client) GetSecurityNamespaces(ctx context.Context) (*SecurityNamespacesResult, error) {
	pathSegments := []string{pathApis, pathSecurityNamespaces}
	resp, err := c.azdoClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var securityNamespaces *SecurityNamespacesResult
	err = c.azdoClient.ParseJSON(ctx, resp, &securityNamespaces)
	return securityNamespaces, err
}

func (c *Client) RemoveAccessControlLists(ctx context.Context, namespaceId string, token string) error {
	pathSegments := []string{pathApis, pathAccessControlLists, namespaceId}
	queryParams := url.Values{"tokens": []string{token}}
	_, err := c.azdoClient.DeleteJSON(ctx, pathSegments, queryParams, networking.ApiVersion70)
	return err
}

func (c *Client) SetAccessControlLists(ctx context.Context, namespaceId string, accessControlList *[]AccessControlList) error {
	pathSegments := []string{pathApis, pathAccessControlLists, namespaceId}
	body := AccessControlListCollection{
		Value: accessControlList,
	}
	_, err := c.azdoClient.PostJSON(ctx, pathSegments, nil, body, networking.ApiVersion70)
	return err
}
