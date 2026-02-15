---
title: "External Secrets on AKS with Azure Key Vault"
description: "This preset deploys the External Secrets Operator (ESO) on an AKS cluster with Azure Key Vault as the backing store. Authentication uses a user-assigned managed identity bound to the ESO service..."
type: "preset"
rank: "03"
presetSlug: "03-aks-key-vault"
componentSlug: "external-secrets"
componentTitle: "External Secrets"
provider: "kubernetes"
icon: "package"
order: 3
---

# External Secrets on AKS with Azure Key Vault

This preset deploys the External Secrets Operator (ESO) on an AKS cluster with Azure Key Vault as the backing store. Authentication uses a user-assigned managed identity bound to the ESO service account.

## When to Use

- You run AKS and store secrets in Azure Key Vault
- You want Kubernetes Secrets to be automatically synced from Key Vault
- A managed identity with Key Vault access is available

## Key Configuration Choices

- **Managed Identity** -- ESO authenticates via a user-assigned managed identity; no client secrets stored in the cluster
- **Default poll interval** (10 seconds) -- how often ESO checks for secret changes

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-azure-key-vault-resource-id>` | Full Azure resource ID of the Key Vault | Azure Portal > Key Vault > Properties > Resource ID |
| `<your-managed-identity-client-id>` | Client ID of the managed identity with Key Vault access | Azure Portal > Managed Identities |

## Related Presets

- **01-gke-secrets-manager** -- Use on GKE with Google Cloud Secret Manager
- **02-eks-secrets-manager** -- Use on EKS with AWS Secrets Manager
