# Azure VPN Gateway

Deploys a Virtual WAN VPN Gateway -- the managed site-to-site VPN terminator inside a virtual hub (ARM allows one per hub). Branch sites connect to it through VPN Gateway Connections; capacity is bought in scale units (500 Mbps each) across an active-active instance pair Azure manages. **The gateway bills hourly per scale unit from creation and takes 30-45 minutes to create** -- plan lifecycle around both.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **VPN Gateway** -- the managed terminator inside the hub, with its BGP speaker and instance public IPs (Azure assigns them; there is no public-IP resource to bring)
- **NAT Rules** (optional) -- one ARM child per spec entry, published by name in the `nat_rule_ids` output

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Virtual Hub** to deploy into -- the gateway's region must match the hub's, and the hub must not already hold a VPN gateway (ARM's one-per-hub rule).

## Deploy

### Console

Open the deployment store, find **Azure VPN Gateway**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Hub Gateway** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVpnGateway
metadata:
  name: hub-vpn-gateway
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: network-rg
  name: hub-vpn-gateway
  virtualHubId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/network-rg/providers/Microsoft.Network/virtualHubs/hub-eastus
```

```shell
planton apply -f azure-vpn-gateway.yaml
```

This creates a one-scale-unit (500 Mbps) gateway inside the referenced hub, with Azure-assigned instance public IPs and the default BGP speaker. Expect the create to run 30-45 minutes (ARM's slow path, not a failure); the gateway bills from the moment it exists. An Infra Job tracks the provisioning in real time.

### InfraChart

In a branch-connectivity chart the order is: WAN → hub → **VPN gateway**, plus one site per branch → one connection per site, each wiring to the previous by reference:

```yaml
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  name: hub-vpn-gateway
  virtualHubId:
    valueFrom:
      kind: AzureVirtualHub
      name: hub-eastus
      fieldPath: status.outputs.virtual_hub_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group, WAN, and hub first, then provisions the gateway with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring a gateway. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scale unit** -- aggregate capacity in 500 Mbps units across the instance pair; it also scales the bill. Updates in place, so start at 1 and grow with measured demand.

**Routing preference** -- how tunnel traffic reaches branch endpoints: Microsoft's backbone (default) or the public internet near the gateway. Fixed at creation.

**NAT rules** -- translation that tunnels OPT INTO via a connection link's `egressNatRuleIds`/`ingressNatRuleIds`; needed when branch address spaces overlap. Each rule's ID surfaces in `nat_rule_ids` by name.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureVirtualHub** | `virtualHubId` | `status.outputs.virtual_hub_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `vpn_gateway_id` | ARM ID of the gateway | A connection's `vpnGatewayId` |
| `bgp_asn` | The gateway's BGP ASN (65515 today) | Branch device remote-ASN configuration |
| `public_ip_addresses` | Each instance's public IPv4 | What branch devices dial |
| `private_ip_addresses` | Each instance's private IPv4 | Private-peering tunnel endpoints |
| `nat_rule_ids` | Each NAT rule's ARM ID, keyed by rule name | A connection link's NAT opt-ins (`status.outputs.nat_rule_ids.<rule-name>`) |

The only other output, `vpn_gateway_name`, echoes the gateway's name back; no downstream Infra Component consumes it.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard hub gateway** -- one scale unit, defaults everywhere. Start from the **Standard Hub Gateway** preset.

**Overlapping branches** -- NAT rules plus BGP route translation. Start from the **NAT for Overlapping Branches** preset.

## Works With

- [**Azure Virtual Hub**](/infra-catalog/azure-virtual-hub) -- the hub the gateway deploys into
- [**Azure VPN Site**](/infra-catalog/azure-vpn-site) -- the branches that connect
- [**Azure VPN Gateway Connection**](/infra-catalog/azure-vpn-gateway-connection) -- the tunnels joining sites to this gateway
