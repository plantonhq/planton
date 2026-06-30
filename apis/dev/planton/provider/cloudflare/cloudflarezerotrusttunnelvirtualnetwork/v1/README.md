# CloudflareZeroTrustTunnelVirtualNetwork

Provision a Cloudflare Tunnel virtual network: an isolated routing segment that lets
the same private CIDR (for example `10.0.0.0/8`) be connected through more than one
tunnel without collision. Routes attach a private network to a tunnel *within* a
virtual network, and WARP clients pick which virtual network to reach.

## Why a first-class virtual network

A virtual network is account-scoped and outlives any individual tunnel — many routes
across many tunnels reference it. Modeling it as its own node (rather than a field on
a tunnel) lets the resource graph wire overlapping private networks as explicit,
independently-owned segments, the way a VPC or subnet is its own node.

## Quick start

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareZeroTrustTunnelVirtualNetwork
metadata:
  name: prod-vnet
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: prod-vnet
  comment: Isolates the production 10.0.0.0/8 overlap
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `name` | yes | User-friendly name (1–100 chars) |
| `comment` | no | Remark describing the segment's purpose |
| `isDefaultNetwork` | no | Make this the account default (exactly one at a time); defaults to `false` |

## Stack outputs

| Output | Description |
|---|---|
| `virtual_network_id` | The virtual network UUID (reference it from a tunnel route) |
| `virtual_network_name` | The virtual network name |

## Composition

Reference the virtual network from a `CloudflareZeroTrustTunnelRoute`:

```yaml
virtualNetworkId:
  valueFrom:
    kind: CloudflareZeroTrustTunnelVirtualNetwork
    name: prod-vnet
    fieldPath: status.outputs.virtual_network_id
```
