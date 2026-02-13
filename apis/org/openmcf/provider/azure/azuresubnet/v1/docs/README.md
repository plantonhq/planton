# AzureSubnet: Research & Design Documentation

## 1. What Is an Azure Subnet?

An Azure Subnet is a range of IP addresses within a Virtual Network (VNet). Subnets
segment the VNet's address space into logical sections, each hosting a different tier
of resources. Every Azure resource deployed into a VNet must be placed in a subnet.

Subnets are Azure's fundamental network partitioning mechanism. They provide:
- **Address isolation** -- different workloads get different IP ranges
- **Security boundaries** -- NSGs can be associated per-subnet for traffic filtering
- **Service delegation** -- Azure PaaS services can inject resources into subnets
- **Service endpoints** -- optimized routes to Azure PaaS over the backbone network
- **Private endpoint hosting** -- private IPs for PaaS services within the VNet

### Key Properties

- **Address Prefix**: IPv4 CIDR block within the parent VNet's address space
- **Service Endpoints**: Optimized routes to Azure services (Storage, SQL, Key Vault)
- **Delegation**: Grants a PaaS service exclusive control of the subnet
- **Private Endpoint Policies**: Controls NSG/UDR enforcement on private endpoints
- **Lifecycle**: Independent from the VNet (can be added/removed without VNet changes)

### Azure's Reserved IPs

Azure reserves **5 IP addresses** per subnet for internal services:
- x.x.x.0: Network address
- x.x.x.1: Default gateway
- x.x.x.2-3: Azure DNS mapping
- x.x.x.255: Broadcast (last IP)

A /24 subnet (256 IPs) provides 251 usable addresses.

## 2. Deployment Landscape

### How People Deploy Subnets Today

#### Level 0: Azure Portal (Click-Ops)

The portal provides a GUI for adding subnets to VNets. Users navigate to the VNet
resource, click "Subnets", and add one. This creates undocumented, un-versioned
infrastructure that's impossible to reproduce.

#### Level 1: Azure CLI

```bash
az network vnet subnet create \
  --name app-subnet \
  --resource-group my-rg \
  --vnet-name my-vnet \
  --address-prefix 10.0.1.0/24 \
  --service-endpoints Microsoft.Sql Microsoft.Storage \
  --delegations Microsoft.DBforPostgreSQL/flexibleServers
```

Simple and scriptable but lacks state management and drift detection.

#### Level 2: ARM Templates / Bicep

```bicep
resource subnet 'Microsoft.Network/virtualNetworks/subnets@2023-09-01' = {
  name: 'app-subnet'
  parent: vnet
  properties: {
    addressPrefix: '10.0.1.0/24'
    serviceEndpoints: [{ service: 'Microsoft.Sql' }]
  }
}
```

Azure-native IaC with full lifecycle management. Good for Azure-only shops.

#### Level 3: Terraform

```hcl
resource "azurerm_subnet" "app" {
  name                 = "app-subnet"
  resource_group_name  = "my-rg"
  virtual_network_name = "my-vnet"
  address_prefixes     = ["10.0.1.0/24"]
  service_endpoints    = ["Microsoft.Sql", "Microsoft.Storage"]
}
```

The most popular IaC approach for multi-cloud teams.

#### Level 4: Pulumi

```go
subnet, _ := network.NewSubnet(ctx, "app-subnet", &network.SubnetArgs{
    Name:               pulumi.String("app-subnet"),
    ResourceGroupName:  pulumi.String("my-rg"),
    VirtualNetworkName: pulumi.String("my-vnet"),
    AddressPrefixes:    pulumi.StringArray{pulumi.String("10.0.1.0/24")},
    ServiceEndpoints:   pulumi.StringArray{pulumi.String("Microsoft.Sql")},
})
```

Programmatic IaC with type safety. Good for complex conditional logic.

#### Level 5: OpenMCF (This Component)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: app-subnet
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
  name: app-subnet
  address_prefix: "10.0.1.0/24"
  service_endpoints:
    - Microsoft.Sql
    - Microsoft.Storage
