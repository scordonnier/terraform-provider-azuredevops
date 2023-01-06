data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

data "azuredevops_groups" "sandbox" {
  project_id = data.azuredevops_project.sandbox.id
}