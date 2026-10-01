# EKS Auto Mode Cluster

This preset creates a hands-off Kubernetes platform: EKS Auto Mode
provisions and scales EC2 capacity, provisions EBS volumes, and manages
load balancers for the cluster's workloads. There are no node groups to
size, patch, or roll -- AWS operates the data plane.

## When to Use

- Teams that want Kubernetes without fleet operations (no AMI rollouts,
  no capacity planning, no node patching)
- New clusters with no existing node-group tooling to preserve
- Workloads with spiky or unpredictable capacity needs

## Key Configuration Choices

- **`autoMode.enabled: true`** -- one toggle for AWS's three underlying
  capabilities (compute, block storage, load balancing); the EKS API
  requires them enabled or disabled together, so the spec models them
  honestly as a unit
- **`nodePools: [general-purpose, system]`** -- the built-in pools;
  `system` isolates cluster-critical pods on dedicated capacity. Omit
  both to drive capacity purely from in-cluster NodePool resources
- **`bootstrapSelfManagedAddons: false`** -- Auto Mode runs networking/
  storage/LB itself, so the default add-on bootstrap is unnecessary
  (create-only; decide before the first apply)
- **Auto Mode replaces node groups** -- do not attach `AwsEksNodeGroup`
  resources to this cluster; the two compute models are alternatives

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<cluster-name>` | Name for the cluster | Your environment naming convention |
| `<aws-region>` | AWS region code (e.g., `us-west-2`) | Your deployment region |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |
| `<cluster-role-resource-name>` | Name of the AwsIamRole with `AmazonEKSClusterPolicy` | Your role manifest's `metadata.name` |
| `<node-role-resource-name>` | Name of the AwsIamRole Auto Mode nodes assume | Your role manifest's `metadata.name` |

## Common Additions

- `endpointPublicAccess: false` + `endpointPrivateAccess: true` for a
  private Auto Mode cluster
- `kmsKeyArn` for customer-managed secrets encryption

## Related Presets

- **01-standard** -- explicit `AwsEksNodeGroup` compute instead
- **02-private-endpoint** -- the hardened private control plane
