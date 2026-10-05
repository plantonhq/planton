# AzureExpressRouteCircuitPeering

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureExpressRouteCircuitPeeringSpec** defines a peering (a BGP
routing configuration) on an ExpressRoute circuit. The circuit is the
physical pipe; peerings are what make routes flow through it. A
circuit carries AT MOST ONE peering of each type -- the peering type IS
the ARM child's name -- so a fully-used circuit has up to two peerings
(private + Microsoft).

**Peering types**:
- **AZURE_PRIVATE_PEERING**: routes your VNets' private address space
  over the circuit -- the type virtually every deployment needs. A
  virtual network gateway (type EXPRESS_ROUTE) connects the VNet to
  this peering.
- **MICROSOFT_PEERING**: routes Microsoft public services (Microsoft
  365, Azure public IPs) over the circuit. Requires
  microsoft_peering_config with your registered public prefixes, and
  optionally a route filter to select service communities.
- **AZURE_PUBLIC_PEERING**: deprecated by Azure for new circuits --
  Microsoft peering is its successor.

**Addressing**: IPv4 peering uses a /30 pair (primary + secondary --
one per physical link); IPv6 adds /126 pairs via the `ipv6` block.
The VLAN id must be unique on the circuit.

**Provisioning order**: ARM accepts and STORES peering configuration
on a circuit whose provider state is still "NotProvisioned"
(live-verified for private peering) -- deploying the peering before
the carrier handoff is legal sequencing, and the BGP session simply
cannot establish until the provider completes the cross-connect.
Microsoft peering validates advertised public prefixes server-side
and has not been exercised on an unprovisioned circuit -- do not
assume its create behaves the same way.

**ForceNew fields**: `express_route_circuit_name`, `resource_group`,
and `peering_type` (the ARM identity) -- the routing configuration
itself updates in place.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureExpressRouteCircuitPeering
metadata:
  name: test-express-route-circuit-peering
spec:
  resourceGroup:
    value: test-rg
  # ARM addresses peerings as children of the circuit NAME.
  expressRouteCircuitName:
    value: hq-circuit
  # Private peering: VNet connectivity -- the type virtually every
  # deployment needs. The type is the ARM identity (one per circuit).
  peeringType: AZURE_PRIVATE_PEERING
  # Provider-assigned; unique on the circuit.
  vlanId: 100
  # One /30 per physical link: your router takes the first usable
  # address, Microsoft's edge the second.
  primaryPeerAddressPrefix: "192.168.16.0/30"
  secondaryPeerAddressPrefix: "192.168.16.4/30"
  peerAsn: 65010
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.expressRouteCircuitName` | `string \| valueFrom` | yes |  | AzureExpressRouteCircuit (`status.outputs.express_route_circuit_name`) |
| `spec.peeringType` | `enum` |  |  |  |
| `spec.vlanId` | `int32` |  |  |  |
| `spec.primaryPeerAddressPrefix` | `string` |  |  |  |
| `spec.secondaryPeerAddressPrefix` | `string` |  |  |  |
| `spec.ipv4Enabled` | `bool` |  | `true` |  |
| `spec.peerAsn` | `int64` |  |  |  |
| `spec.sharedKey` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.microsoftPeeringConfig` | `AzureExpressRouteCircuitPeeringMicrosoftConfig` |  |  |  |
| `spec.microsoftPeeringConfig.advertisedPublicPrefixes` | `[]string` | yes |  |  |
| `spec.microsoftPeeringConfig.customerAsn` | `int64` |  |  |  |
| `spec.microsoftPeeringConfig.routingRegistryName` | `string` |  | `NONE` |  |
| `spec.microsoftPeeringConfig.advertisedCommunities` | `[]string` |  |  |  |
| `spec.ipv6` | `AzureExpressRouteCircuitPeeringIpv6` |  |  |  |
| `spec.ipv6.primaryPeerAddressPrefix` | `string` | yes |  |  |
| `spec.ipv6.secondaryPeerAddressPrefix` | `string` | yes |  |  |
| `spec.ipv6.enabled` | `bool` |  | `true` |  |
| `spec.ipv6.routeFilterId` | `string` |  |  |  |
| `spec.ipv6.microsoftPeering` | `AzureExpressRouteCircuitPeeringMicrosoftConfig` |  |  |  |
| `spec.ipv6.microsoftPeering.advertisedPublicPrefixes` | `[]string` | yes |  |  |
| `spec.ipv6.microsoftPeering.customerAsn` | `int64` |  |  |  |
| `spec.ipv6.microsoftPeering.routingRegistryName` | `string` |  | `NONE` |  |
| `spec.ipv6.microsoftPeering.advertisedCommunities` | `[]string` |  |  |  |
| `spec.routeFilterId` | `string` |  |  |  |
| `spec.connections` | `[]AzureExpressRouteCircuitPeeringConnection` |  |  |  |
| `spec.connections[].name` | `string` | yes |  |  |
| `spec.connections[].peerPeeringId` | `string \| valueFrom` | yes |  | AzureExpressRouteCircuitPeering (`status.outputs.express_route_circuit_peering_id`) |
| `spec.connections[].addressPrefixIpv4` | `string` | yes |  |  |
| `spec.connections[].addressPrefixIpv6` | `string` |  |  |  |
| `spec.connections[].authorizationKey` | `string \| valueFrom` (sensitive) |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the parent circuit lives in. Can be a
literal resource-group name or a reference to an
AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.expressRouteCircuitName

