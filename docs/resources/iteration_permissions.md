---
page_title: "azuredevops_iteration_permissions Resource - azuredevops"
subcategory: "Work Items"
description: |-
  Sets permissions on iterations of an existing project within Azure DevOps. All permissions that currently exists will be overwritten.
---

# azuredevops_iteration_permissions (Resource)

Sets permissions on iterations of an existing project within Azure DevOps. All permissions that currently exists will be overwritten.

## Example Usage

```terraform
data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_iteration_permissions" "root" {
  path           = "" // Apply permissions to the root iteration
  principal_name = "[Sandbox]\\Contributors"
  project_id     = data.azuredevops_project.sandbox.id
  permissions = {
    create             = "allow"
    delete             = "notset"
    manage_test_plans  = "notset"
    manage_test_suites = "notset"
    read               = "allow"
    workitems_read     = "allow"
    workitems_write    = "allow"
    write              = "allow"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `path` (String) The path of the iteration.
- `permissions` (Attributes) The permissions to assign. (see [below for nested schema](#nestedatt--permissions))
- `principal_name` (String) The principal name to assign the permissions.
- `project_id` (String) The ID of the project.

### Read-Only

- `principal_descriptor` (String) The principal descriptor to assign the permissions.

<a id="nestedatt--permissions"></a>
### Nested Schema for `permissions`

Required:

- `create` (String) Sets the `CREATE_CHILDREN` permission for the identity. Must be `notset`, `allow` or `deny`.
- `delete` (String) Sets the `DELETE` permission for the identity. Must be `notset`, `allow` or `deny`.
- `read` (String) Sets the `GENERIC_READ` permission for the identity. Must be `notset`, `allow` or `deny`.
- `write` (String) Sets the `GENERIC_WRITE` permission for the identity. Must be `notset`, `allow` or `deny`.