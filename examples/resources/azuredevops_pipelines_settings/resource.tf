data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_pipelines_settings" "sandbox" {
  general = {
    disable_classic_pipeline_creation    = false
    enforce_job_auth_scope               = true
    enforce_job_auth_scope_for_releases  = true
    enforce_referenced_repo_scoped_token = true
    enforce_settable_var                 = true
    publish_pipeline_metadata            = false
    status_badges_are_private            = true
  }
  project_id = data.azuredevops_project.sandbox.id
  retention = {
    days_to_keep_artifacts        = 30
    days_to_keep_pullrequest_runs = 10
    days_to_keep_runs             = 30
  }
}
