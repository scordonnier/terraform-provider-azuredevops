package serviceendpoint

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoint"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
)

func CreateResourceServiceEndpoint(ctx context.Context, args *serviceendpoint.CreateOrUpdateServiceEndpointArgs, projectId string, client *serviceendpoint.Client, resp *resource.CreateResponse) (*serviceendpoint.ServiceEndpoint, error) {
	serviceEndpoint, err := client.CreateServiceEndpoint(ctx, args, projectId)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create service endpoint", err.Error())
		return nil, err
	}
	return serviceEndpoint, nil
}

func ReadResourceServiceEndpoint(ctx context.Context, id string, projectId string, client *serviceendpoint.Client, resp *resource.ReadResponse) (*serviceendpoint.ServiceEndpoint, error) {
	serviceEndpoint, err := client.GetServiceEndpoint(ctx, id, projectId)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return nil, err
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Error looking up service endpoint with Id '%s', %+v", id, err), "")
		return nil, err
	}

	if serviceEndpoint == nil {
		resp.State.RemoveResource(ctx)
		return nil, errors.New("service endpoint does not exist anymore")
	}

	return serviceEndpoint, nil
}

func UpdateResourceServiceEndpoint(ctx context.Context, id string, args *serviceendpoint.CreateOrUpdateServiceEndpointArgs, projectId string, client *serviceendpoint.Client, resp *resource.UpdateResponse) (*serviceendpoint.ServiceEndpoint, error) {
	serviceEndpoint, err := client.UpdateServiceEndpoint(ctx, id, args, projectId)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("Service connection with Id '%s' does not exist", id), "")
			return nil, err
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Error looking up service endpoint with Id '%s', %+v", id, err), "")
		return nil, err
	}

	return serviceEndpoint, nil
}

func DeleteResourceServiceEndpoint(ctx context.Context, id string, projectId string, client *serviceendpoint.Client, resp *resource.DeleteResponse) {
	err := client.DeleteServiceEndpoint(ctx, id, []string{projectId})
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Service connection with Id '%s' failed to delete", id), err.Error())
	}
}
