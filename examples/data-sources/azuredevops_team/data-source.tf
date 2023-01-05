data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

data "azuredevops_team" "sandbox" {
  name       = "Sandbox Team"
  project_id = data.azuredevops_project.sandbox.id
}