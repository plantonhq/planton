---
title: "Preset: Overlapping CIDR isolated in a virtual network"
description: "Advertise a CIDR that overlaps another already-connected network by scoping the route to its own virtual network, so the two never collide."
type: "preset"
rank: "02"
presetSlug: "02-isolated-overlap"
componentSlug: "tunnel-route"
componentTitle: "Tunnel Route"
provider: "cloudflare"
icon: "package"
order: 2
---

# Preset: Overlapping CIDR isolated in a virtual network

Advertise a CIDR that overlaps another already-connected network by scoping the route to
its own virtual network, so the two never collide.

## When to use

- Multiple sites/tenants each use the same private range (e.g. `10.0.0.0/8`) and must be
  reachable independently.

## Key choices

- `virtualNetworkId`: reference a dedicated `CloudflareZeroTrustTunnelVirtualNetwork`;
  each overlapping CIDR goes in its own virtual network through its own tunnel.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<tenant-a-tunnel>` | Tunnel serving tenant A's network |
| `<tenant-a-vnet>` | Virtual network isolating tenant A's CIDR |
