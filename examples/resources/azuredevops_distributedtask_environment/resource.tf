data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_distributedtask_environment" "production" {
  name       = "Production"
  project_id = data.azuredevops_project.sandbox.id
}