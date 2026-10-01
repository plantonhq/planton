# Web Service

This preset runs a load-balanced web service: a referenced task-definition
revision scheduled into a referenced cluster, registered into a
first-class `AwsLbTargetGroup` that an `AwsLbListener` (or
`AwsLbListenerRule`) routes into, guarded by the deployment circuit
breaker, and scaled on CPU between 2 and 10 tasks.

## When to Use

- HTTP/API services behind an ALB -- the standard production shape
- Any service whose deploys should be composition-driven: registering a
  new task-definition revision rolls this service automatically

## Key Configuration Choices

- **Everything composes by reference** -- the cluster, task definition,
  subnets, security group, and target group are all first-class nodes;
  this service only schedules and registers task IPs
- **The target group is NOT created here** -- it belongs to the routing
  layer (`AwsLbTargetGroup` + listener/rule), so multiple services,
  listeners, and blue/green pairs can share and swap it
- **`deploymentCircuitBreaker` with rollback** -- a failing rollout stops
  and reverts instead of thrashing; the zero-configuration deployment
  guard every production service should carry
- **`healthCheckGracePeriodSeconds: 90`** -- slow-booting apps are not
  killed by the load balancer mid-startup
- **CPU target tracking at 70%** -- headroom for spikes without paying
  for idle; desired_count seeds 2 and the scaler owns it from there

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<service-name>` | Name for the ECS service | Your service's name (e.g., `api`) |
| `<aws-region>` | AWS region code | Your deployment region |
| `<cluster-resource-name>` | Name of the AwsEcsCluster resource | Your cluster manifest's `metadata.name` |
| `<task-definition-resource-name>` | Name of the AwsEcsTaskDefinition resource | Your task-definition manifest's `metadata.name` |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of the private AwsSubnet resources | Your subnet manifests' `metadata.name` |
| `<security-group-resource-name>` | Name of the AwsSecurityGroup resource | Your security-group manifest's `metadata.name` |
| `<target-group-resource-name>` | Name of the AwsLbTargetGroup resource | Your target-group manifest's `metadata.name` |

## Common Additions

- `alarms` referencing `AwsCloudwatchAlarm` nodes to roll back on error
  rates or latency, not just task health
- `requests_per_target` autoscaling (composes the ALB's and target
  group's `arn_suffix` outputs) for the most direct request-serving
  signal
- `capacityProviderStrategy` blending FARGATE and FARGATE_SPOT (drop
  `launchType`) to cut compute cost on interruption-tolerant services
