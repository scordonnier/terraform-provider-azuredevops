data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_iteration_permissions" "root" {
  path           = "" // Apply permissions to the root iteration
  principal_name = "[Sandbox]\\Contributors"
  project_id     = data.azuredevops_project.sandbox.id
  permissions = {
    create             = "allow"
    delete             = "notset"
    manage_test_plans  = "notset"
    manage_test_suites = "notset"
    read               = "allow"
    workitems_read     = "allow"
    workitems_write    = "allow"
    write              = "allow"
  }
}