---
title: "Key Vault"
description: "Key Vault deployment documentation"
icon: "package"
order: 100
componentName: "azurekeyvault"
---

# Azure Key Vault

Deploys an Azure Key Vault with configurable SKU tier, RBAC authorization, purge protection, soft delete retention, and network access controls. The component optionally creates named secret placeholders whose values are set separately via Azure SDK or CLI.

## What Gets Created

When you deploy an AzureKeyVault resource, OpenMCF provisions:

- **Key Vault** — a `keyvault.KeyVault` resource in the specified region and resource group, configured with the chosen SKU tier, RBAC authorization, purge protection, and soft delete retention
- **Network ACLs** — default-deny network rules with optional IP allowlists, VNet subnet rules, and Azure trusted services bypass
- **Secrets** — created for each entry in `secretNames` as empty placeholder `keyvault.Secret` resources; actual values must be set separately via Azure SDK, CLI, or the Key Vault API
- **Azure Tags** — resource metadata tags applied to the vault and all secrets for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the vault will be created (can reference an AzureResourceGroup resource)
- **Network planning** — know which IP ranges and/or VNet subnets need vault access if restricting with network ACLs

## Quick Start

Create a file `keyvault.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureKeyVault
metadata:
  name: my-vault
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureKeyVault.my-vault
spec:
  region: eastus
  resourceGroup: my-rg
```

Deploy:

```shell
openmcf apply -f keyvault.yaml
```

This creates a Standard-tier Key Vault with RBAC authorization enabled, purge protection on, 90-day soft delete retention, and default-deny network ACLs that bypass Azure trusted services.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the Key Vault (e.g., `eastus`, `westeurope`). | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sku` | `enum` | `STANDARD` | Key Vault SKU tier. Values: `STANDARD` (software-protected keys), `PREMIUM` (HSM-backed keys, required for PCI-DSS/FIPS 140-2 Level 3 compliance). |
| `enableRbacAuthorization` | `bool` | `true` | Enables Azure RBAC for authorization instead of vault access policies. Recommended for new deployments. |
| `enablePurgeProtection` | `bool` | `true` | Prevents permanent deletion of the vault during the soft delete retention period. Should always be `true` for production. |
| `softDeleteRetentionDays` | `int` | `90` | Retention period in days for deleted secrets, keys, and certificates. Range: 7–90. |
| `networkAcls.defaultAction` | `enum` | `DENY` | Default action when no rule matches. Values: `DENY` (recommended), `ALLOW`. |
| `networkAcls.bypassAzureServices` | `bool` | `true` | Allows traffic from trusted Azure services (Backup, Site Recovery, Monitor, etc.) even when default action is `DENY`. |
| `networkAcls.ipRules` | `string[]` | `[]` | IP addresses or CIDR ranges allowed to access the vault. Maximum 200 entries. |
| `networkAcls.virtualNetworkSubnetIds` | `string[]` | `[]` | Azure VNet subnet resource IDs allowed to access the vault. Maximum 100 entries. |
| `secretNames` | `string[]` | `[]` | Secret names to create as empty placeholders. Actual values must be set separately. Maximum 100 entries. |

## Examples

### Development Vault with Open Access

A vault for development with network ACLs set to allow all traffic:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureKeyVault
metadata:
  name: dev-vault
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureKeyVault.dev-vault
spec:
  region: eastus
  resourceGroup: dev-rg
  sku: STANDARD
  enablePurgeProtection: false
  softDeleteRetentionDays: 7
  networkAcls:
    defaultAction: ALLOW
```

### Production Vault with Secrets and Network Restrictions

A production vault with purge protection, restricted network access from office IPs and CI/CD runners, and pre-created secret placeholders:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureKeyVault
metadata:
  name: prod-vault
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureKeyVault.prod-vault
spec:
  region: eastus
  resourceGroup: prod-rg
  sku: STANDARD
  enableRbacAuthorization: true
  enablePurgeProtection: true
  softDeleteRetentionDays: 90
  networkAcls:
    defaultAction: DENY
    bypassAzureServices: true
    ipRules:
      - "203.0.113.0/24"
      - "198.51.100.42"
  secretNames:
    - db-connection-string
    - api-key
    - storage-account-key
```

### HSM-Backed Vault with VNet Integration

A Premium-tier vault with HSM-backed keys for compliance workloads, restricted to specific VNet subnets:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureKeyVault
metadata:
  name: compliance-vault
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureKeyVault.compliance-vault
spec:
  region: westeurope
  resourceGroup: compliance-rg
  sku: PREMIUM
  enableRbacAuthorization: true
  enablePurgeProtection: true
  softDeleteRetentionDays: 90
  networkAcls:
    defaultAction: DENY
    bypassAzureServices: true
    virtualNetworkSubnetIds:
      - /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/compliance-rg/providers/Microsoft.Network/virtualNetworks/compliance-vnet/subnets/app
      - /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/compliance-rg/providers/Microsoft.Network/virtualNetworks/compliance-vnet/subnets/data
  secretNames:
    - encryption-key
    - signing-key
    - tls-certificate
```

### Using Foreign Key References

Reference an OpenMCF-managed resource group instead of hardcoding the name:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureKeyVault
metadata:
  name: ref-vault
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureKeyVault.ref-vault
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  secretNames:
    - app-secret
    - db-password
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `vault_id` | `string` | Azure Resource Manager ID of the Key Vault |
| `vault_name` | `string` | Name of the Key Vault |
| `vault_uri` | `string` | URI of the Key Vault (e.g., `https://{vault-name}.vault.azure.net/`). Applications use this to access secrets, keys, and certificates. |
| `secret_id_map` | `map<string, string>` | Map of secret names to their full secret IDs. Only contains secrets created by this stack. |
| `region` | `string` | Azure region where the Key Vault was deployed |
| `resource_group` | `string` | Resource group name where the Key Vault was created |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/resource-group) — provides the resource group for vault placement
- [AzureAksCluster](/docs/catalog/azure/aks-cluster) — AKS clusters can mount vault secrets via the Key Vault CSI driver add-on
- [AzureVpc](/docs/catalog/azure/vpc-virtual-network) — provides VNet subnets for network ACL rules
- [AzureUserAssignedIdentity](/docs/catalog/azure/user-assigned-identity) — managed identities used to authenticate to the vault
