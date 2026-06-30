---
title: "Preset: Private subnet via a tunnel"
description: "Advertise a private subnet through a tunnel so WARP clients can reach hosts in it — the most common private-networking route."
type: "preset"
rank: "01"
presetSlug: "01-private-subnet"
componentSlug: "tunnel-route"
componentTitle: "Tunnel Route"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: Private subnet via a tunnel

Advertise a private subnet through a tunnel so WARP clients can reach hosts in it — the
most common private-networking route.

## When to use

- You want to reach a private network (databases, internal apps) over Cloudflare Tunnel
  without exposing public hostnames.

## Key choices

- `network`: the private CIDR to advertise (use a `/32` or `/128` for a single host).
- `tunnelId`: reference the `CloudflareZeroTrustTunnel` so the graph deploys it first.
- Leave `virtualNetworkId` unset to use the account default; set it (referencing a
  `CloudflareZeroTrustTunnelVirtualNetwork`) when CIDRs overlap.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<tunnel-name>` | Name of the tunnel that serves this network |
