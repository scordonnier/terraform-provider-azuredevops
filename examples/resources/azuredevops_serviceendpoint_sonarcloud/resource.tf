data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_sonarcloud" "sonarcloud" {
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "SonarCloud"
  project_id          = data.azuredevops_project.sandbox.id
  token               = "GTu62azpC#qA2K*X"
}
