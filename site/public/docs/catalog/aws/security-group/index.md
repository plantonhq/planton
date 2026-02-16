---
title: "Security Group"
description: "Security Group deployment documentation"
icon: "package"
order: 100
componentName: "awssecuritygroup"
---

# AWS Security Group

Deploys an AWS EC2 Security Group in a specified VPC with configurable ingress and egress rules supporting IPv4 CIDRs, IPv6 CIDRs, source/destination security group references, and self-referencing rules.

## What Gets Created

When you deploy an AwsSecurityGroup resource, OpenMCF provisions:

- **Security Group** — an `ec2.SecurityGroup` resource in the specified VPC with the given name, description, ingress rules, and egress rules
- **Ingress Rules** — inbound traffic rules mapped from the `ingress` field, each specifying protocol, port range, CIDR blocks, security group references, and self-reference settings
- **Egress Rules** — outbound traffic rules mapped from the `egress` field, each specifying protocol, port range, CIDR blocks, destination security group references, and self-reference settings

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An existing VPC** where the Security Group will be created (can be managed by an AwsVpc resource)
- **Knowledge of required ports and protocols** for your workload

## Quick Start

Create a file `sg.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: my-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsSecurityGroup.my-sg
spec:
  vpcId: vpc-0a1b2c3d4e5f00001
  description: Allow HTTP traffic
```

Deploy:

```shell
openmcf apply -f sg.yaml
```

This creates a Security Group in the specified VPC with no ingress rules (all inbound traffic denied) and default AWS egress behavior.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `vpcId` | `StringValueOrRef` | ID of the VPC where the Security Group is created. Can reference an AwsVpc resource via `valueFrom`. | Required |
| `vpcId.value` | `string` | Direct VPC ID value (e.g., `vpc-12345abcde`) | — |
| `vpcId.valueFrom` | `object` | Foreign key reference to an AwsVpc resource | Default kind: `AwsVpc`, default field: `status.outputs.vpc_id` |
| `description` | `string` | Short explanation of the Security Group's purpose. Cannot be modified after creation without replacement. | Required; max 255 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ingress` | `SecurityGroupRule[]` | `[]` | Inbound traffic rules. If empty, all inbound traffic is denied. |
| `egress` | `SecurityGroupRule[]` | `[]` | Outbound traffic rules. If empty, AWS defaults to allow all outbound traffic. |

Each `SecurityGroupRule` contains:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `protocol` | `string` | — | Protocol for the rule: `"tcp"`, `"udp"`, `"icmp"`, or `"-1"` (all protocols). **Required.** |
| `fromPort` | `int32` | `0` | Starting port in the range. For single-port rules, set equal to `toPort`. Use `0` with protocol `-1`. |
| `toPort` | `int32` | `0` | Ending port in the range. For single-port rules, set equal to `fromPort`. Use `0` with protocol `-1`. |
| `ipv4Cidrs` | `string[]` | `[]` | IPv4 CIDR blocks allowed (ingress) or targeted (egress). Example: `"0.0.0.0/0"`. |
| `ipv6Cidrs` | `string[]` | `[]` | IPv6 CIDR blocks allowed or targeted. Example: `"::/0"`. |
| `sourceSecurityGroupIds` | `string[]` | `[]` | Security Group IDs that can send traffic (for ingress rules). |
| `destinationSecurityGroupIds` | `string[]` | `[]` | Security Group IDs that receive traffic (for egress rules). |
| `selfReference` | `bool` | `false` | When `true`, allows traffic from/to instances associated with the same Security Group. |
| `description` | `string` | `""` | Optional description for this specific rule. Max 255 characters. |

## Examples

### Allow HTTP and HTTPS Inbound

A Security Group that allows inbound HTTP (80) and HTTPS (443) from anywhere, with unrestricted outbound:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: web-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsSecurityGroup.web-sg
spec:
  vpcId: vpc-0a1b2c3d4e5f00001
  description: Web tier - HTTP and HTTPS
  ingress:
    - protocol: tcp
      fromPort: 80
      toPort: 80
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: Allow HTTP from anywhere
    - protocol: tcp
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: Allow HTTPS from anywhere
  egress:
    - protocol: "-1"
      fromPort: 0
      toPort: 0
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: Allow all outbound traffic
```

### SSH Access from a Corporate Network

A Security Group that restricts SSH access to a specific CIDR range:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: bastion-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSecurityGroup.bastion-sg
spec:
  vpcId: vpc-0a1b2c3d4e5f00001
  description: Bastion host - SSH from corporate network
  ingress:
    - protocol: tcp
      fromPort: 22
      toPort: 22
      ipv4Cidrs:
        - "203.0.113.0/24"
      description: SSH from corporate CIDR
  egress:
    - protocol: "-1"
      fromPort: 0
      toPort: 0
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: Allow all outbound
```

### Internal Microservice with Self-Referencing

A Security Group for internal services that allows traffic on a custom port from other instances in the same group:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: internal-svc-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSecurityGroup.internal-svc-sg
spec:
  vpcId: vpc-0a1b2c3d4e5f00001
  description: Internal service mesh communication
  ingress:
    - protocol: tcp
      fromPort: 8080
      toPort: 8080
      selfReference: true
      description: Allow port 8080 from same security group
    - protocol: tcp
      fromPort: 8443
      toPort: 8443
      selfReference: true
      description: Allow port 8443 from same security group
  egress:
    - protocol: "-1"
      fromPort: 0
      toPort: 0
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: Allow all outbound
```

### Database Tier with Source Security Group

A Security Group for a database that only accepts traffic from a specific application Security Group:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: db-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSecurityGroup.db-sg
spec:
  vpcId: vpc-0a1b2c3d4e5f00001
  description: Database tier - PostgreSQL from app tier only
  ingress:
    - protocol: tcp
      fromPort: 5432
      toPort: 5432
      sourceSecurityGroupIds:
        - sg-0a1b2c3d4e5f00099
      description: PostgreSQL from app security group
  egress:
    - protocol: "-1"
      fromPort: 0
      toPort: 0
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: Allow all outbound
```

### Using Foreign Key References

Reference an OpenMCF-managed VPC instead of hardcoding the VPC ID:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecurityGroup
metadata:
  name: ref-sg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSecurityGroup.ref-sg
spec:
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: my-vpc
      fieldPath: status.outputs.vpc_id
  description: ALB security group in managed VPC
  ingress:
    - protocol: tcp
      fromPort: 80
      toPort: 80
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: Allow HTTP
    - protocol: tcp
      fromPort: 443
      toPort: 443
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: Allow HTTPS
  egress:
    - protocol: "-1"
      fromPort: 0
      toPort: 0
      ipv4Cidrs:
        - "0.0.0.0/0"
      description: Allow all outbound
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `security_group_id` | `string` | The unique ID of the created Security Group |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides the VPC where Security Groups are created
- [AwsAlb](/docs/catalog/aws/alb) — attaches Security Groups to Application Load Balancers
- [AwsEksCluster](/docs/catalog/aws/eks-cluster) — uses Security Groups for cluster and node networking
- [AwsEc2Instance](/docs/catalog/aws/ec2-instance) — assigns Security Groups to EC2 instances
- [AwsRdsInstance](/docs/catalog/aws/rds-instance) — uses Security Groups to control database access
