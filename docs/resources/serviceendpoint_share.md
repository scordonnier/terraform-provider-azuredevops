---
page_title: "azuredevops_serviceendpoint_share Resource - azuredevops"
subcategory: ""
description: |-
  Shares a service endpoint with multiple Azure DevOps projects.
---

# azuredevops_serviceendpoint_share (Resource)

Shares a service endpoint with multiple Azure DevOps projects.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) The ID of the service endpoint.
- `name` (String) The name of the service endpoint.
- `project_id` (String) The ID of the project hosting the service endpoint.
- `project_ids` (List of String) The IDs of the projects to share the service endpoint.

### Optional

- `description` (String) The description of the service endpoint.