# AwsNetworkAcl

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsNetworkAclSpec defines one network ACL: the stateless
subnet-level firewall. Rules are evaluated in rule-number order per
direction, first match wins, and - unlike security groups - a rule
can DENY. Replies are NOT tracked: a stateless firewall needs
explicit rules in both directions (the classic NACL trap is allowing
inbound 443 and forgetting the outbound ephemeral-port reply range).

Rules and subnet associations are managed here in-line as the single
declarative owner - the standalone aws_network_acl_rule /
aws_network_acl_association resources carry identical payloads and
fight the in-line form, so this kind never uses them. AWS's own
catch-all rules (32767 for IPv4, 32768 for IPv6) always exist, are
invisible here, and cannot be managed - this spec's rules take
numbers 1-32766 below them.

Deleting the ACL re-parents its subnets to the VPC's default NACL
first (subnets always belong to SOME ACL), then deletes.

## Example

```yaml
# Canonical AwsNetworkAcl example (hack/dev manifest and refgen Example
# source): a web-tier subnet firewall - HTTPS in, ephemeral replies
# out, ICMP diagnostics, and a deny for a known-bad range.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsNetworkAcl
metadata:
  name: web-tier-acl
  id: web-tier-acl
  org: test-org
  env: dev
spec:
  region: us-west-2
  vpcId:
    value: vpc-0123456789abcdef0
  ingress:
    - ruleNo: 90
      action: deny
      protocol: "-1"
      cidrBlock: 198.51.100.0/24
    - ruleNo: 100
      action: allow
      protocol: tcp
      cidrBlock: 0.0.0.0/0
      fromPort: 443
      toPort: 443
    - ruleNo: 110
      action: allow
      protocol: icmp
      cidrBlock: 10.0.0.0/16
      icmpType: 8
      icmpCode: 0
  egress:
    - ruleNo: 100
      action: allow
      protocol: "6"
      cidrBlock: 0.0.0.0/0
      fromPort: 1024
      toPort: 65535
  subnetIds:
    - value: subnet-0123456789abcdef0
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.ingress` | `[]AwsNetworkAclRule` |  |  |  |
| `spec.ingress[].ruleNo` | `int32` |  |  |  |
| `spec.ingress[].action` | `string` |  |  |  |
| `spec.ingress[].protocol` | `string` | yes |  |  |
| `spec.ingress[].cidrBlock` | `string` |  |  |  |
| `spec.ingress[].ipv6CidrBlock` | `string` |  |  |  |
| `spec.ingress[].fromPort` | `int32` |  |  |  |
| `spec.ingress[].toPort` | `int32` |  |  |  |
| `spec.ingress[].icmpType` | `int32` |  |  |  |
| `spec.ingress[].icmpCode` | `int32` |  |  |  |
| `spec.egress` | `[]AwsNetworkAclRule` |  |  |  |
| `spec.egress[].ruleNo` | `int32` |  |  |  |
| `spec.egress[].action` | `string` |  |  |  |
| `spec.egress[].protocol` | `string` | yes |  |  |
| `spec.egress[].cidrBlock` | `string` |  |  |  |
| `spec.egress[].ipv6CidrBlock` | `string` |  |  |  |
| `spec.egress[].fromPort` | `int32` |  |  |  |
| `spec.egress[].toPort` | `int32` |  |  |  |
| `spec.egress[].icmpType` | `int32` |  |  |  |
| `spec.egress[].icmpCode` | `int32` |  |  |  |
| `spec.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |

## Field Details

### spec.region

`string` · required

The AWS region the ACL lives in. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.vpcId

`string | valueFrom` · required

The VPC this ACL belongs to. Fixed for life - changing it replaces
the ACL. Reference an AwsVpc vpc_id output or pass a literal
vpc-... id.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.ingress

`[]AwsNetworkAclRule`

Inbound rules, evaluated in rule_no order (first match wins).
Traffic matching no rule falls through to AWS's invisible
catch-all DENY.

- rule: set exactly one of cidr_block (IPv4) and ipv6_cidr_block (IPv6) - a rule matches one address family; add a second rule for the other
- rule: protocol -1/all matches every port - leave from_port and to_port unset (0)
- rule: icmp_type/icmp_code apply only to ICMP protocols (1/icmp, 58/ipv6-icmp)
- rule: icmp_type -1 (all types) requires icmp_code -1 (all codes)

### spec.ingress[].ruleNo

`int32`

The rule's evaluation position, 1-32766 (lower evaluates first).
Leave gaps (100, 200, ...) so rules can be inserted later without
renumbering. Unique within the direction.

- rule: {"int32":{"lte":32766,"gte":1}}

### spec.ingress[].action

`string`

Whether matching traffic is allowed or denied - NACLs can deny,
which is exactly what security groups cannot do.

- rule: {"string":{"in":["allow","deny"]}}

### spec.ingress[].protocol

`string` · required

The protocol, as an IANA number ("6") or name ("tcp"). "-1" (or
"all") matches every protocol. AWS stores only NUMBERS - the
provider normalizes names when diffing, so a name here never
causes a perpetual diff.

- rule: {"string":{"minLen":"1"}}

### spec.ingress[].cidrBlock

`string`

The IPv4 CIDR this rule matches. Example: "10.0.0.0/16", or
"0.0.0.0/0" for everything. Exactly one of cidr_block and
ipv6_cidr_block per rule.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.ingress[].ipv6CidrBlock

`string`

The IPv6 CIDR this rule matches. Example: "2001:db8::/32", or
"::/0" for everything.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.ingress[].fromPort

`int32`

The port range start (0-65535). With protocol -1/all, leave both
ports unset - the wildcard matches every port. For single-port
rules set from_port == to_port.

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.ingress[].toPort

`int32`

The port range end (0-65535).

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.ingress[].icmpType

`int32` · optional (explicit presence)

The ICMP type this rule matches (-1 for all types). Only for ICMP
protocols. Presence-typed so type 0 (echo reply) is expressible.

- rule: {"int32":{"lte":255,"gte":-1}}

### spec.ingress[].icmpCode

`int32` · optional (explicit presence)

The ICMP code this rule matches (-1 for all codes). Only for ICMP
protocols. Presence-typed so code 0 is expressible.

- rule: {"int32":{"lte":255,"gte":-1}}

### spec.egress

`[]AwsNetworkAclRule`

Outbound rules, evaluated in rule_no order (first match wins).
Remember the reply direction: stateless means outbound replies to
inbound connections need their own allow (ephemeral ports
1024-65535 for most clients).

- rule: set exactly one of cidr_block (IPv4) and ipv6_cidr_block (IPv6) - a rule matches one address family; add a second rule for the other
- rule: protocol -1/all matches every port - leave from_port and to_port unset (0)
- rule: icmp_type/icmp_code apply only to ICMP protocols (1/icmp, 58/ipv6-icmp)
- rule: icmp_type -1 (all types) requires icmp_code -1 (all codes)

### spec.egress[].ruleNo

`int32`

The rule's evaluation position, 1-32766 (lower evaluates first).
Leave gaps (100, 200, ...) so rules can be inserted later without
renumbering. Unique within the direction.

- rule: {"int32":{"lte":32766,"gte":1}}

### spec.egress[].action

`string`

Whether matching traffic is allowed or denied - NACLs can deny,
which is exactly what security groups cannot do.

- rule: {"string":{"in":["allow","deny"]}}

### spec.egress[].protocol

`string` · required

The protocol, as an IANA number ("6") or name ("tcp"). "-1" (or
"all") matches every protocol. AWS stores only NUMBERS - the
provider normalizes names when diffing, so a name here never
causes a perpetual diff.

- rule: {"string":{"minLen":"1"}}

### spec.egress[].cidrBlock

`string`

The IPv4 CIDR this rule matches. Example: "10.0.0.0/16", or
"0.0.0.0/0" for everything. Exactly one of cidr_block and
ipv6_cidr_block per rule.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.egress[].ipv6CidrBlock

`string`

The IPv6 CIDR this rule matches. Example: "2001:db8::/32", or
"::/0" for everything.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.egress[].fromPort

`int32`

The port range start (0-65535). With protocol -1/all, leave both
ports unset - the wildcard matches every port. For single-port
rules set from_port == to_port.

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.egress[].toPort

`int32`

The port range end (0-65535).

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.egress[].icmpType

`int32` · optional (explicit presence)

The ICMP type this rule matches (-1 for all types). Only for ICMP
protocols. Presence-typed so type 0 (echo reply) is expressible.

- rule: {"int32":{"lte":255,"gte":-1}}

### spec.egress[].icmpCode

`int32` · optional (explicit presence)

The ICMP code this rule matches (-1 for all codes). Only for ICMP
protocols. Presence-typed so code 0 is expressible.

- rule: {"int32":{"lte":255,"gte":-1}}

### spec.subnetIds

`[]string | valueFrom`

The subnets this ACL filters. A subnet has exactly ONE ACL:
listing it here atomically replaces its previous association (AWS
has no attach - only replace). Removing a subnet from this list
hands it back to the VPC's default NACL. Reference AwsSubnet
subnet_id outputs or pass literal subnet-... ids.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

## Validation Rules

- `spec.ingress_rule_numbers_unique`: ingress rules must have unique rule_no values
- `spec.egress_rule_numbers_unique`: egress rules must have unique rule_no values

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsNetworkAcl, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.network_acl_id` | `string` | The ACL's id (acl-...) - the provider's import ID. |
| `status.outputs.network_acl_arn` | `string` | The ACL's ARN. |
| `status.outputs.owner_id` | `string` | The AWS account that owns the ACL. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |

## See Also

- [Overview](../README.md)
