# GcpHierarchicalFirewallPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpHierarchicalFirewallPolicySpec creates one hierarchical firewall
policy: an ordered set of firewall rules that lives on the organization
or on a folder and is enforced on every VPC network beneath the node it
is associated with -- in every project, present and future -- before any
network-level policy or legacy VPC firewall rule gets a say. It is the
tool a platform team uses to state the rules nobody in the organization
may undo ("no SSH from the internet", "deny known-malicious sources",
"allow the health-check ranges") and to delegate everything else
downward with goto_next.

The kind composes three Google resources: the policy (a container that
Google names with a server-assigned numeric ID), its rules (one resource
per entry of `rules`, keyed by priority), and its associations (one
resource per entry of `associations`: the organization or folder the
policy is enforced on). A policy with no association exists but governs
nothing; the everyday shape is one policy associated with the node it
lives under, and Google allows the same policy to be associated with
several folders so one rule set can govern many branches of the tree.
A folder or the organization can carry at most one hierarchical policy
association at a time.

Evaluation order Google applies to a packet: hierarchical policies from
the organization down through the folders (lowest priority number first
within each), then the network's global network firewall policy, then
the regional one, then legacy VPC firewall rules. A rule's `goto_next`
hands the decision to the next level; `allow` and `deny` end it. Google
appends two implied goto_next rules at priorities 2147483646 (IPv4) and
2147483647 (IPv6) to every hierarchical policy, so leave those two
numbers alone.

What is deliberately NOT here: network-scoped policies attached to one
VPC (GcpNetworkFirewallPolicy) and the legacy per-network rules
(GcpFirewallRule); address groups and security profile groups, which
are their own Google resources and are named here as literals until the
catalog carries them.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpHierarchicalFirewallPolicy
metadata:
  name: planton-oss-e2e-gcphfwp
  id: planton-oss-e2e-gcphfwp
  org: planton-oss
  env: e2e
  labels:
    managed-by: planton-e2e
    e2e-component: gcphierarchicalfirewallpolicy
  annotations:
    planton.dev/e2e: "true"
  tags:
    - planton-e2e
spec:
  # Where the policy lives: exactly one of organizationId or folderId (a
  # GcpFolder reference or numeric literal). Immutable. The reference
  # manifest names a placeholder organization (the kind is organization-
  # scoped and its profile is deferred); the scenarios take the real one
  # from the harness token.
  parent:
    organizationId: "123456789012"

  # Unique in the organization; defaults to metadata.name.
  shortName: planton-oss-e2e-gcphfwp

  description: Planton E2E hierarchical firewall policy; safe to delete

  # Evaluated lowest priority first. Each rule is its own resource keyed by
  # priority. A logged deny nobody beneath can undo, then delegation.
  rules:
    - priority: 1000
      action: deny
      direction: INGRESS
      description: No SSH from the internet, organization-wide
      enableLogging: true
      match:
        srcIpRanges: ["0.0.0.0/0"]
        layer4Configs:
          - ipProtocol: tcp
            ports: ["22"]
    - priority: 2000
      action: goto_next
      direction: INGRESS
      description: Everything else is decided lower down
      match:
        layer4Configs:
          - ipProtocol: all

  # Where the policy is enforced: the organization itself. A node carries
  # one hierarchical policy association at a time.
  associations:
    - name: planton-oss-e2e-gcphfwp-org
      target:
        organizationId: "123456789012"

  # DELETE (default), PREVENT, ABANDON -- fans to the policy, every rule,
  # and every association.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.parent` | `GcpHierarchicalFirewallPolicyParent` | yes |  |  |
