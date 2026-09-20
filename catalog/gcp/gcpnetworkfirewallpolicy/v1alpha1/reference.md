# GcpNetworkFirewallPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpNetworkFirewallPolicySpec creates one network firewall policy: an
ordered set of firewall rules owned by a project and associated with
one or more of its VPC networks. It is Google's replacement for legacy
per-network firewall rules (GcpFirewallRule) -- one policy holds the
whole rule set, can be attached to several networks, and carries the
matching surface the legacy rules lack: secure tags, address groups,
domain names, threat-intelligence lists, geography, and layer-7
inspection through a security profile group.

One kind covers both scopes. Leave `region` empty for a GLOBAL policy,
which governs all of a network's traffic in every region and is the
right default. Set `region` for a REGIONAL policy, which governs only
that region's traffic on the networks it is associated with -- the
scope an internal managed load balancer's rules must live at
(target_type INTERNAL_MANAGED_LB) and the only scope that accepts the
RDMA and ultra-low-latency policy types. Google models the two scopes as
separate resource families with identical rule and association shapes,
so the spec is one shape and the module picks the family by `region`.

The kind composes three Google resources: the policy, its rules (one
resource per entry of `rules`, keyed by priority), and its associations
(one per entry of `associations`: a VPC network in the same project).
A network can carry one global and one regional (per region) network
firewall policy association at a time.

Evaluation order Google applies to a packet: the hierarchical policies
above the project first (GcpHierarchicalFirewallPolicy), then the
network's GLOBAL network firewall policy, then its REGIONAL one, then
legacy VPC firewall rules -- unless the network's firewall policy
enforcement order has been flipped to put legacy rules before network
policies. A rule's `goto_next` hands the decision to the next level;
`allow` and `deny` end it.

What is deliberately NOT here: organization- and folder-scoped policies
(GcpHierarchicalFirewallPolicy); address groups and security profile
groups, named as literals until the catalog carries them.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpNetworkFirewallPolicy
metadata:
  name: planton-oss-e2e-gcpnfwp
  id: planton-oss-e2e-gcpnfwp
  org: planton-oss
  env: e2e
  labels:
    managed-by: planton-e2e
    e2e-component: gcpnetworkfirewallpolicy
  annotations:
    planton.dev/e2e: "true"
  tags:
    - planton-e2e
spec:
  # The policy's name in GCP; defaults to metadata.name. Project-scoped with
  # no soft-delete reservation, so it stays fixed.
  policyName: planton-oss-e2e-gcpnfwp

  # Empty region = a GLOBAL policy governing every region of the attached
  # network (the everyday choice). See the `regional` scenario for the other
  # resource family.
  region: ""

  description: Planton E2E network firewall policy; safe to delete

  # Evaluated lowest priority first. Each rule is its own resource keyed by
  # priority.
  rules:
    # Identity-Aware Proxy TCP forwarding may reach SSH.
    - priority: 1000
      action: allow
      direction: INGRESS
      ruleName: allow-iap-ssh
      description: IAP TCP forwarding to SSH
      match:
        srcIpRanges: ["35.235.240.0/20"]
        layer4Configs:
          - ipProtocol: tcp
            ports: ["22"]

    # Everything else is dropped and logged.
    - priority: 65000
      action: deny
      direction: INGRESS
      ruleName: deny-all-ingress
      description: Default deny, logged
      enableLogging: true
      match:
        srcIpRanges: ["0.0.0.0/0"]
        layer4Configs:
          - ipProtocol: all

  # Attached to the registry's prerequisite network (a custom-mode VPC with
  # no subnets, so the deny reaches nothing that exists).
  associations:
    - name: planton-oss-e2e-gcpnfwp-prereq
      network:
        valueFrom:
          kind: GcpVpcNetwork
          name: planton-oss-e2e-gcpvpcnetwork-prereq
          fieldPath: status.outputs.network_self_link

  # DELETE (default), PREVENT, ABANDON -- fans to the policy, every rule,
  # and every association.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.policyName` | `string` |  |  |  |
