package distributedtask

import (
	"github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"
)

type CreateOrUpdateEnvironmentArgs struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
}

type EnvironmentInstance struct {
	CreatedBy      *core.IdentityRef               `json:"createdBy,omitempty"`
	CreatedOn      *core.Time                      `json:"createdOn,omitempty"`
	Description    *string                         `json:"description,omitempty"`
	Id             *int                            `json:"id,omitempty"`
	LastModifiedBy *core.IdentityRef               `json:"lastModifiedBy,omitempty"`
	LastModifiedOn *core.Time                      `json:"lastModifiedOn,omitempty"`
	Name           *string                         `json:"name,omitempty"`
	Project        *core.ProjectReference          `json:"project,omitempty"`
	Resources      *[]EnvironmentResourceReference `json:"resources,omitempty"`
}

type EnvironmentReference struct {
	Id   *int    `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

type EnvironmentResourceKubernetes struct {
	ClusterName          *string               `json:"clusterName,omitempty"`
	CreatedBy            *core.IdentityRef     `json:"createdBy,omitempty"`
	CreatedOn            *core.Time            `json:"createdOn,omitempty"`
	EnvironmentReference *EnvironmentReference `json:"environmentReference,omitempty"`
	Id                   *int                  `json:"id,omitempty"`
	LastModifiedBy       *core.IdentityRef     `json:"lastModifiedBy,omitempty"`
	LastModifiedOn       *core.Time            `json:"lastModifiedOn,omitempty"`
	Name                 *string               `json:"name,omitempty"`
	Namespace            *string               `json:"namespace,omitempty"`
	ServiceEndpointId    *string               `json:"serviceEndpointId,omitempty"`
	Type                 *string               `json:"type,omitempty"`
	Tags                 *[]string             `json:"tags,omitempty"`
}

type EnvironmentResourceReference struct {
	Id   *int                     `json:"id,omitempty"`
	Name *string                  `json:"name,omitempty"`
	Tags *[]string                `json:"tags,omitempty"`
	Type *EnvironmentResourceType `json:"type,omitempty"`
}

type EnvironmentResourceType string
