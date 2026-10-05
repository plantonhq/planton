# AzurePrivateDnsResolverVirtualNetworkLink

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzurePrivateDnsResolverVirtualNetworkLinkSpec** defines the
attachment that makes a DNS forwarding ruleset
(AzurePrivateDnsResolverForwardingRuleset) take effect in one
virtual network: once linked, resources in that network resolve the
ruleset's domains through the resolver's outbound endpoints. A
ruleset without links steers nobody; each link adds one network to
its audience.

The link is a first-class resource because it is many-per-ruleset
with its own lifecycle: hub-and-spoke topologies link one ruleset
(owned with the hub's resolver) to hundreds of spoke networks, and
spokes join and leave without touching the ruleset or each other --
linked networks do not even need to be peered with the resolver's
network, or live in the same subscription. Azure allows up to 500
links per ruleset, all in the ruleset's region (enforced by Azure,
not mirrored as validation because Microsoft adjusts limits over
time). Compose it with the ruleset and AzureVirtualNetwork; one
link resource per ruleset-network pair.

Links are free at rest. Everything except metadata is FIXED AT
CREATION.

## Example

```yaml
# Offline-plan test manifest. Exercises the full surface: the link
# with its metadata annotations (links carry no tags -- ARM gives them
# a metadata map instead).
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateDnsResolverVirtualNetworkLink
metadata:
  name: test-resolver-vnet-link
  org: test-org
  env: dev
spec:
  name: spoke-payments
  dnsForwardingRulesetId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/dnsForwardingRulesets/platform-ruleset
  virtualNetworkId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/spoke-rg/providers/Microsoft.Network/virtualNetworks/spoke-vnet
  metadata:
    owner: payments-team
    env: prod
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.dnsForwardingRulesetId` | `string \| valueFrom` | yes |  | AzurePrivateDnsResolverForwardingRuleset (`status.outputs.dns_forwarding_ruleset_id`) |
| `spec.virtualNetworkId` | `string \| valueFrom` | yes |  | AzureVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.metadata` | `map<string, string>` |  |  |  |

## Field Details

### spec.name

`string` · required

The link's name -- its ARM resource name under the parent
ruleset, unique per ruleset. Name it after the network it
attaches ("hub-vnet", "spoke-payments"). Changing the name
replaces the link (a brief forwarding gap for the affected
network, nothing else).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dnsForwardingRulesetId

`string | valueFrom` · required

The DNS forwarding ruleset this link attaches to. Takes the
ruleset's full ARM resource ID; defaults to referencing an
AzurePrivateDnsResolverForwardingRuleset's
dns_forwarding_ruleset_id output in composed environments. The
link is a child resource of the ruleset -- its name and resource
group are derived from this ID, so they can never contradict it.
Changing the ruleset replaces the link.

- references: AzurePrivateDnsResolverForwardingRuleset (`status.outputs.dns_forwarding_ruleset_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePrivateDnsResolverForwardingRuleset, name: <that resource's name>, fieldPath: status.outputs.dns_forwarding_ruleset_id}} -- a bare string does not parse

### spec.virtualNetworkId

`string | valueFrom` · required

The virtual network the ruleset becomes active in. Takes the
network's full ARM resource ID; defaults to referencing an
AzureVirtualNetwork's virtual_network_id output in composed
environments. The network must be in the ruleset's region; it
does NOT need to be peered with the resolver's network (cross-
subscription links are allowed). Changing the network replaces
the link.

- references: AzureVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

### spec.metadata

`map<string, string>`

Free-form key/value annotations stored on the link itself (ARM's
metadata map -- links carry no tags). Useful for recording the
linked network's owner or environment. The only surface updatable
in place.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzurePrivateDnsResolverVirtualNetworkLink, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.virtual_network_link_id` | `string` | The link's ARM resource ID. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/dnsForwardingRulesets/{ruleset}/virtualNetworkLinks/{name} |
| `status.outputs.virtual_network_link_name` | `string` | The link's name. Echoed from the spec for convenience. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.dnsForwardingRulesetId` | AzurePrivateDnsResolverForwardingRuleset | `status.outputs.dns_forwarding_ruleset_id` |
| `spec.virtualNetworkId` | AzureVirtualNetwork | `status.outputs.virtual_network_id` |

## See Also

- [Overview](../README.md)
