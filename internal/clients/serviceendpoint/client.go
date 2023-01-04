package serviceendpoint

import (
	"context"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"net/url"
)

const (
	pathApis            = "_apis"
	pathEndpoints       = "endpoints"
	pathServiceEndpoint = "serviceendpoint"
)

type Client struct {
	restClient *networking.RestClient
}

func NewClient(restClient *networking.RestClient) *Client {
	return &Client{
		restClient: restClient,
	}
}

func (c *Client) CreateServiceEndpoint(ctx context.Context, args *CreateOrUpdateServiceEndpointArgs, projectId string) (*ServiceEndpoint, error) {
	pathSegments := []string{projectId, pathApis, pathServiceEndpoint, pathEndpoints}
	serviceEndpoint := c.createOrUpdateServiceEndpoint(ctx, nil, args, projectId)
	resp, err := c.restClient.PostJSON(ctx, pathSegments, nil, serviceEndpoint, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	err = c.restClient.ParseJSON(ctx, resp, &serviceEndpoint)
	return serviceEndpoint, err
}

func (c *Client) DeleteServiceEndpoint(ctx context.Context, id string, projectIds []string) error {
	pathSegments := []string{pathApis, pathServiceEndpoint, pathEndpoints, id}
	queryParams := url.Values{"projectIds": projectIds}
	_, err := c.restClient.DeleteJSON(ctx, pathSegments, queryParams, networking.ApiVersion70)
	return err
}

func (c *Client) GetServiceEndpoint(ctx context.Context, id string, projectId string) (*ServiceEndpoint, error) {
	pathSegments := []string{projectId, pathApis, pathServiceEndpoint, pathEndpoints, id}
	resp, err := c.restClient.GetJSON(ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	var serviceEndpoint *ServiceEndpoint
	err = c.restClient.ParseJSON(ctx, resp, &serviceEndpoint)
	return serviceEndpoint, err
}

func (c *Client) ShareServiceEndpoint(ctx context.Context, id string, name string, description string, projectId string, projectIds []string) error {
	var serviceEndpointProjectReferences []ServiceEndpointProjectReference
	for _, projectId := range projectIds {
		serviceEndpointProjectReferences = append(serviceEndpointProjectReferences, ServiceEndpointProjectReference{
			Description: &description,
			Name:        &name,
			ProjectReference: &core.ProjectReference{
				Id: utils.UUID(projectId),
			},
		})
	}

	pathSegments := []string{projectId, pathApis, pathServiceEndpoint, pathEndpoints, id}
	_, err := c.restClient.PatchJSON(ctx, pathSegments, nil, serviceEndpointProjectReferences, networking.ApiVersion70)
	return err
}

func (c *Client) UpdateServiceEndpoint(ctx context.Context, id string, args *CreateOrUpdateServiceEndpointArgs, projectId string) (*ServiceEndpoint, error) {
	pathSegments := []string{pathApis, pathServiceEndpoint, pathEndpoints, id}
	serviceEndpoint, err := c.GetServiceEndpoint(ctx, id, projectId)
	if err != nil {
		return nil, err
	}

	updatedServiceEndpoint := c.createOrUpdateServiceEndpoint(ctx, serviceEndpoint, args, projectId)
	resp, err := c.restClient.PutJSON(ctx, pathSegments, nil, updatedServiceEndpoint, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}

	err = c.restClient.ParseJSON(ctx, resp, &updatedServiceEndpoint)
	return serviceEndpoint, err
}

func (c *Client) createOrUpdateServiceEndpoint(_ context.Context, serviceEndpoint *ServiceEndpoint, args *CreateOrUpdateServiceEndpointArgs, projectId string) *ServiceEndpoint {
	if serviceEndpoint != nil {
		serviceEndpoint.Authorization = getServiceEndpointAuthorization(args)
		serviceEndpoint.Data = getServiceEndpointData(args)
		serviceEndpoint.Description = &args.Description
		serviceEndpoint.Name = &args.Name
		for _, projectReference := range *serviceEndpoint.ServiceEndpointProjectReferences {
			*projectReference.Description = args.Description
			*projectReference.Name = args.Name
		}
		return serviceEndpoint
	}

	return &ServiceEndpoint{
		Authorization: getServiceEndpointAuthorization(args),
		Data:          getServiceEndpointData(args),
		Description:   &args.Description,
		Name:          &args.Name,
		Owner:         utils.String(ServiceEndpointOwnerLibrary),
		ServiceEndpointProjectReferences: &[]ServiceEndpointProjectReference{
			{
				Description: &args.Description,
				Name:        &args.Name,
				ProjectReference: &core.ProjectReference{
					Id: utils.UUID(projectId),
				},
			},
		},
		Type: utils.String(args.Type),
		Url:  getServiceEndpointUrl(args),
	}
}

func getServiceEndpointAuthorization(args *CreateOrUpdateServiceEndpointArgs) *EndpointAuthorization {
	switch args.Type {
	case ServiceEndpointTypeAzureRm:
		return &EndpointAuthorization{
			Parameters: &map[string]string{
				ServiceEndpointAuthorizationParamsAuthenticationType:  "spnKey",
				ServiceEndpointAuthorizationParamsServicePrincipalId:  args.ServicePrincipalId,
				ServiceEndpointAuthorizationParamsServicePrincipalKey: args.ServicePrincipalKey,
				ServiceEndpointAuthorizationParamsServiceTenantId:     args.TenantId,
			},
			Scheme: utils.String(ServiceEndpointAuthorizationSchemeServicePrincipal),
		}
	case ServiceEndpointTypeBitbucket:
		return &EndpointAuthorization{
			Parameters: &map[string]string{
				ServiceEndpointAuthorizationParamsPassword: args.Password,
				ServiceEndpointAuthorizationParamsUserName: args.UserName,
			},
			Scheme: utils.String(ServiceEndpointAuthorizationSchemeUsernamePassword),
		}
	default:
		return nil
	}
}

func getServiceEndpointData(args *CreateOrUpdateServiceEndpointArgs) *map[string]string {
	switch args.Type {
	case ServiceEndpointTypeAzureRm:
		return &map[string]string{
			ServiceEndpointDataCreationMode:     "Manual",
			ServiceEndpointDataEnvironment:      "AzureCloud",
			ServiceEndpointDataScopeLevel:       "Subscription",
			ServiceEndpointDataSubscriptionId:   args.SubscriptionId,
			ServiceEndpointDataSubscriptionName: args.SubscriptionName,
		}
	default:
		return nil
	}
}

func getServiceEndpointUrl(args *CreateOrUpdateServiceEndpointArgs) *string {
	switch args.Type {
	case ServiceEndpointTypeAzureRm:
		return utils.String("https://management.azure.com/")
	case ServiceEndpointTypeBitbucket:
		return utils.String("https://api.bitbucket.org/")
	default:
		return nil
	}
}
