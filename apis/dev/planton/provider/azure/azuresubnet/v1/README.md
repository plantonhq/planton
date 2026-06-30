# AzureSubnet

## Overview

`AzureSubnet` creates a subnet within an existing Azure Virtual Network (VNet).
Subnets partition a VNet's address space into segments for different workloads,
network tiers, or service delegations.

This is the most widely referenced Azure resource in Planton. Eleven downstream
resource types consume `subnet_id` from this component's outputs, making it a
critical building block in Azure infra charts.

Unlike the built-in `nodes_subnet` in AzureVpc (which is purpose-built for AKS),
AzureSubnet provides full control over service endpoints, delegations, and private
endpoint policies -- enabling multi-tier enterprise architectures.

**Note:** Subnets inherit their region from the parent VNet. This spec intentionally
omits the `region` field found on other Azure resources.

## Key Features

- **StringValueOrRef vnet_id** -- references an `AzureVpc` output for dependency wiring
- **Service endpoints** -- optimized routes over Azure backbone to PaaS services
- **Service delegation** -- grants Azure PaaS services permission to inject resources
  (PostgreSQL, MySQL, Container Apps, App Service)
- **Private endpoint policies** -- granular control over NSG/route table enforcement
  on private endpoints (4 policy modes)
- **Private Link Service policies** -- control network policy bypass for PLS
- **Composable outputs** -- exports `subnet_id` referenced by 11 downstream resources

## When to Use

- When an Azure workload needs a dedicated subnet (databases, app gateways, container apps)
- When you need service endpoints for secure PaaS access over Azure backbone
- When a PaaS service requires a delegated subnet (PostgreSQL Flexible Server,
  Container App Environment, App Service VNet integration)
- When building multi-tier network architectures (web, app, data, management tiers)
- As part of `database-stack`, `enterprise-network-foundation`, or
  `container-apps-environment` infra charts

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `resource_group` | StringValueOrRef | Yes | -- | Resource group (same as VNet) |
| `vnet_id` | StringValueOrRef | Yes | -- | Parent VNet ARM resource ID |
| `name` | string | Yes | -- | Subnet name (1-80 characters) |
| `address_prefix` | string | Yes | -- | IPv4 CIDR block |
| `service_endpoints` | repeated string | No | -- | Azure service endpoint names |
| `delegation` | AzureSubnetDelegation | No | -- | Service delegation config |
| `private_endpoint_network_policies` | string | No | Disabled | PE network policy mode |
| `private_link_service_network_policies_enabled` | bool | No | true | PLS network policies |

## Outputs

| Output | Description |
|--------|-------------|
| `subnet_id` | Azure Resource Manager ID (primary, referenced by 11 resource types) |
| `subnet_name` | Name of the subnet |
| `address_prefix` | CIDR block (useful for NSG rules) |

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureSubnet
metadata:
  name: app-subnet
  org: mycompany
  env: production
spec:
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  vnet_id:
    valueFrom:
      kind: AzureVpc
      name: prod-vpc
      fieldPath: status.outputs.vnet_id
  name: prod-app-subnet
  address_prefix: "10.0.1.0/24"
  service_endpoints:
    - Microsoft.Sql
    - Microsoft.Storage
```

## Downstream Usage

Other Azure resources reference this subnet via `StringValueOrRef`:

```yaml
# In an AzurePostgresqlFlexibleServer spec:
spec:
  delegated_subnet_id:
    valueFrom:
      kind: AzureSubnet
      name: db-subnet
      fieldPath: status.outputs.subnet_id
```

## What's NOT Included

- **Region** -- inherited from the parent VNet; including it would be misleading
- **NSG association** -- handled by AzureNetworkSecurityGroup (separate lifecycle)
- **Route table association** -- advanced networking, future iteration
- **NAT Gateway association** -- handled at VNet level by AzureVpc
- **Multiple address prefixes** -- Azure supports this for niche scenarios but 99.9%
  of subnets use a single CIDR
- **Service endpoint policies** -- advanced traffic restriction feature, very niche
- **Default outbound access** -- newer Azure feature for zero-trust, can be added later
