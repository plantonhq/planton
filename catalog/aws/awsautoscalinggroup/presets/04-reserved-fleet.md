# Reserved Fleet with Guaranteed Capacity

This preset runs a fleet on capacity you have already paid for. The group
fills the targeted On-Demand Capacity Reservations first
(`capacity-reservations-first`) and falls back to regular On-Demand only
when the reservations are exhausted, while
`reservations-then-balanced` lets zone balance bend toward the zones
holding the reservations instead of fighting them.

The lifecycle posture matches a fleet that holds real work: the drain
hook pauses terminations for up to ten minutes, and
`terminateHookAbandon: retain` keeps any instance whose drain FAILED
running (outside the group) for a post-mortem instead of destroying the
evidence. The warm-cache hook is attached atomically at group creation
(`applyAtLaunch`), so even the very first instance pauses to warm before
taking traffic.

## When to Use

- Steady baseline fleets whose capacity is covered by On-Demand Capacity
  Reservations (or Savings-Plan-backed reservations)
- Workloads where a failed drain must be debuggable, not silently
  terminated

## Key Configuration Choices

- **`preference: capacity-reservations-first`** -- consume the paid
  capacity before anything else; use `capacity-reservations-only` to fail
  instead of falling back
- **`capacityDistributionStrategy: reservations-then-balanced`** -- zone
  placement follows the reservations first
- **`terminateHookAbandon: retain`** -- keep failed-drain instances for
  post-mortems; retained instances stop counting toward capacity
- **`applyAtLaunch` on the warm-cache hook** -- creation-time attachment
  closes the first-instance race; note AWS makes creation-time hooks
  immutable (changing one replaces the group)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<fleet-name>` | Name for the auto-scaling group | Your fleet's name (e.g., `api-reserved`) |
| `<aws-region>` | AWS region code (e.g., `us-east-1`) | Your deployment region |
| `<subnet-a-resource-name>` / `<subnet-b-resource-name>` | Names of the AwsSubnet resources | Your subnet manifests' `metadata.name` |
| `<launch-template-resource-name>` | Name of the AwsLaunchTemplate resource | Your template manifest's `metadata.name` |
| `cr-<your-reservation-id>` | The Capacity Reservation to fill first | EC2 console, Capacity Reservations |

## Common Additions

- `capacityReservationResourceGroupArns` instead of IDs, to target a
  resource group that collects reservations as they rotate
- A `targetGroups` reference plus `healthCheckType: ELB` when the fleet
  serves a load balancer
- An SNS `notifications` block to observe launches and terminations

## Related Presets

- **01-web-service-behind-alb** -- the load-balanced web fleet baseline
- **02-spot-mixed-fleet** -- the opposite cost posture: Spot-majority
  mixed instances
