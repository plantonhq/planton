# CloudflareZeroTrustTunnelVirtualNetwork

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustTunnelVirtualNetworkSpec configures a Cloudflare Tunnel virtual
network: an isolated routing segment that lets the same private CIDR (for example
10.0.0.0/8) be connected through more than one tunnel without collision. Routes
(CloudflareZeroTrustTunnelRoute) attach a private network to a tunnel within one
virtual network, and WARP clients select which virtual network to reach. A virtual
network is account-scoped and outlives any individual tunnel, so it is its own
first-class kind rather than a field on the tunnel.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustTunnelVirtualNetwork
metadata:
  name: test-vnet
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: prod-vnet
  comment: "Isolates the production 10.0.0.0/8 overlap from staging"
  isDefaultNetwork: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.comment` | `string` |  |  |  |
| `spec.isDefaultNetwork` | `bool` |  | `false` |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this virtual network.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.name

`string` · required

A user-friendly name for the virtual network (shown in the Zero Trust dashboard
and used to disambiguate overlapping routes).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"100"}}

### spec.comment

`string`

Optional remark describing the virtual network's purpose.

- rule: {"string":{"maxLen":"1000"}}

### spec.isDefaultNetwork

`bool` · optional (explicit presence)

When true, this virtual network becomes the account default: routes and WARP
clients that do not name a virtual network use it. Exactly one virtual network
can be the default at a time. Defaults to false.

- default: `false`

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustTunnelVirtualNetwork, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.virtual_network_id` | `string` | The Cloudflare-assigned UUID of the virtual network. A route (CloudflareZeroTrustTunnelRoute) references this value to bind a private network to this segment. |
| `status.outputs.virtual_network_name` | `string` | The virtual network name (echoed for convenience). |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareZeroTrustAccessInfrastructureTarget | `spec.ip.ipv4.virtualNetworkId` | `status.outputs.virtual_network_id` |
| CloudflareZeroTrustAccessInfrastructureTarget | `spec.ip.ipv6.virtualNetworkId` | `status.outputs.virtual_network_id` |
| CloudflareZeroTrustDeviceCustomProfile | `spec.virtualNetworks.allowed` | `status.outputs.virtual_network_id` |
| CloudflareZeroTrustDeviceCustomProfile | `spec.virtualNetworks.defaultVirtualNetworkId` | `status.outputs.virtual_network_id` |
| CloudflareZeroTrustDeviceDefaultProfile | `spec.virtualNetworks.allowed` | `status.outputs.virtual_network_id` |
| CloudflareZeroTrustDeviceDefaultProfile | `spec.virtualNetworks.defaultVirtualNetworkId` | `status.outputs.virtual_network_id` |
| CloudflareZeroTrustGatewayPolicy | `spec.ruleSettings.dnsResolvers.ipv4[].vnetId` | `status.outputs.virtual_network_id` |
| CloudflareZeroTrustGatewayPolicy | `spec.ruleSettings.dnsResolvers.ipv6[].vnetId` | `status.outputs.virtual_network_id` |
| CloudflareZeroTrustTunnelRoute | `spec.virtualNetworkId` | `status.outputs.virtual_network_id` |

## See Also

- [Overview](../README.md)