`string | valueFrom` · required

The parent ExpressRoute circuit, by NAME (ARM addresses peerings as
children of the circuit name, not the circuit id). Can be a literal
name or a reference to an AzureExpressRouteCircuit's name output.

- references: AzureExpressRouteCircuit (`status.outputs.express_route_circuit_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureExpressRouteCircuit, name: <that resource's name>, fieldPath: status.outputs.express_route_circuit_name}} -- a bare string does not parse

### spec.peeringType

`enum`

The peering type -- also the ARM child's NAME, so a circuit carries
at most one peering of each type. Fixed at creation.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_express_route_circuit_peering_type_unspecified` -- Not specified -- invalid: the type is the peering's ARM identity (see the peering_type_required contract).
- `AZURE_PRIVATE_PEERING` -- Private connectivity to your VNets -- what an EXPRESS_ROUTE virtual network gateway connects to. Wire value "AzurePrivatePeering".
- `AZURE_PUBLIC_PEERING` -- DEPRECATED by Azure for new circuits -- Microsoft peering is the successor. Wire value "AzurePublicPeering".
- `MICROSOFT_PEERING` -- Microsoft public services (Microsoft 365, Azure public IPs) over the circuit. Wire value "MicrosoftPeering".

### spec.vlanId

`int32`

The 802.1Q VLAN id the provider tags this peering's traffic with,
1-4094. Must be unique across the circuit's peerings (your
connectivity provider assigns or confirms it).

- rule: {"int32":{"lte":4094,"gte":1}}

### spec.primaryPeerAddressPrefix

`string`

The IPv4 /30 for the PRIMARY link's point-to-point BGP session --
your router gets the first usable address, Microsoft's the second.
Set with secondary_peer_address_prefix (the pair travels together).

### spec.secondaryPeerAddressPrefix

`string`

The IPv4 /30 for the SECONDARY link's point-to-point BGP session.
Set with primary_peer_address_prefix.

### spec.ipv4Enabled

`bool` · optional (explicit presence)

Whether the IPv4 peering is enabled. Disabling keeps the
configuration but withdraws the routes -- useful for maintenance
without tearing the peering down.

- default: `true`

### spec.peerAsn

`int64`

Your side's BGP Autonomous System Number. Leave 0 to let Azure
record the ASN your router presents; set it when the provider
requires it declared up front.

- rule: {"int64":{"gte":"0"}}

### spec.sharedKey

`string | valueFrom` · sensitive

The MD5 hash key for the BGP sessions, 1-25 characters. Reference a
secret rather than embedding the literal in manifests. ARM never
returns it on reads.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.microsoftPeeringConfig

`AzureExpressRouteCircuitPeeringMicrosoftConfig`

MICROSOFT_PEERING only: the public-prefix advertisement contract --
which of your registered public prefixes Microsoft accepts routes
for. REQUIRED (by ARM) when the peering type is MICROSOFT_PEERING
and IPv4 prefixes are configured.

