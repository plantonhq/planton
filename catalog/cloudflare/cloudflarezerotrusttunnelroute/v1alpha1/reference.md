# CloudflareZeroTrustTunnelRoute

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustTunnelRouteSpec configures a Cloudflare Tunnel route: it advertises
a private IP range (CIDR) as reachable through a specific tunnel, within a virtual
network. WARP clients and other tunnels can then reach that range. A route has an
independent lifecycle from the tunnel (you add or remove reachable networks without
touching the tunnel), and a tunnel commonly carries many routes — so it is its own
first-class kind, wired to its tunnel and virtual network by reference.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustTunnelRoute
metadata:
  name: test-route
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  network: "10.0.0.0/24"
  tunnelId:
    value: "b8f2e1c0-1111-2222-3333-444455556666"
  comment: "Private app subnet via prod tunnel"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.network` | `string` | yes |  |  |
| `spec.tunnelId` | `string \| valueFrom` | yes |  | CloudflareZeroTrustTunnel (`status.outputs.tunnel_id`) |
| `spec.virtualNetworkId` | `string \| valueFrom` |  |  | CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.comment` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this route.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.network

`string` · required

The private IPv4 or IPv6 range advertised by this route, in CIDR notation
(for example "10.0.0.0/24" or "2001:db8::/48"). Single hosts use a /32 or /128.

- rule: network must be an IPv4 or IPv6 range in CIDR notation, e.g. 10.0.0.0/24
- rule: {"required":true}

### spec.tunnelId

`string | valueFrom` · required

The tunnel that serves this network: a literal tunnel UUID, or a reference to a
CloudflareZeroTrustTunnel resource.

- references: CloudflareZeroTrustTunnel (`status.outputs.tunnel_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustTunnel, name: <that resource's name>, fieldPath: status.outputs.tunnel_id}} -- a bare string does not parse

### spec.virtualNetworkId

`string | valueFrom`

The virtual network this route belongs to: a literal virtual network UUID, or a
reference to a CloudflareZeroTrustTunnelVirtualNetwork. Omit to use the account's
default virtual network. Use distinct virtual networks to advertise overlapping
CIDRs through different tunnels without collision.

- references: CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustTunnelVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

### spec.comment

`string`

Optional remark describing the route.

- rule: {"string":{"maxLen":"1000"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustTunnelRoute, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.route_id` | `string` | The Cloudflare-assigned UUID of the route. |
| `status.outputs.network` | `string` | The private CIDR advertised by this route (echoed for convenience). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.tunnelId` | CloudflareZeroTrustTunnel | `status.outputs.tunnel_id` |
| `spec.virtualNetworkId` | CloudflareZeroTrustTunnelVirtualNetwork | `status.outputs.virtual_network_id` |

## See Also

- [Overview](../README.md)
