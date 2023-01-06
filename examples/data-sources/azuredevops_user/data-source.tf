data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

data "azuredevops_user" "someone" {
  mail_address = "someone@noreply.com"
  project_id   = data.azuredevops_project.sandbox.id
}