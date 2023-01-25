data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_npm" "npm-registry" {
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "Npm-Registry"
  password            = "GTu62azpC#qA2K*X"
  project_id          = data.azuredevops_project.sandbox.id
  url                 = "https://registry.npmjs.org/"
  username            = "username"
}

resource "azuredevops_serviceendpoint_npm" "npm-registry" {
  access_token        = "GTu62azpC#qA2K*X"
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "Npm-Registry"
  project_id          = data.azuredevops_project.sandbox.id
  url                 = "https://registry.npmjs.org/"
}
