---
page_title: "azuredevops_project_features Data Source - azuredevops"
subcategory: "Projects"
description: |-
  Use this data source to access information about features of an existing project within Azure DevOps.
---

# azuredevops_project_features (Data Source)

Use this data source to access information about features of an existing project within Azure DevOps.

## Example Usage

```terraform
data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

data "azuredevops_project_features" "sandbox" {
  project_id = data.azuredevops_project.sandbox.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project_id` (String) The ID of the project.

### Read-Only

- `artifacts` (String) If enabled, gives access to Azure Artifacts.
- `boards` (String) If enabled, gives access to Azure Boards.
- `pipelines` (String) If enabled, gives access to Azure Pipelines.
- `repositories` (String) If enabled, gives access to Azure Repos.
- `testplans` (String) If enabled, gives access to Azure Test Plans.