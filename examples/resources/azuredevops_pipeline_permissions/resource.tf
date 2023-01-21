data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_pipeline_permissions" "sandbox" {
  project_id     = data.azuredevops_project.sandbox.id
  principal_name = "[Sandbox]\\Contributors"
  permissions = {
    administer_build_permissions      = "notset"
    delete_build_definition           = "allow"
    delete_builds                     = "allow"
    destroy_builds                    = "allow"
    edit_build_definition             = "allow"
    edit_build_quality                = "allow"
    manage_build_qualities            = "notset"
    manage_build_queue                = "notset"
    override_build_checkin_validation = "notset"
    queue_builds                      = "allow"
    retain_indefinitely               = "allow"
    stop_builds                       = "allow"
    update_build_information          = "allow"
    view_build_definition             = "allow"
    view_builds                       = "allow"
  }
}