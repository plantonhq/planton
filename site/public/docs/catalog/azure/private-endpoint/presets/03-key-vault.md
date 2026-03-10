---
title: "Key Vault Private Endpoint"
description: "This preset creates a Private Endpoint for Azure Key Vault, enabling private network access to secrets, keys, and certificates. The `vault` subresource connects the endpoint to the Key Vault's data..."
type: "preset"
rank: "03"
presetSlug: "03-key-vault"
componentSlug: "private-endpoint"
componentTitle: "Private Endpoint"
provider: "azure"
icon: "package"
order: 3
---

# Key Vault Private Endpoint

This preset creates a Private Endpoint for Azure Key Vault, enabling private network access to secrets, keys, and certificates. The `vault` subresource connects the endpoint to the Key Vault's data plane. Combined with a Private DNS Zone for `privatelink.vaultcore.azure.net`, clients in the VNet resolve the Key Vault FQDN to its private IP — no public internet exposure.

## When to Use

- Key Vault instances storing production secrets, TLS certificates, or encryption keys
- Zero-trust architectures where secret access must not traverse the public internet
- AKS clusters, Function Apps, or Web Apps with managed identity accessing Key Vault over private networks
- Compliance requirements (PCI-DSS, HIPAA) mandating private secret storage access

## Key Configuration Choices

- **Subresource: vault** (`subresourceNames: [vault]`) -- Connects to the Key Vault's data plane (secrets, keys, certificates). This is the only subresource for Key Vault
- **Subnet** (`subnetId`) -- The private endpoint's NIC is placed in this subnet with a private IP. Use a dedicated "private-endpoints" subnet or a shared services subnet
- **Private DNS Zone** (`privateDnsZoneId`) -- Must be `privatelink.vaultcore.azure.net` for Key Vault DNS resolution. Create with `AzurePrivateDnsZone`
- **Key Vault firewall** -- After creating the private endpoint, disable public access on the Key Vault (`networkAcls.defaultAction: Deny`) to ensure all access goes through the private endpoint

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the Key Vault's region) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-pe-name>` | Name for the private endpoint (e.g., "pe-kv-prod") | Choose a descriptive name |
| `<subnet-resource-id>` | ARM resource ID of the subnet | `AzureSubnet` status outputs |
| `<key-vault-resource-id>` | ARM resource ID of the Key Vault | `AzureKeyVault` status outputs |
| `<private-dns-zone-id>` | ARM resource ID of the `privatelink.vaultcore.azure.net` DNS zone | `AzurePrivateDnsZone` status outputs |

## Related Presets

- **01-sql-server** -- Private Endpoint for Azure SQL Server
- **02-storage-account** -- Private Endpoint for Azure Storage Account blob service
