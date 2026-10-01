# AwsSecurityGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsSecurityGroupSpec defines an AWS EC2 Security Group in a specified VPC.
A security group is a stateful virtual firewall attached to network
interfaces: return traffic for an allowed connection is automatically
permitted, so rules only describe the initiating direction. Rules are
authored inline on the group -- they live and die with it -- and other
resources compose onto the group by referencing its exported
security_group_id.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSecurityGroup
metadata:
  name: awssecuritygroup-demo
spec:
  region: us-west-2
  vpcId:
    value: vpc-0123456789abcdef0
  description: Demo security group
  ingress:
    - protocol: tcp
      fromPort: 80
      toPort: 80
      ipv4Cidrs:
        - 0.0.0.0/0
      description: Allow HTTP inbound
  egress:
    - protocol: tcp
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - 0.0.0.0/0
      description: Allow HTTPS outbound
  # Share the same group into another VPC (same account and region) --
  # resources there attach it instead of maintaining a per-VPC copy.
  additionalVpcIds:
    - value: vpc-0f9e8d7c6b5a43210
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.description` | `string` | yes |  |  |
| `spec.ingress` | `[]SecurityGroupRule` |  |  |  |
| `spec.ingress[].protocol` | `string` | yes |  |  |
| `spec.ingress[].fromPort` | `int32` |  |  |  |
| `spec.ingress[].toPort` | `int32` |  |  |  |
| `spec.ingress[].ipv4Cidrs` | `[]string` |  |  |  |
| `spec.ingress[].ipv6Cidrs` | `[]string` |  |  |  |
| `spec.ingress[].sourceSecurityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.ingress[].destinationSecurityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.ingress[].prefixListIds` | `[]string` |  |  |  |
| `spec.ingress[].selfReference` | `bool` |  |  |  |
| `spec.ingress[].description` | `string` |  |  |  |
| `spec.egress` | `[]SecurityGroupRule` |  |  |  |
| `spec.egress[].protocol` | `string` | yes |  |  |
| `spec.egress[].fromPort` | `int32` |  |  |  |
| `spec.egress[].toPort` | `int32` |  |  |  |
| `spec.egress[].ipv4Cidrs` | `[]string` |  |  |  |
| `spec.egress[].ipv6Cidrs` | `[]string` |  |  |  |
| `spec.egress[].sourceSecurityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.egress[].destinationSecurityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.egress[].prefixListIds` | `[]string` |  |  |  |
| `spec.egress[].selfReference` | `bool` |  |  |  |
| `spec.egress[].description` | `string` |  |  |  |
| `spec.revokeRulesOnDelete` | `bool` |  |  |  |
| `spec.additionalVpcIds` | `[]string \| valueFrom` |  |  | AwsVpc (`status.outputs.vpc_id`) |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.vpcId

`string | valueFrom` · required

vpc_id is the ID of the VPC where this Security Group will be created.
Example: "vpc-12345abcde"
Required: every security group belongs to exactly one VPC.
ForceNew: changing the VPC forces group replacement.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.description

`string` · required

description provides a short explanation of this Security Group's purpose.
Required by AWS. ForceNew: AWS does not allow editing a group description
in place, so changing it forces group replacement.
Example: "Allows inbound HTTP and SSH for web tier"

- rule: Description must not exceed 255 characters
- rule: AWS allows only letters, digits, spaces and ._-:/()#,@[]+=&;{}!$* in a security group description
- rule: {"required":true}

### spec.ingress

`[]SecurityGroupRule`

ingress defines the inbound traffic rules for this Security Group.
If empty, inbound traffic is fully restricted (deny all).

- rule: when protocol is '-1' (all protocols), from_port and to_port must both be 0

### spec.ingress[].protocol

`string` · required

protocol indicates the protocol for the rule.
Common values: "tcp", "udp", "icmp", "icmpv6", or "-1" (all protocols).
IANA protocol numbers are also accepted.

- rule: {"required":true}

### spec.ingress[].fromPort

`int32`

from_port is the starting port in the range. For single-port rules,
from_port == to_port. For ICMP/ICMPv6, from_port is the ICMP TYPE
(-1 means all types). For all-protocol rules (protocol "-1"), both ports
must be 0.

- rule: {"int32":{"lte":65535,"gte":-1}}

### spec.ingress[].toPort

`int32`

to_port is the ending port in the range. For single-port rules,
to_port == from_port. For ICMP/ICMPv6, to_port is the ICMP CODE
(-1 means all codes). For all-protocol rules (protocol "-1"), both ports
must be 0.

- rule: {"int32":{"lte":65535,"gte":-1}}

### spec.ingress[].ipv4Cidrs

`[]string`

ipv4_cidrs is the list of IPv4 CIDR blocks allowed (ingress) or targeted (egress).
Examples: "10.0.0.0/16", "0.0.0.0/0"
If empty, no IPv4 CIDRs are included in this rule.

### spec.ingress[].ipv6Cidrs

`[]string`

ipv6_cidrs is the list of IPv6 CIDR blocks allowed or targeted.
Example: "::/0"
If empty, no IPv6 CIDRs are included in this rule.

### spec.ingress[].sourceSecurityGroupIds

`[]string | valueFrom`

source_security_group_ids is the list of Security Group IDs that can send traffic (for ingress).
Typically used for internal traffic between resources. Can reference other AwsSecurityGroup resources.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.ingress[].destinationSecurityGroupIds

`[]string | valueFrom`

destination_security_group_ids is the list of Security Group IDs that receive traffic (for egress).
Useful for restricting outbound traffic to specific groups. Can reference other AwsSecurityGroup resources.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.ingress[].prefixListIds

`[]string`

prefix_list_ids is the list of managed prefix list IDs allowed (ingress)
or targeted (egress). Example: "pl-63a5400a". Managed prefix lists name a
set of CIDRs by a stable ID: AWS-managed lists cover services like S3 and
DynamoDB gateway endpoints (so an egress rule can target "the S3 service"
instead of hardcoding its CIDRs), and customer-managed lists let network
teams maintain shared CIDR sets (office ranges, partner networks) that
many groups reference without copying.

- rule: {"repeated":{"unique":true,"items":{"string":{"pattern":"^pl-[0-9a-f]+$"}}}}

### spec.ingress[].selfReference

`bool`

self_reference indicates whether to allow traffic from/to the same Security Group.
This is equivalent to referencing the group's own ID -- the standard pattern
for intra-cluster traffic (nodes of one cluster talking to each other).

### spec.ingress[].description

`string`

description is an optional explanation of this specific rule,
aiding in clarity and maintenance. Max 255 chars, from the same character
set AWS allows in a group description.

- rule: Rule description must not exceed 255 characters
- rule: AWS allows only letters, digits, spaces and ._-:/()#,@[]+=&;{}!$* in a security group rule description

### spec.egress

`[]SecurityGroupRule`

egress defines the outbound traffic rules for this Security Group.
If empty, ALL outbound traffic is denied: the module revokes the allow-all
egress rule AWS adds to every new group, so the manifest is the complete
statement of what the group permits. Add an explicit all-traffic egress
rule (protocol "-1", 0.0.0.0/0) to restore the AWS default behavior.

- rule: when protocol is '-1' (all protocols), from_port and to_port must both be 0

### spec.egress[].protocol

`string` · required

protocol indicates the protocol for the rule.
Common values: "tcp", "udp", "icmp", "icmpv6", or "-1" (all protocols).
IANA protocol numbers are also accepted.

- rule: {"required":true}

### spec.egress[].fromPort

`int32`

from_port is the starting port in the range. For single-port rules,
from_port == to_port. For ICMP/ICMPv6, from_port is the ICMP TYPE
(-1 means all types). For all-protocol rules (protocol "-1"), both ports
must be 0.

- rule: {"int32":{"lte":65535,"gte":-1}}

### spec.egress[].toPort

`int32`

to_port is the ending port in the range. For single-port rules,
to_port == from_port. For ICMP/ICMPv6, to_port is the ICMP CODE
(-1 means all codes). For all-protocol rules (protocol "-1"), both ports
must be 0.

- rule: {"int32":{"lte":65535,"gte":-1}}

### spec.egress[].ipv4Cidrs

`[]string`

ipv4_cidrs is the list of IPv4 CIDR blocks allowed (ingress) or targeted (egress).
Examples: "10.0.0.0/16", "0.0.0.0/0"
If empty, no IPv4 CIDRs are included in this rule.

### spec.egress[].ipv6Cidrs

`[]string`

ipv6_cidrs is the list of IPv6 CIDR blocks allowed or targeted.
Example: "::/0"
If empty, no IPv6 CIDRs are included in this rule.

### spec.egress[].sourceSecurityGroupIds

`[]string | valueFrom`

source_security_group_ids is the list of Security Group IDs that can send traffic (for ingress).
Typically used for internal traffic between resources. Can reference other AwsSecurityGroup resources.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.egress[].destinationSecurityGroupIds

`[]string | valueFrom`

destination_security_group_ids is the list of Security Group IDs that receive traffic (for egress).
Useful for restricting outbound traffic to specific groups. Can reference other AwsSecurityGroup resources.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.egress[].prefixListIds

`[]string`

prefix_list_ids is the list of managed prefix list IDs allowed (ingress)
or targeted (egress). Example: "pl-63a5400a". Managed prefix lists name a
set of CIDRs by a stable ID: AWS-managed lists cover services like S3 and
DynamoDB gateway endpoints (so an egress rule can target "the S3 service"
instead of hardcoding its CIDRs), and customer-managed lists let network
teams maintain shared CIDR sets (office ranges, partner networks) that
many groups reference without copying.

- rule: {"repeated":{"unique":true,"items":{"string":{"pattern":"^pl-[0-9a-f]+$"}}}}

### spec.egress[].selfReference

`bool`

self_reference indicates whether to allow traffic from/to the same Security Group.
This is equivalent to referencing the group's own ID -- the standard pattern
for intra-cluster traffic (nodes of one cluster talking to each other).

### spec.egress[].description

`string`

description is an optional explanation of this specific rule,
aiding in clarity and maintenance. Max 255 chars, from the same character
set AWS allows in a group description.

- rule: Rule description must not exceed 255 characters
- rule: AWS allows only letters, digits, spaces and ._-:/()#,@[]+=&;{}!$* in a security group rule description

### spec.revokeRulesOnDelete

`bool`

revoke_rules_on_delete forcibly revokes this group's rules (and rules in
OTHER groups that reference this one) before deleting the group. Without
it, deleting a group that is still referenced by another group's rules
fails with a DependencyViolation. Enable for groups that are referenced
cross-group (e.g. an app tier referenced by a database tier) so teardown
never requires manual rule surgery. Safe to toggle in place.

### spec.additionalVpcIds

`[]string | valueFrom`

additional_vpc_ids shares this security group into OTHER VPCs beyond its
home vpc_id, so resources in those VPCs can attach the same group instead
of maintaining copied groups per VPC (one firewall definition, many VPCs).
Each entry must be a different VPC in the same account and region as the
group's own VPC, and must not repeat vpc_id or another entry — AWS rejects
duplicates at apply (entries may be references, so this is not validated
here). Reference AwsVpc vpc_id outputs or pass literal "vpc-..." ids.
Adding or removing an entry associates/disassociates in place; the group
itself is never replaced. Rules referencing security groups from a
DIFFERENT VPC than the one a packet traverses are ignored by AWS in that
VPC — prefer CIDR/prefix-list rules on multi-VPC groups.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

## Validation Rules

- `ingress_rules_use_source_groups`: ingress rules take source_security_group_ids -- destination_security_group_ids applies only to egress rules
- `egress_rules_use_destination_groups`: egress rules take destination_security_group_ids -- source_security_group_ids applies only to ingress rules

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSecurityGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.security_group_id` | `string` | the unique ID of the security group (sg-...). The join key other resources reference to attach this group. |
| `status.outputs.security_group_arn` | `string` | the ARN of the security group -- the form IAM policy conditions and resource-level permissions expect. |
| `status.outputs.owner_id` | `string` | the AWS account ID that owns the security group. Needed when another account references this group in a cross-account rule ("<owner_id>/<group_id>"). |
| `status.outputs.additional_vpc_association_ids` | `map<string, string>` | association IDs of the group's shares into other VPCs, keyed by the resolved VPC id from spec.additional_vpc_ids (import id form "<group_id>,<vpc_id>"); empty when the group lives in one VPC. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.ingress[].sourceSecurityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.ingress[].destinationSecurityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.egress[].sourceSecurityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.egress[].destinationSecurityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.additionalVpcIds` | AwsVpc | `status.outputs.vpc_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAlb | `spec.securityGroups` | `status.outputs.security_group_id` |
| AwsAppRunnerVpcConnector | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsBatchComputeEnvironment | `spec.computeResources.securityGroupIds` | `status.outputs.security_group_id` |
| AwsBedrockAgentCoreEvaluation | `spec.harnesses[].runtimeEnvironment.network.vpcConfig.securityGroups` | `status.outputs.security_group_id` |
| AwsBedrockAgentCoreEvaluation | `spec.harnesses[].customJwtAuthorizer.privateEndpoint.managedVpc.securityGroupIds` | `status.outputs.security_group_id` |
| AwsBedrockAgentCoreEvaluation | `spec.harnesses[].customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.securityGroupIds` | `status.outputs.security_group_id` |
| AwsBedrockAgentCoreGateway | `spec.customJwtAuthorizer.privateEndpoint.managedVpc.securityGroupIds` | `status.outputs.security_group_id` |
| AwsBedrockAgentCoreGateway | `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.securityGroupIds` | `status.outputs.security_group_id` |
| AwsBedrockAgentCoreGateway | `spec.targets[].privateEndpoint.managedVpc.securityGroupIds` | `status.outputs.security_group_id` |
| AwsBedrockAgentCoreRuntime | `spec.network.vpcConfig.securityGroups` | `status.outputs.security_group_id` |
| AwsBedrockAgentCoreRuntime | `spec.customJwtAuthorizer.privateEndpoint.managedVpc.securityGroupIds` | `status.outputs.security_group_id` |
| AwsBedrockAgentCoreRuntime | `spec.customJwtAuthorizer.privateEndpointOverrides[].privateEndpoint.managedVpc.securityGroupIds` | `status.outputs.security_group_id` |
| AwsBedrockAgentCoreTools | `spec.browsers[].network.vpcConfig.securityGroups` | `status.outputs.security_group_id` |
| AwsBedrockAgentCoreTools | `spec.codeInterpreters[].network.vpcConfig.securityGroups` | `status.outputs.security_group_id` |
| AwsBedrockCustomModel | `spec.vpcConfig.securityGroupIds` | `status.outputs.security_group_id` |
| AwsClientVpn | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsCloudwatchSynthetics | `spec.canary.vpcConfig.securityGroupIds` | `status.outputs.security_group_id` |
| AwsCodeBuildProject | `spec.environment.dockerServer.securityGroupIds` | `status.outputs.security_group_id` |
| AwsCodeBuildProject | `spec.vpcConfig.securityGroupIds` | `status.outputs.security_group_id` |
| AwsDocumentDb | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsEc2Instance | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsEcsCluster | `spec.managedInstancesCapacityProviders[].instanceLaunchTemplate.networkConfiguration.securityGroups` | `status.outputs.security_group_id` |
| AwsEcsService | `spec.network.securityGroups` | `status.outputs.security_group_id` |
| AwsEksCluster | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsEksNodeGroup | `spec.remoteAccess.sourceSecurityGroupIds` | `status.outputs.security_group_id` |
| AwsElasticFileSystem | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsEventBridgePipe | `spec.sourceParameters.selfManagedKafka.vpc.securityGroups` | `status.outputs.security_group_id` |
| AwsEventBridgePipe | `spec.targetParameters.ecsTask.networkConfiguration.securityGroups` | `status.outputs.security_group_id` |
| AwsEventBridgeRule | `spec.targets[].ecsTarget.networkConfiguration.securityGroups` | `status.outputs.security_group_id` |
| AwsEventBridgeScheduler | `spec.target.ecsParameters.networkConfiguration.securityGroups` | `status.outputs.security_group_id` |
| AwsFsxLustreFileSystem | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsFsxOntapFileSystem | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsFsxOpenzfsFileSystem | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsFsxWindowsFileSystem | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsHttpApiVpcLink | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsKinesisFirehose | `spec.opensearch.vpcConfig.securityGroupIds` | `status.outputs.security_group_id` |
| AwsKinesisFirehose | `spec.opensearchServerless.vpcConfig.securityGroupIds` | `status.outputs.security_group_id` |
| AwsLambda | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsLaunchTemplate | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsLaunchTemplate | `spec.networkInterfaces[].securityGroupIds` | `status.outputs.security_group_id` |
| AwsManagedPrometheusScraper | `spec.sourceEks.securityGroupIds` | `status.outputs.security_group_id` |
| AwsManagedPrometheusScraper | `spec.sourceVpc.securityGroupIds` | `status.outputs.security_group_id` |
| AwsMemcachedElasticache | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsMemorydbCluster | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsMskCluster | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsMskServerlessCluster | `spec.vpcConfigs[].securityGroupIds` | `status.outputs.security_group_id` |
| AwsMwaaEnvironment | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsNeptuneCluster | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsNlb | `spec.securityGroups` | `status.outputs.security_group_id` |
| AwsOpenSearchDomain | `spec.vpcOptions.securityGroupIds` | `status.outputs.security_group_id` |
| AwsPlantonRunner | `spec.securityGroups` | `status.outputs.security_group_id` |
| AwsRdsCluster | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsRdsInstance | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsRdsInstance | `spec.options[].vpcSecurityGroupMemberships` | `status.outputs.security_group_id` |
| AwsRdsProxy | `spec.vpcSecurityGroupIds` | `status.outputs.security_group_id` |
| AwsRdsProxy | `spec.endpoints[].vpcSecurityGroupIds` | `status.outputs.security_group_id` |
| AwsRedisElasticache | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsRedshiftCluster | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsRedshiftCluster | `spec.endpointAccesses[].vpcSecurityGroupIds` | `status.outputs.security_group_id` |
| AwsRedshiftServerlessWorkgroup | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsRedshiftServerlessWorkgroup | `spec.endpointAccesses[].vpcSecurityGroupIds` | `status.outputs.security_group_id` |
| AwsRoute53ResolverEndpoint | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsSagemakerDomain | `spec.defaultUserSettings.securityGroupIds` | `status.outputs.security_group_id` |
| AwsSagemakerDomain | `spec.defaultSpaceSettings.securityGroupIds` | `status.outputs.security_group_id` |
| AwsSagemakerDomain | `spec.domainSecurityGroupIds` | `status.outputs.security_group_id` |
| AwsSagemakerDomain | `spec.userProfiles[].userSettings.securityGroupIds` | `status.outputs.security_group_id` |
| AwsSagemakerModel | `spec.vpcConfig.securityGroupIds` | `status.outputs.security_group_id` |
| AwsSagemakerNotebookInstance | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsSecurityGroup | `spec.ingress[].sourceSecurityGroupIds` | `status.outputs.security_group_id` |
| AwsSecurityGroup | `spec.ingress[].destinationSecurityGroupIds` | `status.outputs.security_group_id` |
| AwsSecurityGroup | `spec.egress[].sourceSecurityGroupIds` | `status.outputs.security_group_id` |
| AwsSecurityGroup | `spec.egress[].destinationSecurityGroupIds` | `status.outputs.security_group_id` |
| AwsServerlessElasticache | `spec.securityGroupIds` | `status.outputs.security_group_id` |
| AwsVpcEndpoint | `spec.securityGroupIds` | `status.outputs.security_group_id` |

## See Also

- [Overview](../README.md)
