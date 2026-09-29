# On-Demand General Pool

This preset runs the workhorse node pool of a typical cluster: On-Demand
AL2023 nodes across two availability zones, surge-enabled version
rollouts, and managed node auto-repair. Everything composes by
reference -- the cluster, the node role, and the subnets are all
first-class resources.

## When to Use

- The default compute pool for stateless and stateful services
- Clusters that need predictable capacity (no Spot interruptions)

## Key Configuration Choices

- **`amiType: AL2023_x86_64_STANDARD`** -- the current-generation
  EKS-optimized Amazon Linux family
- **`updateStrategy: MINIMAL` + `maxUnavailablePercentage: 25`** --
  version updates launch replacements before terminating, so capacity
  never dips; a quarter of the pool rolls at a time
- **`nodeRepairConfig.enabled: true`** -- EKS replaces or reboots nodes
  the cluster reports unhealthy, without operator intervention
- **The node role carries its own policies** -- attach
  `AmazonEKSWorkerNodePolicy`, `AmazonEC2ContainerRegistryReadOnly`, and
  `AmazonEKS_CNI_Policy` on the referenced `AwsIamRole`; the node group
  never modifies a role it references

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<node-group-name>` | Name for the pool | Your naming convention (e.g., `general`) |
| `<aws-region>` | AWS region code (e.g., `us-west-2`) | Your deployment region |
| `<cluster-resource-name>` | Name of the AwsEksCluster resource | Your cluster manifest's `metadata.name` |
| `<node-role-resource-name>` | Name of the AwsIamRole with the worker policies | Your role manifest's `metadata.name` |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |

## Common Additions

- Pin `version` during a control-plane upgrade, then bump it to roll the
  nodes on your schedule
- Add `labels` (e.g., `pool: general`) so workloads can target the pool
  with nodeSelectors

## Related Presets

- **02-spot-cost-optimized** -- interruptible capacity at steep discount
- **03-launch-template** -- custom AMI/IMDSv2/encrypted-volume mechanics
  from an AwsLaunchTemplate
