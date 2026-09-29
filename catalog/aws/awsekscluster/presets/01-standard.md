# Standard EKS Cluster

This preset creates a two-AZ EKS control plane with the modern
access-entries authentication model, audit-grade control-plane logging,
and standard upgrade support. The API endpoint stays publicly reachable
(the AWS default), which is what most teams want for `kubectl` access
from developer machines and CI/CD.

## When to Use

- Standard production Kubernetes clusters where the API server needs to
  be reachable from outside the VPC
- Clusters that will have `AwsEksNodeGroup` compute attached
- Teams starting fresh (no legacy aws-auth ConfigMap to migrate)

## Key Configuration Choices

- **`authenticationMode: API`** -- access entries only, the modern model;
  IAM principals are granted cluster access as first-class EKS resources
  instead of edits to a ConfigMap
- **`enabledClusterLogTypes: [audit, authenticator]`** -- the two
  highest-value control-plane logs (who did what, who got in) without the
  CloudWatch ingestion cost of all five types
- **`upgradeSupportType: STANDARD`** -- upgrade on the normal schedule
  and never pay the extended-support surcharge
- **Version left unset** -- AWS picks its current default Kubernetes
  version; pin `version` when you need a specific minor
- **The cluster role carries its own policy** -- attach
  `AmazonEKSClusterPolicy` on the referenced `AwsIamRole`
  (`managedPolicyArns`); the cluster never modifies a role it references

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<cluster-name>` | Name for the cluster | Your environment naming convention |
| `<aws-region>` | AWS region code (e.g., `us-west-2`) | Your deployment region |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |
| `<cluster-role-resource-name>` | Name of the AwsIamRole with `AmazonEKSClusterPolicy` | Your role manifest's `metadata.name` |

## Common Additions

- Restrict the public endpoint with `publicAccessCidrs` (office/VPN
  egress ranges) -- the cheapest hardening step for a public cluster
- Add `kmsKeyArn` (an `AwsKmsKey` reference) for customer-managed
  envelope encryption of Kubernetes secrets -- decide at creation; it
  cannot be disabled later
- Add `deletionProtection: true` on shared/production clusters

## Related Presets

- **02-private-endpoint** -- API server reachable only inside the VPC
- **03-auto-mode** -- AWS manages compute/storage/load balancing itself
