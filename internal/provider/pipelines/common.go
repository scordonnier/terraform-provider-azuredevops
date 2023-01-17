package pipelines

type PipelineGeneralSettings struct {
	DisableClassicPipelineCreation   *bool `tfsdk:"disable_classic_pipeline_creation"`
	EnforceJobAuthScope              *bool `tfsdk:"enforce_job_auth_scope"`
	EnforceJobAuthScopeForReleases   *bool `tfsdk:"enforce_job_auth_scope_for_releases"`
	EnforceReferencedRepoScopedToken *bool `tfsdk:"enforce_referenced_repo_scoped_token"`
	EnforceSettableVar               *bool `tfsdk:"enforce_settable_var"`
	PublishPipelineMetadata          *bool `tfsdk:"publish_pipeline_metadata"`
	StatusBadgesArePrivate           *bool `tfsdk:"status_badges_are_private"`
}

type PipelineRetentionSettings struct {
	DaysToKeepArtifacts       *int `tfsdk:"days_to_keep_artifacts"`
	DaysToKeepPullRequestRuns *int `tfsdk:"days_to_keep_pullrequest_runs"`
	DaysToKeepRuns            *int `tfsdk:"days_to_keep_runs"`
}
