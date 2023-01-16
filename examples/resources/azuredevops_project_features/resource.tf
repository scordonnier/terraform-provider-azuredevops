data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_project_features" "sandbox" {
  artifacts    = "disabled"
  boards       = "enabled"
  pipelines    = "enabled"
  project_id   = data.azuredevops_project.sandbox.id
  repositories = "enabled"
  testplans    = "disabled"
}
