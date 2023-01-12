package pipelines

import "github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"

type Permission struct {
	Authorized   *bool             `json:"authorized,omitempty"`
	AuthorizedBy *core.IdentityRef `json:"authorizedBy,omitempty"`
	AuthorizedOn *core.Time        `json:"authorizedOn,omitempty"`
}

type PipelinePermission struct {
	Authorized   *bool             `json:"authorized,omitempty"`
	AuthorizedBy *core.IdentityRef `json:"authorizedBy,omitempty"`
	AuthorizedOn *core.Time        `json:"authorizedOn,omitempty"`
	Id           *int              `json:"id,omitempty"`
}

type Resource struct {
	Id   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`
}

type ResourcePipelinePermissions struct {
	AllPipelines *Permission           `json:"allPipelines,omitempty"`
	Pipelines    *[]PipelinePermission `json:"pipelines,omitempty"`
	Resource     *Resource             `json:"resource,omitempty"`
}
