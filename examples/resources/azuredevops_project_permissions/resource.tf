data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_project_permissions" "sandbox" {
  project_id = data.azuredevops_project.sandbox.id
  permissions = [
    {
      identity_name = "[Sandbox]\\Contributors"
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
    },
    {
      identity_name = "user@noreply.com"
      boards = {
        bypass_rules                = "allow"
        change_process              = "allow"
        workitem_delete             = "allow"
        workitem_move               = "allow"
        workitem_permanently_delete = "allow"
      }
      general = {
        delete                 = "allow"
        manage_properties      = "allow"
        rename                 = "allow"
        read                   = "allow"
        suppress_notifications = "allow"
        update_visibility      = "allow"
        write                  = "allow"
      }
      test_plans = {
        delete_test_results        = "allow"
        manage_test_configurations = "allow"
        manage_test_environments   = "allow"
        publish_test_results       = "allow"
        view_test_results          = "allow"
      }
    }
  ]
}
