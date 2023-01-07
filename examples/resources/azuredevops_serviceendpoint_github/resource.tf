data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_github" "private" {
  access_token = "GTu62azpC#qA2K*X"
  description  = "Managed by Terraform"
  name         = "GitHub-Private"
  project_id   = data.azuredevops_project.sandbox.id
}
