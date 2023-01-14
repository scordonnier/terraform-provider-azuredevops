data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

data "azuredevops_area" "it-department" {
  path       = "IT Department"
  project_id = data.azuredevops_project.sandbox.id
}

data "azuredevops_area" "architecture" {
  path       = "IT Department/Architecture"
  project_id = data.azuredevops_project.sandbox.id
}