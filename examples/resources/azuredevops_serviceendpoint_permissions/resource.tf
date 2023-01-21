data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_permissions" "sandbox" {
  principal_name = "[Sandbox]\\Contributors"
  project_id     = data.azuredevops_project.sandbox.id
  permissions = {
    administer         = "notset"
    create             = "notset"
    use                = "allow"
    view_authorization = "allow"
    view_endpoint      = "allow"
  }
}