resource "azuredevops_agent_pool" "production" {
  auto_provision = false
  auto_update    = true
  name           = "Production"
}