```

Declarative, Kubernetes-style API that enables infra chart composition with
`StringValueOrRef` dependency wiring.

## 3. 80/20 Analysis: What We Include and What We Skip

### Included (80% of Use Cases)

| Feature | Rationale |
|---------|-----------|
| Address prefix (singular) | 99.9% of subnets use a single CIDR |
| Service endpoints | Common for secure PaaS access |
| Service delegation | Required for PostgreSQL, MySQL, Container Apps, App Service |
| Private endpoint network policies | Enterprise private link architectures |
| Private Link Service policies | Required for PLS providers |

### Excluded (20% Niche / Advanced)

| Feature | Rationale |
|---------|-----------|
| Multiple address prefixes | Azure supports this but it's extremely rare |
| Service endpoint policies | Advanced traffic restriction, very niche |
| Default outbound access | Newer feature for zero-trust; can add later |
| IP address pool (IPAM) | Azure IPAM integration; enterprise-only feature |
| Sharing scope | Multi-tenant subnet sharing; preview feature |
| NSG association | Separate lifecycle, handled by AzureNetworkSecurityGroup |
| Route table association | Advanced networking, future iteration |
| NAT Gateway association | Handled at VNet level by AzureVpc |

## 4. Why No Region Field

Unlike every other Azure resource in OpenMCF, AzureSubnet deliberately omits
the `region` field. Here's why:

1. **Subnets don't have their own region** -- they inherit from the parent VNet
2. **Azure's API doesn't accept a region parameter** for subnet creation
3. **Neither Terraform nor Pulumi accept a region** on the subnet resource
4. **Including it would be misleading** -- users could provide a different region
   than the VNet, which would be silently ignored or fail

This is a deliberate, well-reasoned deviation from the standard Azure pattern.

## 5. VNet ID Reference Design

AzureSubnet references the parent VNet via `StringValueOrRef vnet_id`, which
resolves to the VNet's ARM resource ID. The IaC modules extract the VNet name
from the ARM ID using string parsing:

```
ARM ID: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/virtualNetworks/{name}
VNet Name: last segment after splitting by "/"
```

This follows the same pattern as GcpSubnetwork, which references `vpc_self_link`
(a full GCP URL) and parses it for the VPC name.

## 6. Downstream Consumers (11 Resource Types)

AzureSubnet is the most widely referenced Azure resource. These resource types
consume `subnet_id`:

| Resource | Use Case |
|----------|----------|
| AzureAksCluster | AKS node pool placement |
| AzureContainerAppEnvironment | Infrastructure subnet for Container Apps |
| AzurePostgresqlFlexibleServer | Delegated subnet for private access |
| AzureMysqlFlexibleServer | Delegated subnet for private access |
| AzureRedisCache | Premium SKU VNet injection |
| AzurePrivateEndpoint | Private IP for PaaS services |
| AzureApplicationGateway | Dedicated subnet for App GW |
| AzureLoadBalancer | Internal LB frontend placement |
| AzureVirtualMachine | VM NIC placement |
| AzureFunctionApp | VNet integration subnet |
| AzureLinuxWebApp | VNet integration subnet |

## 7. Infra Chart Integration

### Database Stack

```
AzureResourceGroup (Layer 0)
├── AzureVpc (Layer 1)
│   ├── AzureSubnet [db-delegated] (Layer 2)  <-- THIS RESOURCE
│   │   └── AzurePostgresqlFlexibleServer (Layer 3)
│   ├── AzureSubnet [pe-subnet] (Layer 2)
│   │   └── AzurePrivateEndpoint (Layer 3)
│   └── AzureSubnet [redis-subnet] (Layer 2)
│       └── AzureRedisCache (Layer 3)
└── AzurePrivateDnsZone (Layer 1)
```

### Enterprise Network Foundation

```
AzureResourceGroup (Layer 0)
├── AzureVpc (Layer 1)
│   ├── AzureSubnet [app-gw] (Layer 2)       <-- App Gateway dedicated
│   ├── AzureSubnet [app-tier] (Layer 2)      <-- Application workloads
│   ├── AzureSubnet [db-tier] (Layer 2)       <-- Database delegations
│   ├── AzureSubnet [pe-tier] (Layer 2)       <-- Private endpoints
│   └── AzureSubnet [mgmt] (Layer 2)          <-- Management/bastion
├── AzurePublicIp (Layer 1)
└── AzureLogAnalyticsWorkspace (Layer 1)
```

### Container Apps Environment

```
AzureResourceGroup (Layer 0)
├── AzureVpc (Layer 1)
│   └── AzureSubnet [cae-infra] (Layer 2)     <-- /23 for Container Apps
│       └── AzureContainerAppEnvironment (Layer 3)
│           └── AzureContainerApp (Layer 4)
└── AzureLogAnalyticsWorkspace (Layer 1)
```

## 8. Design Decisions

### Why Standalone Resource (DD01)

AzureSubnet exists as a standalone resource rather than being embedded in AzureVpc
because subnets have independent lifecycles. Different subnets need different
configurations (delegations, service endpoints, policies), and bundling them all
into the VPC spec would force VPC redeployment when a subnet changes. See DD01.

### Why Singular address_prefix

Azure's API supports `address_prefixes` (plural), but multiple CIDRs per subnet
is an extremely rare edge case introduced for backwards-compatibility with legacy
configurations. Using singular `address_prefix` keeps the spec clean and matches
the 80/20 principle. The IaC modules wrap it in a single-element list for the provider.

### Why private_endpoint_network_policies as String (Not Bool)

Azure deprecated the boolean field `private_endpoint_network_policies_enabled` in
favor of a string enum with four values: `Disabled`, `Enabled`,
`NetworkSecurityGroupEnabled`, and `RouteTableEnabled`. The string enum provides
granular control -- you can apply NSG policies without route table policies, or
vice versa. Using the current API surface prevents future deprecation issues.

## 9. Scope Boundaries

### What This Component Does

- Creates a subnet in an existing VNet
- Configures service endpoints for optimized PaaS access
- Configures service delegation for PaaS resource injection
- Sets private endpoint and Private Link Service network policies
- Tags the subnet with OpenMCF metadata (via parent VNet tagging)
- Exports subnet_id, subnet_name, and address_prefix for downstream consumption

### What This Component Does NOT Do

- **NSG association** -- handled by AzureNetworkSecurityGroup
- **Route table association** -- future iteration
- **NAT Gateway association** -- handled by AzureVpc
- **Multiple address prefixes** -- niche feature, not included
- **Create the parent VNet** -- VNet must exist first (AzureVpc)
