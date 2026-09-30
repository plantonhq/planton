# Blue/Green with Canary

This preset turns on ECS-native blue/green deployments: new revisions
stand up in the green target group, take 5% canary traffic for 5 minutes,
then all traffic, and bake 10 minutes before the blue side drains -- with
an instant rollback path the whole way. No CodeDeploy required; ECS swaps
the referenced production listener rule between the two target-group
nodes itself.

## When to Use

- Revenue-critical or hard-to-roll-back services where a bad deploy must
  never take full traffic
- Services whose regressions show up in minutes under real traffic --
  the canary + bake windows are exactly for catching those

## Key Configuration Choices

- **Two first-class target groups** -- blue (`targetGroupArn`) and green
  (`alternateTargetGroupArn`) are both `AwsLbTargetGroup` nodes; ECS
  registers new tasks into green and swaps the rule
- **`productionListenerRule` by reference** -- the `AwsLbListenerRule`
  whose forward action ECS rewrites during the swap; it must forward to
  the blue group initially
- **`canaryPercent: 5` then bake** -- a 5% blast radius for the first 5
  minutes; alarms or a manual abort roll back with traffic still 95%
  on blue
- **`bakeTimeInMinutes: 10`** -- after full shift, blue stays warm for
  10 minutes; rollback within the window is a rule swap, not a redeploy
- **`roleArn`** -- the IAM role ECS assumes to modify the listener
  rules; it needs elasticloadbalancing modify permissions on them

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<service-name>` | Name for the ECS service | Your service's name |
| `<aws-region>` | AWS region code | Your deployment region |
| `<cluster-resource-name>` | Name of the AwsEcsCluster resource | Your cluster manifest's `metadata.name` |
| `<task-definition-resource-name>` | Name of the AwsEcsTaskDefinition resource | Your task-definition manifest's `metadata.name` |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of the private AwsSubnet resources | Your subnet manifests' `metadata.name` |
| `<blue-target-group-resource-name>` / `<green-target-group-resource-name>` | Names of the two AwsLbTargetGroup resources | Your target-group manifests' `metadata.name` |
| `<production-rule-resource-name>` | Name of the AwsLbListenerRule ECS swaps | Your listener-rule manifest's `metadata.name` |
| `<blue-green-role-resource-name>` | Name of the AwsIamRole ECS assumes for the swap | Your role manifest's `metadata.name` |

## Common Additions

- `alarms` with rollback -- alarms firing during the canary or bake
  trigger the rollback automatically
- `deploymentConfiguration.lifecycleHooks` invoking a Lambda test suite
  at `POST_TEST_TRAFFIC_SHIFT` before production traffic moves
- A `testListenerRule` reference so smoke traffic reaches green before
  any production shift
