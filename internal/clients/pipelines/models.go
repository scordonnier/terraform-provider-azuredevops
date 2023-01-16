package pipelines

import "github.com/scordonnier/terraform-provider-azuredevops/internal/clients/core"

type Permission struct {
	Authorized   *bool             `json:"authorized,omitempty"`
	AuthorizedBy *core.IdentityRef `json:"authorizedBy,omitempty"`
	AuthorizedOn *core.Time        `json:"authorizedOn,omitempty"`
}

type PipelineGeneralSettings struct {
	DisableClassicPipelineCreation   *bool `json:"disableClassicPipelineCreation,omitempty"`
	EnforceJobAuthScope              *bool `json:"enforceJobAuthScope,omitempty"`
	EnforceJobAuthScopeForReleases   *bool `json:"enforceJobAuthScopeForReleases,omitempty"`
	EnforceReferencedRepoScopedToken *bool `json:"enforceReferencedRepoScopedToken,omitempty"`
	EnforceSettableVar               *bool `json:"enforceSettableVar,omitempty"`
	PublishPipelineMetadata          *bool `json:"publishPipelineMetadata,omitempty"`
	StatusBadgesArePrivate           *bool `json:"statusBadgesArePrivate,omitempty"`
}

type PipelinePermission struct {
	Authorized   *bool             `json:"authorized,omitempty"`
	AuthorizedBy *core.IdentityRef `json:"authorizedBy,omitempty"`
	AuthorizedOn *core.Time        `json:"authorizedOn,omitempty"`
	Id           *int              `json:"id,omitempty"`
}

type PipelineRetentionSettings struct {
	PurgeArtifacts               *RetentionSetting `json:"purgeArtifacts,omitempty"`
	PurgePullRequestRuns         *RetentionSetting `json:"purgePullRequestRuns,omitempty"`
	PurgeRuns                    *RetentionSetting `json:"purgeRuns,omitempty"`
	RetainRunsPerProtectedBranch *RetentionSetting `json:"retainRunsPerProtectedBranch,omitempty"`
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

type RetentionSetting struct {
	Max   *int `json:"max,omitempty"`
	Min   *int `json:"min,omitempty"`
	Value *int `json:"value,omitempty"`
}

type UpdatePipelineRetentionSettings struct {
	PurgeArtifacts               *RetentionSetting `json:"artifactsRetention,omitempty"`
	PurgePullRequestRuns         *RetentionSetting `json:"pullRequestRunRetention,omitempty"`
	PurgeRuns                    *RetentionSetting `json:"runRetention,omitempty"`
	RetainRunsPerProtectedBranch *RetentionSetting `json:"retainRunsPerProtectedBranch,omitempty"`
}
