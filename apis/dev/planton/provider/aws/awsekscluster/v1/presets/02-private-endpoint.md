# Private Endpoint EKS Cluster

This preset creates an EKS cluster with the API server endpoint restricted to VPC-internal access only. The Kubernetes API is not reachable from the public internet. Use this for security-sensitive environments where all cluster access must go through a VPN, bastion host, or AWS PrivateLink.

## When to Use

- Security-sensitive environments that require the Kubernetes API to be inaccessible from the internet
- Compliance requirements (PCI-DSS, HIPAA) mandating private-only control plane access
- Teams that access clusters exclusively through VPN or bastion hosts

## Key Configuration Choices

- **Private API endpoint** (`disablePublicEndpoint: true`) -- API server is only accessible from within the VPC or peered networks
- **Control plane logs enabled** -- Essential for auditing when the cluster is not publicly accessible
- **No public access CIDRs** -- Not needed since the public endpoint is disabled entirely

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<eks-cluster-role-arn>` | IAM role ARN with `AmazonEKSClusterPolicy` attached | AWS IAM console or `AwsIamRole` status outputs |

## Related Presets

- **01-standard** -- Use instead when the API server needs to be accessible from outside the VPC
