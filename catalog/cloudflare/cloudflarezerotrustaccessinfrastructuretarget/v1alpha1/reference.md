# CloudflareZeroTrustAccessInfrastructureTarget

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustAccessInfrastructureTargetSpec registers an
infrastructure target: a server (by hostname and private IP) that Access
infrastructure applications grant SSH access to. Targets are the inventory
layer -- an infrastructure application selects targets by hostname or IP,
and Access brokers short-lived SSH access to them through the account's
tunnels.

A target is a plain CRUD object (real create, update, delete), unlike the
Zero Trust settings singletons.

## Example

```yaml
# Complete example manifest for CloudflareZeroTrustAccessInfrastructureTarget.
# Registers a server (hostname + private IP) that Access infrastructure
# applications broker SSH access to. The virtual-network reference is
# optional: omitted, the address joins the account's default virtual network.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessInfrastructureTarget
metadata:
  name: prod-db-1
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  hostname: prod-db-1
  ip:
    ipv4:
      ip_addr: 10.0.10.5
      virtual_network_id:
        value: f70ff985-a4ef-4643-bbbc-4a0ed4fc8415
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.hostname` | `string` | yes |  |  |
| `spec.ip` | `CloudflareZeroTrustAccessInfrastructureTargetIp` | yes |  |  |
| `spec.ip.ipv4` | `CloudflareZeroTrustAccessInfrastructureTargetIpInfo` |  |  |  |
| `spec.ip.ipv4.ipAddr` | `string` | yes |  |  |
| `spec.ip.ipv4.virtualNetworkId` | `string \| valueFrom` |  |  | CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.ip.ipv6` | `CloudflareZeroTrustAccessInfrastructureTargetIpInfo` |  |  |  |
| `spec.ip.ipv6.ipAddr` | `string` | yes |  |  |
| `spec.ip.ipv6.virtualNetworkId` | `string \| valueFrom` |  |  | CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`) |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the target belongs to.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.hostname

`string` · required

The target's hostname: up to 255 characters, letters, digits, dashes and
periods only, starting and ending alphanumeric, case-insensitive.
Hostnames are how infrastructure applications select groups of targets,
so a consistent naming scheme (e.g. "prod-db-1") is the access-control
surface itself.

- rule: hostname may contain only letters, digits, dashes and periods, and must start and end with a letter or digit (e.g. prod-db-1)
- rule: {"required":true,"string":{"maxLen":"255"}}

### spec.ip

`CloudflareZeroTrustAccessInfrastructureTargetIp` · required

The target's IP addressing. Declare the IPv4 arm, the IPv6 arm, or both
-- each binds one address inside one virtual network.

- rule: {"required":true}
- rule: declare at least one of ipv4 or ipv6

### spec.ip.ipv4

`CloudflareZeroTrustAccessInfrastructureTargetIpInfo`

The IPv4 address of the target.

### spec.ip.ipv4.ipAddr

`string` · required

The IP address. Cloudflare's schema leaves this optional, but a declared
family without an address is meaningless -- this spec requires it.

- rule: {"required":true}

### spec.ip.ipv4.virtualNetworkId

`string | valueFrom`

The virtual network the address belongs to: a literal virtual-network
UUID, or a reference to a CloudflareZeroTrustTunnelVirtualNetwork.
Omit to use the account's default virtual network.

- references: CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustTunnelVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

### spec.ip.ipv6

`CloudflareZeroTrustAccessInfrastructureTargetIpInfo`

The IPv6 address of the target.

### spec.ip.ipv6.ipAddr

`string` · required

The IP address. Cloudflare's schema leaves this optional, but a declared
family without an address is meaningless -- this spec requires it.

- rule: {"required":true}

### spec.ip.ipv6.virtualNetworkId

`string | valueFrom`

The virtual network the address belongs to: a literal virtual-network
UUID, or a reference to a CloudflareZeroTrustTunnelVirtualNetwork.
Omit to use the account's default virtual network.

- references: CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustTunnelVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustAccessInfrastructureTarget, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.target_id` | `string` | The Cloudflare-assigned UUID of the target. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.ip.ipv4.virtualNetworkId` | CloudflareZeroTrustTunnelVirtualNetwork | `status.outputs.virtual_network_id` |
| `spec.ip.ipv6.virtualNetworkId` | CloudflareZeroTrustTunnelVirtualNetwork | `status.outputs.virtual_network_id` |

## See Also

- [Overview](../README.md)
