# CloudflareZeroTrustTunnelRoute

Advertise a private IP range (CIDR) as reachable through a Cloudflare Tunnel, within a
virtual network. Once a route exists, WARP clients and other tunnels can reach hosts in
that range over the private network.

## Why a first-class route

A route has a lifecycle independent of its tunnel — you expose or withdraw a subnet
without touching the tunnel — and a tunnel commonly carries many routes. Modeling it as
its own node lets the resource graph wire `route → tunnel` and `route → virtual network`
as explicit edges, the way a subnet/route is its own node.

## Quick start

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareZeroTrustTunnelRoute
metadata:
  name: app-subnet
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  network: 10.0.0.0/24
  tunnelId:
    valueFrom:
      kind: CloudflareZeroTrustTunnel
      name: prod-tunnel
      fieldPath: status.outputs.tunnel_id
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `network` | yes | Private IPv4/IPv6 range in CIDR notation (e.g. `10.0.0.0/24`, `2001:db8::/48`) |
| `tunnelId` | yes | Tunnel that serves this network (literal UUID or `CloudflareZeroTrustTunnel` ref) |
| `virtualNetworkId` | no | Virtual network UUID (literal or `CloudflareZeroTrustTunnelVirtualNetwork` ref); omit for the account default |
| `comment` | no | Remark describing the route |

## Stack outputs

| Output | Description |
|---|---|
| `route_id` | The route UUID |
| `network` | The advertised CIDR |

## Composition

A route sits above a tunnel and (optionally) a virtual network. Use distinct virtual
networks to advertise overlapping CIDRs through different tunnels without collision.
