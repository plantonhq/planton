# Scheduled Scale with Warm Pool

This preset shapes capacity around a business calendar: four instances
minimum during weekday business hours, scale-to-zero overnight, and a
warm pool of stopped, pre-initialized instances so the morning ramp (and
any intra-day spike) serves in seconds instead of boot minutes -- at
near-zero cost while stopped.

## When to Use

- Internal tools, dashboards, and dev/test environments with
  predictable usage windows
- Slow-booting workloads (heavy AMIs, long user-data) where a cold
  launch takes minutes

## Key Configuration Choices

- **Explicit `desiredCapacity: 0` on the overnight action** -- absent
  values mean "leave unchanged", so scale-to-zero must be stated;
  optional fields make explicit zero expressible
- **`timeZone`** -- cron fires in local time, DST included; without it
  the schedule drifts an hour twice a year
- **`poolState: Stopped`** -- pooled instances cost only EBS storage;
  they start in seconds with boot work already done
- **`reuseOnScaleIn: true`** -- the overnight scale-down refills the
  pool instead of terminating warm instances, so tomorrow's ramp reuses
  today's boot

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<fleet-name>` | Name for the group | Your workload's name |
| `<aws-region>` | AWS region code (e.g., `us-east-1`) | Your deployment region |
| `<private-subnet-a-resource-name>` / `<private-subnet-b-resource-name>` | Names of two AwsSubnet resources in different AZs | Your subnet manifests' `metadata.name` |
| `<launch-template-resource-name>` | Name of the AwsLaunchTemplate resource | Your template manifest's `metadata.name` |
| `<iana-time-zone>` | IANA zone for the schedule (e.g., `America/New_York`) | Your business locale |

## Common Additions

- Add a target-tracking policy for intra-day demand on top of the
  scheduled floor
- Use `poolState: Hibernated` for JVM-style workloads that benefit from
  restored RAM
- Add a one-shot action (`startTime`, no `recurrence`) to pre-provision
  for a known event

## Related Presets

- **01-web-service-behind-alb** -- load-balanced fleet with ELB health and rolling refresh
- **02-spot-mixed-fleet** -- blend an On-Demand base with Spot capacity
