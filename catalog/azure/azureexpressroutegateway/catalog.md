# Azure ExpressRoute Gateway

Deploys an ExpressRoute Gateway -- the Virtual WAN on-ramp for ExpressRoute circuits -- together with its connections, each joining one circuit peering to the hub. The gateway bills per scale unit from creation and takes roughly 30 minutes to provision; a hub holds at most one, and each connection requires the circuit's provider side to already be provisioned.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ExpressRoute Gateway** -- the gateway in the hub, with its scale-unit floor and non-WAN-traffic policy
- **ExpressRoute Connections** (optional) -- one per `connections` entry: the join between a circuit's private peering and the hub, with authorization key, routing block, and weight
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) merged with your `tags`, applied to the gateway

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the gateway will be created.
- **An Azure Virtual Hub** the gateway deploys into (must be in the same region).
- **For connections**: an ExpressRoute circuit whose provider side is PROVISIONED -- ARM rejects a connection to an unprovisioned circuit.

## Deploy

### Console

Open the deployment store, find **Azure ExpressRoute Gateway**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Gateway** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureExpressRouteGateway
metadata:
  name: hub-er-gateway
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "network-rg"
  name: hub-er-gateway
  virtualHubId:
    valueFrom:
      kind: AzureVirtualHub
      name: hub-eastus
      fieldPath: status.outputs.virtual_hub_id
  scaleUnits: 1
```

```shell
planton apply -f azure-express-route-gateway.yaml
```

This creates a one-scale-unit ExpressRoute gateway in the `hub-eastus` Virtual WAN hub, ready for circuit connections. An Infra Job tracks the provisioning in real time. The gateway bills hourly per scale unit from creation, and ARM takes roughly 30 minutes to provision one.

### InfraChart

In a hybrid-connectivity chart, the gateway follows the hub: WAN → hub → **ExpressRoute gateway** → connections referencing circuit peerings. Wire the gateway to its hub, and each connection to its circuit's private peering, deployed in the same InfraPipeline:

```yaml
spec:
  virtualHubId:
    valueFrom:
      kind: AzureVirtualHub
      name: hub-eastus
      fieldPath: status.outputs.virtual_hub_id
  connections:
    - name: dc-primary
      expressRouteCircuitPeeringId:
        valueFrom:
          kind: AzureExpressRouteCircuitPeering
          name: hq-private-peering
          fieldPath: status.outputs.express_route_circuit_peering_id
```

The InfraPipeline resolves the dependency graph, deploys the hub and the peering first, then provisions the gateway and its connections.

## Key Configuration

These are the most important decisions when configuring a gateway. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scale units** -- the MINIMUM (1-10); each unit carries ~2 Gbps of circuit-to-WAN throughput and bills hourly. ARM auto-scales above the floor under load. Updatable in place.

**Connections** -- each joins one circuit PRIVATE PEERING to the hub. A cross-subscription circuit needs the `authorizationKey` its owner generated; a same-subscription circuit needs none.

**Non-Virtual-WAN traffic** -- off by default; on lets classic VNets connected to the same circuit exchange traffic through this gateway.

**Connection routing** -- leave `routing` unset for ARM's default: associate with and propagate to the hub's built-in default route table. A configured block must name an associated route table or propagation targets; route maps shape what arrives from and is advertised to the circuit, and `routingWeight` (0-32000) breaks ties when the same prefix is reachable over multiple connections.

**One-way doors** -- `name`, `region`, `resourceGroup`, and `virtualHubId` are fixed at creation: changing any of them replaces the gateway, and a replacement is another ~30-minute build. `scaleUnits`, the traffic policy, and the connections list edit in place.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureVirtualHub** | `virtualHubId` | `status.outputs.virtual_hub_id` |
| **AzureExpressRouteCircuitPeering** (per connection) | `connections[].expressRouteCircuitPeeringId` | `status.outputs.express_route_circuit_peering_id` |
| **AzureVirtualHub** (optional) | `connections[].routing.associatedRouteTableId` | `status.outputs.default_route_table_id` or `status.outputs.route_table_ids.<name>` |

### What This Kind Provides

After provisioning, `status.outputs` surfaces the gateway's ARM ID and name (`express_route_gateway_id`, `express_route_gateway_name`) and a name-keyed map of connection ARM IDs (`connection_ids`). No catalog kind consumes these via ValueFromRef -- the gateway is the end of the hybrid-connectivity chain, and its connections are declared inline rather than as separate resources -- so the outputs exist for inspection and external tooling.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard gateway** -- one scale unit in the hub, connections added when circuits are ready. Start from the **Standard Gateway** preset.

**Circuit connection** -- the gateway joined to a provisioned circuit's private peering. Start from the **Circuit Connection** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the gateway is created in
- [**Azure Virtual Hub**](/infra-catalog/azure-virtual-hub) -- the hub the gateway deploys into
- [**Azure ExpressRoute Circuit Peering**](/infra-catalog/azure-express-route-circuit-peering) -- the private peering each connection joins
