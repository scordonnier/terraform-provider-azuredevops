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
