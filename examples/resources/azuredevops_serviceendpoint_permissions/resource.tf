data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_permissions" "sandbox" {
  project_id = data.azuredevops_project.sandbox.id
  permissions = [
    {
      identity_name = "[Sandbox]\\Contributors"

      administer         = "notset"
      create             = "notset"
      use                = "allow"
      view_authorization = "allow"
      view_endpoint      = "allow"
    },
    {
      identity_name = "user@noreply.com"

      administer         = "allow"
      create             = "allow"
      use                = "allow"
      view_authorization = "allow"
      view_endpoint      = "allow"
    }
  ]
}