| `spec.region` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.policyType` | `string` |  |  |  |
| `spec.rules` | `[]GcpNetworkFirewallPolicyRule` |  |  |  |
| `spec.rules[].priority` | `int32` | yes |  |  |
| `spec.rules[].action` | `string` | yes |  |  |
| `spec.rules[].direction` | `string` | yes |  |  |
| `spec.rules[].description` | `string` |  |  |  |
| `spec.rules[].ruleName` | `string` |  |  |  |
| `spec.rules[].disabled` | `bool` |  |  |  |
| `spec.rules[].enableLogging` | `bool` |  |  |  |
| `spec.rules[].match` | `GcpNetworkFirewallPolicyRuleMatch` | yes |  |  |
| `spec.rules[].match.layer4Configs` | `[]GcpNetworkFirewallPolicyLayer4Config` | yes |  |  |
| `spec.rules[].match.layer4Configs[].ipProtocol` | `string` | yes |  |  |
| `spec.rules[].match.layer4Configs[].ports` | `[]string` |  |  |  |
| `spec.rules[].match.srcIpRanges` | `[]string` |  |  |  |
| `spec.rules[].match.destIpRanges` | `[]string` |  |  |  |
| `spec.rules[].match.srcAddressGroups` | `[]string` |  |  |  |
| `spec.rules[].match.destAddressGroups` | `[]string` |  |  |  |
| `spec.rules[].match.srcFqdns` | `[]string` |  |  |  |
| `spec.rules[].match.destFqdns` | `[]string` |  |  |  |
| `spec.rules[].match.srcRegionCodes` | `[]string` |  |  |  |
| `spec.rules[].match.destRegionCodes` | `[]string` |  |  |  |
| `spec.rules[].match.srcThreatIntelligences` | `[]string` |  |  |  |
| `spec.rules[].match.destThreatIntelligences` | `[]string` |  |  |  |
| `spec.rules[].match.srcSecureTags` | `[]string \| valueFrom` |  |  | GcpTagValue (`status.outputs.name`) |
| `spec.rules[].match.srcNetworks` | `[]string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.rules[].match.srcNetworkContext` | `string` |  |  |  |
| `spec.rules[].match.destNetworkContext` | `string` |  |  |  |
| `spec.rules[].targetType` | `string` |  |  |  |
| `spec.rules[].targetForwardingRules` | `[]string \| valueFrom` |  |  | GcpGlobalForwardingRule (`status.outputs.self_link`) |
| `spec.rules[].targetSecureTags` | `[]string \| valueFrom` |  |  | GcpTagValue (`status.outputs.name`) |
| `spec.rules[].targetServiceAccounts` | `[]string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.rules[].securityProfileGroup` | `string` |  |  |  |
| `spec.rules[].tlsInspect` | `bool` |  |  |  |
| `spec.associations` | `[]GcpNetworkFirewallPolicyAssociation` |  |  |  |
| `spec.associations[].name` | `string` |  |  |  |
| `spec.associations[].network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the policy and its rules are created in, and whose
networks it can be associated with: a reference to a GcpProject or
the project ID as a literal. Empty means the provider's default
project. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.policyName

`string`

The policy's name in GCP, unique among the project's network firewall
policies of the same scope. Defaults to metadata.name when empty. RFC
1035: 1-63 characters, lowercase letters, digits, and hyphens,
starting with a letter and not ending with a hyphen. Immutable: a
rename recreates the policy with every rule and association.

- rule: policy_name must be 1-63 characters: lowercase letters, digits, and hyphens, starting with a letter and not ending with a hyphen (RFC 1035)

### spec.region

`string`

Empty for a GLOBAL policy (all regions of the associated networks; the
right default). A region name such as us-central1 for a REGIONAL
policy, which governs only that region's traffic -- pick regional when
the rules target an internal managed load balancer (INTERNAL_MANAGED_LB
rules are regional) or need an RDMA or ULL policy type. Immutable: a
policy cannot move between scopes or regions.

- rule: region must be a valid GCP region name such as us-central1, or empty for a global policy

### spec.description

`string`

