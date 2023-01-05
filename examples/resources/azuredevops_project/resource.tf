data "azuredevops_process" "agile" {
  name = "Agile"
}

resource "azuredevops_project" "Sandbox" {
  description         = "Managed by Terraform"
  name                = "Sandbox"
  process_template_id = data.azuredevops_process.agile.id
  version_control     = "Git"
  visibility          = "private"
}
