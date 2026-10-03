# Azure Local Network Gateway

Deploys a local network gateway -- Azure's description of the ON-PREMISES side of a site-to-site VPN: the VPN device's public endpoint (IP or FQDN) and the address space reachable behind it. It deploys nothing on-premises and costs nothing to keep; it is the address-book entry an AzureVirtualNetworkGatewayConnection points at, one per site.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Local Network Gateway** -- the ARM object carrying the site's endpoint, reachable prefixes, and (optionally) its BGP speaker settings
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) merged with your `tags`, applied to the object

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the object will be created.
- **The site's facts**: the device's public IPv4 address (or FQDN) and the CIDR ranges behind it -- or its BGP speaker's ASN and tunnel-interior peering address.

## Deploy

### Console

Open the deployment store, find **Azure Local Network Gateway**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Static Site** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureLocalNetworkGateway
metadata:
  name: hq-datacenter
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "network-rg"
  name: hq-datacenter
  gatewayAddress: "198.51.100.4"
  addressSpaces:
    - "192.168.0.0/16"
```

```shell
planton apply -f azure-local-network-gateway.yaml
```

This creates the address-book entry for the HQ site -- the static endpoint at 198.51.100.4 and the 192.168.0.0/16 space behind it. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying a complete site-to-site story as one chart, the connection wires to this description by reference:

```yaml
# On the AzureVirtualNetworkGatewayConnection:
spec:
  localNetworkGatewayId:
    valueFrom:
      kind: AzureLocalNetworkGateway
      name: hq-datacenter
      fieldPath: status.outputs.local_network_gateway_id
```

The InfraPipeline deploys this description (seconds) alongside the gateway (25-45 minutes), then the connection that ties them together.

## Key Configuration

These are the most important decisions when configuring a site description. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Endpoint** -- exactly one of `gatewayAddress` (static public IPv4) or `gatewayFqdn` (re-resolved name, for sites whose address changes). Spec-enforced.

**Routing source** -- `addressSpaces` (static prefixes Azure routes into the tunnel), `bgpSettings` (the site advertises its own routes), or both. An object with neither routes nothing, and ARM rejects it -- spec-enforced upfront.

**BGP peering address** -- lives INSIDE the tunnel (the device's tunnel interface), never the device's public address. The ASN must differ from the Azure gateway's (65515-65520 are Azure-reserved).

**Naming** -- name descriptions after the site ("hq-datacenter", "branch-london"); a connection references exactly one site.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `local_network_gateway_id` | Azure Resource Manager ID of the description | AzureVirtualNetworkGatewayConnection's `localNetworkGatewayId`; AzureVirtualNetworkGateway's `defaultLocalNetworkGatewayId` (forced tunneling) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Static site** -- A static public IP and the prefixes behind it: the classic datacenter description. Start from the **Static Site** preset.

**BGP site** -- An FQDN-addressed site whose routes arrive over BGP. Start from the **BGP Site** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the description is created in
- [**Azure Virtual Network Gateway**](/infra-catalog/azure-virtual-network-gateway) -- the Azure-side appliance tunnels terminate on
- [**Azure Virtual Network Gateway Connection**](/infra-catalog/azure-virtual-network-gateway-connection) -- the tunnel that consumes this description
