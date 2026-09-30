# Web Service Behind ALB

This preset runs a load-balanced web fleet: instances register into an
`AwsLbTargetGroup` on launch, ELB health checks replace anything the
application health check fails, CPU target tracking holds utilization at
60%, and a surge-enabled instance refresh turns every launch-template
change into a zero-downtime rollout.

## When to Use

- HTTP/API services on EC2 behind an ALB (or NLB) target group
- Any fleet where "the port is open" is not the same as "the service is
  healthy"

## Key Configuration Choices

- **`healthCheckType: ELB`** -- trusts the target group's application
  health check; a wedged-but-running process gets replaced instead of
  serving errors forever
- **`minHealthyPercentage: 100` / `maxHealthyPercentage: 110`** --
  launch-before-terminate: the fleet never dips below full strength
  during a rollout, at the cost of brief 10% over-capacity
- **`OldestLaunchTemplate` termination policy** -- scale-in removes
  outdated instances first, so routine scale cycles also converge the
  fleet onto the newest template
- **`autoRollback: true`** -- a failed refresh returns the fleet to the
  previous configuration automatically

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<service-name>` | Name for the group | Your service's name (e.g., `api`, `web`) |
| `<aws-region>` | AWS region code (e.g., `us-east-1`) | Your deployment region |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |
| `<launch-template-resource-name>` | Name of the AwsLaunchTemplate resource | Your template manifest's `metadata.name` |
| `<target-group-resource-name>` | Name of the AwsLbTargetGroup resource | Your target-group manifest's `metadata.name` |

## Common Additions

- Add `defaultInstanceWarmupSeconds` (e.g., 60) so target tracking and
  refresh count fresh instances sooner
- Add an `ALBRequestCountPerTarget` tracking policy with a
  `resourceLabel` for request-based scaling
- Add `maxInstanceLifetimeSeconds: 604800` (7 days) for continuous fleet
  hygiene

## Related Presets

- **02-spot-mixed-fleet** -- blend an On-Demand base with Spot capacity
- **03-scheduled-scale** -- time-based capacity with a warm pool
