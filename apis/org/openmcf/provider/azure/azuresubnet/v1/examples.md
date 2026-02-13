# AzureSubnet Examples

## Minimal Configuration

The simplest possible subnet -- a name, resource group, VNet, and CIDR block.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: my-subnet
spec:
  resource_group: my-rg
  vnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet
  name: my-subnet
  address_prefix: "10.0.1.0/24"
```

## With Service Endpoints

A subnet with service endpoints for secure access to Azure SQL and Storage over
the Azure backbone network (bypassing the public internet).

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: app-subnet
  org: mycompany
  env: production
spec:
  resource_group: prod-network-rg
  vnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
  name: prod-app-subnet
  address_prefix: "10.0.1.0/24"
  service_endpoints:
    - Microsoft.Sql
    - Microsoft.Storage
    - Microsoft.KeyVault
```

## Delegated Subnet for PostgreSQL Flexible Server

A subnet delegated to Azure PostgreSQL Flexible Server. Delegated subnets are
dedicated to a single service and cannot host other resource types.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: pg-subnet
  org: mycompany
  env: production
spec:
  resource_group: prod-network-rg
  vnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
  name: prod-pg-subnet
  address_prefix: "10.0.2.0/24"
  delegation:
    name: postgresql
    service_name: Microsoft.DBforPostgreSQL/flexibleServers
    actions:
      - Microsoft.Network/virtualNetworks/subnets/action
```

## Delegated Subnet for Container App Environment

A subnet for Azure Container App Environment. Container Apps require a /23 or
larger subnet when using VNet integration.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: cae-subnet
  org: mycompany
  env: production
spec:
  resource_group: prod-network-rg
  vnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
  name: prod-container-apps-subnet
  address_prefix: "10.0.4.0/23"
  delegation:
    name: container-apps
    service_name: Microsoft.App/environments
```

## Private Endpoint Subnet

A subnet configured for private endpoints with network policies disabled (the
Azure default). Private endpoints allow PaaS services to be accessed via private
IP addresses within the VNet.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: pe-subnet
  org: mycompany
  env: production
spec:
  resource_group: prod-network-rg
  vnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
  name: prod-private-endpoints-subnet
  address_prefix: "10.0.3.0/24"
  private_endpoint_network_policies: Disabled
```

## Zero-Trust Private Endpoint Subnet

A subnet with NSG-enforced policies on private endpoints. Use this in zero-trust
architectures where all traffic -- including private endpoint traffic -- must pass
through NSG rules.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: zt-pe-subnet
  org: mycompany
  env: production
spec:
  resource_group: prod-network-rg
  vnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
  name: prod-zt-pe-subnet
  address_prefix: "10.0.8.0/24"
  private_endpoint_network_policies: NetworkSecurityGroupEnabled
```

## Application Gateway Subnet

A dedicated subnet for Azure Application Gateway. App Gateway requires a /27 or
larger dedicated subnet with no other resources.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: appgw-subnet
  org: mycompany
  env: production
spec:
  resource_group: prod-network-rg
  vnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
  name: prod-appgw-subnet
  address_prefix: "10.0.9.0/27"
```

## Infra Chart Wiring: Multi-Tier Enterprise Network

This example shows how AzureSubnet fits into an enterprise network architecture
with proper `valueFrom` wiring between resources.

### Resource Group (Layer 0)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: network-rg
  org: mycompany
  env: production
spec:
  name: prod-network-rg
  region: eastus
```

### VNet (Layer 1)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureVpc
metadata:
  name: prod-vpc
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  address_space_cidr: "10.0.0.0/16"
  nodes_subnet_cidr: "10.0.0.0/18"
```

### App Tier Subnet (Layer 2)

```yaml
apiVersion: azure.openmcf.org/v1
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
  address_prefix: "10.0.64.0/18"
  service_endpoints:
    - Microsoft.Sql
    - Microsoft.Storage
```

### Database Tier Subnet (Layer 2)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: db-subnet
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
  name: prod-db-subnet
  address_prefix: "10.0.128.0/24"
  delegation:
    name: postgresql
    service_name: Microsoft.DBforPostgreSQL/flexibleServers
    actions:
      - Microsoft.Network/virtualNetworks/subnets/action
```

### PostgreSQL (Layer 3) -- references subnet

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: prod-pg
  org: mycompany
  env: production
spec:
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  delegated_subnet_id:
    valueFrom:
      kind: AzureSubnet
      name: db-subnet
      fieldPath: status.outputs.subnet_id
  # ... other fields
```

## Subnet Sizing Guide

| Use Case | Recommended Size | Usable IPs | Notes |
|----------|-----------------|------------|-------|
| AKS Nodes | /18 | 16,379 | Large pool for pods+nodes |
| App Services / Functions | /24 | 251 | One IP per integrated app |
| PostgreSQL / MySQL Flexible | /24 | 251 | One IP per server |
| Container App Environment | /23 | 507 | Requires /23 minimum |
| Application Gateway | /27 | 27 | Minimum recommended |
| Private Endpoints | /24 | 251 | One IP per endpoint |
| Management / Bastion | /26 | 59 | Small admin subnet |

**Remember:** Azure reserves 5 IPs per subnet (first 4 + last) for internal services.

## Best Practices

1. **Plan CIDR ranges upfront** -- subnet CIDRs cannot overlap within a VNet and
   cannot be changed after creation without destroying the subnet

2. **Use service endpoints for PaaS access** -- they route traffic over Azure backbone
   instead of the public internet, improving security and latency

3. **Dedicate subnets for delegated services** -- PostgreSQL, MySQL, and Container
   Apps require exclusive use of their delegated subnet

4. **Size App Gateway subnets generously** -- while /27 is the minimum, /26 or /25
   provides room for autoscaling

5. **Use private endpoint subnets for data services** -- combine with
   AzurePrivateDnsZone for private DNS resolution within the VNet
