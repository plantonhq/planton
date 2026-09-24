# Annual Enterprise Commitment

## Use Case

The common first commitment: 100 Enterprise slots in `US` for one year, renewing annually, set to leave management gracefully because Google will not delete it early.

## When to Use

- Steady BigQuery load covered by an Enterprise reservation
- Lowering the rate of a reservation's baseline

## What This Creates

- A one-year, 100-slot Enterprise commitment in `US` with annual renewal and `ABANDON` on destroy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `slotCount` | `100` | Match the steady baseline of the reservations in this admin project. |
| `renewalPlan` | `ANNUAL` | What the commitment becomes after the term. |
