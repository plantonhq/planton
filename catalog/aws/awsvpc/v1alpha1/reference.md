# AwsVpc

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsVpcSpec defines an AWS Virtual Private Cloud (VPC): an isolated virtual
network in which other AWS resources are launched.

A VPC is the networking foundation for nearly everything else on AWS, but it
is itself a thin construct: an IP address space (one primary IPv4 CIDR, plus
optional secondary IPv4/IPv6 CIDRs), a tenancy mode, a few DNS toggles, and
an optional encryption-in-transit control. The things that make a network
useful -- subnets, internet gateways, NAT gateways, route tables -- are
separate, independently ownable components that reference this VPC. Keeping
the VPC a clean building block is what lets a topology be composed from
first-class nodes rather than hidden inside one opaque resource.

The primary IPv4 CIDR can be specified explicitly (cidr_block) or allocated
from an IPAM pool (ipv4_ipam_pool_id [+ ipv4_netmask_length]). IPv6 is opt-in
and supports either an Amazon-provided /56 (assign_generated_ipv6_cidr_block)
or an IPAM-allocated block (ipv6_ipam_pool_id). Additional address space is
attached in place through secondary_ipv4_cidrs / secondary_ipv6_cidrs. The
cross-field rules below mirror what AWS itself enforces.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsVpc
metadata:
  name: awsvpc-demo
