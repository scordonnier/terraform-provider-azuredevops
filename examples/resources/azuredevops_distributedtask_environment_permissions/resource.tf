data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_distributedtask_environment" "production" {
  name       = "Production"
  project_id = data.azuredevops_project.sandbox.id
}

resource "azuredevops_distributedtask_environment_permissions" "production" {
  id         = azuredevops_distributedtask_environment.production.id
  project_id = azuredevops_distributedtask_environment.production.project_id
  permissions {
    identity_name = "[Sandbox]\\Contributors"
    identity_type = "group"

    administer     = "notset"
    create         = "notset"
    manage         = "notset"
    manage_history = "notset"
    use            = "allow"
    view           = "allow"
  }
  permissions {
    identity_name = "user@noreply.com"
    identity_type = "user"

    administer     = "allow"
    create         = "allow"
    manage         = "allow"
    manage_history = "allow"
    use            = "allow"
    view           = "allow"
  }
}