data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_azurerm" "production" {
  description           = "Managed by Terraform"
  name                  = "AzureRM-Production"
  project_id            = data.azuredevops_project.sandbox.id
  service_principal_id  = "00000000-0000-0000-0000-000000000000"
  service_principal_key = "GTu62azpC#qA2K*X"
  subscription_id       = "00000000-0000-0000-0000-000000000000"
  subscription_name     = "Azure Subscription Name"
  tenant_id             = "00000000-0000-0000-0000-000000000000"
}

resource "azuredevops_serviceendpoint_share" "production" {
  description = azuredevops_serviceendpoint_azurerm.production.description
  id          = azuredevops_serviceendpoint_azurerm.production.id
  name        = azuredevops_serviceendpoint_azurerm.production.name
  project_id  = azuredevops_serviceendpoint_azurerm.production.project_id
  project_ids = [
    "00000000-0000-0000-0000-000000000000",
    "11111111-1111-1111-1111-111111111111",
  ]
}