What this policy enforces and who owns it -- the operator reading a
denied connection's log sees the policy, not this manifest. Mutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.policyType

`string` · optional (explicit presence)

Which kind of network the policy can be associated with; a network
accepts a policy only when its network profile carries the matching
type. VPC_POLICY is the ordinary VPC network and Google's default when
unset -- leave this empty unless the network is one of the special
profiles. Regional policies only: RDMA_ROCE_POLICY (RoCE RDMA
networks), RDMA_FALCON_POLICY (Falcon RDMA networks), ULL_POLICY
(ultra-low-latency networks). Immutable. Both engines send it only
when set, so an unset value never fights Google's default on re-plan.

- rule: policy_type must be one of: VPC_POLICY, RDMA_ROCE_POLICY, RDMA_FALCON_POLICY, ULL_POLICY

### spec.rules

`[]GcpNetworkFirewallPolicyRule`

The ordered rule set, evaluated from the LOWEST priority number up:
the first rule whose match fits the packet decides (allow, deny, or
goto_next, which delegates to the next level). Each entry is its own
Google resource keyed by `priority`, so changing a rule's priority
RECREATES that rule (the old priority is freed, the new one is
created) while every other rule is left alone; change a rule's content
in place and it is updated in place. An empty list is a valid policy
that decides nothing.

- rule: security_profile_group is required when action is apply_security_profile_group and must be empty for every other action
- rule: tls_inspect can be true only when action is apply_security_profile_group
- rule: enable_logging cannot be set on a goto_next rule -- Google logs only the rule that decides
- rule: target_forwarding_rules is required when target_type is INTERNAL_MANAGED_LB and must be empty otherwise

### spec.rules[].priority

`int32` · required · optional (explicit presence)

The rule's position in the evaluation order: 0 is evaluated first,
2147483647 last. Unique within the policy, and a rule's identity in
Google: changing it recreates the rule. Leave a gap between rules
(1000, 2000, ...) so a rule can be slotted in later without
renumbering.

- rule: {"required":true,"int32":{"lte":2147483647,"gte":0}}

### spec.rules[].action

`string` · required

What happens to a matching packet:
  "allow"     -- permit it; evaluation of every lower level stops
  "deny"      -- drop it; evaluation stops
  "goto_next" -- delegate the decision to the next level (the
                 regional policy after a global one, then legacy VPC
                 rules)
  "apply_security_profile_group" -- send the traffic through the
                 Cloud NGFW security profile group named in
                 security_profile_group for layer-7 inspection, which
                 then allows or denies it
Lowercase, exactly as Google's API spells them.

- rule: action must be one of: allow, deny, goto_next, apply_security_profile_group
- rule: {"required":true}

### spec.rules[].direction

`string` · required

Which traffic the rule looks at: INGRESS (arriving at the targets;
`match` describes the SOURCE with src_* fields and may name the
destination) or EGRESS (leaving the targets; `match` describes the
DESTINATION with dest_* fields and may name the source). An ingress
rule's src_secure_tags and src_networks describe where traffic comes
from, so they carry no meaning on an egress rule and Google ignores
them there.

- rule: direction must be INGRESS or EGRESS
- rule: {"required":true}

### spec.rules[].description

`string`

What this rule is for, in the words an operator reading a firewall
log will understand. Mutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.rules[].ruleName

`string`

A label for the rule shown by `gcloud` and the console beside its
priority. Mutable, and NOT the rule's identity (priority is), so it
can be renamed freely. Optional.

- rule: {"string":{"maxLen":"63"}}

### spec.rules[].disabled

`bool`

When true the rule is kept but not enforced -- traffic behaves as if
the rule did not exist. The safe way to switch a rule off without
losing it (and its priority slot). Mutable.

### spec.rules[].enableLogging

`bool`

Log every connection this rule decides, to Cloud Logging (and from
there to BigQuery or Pub/Sub through a GcpLoggingSink). Logs carry the
rule, the policy, both endpoints, and the decision. Cannot be set on a
goto_next rule: Google logs only the rule that decides. Mutable.

### spec.rules[].match

