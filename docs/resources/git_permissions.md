---
page_title: "azuredevops_git_permissions Resource - azuredevops"
subcategory: "Git"
description: |-
  Sets permissions on repositories within an Azure DevOps project. All permissions that currently exists will be overwritten.
---

# azuredevops_git_permissions (Resource)

Sets permissions on repositories within an Azure DevOps project. All permissions that currently exists will be overwritten.

## Example Usage

```terraform
data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_git_permissions" "sandbox" {
  principal_name = "[Sandbox]\\Readers"
  project_id     = data.azuredevops_project.sandbox.id
  permissions = {
    administer                = "notset"
    create_branch             = "notset"
    create_repository         = "notset"
    create_tag                = "notset"
    contribute                = "notset"
    delete_repository         = "notset"
    edit_policies             = "notset"
    force_push                = "notset"
    manage_note               = "notset"
    manage_permissions        = "notset"
    policy_exempt             = "notset"
    pullrequest_bypass_policy = "notset"
    pullrequest_contribute    = "allow"
    read                      = "allow"
    remove_others_locks       = "notset"
    rename_repository         = "notset"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `permissions` (Attributes) The permissions to assign. (see [below for nested schema](#nestedatt--permissions))
- `principal_name` (String) The principal name to assign the permissions.
- `project_id` (String) The ID of the project.

### Optional

- `id` (String) The ID of the repository. If you omit the value, the permissions are applied to the repositories page and by default all repositories inherit permissions from there.

### Read-Only

- `principal_descriptor` (String) The principal descriptor to assign the permissions.

<a id="nestedatt--permissions"></a>
### Nested Schema for `permissions`

Required:

- `administer` (String) Sets the `Administer` permission for the identity. Must be `notset`, `allow` or `deny`.
- `contribute` (String) Sets the `GenericContribute` permission for the identity. Must be `notset`, `allow` or `deny`.
- `create_branch` (String) Sets the `CreateBranch` permission for the identity. Must be `notset`, `allow` or `deny`.
- `create_repository` (String) Sets the `CreateRepository` permission for the identity. Must be `notset`, `allow` or `deny`.
- `create_tag` (String) Sets the `CreateTag` permission for the identity. Must be `notset`, `allow` or `deny`.
- `delete_repository` (String) Sets the `DeleteRepository` permission for the identity. Must be `notset`, `allow` or `deny`.
- `edit_policies` (String) Sets the `EditPolicies` permission for the identity. Must be `notset`, `allow` or `deny`.
- `force_push` (String) Sets the `ForcePush` permission for the identity. Must be `notset`, `allow` or `deny`.
- `manage_note` (String) Sets the `ManageNote` permission for the identity. Must be `notset`, `allow` or `deny`.
- `manage_permissions` (String) Sets the `ManagePermissions` permission for the identity. Must be `notset`, `allow` or `deny`.
- `policy_exempt` (String) Sets the `PolicyExempt` permission for the identity. Must be `notset`, `allow` or `deny`.
- `pullrequest_bypass_policy` (String) Sets the `PullRequestBypassPolicy` permission for the identity. Must be `notset`, `allow` or `deny`.
- `pullrequest_contribute` (String) Sets the `PullRequestContribute` permission for the identity. Must be `notset`, `allow` or `deny`.
- `read` (String) Sets the `GenericRead` permission for the identity. Must be `notset`, `allow` or `deny`.
- `remove_others_locks` (String) Sets the `RemoveOthersLocks` permission for the identity. Must be `notset`, `allow` or `deny`.
- `rename_repository` (String) Sets the `RenameRepository` permission for the identity. Must be `notset`, `allow` or `deny`.