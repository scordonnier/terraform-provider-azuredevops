data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_bitbucket" "private" {
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "Bitbucket-Private"
  password            = "GTu62azpC#qA2K*X"
  project_id          = data.azuredevops_project.sandbox.id
  username            = "username"
}
