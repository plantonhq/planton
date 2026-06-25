# Preset: Private-network connector (for WARP access)

A tunnel with no public hostnames, used purely to make private IP ranges reachable to
WARP clients via routes.

## When to use

- You want enrolled devices to reach internal subnets (databases, internal apps) over the
  private network, with no public ingress.

## Key choices

- No `ingress` rules — reachability comes from `CloudflareZeroTrustTunnelRoute` resources
  that reference this tunnel's `status.outputs.tunnel_id`.
- Run the connector with `status.outputs.tunnel_token`.

## Composition

Pair this with one or more `CloudflareZeroTrustTunnelRoute` resources (optionally scoped
to a `CloudflareZeroTrustTunnelVirtualNetwork` when CIDRs overlap).

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
