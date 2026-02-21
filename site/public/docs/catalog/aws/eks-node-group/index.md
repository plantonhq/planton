---
title: "EKS Node Group"
description: "EKS Node Group deployment documentation"
icon: "package"
order: 100
componentName: "awseksnodegroup"
---

# AWS EKS Node Group

Deploys an AWS EKS managed node group into an existing EKS cluster, provisioning EC2 worker nodes with configurable instance types, auto-scaling, and optional SSH access.

## What Gets Created

When you deploy an AwsEksNodeGroup resource, OpenMCF provisions:

- **EKS Managed Node Group** — an `aws_eks_node_group` resource attached to the specified EKS cluster, running EC2 instances in the provided subnets with the configured scaling parameters, instance type, capacity type, and disk size
- **Auto Scaling Group** — AWS automatically creates and manages an ASG behind the node group to enforce the min/max/desired node counts
- **Remote Access Configuration** — created only when `sshKeyName` is provided, configures the EC2 Key Pair on the nodes to allow SSH access

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An existing EKS cluster** (e.g., created by an AwsEksCluster resource)
- **An IAM role** with the required EKS worker node policies (`AmazonEKSWorkerNodePolicy`, `AmazonEKS_CNI_Policy`, `AmazonEC2ContainerRegistryReadOnly`)
- **At least two subnets** in different Availability Zones (typically private subnets in the cluster's VPC)

## Quick Start

Create a file `eks-nodegroup.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: my-nodegroup
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEksNodeGroup.my-nodegroup
spec:
  region: us-west-2
  clusterName: my-eks-cluster
  nodeRoleArn: arn:aws:iam::123456789012:role/eks-node-role
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  instanceType: t3.medium
  scaling:
    minSize: 1
    maxSize: 3
    desiredSize: 2
```

Deploy:

```shell
openmcf apply -f eks-nodegroup.yaml
```

This creates a managed node group with two `t3.medium` on-demand instances in the specified EKS cluster.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the node group will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `clusterName` | `StringValueOrRef` | Name of the EKS cluster to attach this node group to. Can reference an AwsEksCluster resource via `valueFrom`. | Required |
| `nodeRoleArn` | `StringValueOrRef` | ARN of the IAM role for the EC2 worker nodes. Must have EKS worker node policies. Can reference an AwsIamRole resource via `valueFrom`. | Required |
| `subnetIds` | `StringValueOrRef[]` | Subnet IDs where worker nodes are launched. Typically private subnets across multiple AZs. Can reference an AwsVpc resource via `valueFrom`. | Minimum 2 items |
| `instanceType` | `string` | EC2 instance type for the worker nodes (e.g., `t3.medium`, `m5.xlarge`). | Required |
| `scaling` | `object` | Auto-scaling configuration for the node group. | Required |
| `scaling.minSize` | `int32` | Minimum number of nodes in the group. | >= 1 |
| `scaling.maxSize` | `int32` | Maximum number of nodes allowed in the group. | >= 1 |
| `scaling.desiredSize` | `int32` | Initial target number of nodes. Should be between `minSize` and `maxSize`. | >= 1 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `capacityType` | `enum` | `on_demand` | Instance purchasing model. Valid values: `on_demand`, `spot`. |
| `diskSizeGb` | `int32` | `100` | EBS root volume size in GiB for each node. |
| `sshKeyName` | `string` | — | Name of an existing EC2 Key Pair to enable SSH access to the nodes. Max 255 characters. |
| `labels` | `map<string, string>` | `{}` | Kubernetes labels applied to the node group and its nodes. Keys and values max 63 characters each. |

## Examples

### Spot Instance Node Group

Use Spot instances for cost savings on fault-tolerant workloads:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: spot-nodegroup
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEksNodeGroup.spot-nodegroup
spec:
  region: us-west-2
  clusterName: my-eks-cluster
  nodeRoleArn: arn:aws:iam::123456789012:role/eks-node-role
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  instanceType: m5.large
  scaling:
    minSize: 2
    maxSize: 10
    desiredSize: 4
  capacityType: spot
```

### Node Group with SSH Access and Labels

Enable SSH for debugging and add Kubernetes labels for workload scheduling:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: labeled-nodegroup
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AwsEksNodeGroup.labeled-nodegroup
spec:
  region: us-west-2
  clusterName: my-eks-cluster
  nodeRoleArn: arn:aws:iam::123456789012:role/eks-node-role
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  instanceType: c5.xlarge
  scaling:
    minSize: 3
    maxSize: 6
    desiredSize: 3
  diskSizeGb: 200
  sshKeyName: ops-keypair
  labels:
    team: data-platform
    workload: batch
```

### Production Node Group with Large Disks

High-capacity node group for production workloads with large container images:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: prod-nodegroup
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEksNodeGroup.prod-nodegroup
spec:
  region: us-west-2
  clusterName: prod-eks-cluster
  nodeRoleArn: arn:aws:iam::123456789012:role/prod-eks-node-role
  subnetIds:
    - subnet-prod-az1
    - subnet-prod-az2
    - subnet-prod-az3
  instanceType: m5.2xlarge
  scaling:
    minSize: 3
    maxSize: 20
    desiredSize: 6
  capacityType: on_demand
  diskSizeGb: 500
  labels:
    environment: production
    tier: compute
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEksNodeGroup
metadata:
  name: ref-nodegroup
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEksNodeGroup.ref-nodegroup
spec:
  region: us-west-2
  clusterName:
    valueFrom:
      kind: AwsEksCluster
      name: my-cluster
      field: metadata.name
  nodeRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: eks-node-role
      field: status.outputs.role_arn
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets[0].id
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        field: status.outputs.private_subnets[1].id
  instanceType: t3.large
  scaling:
    minSize: 2
    maxSize: 8
    desiredSize: 3
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `nodegroup_name` | `string` | Name of the created EKS managed node group |
| `asg_name` | `string` | Name of the underlying AWS Auto Scaling Group managing the nodes |
| `remote_access_sg_id` | `string` | ID of the security group for SSH access (present only when `sshKeyName` is set) |
| `instance_profile_arn` | `string` | ARN of the EC2 instance profile associated with the nodes |

## Related Components

- [AwsEksCluster](/docs/catalog/aws/eks-cluster) — provides the EKS cluster that this node group attaches to
- [AwsIamRole](/docs/catalog/aws/iam-role) — supplies the IAM role assumed by the worker nodes
- [AwsVpc](/docs/catalog/aws/vpc) — provides the subnets for node placement
