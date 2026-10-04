# AzurePrivateDnsResolverForwardingRuleset

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzurePrivateDnsResolverForwardingRulesetSpec** defines a DNS
forwarding ruleset -- the rule book that decides which DNS names
leave Azure and which servers answer them. Each rule pairs a domain
(say "corp.contoso.com.") with the on-premises (or other custom)
DNS servers that own it; queries for everything else resolve
normally inside Azure.

A ruleset attaches to the OUTBOUND endpoints of a DNS Private
Resolver (AzurePrivateDnsResolver) -- that is the pipe its rules
steer -- and takes effect in a virtual network only once the
network is linked to it
(AzurePrivateDnsResolverVirtualNetworkLink). Linked networks do NOT
need to be peered with the resolver's network: in hub-and-spoke,
spokes link to the hub's ruleset and their queries egress through
the hub's outbound endpoint.

**Service limits worth planning around** (enforced by Azure, not
mirrored as validation because Microsoft adjusts them over time):
a ruleset binds AT MOST 2 outbound endpoints, and both must belong
to the SAME resolver; a ruleset carries up to 1,000 rules and up to
500 virtual-network links, all in the ruleset's own region.
Rulesets, rules, and links are free at rest -- only endpoint hours
and query volume bill.

## Example

```yaml
# Offline-plan test manifest. Exercises the full surface: two rules
# (one live with two ordered targets and an explicit non-standard
# port, one PARKED with enabled false and metadata annotations), the
# endpoint binding, and user tags merged over the derived ones.
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePrivateDnsResolverForwardingRuleset
metadata:
  name: test-forwarding-ruleset
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: platform-rg
  name: test-ruleset
  outboundEndpointIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/dnsResolvers/platform-resolver/outboundEndpoints/outbound
  forwardingRules:
    - name: corp-domain
      domainName: corp.contoso.com.
      targetDnsServers:
        - ipAddress: "10.100.0.10"
        - ipAddress: "10.100.0.11"
          port: 5353
      metadata:
        owner: network-team
    - name: parked-domain
      domainName: internal.fabrikam.com.
      enabled: false
      targetDnsServers:
        - ipAddress: "172.16.5.53"
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.outboundEndpointIds` | `[]string \| valueFrom` | yes |  | AzurePrivateDnsResolver (`status.outputs.outbound_endpoint_id`) |
| `spec.forwardingRules` | `[]AzurePrivateDnsResolverForwardingRule` |  |  |  |
| `spec.forwardingRules[].name` | `string` | yes |  |  |
| `spec.forwardingRules[].domainName` | `string` | yes |  |  |
| `spec.forwardingRules[].targetDnsServers` | `[]AzurePrivateDnsResolverForwardingRuleTargetDnsServer` | yes |  |  |
| `spec.forwardingRules[].targetDnsServers[].ipAddress` | `string` | yes |  |  |
| `spec.forwardingRules[].targetDnsServers[].port` | `int32` |  | `53` |  |
| `spec.forwardingRules[].enabled` | `bool` |  | `true` |  |
| `spec.forwardingRules[].metadata` | `map<string, string>` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the ruleset lives in, e.g. "eastus". Must match
the region of the resolver whose outbound endpoints it binds, and
only virtual networks in this region can link to it. Changing the
region replaces the ruleset.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the ruleset is created in. Can be a
literal resource-group name or a reference to an
AzureResourceGroup's name output. Changing it replaces the
ruleset.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The ruleset's name, unique within the resource group. Changing
the name replaces the ruleset.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.outboundEndpointIds

`[]string | valueFrom` · required

The outbound endpoint(s) this ruleset steers, by ARM id --
defaults to referencing an AzurePrivateDnsResolver's primary
outbound_endpoint_id output (the resolver publishes every
endpoint's id in its outbound_endpoint_ids map; supply additional
ids as literals or explicit references). Azure accepts at most 2,
both from the SAME resolver. Updatable in place.

- references: AzurePrivateDnsResolver (`status.outputs.outbound_endpoint_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePrivateDnsResolver, name: <that resource's name>, fieldPath: status.outputs.outbound_endpoint_id}} -- a bare string does not parse

### spec.forwardingRules

`[]AzurePrivateDnsResolverForwardingRule`

The forwarding rules -- one per domain whose queries leave Azure.
A ruleset with no rules forwards nothing (legal, and a fine way
to stage the plumbing before the rules). Azure allows up to
1,000. Rules are ADDED, REMOVED, and EDITED in place -- except a
rule's domain_name, which replaces that rule.

### spec.forwardingRules[].name

`string` · required

The rule's name, unique on the ruleset. Its ARM id composes as
{ruleset_id}/forwardingRules/{name}.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.forwardingRules[].domainName

`string` · required

The DNS suffix this rule captures -- the rule matches this domain
and every name under it. ARM stores it as a fully qualified name
WITH the trailing dot ("corp.contoso.com."); write it that way.
The ONLY rule field that cannot change in place: editing the
domain replaces the rule (a brief forwarding gap for that domain,
nothing else).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.forwardingRules[].targetDnsServers

`[]AzurePrivateDnsResolverForwardingRuleTargetDnsServer` · required

The DNS servers that own the domain -- typically the on-premises
domain controllers or resolvers. Azure tries them in order;
supply at least one (Azure caps a rule at 6). Updatable in place.

- rule: {"repeated":{"minItems":"1"}}

### spec.forwardingRules[].targetDnsServers[].ipAddress

`string` · required

The server's IPv4 address, reachable from the outbound endpoint's
subnet (over VPN/ExpressRoute for on-premises targets).

- rule: {"required":true,"string":{"ipv4":true}}

### spec.forwardingRules[].targetDnsServers[].port

`int32` · optional (explicit presence)

The server's DNS port. Unspecified applies 53, the standard DNS
port and ARM's default -- the modules always send the effective
value explicitly.

- default: `53`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.forwardingRules[].enabled

`bool` · optional (explicit presence)

Whether the rule is active. Unspecified applies true, the
provider's default -- set false to park a rule without deleting
it (its domain resolves inside Azure again). Updatable in place.

- default: `true`

### spec.forwardingRules[].metadata

`map<string, string>`

Free-form key/value annotations stored on the rule itself (ARM's
metadata map -- rules carry no tags). Useful for recording the
rule's owner or ticket. Updatable in place.

### spec.tags

`map<string, string>`

Free-form tags applied to the ruleset, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins. (Individual
rules carry no tags -- ARM gives them a metadata map instead.)

## Validation Rules

- `forwarding_rule_names_unique`: Forwarding rule names must be unique within the ruleset

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzurePrivateDnsResolverForwardingRuleset, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.dns_forwarding_ruleset_id` | `string` | The ruleset's ARM resource ID (.../providers/Microsoft.Network/dnsForwardingRulesets/{name}) -- what AzurePrivateDnsResolverVirtualNetworkLink references. |
| `status.outputs.dns_forwarding_ruleset_name` | `string` | The ruleset's name within its resource group. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.outboundEndpointIds` | AzurePrivateDnsResolver | `status.outputs.outbound_endpoint_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzurePrivateDnsResolverVirtualNetworkLink | `spec.dnsForwardingRulesetId` | `status.outputs.dns_forwarding_ruleset_id` |

## See Also

- [Overview](../README.md)
