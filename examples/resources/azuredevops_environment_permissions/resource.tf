data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_environment" "production" {
  name       = "Production"
  project_id = data.azuredevops_project.sandbox.id
}

resource "azuredevops_environment_permissions" "production" {
  id             = azuredevops_environment.production.id
  principal_name = "[Sandbox]\\Contributors"
  project_id     = azuredevops_environment.production.project_id
  permissions = {
    administer     = "notset"
    create         = "notset"
    manage         = "notset"
    manage_history = "notset"
    use            = "allow"
    view           = "allow"
  }
}