### spec.microsoftPeeringConfig.advertisedPublicPrefixes

`[]string` · required

The public IPv4 prefixes you will advertise to Microsoft. Each must
be registered to you (or to customer_asn) in an internet routing
registry -- Microsoft validates ownership before activating the
peering.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.microsoftPeeringConfig.customerAsn

`int64`

The customer ASN the prefixes are registered under, when routes are
advertised on behalf of a downstream customer of yours. 0 (the
default) means the prefixes are registered to peer_asn itself.

- rule: {"int64":{"gte":"0"}}

### spec.microsoftPeeringConfig.routingRegistryName

`string` · optional (explicit presence)

The internet routing registry Microsoft validates prefix ownership
against (e.g. "ARIN", "RIPE", "AFRINIC"). "NONE" (the default) lets
Microsoft use its standard validation.

- default: `NONE`

### spec.microsoftPeeringConfig.advertisedCommunities

`[]string`

BGP community values tagged onto the advertised routes (e.g.
"12076:20000" service selectors).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.ipv6

`AzureExpressRouteCircuitPeeringIpv6`

The IPv6 half of the peering: /126 address pairs and (for Microsoft
peering) its own advertisement contract. Private and Microsoft
peering only -- ARM rejects IPv6 on the deprecated public peering.

### spec.ipv6.primaryPeerAddressPrefix

`string` · required

The IPv6 /126 for the PRIMARY link's BGP session.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.ipv6.secondaryPeerAddressPrefix

`string` · required

The IPv6 /126 for the SECONDARY link's BGP session.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.ipv6.enabled

`bool` · optional (explicit presence)

Whether the IPv6 peering is enabled. Disabling withdraws the IPv6
routes while keeping the configuration.

- default: `true`

### spec.ipv6.routeFilterId

`string`

MICROSOFT_PEERING only: a Route Filter's ARM id for the IPv6
session's service communities.

### spec.ipv6.microsoftPeering

`AzureExpressRouteCircuitPeeringMicrosoftConfig`

MICROSOFT_PEERING only: the IPv6 advertisement contract (public
IPv6 prefixes registered to you).

### spec.ipv6.microsoftPeering.advertisedPublicPrefixes

`[]string` · required

The public IPv4 prefixes you will advertise to Microsoft. Each must
be registered to you (or to customer_asn) in an internet routing
registry -- Microsoft validates ownership before activating the
peering.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.ipv6.microsoftPeering.customerAsn

`int64`

The customer ASN the prefixes are registered under, when routes are
advertised on behalf of a downstream customer of yours. 0 (the
default) means the prefixes are registered to peer_asn itself.

- rule: {"int64":{"gte":"0"}}

### spec.ipv6.microsoftPeering.routingRegistryName

`string` · optional (explicit presence)

The internet routing registry Microsoft validates prefix ownership
against (e.g. "ARIN", "RIPE", "AFRINIC"). "NONE" (the default) lets
Microsoft use its standard validation.

- default: `NONE`

### spec.ipv6.microsoftPeering.advertisedCommunities

`[]string`

BGP community values tagged onto the advertised routes (e.g.
"12076:20000" service selectors).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.routeFilterId

`string`

MICROSOFT_PEERING only: a Route Filter's ARM id selecting which
Microsoft service communities (BGP communities for Microsoft 365,
Azure regions) are advertised to you. Without one, Microsoft
peering advertises nothing.

### spec.connections

`[]AzureExpressRouteCircuitPeeringConnection`

GLOBAL REACH: connections from THIS circuit's private peering to
OTHER circuits' private peerings, linking the on-premises sites
behind them across the Microsoft backbone. Each entry creates an
ARM circuit-connection child under this peering. Private peering
only.

### spec.connections[].name

`string` · required

The connection's name, unique on the peering. 1-80 characters; must
begin with a letter or number, end with a letter, number, or
underscore, and may contain only letters, numbers, underscores,
periods, or hyphens. Fixed at creation.

- rule: Connection names start with a letter or number, end with a letter, number, or underscore, and may contain alphanumerics, underscores, periods, and hyphens
- rule: {"required":true,"string":{"minLen":"1","maxLen":"80"}}

