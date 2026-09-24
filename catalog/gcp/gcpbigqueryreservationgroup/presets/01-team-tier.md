# Team Tier

## Use Case

One team's or one priority tier's reservations share their idle slots with each other before the rest of the admin project sees them.

## When to Use

- Several reservations owned by one team
- Separating a production tier from ad-hoc capacity

## What This Creates

- The `analytics-tier` group in `US` for reservations to join

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `US` | The same location as its reservations. Fixed at creation. |
