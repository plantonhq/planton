# Private-Endpoint EKS Cluster

This preset creates a hardened, fully private EKS control plane: the API
server is reachable only from inside the VPC, Kubernetes secrets are
envelope-encrypted with a customer-managed KMS key, and deletion
protection guards the control plane. The configuration for regulated
environments and security-first platform teams.

## When to Use

- Clusters whose API server must never be internet-reachable
  (compliance, internal platforms, regulated workloads)
- Environments where operators reach the cluster through a VPN, bastion,
  or VPC-connected CI runners

## Key Configuration Choices

- **`endpointPublicAccess: false` + `endpointPrivateAccess: true`** --
  the pair that makes a cluster private; disabling public without
  enabling private would leave the API server unreachable
- **`kmsKeyArn`** -- customer-managed envelope encryption for secrets in
  etcd; a one-way door (cannot be disabled or re-keyed on a live
  cluster), so it belongs in the manifest from day one
- **`deletionProtection: true`** -- the EKS API refuses to delete the
  cluster until this is explicitly turned off
- **`authenticationMode: API`** -- access entries only; no legacy
  ConfigMap surface

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<cluster-name>` | Name for the cluster | Your environment naming convention |
| `<aws-region>` | AWS region code (e.g., `us-west-2`) | Your deployment region |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |
| `<cluster-role-resource-name>` | Name of the AwsIamRole with `AmazonEKSClusterPolicy` | Your role manifest's `metadata.name` |
| `<kms-key-resource-name>` | Name of the AwsKmsKey for secrets encryption | Your KMS key manifest's `metadata.name` |

## Common Additions

- `controlPlaneEgressMode: CUSTOMER_ROUTED` to route control-plane
  egress through your inspection/firewall VPC layout (one-way: reverting
  to AWS_MANAGED replaces the cluster)
- `serviceIpv4Cidr` when the default service ranges would collide with
  peered networks -- create-only, so decide before the first apply

## Related Presets

- **01-standard** -- public endpoint for developer-facing clusters
- **03-auto-mode** -- AWS manages compute/storage/load balancing itself
