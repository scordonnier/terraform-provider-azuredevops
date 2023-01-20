data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_area_permissions" "root" {
  path       = "" // Apply permissions to the root area
  project_id = data.azuredevops_project.sandbox.id
  permissions = [
    {
      identity_name = "[Sandbox]\\Contributors"

      create             = "allow"
      delete             = "notset"
      manage_test_plans  = "notset"
      manage_test_suites = "notset"
      read               = "allow"
      workitems_read     = "allow"
      workitems_write    = "allow"
      write              = "allow"
    },
    {
      identity_name = "user@noreply.com"

      create             = "allow"
      delete             = "allow"
      manage_test_plans  = "allow"
      manage_test_suites = "allow"
      read               = "allow"
      workitems_read     = "allow"
      workitems_write    = "allow"
      write              = "allow"
    }
  ]
}