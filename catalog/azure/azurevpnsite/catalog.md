# Azure VPN Site

Deploys a VPN Site -- the Virtual WAN address-book entry for one branch location: its internet links (each with a public endpoint and optional BGP speaker), the address space reachable behind it, and the device that terminates the tunnels. The site is free and deploys nothing at the branch; a VPN Gateway Connection points at it to build the actual tunnels. The classic-world sibling (without a Virtual WAN) is Azure Local Network Gateway.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **VPN Site** -- the ARM description of the branch, including its links (ARM assigns each link an ID the `link_ids` output republishes by name)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Virtual WAN** the site belongs to (sites are WAN-scoped so any hub's VPN gateway in the WAN can connect to them).
- **The branch's real details**: each link's public IP or FQDN, and either the reachable prefixes (`addressCidrs`) or per-link BGP.

## Deploy

### Console

Open the deployment store, find **Azure VPN Site**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Single-Link Branch** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVpnSite
metadata:
  name: branch-london
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  name: branch-london
  virtualWanId:
    valueFrom:
      kind: AzureVirtualWan
      name: corp-wan
      fieldPath: status.outputs.virtual_wan_id
  addressCidrs:
    - "192.168.10.0/24"
  links:
    - name: primary-isp
      ipAddress: "203.0.113.10"
      speedInMbps: 200
```

```shell
planton apply -f azure-vpn-site.yaml
```

This creates the address-book entry for a single-ISP London branch: one link at 203.0.113.10 with a /24 behind it, ready for a connection to point at. The site is free and provisions in seconds. An Infra Job tracks the provisioning in real time.

### InfraChart

In a branch-connectivity chart the order is: WAN → hub → VPN gateway, plus one **site** per branch → one connection per site. Wire the site's references with ValueFromRef:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: network-rg
      fieldPath: status.outputs.resource_group_name
  virtualWanId:
    valueFrom:
      kind: AzureVirtualWan
      name: corp-wan
      fieldPath: status.outputs.virtual_wan_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group and WAN first, then creates the site -- and each branch's connection references `vpn_site_id` and the name-keyed `link_ids`.

## Key Configuration

These are the most important decisions when configuring a site. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Links** -- the connectable unit. Each link needs a public endpoint (IP or FQDN); its ARM ID surfaces in `link_ids` keyed by the link's name, which is exactly what a connection's `vpnLinks` reference. Two links model a dual-ISP branch.

**Routing source** -- static `addressCidrs` (Azure routes those prefixes into the tunnels) or per-link `bgp` (the branch advertises its prefixes), or both. A site with neither routes nothing.

**Device metadata** -- `deviceVendor`/`deviceModel` are informational (portal display, SD-WAN partners); they change no behavior.

**What replaces the site** -- `name`, `region`, `resourceGroup`, and `virtualWanId` are fixed at creation; everything else updates in place. But a connection pins each link by ARM ID, so renaming or removing a CONNECTED link is a far-side change that breaks the tunnel -- coordinate link edits with the connection, not as a site-only edit.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureVirtualWan** | `virtualWanId` | `status.outputs.virtual_wan_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `vpn_site_id` | ARM ID of the site | A connection's `remoteVpnSiteId` |
| `link_ids` | ARM ID of each link, keyed by link name | A connection link's `vpnSiteLinkId` (`status.outputs.link_ids.primary-isp`) |

The outputs also carry `vpn_site_name` -- connections reference the site by ARM ID, so the name has no ValueFromRef consumer.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single-link branch** -- one ISP, static prefixes. Start from the **Single-Link Branch** preset.

**Dual-link BGP branch** -- two ISPs with per-link BGP for active-active connectivity. Start from the **Dual-Link BGP Branch** preset.

## Works With

- [**Azure Virtual WAN**](/infra-catalog/azure-virtual-wan) -- the WAN the site belongs to
- [**Azure VPN Gateway**](/infra-catalog/azure-vpn-gateway) -- the hub gateway branches connect to
- [**Azure VPN Gateway Connection**](/infra-catalog/azure-vpn-gateway-connection) -- the tunnels that point at this site
