# CloudflareZeroTrustTunnelRoute — Research & Design Notes

## Purpose

A Cloudflare Tunnel route (`cloudflare_zero_trust_tunnel_cloudflared_route`) tells
Cloudflare that a private IP range is reachable through a particular tunnel. It is the
private-networking half of Cloudflare Tunnel (as opposed to public-hostname ingress):
WARP clients and other tunnels route traffic for that CIDR through the connector.

## 80/20 scope

The provider resource is small — `account_id`, `network`, `tunnel_id`,
`virtual_network_id`, `comment`. All are modeled. `network` and `tunnel_id` are
required; `virtual_network_id` defaults to the account's default virtual network when
omitted.

## Composition

- `tunnel_id` references `CloudflareZeroTrustTunnel.status.outputs.tunnel_id`.
- `virtual_network_id` references
  `CloudflareZeroTrustTunnelVirtualNetwork.status.outputs.virtual_network_id`.
- Overlapping CIDRs are disambiguated by placing their routes in different virtual
  networks — the canonical reason virtual networks exist.

## Engine parity

Both engines create the one provider resource with identical inputs and emit the same
outputs (`route_id`, `network`). The `id` attribute is the route UUID in both engines.

## Gotchas

- `network` must be a valid CIDR; the spec enforces a basic IPv4/IPv6-with-prefix shape,
  and the provider rejects malformed ranges at apply.
- A route is unique per (tunnel, network, virtual network). Re-advertising the same CIDR
  in the same virtual network through a second tunnel requires a different virtual
  network.
