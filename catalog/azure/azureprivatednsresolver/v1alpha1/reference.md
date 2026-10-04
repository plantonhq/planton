# AzurePrivateDnsResolver

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzurePrivateDnsResolverSpec** defines an Azure DNS Private Resolver
-- the managed DNS proxy that makes names resolve ACROSS the
hybrid boundary without anyone running DNS server VMs. It anchors to
one virtual network and exposes up to two kinds of endpoints:

- **Inbound endpoints** give on-premises (or other-cloud) DNS
  forwarders a private IP inside the network to send queries TO --
  Azure answers them with the network's private DNS view (private
  zones, private endpoints, VM records).
- **Outbound endpoints** carry queries OUT of Azure toward
  on-premises DNS servers, steered by the forwarding rules of the
  DNS forwarding rulesets attached to them
  (AzurePrivateDnsResolverForwardingRuleset).

**Every endpoint needs its own dedicated subnet**, delegated to
"Microsoft.Network/dnsResolvers", sized /28 to /24, hosting nothing
else -- one endpoint per subnet, validated by ARM at deploy time
(the delegation lives on the referenced subnet, so it cannot be
checked here).

**Service limits worth planning around** (enforced by Azure, not
mirrored as validation because Microsoft adjusts them over time):
a virtual network hosts AT MOST ONE resolver; a resolver carries up
to 5 inbound and 5 outbound endpoints; each endpoint serves ~10,000
queries/second. Endpoints bill hourly from the moment they
provision; the resolver object itself is free.

Everything except tags is FIXED AT CREATION -- the resolver and its
endpoints have no ARM update surface beyond tags.

## Example

```yaml
# Offline-plan test manifest. Exercises the full surface: the resolver
# on its network, one STATIC inbound endpoint (pinned address -- the
# allocation contract's non-default arm), one outbound endpoint, and
# user tags merged over the derived ones. (The dynamic-inbound default
# arm is exercised by the live scenario, whose endpoint leaves the
# address to Azure.)
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateDnsResolver
metadata:
  name: test-private-dns-resolver
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: platform-rg
  name: test-dns-resolver
  virtualNetworkId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/virtualNetworks/platform-vnet
  inboundEndpoints:
    - name: inbound
      subnetId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/virtualNetworks/platform-vnet/subnets/dns-inbound
      privateIpAllocationMethod: STATIC
      privateIpAddress: "10.0.4.4"
  outboundEndpoints:
    - name: outbound
      subnetId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/virtualNetworks/platform-vnet/subnets/dns-outbound
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.virtualNetworkId` | `string \| valueFrom` | yes |  | AzureVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.inboundEndpoints` | `[]AzurePrivateDnsResolverInboundEndpoint` |  |  |  |
| `spec.inboundEndpoints[].name` | `string` | yes |  |  |
| `spec.inboundEndpoints[].subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.inboundEndpoints[].privateIpAllocationMethod` | `enum` |  |  |  |
| `spec.inboundEndpoints[].privateIpAddress` | `string` |  |  |  |
| `spec.outboundEndpoints` | `[]AzurePrivateDnsResolverOutboundEndpoint` |  |  |  |
| `spec.outboundEndpoints[].name` | `string` | yes |  |  |
| `spec.outboundEndpoints[].subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the resolver lives in, e.g. "eastus". MUST match
the region of the virtual network it anchors to (ARM rejects a
mismatch at deploy time). Composed endpoints deploy into the same
region. Changing the region replaces the resolver.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the resolver is created in. Can be a
literal resource-group name or a reference to an
AzureResourceGroup's name output. Changing it replaces the
resolver.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The resolver's name, unique within the resource group. Changing
the name replaces the resolver.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.virtualNetworkId

`string | valueFrom` · required

The virtual network the resolver anchors to -- references an
AzureVirtualNetwork's ARM id. All endpoint subnets must belong to
THIS network, and Azure allows only ONE resolver per virtual
network (a second create is rejected at deploy time). Changing
the network replaces the resolver.

- references: AzureVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

### spec.inboundEndpoints

`[]AzurePrivateDnsResolverInboundEndpoint`

Inbound endpoints -- the private IPs on-premises DNS forwarders
send queries TO. Most deployments carry exactly one; Azure allows
up to 5 per resolver (each on its own delegated subnet). The
FIRST endpoint declared here is the resolver's primary: its
resolved IP surfaces as the singular inbound_endpoint_ip output.
Endpoints are FIXED AT CREATION (adding or removing entries
creates or destroys endpoints; editing one replaces it).

