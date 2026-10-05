# GcpFirewallRule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpFirewallRuleSpec defines the configuration for a Google Compute Engine firewall rule.

A firewall rule controls traffic to and from VM instances in a VPC network.
Each rule either allows or denies traffic matching a set of protocol/port combinations,
filtered by source/destination ranges and instance tags or service accounts.

GCP constraint: tag-based targeting (source_tags, target_tags) and service-account-based
targeting (source_service_accounts, target_service_accounts) are mutually exclusive.
You cannot use both in the same rule.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirewallRule
metadata:
  name: test-firewall
  annotations:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: test-org
    pulumi.planton.dev/project: test-project
    pulumi.planton.dev/stack.name: dev.GcpFirewallRule.test-firewall
spec:
  projectId:
    value: my-gcp-project-123
  network:
    value: default
  ruleName: allow-http-ingress
  direction: INGRESS
  action: ALLOW
  rules:
    - protocol: tcp
      ports:
        - "80"
        - "443"
  priority: 1000
  description: Allow HTTP and HTTPS traffic from the internet
  sourceRanges:
    - "0.0.0.0/0"
  targetTags:
    - web-server
  # Resource Manager tags for org-policy conditions and IAM scoping
  # (tagKeys/tagValues IDs, not short names). Changing them REPLACES the
  # firewall rule — plan tag changes deliberately.
  resourceManagerTags:
    tagKeys/281475123456789: tagValues/281476987654321
  # Delete the rule on destroy (GCP's default, made explicit). PREVENT
  # protects a rule other teams' traffic may silently depend on.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.ruleName` | `string` | yes |  |  |
| `spec.direction` | `string` | yes |  |  |
| `spec.action` | `string` | yes |  |  |
| `spec.rules` | `[]GcpFirewallProtocolPort` | yes |  |  |
| `spec.rules[].protocol` | `string` | yes |  |  |
| `spec.rules[].ports` | `[]string` |  |  |  |
| `spec.priority` | `int32` |  | `1000` |  |
| `spec.description` | `string` |  |  |  |
| `spec.sourceRanges` | `[]string` |  |  |  |
| `spec.destinationRanges` | `[]string` |  |  |  |
| `spec.sourceTags` | `[]string` |  |  |  |
| `spec.targetTags` | `[]string` |  |  |  |
| `spec.sourceServiceAccounts` | `[]string` |  |  |  |
| `spec.targetServiceAccounts` | `[]string` |  |  |  |
| `spec.disabled` | `bool` |  |  |  |
| `spec.logConfig` | `GcpFirewallLogConfig` |  |  |  |
| `spec.logConfig.metadata` | `string` | yes |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project ID in which to create this firewall rule.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.network

`string | valueFrom` · required

VPC network to which this firewall rule applies.
Accepts a network name (e.g., "default") or a full self-link URL.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.ruleName

`string` · required

Name of the firewall rule in GCP.
Must be 1-63 characters, lowercase letters, numbers, or hyphens.
Must start with a lowercase letter and end with a letter or number.
Example: "allow-http-ingress", "deny-all-egress"

- rule: {"required":true,"string":{"pattern":"^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"}}

### spec.direction

`string` · required

Traffic direction this rule applies to.
"INGRESS" matches inbound traffic to instances; "EGRESS" matches outbound traffic.

- rule: direction must be INGRESS or EGRESS
- rule: {"required":true}

### spec.action

`string` · required

Action to take when the rule matches traffic.
"ALLOW" permits the matched traffic; "DENY" blocks it.

- rule: action must be ALLOW or DENY
- rule: {"required":true}

### spec.rules

`[]GcpFirewallProtocolPort` · required

Protocol and port combinations this rule matches.
At least one rule must be specified.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].protocol

`string` · required

IP protocol to match. Accepted values: "tcp", "udp", "icmp", "esp", "ah", "sctp", "ipip", "all",
or an IANA protocol number (e.g., "6" for TCP).

- rule: {"required":true}

### spec.rules[].ports

`[]string`

Ports or port ranges to match. Only applicable when protocol is "tcp",
"udp", or "sctp" — including their IANA numbers ("6", "17", "132").
Each entry is a single port (e.g., "80") or a range (e.g., "8000-9000").
Omit for protocols that do not use ports (icmp, esp, ah, etc.).

