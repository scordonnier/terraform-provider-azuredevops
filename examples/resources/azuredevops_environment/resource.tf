data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_environment" "production" {
  grant_all_pipelines = true
  name                = "Production"
  project_id          = data.azuredevops_project.sandbox.id
}