# On-Demand General Purpose Node Group

This preset creates an EKS managed node group using on-demand `t3.medium` instances across two Availability Zones. The group scales between 2 and 5 nodes with 100 GiB root disks. This is the standard starting point for most Kubernetes workloads that need predictable compute capacity without Spot interruptions.

## When to Use

- General-purpose Kubernetes workloads (web servers, API services, background workers)
- Production clusters where Spot interruptions are not acceptable
- Starting point before right-sizing to larger or more specialized instance types

## Key Configuration Choices

- **On-demand instances** (`capacityType: on_demand`) -- No Spot interruptions; predictable availability
- **t3.medium** (`instanceType`) -- 2 vCPUs, 4 GiB RAM; burstable performance suitable for most workloads
- **2-5 nodes** (`scaling`) -- Starts with 2 nodes for HA; scales to 5 under load via Cluster Autoscaler or Karpenter
- **100 GiB disk** (`diskSizeGb: 100`) -- Sufficient for container images, logs, and ephemeral storage
- **Multi-AZ** -- Nodes span two AZs for high availability

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<eks-cluster-name>` | Name of the EKS cluster to attach this node group to | `AwsEksCluster` metadata name |
| `<node-role-arn>` | IAM role ARN with `AmazonEKSWorkerNodePolicy`, `AmazonEKS_CNI_Policy`, and `AmazonEC2ContainerRegistryReadOnly` | AWS IAM console or `AwsIamRole` status outputs |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |

## Related Presets

- **02-spot-cost-optimized** -- Use instead to reduce costs by running nodes on Spot instances (suitable for fault-tolerant workloads)