### spec.priority

`int32` · optional (explicit presence)

Rule priority. Range: 0-65535. Lower values indicate higher priority.
At the same priority, DENY rules take precedence over ALLOW rules.
Default: 1000

- default: `1000`
- rule: {"int32":{"lte":65535,"gte":0}}

### spec.description

`string`

Human-readable description for the firewall rule.
Example: "Allow HTTP/HTTPS traffic from the internet"

### spec.sourceRanges

`[]string`

Source IPv4 or IPv6 CIDR ranges for INGRESS rules.
Traffic is only matched when the source IP falls within one of these ranges.
For INGRESS rules, at least one of source_ranges, source_tags, or source_service_accounts is required.
Example: ["0.0.0.0/0"] to match all IPv4 traffic, ["10.0.0.0/8"] for internal.

### spec.destinationRanges

`[]string`

Destination IPv4 or IPv6 CIDR ranges for EGRESS rules.
Traffic is only matched when the destination IP falls within one of these ranges.
If omitted on EGRESS rules, GCP defaults to ["0.0.0.0/0"] (all destinations).

### spec.sourceTags

`[]string`

Source instance network tags for INGRESS rules.
Traffic from instances with any of these tags is matched.
Cannot be combined with source_service_accounts or target_service_accounts.

### spec.targetTags

`[]string`

Target instance network tags.
The rule applies only to instances that have one of these tags.
If omitted, the rule applies to all instances in the network.
Cannot be combined with source_service_accounts or target_service_accounts.

### spec.sourceServiceAccounts

`[]string`

Source service accounts for INGRESS rules. Max 10.
Traffic from instances running as any of these service accounts is matched.
Cannot be combined with source_tags or target_tags.

- rule: {"repeated":{"maxItems":"10"}}

### spec.targetServiceAccounts

`[]string`

Target service accounts. Max 10.
The rule applies only to instances running as one of these service accounts.
If omitted, the rule applies to all instances in the network.
Cannot be combined with source_tags or target_tags.

- rule: {"repeated":{"maxItems":"10"}}

### spec.disabled

`bool`

Whether the firewall rule is disabled.
A disabled rule exists in the configuration but is not enforced.
Useful for temporarily suspending a rule without deleting it.

### spec.logConfig

`GcpFirewallLogConfig`

Logging configuration. When present, firewall logging is enabled.
Omit this field to disable logging (the default).

### spec.logConfig.metadata

`string` · required

Metadata inclusion mode for firewall logs.

- rule: metadata must be EXCLUDE_ALL_METADATA or INCLUDE_ALL_METADATA
- rule: {"required":true}

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the firewall rule at create time, for
org-policy conditions and IAM scoping. Keys are "tagKeys/{tag_key_id}"
and values "tagValues/{tag_value_id}" (IDs, not short names). Changing
tags REPLACES the firewall rule (the provider sends them through a
create-only params block) -- plan tag changes deliberately.

### spec.deletionPolicy

`string`

What happens to the firewall rule in GCP when this resource is destroyed.
  "DELETE"  -- (GCP's default when unset) the rule is deleted; traffic
               it allowed is cut over to the next matching rule
  "PREVENT" -- destroy FAILS; protects a rule other teams' traffic
               may silently depend on
  "ABANDON" -- the rule is removed from management but keeps enforcing
               in GCP (free at rest; clean it up manually)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `ingress_requires_source`: INGRESS rules must specify at least one of source_ranges, source_tags, or source_service_accounts
- `tags_and_service_accounts_mutually_exclusive`: tag-based targeting (source_tags, target_tags) and service-account-based targeting (source_service_accounts, target_service_accounts) cannot be used together

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpFirewallRule, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.firewall_self_link` | `string` | Full self-link URI of the created firewall rule. Example: "projects/my-project/global/firewalls/allow-http-ingress" Useful for expressing depends_on or uses relationships in infra charts. |
| `status.outputs.firewall_name` | `string` | Name of the firewall rule as it exists in GCP. |
| `status.outputs.creation_timestamp` | `string` | RFC3339 timestamp of when the firewall rule was created. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |

## See Also

- [Overview](../README.md)
