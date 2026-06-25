# Cloudflare Tunnel Route

Advertise a private CIDR as reachable through a Cloudflare Tunnel, within a virtual
network.

## What Gets Created

- A `cloudflare_zero_trust_tunnel_cloudflared_route`.

## Prerequisites

- A Cloudflare account ID.
- A tunnel (CloudflareZeroTrustTunnel) to serve the network.
- Optionally, a virtual network (CloudflareZeroTrustTunnelVirtualNetwork) to isolate
  overlapping CIDRs.

## Configuration Reference

**Required**

- `accountId` — Cloudflare account ID.
- `network` — private CIDR to advertise.
- `tunnelId` — tunnel that serves the network.

**Optional**

- `virtualNetworkId` — virtual network to scope the route (defaults to the account default).
- `comment` — remark describing the route.

## Stack Outputs

| Output | Description |
|---|---|
| `route_id` | The route UUID |
| `network` | The advertised CIDR |

## Related Components

- CloudflareZeroTrustTunnel
- CloudflareZeroTrustTunnelVirtualNetwork
