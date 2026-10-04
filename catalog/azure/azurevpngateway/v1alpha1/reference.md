# AzureVpnGateway

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureVpnGatewaySpec** defines a Virtual WAN VPN gateway -- the
managed site-to-site VPN terminator that lives INSIDE a virtual hub
(ARM allows one per hub). Branches described by AzureVpnSite objects
connect to it through AzureVpnGatewayConnection tunnels; the hub's
routing then distributes the branch routes. The classic-world
sibling (in a VNet's GatewaySubnet instead of a hub) is
AzureVirtualNetworkGateway.

**Capacity is scale units**: each scale unit buys 500 Mbps of
aggregate throughput across an active-active instance pair Azure
manages for you (no SKU tiers, no generation matrix -- the hub
handles that). The gateway bills from creation and provisions in
tens of minutes; plan lifecycle around both.

**BGP is instance-aware**: Azure runs two gateway instances and each
gets its own peering address. The asn/peer_weight pair is set at
creation; custom APIPA addresses per instance (for tunnels that need
fixed 169.254.x.x peering, e.g. AWS site-to-site) are the one part
Azure applies AFTER the gateway exists, so they update in place.

**ForceNew fields**: `name`, `region`, `resource_group`,
`virtual_hub_id`, `routing_preference`, and bgp_settings'
`asn`/`peer_weight` -- changing any of them replaces the gateway (a
30-45 minute create plus a 10-20 minute delete, and every connection
on it). `scale_unit`, tags, NAT rules, and the custom APIPA
addresses update in place.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureVpnGateway
metadata:
  name: test-vpn-gateway
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  name: hub-vpn-gateway
  virtualHubId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus
  # The unset optional fields apply ARM's defaults: routing preference
  # "Microsoft Network", scale unit 1.
  bgpRouteTranslationForNatEnabled: true
  bgpSettings:
    asn: 65515
    peerWeight: 10
    instance0BgpPeeringAddress:
      customIps:
        - "169.254.21.5"
    instance1BgpPeeringAddress:
      customIps:
        - "169.254.22.5"
  natRules:
    - name: branch-overlap
      mode: INGRESS_SNAT
      type: STATIC_NAT
      externalMappings:
        - addressSpace: "100.64.10.0/24"
      internalMappings:
        - addressSpace: "192.168.10.0/24"
  tags:
    purpose: hack-test
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.virtualHubId` | `string \| valueFrom` | yes |  | AzureVirtualHub (`status.outputs.virtual_hub_id`) |
| `spec.routingPreference` | `enum` |  | `MICROSOFT_NETWORK` |  |
| `spec.scaleUnit` | `int32` |  | `1` |  |
| `spec.bgpRouteTranslationForNatEnabled` | `bool` |  |  |  |
| `spec.bgpSettings` | `AzureVpnGatewayBgpSettings` |  |  |  |
| `spec.bgpSettings.asn` | `int64` |  |  |  |
| `spec.bgpSettings.peerWeight` | `int32` |  |  |  |
| `spec.bgpSettings.instance0BgpPeeringAddress` | `AzureVpnGatewayInstanceBgpPeeringAddress` |  |  |  |
| `spec.bgpSettings.instance0BgpPeeringAddress.customIps` | `[]string` | yes |  |  |
| `spec.bgpSettings.instance1BgpPeeringAddress` | `AzureVpnGatewayInstanceBgpPeeringAddress` |  |  |  |
| `spec.bgpSettings.instance1BgpPeeringAddress.customIps` | `[]string` | yes |  |  |
| `spec.natRules` | `[]AzureVpnGatewayNatRule` |  |  |  |
| `spec.natRules[].name` | `string` | yes |  |  |
| `spec.natRules[].mode` | `enum` |  |  |  |
| `spec.natRules[].type` | `enum` |  |  |  |
| `spec.natRules[].externalMappings` | `[]AzureVpnGatewayNatRuleMapping` | yes |  |  |
| `spec.natRules[].externalMappings[].addressSpace` | `string` | yes |  |  |
| `spec.natRules[].externalMappings[].portRange` | `string` |  |  |  |
| `spec.natRules[].internalMappings` | `[]AzureVpnGatewayNatRuleMapping` | yes |  |  |
| `spec.natRules[].internalMappings[].addressSpace` | `string` | yes |  |  |
| `spec.natRules[].internalMappings[].portRange` | `string` |  |  |  |
| `spec.natRules[].ipConfiguration` | `enum` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the gateway lives in. Must match the hub's region
(the gateway deploys into the hub). Changing it replaces the
gateway.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the gateway is created in. Can be a
literal resource-group name or a reference to an
AzureResourceGroup's name output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The gateway's name, unique within the resource group. Changing the
name replaces the gateway.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.virtualHubId

`string | valueFrom` · required

The virtual hub the gateway deploys into -- references an
AzureVirtualHub's ARM ID. One VPN gateway per hub (ARM's rule).
Fixed at creation.

- references: AzureVirtualHub (`status.outputs.virtual_hub_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualHub, name: <that resource's name>, fieldPath: status.outputs.virtual_hub_id}} -- a bare string does not parse

### spec.routingPreference

`enum` · optional (explicit presence)

How tunnel traffic reaches the internet-facing branch endpoints:
MICROSOFT_NETWORK (ARM's default -- ride Microsoft's backbone as
long as possible) or INTERNET (hot-potato: exit to the public
internet close to the gateway). Fixed at creation.

- default: `MICROSOFT_NETWORK`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_vpn_gateway_routing_preference_unspecified` -- Not specified -- MICROSOFT_NETWORK (ARM's default) applies.
- `MICROSOFT_NETWORK` -- Ride Microsoft's backbone as long as possible (wire value "Microsoft Network" -- ARM's default).
- `INTERNET` -- Exit to the public internet close to the gateway (wire value "Internet").

### spec.scaleUnit

`int32` · optional (explicit presence)

The gateway's aggregate capacity in scale units (500 Mbps each,
across the managed instance pair). Unset applies the provider's
default of 1. Updates in place.

- default: `1`
- rule: {"int32":{"gte":0}}

### spec.bgpRouteTranslationForNatEnabled

`bool`

Translate BGP-learned routes to the NAT-translated address space
(post-NAT prefixes are what BGP advertises). Only meaningful when
nat_rules are configured on BGP-enabled tunnels. Off is ARM's
default. Updates in place.

### spec.bgpSettings

`AzureVpnGatewayBgpSettings`

The gateway's BGP speaker. Leave unset to accept Azure's defaults
(ASN 65515, weight 0); set it to pin custom APIPA peering
addresses per instance. asn and peer_weight are fixed at creation.

### spec.bgpSettings.asn

`int64`

The gateway's Autonomous System Number. Azure currently pins
Virtual WAN VPN gateways at 65515 -- set 65515 (or leave the whole
bgp_settings unset) unless Azure lifts the restriction. Fixed at
creation.

- rule: {"int64":{"gte":"1"}}

### spec.bgpSettings.peerWeight

`int32`

The weight added to routes learned from this gateway, 0-100.
Fixed at creation.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.bgpSettings.instance0BgpPeeringAddress

`AzureVpnGatewayInstanceBgpPeeringAddress`

Custom APIPA (169.254.x.x) BGP peering addresses for the FIRST
gateway instance. Azure assigns the defaults at creation and
applies these on top afterwards -- they update in place.

### spec.bgpSettings.instance0BgpPeeringAddress.customIps

`[]string` · required

The custom APIPA addresses (169.254.21.0-169.254.22.255 per
Azure's rule) this instance peers from -- what a connection link's
custom_bgp_addresses select among.

- rule: {"repeated":{"minItems":"1","items":{"string":{"ipv4":true}}}}

### spec.bgpSettings.instance1BgpPeeringAddress

`AzureVpnGatewayInstanceBgpPeeringAddress`

Custom APIPA BGP peering addresses for the SECOND gateway
instance (the gateway always runs an active-active pair).

### spec.bgpSettings.instance1BgpPeeringAddress.customIps

`[]string` · required

The custom APIPA addresses (169.254.21.0-169.254.22.255 per
Azure's rule) this instance peers from -- what a connection link's
custom_bgp_addresses select among.

- rule: {"repeated":{"minItems":"1","items":{"string":{"ipv4":true}}}}

### spec.natRules

`[]AzureVpnGatewayNatRule`

NAT rules on the gateway -- address translation that tunnels opt
into via a connection link's egress/ingress NAT rule id lists
(needed when branch address spaces overlap). Each rule deploys as
its own ARM child of the gateway; the gateway publishes each
rule's ARM id in the nat_rule_ids output.

### spec.natRules[].name

`string` · required

The rule's name, unique on the gateway. The rule's ARM id surfaces
in the gateway's nat_rule_ids output under this name. Changing the
name replaces the rule.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.natRules[].mode

`enum`

The translation direction: EGRESS_SNAT (translate the Azure-side
source -- the default) or INGRESS_SNAT (translate the branch-side
source). Fixed at creation.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_vpn_gateway_nat_rule_mode_unspecified` -- Not specified -- uses EGRESS_SNAT, the default direction.
- `EGRESS_SNAT` -- Translate the Azure-side source address space (wire value "EgressSnat").
- `INGRESS_SNAT` -- Translate the branch-side source address space (wire value "IngressSnat").

### spec.natRules[].type

`enum`

The translation type: STATIC_NAT (one-to-one, no port translation
-- the default) or DYNAMIC_NAT (many-to-one with port
translation). Fixed at creation.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_vpn_gateway_nat_rule_type_unspecified` -- Not specified -- uses STATIC_NAT, one-to-one translation.
- `STATIC_NAT` -- One-to-one address translation without ports (wire value "Static").
- `DYNAMIC_NAT` -- Many-to-one translation with port translation (wire value "Dynamic").

### spec.natRules[].externalMappings

`[]AzureVpnGatewayNatRuleMapping` · required

The external (post-translation, as seen by the remote side)
mappings.

- rule: {"repeated":{"minItems":"1"}}

### spec.natRules[].externalMappings[].addressSpace

`string` · required

The address space in CIDR notation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.natRules[].externalMappings[].portRange

`string`

The port range (e.g. "100-200"). Dynamic rules only.

### spec.natRules[].internalMappings

`[]AzureVpnGatewayNatRuleMapping` · required

The internal (pre-translation, Azure-side) mappings.

- rule: {"repeated":{"minItems":"1"}}

### spec.natRules[].internalMappings[].addressSpace

`string` · required

The address space in CIDR notation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.natRules[].internalMappings[].portRange

`string`

The port range (e.g. "100-200"). Dynamic rules only.

### spec.natRules[].ipConfiguration

`enum`

Pin the rule to one gateway instance (INSTANCE_0 or INSTANCE_1).
Rarely needed -- leave unspecified to apply on both.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_vpn_gateway_nat_rule_ip_configuration_unspecified` -- Not specified -- the rule applies on both instances.
- `INSTANCE_0` -- The first gateway instance (wire value "Instance0").
- `INSTANCE_1` -- The second gateway instance (wire value "Instance1").

### spec.tags

`map<string, string>`

Free-form tags applied to the gateway, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins.

## Validation Rules

- `nat_rule_names_unique`: NAT rule names must be unique on the gateway -- each is the key the nat_rule_ids output uses

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureVpnGateway, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.vpn_gateway_id` | `string` | The Azure Resource Manager ID of the gateway -- what a connection references as its vpn_gateway_id. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/vpnGateways/{name} |
| `status.outputs.vpn_gateway_name` | `string` | The name of the gateway. |
| `status.outputs.bgp_asn` | `int64` | The gateway's BGP autonomous system number (65515 on today's Virtual WAN gateways) -- what branch devices configure as the remote ASN. |
| `status.outputs.public_ip_addresses` | `[]string` | The PUBLIC IPv4 address of each gateway instance (the active-active pair) -- what branch devices dial as their tunnel peer. Azure assigns them; there is no public-IP resource to bring. |
| `status.outputs.private_ip_addresses` | `[]string` | The private IPv4 address of each gateway instance -- the tunnel endpoints when connections use local_azure_ip_address_enabled (private peering over ExpressRoute). |
| `status.outputs.nat_rule_ids` | `map<string, string>` | The ARM ID of each NAT rule on the gateway, keyed by the rule's name from the spec -- what a connection link's egress/ingress_nat_rule_ids reference. Example valueFrom fieldPath: status.outputs.nat_rule_ids.branch-overlap |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.virtualHubId` | AzureVirtualHub | `status.outputs.virtual_hub_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureVpnGatewayConnection | `spec.vpnGatewayId` | `status.outputs.vpn_gateway_id` |
| AzureVpnGatewayConnection | `spec.vpnLinks[].egressNatRuleIds` | `status.outputs.nat_rule_ids` |
| AzureVpnGatewayConnection | `spec.vpnLinks[].ingressNatRuleIds` | `status.outputs.nat_rule_ids` |

## See Also

- [Overview](../README.md)
