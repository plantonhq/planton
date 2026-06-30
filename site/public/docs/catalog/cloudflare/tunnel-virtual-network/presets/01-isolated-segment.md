---
title: "Preset: Isolated routing segment"
description: "A standard, non-default virtual network used to isolate a set of routes (and the overlapping private CIDRs behind them) from other segments in the account."
type: "preset"
rank: "01"
presetSlug: "01-isolated-segment"
componentSlug: "tunnel-virtual-network"
componentTitle: "Tunnel Virtual Network"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: Isolated routing segment

A standard, non-default virtual network used to isolate a set of routes (and the
overlapping private CIDRs behind them) from other segments in the account.

## When to use

- You need to connect a private network whose CIDR overlaps another already-connected
  network, and you want each reachable through its own tunnel.
- You are partitioning routing by environment (prod vs staging) or tenant.

## Key choices

- `name`: a descriptive segment name; WARP clients and routes refer to it.
- Leave `isDefaultNetwork` unset (defaults to `false`) for a normal segment. Set it on
  exactly one virtual network if you want unscoped routes/clients to land here.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |

## Composition

Attach a private network to a tunnel within this segment with a
`CloudflareZeroTrustTunnelRoute` that references
`status.outputs.virtual_network_id`.
