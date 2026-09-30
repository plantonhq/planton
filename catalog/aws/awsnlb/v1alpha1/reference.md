# AwsNlb

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsNlbSpec defines a Network Load Balancer: the Layer-4 entry point for
TCP/UDP/TLS traffic, with static IP addresses and millions of requests per
second of headroom.

The load balancer itself carries no routing configuration -- that is
deliberate. Listeners (AwsLbListener) attach to it and own ports,
protocols, and TLS material; target groups (AwsLbTargetGroup) receive the
connections. NLB listeners only forward (AWS rejects every other action
type at Layer 4), and rules do not apply -- routing is purely by
port/protocol. This spec owns only what is truly load-balancer-wide:
node placement with optional static IPs, security groups, and traffic
distribution behavior.

Key NLB characteristics versus ALB:
- Static IP addresses via Elastic IP allocation per subnet (internet-facing)
- TCP, UDP, TLS, and TCP_UDP listener protocols
- Cross-zone load balancing is configurable (default: disabled, unlike ALB)
- Security groups are optional (unlike ALB where they are effectively
  required) -- but once attached, they can never be fully removed

The NLB name comes from metadata.name. AWS limits the name to 32
characters; both IaC modules truncate longer names deterministically.
Changing "internal" replaces the load balancer.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsNlb
metadata:
  name: awsnlb-demo
spec:
  region: us-west-2
  # Two mappings with Elastic IPs exercise the static-IP shape that
  # differentiates NLB from ALB; the behavior attributes exercise the
  # enum-validated strings so the fixture proves the full variable contract.
  subnetMappings:
    - subnetId:
        value: subnet-12345678
      allocationId:
        value: eipalloc-0123456789abcdef0
    - subnetId:
        value: subnet-12345679
      allocationId:
        value: eipalloc-0fedcba9876543210
  internal: false
  crossZoneLoadBalancingEnabled: true
  ipAddressType: ipv4
  dnsRecordClientRoutingPolicy: availability_zone_affinity
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.subnetMappings` | `[]AwsNlbSubnetMapping` | yes |  |  |
| `spec.subnetMappings[].subnetId` | `string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.subnetMappings[].allocationId` | `string \| valueFrom` |  |  | AwsElasticIp (`status.outputs.allocation_id`) |
| `spec.subnetMappings[].privateIpv4Address` | `string` |  |  |  |
| `spec.subnetMappings[].ipv6Address` | `string` |  |  |  |
| `spec.securityGroups` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.internal` | `bool` |  |  |  |
| `spec.deleteProtectionEnabled` | `bool` |  |  |  |
| `spec.crossZoneLoadBalancingEnabled` | `bool` |  |  |  |
| `spec.ipAddressType` | `string` |  | `ipv4` |  |
| `spec.dnsRecordClientRoutingPolicy` | `string` |  |  |  |
| `spec.zonalShiftEnabled` | `bool` |  |  |  |
| `spec.enforceSecurityGroupInboundRulesOnPrivateLinkTraffic` | `string` |  |  |  |
| `spec.accessLogs` | `AwsNlbAccessLogs` |  |  |  |
| `spec.accessLogs.bucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.accessLogs.prefix` | `string` |  |  |  |
| `spec.dns` | `AwsNlbDns` |  |  |  |
| `spec.dns.enabled` | `bool` |  |  |  |
| `spec.dns.route53ZoneId` | `string \| valueFrom` |  |  | AwsRoute53Zone (`status.outputs.zone_id`) |
| `spec.dns.hostnames` | `[]string` |  |  |  |
| `spec.minimumLoadBalancerCapacityUnits` | `int32` |  |  |  |
| `spec.secondaryIpsAutoAssignedPerSubnet` | `int32` |  |  |  |
| `spec.enablePrefixForIpv6SourceNat` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the NLB is created.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.subnetMappings

`[]AwsNlbSubnetMapping` · required

Subnet mappings define where NLB nodes are placed and optionally assign
static IP addresses. Each mapping corresponds to one Availability Zone.

For internet-facing NLBs, provide an allocation_id (Elastic IP) per
subnet to get static public IPs -- a primary reason to choose NLB over
ALB. For internal NLBs, optionally pin a private_ipv4_address per subnet.