spec:
  region: us-west-2
  cidrBlock: 10.0.0.0/16
  enableDnsHostnames: true
  # Additional address space grows in place: an explicit block, an IPAM-sized
  # allocation, or a pool-pinned block per entry.
  secondaryIpv4Cidrs:
    - cidrBlock: 10.1.0.0/16
    - ipamPoolId:
        value: ipam-pool-0demo1234abcd
      netmaskLength: 20
  # Additional IPv6 ranges: one source per entry (Amazon-provided here).
  secondaryIpv6Cidrs:
    - assignGenerated: true
  # VPC Encryption Control: observe unencrypted traffic paths before
  # enforcing (move to enforce with exclusions once findings are clean).
  encryptionControl:
    mode: monitor
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.cidrBlock` | `string` |  |  |  |
| `spec.secondaryIpv4Cidrs` | `[]AwsVpcSecondaryIpv4Cidr` |  |  |  |
| `spec.secondaryIpv4Cidrs[].cidrBlock` | `string` |  |  |  |
| `spec.secondaryIpv4Cidrs[].ipamPoolId` | `string \| valueFrom` |  |  |  |
| `spec.secondaryIpv4Cidrs[].netmaskLength` | `int32` |  |  |  |
| `spec.ipv4IpamPoolId` | `string \| valueFrom` |  |  |  |
| `spec.ipv4NetmaskLength` | `int32` |  |  |  |
| `spec.instanceTenancy` | `string` |  | `default` |  |
| `spec.enableDnsSupport` | `bool` |  | `true` |  |
| `spec.enableDnsHostnames` | `bool` |  | `true` |  |
| `spec.enableNetworkAddressUsageMetrics` | `bool` |  |  |  |
| `spec.assignGeneratedIpv6CidrBlock` | `bool` |  |  |  |
| `spec.ipv6CidrBlock` | `string` |  |  |  |
| `spec.ipv6CidrBlockNetworkBorderGroup` | `string` |  |  |  |
| `spec.ipv6IpamPoolId` | `string \| valueFrom` |  |  |  |
| `spec.ipv6NetmaskLength` | `int32` |  |  |  |
| `spec.secondaryIpv6Cidrs` | `[]AwsVpcSecondaryIpv6Cidr` |  |  |  |
| `spec.secondaryIpv6Cidrs[].assignGenerated` | `bool` |  |  |  |
| `spec.secondaryIpv6Cidrs[].ipv6Pool` | `string` |  |  |  |
| `spec.secondaryIpv6Cidrs[].ipamPoolId` | `string \| valueFrom` |  |  |  |
| `spec.secondaryIpv6Cidrs[].cidrBlock` | `string` |  |  |  |
| `spec.secondaryIpv6Cidrs[].netmaskLength` | `int32` |  |  |  |
| `spec.encryptionControl` | `AwsVpcEncryptionControl` |  |  |  |
| `spec.encryptionControl.mode` | `string` | yes |  |  |
| `spec.encryptionControl.excludeInternetGateway` | `bool` |  |  |  |
| `spec.encryptionControl.excludeEgressOnlyInternetGateway` | `bool` |  |  |  |
| `spec.encryptionControl.excludeNatGateway` | `bool` |  |  |  |
| `spec.encryptionControl.excludeVirtualPrivateGateway` | `bool` |  |  |  |
| `spec.encryptionControl.excludeVpcPeering` | `bool` |  |  |  |
| `spec.encryptionControl.excludeVpcLattice` | `bool` |  |  |  |
| `spec.encryptionControl.excludeLambda` | `bool` |  |  |  |
| `spec.encryptionControl.excludeElasticFileSystem` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the VPC will be created.
Example: "us-west-2", "eu-west-1", "ap-southeast-1".
For the list of regions, see:
https://aws.amazon.com/about-aws/global-infrastructure/regions_az/

- rule: {"string":{"minLen":"1"}}

### spec.cidrBlock

`string`

The primary IPv4 CIDR block for the VPC. This is the VPC's main address
range and is immutable: changing it replaces the VPC. The mask must be
between /16 and /28 (AWS limits), e.g. "10.0.0.0/16" yields 65,536
addresses. Leave empty only when allocating the primary CIDR from IPAM
(set ipv4_ipam_pool_id instead); exactly one of cidr_block or
ipv4_ipam_pool_id must be provided.

### spec.secondaryIpv4Cidrs

`[]AwsVpcSecondaryIpv4Cidr`

Additional IPv4 address ranges to associate with the VPC beyond the
primary one. Each entry is associated as its own resource and can be
added or removed without recreating the VPC -- the standard way to grow
a network whose original range ran out of subnet space, or to carve a
distinct range (such as the 100.64.0.0/10 shared space) for specific
workloads. Each entry names an explicit CIDR or an IPAM allocation.

- rule: set cidr_block or ipam_pool_id (or both, to pin a specific block from the pool)
- rule: cidr_block must be an IPv4 CIDR with a /16-/28 mask, e.g. 10.1.0.0/16
- rule: netmask_length must be between 16 and 28, requires ipam_pool_id, and is mutually exclusive with cidr_block

### spec.secondaryIpv4Cidrs[].cidrBlock

`string`

An explicit IPv4 CIDR block to associate (e.g. "10.1.0.0/16" or
"100.64.0.0/16"). The mask must be between /16 and /28 and the range must
not overlap the VPC's other CIDRs. When ipam_pool_id is also set, this
pins a specific block from that pool. Immutable: changing it replaces the
association (the VPC itself is untouched).

### spec.secondaryIpv4Cidrs[].ipamPoolId

`string | valueFrom`

The IPAM pool to allocate this secondary CIDR from, instead of (or in
addition to, when pinning a specific block) cidr_block. Supply a literal
ipam-pool-id; there is no IPAM pool catalog kind yet. Immutable: changing
it replaces the association.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.secondaryIpv4Cidrs[].netmaskLength

`int32`

The netmask length of the CIDR to allocate from ipam_pool_id (e.g. 16 for
a /16). Must be between 16 and 28. Requires ipam_pool_id and is mutually
exclusive with cidr_block. Immutable: changing it replaces the
association.

### spec.ipv4IpamPoolId

`string | valueFrom`

The IPAM (IP Address Manager) pool to allocate the primary IPv4 CIDR from,
instead of specifying cidr_block directly. Pair with ipv4_netmask_length to
let IPAM choose a block of the requested size, or with cidr_block to take a
specific block from the pool. Supply a literal ipam-pool-id; there is no
IPAM pool catalog kind yet. Immutable: changing it replaces the VPC.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.ipv4NetmaskLength

`int32`

The netmask length of the primary IPv4 CIDR to allocate from
ipv4_ipam_pool_id (e.g. 16 for a /16). Must be between 16 and 28. Requires
ipv4_ipam_pool_id and is mutually exclusive with cidr_block. Immutable:
changing it replaces the VPC.

### spec.instanceTenancy

`string`

The tenancy of instances launched into this VPC: "default" (instances may
run on shared hardware) or "dedicated" (every instance runs on
single-tenant hardware, at higher cost). Leave empty for the AWS default
("default"). Note that AWS only supports changing "dedicated" -> "default"
in place; changing "default" -> "dedicated" replaces the VPC.

- default: `default`
- rule: instance_tenancy must be one of: default, dedicated

### spec.enableDnsSupport

`bool` · optional (explicit presence)

Whether the Amazon-provided DNS server (the .2 resolver) answers DNS
queries from within the VPC. AWS enables this by default, and disabling it
breaks name resolution for most workloads, so when this field is unset the
VPC keeps DNS resolution ON. Set it explicitly to false only to deliberately
turn the Amazon resolver off.

- default: `true`

### spec.enableDnsHostnames

`bool`

Whether instances with public IP addresses receive public DNS hostnames.
AWS leaves this off by default; turn it on for VPCs whose instances should
be reachable by DNS name. See:
https://docs.aws.amazon.com/vpc/latest/userguide/vpc-dns.html#vpc-dns-hostnames

- default: `true`

### spec.enableNetworkAddressUsageMetrics

`bool`

Whether to enable Network Address Usage (NAU) metrics for the VPC. NAU is a
CloudWatch metric that tracks how the VPC's IP address space is consumed,
useful for capacity planning in large networks. Off by default.

### spec.assignGeneratedIpv6CidrBlock

`bool`

Request an Amazon-provided IPv6 /56 CIDR block for the VPC. This is the
simplest way to make a VPC dual-stack. Mutually exclusive with the IPAM
IPv6 fields (ipv6_ipam_pool_id / ipv6_cidr_block / ipv6_netmask_length).

### spec.ipv6CidrBlock

`string`

An explicit IPv6 CIDR block to allocate from ipv6_ipam_pool_id (e.g.
"2600:1f18:abcd:1200::/56"). The prefix length must be one of /44, /48,
/52, /56, or /60 (the AWS VPC IPv6 sizes). Requires ipv6_ipam_pool_id (a
bring-your-own IPv6 block without IPAM is not supported on the VPC
resource) and is mutually exclusive with ipv6_netmask_length.

### spec.ipv6CidrBlockNetworkBorderGroup

`string`

The network border group from which to advertise an Amazon-provided IPv6
CIDR (a Local Zone / Wavelength advertisement scope). Only valid together
with assign_generated_ipv6_cidr_block.

### spec.ipv6IpamPoolId

`string | valueFrom`

The IPAM pool to allocate the IPv6 CIDR from. Pair with ipv6_netmask_length
(the size to allocate) or ipv6_cidr_block (a specific block). Supply a
literal ipam-pool-id; there is no IPAM pool catalog kind yet. Mutually
exclusive with assign_generated_ipv6_cidr_block.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.ipv6NetmaskLength

`int32`

The netmask length of the IPv6 CIDR to allocate from ipv6_ipam_pool_id.
Must be one of 44, 48, 52, 56, or 60. Requires ipv6_ipam_pool_id and is
mutually exclusive with ipv6_cidr_block.

### spec.secondaryIpv6Cidrs

`[]AwsVpcSecondaryIpv6Cidr`

Additional IPv6 address ranges to associate with the VPC beyond the one
configured above. Each entry is associated as its own resource and can be
added or removed without recreating the VPC. Each entry names exactly one
source: an Amazon-provided block, a BYOIP public pool, or an IPAM pool.

- rule: set exactly one IPv6 source: assign_generated, ipv6_pool, or ipam_pool_id
- rule: cidr_block cannot be combined with assign_generated
- rule: cidr_block must be an IPv6 CIDR with a /44, /48, /52, /56, or /60 prefix
- rule: netmask_length must be one of 44, 48, 52, 56, 60, requires ipam_pool_id, and is mutually exclusive with cidr_block

### spec.secondaryIpv6Cidrs[].assignGenerated

`bool`

Request an Amazon-provided IPv6 /56 block. Mutually exclusive with the
pool sources and with cidr_block (Amazon chooses the range).

### spec.secondaryIpv6Cidrs[].ipv6Pool

`string`

The BYOIP public IPv6 pool (ipv6pool-ec2-...) to take the range from --
address space you brought to AWS outside of IPAM. Pair with cidr_block to
pin the specific block. Immutable: changing it replaces the association.

### spec.secondaryIpv6Cidrs[].ipamPoolId

`string | valueFrom`

The IPAM pool to allocate the range from. Pair with netmask_length (the
size to allocate) or cidr_block (a specific block). Supply a literal
ipam-pool-id; there is no IPAM pool catalog kind yet. Immutable: changing
it replaces the association.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.secondaryIpv6Cidrs[].cidrBlock

`string`

An explicit IPv6 CIDR block, used with ipv6_pool or ipam_pool_id to pin
the exact range (e.g. "2600:1f18:abcd:1200::/56"). The prefix length must
be one of /44, /48, /52, /56, or /60. Mutually exclusive with
netmask_length and with assign_generated. Immutable: changing it replaces
the association.

### spec.secondaryIpv6Cidrs[].netmaskLength

`int32`

The netmask length to allocate from ipam_pool_id. Must be one of 44, 48,
52, 56, or 60 (AWS only sizes IPAM allocations -- pool and Amazon blocks
pin their own size). Mutually exclusive with cidr_block. Immutable:
changing it replaces the association.

### spec.encryptionControl

`AwsVpcEncryptionControl`

VPC Encryption Control: monitor or enforce encryption in transit for
traffic entering, leaving, and moving within the VPC. Omit to leave the
feature off (the AWS default). Start in "monitor" mode to observe, then
move to "enforce" once the exclusions below cover every path that cannot
encrypt.

- rule: exclusions only apply when mode is 'enforce'

### spec.encryptionControl.mode

`string` · required

The control mode: "monitor" (observe and report unencrypted paths) or
"enforce" (block unencrypted traffic, honoring the exclusions below).

- rule: {"required":true,"string":{"in":["monitor","enforce"]}}

### spec.encryptionControl.excludeInternetGateway

`bool`

Exclude internet gateway traffic from enforcement.

### spec.encryptionControl.excludeEgressOnlyInternetGateway

`bool`

Exclude egress-only internet gateway traffic from enforcement.

### spec.encryptionControl.excludeNatGateway

`bool`

Exclude NAT gateway traffic from enforcement.

### spec.encryptionControl.excludeVirtualPrivateGateway

`bool`

Exclude virtual private gateway (Site-to-Site VPN) traffic from
enforcement.

### spec.encryptionControl.excludeVpcPeering

`bool`

Exclude VPC peering traffic from enforcement.

### spec.encryptionControl.excludeVpcLattice

`bool`

Exclude VPC Lattice traffic from enforcement.

### spec.encryptionControl.excludeLambda

`bool`

Exclude Lambda function traffic from enforcement.

### spec.encryptionControl.excludeElasticFileSystem

`bool`

Exclude Elastic File System (EFS) traffic from enforcement.

## Validation Rules

- `ipv4_primary_source_required`: set exactly one primary IPv4 source: cidr_block or ipv4_ipam_pool_id
- `ipv4_cidr_block_vs_netmask_exclusive`: cidr_block and ipv4_netmask_length are mutually exclusive
- `ipv4_cidr_block_format`: cidr_block must be an IPv4 CIDR with a /16-/28 mask, e.g. 10.0.0.0/16
- `ipv4_netmask_length_valid`: ipv4_netmask_length must be between 16 and 28 and requires ipv4_ipam_pool_id
- `ipv6_amazon_vs_ipam_exclusive`: assign_generated_ipv6_cidr_block cannot be combined with the IPAM IPv6 fields (ipv6_ipam_pool_id/ipv6_cidr_block/ipv6_netmask_length)
- `ipv6_cidr_block_requires_ipam`: ipv6_cidr_block requires ipv6_ipam_pool_id
- `ipv6_cidr_block_format`: ipv6_cidr_block must be an IPv6 CIDR with a /44, /48, /52, /56, or /60 prefix, e.g. 2600:1f18:abcd:1200::/56
- `ipv6_netmask_length_valid`: ipv6_netmask_length must be one of 44, 48, 52, 56, 60 and requires ipv6_ipam_pool_id
- `ipv6_cidr_block_vs_netmask_exclusive`: ipv6_cidr_block and ipv6_netmask_length are mutually exclusive
- `ipv6_border_group_requires_amazon`: ipv6_cidr_block_network_border_group requires assign_generated_ipv6_cidr_block

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsVpc, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.vpc_id` | `string` | The ID of the VPC (e.g. "vpc-0abc123"). The primary handle other resources reference via status.outputs.vpc_id. |
| `status.outputs.vpc_arn` | `string` | The ARN of the VPC. |
| `status.outputs.cidr_block` | `string` | The primary IPv4 CIDR block of the VPC (e.g. "10.0.0.0/16"). Reflects the allocated block even when it was assigned from IPAM. |
| `status.outputs.ipv6_cidr_block` | `string` | The IPv6 CIDR block associated with the VPC (e.g. "2600:1f18:abcd:1200::/56"), empty when the VPC is IPv4-only. |
| `status.outputs.owner_id` | `string` | The AWS account ID that owns the VPC. |
| `status.outputs.main_route_table_id` | `string` | The ID of the VPC's main route table. Subnets with no explicit route-table association use this table. |
| `status.outputs.default_security_group_id` | `string` | The ID of the default security group created with the VPC. |
| `status.outputs.default_network_acl_id` | `string` | The ID of the default network ACL created with the VPC. |
| `status.outputs.default_route_table_id` | `string` | The ID of the default route table created with the VPC (equal to the main route table). |
| `status.outputs.region` | `string` | The region the VPC was created in (mirrors spec.region, included for convenience). |
| `status.outputs.secondary_ipv4_cidr_association_ids` | `map<string, string>` | The association IDs ("vpc-cidr-assoc-...") of the secondary IPv4 CIDR blocks, keyed exactly as the module keys each association: by the entry's literal CIDR, or "ipam-<index>" for IPAM-sized entries. These are the handles `aws ec2 disassociate-vpc-cidr-block` and state import take; AWS generates them at association time. |
| `status.outputs.secondary_ipv6_cidr_association_ids` | `map<string, string>` | The association IDs of the secondary IPv6 CIDR blocks, keyed exactly as the module keys each association: by the entry's pinned CIDR, or "ipv6-<index>" for Amazon-provided/BYOIP-pool/IPAM entries. |
| `status.outputs.encryption_control_id` | `string` | The ID ("vpcec-...") of the VPC Encryption Control resource, empty when spec.encryption_control is not configured. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAppRunnerService | `spec.vpcIngressConnections[].vpcId` | `status.outputs.vpc_id` |
| AwsBedrockAgentCoreEvaluation | `spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.vpcId` | `status.outputs.vpc_id` |
| AwsBedrockAgentCoreEvaluation | `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.vpcId` | `status.outputs.vpc_id` |
| AwsBedrockAgentCoreGateway | `spec.customJwtAuthorizer.privateEndpoint.managedVpc.vpcId` | `status.outputs.vpc_id` |
| AwsBedrockAgentCoreGateway | `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.vpcId` | `status.outputs.vpc_id` |
| AwsBedrockAgentCoreGateway | `spec.targets[].privateEndpoint.managedVpc.vpcId` | `status.outputs.vpc_id` |
| AwsBedrockAgentCoreRuntime | `spec.customJwtAuthorizer.privateEndpoint.managedVpc.vpcId` | `status.outputs.vpc_id` |
| AwsBedrockAgentCoreRuntime | `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.vpcId` | `status.outputs.vpc_id` |
| AwsClientVpn | `spec.vpcId` | `status.outputs.vpc_id` |
| AwsCloudMapNamespace | `spec.vpcId` | `status.outputs.vpc_id` |
| AwsCodeBuildProject | `spec.vpcConfig.vpcId` | `status.outputs.vpc_id` |
| AwsEgressOnlyInternetGateway | `spec.vpcId` | `status.outputs.vpc_id` |
| AwsInternetGateway | `spec.vpcId` | `status.outputs.vpc_id` |
| AwsLbTargetGroup | `spec.vpcId` | `status.outputs.vpc_id` |
| AwsNatGateway | `spec.vpcId` | `status.outputs.vpc_id` |
| AwsNetworkAcl | `spec.vpcId` | `status.outputs.vpc_id` |
| AwsRedshiftCluster | `spec.endpointAuthorizations[].vpcIds` | `status.outputs.vpc_id` |
| AwsRoute53ResolverEndpoint | `spec.rules[].vpcIds` | `status.outputs.vpc_id` |
| AwsRoute53ResolverFirewall | `spec.vpcAssociations[].vpcId` | `status.outputs.vpc_id` |
| AwsRoute53ResolverQueryLog | `spec.vpcIds` | `status.outputs.vpc_id` |
| AwsRoute53Zone | `spec.vpcAssociations[].vpcId` | `status.outputs.vpc_id` |
| AwsSagemakerDomain | `spec.vpcId` | `status.outputs.vpc_id` |
| AwsSecurityGroup | `spec.vpcId` | `status.outputs.vpc_id` |
| AwsSecurityGroup | `spec.additionalVpcIds` | `status.outputs.vpc_id` |
| AwsSubnet | `spec.vpcId` | `status.outputs.vpc_id` |
| AwsTransitGatewayVpcAttachment | `spec.vpcId` | `status.outputs.vpc_id` |
| AwsVpcEndpoint | `spec.vpcId` | `status.outputs.vpc_id` |
| AwsVpcPeering | `spec.request.vpcId` | `status.outputs.vpc_id` |
| AwsVpcPeering | `spec.request.peerVpcId` | `status.outputs.vpc_id` |
| KubernetesNetworkPolicy | `spec.ingressRules[].from[].ipBlock.cidr` | `status.outputs.cidr_block` |
| KubernetesNetworkPolicy | `spec.ingressRules[].from[].ipBlock.except` | `status.outputs.cidr_block` |
| KubernetesNetworkPolicy | `spec.egressRules[].to[].ipBlock.cidr` | `status.outputs.cidr_block` |
| KubernetesNetworkPolicy | `spec.egressRules[].to[].ipBlock.except` | `status.outputs.cidr_block` |

## See Also

- [Overview](../README.md)
