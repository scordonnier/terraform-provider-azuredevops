resource "azuredevops_agent_pool" "production" {
  auto_provision = false
  auto_update    = true
  name           = "Production"
}

data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_agent_queue" "production" {
  agent_pool_id       = azuredevops_agent_pool.production.id
  grant_all_pipelines = true
  name                = azuredevops_agent_pool.production.name
  project_id          = data.azuredevops_project.sandbox.id
}
