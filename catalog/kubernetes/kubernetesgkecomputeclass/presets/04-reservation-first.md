# Reservation First

A class that spends capacity you have already paid for before it buys more. It fills a named Compute Engine reservation, then any matching reservation the project can use, and only then creates on-demand nodes, falling back to a second family if the first is out.

## When to Use

- A team with committed-use reservations that sit partly idle.
- Capacity you must guarantee for a launch or a peak, held in a reservation, with on-demand as overflow.

## Key Configuration Choices

- **A Specific reservation first.** GKE checks that the named reservation exists and that the cluster's project can use it; until the placeholder names one, the class reports `CrdMisconfigured` (ReservationNotFound) and schedules nothing.
- **Specific never falls back by itself.** `affinity: Specific` consumes only the named reservation; the rule is skipped when the reservation is full, and the rules after it are the fallback.
- **The machine type must match the reservation.** A reservation holds one machine type, so the first two rules name it exactly.
- **`AnyBestEffort` second.** It uses any other matching non-specific reservation and falls back to on demand, which covers shared reservations.
- **A second family last.** If on-demand n2d runs out too, c3d at the same size takes over.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<reservation-name>` | The Compute Engine reservation to consume | `gcloud compute reservations list` |
| `<reservation-zone>` | The zone the reservation holds capacity in | Same listing, ZONE column |

## Related Presets

- **01-stockout-resilient-default**: the on-demand cluster default.
