# Azure ExpressRoute Circuit Peering

Deploys an ExpressRoute circuit peering -- the BGP routing configuration that makes routes flow through a circuit. Private peering carries your VNets' address space (what an ExpressRoute-type virtual network gateway connects to); Microsoft peering carries Microsoft 365 and Azure public services. A circuit holds at most one peering of each type, and the peering can be declared before the provider finishes the cross-connect -- ARM stores the configuration, and routes flow once the circuit reads Provisioned.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ExpressRoute Circuit Peering** -- the ARM child of the circuit (named by its type) carrying the VLAN, BGP session addressing, and type-specific configuration
- **Global Reach Connections** -- one per `connections` entry: links from this private peering to other circuits' private peerings

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An ExpressRoute circuit** -- the peering may be configured while the circuit's provider state is still NotProvisioned (ARM stores the configuration), but routes only flow after the connectivity provider completes the cross-connect and the circuit reads Provisioned.
- **The session facts from your provider**: the VLAN id and the /30 address pairs (one per physical link), plus -- for Microsoft peering -- your registered public prefixes.

## Deploy

### Console

Open the deployment store, find **Azure ExpressRoute Circuit Peering**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Private Peering** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureExpressRouteCircuitPeering
metadata:
  name: hq-private-peering
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "network-rg"
  expressRouteCircuitName:
    valueFrom:
      kind: AzureExpressRouteCircuit
      name: hq-circuit
      fieldPath: status.outputs.express_route_circuit_name
  peeringType: AZURE_PRIVATE_PEERING
  vlanId: 100
  primaryPeerAddressPrefix: "192.168.16.0/30"
  secondaryPeerAddressPrefix: "192.168.16.4/30"
```

```shell
planton apply -f azure-express-route-circuit-peering.yaml
```

This configures private peering on the `hq-circuit` circuit: two BGP sessions on VLAN 100, one /30 per physical link, with Microsoft's ASN and edge-port identifiers surfacing in the outputs. An Infra Job tracks the provisioning in real time.

### InfraChart

In a hybrid-connectivity chart the peering sits between the circuit and the gateway: circuit → this peering → EXPRESS_ROUTE-type virtual network gateway → gateway connection. Wire the peering to a circuit deployed in the same InfraPipeline:

```yaml
spec:
  expressRouteCircuitName:
    valueFrom:
      kind: AzureExpressRouteCircuit
      name: hq-circuit
      fieldPath: status.outputs.express_route_circuit_name
```

The InfraPipeline resolves the dependency graph, deploys the circuit first, then configures the peering on it.

## Key Configuration

These are the most important decisions when configuring a peering. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Peering type** -- `AZURE_PRIVATE_PEERING` for VNet connectivity (the common case) or `MICROSOFT_PEERING` for Microsoft public services. The type IS the peering's ARM identity: one of each per circuit, fixed at creation. Public peering exists only for legacy imports -- Azure deprecated it.

**Session addressing** -- the /30 pair (`primaryPeerAddressPrefix` + `secondaryPeerAddressPrefix`, one per physical link) and the provider-assigned `vlanId` (1-4094, unique on the circuit). Your router takes each /30's first usable address, Microsoft's the second.

**Microsoft peering** -- requires `microsoftPeeringConfig` with public prefixes REGISTERED to you (Microsoft validates ownership against internet routing registries before activating) and typically a `routeFilterId` -- without a route filter, Microsoft peering advertises nothing.

**Global Reach** -- `connections` entries link this private peering to other circuits' private peerings (far side by ARM id, a /29 for tunnel addressing, an authorization key when the far circuit is in another subscription).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureExpressRouteCircuit** | `expressRouteCircuitName` | `status.outputs.express_route_circuit_name` |
| **AzureExpressRouteCircuitPeering** | `connections[].peerPeeringId` | `status.outputs.express_route_circuit_peering_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `express_route_circuit_peering_id` | Azure Resource Manager ID of the peering | AzureExpressRouteGateway circuit connections, and another circuit's Global Reach `peerPeeringId` |
| `azure_asn` | Microsoft's BGP ASN (12076 on public Azure) | Router neighbor configuration |
| `primary_azure_port` / `secondary_azure_port` | Microsoft-edge port identifiers | Provider troubleshooting on the physical links |

`connection_ids` is also exported -- the ARM ID of each Global Reach connection created from the `connections` list, keyed by name; it exists for inspection and has no ValueFromRef consumers.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private peering** -- VNet connectivity over the circuit: the standard hybrid path. Start from the **Private Peering** preset.

**Microsoft peering** -- Microsoft 365 / Azure public services with the advertisement contract. Start from the **Microsoft Peering** preset.

## Works With

- [**Azure ExpressRoute Circuit**](/infra-catalog/azure-express-route-circuit) -- the parent circuit this peering configures
- [**Azure Virtual Network Gateway**](/infra-catalog/azure-virtual-network-gateway) -- the EXPRESS_ROUTE-type gateway that consumes private peering
- [**Azure Virtual Network Gateway Connection**](/infra-catalog/azure-virtual-network-gateway-connection) -- the link between a gateway and the circuit
- [**Azure ExpressRoute Gateway**](/infra-catalog/azure-express-route-gateway) -- the Virtual WAN gateway whose circuit connections reference this peering's ID
