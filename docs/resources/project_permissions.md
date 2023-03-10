---
page_title: "azuredevops_project_permissions Resource - azuredevops"
subcategory: "Projects"
description: |-
  Sets permissions on projects within Azure DevOps. All permissions that currently exists will be overwritten.
---

# azuredevops_project_permissions (Resource)

Sets permissions on projects within Azure DevOps. All permissions that currently exists will be overwritten.

## Example Usage

```terraform
data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_project_permissions" "sandbox" {
  principal_name = "[Sandbox]\\Contributors"
  project_id     = data.azuredevops_project.sandbox.id
  permissions = {
    boards = {
      bypass_rules                = "notset"
      change_process              = "notset"
      workitem_delete             = "allow"
      workitem_move               = "allow"
      workitem_permanently_delete = "allow"
    }
    general = {
      delete                 = "notset"
      manage_properties      = "notset"
      rename                 = "notset"
      read                   = "allow"
      suppress_notifications = "notset"
      update_visibility      = "notset"
      write                  = "notset"
    }
    test_plans = {
      delete_test_results        = "allow"
      manage_test_configurations = "allow"
      manage_test_environments   = "allow"
      publish_test_results       = "allow"
      view_test_results          = "allow"
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `permissions` (Attributes) The permissions to assign. (see [below for nested schema](#nestedatt--permissions))
- `principal_name` (String) The principal name to assign the permissions.
- `project_id` (String) The ID of the project.

### Read-Only

- `principal_descriptor` (String) The principal descriptor to assign the permissions.

<a id="nestedatt--permissions"></a>
### Nested Schema for `permissions`

Required:

- `boards` (Attributes) (see [below for nested schema](#nestedatt--permissions--boards))
- `general` (Attributes) (see [below for nested schema](#nestedatt--permissions--general))
- `test_plans` (Attributes) (see [below for nested schema](#nestedatt--permissions--test_plans))

<a id="nestedatt--permissions--boards"></a>
### Nested Schema for `permissions.boards`

Required:

- `bypass_rules` (String) Sets the `BYPASS_RULES` permission for the identity. Must be `notset`, `allow` or `deny`.
- `change_process` (String) Sets the `CHANGE_PROCESS` permission for the identity. Must be `notset`, `allow` or `deny`.
- `workitem_delete` (String) Sets the `WORK_ITEM_DELETE` permission for the identity. Must be `notset`, `allow` or `deny`.
- `workitem_move` (String) Sets the `WORK_ITEM_MOVE` permission for the identity. Must be `notset`, `allow` or `deny`.
- `workitem_permanently_delete` (String) Sets the `WORK_ITEM_PERMANENTLY_DELETE` permission for the identity. Must be `notset`, `allow` or `deny`.


<a id="nestedatt--permissions--general"></a>
### Nested Schema for `permissions.general`

Required:

- `delete` (String) Sets the `DELETE` permission for the identity. Must be `notset`, `allow` or `deny`.
- `manage_properties` (String) Sets the `MANAGE_PROPERTIES` permission for the identity. Must be `notset`, `allow` or `deny`.
- `read` (String) Sets the `GENERIC_READ` permission for the identity. Must be `notset`, `allow` or `deny`.
- `rename` (String) Sets the `RENAME` permission for the identity. Must be `notset`, `allow` or `deny`.
- `suppress_notifications` (String) Sets the `SUPPRESS_NOTIFICATIONS` permission for the identity. Must be `notset`, `allow` or `deny`.
- `update_visibility` (String) Sets the `UPDATE_VISIBILITY` permission for the identity. Must be `notset`, `allow` or `deny`.
- `write` (String) Sets the `GENERIC_WRITE` permission for the identity. Must be `notset`, `allow` or `deny`.


<a id="nestedatt--permissions--test_plans"></a>
### Nested Schema for `permissions.test_plans`

Required:

- `delete_test_results` (String) Sets the `DELETE_TEST_RESULTS` permission for the identity. Must be `notset`, `allow` or `deny`.
- `manage_test_configurations` (String) Sets the `MANAGE_TEST_CONFIGURATIONS` permission for the identity. Must be `notset`, `allow` or `deny`.
- `manage_test_environments` (String) Sets the `MANAGE_TEST_ENVIRONMENTS` permission for the identity. Must be `notset`, `allow` or `deny`.
- `publish_test_results` (String) Sets the `PUBLISH_TEST_RESULTS` permission for the identity. Must be `notset`, `allow` or `deny`.
- `view_test_results` (String) Sets the `VIEW_TEST_RESULTS` permission for the identity. Must be `notset`, `allow` or `deny`.