`GcpNetworkFirewallPolicyRuleMatch` · required

The condition a packet must satisfy for this rule to decide it. At
least one layer4_configs entry is required; every other field narrows
the match and an empty field matches everything.

- rule: {"required":true}

### spec.rules[].match.layer4Configs

`[]GcpNetworkFirewallPolicyLayer4Config` · required

Protocols and ports the rule applies to. At least one entry; a packet
matches when it fits any entry.

- rule: {"repeated":{"minItems":"1"}}
- rule: ports can be set only when ip_protocol is tcp, udp, 6, or 17 -- other protocols have no ports

### spec.rules[].match.layer4Configs[].ipProtocol

`string` · required

The IP protocol: a well-known name (tcp, udp, icmp, esp, ah, ipip,
sctp), `all` for every protocol, or a protocol number (0-255).

- rule: ip_protocol must be tcp, udp, icmp, esp, ah, ipip, sctp, all, or an IP protocol number 0-255
- rule: {"required":true}

### spec.rules[].match.layer4Configs[].ports

`[]string`

Ports or port ranges the rule applies to, for tcp and udp only:
"22", "80", "8000-8999". Empty means every port of the protocol.

- rule: {"repeated":{"items":{"string":{"pattern":"^[0-9]{1,5}(-[0-9]{1,5})?$"}}}}

### spec.rules[].match.srcIpRanges

`[]string`

Source addresses in CIDR form (IPv4 or IPv6, e.g. 10.0.0.0/8,
0.0.0.0/0, 2001:db8::/32). Maximum 5000. On an INGRESS rule the
everyday selector ("from the internet", "from on-premises").

- rule: {"repeated":{"maxItems":"5000","items":{"string":{"ipWithPrefixlen":true}}}}

### spec.rules[].match.destIpRanges

`[]string`

Destination addresses in CIDR form. Maximum 5000. On an EGRESS rule
the everyday selector ("to the internet", "to the database subnet").

- rule: {"repeated":{"maxItems":"5000","items":{"string":{"ipWithPrefixlen":true}}}}

### spec.rules[].match.srcAddressGroups

`[]string`

Address groups whose addresses count as the source, as Network
Security address-group names
(`projects/{project}/locations/global/addressGroups/{name}` or
`organizations/{id}/locations/global/addressGroups/{name}`). Maximum
10. Literals: address groups are not yet a catalog kind.

- rule: {"repeated":{"maxItems":"10"}}

### spec.rules[].match.destAddressGroups

`[]string`

Address groups whose addresses count as the destination. Maximum 10.

- rule: {"repeated":{"maxItems":"10"}}

### spec.rules[].match.srcFqdns

`[]string`

Fully qualified domain names whose resolved addresses count as the
source (Google resolves them continuously). Maximum 100.

- rule: {"repeated":{"maxItems":"100"}}

### spec.rules[].match.destFqdns

`[]string`

Fully qualified domain names whose resolved addresses count as the
destination -- the way an egress rule allows `api.example.com` without
pinning its IPs. Maximum 100.

- rule: {"repeated":{"maxItems":"100"}}

### spec.rules[].match.srcRegionCodes

`[]string`

Countries whose IP space counts as the source, as ISO 3166-1 alpha-2
codes (US, DE, ...). Maximum 5000. Geo-blocking on ingress.

- rule: {"repeated":{"maxItems":"5000","items":{"string":{"pattern":"^[A-Z]{2}$"}}}}

### spec.rules[].match.destRegionCodes

`[]string`

Countries whose IP space counts as the destination, as ISO 3166-1
alpha-2 codes. Maximum 5000.

- rule: {"repeated":{"maxItems":"5000","items":{"string":{"pattern":"^[A-Z]{2}$"}}}}

### spec.rules[].match.srcThreatIntelligences

`[]string`

Google Threat Intelligence lists whose addresses count as the source,
by list name: iplist-tor-exit-nodes, iplist-known-malicious-ips,
iplist-search-engines-crawlers, iplist-vpn-providers,
iplist-anon-proxies, iplist-crypto-miners, iplist-public-clouds and
the per-cloud lists (iplist-public-clouds-aws, -azure, -gcp). A deny
rule on iplist-known-malicious-ips is the classic use.

