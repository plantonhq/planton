# Azure DNS Private Resolver

Deploys Azure DNS Private Resolver -- the managed DNS proxy that resolves names across the hybrid boundary without anyone running DNS server VMs: on-premises queries INTO Azure through inbound endpoints, Azure queries OUT to on-premises DNS through outbound endpoints. The resolver anchors to one virtual network (Azure allows at most one per network), each endpoint occupies its own delegated subnet, and everything except tags is fixed at creation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **DNS Private Resolver** -- the managed proxy, anchored to one virtual network
- **Inbound endpoints** (optional, up to 5) -- private IPs that answer DNS queries sent to them with the network's private DNS view; one per dedicated delegated subnet
- **Outbound endpoints** (optional, up to 5) -- egress points for queries leaving Azure, steered by the forwarding rulesets that bind them; one per dedicated delegated subnet

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An Azure Virtual Network** the resolver anchors to -- Azure allows at most ONE resolver per network.
- **One dedicated Azure Subnet per endpoint**, delegated to `Microsoft.Network/dnsResolvers`, sized /28 to /24, hosting nothing else.

### Azure Subscription

- **Endpoints bill hourly** from the moment they provision; the resolver object itself is free.
- **Everything except tags is create-only** -- plan the endpoint layout up front; edits replace the affected endpoint.
- **The region must match the anchor network's region** -- ARM rejects a mismatch at deploy time, and endpoints deploy into the same region.

## Deploy

### Console

Open the deployment store, find **Azure DNS Private Resolver**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the anchor network reference, and the inbound and outbound endpoint lists with their delegated subnets. Start from the **Hybrid Resolver** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateDnsResolver
metadata:
  name: hub-dns-resolver
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: prod-network
  name: hub-dns-resolver
  virtualNetworkId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network/providers/Microsoft.Network/virtualNetworks/hub-vnet
  inboundEndpoints:
    - name: inbound
      subnetId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network/providers/Microsoft.Network/virtualNetworks/hub-vnet/subnets/dns-inbound
  outboundEndpoints:
    - name: outbound
      subnetId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network/providers/Microsoft.Network/virtualNetworks/hub-vnet/subnets/dns-outbound
```

```shell
planton apply -f resolver.yaml
```

This creates the resolver anchored to `hub-vnet` with one inbound endpoint (a dynamically assigned private IP that answers with Azure's private DNS view) and one outbound endpoint (the egress point forwarding rulesets bind). An Infra Job tracks the provisioning in real time.

### InfraChart

When the network stack is Infra Components in the same chart, wire everything by reference:

```yaml
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      name: prod-network
  name: hub-dns-resolver
  virtualNetworkId:
    valueFrom:
      name: hub-vnet
  inboundEndpoints:
    - name: inbound
      subnetId:
        valueFrom:
          name: dns-inbound-subnet
  outboundEndpoints:
    - name: outbound
      subnetId:
        valueFrom:
          name: dns-outbound-subnet
