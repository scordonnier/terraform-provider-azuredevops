---
page_title: "azuredevops_project Resource - azuredevops"
subcategory: "Projects"
description: |-
  Manage a project within Azure DevOps.
---

# azuredevops_project (Resource)

Manage a project within Azure DevOps.

## Example Usage

```terraform
data "azuredevops_process" "agile" {
  name = "Agile"
}

resource "azuredevops_project" "Sandbox" {
  description         = "Managed by Terraform"
  name                = "Sandbox"
  process_template_id = data.azuredevops_process.agile.id
  version_control     = "Git"
  visibility          = "private"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the project.
- `process_template_id` (String) The process template ID of the project.
- `version_control` (String) Specifies the visibility of the project. Must be `Git` or `Tfvc`.
- `visibility` (String) Specifies the visibility of the project. Must be `private` or `public`.

### Optional

- `description` (String) The description of the project.

### Read-Only

- `id` (String) The ID of the project.