### spec.connections[].peerPeeringId

`string | valueFrom` · required

The FAR side: the other circuit's private peering, by ARM id.
References another AzureExpressRouteCircuitPeering's id output.
Fixed at creation.

- references: AzureExpressRouteCircuitPeering (`status.outputs.express_route_circuit_peering_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureExpressRouteCircuitPeering, name: <that resource's name>, fieldPath: status.outputs.express_route_circuit_peering_id}} -- a bare string does not parse

### spec.connections[].addressPrefixIpv4

`string` · required

A /29 IPv4 block used for the connection's tunnel addressing. Must
not overlap either side's address space. Fixed at creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.connections[].addressPrefixIpv6

`string`

A /125 IPv6 block for the connection's tunnel addressing.
ARM-enforced at deploy time: not allowed when the parent circuit is
ExpressRoute-Direct (port) based.

### spec.connections[].authorizationKey

`string | valueFrom` · sensitive

The authorization key redeemed when the FAR circuit belongs to a
different subscription -- issued by that circuit's authorizations
list. A UUID. Reference a secret rather than embedding the literal
in manifests. ARM masks it on reads, so an imported connection
legitimately plans an in-place update on it.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

## Validation Rules

- `peering_type_required`: Choose the peering type explicitly -- AZURE_PRIVATE_PEERING for VNet connectivity (the common case) or MICROSOFT_PEERING for Microsoft public services
- `address_prefix_pair_travels_together`: primary_peer_address_prefix and secondary_peer_address_prefix are set together -- one /30 per physical link
- `route_filter_is_microsoft_peering_only`: route_filter_id applies to MICROSOFT_PEERING only -- private peering carries no service communities to filter
- `microsoft_config_is_microsoft_peering_only`: microsoft_peering_config applies to MICROSOFT_PEERING only
- `microsoft_ipv4_requires_config`: Microsoft peering with IPv4 prefixes needs microsoft_peering_config (the advertised-prefix contract)
- `microsoft_config_requires_ipv4_prefixes`: microsoft_peering_config configures the IPv4 session -- set primary_peer_address_prefix and secondary_peer_address_prefix with it
- `ipv6_not_on_public_peering`: IPv6 configuration applies to AZURE_PRIVATE_PEERING and MICROSOFT_PEERING only (public peering is deprecated and IPv4-only)
- `connections_are_private_peering_only`: Global Reach connections link PRIVATE peerings -- set peering_type to AZURE_PRIVATE_PEERING or remove connections
- `connection_names_unique`: Global Reach connection names must be unique on the peering

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureExpressRouteCircuitPeering, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.express_route_circuit_peering_id` | `string` | The Azure Resource Manager ID of the peering -- what a Global Reach connection on ANOTHER circuit references as peer_peering_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/expressRouteCircuits/{circuit}/peerings/{type} |
| `status.outputs.azure_asn` | `int64` | Microsoft's Autonomous System Number on this peering (12076 on public Azure) -- configure it as the BGP neighbor ASN on your routers. |
| `status.outputs.primary_azure_port` | `string` | The Microsoft-edge identifier of the PRIMARY physical port the peering rides on. |
| `status.outputs.secondary_azure_port` | `string` | The Microsoft-edge identifier of the SECONDARY physical port. |
| `status.outputs.connection_ids` | `map<string, string>` | The ARM ID of each Global Reach connection created from this peering's `connections` list, keyed by the connection's name. Example valueFrom fieldPath: status.outputs.connection_ids.hq-to-branch |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.expressRouteCircuitName` | AzureExpressRouteCircuit | `status.outputs.express_route_circuit_name` |
| `spec.connections[].peerPeeringId` | AzureExpressRouteCircuitPeering | `status.outputs.express_route_circuit_peering_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureExpressRouteCircuitPeering | `spec.connections[].peerPeeringId` | `status.outputs.express_route_circuit_peering_id` |
| AzureExpressRouteGateway | `spec.connections[].expressRouteCircuitPeeringId` | `status.outputs.express_route_circuit_peering_id` |

## See Also

- [Overview](../README.md)
