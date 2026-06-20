---
title: "EC2 Instance"
description: "EC2 Instance deployment documentation"
icon: "package"
order: 100
componentName: "awsec2instance"
---

# AWS EC2 Instance

Deploys a single AWS EC2 virtual machine instance in a private subnet with configurable networking, IAM, and access method. The component supports three connection methods (SSM, Bastion SSH, and EC2 Instance Connect) and auto-generates an SSH key pair when needed.

## What Gets Created

When you deploy an AwsEc2Instance resource, OpenMCF provisions:

- **EC2 Instance** — an `aws:ec2:Instance` resource launched with the specified AMI, instance type, subnet, and security groups, with a configurable root EBS volume
- **TLS Private Key** — a `tls:PrivateKey` RSA-4096 key pair, created only when `connectionMethod` is `BASTION` or `INSTANCE_CONNECT` and no `keyName` is provided
- **EC2 Key Pair** — an `aws:ec2:KeyPair` resource registered in AWS from the generated public key, created only alongside the TLS private key above

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A VPC with at least one private subnet** where the instance will be placed
- **At least one security group** controlling inbound/outbound traffic for the instance
- **An AMI ID** for the desired operating system (e.g., Amazon Linux, Ubuntu)
- **An IAM instance profile ARN** if using SSM as the connection method
- **An EC2 key pair name** if using BASTION or INSTANCE_CONNECT and you want to supply your own key (otherwise one is auto-generated)

## Quick Start

Create a file `ec2-instance.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: my-instance
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEc2Instance.my-instance
spec:
  region: us-west-2
  instanceName: my-instance
  amiId: ami-0abcdef1234567890
  instanceType: t3.small
  subnetId: subnet-0a1b2c3d4e5f00001
  securityGroupIds:
    - sg-0a1b2c3d4e5f00001
  iamInstanceProfileArn: arn:aws:iam::123456789012:instance-profile/ssm-profile
```

Deploy:

```shell
openmcf apply -f ec2-instance.yaml
```

This creates an EC2 instance in a private subnet using SSM (the default connection method) for shell access.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the resource will be created (e.g., `us-west-2`) | Must be a valid AWS region |
| `instanceName` | `string` | Name tag for the EC2 instance | Min length 1 |
| `amiId` | `string` | Amazon Machine Image ID (e.g., `ami-0abcdef1234567890`) | Must start with `ami-` |
| `instanceType` | `string` | EC2 instance type determining vCPU and memory (e.g., `t3.small`, `m5.large`) | Min length 1 |
| `subnetId` | `StringValueOrRef` | Subnet ID where the instance will reside. Can reference an AwsVpc resource via `valueFrom`. | Required |
| `securityGroupIds` | `StringValueOrRef[]` | Security group IDs to attach to the instance. Can reference AwsSecurityGroup resources via `valueFrom`. | Minimum 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `connectionMethod` | `enum` | `SSM` | How to connect to the instance. Valid values: `SSM`, `BASTION`, `INSTANCE_CONNECT`. |
| `iamInstanceProfileArn` | `StringValueOrRef` | — | ARN of an IAM instance profile to attach. Required when `connectionMethod` is `SSM`. Can reference an AwsIamRole resource via `valueFrom`. |
| `keyName` | `string` | — | Name of an existing EC2 key pair for SSH access. Required when `connectionMethod` is `BASTION` or `INSTANCE_CONNECT` (if omitted, a key pair is auto-generated). |
| `rootVolumeSizeGb` | `int32` | `30` | Size of the root EBS volume in GiB. Must be greater than 0. |
| `tags` | `map<string, string>` | `{}` | Custom tags to apply to the EC2 instance. Merged with OpenMCF-managed tags (user tags take precedence on collision). |
| `userData` | `string` | — | Cloud-init or shell script to run on first boot. Maximum 32 KiB. |
| `ebsOptimized` | `bool` | `false` | Enables dedicated EBS throughput for I/O-intensive workloads. |
| `disableApiTermination` | `bool` | `false` | Prevents accidental instance termination via the AWS API. |

