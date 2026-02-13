---
title: "Privateendpoint"
description: "Privateendpoint deployment documentation"
icon: "package"
order: 100
componentName: "azureprivateendpoint"
---

# Azure Private Endpoint

Deploys an Azure Private Endpoint that provides secure, private connectivity to Azure PaaS services over the Microsoft backbone network using a private IP address from your VNet. The component optionally creates a Private DNS Zone Group to automatically register the private endpoint's IP as an A-record in a linked private DNS zone.

## What Gets Created

When you deploy an AzurePrivateEndpoint resource, OpenMCF provisions:

- **Private Endpoint** — a `privatelink.Endpoint` resource in the specified region and resource group, with an auto-approved private service connection to the target Azure service and a private IP allocated from the designated subnet
- **Private Service Connection** — an auto-approved connection from the endpoint to the target resource, named `{metadata.name}-connection`, mapping one or more sub-resource group IDs
- **Private DNS Zone Group** (conditional) — when `privateDnsZoneId` is provided, a DNS zone group named `{metadata.name}-dns-zone-group` that registers the private endpoint's IP as an A-record in the specified private DNS zone, ensuring in-VNet DNS resolution routes to the private IP instead of the public one
- **Azure Tags** — resource metadata tags applied to the endpoint for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the endpoint will be created (can reference an AzureResourceGroup resource)
- **A Subnet** in a VNet from which the private IP will be allocated; the subnet must have private endpoint network policies configured appropriately
- **A Private Link-enabled target resource** — the Azure PaaS service or custom Private Link Service to connect to (e.g., PostgreSQL Flexible Server, Key Vault, Storage Account)
- **A Private DNS Zone** (optional) — required if you want automatic DNS registration of the private endpoint's IP address

## Quick Start

Create a file `private-endpoint.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: my-pe
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzurePrivateEndpoint.my-pe
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-private-endpoint
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet
  privateConnectionResourceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-postgres
  subresourceNames:
    - postgresqlServer
```

Deploy:

```shell
openmcf apply -f private-endpoint.yaml
```

This creates a private endpoint in the specified subnet with an auto-approved connection to the target PostgreSQL Flexible Server, allocating a private IP from the subnet.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the Private Endpoint (e.g., `eastus`, `westeurope`). Must be in the same region as the subnet. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Name of the Private Endpoint. Must be unique within the resource group. Allows letters, numbers, underscores, periods, and hyphens; must start and end with an alphanumeric character. | Required, 1-80 characters |
| `subnetId` | `StringValueOrRef` | Azure Resource Manager ID of the subnet from which a private IP will be allocated. Can reference an AzureSubnet resource via `valueFrom`. | Required |
| `privateConnectionResourceId` | `StringValueOrRef` | Azure Resource Manager ID of the Private Link-enabled target resource. Can reference any OpenMCF resource that supports Private Link via `valueFrom` (polymorphic -- no default kind). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `subresourceNames` | `string[]` | `[]` | Sub-resource group IDs that the endpoint connects to. Common values: `postgresqlServer`, `mysqlServer`, `sqlServer`, `vault`, `blob`, `table`, `queue`, `file`, `Sql` (Cosmos DB SQL API), `MongoDB` (Cosmos DB Mongo API), `redisCache`, `registry`. Most endpoints use exactly one sub-resource. |
| `privateDnsZoneId` | `StringValueOrRef` | not set | Azure Resource Manager ID of a Private DNS Zone for automatic A-record registration. Can reference an AzurePrivateDnsZone resource via `valueFrom`. When omitted, no DNS zone group is created. |

## Examples

### Private Endpoint for PostgreSQL with DNS

A private endpoint connecting to a PostgreSQL Flexible Server with automatic DNS registration in the corresponding privatelink zone:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: postgres-pe
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePrivateEndpoint.postgres-pe
spec:
  region: eastus
  resourceGroup: prod-rg
  name: postgres-private-endpoint
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/pe-subnet
  privateConnectionResourceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/prod-postgres
  subresourceNames:
    - postgresqlServer
  privateDnsZoneId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/privateDnsZones/privatelink.postgres.database.azure.com
