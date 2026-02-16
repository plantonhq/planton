---
title: "Premium Key Vault with Network Restrictions"
description: "This preset creates an Azure Key Vault with Premium SKU (HSM-backed keys), Azure RBAC, and network access control. Network ACLs restrict access to specific IP ranges and VNet subnets, with a..."
type: "preset"
rank: "02"
presetSlug: "02-premium-network-restricted"
componentSlug: "key-vault"
componentTitle: "Key Vault"
provider: "azure"
icon: "package"
order: 2
---

# Premium Key Vault with Network Restrictions

This preset creates an Azure Key Vault with Premium SKU (HSM-backed keys), Azure RBAC, and network access control. Network ACLs restrict access to specific IP ranges and VNet subnets, with a default-deny posture. This is the configuration for compliance-regulated workloads requiring FIPS 140-2 Level 3 validated key protection and strict network isolation.

## When to Use

- Compliance workloads (PCI-DSS, HIPAA, FedRAMP, financial services)
- Environments requiring HSM-backed keys for cryptographic operations
- Zero-trust architectures where Key Vault access must be restricted to specific networks

## Key Configuration Choices

- **Premium SKU** (`sku: PREMIUM`) -- HSM-backed keys with FIPS 140-2 Level 3 validation. Required for regulated industries
- **Network ACLs** (`networkAcls.defaultAction: DENY`) -- Denies all access except from specified IPs and VNet subnets. Azure trusted services bypass this restriction
- **IP rules** -- Allowlist for office IPs, VPN gateways, or CI/CD runner IPs
- **VNet subnet rules** -- Allowlist for specific Azure subnets using service endpoints
- **All other settings** match 01-standard-rbac (RBAC, purge protection, 90-day soft delete)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-office-cidr>` | Office or VPN IP range (e.g., `203.0.113.0/24`) | Your network team |
| `<your-subnet-resource-id>` | Full ARM resource ID of an allowed subnet | Azure portal or `AzureSubnet` status outputs |
| `<your-secret-name-1>` | Name of the first secret to create | Your application configuration |
| `<your-secret-name-2>` | Name of the second secret to create | Your application configuration |

## Related Presets

- **01-standard-rbac** -- Use instead for standard workloads without HSM or network isolation requirements
