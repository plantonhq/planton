# AzureUserAssignedIdentity Terraform Module

## Overview

This Terraform module provisions an Azure User-Assigned Managed Identity with optional
RBAC role assignments using the `azurerm` provider. It creates a single
`azurerm_user_assigned_identity` and zero or more `azurerm_role_assignment` resources.

## Resources Created

- `azurerm_user_assigned_identity.main` -- the managed identity
- `azurerm_role_assignment.main` (N) -- one per role assignment in the spec

## Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | OpenMCF metadata (name, org, env) |
| `spec` | object | Identity specification (region, resource_group, name, role_assignments) |

## Outputs

| Output | Description |
|--------|-------------|
| `identity_id` | Azure Resource Manager ID |
| `principal_id` | Service Principal Object ID |
| `client_id` | Application/Client ID |
| `tenant_id` | Azure AD Tenant ID |

## Usage

```hcl
module "identity" {
  source = "./iac/tf"

  metadata = {
    name = "platform-identity"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region         = "eastus"
    resource_group = "prod-identity-rg"
    name           = "prod-platform-identity"
    role_assignments = [
      {
        scope                = "/subscriptions/.../providers/Microsoft.KeyVault/vaults/prod-kv"
        role_definition_name = "Key Vault Secrets User"
      },
      {
        scope                = "/subscriptions/.../providers/Microsoft.Storage/storageAccounts/prodstorage"
        role_definition_name = "Storage Blob Data Contributor"
      }
    ]
  }
}
```