```

### Private Endpoint for Key Vault without DNS

A private endpoint for Key Vault where DNS is managed externally:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: vault-pe
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePrivateEndpoint.vault-pe
spec:
  region: eastus
  resourceGroup: prod-rg
  name: vault-private-endpoint
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/pe-subnet
  privateConnectionResourceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.KeyVault/vaults/prod-vault
  subresourceNames:
    - vault
```

### Private Endpoint for Storage Account Blob

A private endpoint for Azure Blob Storage with DNS zone group:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: blob-pe
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePrivateEndpoint.blob-pe
spec:
  region: westeurope
  resourceGroup: data-rg
  name: blob-private-endpoint
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/data-rg/providers/Microsoft.Network/virtualNetworks/data-vnet/subnets/pe-subnet
  privateConnectionResourceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/data-rg/providers/Microsoft.Storage/storageAccounts/prodstorage
  subresourceNames:
    - blob
  privateDnsZoneId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/data-rg/providers/Microsoft.Network/privateDnsZones/privatelink.blob.core.windows.net
```

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding Azure resource IDs. The `privateConnectionResourceId` field is polymorphic and can reference any resource kind that supports Private Link:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: ref-pe
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePrivateEndpoint.ref-pe
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-private-endpoint
  subnetId:
    valueFrom:
      kind: AzureSubnet
      name: pe-subnet
      field: status.outputs.subnet_id
  privateConnectionResourceId:
    valueFrom:
      kind: AzurePostgresqlFlexibleServer
      name: prod-postgresql
      field: status.outputs.server_id
  subresourceNames:
    - postgresqlServer
  privateDnsZoneId:
    valueFrom:
      kind: AzurePrivateDnsZone
      name: pg-dns-zone
      field: status.outputs.zone_id
```

### Multiple Sub-Resources for Cosmos DB

A private endpoint connecting to Azure Cosmos DB using the SQL API sub-resource:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: cosmos-pe
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePrivateEndpoint.cosmos-pe
spec:
  region: eastus
  resourceGroup: app-rg
  name: cosmos-private-endpoint
  subnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Network/virtualNetworks/app-vnet/subnets/pe-subnet
  privateConnectionResourceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.DocumentDB/databaseAccounts/prod-cosmos
  subresourceNames:
    - Sql
  privateDnsZoneId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Network/privateDnsZones/privatelink.documents.azure.com
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `private_endpoint_id` | `string` | Azure Resource Manager ID of the Private Endpoint |
| `private_ip_address` | `string` | Private IP address allocated from the subnet. This is the IP that the target service's FQDN should resolve to within the VNet. If a DNS zone group is configured, this IP is automatically registered as an A-record. |
| `network_interface_id` | `string` | Azure Resource Manager ID of the network interface created for this Private Endpoint. Useful for advanced networking diagnostics. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group for endpoint placement
- [AzureVpc](/docs/catalog/azure/azurevpc) -- provides the VNet containing the subnet for private IP allocation
- [AzureSubnet](/docs/catalog/azure/azuresubnet) -- provides the subnet from which the private endpoint's IP is allocated
- [AzurePrivateDnsZone](/docs/catalog/azure/azureprivatednszone) -- provides the DNS zone for automatic A-record registration of the private endpoint's IP
- [AzurePostgresqlFlexibleServer](/docs/catalog/azure/azurepostgresqlflexibleserver) -- a common Private Link-enabled target for database workloads
- [AzureKeyVault](/docs/catalog/azure/azurekeyvault) -- a common Private Link-enabled target for secrets management
- [AzureStorageAccount](/docs/catalog/azure/azurestorageaccount) -- a common Private Link-enabled target for blob, table, queue, and file storage
