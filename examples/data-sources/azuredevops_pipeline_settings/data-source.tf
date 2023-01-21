data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

data "azuredevops_pipeline_settings" "sandbox" {
  project_id = data.azuredevops_project.sandbox.id
}
