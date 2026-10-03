# Azure Virtual Network Peering

Deploys one direction of an Azure Virtual Network Peering — private, low-latency connectivity between two virtual networks over the Microsoft backbone, without gateways, public IPs, or encryption overhead. Peered networks exchange traffic as if they were one network while remaining separately owned — the building block of hub-and-spoke topologies. **One resource models ONE DIRECTION**, exactly as ARM does: connectivity flows only once both directions exist, so a working pair is two of these resources with local and remote swapped, typically stamped from the same chart.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Virtual Network Peering** -- one direction, written as an ARM child of the LOCAL network (the module derives the resource group and network name from the local network's ARM ID — this spec carries no placement fields)
- **Traffic posture** -- the four connectivity dials (network access, forwarded traffic, gateway transit, use remote gateways), each an explicit position or Azure's default
- **Scope** -- the networks' complete address spaces (the standard shape), or subnet-scoped peering restricted to the subnets you list on each side

No tags — peerings are not tracked ARM resources. The reciprocal direction is NOT created here; declare it as its own resource on the remote network.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **Both virtual networks** must exist (reference AzureVirtualNetwork Infra Components, or pass literal ARM IDs — cross-subscription works unchanged).
- **Non-overlapping address spaces** — Azure rejects peerings between networks whose CIDR ranges overlap.
- **For gateway transit**: the hub network needs a VPN/ExpressRoute gateway; the spoke network must have none of its own.

## Deploy

### Console

Open the deployment store, find **Azure Virtual Network Peering**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Hub to Spoke** preset in the [Presets](#presets) tab, and pair it with **Spoke to Hub** for the reciprocal direction.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualNetworkPeering
metadata:
  name: hub-to-spoke1
  org: acme-corp
  env: prod
spec:
  name: hub-to-spoke1
  virtualNetworkId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-network-hub/providers/Microsoft.Network/virtualNetworks/hub-vnet
  remoteVirtualNetworkId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-spoke1/providers/Microsoft.Network/virtualNetworks/spoke1-vnet
  allowForwardedTraffic: true
  allowGatewayTransit: true
```

```shell
planton apply -f peering.yaml
```

This declares the HUB side. Apply the reciprocal manifest (local and remote swapped, `useRemoteGateways: true`) on the spoke to complete the pair. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire both directions from one chart:

```yaml
spec:
  virtualNetworkId:
    valueFrom:
      kind: AzureVirtualNetwork
      name: hub-vnet
      fieldPath: status.outputs.virtual_network_id
  remoteVirtualNetworkId:
    valueFrom:
      kind: AzureVirtualNetwork
      name: spoke1-vnet
      fieldPath: status.outputs.virtual_network_id
```

The InfraPipeline resolves the dependency graph, deploys both networks first, then both peering directions — the whole edge lands in one deploy.

## Key Configuration

These are the most important decisions when configuring a peering. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The two networks** -- `virtualNetworkId` is the LOCAL side this peering is written on (its ARM ID carries the resource group); `remoteVirtualNetworkId` is the far side. Both fixed at creation — pointing elsewhere is a new edge.

**Traffic dials** -- four tri-states with Azure defaults: `allowVirtualNetworkAccess` (default true — the reason peering exists), `allowForwardedTraffic` (default false — set true on hub-to-spoke directions so appliance-relayed traffic is admitted), `allowGatewayTransit` (default false — the HUB side offers its gateway), and `useRemoteGateways` (default false — the SPOKE side rides the hub's gateway; requires transit on the hub's peering, no local gateway, and no global peering). Transit and use-remote-gateways cannot both be true on one direction.

**Scope** -- `peerCompleteVirtualNetworksEnabled` defaults to true (whole networks). Set it explicitly false for subnet-scoped peering and list the participating subnets in `localSubnetNames`/`remoteSubnetNames` — shared-services patterns without exposing whole networks. `onlyIpv6PeeringEnabled` restricts a scoped dual-stack peering to the IPv6 space.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureVirtualNetwork** (local) | `virtualNetworkId` | `status.outputs.virtual_network_id` |
| **AzureVirtualNetwork** (remote) | `remoteVirtualNetworkId` | `status.outputs.virtual_network_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `virtual_network_name` | The LOCAL network's name, derived from its ARM ID | Chart composition without re-parsing IDs |
| `resource_group_name` | The local network's resource group, derived from its ARM ID | Chart composition without re-parsing IDs |

The peering also surfaces `peering_id` and `peering_name` -- identity echoes of the edge itself; no downstream Infra Component consumes them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Hub to spoke** -- the hub's direction: forwarded traffic admitted, gateway transit offered. Start from the **Hub to Spoke** preset.

**Spoke to hub** -- the reciprocal direction: rides the hub's gateway (`useRemoteGateways: true`). Start from the **Spoke to Hub** preset.

**Subnet-scoped** -- only the listed subnets participate — shared-services exposure without opening whole networks. Start from the **Subnet-Scoped Peering** preset.

## Works With

- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) -- both endpoints of the edge, referenced by their `virtual_network_id` outputs
- [**Azure Firewall**](/infra-catalog/azure-firewall) -- the hub appliance that relays spoke-to-spoke traffic (pair with `allowForwardedTraffic` on hub-to-spoke directions)
- [**Azure Route Table**](/infra-catalog/azure-route-table) -- steers spoke traffic through the hub appliance the peering carries it to
