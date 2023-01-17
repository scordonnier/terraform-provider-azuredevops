data "azuredevops_project" "sandbox" {
  name = "Sandbox"
}

resource "azuredevops_serviceendpoint_kubernetes" "production" {
  grant_all_pipelines = true
  description         = "Managed by Terraform"
  kubeconfig = {
    accept_untrusted_certs = false
    yaml_content           = "apiVersion: v1\nclusters:\n-cluster:\n ..."
  }
  name       = "K8S-Production"
  project_id = data.azuredevops_project.sandbox.id
}

resource "azuredevops_environment" "production" {
  description = "Managed by Terraform"
  name        = "Production"
  project_id  = data.azuredevops_project.sandbox.id
}

resource "azuredevops_environment_kubernetes" "production-api-backend" {
  environment_id      = azuredevops_environment.production.id
  name                = "API Backend"
  namespace           = "api-backend"
  project_id          = azuredevops_environment.production.project_id
  service_endpoint_id = azuredevops_serviceendpoint_kubernetes.production.id
}