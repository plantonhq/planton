# Standard EKS Cluster

This preset creates an EKS cluster with a publicly accessible API endpoint and control plane logging enabled. The cluster spans two Availability Zones for high availability. This is the most common EKS configuration, suitable for most production workloads where `kubectl` access from developer machines and CI/CD pipelines is needed.

## When to Use

- Standard production Kubernetes clusters where the API server needs to be accessible from outside the VPC
- Teams that access `kubectl` from local machines, CI/CD systems, or VPNs
- Clusters that will later have node groups attached via `AwsEksNodeGroup`

## Key Configuration Choices

- **Public API endpoint** (`disablePublicEndpoint: false`) -- API server is accessible from the internet; restrict with `publicAccessCidrs` if needed
- **Control plane logs enabled** (`enableControlPlaneLogs: true`) -- API server, audit, authenticator, controller manager, and scheduler logs sent to CloudWatch
- **Two subnets across AZs** -- Minimum required for EKS high availability
- **Kubernetes 1.30** (`version: "1.30"`) -- Update to the latest supported version for your environment

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<eks-cluster-role-arn>` | IAM role ARN with `AmazonEKSClusterPolicy` attached | AWS IAM console or `AwsIamRole` status outputs |

## Related Presets

- **02-private-endpoint** -- Use instead when the API server should only be accessible within the VPC
