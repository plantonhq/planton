# Annual Enterprise Commitment

## Use Case

The common first commitment: 100 Enterprise slots in `US` for one year that end at the term (`renewalPlan: NONE`), set to leave management gracefully because Google will not delete a commitment before its term ends.

## When to Use

- Steady BigQuery load covered by an Enterprise reservation
- Lowering the rate of a reservation's baseline

## What This Creates

- A one-year, 100-slot Enterprise commitment in `US` that Google removes at the end of its term, with `ABANDON` on destroy and on any replacing change

Once Google removes it at the end of the term, remove this block (or destroy it with `ABANDON` before then): a deploy of the unchanged manifest after the term buys a new commitment.

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `slotCount` | `100` | Match the steady baseline of the reservations in this admin project. |
| `projectId` | `bq-admin-project` | A stand-in: name your BigQuery admin project, or reference a GcpProject. |
| `renewalPlan` | `NONE` | `ANNUAL` renews it every year, and under `ABANDON` it keeps renewing, and billing, with no manifest owning it. With a renewal, keep `plan` equal to `renewalPlan`. |
