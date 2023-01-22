package serviceendpoints

import (
	"context"
	"fmt"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/networking"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"net/url"
	"strings"
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
	serviceEndpoint, _, err := networking.PostJSON[ServiceEndpoint](c.restClient, ctx, pathSegments, nil, serviceEndpoint, networking.ApiVersion70)
	return serviceEndpoint, err
}

func (c *Client) DeleteServiceEndpoint(ctx context.Context, id string, projectIds []string) error {
	pathSegments := []string{pathApis, pathServiceEndpoint, pathEndpoints, id}
	queryParams := url.Values{"projectIds": []string{strings.Join(projectIds, ",")}}
	_, _, err := networking.DeleteJSON[networking.NoJSON](c.restClient, ctx, pathSegments, queryParams, networking.ApiVersion70)
	return err
}

func (c *Client) GetServiceEndpoint(ctx context.Context, id string, projectId string) (*ServiceEndpoint, error) {
	pathSegments := []string{projectId, pathApis, pathServiceEndpoint, pathEndpoints, id}
	serviceEndpoint, _, err := networking.GetJSON[ServiceEndpoint](c.restClient, ctx, pathSegments, nil, networking.ApiVersion70)
	if err != nil {
		return nil, err
	}
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
	_, _, err := networking.PatchJSON[networking.NoJSON](c.restClient, ctx, pathSegments, nil, serviceEndpointProjectReferences, networking.ApiVersion70)
	return err
}

func (c *Client) UpdateServiceEndpoint(ctx context.Context, id string, args *CreateOrUpdateServiceEndpointArgs, projectId string) (*ServiceEndpoint, error) {
	pathSegments := []string{pathApis, pathServiceEndpoint, pathEndpoints, id}
	serviceEndpoint, err := c.GetServiceEndpoint(ctx, id, projectId)
	if err != nil {
		return nil, err
	}

	updatedServiceEndpoint := c.createOrUpdateServiceEndpoint(ctx, serviceEndpoint, args, projectId)
	updatedServiceEndpoint, _, err = networking.PutJSON[ServiceEndpoint](c.restClient, ctx, pathSegments, nil, updatedServiceEndpoint, networking.ApiVersion70)
	return updatedServiceEndpoint, err
}

// Private Methods

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
				ServiceEndpointAuthorizationParamsUserName: args.Username,
			},
			Scheme: utils.String(ServiceEndpointAuthorizationSchemeUsernamePassword),
		}
	case ServiceEndpointTypeGitHub:
		return &EndpointAuthorization{
			Parameters: &map[string]string{
				ServiceEndpointAuthorizationParamsAccessToken: args.Token,
			},
			Scheme: utils.String(ServiceEndpointAuthorizationSchemeToken),
		}
	case ServiceEndpointTypeJFrogArtifactory,
		ServiceEndpointTypeJFrogDistribution,
		ServiceEndpointTypeJFrogPlatform,
		ServiceEndpointTypeJFrogXray:
		if args.Token != "" {
			return &EndpointAuthorization{
				Parameters: &map[string]string{
					ServiceEndpointAuthorizationParamsApiToken: args.Token,
				},
				Scheme: utils.String(ServiceEndpointAuthorizationSchemeToken),
			}
		} else {
			return &EndpointAuthorization{
				Parameters: &map[string]string{
					ServiceEndpointAuthorizationParamsPassword: args.Password,
					ServiceEndpointAuthorizationParamsUserName: args.Username,
				},
				Scheme: utils.String(ServiceEndpointAuthorizationSchemeUsernamePassword),
			}
		}
	case ServiceEndpointTypekubernetes:
		return &EndpointAuthorization{
			Parameters: &map[string]string{
				ServiceEndpointAuthorizationParamsClusterContext: args.ClusterContext,
				ServiceEndpointAuthorizationParamsKubeconfig:     args.Kubeconfig,
			},
			Scheme: utils.String(ServiceEndpointAuthorizationSchemeKubernetes),
		}
	case ServiceEndpointTypeVsAppCenter:
		return &EndpointAuthorization{
			Parameters: &map[string]string{
				ServiceEndpointAuthorizationParamsApiToken: args.Token,
			},
			Scheme: utils.String(ServiceEndpointAuthorizationSchemeToken),
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
	case ServiceEndpointTypekubernetes:
		return &map[string]string{
			ServiceEndpointDataAuthorizationType:           "Kubeconfig",
			ServiceEndpointDataAcceptUntrustedCertificates: fmt.Sprintf("%v", args.AcceptUntrustedCertificates),
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
	case ServiceEndpointTypeGitHub:
		return utils.String("https://github.com/")
	case ServiceEndpointTypeJFrogArtifactory:
		return utils.String(strings.TrimSuffix(args.Url, "/") + "/artifactory")
	case ServiceEndpointTypeJFrogDistribution:
		return utils.String(strings.TrimSuffix(args.Url, "/") + "/distribution")
	case ServiceEndpointTypeJFrogPlatform:
		return utils.String(strings.TrimSuffix(args.Url, "/"))
	case ServiceEndpointTypeJFrogXray:
		return utils.String(strings.TrimSuffix(args.Url, "/") + "/xray")
	case ServiceEndpointTypekubernetes:
		return utils.String(args.Url)
	case ServiceEndpointTypeVsAppCenter:
		return utils.String("https://api.appcenter.ms/v0.1")
	default:
		return nil
	}
}
