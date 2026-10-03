# AWS Security Group

Deploys an EC2 Security Group within a specified VPC with configurable ingress and egress rules supporting IPv4/IPv6 CIDRs, cross-group references, and self-referencing. Security groups are stateful -- return traffic for an allowed connection is automatically permitted, so rules only describe the initiating direction. One posture choice to know up front: when `egress` is empty, the module revokes AWS's default allow-all outbound rule, making the manifest the complete statement of what the group permits. The group can also be shared into additional VPCs in the same account and region, so one firewall definition serves many networks.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Security Group** -- an EC2 security group in the specified VPC with a description and name matching your manifest's `metadata.name`
- **Ingress Rules** -- one rule per entry in the `ingress` array, supporting protocol, port ranges, IPv4/IPv6 CIDRs, source security group references, and self-referencing
- **Egress Rules** -- one rule per entry in the `egress` array, supporting protocol, port ranges, IPv4/IPv6 CIDRs, destination security group references, and self-referencing
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the security group

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A VPC** -- the security group must belong to exactly one VPC. Provide the VPC ID directly or reference an AwsVpc Infra Component via ValueFromRef.
- **Source/destination security groups** (optional) -- if rules reference other security groups, those groups must exist in the same VPC. Provide their IDs directly or reference other AwsSecurityGroup Infra Components via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **AWS Security Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Web Tier Security Group** preset in the [Presets](#presets) tab to pre-populate a common configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSecurityGroup
metadata:
  name: web-tier-sg
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  vpcId:
    value: "vpc-0a1b2c3d4e5f00001"
  description: "Allow HTTP/HTTPS from the internet"
  ingress:
    - protocol: tcp
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "HTTPS from anywhere"
  egress:
    - protocol: "-1"
      fromPort: 0
      toPort: 0
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: "All outbound traffic"
```

```shell
planton apply -f security-group.yaml
```

This creates a security group allowing inbound HTTPS from all sources and unrestricted outbound traffic. No source or destination security group references are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the security group to a VPC deployed in the same InfraPipeline:

```yaml
spec:
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: production-vpc
      fieldPath: status.outputs.vpc_id
```

The InfraPipeline resolves the dependency graph, deploys the VPC first, then provisions the security group with the resolved VPC ID.

## Key Configuration

These are the most important decisions when configuring a security group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Ingress rules** -- Each entry in `ingress` defines an inbound traffic rule. Combine `protocol`, `fromPort`, and `toPort` with `ipv4Cidrs`/`ipv6Cidrs` for CIDR-based access, `sourceSecurityGroupIds` for group-based access, or `prefixListIds` for managed prefix lists. Use `selfReference: true` for intra-group traffic (e.g., cluster nodes communicating with each other).

**Egress rules** -- If `egress` is empty, ALL outbound traffic is denied: the module revokes the allow-all egress rule AWS adds to every new group, so the manifest is the complete statement of what the group permits. Add an explicit all-traffic egress rule (`protocol: "-1"`, `0.0.0.0/0`) to restore the AWS default behavior.

**Security group references** -- The `sourceSecurityGroupIds` (ingress) and `destinationSecurityGroupIds` (egress) fields accept ValueFromRef references to other AwsSecurityGroup resources. This enables declarative, dependency-tracked security policies across tiers (web -> app -> database) without hardcoding security group IDs.

**Managed prefix lists** -- The per-rule `prefixListIds` field names a set of CIDRs by a stable ID (`pl-...`). AWS-managed lists cover services like S3 and DynamoDB gateway endpoints, so an egress rule can target "the S3 service" instead of hardcoding its CIDRs; customer-managed lists let network teams maintain shared CIDR sets (office ranges, partner networks) that many groups reference without copying.

**Description immutability** -- The `description` field cannot be modified after creation without replacing the security group. Choose a descriptive value upfront, such as the tier name and purpose.

**Deletion posture** -- `revokeRulesOnDelete` is off by default: deleting a group that other groups' rules still reference fails with a DependencyViolation, surfacing the dependency. Enable it for groups referenced cross-group in environments that are torn down whole, so teardown never requires manual rule surgery. Safe to toggle in place.

**Multi-VPC sharing** -- `additionalVpcIds` shares the group into other VPCs in the same account and region, so workloads there attach the same group instead of maintaining drifting per-VPC copies. Associations add and remove in place. One rule of thumb: AWS ignores a rule that references a security group from a different VPC than the one a packet traverses, so shared groups should carry CIDR or prefix-list rules, which behave identically everywhere.

**Rule direction is validated** -- `sourceSecurityGroupIds` belongs to ingress rules and `destinationSecurityGroupIds` to egress rules; a value in the wrong direction would be silently dropped by the module, so validation rejects it up front.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsVpc** | `vpcId` | `status.outputs.vpc_id` |
| **AwsVpc** (optional) | `additionalVpcIds` | `status.outputs.vpc_id` |
| **AwsSecurityGroup** (optional) | `ingress[].sourceSecurityGroupIds` | `status.outputs.security_group_id` |
| **AwsSecurityGroup** (optional) | `egress[].destinationSecurityGroupIds` | `status.outputs.security_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `security_group_id` | ID of the created security group | EKS cluster security, RDS instance access, EC2 instance network interfaces |
| `security_group_arn` | Amazon Resource Name of the group | IAM policies and support tooling |
| `owner_id` | The AWS account that owns the group | Cross-account rule references (account/group-id pairs) |
| `additional_vpc_association_ids` | Association ids of the group's multi-VPC shares, keyed by VPC id | Attachment evidence, import tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Web tier** -- Allows inbound HTTP (80) and HTTPS (443) from the internet with unrestricted outbound. The standard pattern for public-facing load balancers and web servers. Start from the **Web Tier Security Group** preset.

**Database tier** -- Restricts inbound access to specific ports (e.g., 5432 for PostgreSQL, 3306 for MySQL) from application-tier security groups only. No public CIDR access. Start from the **Database Tier Security Group** preset.

**Bastion host** -- Allows inbound SSH (22) from a restricted set of administrator IP addresses and full outbound access for management tasks. Start from the **Bastion Host Security Group** preset.

**Shared multi-VPC group** -- One baseline firewall definition associated with several VPCs, attached by workloads everywhere it is shared. Start from the **Shared Multi-VPC Security Group** preset.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) -- provides the VPC where the security group is created
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- provides source or destination security group IDs for cross-group rule references