At least one subnet mapping is required. AWS recommends at least two for
high availability across Availability Zones.

- rule: {"required":true,"repeated":{"minItems":"1"}}

### spec.subnetMappings[].subnetId

`string | valueFrom` · required

Subnet ID where the NLB node is placed. Must be in a VPC that supports
the NLB's scheme (public subnet for internet-facing, private for internal).

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.subnetMappings[].allocationId

`string | valueFrom`

Elastic IP allocation ID for a static public IP address on this NLB node.
Only valid for internet-facing NLBs. Each subnet mapping can have at most
one Elastic IP.

- references: AwsElasticIp (`status.outputs.allocation_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsElasticIp, name: <that resource's name>, fieldPath: status.outputs.allocation_id}} -- a bare string does not parse

### spec.subnetMappings[].privateIpv4Address

`string`

Specific private IPv4 address for the NLB node in this subnet. Only valid
for internal NLBs. The address must belong to the subnet's CIDR range.
When omitted, AWS assigns a private IP automatically.

### spec.subnetMappings[].ipv6Address

`string`

Specific IPv6 address for the NLB node in this subnet, for dualstack
NLBs (ip_address_type = "dualstack"). Must fall within the subnet's IPv6
CIDR. When omitted, AWS assigns one automatically. Pinning it gives the
IPv6 side the same static-address story allocation_id gives the IPv4
side.

### spec.securityGroups

`[]string | valueFrom`

Security group IDs to attach to the NLB. Unlike ALB, security groups are
optional for NLB. When omitted, the NLB accepts all traffic on configured
listener ports. When provided, security group rules filter inbound
traffic.

Important: once security groups are attached to an NLB, they cannot be
fully removed -- at least one must remain. Plan accordingly.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.internal

`bool`

When true, creates an internal NLB accessible only within the VPC.
When false (default), creates an internet-facing NLB with public DNS.
Immutable: changing the scheme replaces the load balancer.

### spec.deleteProtectionEnabled

`bool`

Prevents deletion of the NLB while enabled. Recommended for production:
deleting an NLB silently orphans every listener attached to it, and any
Elastic IPs pinned to it start billing as unattached.

### spec.crossZoneLoadBalancingEnabled

`bool`

Distribute traffic evenly across all registered targets in all enabled
Availability Zones. Default is false for NLB (unlike ALB where
cross-zone is always enabled). Enable this when target distribution
across AZs is uneven -- and budget for the inter-AZ data transfer it
introduces.

### spec.ipAddressType

`string`

IP address type for the NLB. Controls whether the NLB uses IPv4 only or
dual-stack (IPv4 + IPv6).
Valid values: "ipv4" (default), "dualstack".

- default: `ipv4`

### spec.dnsRecordClientRoutingPolicy

`string`

Controls how DNS queries from clients are routed to NLB nodes across
Availability Zones. Affects latency and cross-zone traffic costs.

Valid values:
- "any_availability_zone" (default): Clients may be routed to any AZ.
- "availability_zone_affinity": Clients are routed to the AZ of the
  resolver, reducing cross-zone traffic. Best when targets are evenly
  distributed.
- "partial_availability_zone_affinity": 85% of requests stay in the
  resolver's AZ, 15% spill to other AZs. Balances affinity with
  availability.

### spec.zonalShiftEnabled

`bool`

Allows Amazon Application Recovery Controller to shift this NLB's
traffic away from an impaired Availability Zone.

### spec.enforceSecurityGroupInboundRulesOnPrivateLinkTraffic

`string`

Whether inbound security-group rules are enforced on traffic arriving
through PrivateLink VPC endpoints. Valid values: "on", "off". AWS
default: "on" for NLBs created with security groups. Only meaningful
when security groups are attached.

### spec.accessLogs

`AwsNlbAccessLogs`

Access logs delivered to S3. NLB access logs only capture TLS-listener
traffic (an AWS limitation -- plain TCP/UDP flows are not logged). The
bucket must carry the ELB log-delivery bucket policy. When omitted,
access logging is off.

### spec.accessLogs.bucket

`string | valueFrom` · required

The S3 bucket receiving the logs.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.accessLogs.prefix

`string`

Key prefix inside the bucket (e.g. "nlb/production"), for sharing one
bucket across several load balancers.

### spec.dns

`AwsNlbDns`

Optional Route53 DNS configuration. When enabled, creates alias A records
pointing the specified hostnames to the NLB's DNS name.

### spec.dns.enabled

`bool`

When true, creates Route53 alias records for the NLB.

### spec.dns.route53ZoneId

`string | valueFrom`

Route53 hosted zone ID where alias records are created.
Required when enabled is true.

- references: AwsRoute53Zone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRoute53Zone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.dns.hostnames

`[]string`

Domain names that will point to the NLB via Route53 alias records.
Each hostname gets its own A record aliased to the NLB's DNS name.

- rule: {"repeated":{"unique":true}}

### spec.minimumLoadBalancerCapacityUnits

`int32` · optional (explicit presence)

Reserved load balancer capacity, in Load Balancer Capacity Units (LCUs).
Pre-provisions the NLB for a known traffic level (product launches,
failover targets) instead of waiting for organic scaling. Reserved
capacity bills for the reservation whether used or not. Leave unset for
normal on-demand scaling.

- rule: {"int32":{"gte":1}}

### spec.secondaryIpsAutoAssignedPerSubnet

`int32` · optional (explicit presence)

Number of secondary private IPv4 addresses (0-7) AWS auto-assigns to each
NLB node, raising the source-port budget for very high connection counts
to a single target. Provider-verified caveat: DECREASING this value
replaces the load balancer (AWS cannot release secondary IPs in place) --
plan capacity before setting it. Leave unset for the AWS default of 0.

- rule: {"int32":{"lte":7,"gte":0}}

### spec.enablePrefixForIpv6SourceNat

`string`

Whether NLB nodes source-NAT IPv6 traffic through a /80 prefix per AZ
instead of a single IPv6 address -- required for UDP listeners on a
dualstack NLB (the per-address port budget cannot serve UDP flow hashing).
Valid values: "on", "off". Only meaningful when ip_address_type is
"dualstack".

## Validation Rules

- `ip_address_type_valid`: ip_address_type must be 'ipv4' or 'dualstack' when set
- `dns_record_client_routing_policy_valid`: dns_record_client_routing_policy must be 'any_availability_zone', 'availability_zone_affinity', or 'partial_availability_zone_affinity' when set
- `private_link_enforcement_valid`: enforce_security_group_inbound_rules_on_private_link_traffic must be 'on' or 'off' when set
- `ipv6_source_nat_prefix_valid`: enable_prefix_for_ipv6_source_nat must be 'on' or 'off' when set
- `ipv6_source_nat_requires_dualstack`: enable_prefix_for_ipv6_source_nat requires ip_address_type 'dualstack'

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsNlb, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.load_balancer_arn` | `string` | ARN of the Network Load Balancer. The primary handle other resources reference via status.outputs.load_balancer_arn -- listeners attach through it, and IAM policies, CloudWatch alarms, and Global Accelerator endpoints all take this value. |
| `status.outputs.load_balancer_name` | `string` | Name assigned to the NLB (metadata.name, truncated to the 32-character AWS limit when necessary), for console URLs and CLI queries. |
| `status.outputs.load_balancer_dns_name` | `string` | DNS name automatically assigned by AWS (e.g., "my-nlb-abc123.elb.us-east-1.amazonaws.com"). Use this to create CNAME records or as the target for Route53 alias records. |
| `status.outputs.load_balancer_hosted_zone_id` | `string` | Route53 hosted zone ID for the NLB's DNS name. Required when creating Route53 alias records that point to this NLB. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetMappings[].subnetId` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.subnetMappings[].allocationId` | AwsElasticIp | `status.outputs.allocation_id` |
| `spec.securityGroups` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.accessLogs.bucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.dns.route53ZoneId` | AwsRoute53Zone | `status.outputs.zone_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsLbListener | `spec.loadBalancerArn` | `status.outputs.load_balancer_arn` |
| AwsRestApiVpcLink | `spec.targetArn` | `status.outputs.load_balancer_arn` |

## See Also

- [Overview](../README.md)
