---
title: "Standard Key Vault with RBAC"
description: "This preset creates an Azure Key Vault with Standard SKU, Azure RBAC authorization, purge protection, and 90-day soft delete retention. This is the recommended configuration for most production..."
type: "preset"
rank: "01"
presetSlug: "01-standard-rbac"
componentSlug: "key-vault"
componentTitle: "Key Vault"
provider: "azure"
icon: "package"
order: 1
---

# Standard Key Vault with RBAC

This preset creates an Azure Key Vault with Standard SKU, Azure RBAC authorization, purge protection, and 90-day soft delete retention. This is the recommended configuration for most production workloads -- RBAC provides fine-grained access control through Azure AD and is the modern replacement for vault access policies.

## When to Use

- Storing application secrets, API keys, connection strings, and certificates
- Environments where Azure RBAC is the primary authorization mechanism
- Standard production workloads without HSM compliance requirements

## Key Configuration Choices

- **Standard SKU** (`sku: STANDARD`) -- Software-protected keys and secrets. Sufficient for most applications
- **RBAC authorization** (`enableRbacAuthorization: true`) -- Uses Azure AD RBAC roles (e.g., "Key Vault Secrets User") instead of legacy vault access policies
- **Purge protection** (`enablePurgeProtection: true`) -- Prevents permanent deletion of the vault during the soft-delete retention period
- **90-day soft delete** (`softDeleteRetentionDays: 90`) -- Maximum retention for deleted secrets, keys, and certificates
- **No network restrictions** -- Network ACLs are not configured, allowing access from any network. Add `networkAcls` for network-isolated environments

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-secret-name-1>` | Name of the first secret to create (values set post-deployment) | Your application configuration |
| `<your-secret-name-2>` | Name of the second secret to create | Your application configuration |

## Related Presets

- **02-premium-network-restricted** -- Use instead for compliance workloads requiring HSM-backed keys and network isolation
