## v0.6.2

FEATURES:

**New Resource** `azuredevops_serviceendpoint_sonarcloud`<br/>

And also various improvements and bug fixes.

## v0.6.1

FEATURES:

**Update Resource** `azuredevops_environment` : Add new property `grant_all_pipelines`<br/>

## v0.6.0

FEATURES:

**New Resource** `azuredevops_agent_pool`<br/>
**New Resource** `azuredevops_agent_queue`<br/>
**New Resource** `azuredevops_serviceendpoint_dockerregistry`<br/>
**New Resource** `azuredevops_serviceendpoint_generic`<br/>
**New Resource** `azuredevops_serviceendpoint_npm`<br/>
**New Resource** `azuredevops_serviceendpoint_nuget`<br/>

## v0.5.0

FEATURES:

**New Resource** `azuredevops_area_permissions`<br/>
**New Resource** `azuredevops_git_permissions`<br/>
**New Resource** `azuredevops_iteration_permissions`<br/>
**New Resource** `azuredevops_pipeline_permissions`<br/>
**New Resource** `azuredevops_project_permissions`<br/>
**New Resource** `azuredevops_serviceendpoint_jfrog`<br/>
**New Resource** `azuredevops_serviceendpoint_permissions`<br/>

BREAKING CHANGES:

**Rename (Data Source / Resource)** `azuredevops_pipelines_settings` to `azuredevops_pipeline_settings`<br/>
**Resource** `azuredevops_environment_permissions` : Update schema to specify permissions as new resources<br/>

## v0.4.1

BUG FIXES:

**Resource** `azuredevops_project_features` : Crash when specifying values using variables<br/>
**Resource** `azuredevops_serviceendpoint_XXX` : Service endpoints not deleted everywhere when shared with other projects<br/>

## v0.4.0

FEATURES:

**New Data Source** `azuredevops_area`<br/>
**New Data Source** `azuredevops_iteration`<br/>
**New Data Source** `azuredevops_pipelines_settings`<br/>
**New Data Source** `azuredevops_project_features`<br/>

**New Resource** `azuredevops_area`<br/>
**New Resource** `azuredevops_environment_kubernetes`<br/>
**New Resource** `azuredevops_iteration`<br/>
**New Resource** `azuredevops_pipelines_settings`<br/>
**New Resource** `azuredevops_project_features`<br/>
**New Resource** `azuredevops_serviceendpoint_kubernetes`<br/>

**Update Resource** `azuredevops_serviceendpoint_azurerm `<br/>
**Update Resource** `azuredevops_serviceendpoint_bitbucket `<br/>
**Update Resource** `azuredevops_serviceendpoint_github `<br/>
**Update Resource** `azuredevops_serviceendpoint_vsappcenter `<br/>

## v0.3.0

FEATURES:

**New Resource** `azuredevops_group_membership`<br/>
**New Resource** `azuredevops_serviceendpoint_github`<br/>
**New Resource** `azuredevops_serviceendpoint_vsappcenter`<br/>

And also various improvements and bug fixes.

## v0.2.0

FEATURES:

**New Data Source** `azuredevops_group`<br/>
**New Data Source** `azuredevops_groups`<br/>
**New Data Source** `azuredevops_process`<br/>
**New Data Source** `azuredevops_team`<br/>
**New Data Source** `azuredevops_teams`<br/>
**New Data Source** `azuredevops_user`<br/>
**New Data Source** `azuredevops_users`<br/>

**New Resource** `azuredevops_group`<br/>
**New Resource** `azuredevops_project`<br/>
**New Resource** `azuredevops_team`<br/>

BREAKING CHANGES:

**Rename** `azuredevops_distributedtask_environment` to `azuredevops_environment`<br/>
**Rename** `azuredevops_distributedtask_environment_permissions` to `azuredevops_environment_permissions`<br/>

## v0.1.0

FEATURES:

**New Data Source** `azuredevops_project`<br/>

**New Resource** `azuredevops_distributedtask_environment`<br/>
**New Resource** `azuredevops_distributedtask_environment_permissions`<br/> 
**New Resource** `azuredevops_serviceendpoint_azurerm`<br/>
**New Resource** `azuredevops_serviceendpoint_bitbucket`<br/>
**New Resource** `azuredevops_serviceendpoint_share`<br/>
