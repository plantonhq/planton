# Azure VPN Gateway Connection

Deploys a VPN Gateway Connection -- the tunnel bundle joining one branch (a VPN Site) to a Virtual WAN hub's VPN gateway. Each `vpnLinks` entry is one tunnel, pinned to one of the site's links and carrying that tunnel's own IPsec, BGP, and NAT choices. The connection is free and provisions in minutes; **a tunnel reaches Connected only when the branch device negotiates** -- deployment success and tunnel establishment are different events.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **VPN Gateway Connection** -- the ARM child of the gateway carrying the tunnels, their routing configuration, and optional traffic selectors

The connection carries no region, resource group, or tags of its own -- ARM derives all three through the gateway, and the provider exposes no tags surface on it.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure VPN Gateway** in a virtual hub -- the connection is its child.
- **An Azure VPN Site** with at least one link -- each tunnel pins to a link by ARM ID.
- **The branch device**, configured with the gateway's public IPs and the agreed pre-shared key -- without it the tunnels provision but never connect.

## Deploy

### Console

Open the deployment store, find **Azure VPN Gateway Connection**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Branch Connection** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVpnGatewayConnection
metadata:
  name: branch-london
  org: acme-corp
  env: prod
spec:
  name: branch-london
  vpnGatewayId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/network-rg/providers/Microsoft.Network/vpnGateways/hub-vpn-gateway
  remoteVpnSiteId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/network-rg/providers/Microsoft.Network/vpnSites/branch-london
  vpnLinks:
    - name: primary-isp
      vpnSiteLinkId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/network-rg/providers/Microsoft.Network/vpnSites/branch-london/vpnSiteLinks/primary-isp
```

```shell
planton apply -f azure-vpn-gateway-connection.yaml
```

This creates a one-tunnel connection joining the branch site's primary link to the hub gateway with Azure's default IPsec proposals and a generated shared key. The connection provisions in minutes and is free; watch the tunnel's connection state separately. An Infra Job tracks the provisioning in real time.

### InfraChart

In a branch-connectivity chart the order is: WAN → hub → VPN gateway, plus one site per branch → one **connection** per site, each wiring to the previous by reference:

```yaml
spec:
  name: branch-london
  vpnGatewayId:
    valueFrom:
      kind: AzureVpnGateway
      name: hub-vpn-gateway
      fieldPath: status.outputs.vpn_gateway_id
  remoteVpnSiteId:
    valueFrom:
      kind: AzureVpnSite
      name: branch-london
      fieldPath: status.outputs.vpn_site_id
  vpnLinks:
    - name: primary-isp
      vpnSiteLinkId:
        valueFrom:
          kind: AzureVpnSite
          name: branch-london
          fieldPath: status.outputs.link_ids.primary-isp
```

The InfraPipeline resolves the dependency graph, deploys the gateway and site first, then provisions the connection with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring a connection. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**vpnLinks** -- one entry per site link being connected, each referencing `status.outputs.link_ids.<link-name>` on the site. Two entries (a dual-ISP site) is Virtual WAN's active-active branch pattern.

**Shared key** -- reference a secret per link, or omit it and let Azure generate one. Both tunnel ends must agree; a mismatched key is the classic provisioned-but-never-Connected cause.

**Routing** -- unset means ARM's default (associate with and propagate to the hub's default table). A configured block must name its `associatedRouteTableId` -- the provider's own contract.

**BGP per link** -- `bgpEnabled` is fixed at tunnel creation and requires the site link to carry its `bgp` block. Static-vs-BGP is a per-branch decision made on BOTH objects.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureVpnGateway** | `vpnGatewayId` | `status.outputs.vpn_gateway_id` |
| **AzureVpnSite** | `remoteVpnSiteId` | `status.outputs.vpn_site_id` |
| **AzureVpnSite** | `vpnLinks[].vpnSiteLinkId` | `status.outputs.link_ids.<link-name>` |
| **AzureVpnGateway** (optional) | `vpnLinks[].egressNatRuleIds` / `ingressNatRuleIds` | `status.outputs.nat_rule_ids.<rule-name>` |
| **AzureVirtualHub** (optional) | `routing.associatedRouteTableId` | `status.outputs.default_route_table_id` or `status.outputs.route_table_ids.<name>` |
| **AzureVirtualHub** (optional) | `routing.inboundRouteMapId` / `outboundRouteMapId` | `status.outputs.route_map_ids.<name>` |

### What This Kind Provides

The connection's `status.outputs` carry only `connection_id` and `connection_name` -- identity echoes of the resource itself. No downstream Infra Component consumes them: the connection is a leaf of the branch-connectivity graph, and the tunnels' live state is runtime telemetry, not a provisioning output.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard branch connection** -- one tunnel, Azure's default IPsec proposals, generated key. Start from the **Standard Branch Connection** preset.

**Pinned IPsec** -- compliance-grade cipher pinning with a referenced pre-shared key. Start from the **Pinned IPsec Connection** preset.

## Works With

- [**Azure VPN Gateway**](/infra-catalog/azure-vpn-gateway) -- the gateway this connection belongs to
- [**Azure VPN Site**](/infra-catalog/azure-vpn-site) -- the branch being connected
- [**Azure Virtual Hub**](/infra-catalog/azure-virtual-hub) -- whose route tables and route maps the routing block references
