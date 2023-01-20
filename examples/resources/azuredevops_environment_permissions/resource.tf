data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_environment" "production" {
  name       = "Production"
  project_id = data.azuredevops_project.sandbox.id
}

resource "azuredevops_environment_permissions" "production" {
  id         = azuredevops_environment.production.id
  project_id = azuredevops_environment.production.project_id
  permissions = [
    {
      identity_name = "[Sandbox]\\Contributors"

      administer     = "notset"
      create         = "notset"
      manage         = "notset"
      manage_history = "notset"
      use            = "allow"
      view           = "allow"
    },
    {
      identity_name = "user@noreply.com"

      administer     = "allow"
      create         = "allow"
      manage         = "allow"
      manage_history = "allow"
      use            = "allow"
      view           = "allow"
    }
  ]
}