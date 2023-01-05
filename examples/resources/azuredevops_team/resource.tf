data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_team" "developers" {
  description = "Description of the Developers Team"
  name        = "Developers Team"
  project_id  = data.azuredevops_project.sandbox.id
}