### spec.rules[].match.destThreatIntelligences

`[]string`

Google Threat Intelligence lists whose addresses count as the
destination, by list name (the same names as src_threat_intelligences).

### spec.rules[].match.srcSecureTags

`[]string | valueFrom`

Source VMs carrying one of these secure tags -- references to
GcpTagValue resources (their `name` output, `tagValues/{numeric_id}`)
or those names as literals. Maximum 256. INGRESS rules only: on an
ingress rule with src_secure_tags and no src_ip_ranges, if every tag
is INEFFECTIVE the rule is ignored. Micro-segmentation without IP
ranges: "frontend-tagged VMs may reach backend-tagged VMs on 8080".

- references: GcpTagValue (`status.outputs.name`)
- rule: {"repeated":{"maxItems":"256"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpTagValue, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.rules[].match.srcNetworks

`[]string | valueFrom`

Source VPC networks -- references to GcpVpcNetwork resources (their
network_self_link output) or network URLs as literals. INGRESS rules
only: matches traffic that originates inside one of these networks
(peered or shared), the selector behind src_network_context
VPC_NETWORKS.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.rules[].match.srcNetworkContext

`string` · optional (explicit presence)

Where the traffic comes from, as a class rather than an address:
INTERNET (outside Google Cloud), INTRA_VPC (inside this network),
NON_INTERNET (any non-internet source), VPC_NETWORKS (the networks in
src_networks), or UNSPECIFIED. Unset means the rule does not narrow on
context (Google's default); both engines send it only when set.

- rule: src_network_context must be one of: UNSPECIFIED, INTERNET, INTRA_VPC, NON_INTERNET, VPC_NETWORKS

### spec.rules[].match.destNetworkContext

`string` · optional (explicit presence)

Where the traffic is going, as a class: INTERNET, INTRA_VPC,
NON_INTERNET, VPC_NETWORKS, or UNSPECIFIED. Unset means no narrowing
(Google's default); sent only when set. The egress-side twin of
src_network_context -- "to the internet" without listing 0.0.0.0/0.

- rule: dest_network_context must be one of: UNSPECIFIED, INTERNET, INTRA_VPC, NON_INTERNET, VPC_NETWORKS

### spec.rules[].targetType

`string` · optional (explicit presence)

What the rule applies to: INSTANCES (VMs on the associated networks;
Google's default when unset) or INTERNAL_MANAGED_LB (the internal
Application Load Balancers named in target_forwarding_rules -- a
regional policy's feature). Both engines send it only when set, so an
unset value never fights Google's default on re-plan.

- rule: target_type must be INSTANCES or INTERNAL_MANAGED_LB

### spec.rules[].targetForwardingRules

`[]string | valueFrom`

The internal managed load balancers the rule applies to when
target_type is INTERNAL_MANAGED_LB -- references to forwarding-rule
resources (a GcpGlobalForwardingRule's self_link output) or
forwarding-rule self-links as literals. Required for that target
type, forbidden for INSTANCES. An internal managed load balancer's
forwarding rule is REGIONAL: reference a GcpGlobalForwardingRule
declared with a region (the kind builds the regional rule behind its
region selector; its name records the global rule it began as).

- references: GcpGlobalForwardingRule (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGlobalForwardingRule, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.rules[].targetSecureTags

`[]string | valueFrom`

Restrict the rule to VMs carrying one of these secure tags --
references to GcpTagValue resources (their `name` output,
`tagValues/{numeric_id}`) or those names as literals. Maximum 256. A
secure tag is a Resource Manager tag value bound to the VM; a tag
whose value or network was deleted is INEFFECTIVE, and a rule whose
target tags are all ineffective is ignored. Empty (with
target_service_accounts also empty) means every VM on the associated
networks.

- references: GcpTagValue (`status.outputs.name`)
- rule: {"repeated":{"maxItems":"256"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpTagValue, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.rules[].targetServiceAccounts

`[]string | valueFrom`

Restrict the rule to VMs running as one of these service accounts --
references to GcpServiceAccount resources (their email output) or
email addresses as literals.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.rules[].securityProfileGroup

`string`

The Cloud NGFW security profile group that inspects the traffic when
action is apply_security_profile_group, as its full resource URL:
`//networksecurity.googleapis.com/projects/{project}/locations/global/securityProfileGroups/{name}`
(organization-owned groups use `organizations/{id}` in place of
`projects/{project}`). Required for that action, forbidden for every
other. A literal: security profile groups are not yet a catalog kind.

### spec.rules[].tlsInspect

`bool`

Decrypt TLS traffic before the security profile group inspects it
(Cloud NGFW TLS inspection; needs a TLS inspection policy on the
network). Only with action apply_security_profile_group.

### spec.associations

`[]GcpNetworkFirewallPolicyAssociation`

The VPC networks this policy is enforced on -- each entry is one
association resource attaching the policy to one network in the
same project. A network carries at most one global and, per region,
one regional network firewall policy at a time; associating a second
fails until the first is detached. A policy with no association
exists but governs nothing.

### spec.associations[].name

`string`

The association's name, unique within the policy. Defaults to
`<policy_name>-<n>` (n = the entry's position, starting at 1) when
empty.

- rule: name must be 1-63 characters: lowercase letters, digits, and hyphens, starting with a letter and not ending with a hyphen (RFC 1035)

### spec.associations[].network

`string | valueFrom` · required

The VPC network the policy is enforced on: a reference to a
GcpVpcNetwork resource (its network_self_link output) or the network's
self-link as a literal. Must be in the policy's project. An attachment
edge, not a placement: the policy does not live in the network.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What destroying this resource does in GCP, applied to the policy, to
every rule, and to every association together:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- associations are detached, rules removed, and the
               policy deleted; traffic falls through to the next
               level as if the policy had never existed
  "PREVENT" -- destroy FAILS; the guard for a network's baseline deny
  "ABANDON" -- everything is removed from management but keeps
               existing and enforcing in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `rule_priorities_unique`: every rule must have a distinct priority -- Google keys a policy's rules by priority
- `association_names_unique`: every association must have a distinct name within the policy
- `policy_type_matches_scope`: policy_type RDMA_ROCE_POLICY, RDMA_FALCON_POLICY, and ULL_POLICY exist only for a regional policy (region set); a global policy accepts VPC_POLICY only

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpNetworkFirewallPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_name` | `string` | The policy's name in GCP, as declared (or defaulted from metadata.name) -- what `gcloud compute network-firewall-policies describe` takes. |
| `status.outputs.policy_id` | `string` | The policy's server-assigned numeric ID. |
| `status.outputs.self_link` | `string` | The policy's self-link URL: `.../global/firewallPolicies/{name}` for a global policy, `.../regions/{region}/firewallPolicies/{name}` for a regional one. |
| `status.outputs.region` | `string` | The region the policy lives in, or empty for a global policy -- the value that tells a consumer which resource family the policy is. |
| `status.outputs.rule_tuple_count` | `int64` | Google's complexity measure for the whole rule set -- the sum of each rule's tuple count (protocols x ports x sources x targets). A project has a quota of tuples per policy; watch this before adding wide rules. |
| `status.outputs.association_names` | `[]string` | The association names, in the order declared -- what `gcloud compute network-firewall-policies associations list` shows and what a detach targets. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.rules[].match.srcSecureTags` | GcpTagValue | `status.outputs.name` |
| `spec.rules[].match.srcNetworks` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.rules[].targetForwardingRules` | GcpGlobalForwardingRule | `status.outputs.self_link` |
| `spec.rules[].targetSecureTags` | GcpTagValue | `status.outputs.name` |
| `spec.rules[].targetServiceAccounts` | GcpServiceAccount | `status.outputs.email` |
| `spec.associations[].network` | GcpVpcNetwork | `status.outputs.network_self_link` |

## See Also

- [Overview](../README.md)
