data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_jfrog" "artifactory" {
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "JFrog-Artitfactory"
  password            = "GTu62azpC#qA2K*X"
  project_id          = data.azuredevops_project.sandbox.id
  service             = "artifactory"
  url                 = "https://my.jfrog.io"
  username            = "username"
}

resource "azuredevops_serviceendpoint_jfrog" "distribution" {
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "JFrog-Distribution"
  password            = "GTu62azpC#qA2K*X"
  project_id          = data.azuredevops_project.sandbox.id
  service             = "distribution"
  url                 = "https://my.jfrog.io"
  username            = "username"
}

resource "azuredevops_serviceendpoint_jfrog" "platform" {
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "JFrog-Platform"
  password            = "GTu62azpC#qA2K*X"
  project_id          = data.azuredevops_project.sandbox.id
  service             = "platform"
  url                 = "https://my.jfrog.io"
  username            = "username"
}

resource "azuredevops_serviceendpoint_jfrog" "xray" {
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "JFrog-Xray"
  password            = "GTu62azpC#qA2K*X"
  project_id          = data.azuredevops_project.sandbox.id
  service             = "xray"
  url                 = "https://my.jfrog.io"
  username            = "username"
}
