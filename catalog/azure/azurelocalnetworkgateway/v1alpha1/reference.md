# AzureLocalNetworkGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureLocalNetworkGatewaySpec** defines a local network gateway --
Azure's DESCRIPTION of the on-premises side of a site-to-site VPN: the
VPN device's public endpoint (IP or FQDN) and the address space
reachable behind it. It deploys nothing on-premises and costs nothing
to keep; it is the address book entry an
AzureVirtualNetworkGatewayConnection points at.

**Routing comes from exactly one of two places**: static
address_spaces (Azure routes those prefixes into the tunnel) or
bgp_settings (the on-premises device advertises its prefixes over
BGP). ARM requires at least one of the two; with BGP, address_spaces
is usually left empty so learned routes win.

**ForceNew fields**: `name`, `region`, and `resource_group` -- every
other field updates in place.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureLocalNetworkGateway
metadata:
  name: test-local-network-gateway
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: hq-datacenter
  # The on-premises VPN device's public endpoint (exactly one of
  # gatewayAddress or gatewayFqdn).
  gatewayAddress: "203.0.113.10"
  # The prefixes reachable behind the device -- Azure routes these into
  # the tunnel. Leave empty only when bgpSettings carries the routing.
  addressSpaces:
    - "192.168.100.0/24"
    - "192.168.101.0/24"
  bgpSettings:
    asn: 65010
    # An IP INSIDE the tunnel (the device's tunnel interface), not the
    # device's public address.
    bgpPeeringAddress: "10.255.255.1"
    peerWeight: 0
  tags:
    purpose: hack-test
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.gatewayAddress` | `string` |  |  |  |
| `spec.gatewayFqdn` | `string` |  |  |  |
| `spec.addressSpaces` | `[]string` |  |  |  |
| `spec.bgpSettings` | `AzureLocalNetworkGatewayBgpSettings` |  |  |  |
| `spec.bgpSettings.asn` | `int64` |  |  |  |
| `spec.bgpSettings.bgpPeeringAddress` | `string` | yes |  |  |
| `spec.bgpSettings.peerWeight` | `int32` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the local network gateway object lives in, e.g.
"eastus". By convention the region of the virtual network gateway
that connects to it. Changing the region replaces the object.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the local network gateway is created in.
Can be a literal resource-group name or a reference to an
AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The local network gateway's name, unique within the resource group.
Name it after the site it describes ("hq-datacenter",
"branch-london") -- a connection references it per site. Changing
the name replaces the object.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.gatewayAddress

`string`

The on-premises VPN device's PUBLIC IPv4 address -- where Azure
sends tunnel traffic. Exactly one of gateway_address or
gateway_fqdn.

- rule: gateway_address must be an IPv4 address

### spec.gatewayFqdn

`string`

The on-premises VPN device's public FQDN (for sites whose public IP
changes -- Azure re-resolves it periodically). Exactly one of
gateway_address or gateway_fqdn.

### spec.addressSpaces

`[]string`

The address space reachable behind the on-premises device, in CIDR
notation -- the prefixes Azure routes into the tunnel. Order is not
significant. Leave empty only when bgp_settings carries the routing
(ARM requires at least one of the two).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.bgpSettings

`AzureLocalNetworkGatewayBgpSettings`

The on-premises BGP speaker -- set it to exchange routes dynamically
instead of (or alongside) static address_spaces. The connection must
enable BGP too.

### spec.bgpSettings.asn

`int64`

The on-premises BGP speaker's Autonomous System Number. Must differ
from the Azure gateway's ASN; 65515-65520 are Azure-reserved.

- rule: {"int64":{"gte":"1"}}

### spec.bgpSettings.bgpPeeringAddress

`string` · required

The on-premises BGP speaker's peering address -- an IP INSIDE the
tunnel (typically on the device's tunnel interface), not the
device's public address.

- rule: {"required":true,"string":{"ipv4":true}}

### spec.bgpSettings.peerWeight

`int32`

The weight added to routes learned from this peer, 0-100.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.tags

`map<string, string>`

Free-form tags applied to the object, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins.

## Validation Rules

- `exactly_one_gateway_endpoint`: Describe the on-premises endpoint with exactly one of gateway_address (static public IP) or gateway_fqdn (re-resolved name)
- `routing_source_required`: Azure needs a routing source for the site: static address_spaces, bgp_settings, or both -- an empty object routes nothing

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureLocalNetworkGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.local_network_gateway_id` | `string` | The Azure Resource Manager ID of the local network gateway -- what connections reference as local_network_gateway_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/localNetworkGateways/{name} |
| `status.outputs.local_network_gateway_name` | `string` | The name of the local network gateway resource. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureVirtualNetworkGateway | `spec.defaultLocalNetworkGatewayId` | `status.outputs.local_network_gateway_id` |
| AzureVirtualNetworkGatewayConnection | `spec.localNetworkGatewayId` | `status.outputs.local_network_gateway_id` |

## See Also

- [Overview](../README.md)
