# GcpBigQueryReservationGroup — Pulumi Implementation

This directory contains the Pulumi implementation for a BigQuery reservation group from the Planton spec: one `gcp.projects.Service` and one `gcp.bigquery.ReservationGroup`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `reservationGroup` |
| `module/locals.go` | The defaulted name |
| `module/reservation_group.go` | Enables the API; creates the group; exports the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`name`** -- `spec.reservation_group_name`, defaulting to `metadata.name`.
- **`location`** -- sent only when set (Google's default is `US`).
- **`name` output** -- the resource's id, the full path.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
