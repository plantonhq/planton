# Azure Subnet

Deploys a standalone Azure Subnet within an existing Virtual Network -- the composition hub of Azure networking. This is the most widely referenced Azure networking resource: downstream consumers include AKS clusters, Container App environments, PostgreSQL/MySQL Flexible Servers, Redis Cache, Application Gateways, Load Balancers, private endpoints, and VMs, all of which deploy INTO a subnet by referencing its ID. The subnet also declares the attachments that make a segment production-grade -- a network security group, a route table, and a NAT gateway -- because that is Azure's own model: one of each serves many subnets.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Subnet** -- a subnet within the referenced VNet using self-planned address prefixes or a delegated Azure Network Manager IPAM allocation
- **Attachments** -- associations to a route table, network security group, and/or NAT gateway when the corresponding IDs are provided
- **Service Endpoints** -- created only when `serviceEndpoints` entries are provided; optimized routes over the Azure backbone to specified Azure services (Storage, SQL, Key Vault, etc.), optionally narrowed by `serviceEndpointPolicyIds`
- **Service Delegations** -- created only when `delegations` is configured; each hands the subnet to an Azure PaaS service (PostgreSQL Flexible Server, Container Apps, etc.), permitting it to inject service-managed resources
- **Network Policies** -- private endpoint / Private Link Service network policy settings, the default outbound access posture, and the cross-tenant sharing scope

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Virtual Network** -- the subnet is an ARM child of an existing VNet. Provide the VNet's ARM ID directly or reference an AzureVirtualNetwork Infra Component via ValueFromRef. The network's ID carries the resource group and region, so the subnet models neither.
- **CIDR planning** -- every `addressPrefixes` block must fall within the VNet's address space and must not overlap other subnets. Azure reserves 5 IPs per subnet for internal use.

## Deploy

### Console

Open the deployment store, find **Azure Subnet**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **General-Purpose Subnet** preset in the [Presets](#presets) tab to pre-populate a /24 subnet with common service endpoints.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureSubnet
metadata:
  name: app-subnet
  org: acme-corp
  env: prod
spec:
  virtualNetworkId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network/providers/Microsoft.Network/virtualNetworks/prod-vnet
  name: app
  addressPrefixes:
    - "10.0.1.0/24"
  serviceEndpoints:
    - "Microsoft.Storage"
    - "Microsoft.KeyVault"
```

```shell
planton apply -f azure-subnet.yaml
```

This creates a /24 subnet with service endpoints for Storage and Key Vault. No delegation, attachments, or custom network policies are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the subnet to a VNet -- and its production attachments to a NAT gateway or NSG -- deployed in the same InfraPipeline:

```yaml
spec:
  virtualNetworkId:
    valueFrom:
      kind: AzureVirtualNetwork
      name: production-vnet
      fieldPath: status.outputs.virtual_network_id
  natGatewayId:
    valueFrom:
      kind: AzureNatGateway
      name: production-nat
      fieldPath: status.outputs.nat_gateway_id
```

The InfraPipeline resolves the dependency graph, deploys the network and gateway first, then provisions the subnet with the resolved values.

## Key Configuration

These are the most important decisions when configuring a subnet. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Address source** -- exactly one of `addressPrefixes` (self-planned CIDR blocks) or `ipAddressPool` (delegated allocation from an Azure Network Manager IPAM pool). A `/24` provides 251 usable IPs (standard for most workloads), `/27` is the minimum for Application Gateway, and `/21` or larger is recommended for Container App environments.

**Attachments** -- `networkSecurityGroupId` filters the subnet's traffic (Azure's primary network access control), `routeTableId` replaces default routing with user-defined routes (the firewall-egress seam), and `natGatewayId` gives every workload stable SNAT-backed egress. All three are optional, updatable in place, and declared subnet-side because one NSG/table/gateway serves many subnets.

**Outbound posture** -- `defaultOutboundAccessEnabled` controls Azure's implicit outbound internet access, which Microsoft is retiring for new subnets: production subnets should set it explicitly `false` and route egress through the NAT gateway, load balancer outbound rules, or a firewall route.

**Service endpoints and delegations** -- `serviceEndpoints` creates optimized backbone routes to Azure PaaS services (`Microsoft.Storage`, `Microsoft.Sql`, `Microsoft.KeyVault`, ...). `delegations` dedicates the subnet to one PaaS service -- PostgreSQL Flexible Server requires `Microsoft.DBforPostgreSQL/flexibleServers`, Container Apps requires `Microsoft.App/environments`; a delegated subnet cannot be shared with other resource types.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureVirtualNetwork** | `virtualNetworkId` | `status.outputs.virtual_network_id` |
| **AzureRouteTable** (optional) | `routeTableId` | `status.outputs.route_table_id` |
| **AzureNetworkSecurityGroup** (optional) | `networkSecurityGroupId` | `status.outputs.network_security_group_id` |
| **AzureNatGateway** (optional) | `natGatewayId` | `status.outputs.nat_gateway_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `subnet_id` | Azure resource ID of the subnet | AKS clusters, Container App environments, PostgreSQL/MySQL Flexible Servers, Redis Cache, Load Balancers, Application Gateways, Private Endpoints, VMs |
| `address_prefixes` | CIDR blocks actually assigned (echoes the spec for self-managed subnets; carries the IPAM-provisioned ranges otherwise) | NSG rules, firewall rules, network planning |
| `virtual_network_name` | Parent network's name, derived from its ARM ID | Composing sibling resources without re-parsing the ID |
| `resource_group_name` | Resource group of the subnet and its parent network | Composing sibling resources without re-parsing the ID |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**General-purpose subnet** -- A /24 subnet with service endpoints for Storage, Key Vault, and SQL. No delegation, keeping the subnet flexible for VMs, load balancers, private endpoints, and other resource types. Start from the **General-Purpose Subnet** preset.

**PostgreSQL delegated subnet** -- A /28 subnet delegated to `Microsoft.DBforPostgreSQL/flexibleServers` for VNet-integrated PostgreSQL deployments. The delegation makes the subnet exclusive to PostgreSQL Flexible Server. Start from the **PostgreSQL Delegated Subnet** preset.

**Container Apps delegated subnet** -- A /21 subnet delegated to `Microsoft.App/environments` with 2,048 IPs for Container App scale-out. Required for VNet-integrated Container App environments. Start from the **Delegated Container Apps Subnet** preset.

## Works With

- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) -- the parent network the subnet partitions
- [**Azure Network Security Group**](/infra-catalog/azure-network-security-group) -- filters the subnet's traffic when attached
- [**Azure Route Table**](/infra-catalog/azure-route-table) -- steers the subnet's egress when attached
- [**Azure NAT Gateway**](/infra-catalog/azure-nat-gateway) -- owns the subnet's outbound connectivity when attached
