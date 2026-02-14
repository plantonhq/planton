# Spot Cost-Optimized Node Group

This preset creates an EKS managed node group using Spot instances for up to 70% cost savings compared to on-demand. The `node-lifecycle: spot` label enables workload targeting via node selectors or tolerations, so only fault-tolerant workloads land on these nodes.

## When to Use

- Stateless or fault-tolerant workloads that can handle Spot interruptions (web servers, batch jobs, CI/CD runners)
- Cost-sensitive environments where reducing compute spend is a priority
- Mixed node group strategies: pair with an on-demand node group for critical workloads

## Key Configuration Choices

- **Spot instances** (`capacityType: spot`) -- Up to 70% cheaper than on-demand; AWS may reclaim instances with 2-minute notice
- **t3.large** (`instanceType`) -- 2 vCPUs, 8 GiB RAM; larger than the on-demand preset to provide headroom for Spot availability
- **2-10 nodes** (`scaling`) -- Wider scaling range to accommodate Spot capacity fluctuations
- **Spot label** (`labels: {node-lifecycle: spot}`) -- Enables Kubernetes scheduling decisions based on node lifecycle type

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<eks-cluster-name>` | Name of the EKS cluster | `AwsEksCluster` metadata name |
| `<node-role-arn>` | IAM role ARN with EKS worker node policies | AWS IAM console or `AwsIamRole` status outputs |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |

## Related Presets

- **01-on-demand-general** -- Use instead for workloads that cannot tolerate Spot interruptions
