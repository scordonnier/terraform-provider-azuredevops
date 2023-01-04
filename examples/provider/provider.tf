terraform {
  required_providers {
    azuredevops = {
      source  = "scordonnier/azuredevops"
      version = "0.1.0"
    }
  }
}

provider "azuredevops" {
  organization_url      = "https://dev.azure.com/[ORGANIZATION_NAME]"
  personal_access_token = "[PERSONAL_ACCESS_TOKEN]"
}