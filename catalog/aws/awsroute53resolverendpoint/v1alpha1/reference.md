# AwsRoute53ResolverEndpoint

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsRoute53ResolverEndpointSpec defines one Route 53 Resolver endpoint
- the hybrid-DNS bridge between a VPC and networks outside it - with
its forwarding rules and their VPC associations managed in-line.

An INBOUND endpoint exposes ENI IP addresses that on-prem resolvers
forward queries TO. An OUTBOUND endpoint sends queries FROM the VPC
to on-prem (or other) name servers, steered by the forwarding rules
below. An INBOUND_DELEGATION endpoint serves the delegation feature
that pairs with DELEGATE rules.

Rules ride the endpoint because forwarding is what an outbound
endpoint exists for: FORWARD rules bind to this endpoint and carry
target name servers; SYSTEM rules override a broader FORWARD rule
for a subdomain (recursive resolution instead of forwarding) and are
meaningless without the forwarding they punch a hole in. Each rule
associates to the VPCs that should honor it.

Associating a rule RAM-shared FROM another account is out of scope
here (single-account posture) - it lands as its own arm on demand.

## Example

```yaml
# Canonical AwsRoute53ResolverEndpoint example (hack/dev manifest and
# refgen Example source): an outbound endpoint forwarding corp.example.com
# queries to two on-prem name servers, with the rule associated to one VPC
# and a SYSTEM override keeping a subdomain on recursive resolution.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53ResolverEndpoint
metadata:
  name: corp-outbound
  id: corp-outbound
  org: test-org
  env: dev
spec:
  region: us-west-2
  direction: OUTBOUND
  ipAddresses:
    - subnetId:
        value: subnet-0123456789abcdef0
    - subnetId:
        value: subnet-0fedcba9876543210
  securityGroupIds:
    - value: sg-0123456789abcdef0
  rules:
    - name: corp-forward
      domainName: corp.example.com
      ruleType: FORWARD
      targetIps:
        - ip: 10.20.0.53
        - ip: 10.20.1.53
          port: 8053
      vpcIds:
        - value: vpc-0123456789abcdef0
    - name: cloud-native-subdomain
      domainName: aws.corp.example.com
      ruleType: SYSTEM
      vpcIds:
        - value: vpc-0123456789abcdef0
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.direction` | `string` |  |  |  |
| `spec.ipAddresses` | `[]AwsRoute53ResolverEndpointIpAddress` | yes |  |  |
| `spec.ipAddresses[].subnetId` | `string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.ipAddresses[].ip` | `string` |  |  |  |
| `spec.ipAddresses[].ipv6` | `string` |  |  |  |
| `spec.securityGroupIds` | `[]string \| valueFrom` | yes |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.endpointType` | `string` |  |  |  |
| `spec.protocols` | `[]string` |  |  |  |
| `spec.rniEnhancedMetricsEnabled` | `bool` |  |  |  |
| `spec.targetNameServerMetricsEnabled` | `bool` |  |  |  |
| `spec.rules` | `[]AwsRoute53ResolverEndpointRule` |  |  |  |
| `spec.rules[].name` | `string` | yes |  |  |
| `spec.rules[].domainName` | `string` | yes |  |  |
| `spec.rules[].ruleType` | `string` |  |  |  |
| `spec.rules[].targetIps` | `[]AwsRoute53ResolverEndpointRuleTarget` |  |  |  |
| `spec.rules[].targetIps[].ip` | `string` |  |  |  |
| `spec.rules[].targetIps[].ipv6` | `string` |  |  |  |
| `spec.rules[].targetIps[].port` | `int64` |  |  |  |
| `spec.rules[].targetIps[].protocol` | `string` |  |  |  |
| `spec.rules[].vpcIds` | `[]string \| valueFrom` |  |  | AwsVpc (`status.outputs.vpc_id`) |

## Field Details

### spec.region

`string` · required

The AWS region the endpoint (and its rules) live in. Example:
"us-west-2".

- rule: {"string":{"minLen":"1"}}

### spec.direction

`string`

Which way DNS queries flow through this endpoint. INBOUND receives
queries from outside the VPC (on-prem resolvers target the ENI
IPs); OUTBOUND sends queries out through the forwarding rules;
INBOUND_DELEGATION serves delegated subdomains (pairs with
DELEGATE rules). Fixed for life.

- rule: {"string":{"in":["INBOUND","OUTBOUND","INBOUND_DELEGATION"]}}

### spec.ipAddresses

`[]AwsRoute53ResolverEndpointIpAddress` · required

The ENIs this endpoint answers or originates on: one entry per
subnet (2-10, different AZs recommended by AWS). Each entry may
pin a fixed private IPv4/IPv6 within its subnet; unset lets AWS
pick one.

- rule: {"repeated":{"minItems":"2","maxItems":"10"}}

### spec.ipAddresses[].subnetId

`string | valueFrom` · required

The subnet the ENI lives in. Reference an AwsSubnet subnet_id
output or pass a literal subnet-... id.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.ipAddresses[].ip

`string`

Pin a specific private IPv4 within the subnet. Unset lets AWS
pick a free address.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"ipv4":true}}

### spec.ipAddresses[].ipv6

`string`

Pin a specific IPv6 within the subnet (IPV6/DUALSTACK endpoints).
Unset lets AWS pick a free address.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"ipv6":true}}

### spec.securityGroupIds

`[]string | valueFrom` · required

Security groups controlling DNS traffic to/from the endpoint
ENIs (allow TCP+UDP 53 from the peers that will use it).
Reference AwsSecurityGroup security_group_id outputs or pass
literal sg-... ids. Fixed for life at the provider.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"required":true,"repeated":{"minItems":"1","maxItems":"64"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.endpointType

`string`

The endpoint's IP family. Unset lets AWS default (IPV4).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["IPV4","IPV6","DUALSTACK"]}}

### spec.protocols

`[]string`

The DNS transport protocols the endpoint speaks (at most 2).
"Do53" is plain DNS; "DoH" is DNS-over-HTTPS; "DoH-FIPS" is the
FIPS-compliant variant (inbound only). Unset lets AWS default
(Do53).

- rule: protocols must be unique
- rule: {"repeated":{"maxItems":"2","items":{"string":{"in":["Do53","DoH","DoH-FIPS"]}}}}

### spec.rniEnhancedMetricsEnabled

`bool` · optional (explicit presence)

Publish enhanced resolver network interface metrics to CloudWatch.
Unset leaves AWS's default; an explicit false turns them off.

### spec.targetNameServerMetricsEnabled

`bool` · optional (explicit presence)

Publish per-target-name-server metrics to CloudWatch (outbound
endpoints). Unset leaves AWS's default; an explicit false turns
them off.

### spec.rules

`[]AwsRoute53ResolverEndpointRule`

The forwarding rules managed with this endpoint, keyed by name.
Each rule steers queries for one domain and associates to the
VPCs that should honor it.

- rule: rule_type FORWARD requires at least one target_ips entry - the name servers queries forward to
- rule: rule_type SYSTEM does not accept target_ips - it restores recursive resolution for the domain
- rule: each target_ips entry sets exactly one of ip and ipv6

### spec.rules[].name

`string` · required

Rule name - the for_each key on both engines and the key in the
rule_ids / association ID outputs. Also becomes the rule's AWS
name.

- rule: {"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z_-]+$"}}

### spec.rules[].domainName

`string` · required

The domain whose queries this rule steers, e.g. "corp.example.com".
AWS strips a trailing dot; the most specific matching rule wins.

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.rules[].ruleType

`string`

What the rule does with matching queries. FORWARD sends them to
target_ips through this (outbound) endpoint. SYSTEM restores
recursive resolution for a subdomain of a forwarded domain.
DELEGATE serves the delegation feature (pairs with an
INBOUND_DELEGATION endpoint; shape enforced server-side).
RECURSIVE rules are Resolver-owned (the autodefined rules) - AWS
rejects user-created ones, so the vocabulary excludes it.

- rule: {"string":{"in":["FORWARD","SYSTEM","DELEGATE"]}}

### spec.rules[].targetIps

`[]AwsRoute53ResolverEndpointRuleTarget`

The name servers FORWARD queries go to (on-prem or other
resolvers reachable from the endpoint's subnets).

### spec.rules[].targetIps[].ip

`string`

The target's IPv4 address.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"ipv4":true}}

### spec.rules[].targetIps[].ipv6

`string`

The target's IPv6 address.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"ipv6":true}}

### spec.rules[].targetIps[].port

`int64`

The port queries are sent to. Unset means AWS's default (53).

- rule: port must be between 1 and 65535

### spec.rules[].targetIps[].protocol

`string`

The transport to this target. Unset means AWS's default (Do53).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Do53","DoH","DoH-FIPS"]}}

### spec.rules[].vpcIds

`[]string | valueFrom`

The VPCs this rule is associated with (the rule takes effect for
queries originating in these VPCs). Reference AwsVpc vpc_id
outputs or pass literal vpc-... ids.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

## Validation Rules

- `spec.forward_rules_require_outbound`: rules with rule_type FORWARD require direction OUTBOUND - forwarding is an outbound endpoint's job
- `spec.rule_names_unique`: rule names must be unique within the endpoint

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRoute53ResolverEndpoint, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.endpoint_id` | `string` | The endpoint's id (rslvr-in-... / rslvr-out-...) - the provider's import ID. |
| `status.outputs.endpoint_arn` | `string` | The endpoint's ARN. |
| `status.outputs.host_vpc_id` | `string` | The VPC the endpoint's subnets belong to (AWS derives it). |
| `status.outputs.ip_addresses` | `[]string` | The ENI IP addresses the endpoint answers or originates on - what on-prem resolvers target (inbound) or what firewalls must allow (outbound). |
| `status.outputs.rule_ids` | `map<string, string>` | AWS-generated rule IDs (rslvr-rr-...) keyed by rule name - what rule associations and imports reference. |
| `status.outputs.rule_association_ids` | `map<string, string>` | AWS-generated rule association IDs (rslvr-rrassoc-...) keyed by "{rule_name}//{vpc_id}". |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.ipAddresses[].subnetId` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.rules[].vpcIds` | AwsVpc | `status.outputs.vpc_id` |

## See Also

- [Overview](../README.md)