- rule: STATIC allocation requires private_ip_address (an address from the endpoint subnet's range)
- rule: private_ip_address can only be set with STATIC allocation -- DYNAMIC lets Azure pick the address

### spec.inboundEndpoints[].name

`string` · required

The endpoint's name, unique on the resolver. Its ARM id composes
as {resolver_id}/inboundEndpoints/{name}.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.inboundEndpoints[].subnetId

`string | valueFrom` · required

The dedicated subnet the endpoint occupies -- references an
AzureSubnet's ARM id. The subnet must belong to the resolver's
virtual network, be delegated to "Microsoft.Network/dnsResolvers",
be sized /28 to /24, and host NOTHING else (one endpoint per
subnet; ARM validates all of this at deploy time because the
delegation lives on the referenced subnet).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.inboundEndpoints[].privateIpAllocationMethod

`enum`

How the endpoint's private IP is assigned. Unspecified applies
DYNAMIC (Azure picks a free address from the subnet), the
provider's default. STATIC pins the address you supply in
private_ip_address -- pick static when on-premises forwarder
configs must never change.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_private_dns_resolver_ip_allocation_method_unspecified` -- Not specified -- Azure assigns a free address from the subnet (DYNAMIC), the provider's default.
- `DYNAMIC` -- Azure picks a free address from the endpoint subnet's range.
- `STATIC` -- Pin the address supplied in private_ip_address -- forwarder configurations pointing at the endpoint never need to change.

### spec.inboundEndpoints[].privateIpAddress

`string`

The static private IP to pin, from the endpoint subnet's range.
REQUIRED with STATIC allocation; FORBIDDEN with DYNAMIC (the
provider rejects the combination before touching ARM -- mirrored
here so the failure is instant). The effective address -- static
or dynamically assigned -- surfaces in the outputs either way.

### spec.outboundEndpoints

`[]AzurePrivateDnsResolverOutboundEndpoint`

Outbound endpoints -- the egress points queries leave Azure
through, steered by attached forwarding rulesets. Most
deployments carry exactly one; Azure allows up to 5 per resolver
(each on its own delegated subnet). The FIRST endpoint declared
here is the resolver's primary: its ARM id surfaces as the
singular outbound_endpoint_id output that
AzurePrivateDnsResolverForwardingRuleset references by default.

### spec.outboundEndpoints[].name

`string` · required

The endpoint's name, unique on the resolver. Its ARM id composes
as {resolver_id}/outboundEndpoints/{name} -- the value forwarding
rulesets bind.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.outboundEndpoints[].subnetId

`string | valueFrom` · required

The dedicated subnet the endpoint occupies -- references an
AzureSubnet's ARM id. Same contract as the inbound endpoint's
subnet: the resolver's network, delegated to
"Microsoft.Network/dnsResolvers", /28 to /24, nothing else in it,
one endpoint per subnet (ARM validates at deploy time).

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.tags

`map<string, string>`

Free-form tags applied to the resolver AND its endpoints, merged
over the Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins. The only surface
updatable in place.

## Validation Rules

- `inbound_endpoint_names_unique`: Inbound endpoint names must be unique within the resolver
- `outbound_endpoint_names_unique`: Outbound endpoint names must be unique within the resolver

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzurePrivateDnsResolver, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.dns_resolver_id` | `string` | The resolver's ARM resource ID (.../providers/Microsoft.Network/dnsResolvers/{name}). |
| `status.outputs.dns_resolver_name` | `string` | The resolver's name within its resource group. |
| `status.outputs.inbound_endpoint_ip` | `string` | The private IP of the FIRST inbound endpoint declared in the spec -- the address on-premises DNS forwarders and VNet DNS settings point at. Empty when the spec declares no inbound endpoints. |
| `status.outputs.inbound_endpoint_ips` | `map<string, string>` | The private IPs of ALL inbound endpoints, keyed by endpoint name. Empty when the spec declares no inbound endpoints. |
| `status.outputs.outbound_endpoint_id` | `string` | The ARM id of the FIRST outbound endpoint declared in the spec ({resolver_id}/outboundEndpoints/{name}) -- what AzurePrivateDnsResolverForwardingRuleset references by default. Empty when the spec declares no outbound endpoints. |
| `status.outputs.outbound_endpoint_ids` | `map<string, string>` | The ARM ids of ALL outbound endpoints, keyed by endpoint name. Empty when the spec declares no outbound endpoints. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.virtualNetworkId` | AzureVirtualNetwork | `status.outputs.virtual_network_id` |
| `spec.inboundEndpoints[].subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.outboundEndpoints[].subnetId` | AzureSubnet | `status.outputs.subnet_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzurePrivateDnsResolverForwardingRuleset | `spec.outboundEndpointIds` | `status.outputs.outbound_endpoint_id` |

## See Also

- [Overview](../README.md)
