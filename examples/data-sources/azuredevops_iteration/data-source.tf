data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

data "azuredevops_iteration" "q1-2023" {
  path       = "Q1-2023"
  project_id = data.azuredevops_project.sandbox.id
}

data "azuredevops_iteration" "q1-2023-sprint1" {
  path       = "Q1-2023/Sprint 1"
  project_id = data.azuredevops_project.sandbox.id
}