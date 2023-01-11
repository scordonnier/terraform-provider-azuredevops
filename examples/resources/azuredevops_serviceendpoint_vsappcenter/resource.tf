data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_vsappcenter" "vsappcenter" {
  api_token           = "GTu62azpC#qA2K*X"
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "Visual Studio App Center"
  project_id          = data.azuredevops_project.sandbox.id
}
