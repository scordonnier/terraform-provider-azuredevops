---
page_title: "Azure DevOps provider"
description: |-
  
---

# Azure DevOps provider

The Azure DevOps provider can be used to configure Azure DevOps projects using [Azure DevOps Services REST API](https://learn.microsoft.com/en-us/rest/api/azure/devops/).

## Example Usage

```terraform
terraform {
  required_providers {
    azuredevops = {
      source  = "scordonnier/azuredevops"
      version = "0.4.0"
    }
  }
}

provider "azuredevops" {
  organization_url      = "https://dev.azure.com/[ORGANIZATION_NAME]"
  personal_access_token = "[PERSONAL_ACCESS_TOKEN]"
}
```
