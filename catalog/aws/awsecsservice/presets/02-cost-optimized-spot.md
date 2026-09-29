# Cost-Optimized Spot Blend

This preset blends Fargate capacity instead of naming a launch type: one
guaranteed on-demand task as the base, then a 1:4 on-demand:Spot split
for everything above it -- roughly 70% Spot at a ~70% discount on those
tasks, with AZ rebalancing keeping the fleet spread after interruptions.

## When to Use

- Interruption-tolerant services (stateless APIs behind retries, queue
  consumers, batch-ish workers) where compute cost matters
- Fleets of 3+ tasks -- the base task keeps at least one on-demand copy
  always running

## Key Configuration Choices

- **`base: 1` on FARGATE** -- the availability floor: one task is always
  on-demand no matter how Spot behaves
- **`weight: 1` vs `weight: 4`** -- tasks beyond the base scale 1:4
  on-demand:Spot; tune the ratio to your interruption tolerance
- **No `launchType`** -- a launch type and a capacity strategy are
  mutually exclusive; the blend replaces the single-type decision
- **`availabilityZoneRebalancing: ENABLED`** -- after an AZ event or a
  wave of Spot reclaims, tasks spread back across zones instead of
  piling up where capacity happened to be
- **Circuit breaker stays on** -- Spot interruptions are normal; failing
  DEPLOYMENTS still stop and roll back

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<service-name>` | Name for the ECS service | Your service's name |
| `<aws-region>` | AWS region code | Your deployment region |
| `<cluster-resource-name>` | Name of the AwsEcsCluster resource (must attach FARGATE + FARGATE_SPOT) | Your cluster manifest's `metadata.name` |
| `<task-definition-resource-name>` | Name of the AwsEcsTaskDefinition resource | Your task-definition manifest's `metadata.name` |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of the private AwsSubnet resources | Your subnet manifests' `metadata.name` |

## Common Additions

- `loadBalancers` wiring (see the web-service preset) when the blend
  fronts user traffic
- `autoscaling` -- the scaler and the blend compose: scaled-out tasks
  follow the same 1:4 split
