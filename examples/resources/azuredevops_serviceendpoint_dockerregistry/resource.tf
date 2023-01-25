data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_dockerregistry" "hub" {
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "Docker-Hub"
  password            = "GTu62azpC#qA2K*X"
  project_id          = data.azuredevops_project.sandbox.id
  url                 = "https://index.docker.io/v1/"
  username            = "username"
}
