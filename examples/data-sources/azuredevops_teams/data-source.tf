data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

data "azuredevops_teams" "sandbox" {
  project_id = data.azuredevops_project.sandbox.id
}