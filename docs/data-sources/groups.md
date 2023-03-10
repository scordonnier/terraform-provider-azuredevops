---
page_title: "azuredevops_groups Data Source - azuredevops"
subcategory: "Users & Groups"
description: |-
  Use this data source to access information about existing groups within an Azure DevOps project.
---

# azuredevops_groups (Data Source)

Use this data source to access information about existing groups within an Azure DevOps project.

## Example Usage

```terraform
data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

data "azuredevops_groups" "sandbox" {
  project_id = data.azuredevops_project.sandbox.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project_id` (String) The ID of the project.

### Read-Only

- `groups` (Attributes List) The list of groups within the project. (see [below for nested schema](#nestedatt--groups))

<a id="nestedatt--groups"></a>
### Nested Schema for `groups`

Read-Only:

- `description` (String) The description of the group.
- `descriptor` (String) The descriptor of the group.
- `display_name` (String) The display name of the group.
- `name` (String) The name of the group.
- `origin` (String) The type of source provider for the group (eg. AD, AAD, MSA).
- `origin_id` (String) The unique identifier from the system of origin.
- `project_id` (String) The project ID of the group.
