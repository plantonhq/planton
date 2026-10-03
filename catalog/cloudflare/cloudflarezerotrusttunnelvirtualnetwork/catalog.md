# Cloudflare Zero Trust Tunnel Virtual Network

Provisions a Cloudflare Tunnel virtual network: an isolated routing segment that lets the same private CIDR (for example `10.0.0.0/8`) be connected through more than one tunnel without collision. Routes (`CloudflareZeroTrustTunnelRoute`) attach a private network to a tunnel within one virtual network, and WARP clients select which virtual network to reach. A virtual network is account-scoped and outlives any individual tunnel.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Virtual Network** -- a named, account-scoped routing segment

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Cloudflare Tunnel edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** -- the account must have Cloudflare Zero Trust (Cloudflare One) set up.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Tunnel Virtual Network**, and click **Deploy**. The creation wizard captures the owning account, a name, an optional comment, and whether this virtual network is the account default. Start from the **Isolated routing segment** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustTunnelVirtualNetwork
metadata:
  name: prod-overlay
  org: acme-corp
  env: prod
spec:
  accountId: a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4
  name: prod-overlay
  comment: prod data center segment
  isDefaultNetwork: false
```

```shell
planton apply -f cloudflare-zero-trust-tunnel-virtual-network.yaml
```

This creates a named routing segment for the prod data center. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a virtual network. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Account (`accountId`)** -- The account that owns the virtual network. Immutable -- changing it replaces the virtual network.

**Name (`name`)** -- A human-readable name shown in the Zero Trust dashboard and used to disambiguate overlapping routes.

**Account Default (`isDefaultNetwork`)** -- When on, routes and WARP clients that do not name a virtual network use this one. Exactly one virtual network can be the default at a time -- promoting this one demotes the previous default.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies -- a virtual network is a self-contained, account-scoped leaf.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `virtual_network_id` | The Cloudflare-assigned UUID | Referenced by a CloudflareZeroTrustTunnelRoute's `virtualNetworkId` to bind a private network to this segment |

`status.outputs` also echoes `virtual_network_name` back from the spec.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Isolated segment** -- a named virtual network keeps a site's overlapping CIDRs separate from other sites. Start from the **Isolated routing segment** preset.

**Default network** -- mark one virtual network as the account default for the common single-overlay case. Start from the **Account default virtual network** preset.

## Works With

- [**Cloudflare Zero Trust Tunnel Route**](/infra-catalog/cloudflare-zero-trust-tunnel-route) -- routes advertise CIDRs within this virtual network
- [**Cloudflare Zero Trust Tunnel**](/infra-catalog/cloudflare-zero-trust-tunnel) -- the tunnel a route binds a network to
