# AwsEcsService

Runs an ECS service: the scheduler that keeps a desired number of copies
of a referenced task definition running in a referenced cluster, replaces
unhealthy tasks, rolls deployments safely, and registers tasks into
referenced load-balancer target groups.

## Purpose

The service is pure scheduling -- everything it touches is a first-class
node it references: the `AwsEcsTaskDefinition` (what runs), the
`AwsEcsCluster` (where it runs), `AwsSubnet`/`AwsSecurityGroup` (the task
network), and `AwsLbTargetGroup` (where traffic arrives, routed by
`AwsLbListener`/`AwsLbListenerRule` nodes). Because the task-definition
reference resolves to a revision-carrying ARN, registering a new revision
rolls the service -- deploys travel through the resource graph.

## Key Features

- **Capacity as a spectrum** -- name a launch type (`FARGATE`, `EC2`,
  `EXTERNAL`) or blend `FARGATE`/`FARGATE_SPOT`/EC2 capacity providers
  with base/weight strategies for cost-optimized fleets.
- **Deployments guarded in layers** -- the deployment circuit breaker
  (stop a failing rollout, roll back), CloudWatch alarm gating (roll
  back on error-rate or latency regressions), and ECS-native
  **blue/green** with canary or linear traffic shifting, bake time, and
  Lambda lifecycle hooks -- no CodeDeploy required.
- **Service Connect** -- sidecar-free service-to-service discovery,
  retries, telemetry, per-request access logs, header-based test-traffic
  routing for blue/green, and optional Private-CA TLS on top of Cloud Map.
- **VPC Lattice attachments** -- register tasks straight into Lattice
  target groups for cross-VPC and cross-account traffic without a load
  balancer in every VPC.
- **Folded autoscaling** -- target tracking on CPU, memory, and ALB
  requests-per-target (composing the ALB's and target group's
  `arn_suffix` outputs), with cooldowns and scale-in control.
- **Managed EBS task volumes** -- real block storage per task, sized,
  encrypted, tagged for cost allocation, and optionally
  snapshot-hydrated at deployment time, against the task definition's
  `configureAtLaunch` volume of the same name.
- **Operations posture** -- ECS Exec, AZ rebalancing, tag propagation,
  placement strategies/constraints for EC2 fleets.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEcsService
metadata:
  name: api
spec:
  region: us-west-2
  clusterArn:
    valueFrom:
      kind: AwsEcsCluster
      name: prod
      fieldPath: status.outputs.cluster_arn
  taskDefinition:
    valueFrom:
      kind: AwsEcsTaskDefinition
      name: api
      fieldPath: status.outputs.task_definition_arn
  desiredCount: 2
  network:
    subnets:
      - valueFrom: { kind: AwsSubnet, name: private-a, fieldPath: status.outputs.subnet_id }
      - valueFrom: { kind: AwsSubnet, name: private-b, fieldPath: status.outputs.subnet_id }
  loadBalancers:
    - targetGroupArn:
        valueFrom: { kind: AwsLbTargetGroup, name: api, fieldPath: status.outputs.target_group_arn }
      containerName: app
      containerPort: 8080
  deploymentCircuitBreaker:
    enable: true
    rollback: true
```

Deploy with:

```shell
planton apply -f service.yaml
```

Both a Pulumi module and a Terraform/OpenTofu module implement this
kind at full behavioral parity; the provisioner is an execution
detail.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
