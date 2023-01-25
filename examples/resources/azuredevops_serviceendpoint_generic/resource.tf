data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_generic" "generic-api" {
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "Generic-API"
  password            = "GTu62azpC#qA2K*X"
  project_id          = data.azuredevops_project.sandbox.id
  url                 = "https://server.domain.com/"
  username            = "username"
}
