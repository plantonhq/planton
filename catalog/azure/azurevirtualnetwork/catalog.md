# Azure Virtual Network

Deploys an Azure Virtual Network (VNet) -- the isolated, private IP address space every network-attached Azure workload lives inside. The network is deliberately just the network: address planning and network-wide policy live here, while subnets, NAT gateways, and private DNS links are their own composable Infra Components referencing this network's outputs. Keeping those as separate nodes means each has its own lifecycle — a hub-and-spoke topology adds subnets and DNS links without ever touching the network resource itself.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Virtual Network** -- the regional network with your address space (self-planned CIDR blocks or delegated Azure Network Manager IPAM allocation)
- **DNS Configuration** -- custom DNS servers applied network-wide when `dnsServers` entries are provided; otherwise Azure's default resolver serves the network
- **DDoS Protection Attachment** -- created only when `ddosProtectionPlan` is configured; attaches an existing (separately billed) DDoS Protection Plan by ARM ID
- **Network-wide Policies** -- virtual network encryption enforcement and private endpoint VNet policies applied based on configuration

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** -- the network must be created inside a resource group. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **CIDR planning** -- pick private ranges (RFC 1918) that do not overlap with networks you will peer with or with on-premises ranges reachable over VPN/ExpressRoute. Blocks can be added later in place, but a block in use by subnets cannot shrink.
- **IPAM (optional)** -- delegated allocation requires an Azure Network Manager IPAM pool; the network requests a size and the pool provisions a non-overlapping range at deploy time.

## Deploy

### Console

Open the deployment store, find **Azure Virtual Network**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Virtual Network** preset in the [Presets](#presets) tab to pre-populate a general-purpose /16 network with Azure's default DNS.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualNetwork
metadata:
  name: prod-hub
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "network-rg"
  name: prod-hub-vnet
  addressSpaces:
    - "10.0.0.0/16"
  tags:
    cost-center: net-1234
```

```shell
planton apply -f azure-virtual-network.yaml
```

This creates a /16 network with Azure's default resolver -- room for dozens of /24 subnets. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the network to a resource group deployed in the same InfraPipeline:

```yaml
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  name: production-vnet
  addressSpaces:
    - "10.0.0.0/16"
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the network with the resolved values.

## Key Configuration

These are the most important decisions when configuring a virtual network. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Address source** -- exactly one of `addressSpaces` (self-planned CIDR blocks, the common choice) or `ipAddressPools` (delegated allocation from Azure Network Manager IPAM, at most two entries -- one per IP version). A /16 starting block means you will likely never renumber; dual-stack networks carry an IPv4 and an IPv6 block side by side.

**DNS servers** -- leave `dnsServers` empty for Azure's default resolver (168.63.129.16), which Azure Private DNS zone resolution requires. Custom servers replace resolution for EVERY workload in the network and must forward to 168.63.129.16 for private zones to keep resolving.

**DDoS protection** -- `ddosProtectionPlan` attaches an existing plan by ARM ID for networks fronting public IPs. Attachment and activation are distinct: `enable: false` keeps the plan attached with protection off, ready to re-activate without re-attaching. Omit the block entirely for Azure's free basic protection.

**Encryption enforcement** -- `encryption` enables VM-to-VM traffic encryption over the Azure backbone. `ALLOW_UNENCRYPTED` is the only enforcement mode ARM currently accepts (`DROP_UNENCRYPTED` is API-defined but not yet generally available).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `virtual_network_id` | Azure Resource Manager ID of the network | AzureSubnet (partitioning), AzurePrivateDnsZoneVirtualNetworkLink (private DNS resolution), AzureVirtualNetworkPeering (connecting networks) |
| `virtual_network_name` | Name of the virtual network | Resources joining by name inside the network |
| `guid` | The stable GUID ARM assigns at creation | BGP community advertisement, network diagnostics |
| `address_spaces` | The address space ACTUALLY carried | For IPAM-allocated networks, the only place the provisioned ranges are visible -- NSG rules, firewall rules, downstream planning |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard network** -- A general-purpose /16 with Azure's default resolver: the networking foundation for any new environment or a spoke in a hub-and-spoke topology. Start from the **Standard Virtual Network** preset.

**Hub with custom DNS** -- Two address blocks (shared services grow fast in hubs), custom DNS servers for on-premises integration, and a raised flow timeout for long-lived gateway traffic. Start from the **Hub Network with Custom DNS** preset.

**DDoS-protected edge network** -- An attached DDoS Protection Plan and virtual network encryption for networks hosting internet-facing workloads. Start from the **DDoS-Protected Edge Network** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the network is created in
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- partitions the network's address space into workload segments
- [**Azure NAT Gateway**](/infra-catalog/azure-nat-gateway) -- provides managed outbound connectivity for the network's subnets
- [**Azure Private DNS Zone Virtual Network Link**](/infra-catalog/azure-private-dns-zone-virtual-network-link) -- makes private DNS zones resolvable from this network
- [**Azure Virtual Network Peering**](/infra-catalog/azure-virtual-network-peering) -- connects this network to other networks