```

The InfraPipeline resolves the dependency graph, provisioning the resource group, network, and delegated subnets before the resolver that occupies them.

## Key Configuration

These are the most important decisions when configuring an Azure DNS Private Resolver. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One resolver per network is a hard wall -- put it in the hub.** In hub-and-spoke, the hub owns THE resolver and every spoke consumes it through ruleset links; spokes need no resolver, no endpoints, not even peering for DNS to flow. Placing a resolver in a spoke buys nothing and burns the spoke's one slot.

**Carve the delegated subnets when you design the network.** Every endpoint needs its own subnet -- delegated to `Microsoft.Network/dnsResolvers`, /28 to /24, carrying nothing else, one endpoint per subnet. ARM enforces all of it at deploy time, and the delegation lives on the subnet resource, so a missing delegation surfaces as a deploy-time failure, not a manifest error. A /28 is all an endpoint ever needs; carve both subnets up front even if you deploy only one endpoint today.

**Everything except tags is create-only.** The resolver and its endpoints have no ARM update surface beyond tags: editing an endpoint replaces it, and changing the region, resource group, name, or network replaces the resolver. Adding or removing endpoints is additive -- endpoints are keyed by name, so extending the list never touches existing siblings.

**Pin the inbound IP if on-premises config is expensive to change.** DYNAMIC allocation (the default) picks a free address at create; that address survives the endpoint's lifetime but not its replacement. If the inbound IP is fanned out across datacenter forwarder configs, use STATIC allocation with a deliberately chosen address from the delegated subnet's range -- STATIC requires `privateIpAddress`, and DYNAMIC forbids it.

**Endpoints are the billing meter and the capacity dial.** Each endpoint bills hourly and serves roughly 10,000 queries/second. One endpoint each way is the right day-one shape -- add endpoints (up to 5 each way) for throughput, not redundancy: a single endpoint is already zone-resilient where the region has zones.

**Declaration order picks the primary.** The FIRST inbound endpoint's IP surfaces as the singular `inbound_endpoint_ip` output, and the FIRST outbound endpoint's ARM ID as `outbound_endpoint_id` -- the values downstream components consume by default. Keep the primary first when adding endpoints.

**Wire deploy order through the outputs.** Forwarding rulesets bind the outbound endpoint by ARM ID -- reference `outbound_endpoint_id` rather than composing the ID by hand, and the reference IS the ordering edge. The same holds for on-premises pointing: read `inbound_endpoint_ip` from the outputs, never guess the subnet's .4 address.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Azure Resource Group | `resourceGroup` | `status.outputs.resource_group_name` |
| Azure Virtual Network | `virtualNetworkId` | `status.outputs.virtual_network_id` |
| Azure Subnet (one per endpoint) | `inboundEndpoints[].subnetId`, `outboundEndpoints[].subnetId` | `status.outputs.subnet_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `inbound_endpoint_ip` | Private IP of the first inbound endpoint | On-premises conditional forwarders and VNet custom DNS settings point at it |
| `outbound_endpoint_id` | ARM ID of the first outbound endpoint | What an Azure DNS Forwarding Ruleset binds by default |

Multi-endpoint deployments also get `inbound_endpoint_ips` and `outbound_endpoint_ids`, keyed by endpoint name; `dns_resolver_id` and `dns_resolver_name` identify the resolver itself.

## Common Patterns

**Hybrid hub resolver** -- One inbound and one outbound endpoint on the hub network: on-premises resolves Azure names through the inbound IP, Azure resolves on-premises names through the outbound endpoint and a forwarding ruleset. Replaces a pair of IaaS DNS forwarder VMs with one managed service. Start from the **Hybrid Resolver** preset.

**Inbound only, pinned IP** -- The datacenter needs Azure's private names but Azure never queries on-premises domains: one STATIC-allocated inbound endpoint whose address survives replacement, so fanned-out forwarder configs never change. Start from the **Inbound Only (Pinned IP)** preset.

**Spokes consume, hub owns** -- Spoke networks get outbound resolution by linking the hub's forwarding ruleset to themselves (Azure DNS Resolver Virtual Network Link) -- no per-spoke resolvers, no per-spoke endpoints, no extra hourly billing.

## Works With

- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) -- the anchor network; one resolver per network, so this is usually the hub.
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- each endpoint's dedicated delegated subnet; reference its `subnet_id` output.
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- where the resolver lives; reference its `resource_group_name` output.
- [**Azure DNS Forwarding Ruleset**](/infra-catalog/azure-private-dns-resolver-forwarding-ruleset) -- binds the outbound endpoint and decides which domains forward where.
- [**Azure DNS Resolver Virtual Network Link**](/infra-catalog/azure-private-dns-resolver-virtual-network-link) -- attaches a ruleset's rules to consuming networks, including spokes.
