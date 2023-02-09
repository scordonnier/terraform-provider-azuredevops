---
page_title: "azuredevops_serviceendpoint_dockerregistry Resource - azuredevops"
subcategory: "Service Endpoints"
description: |-
  Manages a Docker Registry service endpoint within an Azure DevOps project.
---

# azuredevops_serviceendpoint_dockerregistry (Resource)

Manages a Docker Registry service endpoint within an Azure DevOps project.

## Example Usage

```terraform
data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_dockerregistry" "hub" {
  description         = "Managed by Terraform"
  grant_all_pipelines = true
  name                = "Docker-Hub"
  password            = "GTu62azpC#qA2K*X"
  project_id          = data.azuredevops_project.sandbox.id
  url                 = "https://index.docker.io/v1/"
  username            = "username"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `grant_all_pipelines` (Boolean) Set to true to grant access to all pipelines in the project.
- `name` (String) The name of the service endpoint.
- `password` (String, Sensitive) Docker registry password.
- `project_id` (String) The ID of the project.
- `url` (String) Docker registry URL.
- `username` (String, Sensitive) Docker registry username.

### Optional

- `description` (String) The description of the service endpoint.

### Read-Only

- `id` (String) The ID of the service endpoint.