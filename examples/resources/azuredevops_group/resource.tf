data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_group" "developers" {
  description  = "Description of the Developers group"
  display_name = "Developers"
  project_id   = data.azuredevops_project.sandbox.id
}