# GcpBigQueryReservation — Pulumi Implementation

This directory contains the Pulumi implementation for a BigQuery reservation from the Planton spec: one `gcp.projects.Service`, one `gcp.bigquery.Reservation`, and one `gcp.bigquery.ReservationAssignment` per assignment.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `reservation` |
| `module/locals.go` | The defaulted name, the composed assignments in declared order, attribution labels |
| `module/reservation.go` | Enables the API; creates the reservation and its assignments; exports the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`name`** -- `spec.reservation_name`, defaulting to `metadata.name`.
- **`autoscale`** -- the lifted `max_slots`, sent only when positive.
- **`ignore_idle_slots`, `concurrency`, `edition`, `location`, `reservation_group`, `secondary_location`** -- sent only when set.
- **Assignments** -- keyed by `assignee|job_type|principal`; the assignee composed as `projects/`, `folders/`, or `organizations/`; they share the reservation's project, location, and `deletion_policy`.
- **`max_slots`, `scaling_mode`** -- not offered: google-beta-only at the pinned provider.
- **Labels** -- user labels merged under the attribution labels.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
