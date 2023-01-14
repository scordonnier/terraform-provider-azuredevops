data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_iteration" "q1-2023" {
  name        = "2023-Q1"
  parent_path = "" // Create the iteration as a child of the root node
  project_id  = data.azuredevops_project.sandbox.id
}

resource "azuredevops_iteration" "q1-2023-sprint1" {
  finish_date = "2023-01-13T00:00:00Z"
  name        = "Sprint 1"
  parent_path = azuredevops_iteration.q1-2023.path
  project_id  = data.azuredevops_project.sandbox.id
  start_date  = "2023-01-02T00:00:00Z"
}

resource "azuredevops_iteration" "q1-2023-sprint2" {
  finish_date = "2023-01-27T00:00:00Z"
  name        = "Sprint 2"
  parent_path = azuredevops_iteration.q1-2023.path
  project_id  = data.azuredevops_project.sandbox.id
  start_date  = "2023-01-16T00:00:00Z"
}
