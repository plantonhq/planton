# Multi-Role Application Identity

This preset creates an Azure User-Assigned Managed Identity with role assignments for Key Vault, Storage, and Container Registry access. This is the standard pattern for production application workloads (AKS pods, Container Apps) that need to access multiple Azure services using a single credential-free identity.

## When to Use

- Application workloads that read secrets, access storage, and pull container images
- AKS pods with workload identity needing access to multiple services
- Container Apps or Function Apps with broad Azure service integration

## Key Configuration Choices

- **Key Vault Secrets User** -- Read-only access to secrets in a specific Key Vault
- **Storage Blob Data Contributor** -- Read/write/delete blobs in a specific Storage Account
- **AcrPull** -- Pull container images from a specific Azure Container Registry
- **Resource-scoped roles** -- Each role is scoped to a specific resource ID (not subscription-wide), following the principle of least privilege

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-identity-name>` | Name for the managed identity (3-128 chars) | Your naming convention |
| `<key-vault-resource-id>` | Full ARM resource ID of the Key Vault | Azure portal or `AzureKeyVault` status outputs |
| `<storage-account-resource-id>` | Full ARM resource ID of the Storage Account | Azure portal or `AzureStorageAccount` status outputs |
| `<acr-resource-id>` | Full ARM resource ID of the Container Registry | Azure portal or `AzureContainerRegistry` status outputs |

## Related Presets

- **01-standard** -- Use instead when the identity only needs one permission
