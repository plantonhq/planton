# Azure Virtual Hub Connection

Deploys a Virtual Hub Connection -- the attachment that joins one spoke virtual network to a Virtual WAN hub. The connection itself is free; its routing block is where WAN topologies are actually built: route-table association, label-based propagation, static routes toward appliances, and internet security.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Virtual Hub Connection** -- the ARM attachment between the hub and the spoke VNet, with its routing configuration

The connection carries no tags of its own -- ARM addresses it as a child of the hub, and the provider exposes no tags surface on it.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Virtual Hub** the network attaches to.
- **An Azure Virtual Network** (the spoke) whose address space overlaps neither the hub's nor any other connected network's.

## Deploy

### Console

Open the deployment store, find **Azure Virtual Hub Connection**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Any-to-Any Attachment** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualHubConnection
metadata:
  name: spoke-app
  org: acme-corp
  env: prod
spec:
  name: spoke-app
  virtualHubId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/network-rg/providers/Microsoft.Network/virtualHubs/hub-eastus
  remoteVirtualNetworkId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Network/virtualNetworks/app-vnet
```

```shell
planton apply -f azure-virtual-hub-connection.yaml
```

This attaches the spoke VNet to the hub with ARM's default routing -- any-to-any reachability through the hub's built-in route table. The connection provisions in a few minutes and is free; transit through the hub is what bills. An Infra Job tracks the provisioning in real time.

### InfraChart

In a hub-and-spoke chart, one connection per spoke: WAN → hub → **connection** (per VNet), each wiring to the previous by reference:

```yaml
spec:
  name: spoke-app
  virtualHubId:
    valueFrom:
      kind: AzureVirtualHub
      name: hub-eastus
      fieldPath: status.outputs.virtual_hub_id
  remoteVirtualNetworkId:
    valueFrom:
      kind: AzureVirtualNetwork
      name: app-vnet
      fieldPath: status.outputs.virtual_network_id
```

The InfraPipeline resolves the dependency graph, deploys the hub and spoke network first, then provisions the connection with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring a connection. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Routing** -- unset means ARM's default: associate with and propagate to the hub's built-in default route table (any-to-any reachability). Every WAN topology beyond that -- isolation, shared services, service chaining -- is expressed here.

**Association vs. propagation** -- a connection is routed BY exactly one table (`associatedRouteTableId`) and its routes are LEARNED via the tables it propagates to (`propagatedRouteTable`: explicit IDs, labels, or both).

**Internet security** -- `internetSecurityEnabled: true` makes the hub advertise 0.0.0.0/0 to this spoke (typically into a hub firewall via routing intent). Off by default: the spoke keeps its own internet egress.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureVirtualHub** | `virtualHubId` | `status.outputs.virtual_hub_id` |
| **AzureVirtualNetwork** | `remoteVirtualNetworkId` | `status.outputs.virtual_network_id` |
| **AzureVirtualHub** (optional) | `routing.associatedRouteTableId` | `status.outputs.default_route_table_id` or `status.outputs.route_table_ids.<name>` |
| **AzureVirtualHub** (optional) | `routing.inboundRouteMapId` / `outboundRouteMapId` | `status.outputs.route_map_ids.<name>` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `virtual_hub_connection_id` | ARM ID of the connection | A hub BGP peering's `virtualNetworkConnectionId` |

The only other output, `virtual_hub_connection_name`, echoes the connection's name back; no downstream Infra Component consumes it.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Any-to-any attachment** -- the default routing: every connected network reaches every other. Start from the **Any-to-Any Attachment** preset.

**Isolated spoke** -- associate with an isolated table, propagate only to shared targets. Start from the **Isolated Spoke** preset.

## Works With

- [**Azure Virtual Hub**](/infra-catalog/azure-virtual-hub) -- the hub the network attaches to
- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) -- the spoke being attached
