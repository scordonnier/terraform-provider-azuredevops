data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_group" "developers" {
  description  = "Description of the Developers group"
  display_name = "Developers"
  project_id   = data.azuredevops_project.sandbox.id
}

resource "azuredevops_group_membership" "developers" {
  display_name = azuredevops_group.developers.display_name
  project_id   = data.azuredevops_project.sandbox.id
  members = [
    "Someone",             // AAD user or group based on its display name
    "someone@noreply.com", // AAD user or group based on its mail address
    "[Sandbox]\\Readers"   // Azure DevOps Group
  ]
}
