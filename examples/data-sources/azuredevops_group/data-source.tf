data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

data "azuredevops_group" "contributors" {
  display_name = "Contributors"
  project_id   = data.azuredevops_project.sandbox.id
}