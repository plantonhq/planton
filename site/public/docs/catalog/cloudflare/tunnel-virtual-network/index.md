---
title: "Tunnel Virtual Network"
description: "Tunnel Virtual Network deployment documentation"
icon: "package"
order: 100
componentName: "cloudflarezerotrusttunnelvirtualnetwork"
---

# Cloudflare Tunnel Virtual Network

An isolated routing segment for Cloudflare Tunnel that lets overlapping private CIDRs
be reached through separate tunnels.

## What Gets Created

- A `cloudflare_zero_trust_tunnel_cloudflared_virtual_network`.

## Prerequisites

- A Cloudflare account ID.

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.
- `name` — user-friendly name for the virtual network.

**Optional**

- `comment` — remark describing the segment.
- `isDefaultNetwork` — make this the account default (exactly one at a time).

## Stack Outputs

| Output | Description |
|---|---|
| `virtual_network_id` | The virtual network UUID |
| `virtual_network_name` | The virtual network name |

## Related Components

- CloudflareZeroTrustTunnel
- CloudflareZeroTrustTunnelRoute