| `spec.parent.organizationId` | `string` |  |  |  |
| `spec.parent.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.shortName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.rules` | `[]GcpHierarchicalFirewallPolicyRule` |  |  |  |
| `spec.rules[].priority` | `int32` | yes |  |  |
| `spec.rules[].action` | `string` | yes |  |  |
| `spec.rules[].direction` | `string` | yes |  |  |
| `spec.rules[].description` | `string` |  |  |  |
| `spec.rules[].disabled` | `bool` |  |  |  |
| `spec.rules[].enableLogging` | `bool` |  |  |  |
| `spec.rules[].match` | `GcpHierarchicalFirewallPolicyRuleMatch` | yes |  |  |
| `spec.rules[].match.layer4Configs` | `[]GcpHierarchicalFirewallPolicyLayer4Config` | yes |  |  |
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
| `spec.rules[].targetResources` | `[]string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.rules[].targetSecureTags` | `[]string \| valueFrom` |  |  | GcpTagValue (`status.outputs.name`) |
| `spec.rules[].targetServiceAccounts` | `[]string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.rules[].securityProfileGroup` | `string` |  |  |  |
| `spec.rules[].tlsInspect` | `bool` |  |  |  |
| `spec.associations` | `[]GcpHierarchicalFirewallPolicyAssociation` |  |  |  |
| `spec.associations[].name` | `string` |  |  |  |
| `spec.associations[].target` | `GcpHierarchicalFirewallPolicyAttachmentTarget` | yes |  |  |
| `spec.associations[].target.organizationId` | `string` |  |  |  |
| `spec.associations[].target.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.parent

`GcpHierarchicalFirewallPolicyParent` · required

Where the policy lives: directly under the organization or inside a
folder. Exactly one arm. This is the policy's home for IAM and
quota purposes; WHERE IT IS ENFORCED is `associations`, which may
name the same node or others. Immutable: changing the parent
recreates the policy (and with it every rule and association).

- rule: {"required":true}
- rule: set exactly one of organization_id or folder_id -- a hierarchical firewall policy lives directly under the organization or inside one folder

### spec.parent.organizationId

`string`

An organization-level policy: the numeric organization ID (from
`gcloud organizations list`), without the `organizations/` prefix.

- rule: organization_id must be the numeric organization ID, without the organizations/ prefix

### spec.parent.folderId

`string | valueFrom`

A folder-level policy: the folder's numeric ID -- a literal, or a
reference to a GcpFolder resource (its folder_id output). The
reference is how a chart builds a landing zone: the policy waits for
its folder to exist and lives inside it.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.shortName

`string`

The policy's user-facing name, unique among the hierarchical policies
of the organization. Defaults to metadata.name when empty. RFC 1035:
1-63 characters, lowercase letters, digits, and hyphens, starting with
a letter and not ending with a hyphen. Immutable: Google identifies
the policy by a server-assigned numeric ID (the `policy_id` output)
and the short name is fixed at creation.

- rule: short_name must be 1-63 characters: lowercase letters, digits, and hyphens, starting with a letter and not ending with a hyphen (RFC 1035)

### spec.description

`string`

What this policy enforces and who owns it -- the operator reading a
denied connection's log sees the policy, not this manifest. Mutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.rules

`[]GcpHierarchicalFirewallPolicyRule`

The ordered rule set, evaluated from the LOWEST priority number up:
the first rule whose match fits the packet decides (allow, deny, or
goto_next, which delegates to the next policy level). Each entry is
its own Google resource keyed by `priority`, so changing a rule's
priority RECREATES that rule (the old priority is freed, the new one
is created) while every other rule is left alone; change a rule's
content in place and it is updated in place. An empty list is a valid
policy that decides nothing (Google's implied goto_next rules apply).

- rule: security_profile_group is required when action is apply_security_profile_group and must be empty for every other action
- rule: tls_inspect can be true only when action is apply_security_profile_group
- rule: enable_logging cannot be set on a goto_next rule -- Google logs only the rule that decides
- rule: target_secure_tags and target_service_accounts cannot both be set on one rule -- Google selects target VMs by one or the other

### spec.rules[].priority

`int32` · required · optional (explicit presence)

The rule's position in the evaluation order: 0 is evaluated first,
2147483647 last. Unique within the policy, and a rule's identity in
Google: changing it recreates the rule. Leave a gap between rules
(1000, 2000, ...) so a rule can be slotted in later without
renumbering. Priorities 2147483646 and 2147483647 are Google's implied
goto_next rules -- do not use them.

- rule: {"required":true,"int32":{"lte":2147483645,"gte":0}}

### spec.rules[].action

`string` · required

What happens to a matching packet:
  "allow"     -- permit it; evaluation of every lower level stops
  "deny"      -- drop it; evaluation stops
  "goto_next" -- delegate the decision to the next level (the child
                 folder's policy, then the network's policies, then
                 legacy VPC rules); the way an organization carves
                 out exceptions for a branch of the tree
  "apply_security_profile_group" -- send the traffic through the
                 Cloud NGFW security profile group named in
                 security_profile_group for layer-7 inspection, which
                 then allows or denies it
Lowercase, exactly as Google's API spells them.

- rule: action must be one of: allow, deny, goto_next, apply_security_profile_group
- rule: {"required":true}

### spec.rules[].direction

`string` · required

Which traffic the rule looks at: INGRESS (arriving at the target VMs;
`match` describes the SOURCE with src_* fields and may name the
destination) or EGRESS (leaving the target VMs; `match` describes the
DESTINATION with dest_* fields and may name the source). Immutable in
meaning: an ingress rule's src_secure_tags and src_networks describe
where traffic comes from, so they carry no meaning on an egress rule
and Google ignores them there.

- rule: direction must be INGRESS or EGRESS
- rule: {"required":true}

### spec.rules[].description

`string`

What this rule is for, in the words an operator reading a firewall
log will understand. Mutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.rules[].disabled

`bool`

When true the rule is kept but not enforced -- traffic behaves as if
the rule did not exist. The safe way to switch a rule off without
losing it (and its priority slot). Mutable.

### spec.rules[].enableLogging

`bool`

Log every connection this rule decides, to Cloud Logging (and from
there to BigQuery or Pub/Sub through a GcpLoggingSink). Logs carry the
rule, the policy, both endpoints, and the decision. Cannot be set on
a goto_next rule: Google logs only the rule that decides. Mutable.

### spec.rules[].match

`GcpHierarchicalFirewallPolicyRuleMatch` · required

The condition a packet must satisfy for this rule to decide it. At
least one layer4_configs entry is required; every other field narrows
the match and an empty field matches everything.

- rule: {"required":true}

### spec.rules[].match.layer4Configs

`[]GcpHierarchicalFirewallPolicyLayer4Config` · required

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

### spec.rules[].targetResources

`[]string | valueFrom`

Restrict the rule to VMs in these VPC networks -- references to
GcpVpcNetwork resources (their network_self_link output) or network
self-links as literals. Empty means every VM in every network beneath
the association's node. The way one rule in an organization policy
targets one shared VPC without a folder of its own.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.rules[].targetSecureTags

`[]string | valueFrom`

Restrict the rule to VMs carrying one of these secure tags --
references to GcpTagValue resources (their `name` output,
`tagValues/{numeric_id}`) or those names as literals. Maximum 256.
A secure tag is a Resource Manager tag value bound to the VM; a tag
whose value or network was deleted is INEFFECTIVE, and a rule whose
target tags are all ineffective is ignored. Cannot be combined with
target_service_accounts. Empty (with target_service_accounts also
empty) means every VM the rule can reach.

- references: GcpTagValue (`status.outputs.name`)
- rule: {"repeated":{"maxItems":"256"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpTagValue, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.rules[].targetServiceAccounts

`[]string | valueFrom`

Restrict the rule to VMs running as one of these service accounts --
references to GcpServiceAccount resources (their email output) or
email addresses as literals. Cannot be combined with
target_secure_tags.

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

`[]GcpHierarchicalFirewallPolicyAssociation`

Where the policy is ENFORCED: each entry attaches the policy to the
organization or to one folder, so every VPC network in every project
beneath that node evaluates these rules. A node can carry only one
hierarchical policy association at a time -- associating a second
policy with the same folder fails until the first is detached. Each
entry is its own Google resource; adding or removing one touches only
that association. A policy with no association exists but governs
nothing.

### spec.associations[].name

`string`

The association's name, unique within the policy. Defaults to
`<short_name>-<n>` (n = the entry's position, starting at 1) when
empty. Shown by `gcloud compute firewall-policies associations list`.

- rule: name must be 1-63 characters: lowercase letters, digits, and hyphens, starting with a letter and not ending with a hyphen (RFC 1035)

### spec.associations[].target

`GcpHierarchicalFirewallPolicyAttachmentTarget` · required

The organization or folder the policy is enforced on. Exactly one arm.

- rule: {"required":true}
- rule: set exactly one of organization_id or folder_id -- an association attaches the policy to the organization or to one folder

### spec.associations[].target.organizationId

`string`

Enforce on the whole organization: its numeric ID, without the
`organizations/` prefix.

- rule: organization_id must be the numeric organization ID, without the organizations/ prefix

### spec.associations[].target.folderId

`string | valueFrom`

Enforce on one folder and everything beneath it: the folder's numeric
ID -- a literal, or a reference to a GcpFolder (its folder_id output).
An attachment edge, not a placement: the policy does not live in this
folder.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What destroying this resource does in GCP, applied to the policy, to
every rule, and to every association together:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- associations are detached, rules removed, and the
               policy deleted; traffic falls through to the next
               policy level as if the policy had never existed
  "PREVENT" -- destroy FAILS; the guard for an organization's baseline
               deny rules
  "ABANDON" -- everything is removed from management but keeps
               existing and enforcing in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `rule_priorities_unique`: every rule must have a distinct priority -- Google keys a policy's rules by priority
- `association_names_unique`: every association must have a distinct name within the policy

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpHierarchicalFirewallPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_id` | `string` | The policy's server-assigned numeric ID -- Google's `name` for a hierarchical policy, and what `gcloud compute firewall-policies describe` and the firewall logs identify it by. |
| `status.outputs.short_name` | `string` | The policy's user-facing short name, as declared (or defaulted from metadata.name). |
| `status.outputs.self_link` | `string` | The policy's self-link URL, `https://www.googleapis.com/compute/v1/locations/global/firewallPolicies/{policy_id}`. |
| `status.outputs.parent` | `string` | The node the policy lives under, `organizations/{id}` or `folders/{id}`. |
| `status.outputs.rule_tuple_count` | `int64` | Google's complexity measure for the whole rule set -- the sum of each rule's tuple count (protocols x ports x sources x targets). An organization has a quota of tuples across its hierarchical policies; watch this before adding wide rules. |
| `status.outputs.association_names` | `[]string` | The association names, in the order declared -- what `gcloud compute firewall-policies associations list` shows and what a detach targets. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.parent.folderId` | GcpFolder | `status.outputs.folder_id` |
| `spec.rules[].match.srcSecureTags` | GcpTagValue | `status.outputs.name` |
| `spec.rules[].match.srcNetworks` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.rules[].targetResources` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.rules[].targetSecureTags` | GcpTagValue | `status.outputs.name` |
| `spec.rules[].targetServiceAccounts` | GcpServiceAccount | `status.outputs.email` |
| `spec.associations[].target.folderId` | GcpFolder | `status.outputs.folder_id` |

## See Also

- [Overview](../README.md)
