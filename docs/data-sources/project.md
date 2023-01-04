---
page_title: "azuredevops_project Data Source - azuredevops"
subcategory: ""
description: |-
  Use this data source to access information about an existing project within Azure DevOps.
---

# azuredevops_project (Data Source)

Use this data source to access information about an existing project within Azure DevOps.

## Example Usage

```terraform
data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name (or ID) of the project.

### Read-Only

- `id` (String) ID of the project.