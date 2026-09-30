# Namespace on Fargate

This preset sends every pod in one namespace to Fargate -- the simplest
serverless slice of a cluster, with the pod execution role and private
subnets attached by reference.

## When to Use

- Isolating a workload class (batch jobs, untrusted tenants, spiky
  services) from the EC2 node fleet
- Running a namespace serverless without sizing or scaling any node
  group for it
- The first Fargate footprint on a cluster that otherwise runs node
  groups

## Key Configuration Choices

- **One namespace selector** -- everything in the namespace matches;
  add `labels` to the selector to narrow it to labeled pods only
- **Private subnets by reference** -- AWS rejects internet-gateway
  subnets for Fargate; image pulls need a NAT gateway (or VPC
  endpoints) on the subnets' route
- **The role carries its own policy** -- trust
  `eks-fargate-pods.amazonaws.com` and attach
  `AmazonEKSFargatePodExecutionRolePolicy` on the referenced
  `AwsIamRole`; the profile never modifies it

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<profile-resource-name>` | Name for the profile (AWS limit 63 chars) | Your naming convention |
| `<aws-region>` | AWS region code (e.g., `us-west-2`) | Your deployment region |
| `<cluster-resource-name>` | Name of the AwsEksCluster resource | Your cluster manifest's `metadata.name` |
| `<pod-execution-role-resource-name>` | Name of the AwsIamRole for Fargate pod execution | Your role manifest's `metadata.name` |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two private AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |
| `<namespace>` | The Kubernetes namespace to run serverless | Your cluster's namespace layout |

## Common Additions

- More selectors (up to 5) for additional namespaces
- Wildcards (`team-*`) to cover namespace-per-team layouts with one
  selector

## Related Presets

- **02-labeled-workloads** -- label-scoped selection inside shared
  namespaces
