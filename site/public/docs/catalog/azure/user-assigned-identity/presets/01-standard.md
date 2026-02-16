---
title: "Standard Managed Identity"
description: "This preset creates an Azure User-Assigned Managed Identity with a single RBAC role assignment. This is the most common pattern -- a single identity with one targeted permission grant, used for..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "user-assigned-identity"
componentTitle: "User Assigned Identity"
provider: "azure"
icon: "package"
order: 1
---

# Standard Managed Identity

This preset creates an Azure User-Assigned Managed Identity with a single RBAC role assignment. This is the most common pattern -- a single identity with one targeted permission grant, used for workloads like AKS pods, Container Apps, or Function Apps that need to access a single Azure resource (typically Key Vault).

## When to Use

- An application that needs to read secrets from a Key Vault
- AKS workload identity bindings that need a single permission
- Any workload needing credential-free authentication to one Azure resource

## Key Configuration Choices

- **Single role assignment** -- One identity with one permission. The simplest and most auditable pattern
- **Key Vault Secrets User** (`roleDefinitionName: Key Vault Secrets User`) -- Read-only access to Key Vault secrets. Change to the appropriate role for other resources (see 02-multi-role for examples)
- **Scoped to a specific resource** -- The role is granted on a single resource ID, not a subscription or resource group. This follows the principle of least privilege

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-identity-name>` | Name for the managed identity (3-128 chars) | Your naming convention |
| `<target-resource-id>` | Full ARM resource ID of the target resource | Azure portal or target resource's status outputs |

## Related Presets

- **02-multi-role** -- Use instead when a single identity needs permissions across multiple resources
