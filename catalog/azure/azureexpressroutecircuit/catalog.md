# Azure ExpressRoute Circuit

Deploys an ExpressRoute circuit -- the dedicated PRIVATE connection between your infrastructure and Microsoft, bought through a connectivity provider (Equinix, Megaport, ...) or carved from your own ExpressRoute Direct port. Creating the circuit issues the service key your provider uses to provision the physical cross-connect; peerings (AzureExpressRouteCircuitPeering) then make routes flow through it. Azure meters the circuit from the moment the service key is issued, even while the provider side is unprovisioned.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ExpressRoute Circuit** -- the ARM billing/identity object with its SKU, provisioning mode, and generated service key
- **Circuit Authorizations** -- one per `authorizations` entry: ARM-generated keys that let virtual network gateways in OTHER subscriptions connect to this circuit
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) merged with your `tags`, applied to the circuit

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the circuit will be created.
- **The connectivity facts**: your provider's exact name and peering location plus the bandwidth step you bought -- or your ExpressRoute Direct port's ARM ID. `az network express-route list-service-providers` shows the provider vocabulary.

## Deploy

### Console

Open the deployment store, find **Azure ExpressRoute Circuit**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Provider Circuit** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureExpressRouteCircuit
metadata:
  name: hq-circuit
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "network-rg"
  name: hq-circuit
  skuTier: STANDARD
  skuFamily: METERED_DATA
  serviceProviderName: "Equinix"
  peeringLocation: "Washington DC"
  bandwidthInMbps: 1000
```

```shell
planton apply -f azure-express-route-circuit.yaml
```

This creates a 1 Gbps STANDARD metered circuit provisioned through Equinix at the Washington DC peering location; it sits in `NotProvisioned` until you hand the `service_key` output to your provider. An Infra Job tracks the provisioning in real time. **Billing starts at creation**, even before the provider completes the cross-connect.

### InfraChart

In a hybrid-connectivity chart, the circuit anchors the chain: circuit → private peering → ExpressRoute-type virtual network gateway → gateway connection, each wiring to the previous by reference. Wire the circuit itself to a resource group (or, for a Direct circuit, to its port) deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the circuit with the resolved values.

## Key Configuration

These are the most important decisions when configuring a circuit. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Provisioning mode** -- exactly one (spec-enforced): the service-provider trio (`serviceProviderName` + `peeringLocation` + `bandwidthInMbps`, all fixed at creation except bandwidth, which can only GROW) or the ExpressRoute Direct pair (`expressRoutePortId` + `bandwidthInGbps`). The port is a reference: point it at an `AzureExpressRoutePort` component by name (its id output resolves at deploy time), or pass a literal ARM id for a port managed outside Planton.

**SKU** -- `skuTier` sizes reach (LOCAL: the circuit's metro only, no egress fees; STANDARD: the geopolitical area -- the common choice; PREMIUM: global reach and higher route limits) and `skuFamily` sizes billing (METERED_DATA per outbound GB, UNLIMITED_DATA flat -- economical above roughly two-thirds sustained utilization).

**Authorizations** -- each named entry issues an ARM-generated key (surfaced, sensitive, in `authorization_keys`) that a gateway in another subscription redeems. Deleting an entry revokes it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureExpressRoutePort** | `expressRoutePortId` | `status.outputs.express_route_port_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `express_route_circuit_id` | Azure Resource Manager ID of the circuit | AzureVirtualNetworkGatewayConnection's `expressRouteCircuitId` -- the link that attaches a VNet gateway to the circuit |
| `express_route_circuit_name` | Name of the circuit | AzureExpressRouteCircuitPeering's `expressRouteCircuitName` |
| `service_key` | The provisioning credential (sensitive) | Handed to the connectivity provider |
| `service_provider_provisioning_state` | NotProvisioned / Provisioning / Provisioned / Deprovisioning | Gating peering deployment |
| `authorization_keys` | Name-keyed issued keys (sensitive) | Far-side gateway connections in other subscriptions |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Provider circuit** -- The common shape: a carrier-provisioned circuit at a named peering location. Start from the **Provider Circuit** preset.

**Metro-local circuit** -- LOCAL tier for egress-free connectivity in one metro. Start from the **Local Metro Circuit** preset.

**Direct-port circuit** -- carved from your own ExpressRoute Direct port, with no third-party provider in the loop: `expressRoutePortId` plus `bandwidthInGbps`, for the 10/100 Gbps estates that own their cross-connects. Start from the **Direct Port Circuit** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the circuit is created in
- [**Azure ExpressRoute Circuit Peering**](/infra-catalog/azure-express-route-circuit-peering) -- the BGP configuration that makes routes flow
- [**Azure ExpressRoute Port**](/infra-catalog/azure-express-route-port) -- the Direct port a provider-less circuit is carved from
- [**Azure Virtual Network Gateway**](/infra-catalog/azure-virtual-network-gateway) -- the EXPRESS_ROUTE-type gateway that connects a VNet to the circuit's private peering
