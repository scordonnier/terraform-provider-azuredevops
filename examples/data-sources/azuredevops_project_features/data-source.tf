data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

data "azuredevops_project_features" "sandbox" {
  project_id = data.azuredevops_project.sandbox.id
}
