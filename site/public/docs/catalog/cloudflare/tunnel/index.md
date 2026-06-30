---
title: "Tunnel"
description: "Tunnel deployment documentation"
icon: "package"
order: 100
componentName: "cloudflarezerotrusttunnel"
---

# Cloudflare Tunnel

A secure, outbound-only connection from a private network to Cloudflare's edge, exposing
private services via public hostnames and/or WARP-reachable private routes.

## What Gets Created

- A `cloudflare_zero_trust_tunnel_cloudflared`.
- When `configSrc` is `cloudflare` and `ingress` is set, a
  `cloudflare_zero_trust_tunnel_cloudflared_config` (provisioned separately so editing
  ingress never recreates the tunnel).
- The connector run token is read via the tunnel token data source and exported.

## Prerequisites

- A Cloudflare account ID.
- A connector (cloudflared) to run with the exported token.
- For public hostnames, a DNS record CNAME'd to the tunnel CNAME target.

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.
- `name` — tunnel name.

**Optional**

- `configSrc` (`cloudflare` | `local`), `tunnelSecret` (sensitive).
- `ingress[]` (`hostname`, `service`, `path`, `originRequest`) — the last must be a catch-all.
- `originRequest.*` — origin connection settings, including Access enforcement.

## Stack Outputs

| Output | Description |
|---|---|
| `tunnel_id` | The tunnel UUID |
| `tunnel_cname` | CNAME target for public hostnames |
| `tunnel_token` | Connector run token (sensitive) |
| `tunnel_status` | Tunnel status |
| `account_tag` | Account tag |
| `created_on` | Creation timestamp |

## Related Components

- CloudflareZeroTrustTunnelRoute
- CloudflareZeroTrustTunnelVirtualNetwork
- CloudflareZeroTrustAccessApplication
- CloudflareDnsRecord
