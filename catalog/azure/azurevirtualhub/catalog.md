# Azure Virtual Hub

Deploys a Virtual Hub -- the managed regional router of an Azure Virtual WAN -- together with its routing customization: custom route tables, route maps, BGP peerings with network virtual appliances, and routing intent. Spoke networks attach through Virtual Hub Connections; VPN and ExpressRoute gateways deploy into the hub as separate resources that reference it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Virtual Hub** -- the regional router with its address space, tier, routing preference, and router capacity
- **Hub Route Tables** (optional) -- one per `routeTables` entry, with inline static routes toward next-hop resources
- **Route Maps** (optional) -- one per `routeMaps` entry: ordered match/transform rules applied to BGP routes on hub connections
- **BGP Connections** (optional) -- one per `bgpConnections` entry: peerings between the hub router and NVAs in connected spokes
- **Routing Intent** (optional, at most one) -- Internet/private traffic steered through a security appliance in the hub
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) merged with your `tags`, applied to the hub

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the hub will be created.
- **An Azure Virtual WAN** the hub belongs to (this kind models WAN hubs; `virtualWanId` is required).

## Deploy

### Console

Open the deployment store, find **Azure Virtual Hub**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Hub** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVirtualHub
metadata:
  name: hub-eastus
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "network-rg"
  name: hub-eastus
  virtualWanId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/network-rg/providers/Microsoft.Network/virtualWans/global-wan
  addressPrefix: "10.100.0.0/23"
```

```shell
planton apply -f azure-virtual-hub.yaml
```

This creates a Standard-tier hub with a /23 address space in the referenced WAN and no routing customization -- it routes any-to-any through its built-in default table. A Standard hub bills hourly from creation, and ARM takes 15-30 minutes to bring the hub's router to a Provisioned state. An Infra Job tracks the provisioning in real time.

### InfraChart

In a hub-and-spoke chart, the hub is the regional pivot: WAN → **hub** → connections, gateways, and firewall, each wiring to the previous by reference:

```yaml
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  name: hub-eastus
  virtualWanId:
    valueFrom:
      kind: AzureVirtualWan
      name: global-wan
      fieldPath: status.outputs.virtual_wan_id
  addressPrefix: "10.100.0.0/23"
```

The InfraPipeline resolves the dependency graph, deploys the resource group and WAN first, then provisions the hub with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring a hub. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Address prefix** -- the hub's private CIDR (minimum /24, Microsoft recommends /23). The router, gateways, and firewall draw addresses from it, so it must not overlap any connected VNet or branch. Fixed at creation.

**Routing customization** -- custom `routeTables` (associate/propagate targets for connections), `routeMaps` (BGP transformations), and `routingIntent` (firewall steering). A hub with none of these routes any-to-any through its built-in default table.

**Routing intent vs. custom tables** -- routing intent takes over the hub's routing policy; per-connection route-table customization and routing intent are mutually exclusive on ARM's side.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureVirtualWan** | `virtualWanId` | `status.outputs.virtual_wan_id` |
| **AzureFirewall** (optional) | `routingIntent.routingPolicies[].nextHop`, route table route `nextHop` | `status.outputs.firewall_id` |
| **AzureVirtualHubConnection** (optional) | `bgpConnections[].virtualNetworkConnectionId` | `status.outputs.virtual_hub_connection_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `virtual_hub_id` | ARM ID of the hub | A connection's or gateway's `virtualHubId` |
| `default_route_table_id` | ARM ID of the built-in default route table | A connection's `associatedRouteTableId` |
| `virtual_router_asn` | The hub router's BGP ASN (always 65515) | NVA BGP configuration |
| `virtual_router_ips` | The hub router's peering IPv4 addresses | NVA BGP neighbor configuration |
| `route_table_ids` | Custom route table ARM IDs, keyed by name | `status.outputs.route_table_ids.<name>` in connection routing |
| `route_map_ids` | Route map ARM IDs, keyed by name | `status.outputs.route_map_ids.<name>` as inbound/outbound maps |

The hub's remaining outputs (`virtual_hub_name`, `bgp_connection_ids`, `routing_intent_id`) echo names and child IDs back for reference; no downstream Infra Component consumes them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard hub** -- the WAN-attached regional router with ARM's defaults. Start from the **Standard Hub** preset.

**Isolated spokes** -- custom route tables with labels, ready for connections that associate/propagate for isolation. Start from the **Isolated Spokes Hub** preset.

**Secured hub** -- routing intent steering Internet and private traffic through a hub firewall. Start from the **Secured Hub (Routing Intent)** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the hub is created in
- [**Azure Virtual WAN**](/infra-catalog/azure-virtual-wan) -- the WAN the hub belongs to
- [**Azure Virtual Hub Connection**](/infra-catalog/azure-virtual-hub-connection) -- attaches spoke VNets to this hub
- [**Azure ExpressRoute Gateway**](/infra-catalog/azure-express-route-gateway) -- brings ExpressRoute circuits into this hub
- [**Azure Firewall**](/infra-catalog/azure-firewall) -- the security appliance routing intent steers through
