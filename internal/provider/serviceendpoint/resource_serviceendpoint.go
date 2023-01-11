package serviceendpoint

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/pipelines"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/serviceendpoint"
	"github.com/scordonnier/terraform-provider-azuredevops/internal/utils"
	"time"
)

func CreateResourceServiceEndpoint(ctx context.Context, projectId string, args *serviceendpoint.CreateOrUpdateServiceEndpointArgs, serviceEndpointClient *serviceendpoint.Client, pipelineClient *pipelines.Client, resp *resource.CreateResponse) (*serviceendpoint.ServiceEndpoint, error) {
	serviceEndpoint, err := serviceEndpointClient.CreateServiceEndpoint(ctx, args, projectId)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create service endpoint", err.Error())
		return nil, err
	}

	stateRefreshFunc := func() (interface{}, string, error) {
		pendingServiceEndpoint, err := serviceEndpointClient.GetServiceEndpoint(ctx, serviceEndpoint.Id.String(), projectId)
		if err != nil {
			return nil, "Failed", err
		}

		if *pendingServiceEndpoint.IsReady {
			return pendingServiceEndpoint, "Ready", nil
		} else if pendingServiceEndpoint.OperationStatus != nil {
			opStatus := ((pendingServiceEndpoint.OperationStatus).(map[string]interface{})["state"]).(string)
			if opStatus == "Failed" {
				return nil, opStatus, errors.New("failed to create service endpoint")
			}
			return nil, opStatus, nil
		}
		return nil, "Failed", errors.New("failed to create service endpoint")
	}
	stateConf := &utils.StateChangeConf{
		Delay:      1 * time.Second,
		MinTimeout: 5 * time.Second,
		Pending:    []string{"InProgress"},
		Target:     []string{"Ready", "Failed"},
		Refresh:    stateRefreshFunc,
		Timeout:    30 * time.Second,
	}

	readyServiceEndpoint, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		serviceEndpointClient.DeleteServiceEndpoint(ctx, serviceEndpoint.Id.String(), []string{projectId})
		return nil, err
	}

	serviceEndpoint = readyServiceEndpoint.(*serviceendpoint.ServiceEndpoint)
	_, err = pipelineClient.GrantAllPipelines(ctx, projectId, pipelines.PipelinePermissionsResourceTypeEndpoint, serviceEndpoint.Id.String(), args.GrantAllPipelines)
	if err != nil {
		resp.Diagnostics.AddError("Unable to grant service endpoint access to all pipelines", err.Error())
		return nil, err
	}

	return serviceEndpoint, nil
}

func ReadResourceServiceEndpoint(ctx context.Context, id string, projectId string, serviceEndpointClient *serviceendpoint.Client, pipelineClient *pipelines.Client, resp *resource.ReadResponse) (*serviceendpoint.ServiceEndpoint, bool, error) {
	serviceEndpoint, err := serviceEndpointClient.GetServiceEndpoint(ctx, id, projectId)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return nil, false, err
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Error looking up service endpoint with Id '%s', %+v", id, err), "")
		return nil, false, err
	}

	if serviceEndpoint == nil {
		resp.State.RemoveResource(ctx)
		return nil, false, errors.New("service endpoint does not exist anymore")
	}

	permissions, err := pipelineClient.GetPipelinePermissions(ctx, projectId, pipelines.PipelinePermissionsResourceTypeEndpoint, serviceEndpoint.Id.String())
	if err != nil {
		resp.Diagnostics.AddError("Unable to retrieve grant access", err.Error())
		return nil, false, err
	}

	var authorized bool
	if permissions.AllPipelines != nil {
		authorized = *permissions.AllPipelines.Authorized
	}
	return serviceEndpoint, authorized, nil
}

func UpdateResourceServiceEndpoint(ctx context.Context, id string, projectId string, args *serviceendpoint.CreateOrUpdateServiceEndpointArgs, serviceEndpointClient *serviceendpoint.Client, pipelineClient *pipelines.Client, resp *resource.UpdateResponse) (*serviceendpoint.ServiceEndpoint, error) {
	serviceEndpoint, err := serviceEndpointClient.UpdateServiceEndpoint(ctx, id, args, projectId)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError(fmt.Sprintf("Service connection with Id '%s' does not exist", id), "")
			return nil, err
		}

		resp.Diagnostics.AddError(fmt.Sprintf("Error looking up service endpoint with Id '%s', %+v", id, err), "")
		return nil, err
	}

	_, err = pipelineClient.GrantAllPipelines(ctx, projectId, pipelines.PipelinePermissionsResourceTypeEndpoint, serviceEndpoint.Id.String(), args.GrantAllPipelines)
	if err != nil {
		resp.Diagnostics.AddError("Unable to grant service endpoint access to all pipelines", err.Error())
		return nil, err
	}

	return serviceEndpoint, nil
}

func DeleteResourceServiceEndpoint(ctx context.Context, id string, projectId string, serviceEndpointClient *serviceendpoint.Client, resp *resource.DeleteResponse) {
	err := serviceEndpointClient.DeleteServiceEndpoint(ctx, id, []string{projectId})
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Service connection with Id '%s' failed to delete", id), err.Error())
	}
}

func GetServiceEndpointResourceSchemaBase(description string) schema.Schema {
	return schema.Schema{
		MarkdownDescription: description,
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the service endpoint.",
				Optional:            true,
			},
			"grant_all_pipelines": schema.BoolAttribute{
				MarkdownDescription: "Set to true to grant access to all pipelines in the project.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the service endpoint.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service endpoint.",
				Required:            true,
				Validators: []validator.String{
					utils.StringNotEmptyValidator(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the project.",
				Required:            true,
				Validators: []validator.String{
					utils.UUIDStringValidator(),
				},
			},
		},
	}
}
