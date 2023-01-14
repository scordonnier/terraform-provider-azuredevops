data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_area" "it-department" {
  name        = "IT Departmnent"
  parent_path = "" // Create the area as a child of the root node
  project_id  = data.azuredevops_project.sandbox.id
}

resource "azuredevops_area" "architecture" {
  name        = "Architecture"
  parent_path = azuredevops_area.it-department.path
  project_id  = data.azuredevops_project.sandbox.id
}
