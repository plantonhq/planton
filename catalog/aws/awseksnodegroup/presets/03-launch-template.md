# Launch-Template Pool

This preset composes the node group onto a first-class
`AwsLaunchTemplate`: the template owns the launch mechanics (instance
type, IMDSv2 enforcement, encrypted volumes, instance tags), and
promoting a new template version rolls the pool -- the template-driven
fleet-rollout pattern.

## When to Use

- Pools that need hardened launch posture: enforced IMDSv2, encrypted or
  provisioned-IOPS root volumes, instance-level tags
- Golden-AMI workflows where a template version promotion is the rollout
  mechanism
- Any pool whose launch settings should be reviewed and versioned
  independently of the pool's scaling

## Key Configuration Choices

- **`launchTemplate.version: $Default`** -- the pool follows the
  template's default version, so `aws ec2 modify-launch-template
  --default-version N` (or updating the AwsLaunchTemplate's
  `default_version`) rolls the fleet; pin a numeric version instead for
  fully drift-free plans
- **No inline `instanceTypes`/`diskSizeGb`/`remoteAccess`** -- AWS
  forbids them alongside a launch template; the template owns them (the
  spec enforces this before anything reaches AWS)
- **`amiType` stays set** -- valid with a template that does not pin a
  custom AMI; leave it out when the template brings its own image
- **`updateStrategy: MINIMAL`** -- template rollouts surge
  (launch-before-terminate), so capacity never dips

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<node-group-name>` | Name for the pool | Your naming convention |
| `<aws-region>` | AWS region code (e.g., `us-west-2`) | Your deployment region |
| `<cluster-resource-name>` | Name of the AwsEksCluster resource | Your cluster manifest's `metadata.name` |
| `<node-role-resource-name>` | Name of the AwsIamRole with the worker policies | Your role manifest's `metadata.name` |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |
| `<launch-template-resource-name>` | Name of the AwsLaunchTemplate resource | Your template manifest's `metadata.name` |

## Common Additions

- `taints` to dedicate the pool (e.g., GPU workloads on an accelerated
  template)
- `nodeRepairConfig.enabled: true` for managed auto-repair

## Related Presets

- **01-on-demand-general** -- inline launch mechanics for simple pools
- **02-spot-cost-optimized** -- interruptible capacity
