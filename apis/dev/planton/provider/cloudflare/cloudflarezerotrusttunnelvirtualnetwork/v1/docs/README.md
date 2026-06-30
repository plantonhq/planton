# CloudflareZeroTrustTunnelVirtualNetwork — Research & Design Notes

## Purpose

A Cloudflare Tunnel virtual network (`cloudflare_zero_trust_tunnel_cloudflared_virtual_network`)
is an account-scoped routing segment. Its single reason to exist is to disambiguate
overlapping private IP ranges: two sites that both use `10.0.0.0/8` can each be reached
through their own tunnel by placing their routes in different virtual networks. WARP
clients then select which virtual network they are connecting to.

## 80/20 scope

The provider resource is intentionally small — `account_id`, `name`, `comment`,
`is_default_network`. All four are modeled. There are no deeper knobs to expose; this
is a complete representation of the resource.

- `is_default_network`: at most one virtual network per account may be the default.
  Routes and WARP clients that do not name a virtual network fall back to it. The
  provider also exposes a deprecated `is_default` alias for the same concept — only the
  current `is_default_network` field is modeled.

## Composition

- A `CloudflareZeroTrustTunnelRoute` references `status.outputs.virtual_network_id` to
  bind a private CIDR to a tunnel within this segment.
- The virtual network has no upstream dependencies; it sits at the foundation layer of
  a tunnel topology (create it before routes).

## Engine parity

Both engines create the one provider resource with identical inputs and emit the same
outputs (`virtual_network_id`, `virtual_network_name`). The `id` attribute is the
virtual network UUID in both Terraform and Pulumi.

## Gotchas

- `name` is mutable; changing it updates in place.
- Deleting a virtual network that still has routes attached fails — destroy the routes
  first (the resource graph handles this ordering via the route's foreign key).
