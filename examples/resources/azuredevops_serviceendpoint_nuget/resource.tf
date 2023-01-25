data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_nuget" "nuget-feed" {
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "NuGet-Feed"
  password            = "GTu62azpC#qA2K*X"
  project_id          = data.azuredevops_project.sandbox.id
  url                 = "https://api.nuget.org/v3/index.json"
  username            = "username"
}

resource "azuredevops_serviceendpoint_nuget" "nuget-feed" {
  api_key             = "GTu62azpC#qA2K*X"
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "NuGet-Feed"
  project_id          = data.azuredevops_project.sandbox.id
  url                 = "https://api.nuget.org/v3/index.json"
}