## Examples

### SSM-Connected Instance in a Private Subnet

The default configuration for a private instance managed via AWS Systems Manager:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: backend-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEc2Instance.backend-server
spec:
  region: us-west-2
  instanceName: backend-server
  amiId: ami-0abcdef1234567890
  instanceType: t3.medium
  subnetId: subnet-private-az1
  securityGroupIds:
    - sg-app-servers
  iamInstanceProfileArn: arn:aws:iam::123456789012:instance-profile/ssm-access
```

### Bastion Host with SSH Key Pair

An instance configured for direct SSH access via a pre-existing key pair:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: bastion
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEc2Instance.bastion
spec:
  region: us-west-2
  instanceName: bastion
  amiId: ami-0abcdef1234567890
  instanceType: t3.micro
  subnetId: subnet-public-az1
  securityGroupIds:
    - sg-bastion
  connectionMethod: BASTION
  keyName: my-existing-keypair
  rootVolumeSizeGb: 20
```

### Auto-Generated Key Pair with Instance Connect

When no `keyName` is provided for INSTANCE_CONNECT, OpenMCF generates an RSA-4096 key pair automatically and exposes the private key as a stack output:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: dev-box
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEc2Instance.dev-box
spec:
  region: us-west-2
  instanceName: dev-box
  amiId: ami-0abcdef1234567890
  instanceType: m5.large
  subnetId: subnet-private-az1
  securityGroupIds:
    - sg-dev-machines
  connectionMethod: INSTANCE_CONNECT
  rootVolumeSizeGb: 100
  ebsOptimized: true
  userData: |
    #!/bin/bash
    yum update -y
    yum install -y docker
    systemctl start docker
```

### Production Instance with Termination Protection

A hardened instance with custom tags, EBS optimization, and API termination protection:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: prod-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEc2Instance.prod-api
spec:
  region: us-west-2
  instanceName: prod-api
  amiId: ami-0abcdef1234567890
  instanceType: c5.xlarge
  subnetId: subnet-private-az1
  securityGroupIds:
    - sg-prod-api
    - sg-monitoring
  connectionMethod: SSM
  iamInstanceProfileArn: arn:aws:iam::123456789012:instance-profile/prod-ssm
  rootVolumeSizeGb: 200
  ebsOptimized: true
  disableApiTermination: true
  tags:
    env: production
    team: platform
    cost-center: eng-infra
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEc2Instance
metadata:
  name: ref-instance
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEc2Instance.ref-instance
spec:
  region: us-west-2
  instanceName: ref-instance
  amiId: ami-0abcdef1234567890
  instanceType: t3.small
  subnetId:
    valueFrom:
      kind: AwsSubnet
      name: my-private-subnet
      fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: app-sg
        field: status.outputs.security_group_id
  connectionMethod: SSM
  iamInstanceProfileArn:
    valueFrom:
      kind: AwsIamRole
      name: ssm-role
      field: status.outputs.role_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | `string` | Unique EC2 instance identifier (e.g., `i-0123456789abcdef0`) |
| `private_ip` | `string` | Primary private IPv4 address assigned to the instance |
| `private_dns_name` | `string` | Internal DNS hostname within the VPC |
| `availability_zone` | `string` | Availability Zone where the instance is running (e.g., `us-west-2a`) |
| `instance_profile_arn` | `string` | ARN of the IAM instance profile attached to the instance (if any) |
| `ssh_private_key` | `string` | Base64-encoded PEM private key (only present when a key pair was auto-generated) |
| `ssh_public_key` | `string` | OpenSSH-formatted public key (only present when a key pair was auto-generated) |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides the subnet where the instance is placed
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls inbound and outbound traffic for the instance
- [AwsIamRole](/docs/catalog/aws/iam-role) — provides the IAM instance profile for SSM or other AWS API access
- [AwsAlb](/docs/catalog/aws/alb) — routes traffic to the instance via an Application Load Balancer
