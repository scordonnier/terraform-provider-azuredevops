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