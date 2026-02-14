---
title: "Private DNS Zone"
description: "Private DNS Zone deployment documentation"
icon: "package"
order: 100
componentName: "azureprivatednszone"
---

# Azure Private DNS Zone

Deploys an Azure Private DNS Zone with a Virtual Network link for internal name resolution. The component supports both Private Link DNS scenarios (resolving Azure PaaS service private endpoints) and custom internal DNS zones for VM hostname discovery within a VNet.

## What Gets Created

When you deploy an AzurePrivateDnsZone resource, OpenMCF provisions:

- **Private DNS Zone** — a `privatedns.Zone` resource in the specified resource group. Private DNS zones are global Azure resources with no region parameter.
- **Virtual Network Link** — a `privatedns.ZoneVirtualNetworkLink` that connects the zone to a VNet, enabling DNS resolution of zone records from resources within the linked VNet. Without this link the zone is unreachable.
- **Azure Tags** — resource metadata tags applied to both the zone and the VNet link for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the zone will be created (can reference an AzureResourceGroup resource)
- **A Virtual Network** to link to the zone (can reference an AzureVpc resource)
- **Zone name planning** — for Private Link scenarios, the zone name must match the Azure-defined privatelink zone name for the target service (e.g., `privatelink.postgres.database.azure.com` for PostgreSQL Flexible Server)

## Quick Start

Create a file `private-dns-zone.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: my-zone
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzurePrivateDnsZone.my-zone
spec:
  resourceGroup: my-rg
  name: privatelink.postgres.database.azure.com
  vnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet
```

Deploy:

```shell
openmcf apply -f private-dns-zone.yaml
```

This creates a Private DNS Zone for PostgreSQL Private Link resolution, linked to the specified VNet with auto-registration disabled.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | DNS zone name. For Private Link, must match the Azure-defined privatelink zone name for the target service (e.g., `privatelink.postgres.database.azure.com`). For custom internal DNS, use any valid domain (e.g., `contoso.internal`). | Required, must be a valid DNS domain name |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `vnetId` | `StringValueOrRef` | Azure Resource Manager ID of the Virtual Network to link. Format: `/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{name}`. Can reference an AzureVpc resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `registrationEnabled` | `bool` | `false` | Enables auto-registration of VM DNS records in the linked VNet. When true, Azure automatically creates and removes A records for VMs in the linked VNet. Useful for custom internal DNS zones. Should remain false for Private Link zones, where DNS records are managed by the private endpoint resource. |

## Examples

### Private Link Zone for PostgreSQL

A Private DNS Zone for resolving PostgreSQL Flexible Server private endpoints:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: postgres-dns
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePrivateDnsZone.postgres-dns
spec:
  resourceGroup: prod-rg
  name: privatelink.postgres.database.azure.com
  vnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
```

### Private Link Zone for Key Vault

A Private DNS Zone enabling private connectivity to Azure Key Vault:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: keyvault-dns
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePrivateDnsZone.keyvault-dns
spec:
  resourceGroup: prod-rg
  name: privatelink.vaultcore.azure.net
  vnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
```

### Custom Internal DNS with Auto-Registration

An internal DNS zone for VM hostname discovery with auto-registration enabled:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: internal-dns
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzurePrivateDnsZone.internal-dns
spec:
  resourceGroup: dev-rg
  name: contoso.internal
  vnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/dev-rg/providers/Microsoft.Network/virtualNetworks/dev-vnet
  registrationEnabled: true
```

### Private Link Zone for Blob Storage

A Private DNS Zone for resolving Azure Blob Storage private endpoints:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: blob-dns
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePrivateDnsZone.blob-dns
spec:
  resourceGroup: prod-rg
  name: privatelink.blob.core.windows.net
  vnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
```

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: ref-dns
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePrivateDnsZone.ref-dns
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: privatelink.postgres.database.azure.com
  vnetId:
    valueFrom:
      kind: AzureVpc
      name: my-vpc
      field: status.outputs.vnet_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `zone_id` | `string` | Azure Resource Manager ID of the Private DNS Zone. Format: `/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/privateDnsZones/{name}`. Referenced by downstream resources via `StringValueOrRef`. |
| `zone_name` | `string` | Name of the Private DNS Zone (e.g., `privatelink.postgres.database.azure.com`). |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/resource-group) — provides the resource group for zone placement
- [AzureVpc](/docs/catalog/azure/vpc-virtual-network) — provides the Virtual Network to link to the zone
- [AzurePostgresqlFlexibleServer](/docs/catalog/azure/postgresql-flexible-server) — references `zone_id` for VNet-integrated deployment with private DNS resolution
- [AzureMysqlFlexibleServer](/docs/catalog/azure/mysql-flexible-server) — references `zone_id` for VNet-integrated deployment with private DNS resolution
- [AzurePrivateEndpoint](/docs/catalog/azure/private-endpoint) — references `zone_id` for DNS zone group registration, enabling private endpoint FQDN resolution
