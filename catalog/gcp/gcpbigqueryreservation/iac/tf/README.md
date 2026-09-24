# GcpBigQueryReservation — Terraform Implementation

This directory contains the Terraform implementation for a BigQuery reservation from the Planton spec: one `google_project_service` (API enablement), one `google_bigquery_reservation`, and one `google_bigquery_reservation_assignment` per assignment.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the name defaulted from `metadata.name`, null-for-empty optionals, the assignments keyed by assignee, job type, and principal with the assignee composed, the attribution label merge |
| `main.tf` | `google_project_service`, `google_bigquery_reservation`, `google_bigquery_reservation_assignment` (`for_each`) |
| `outputs.tf` | `name` (the reservation's id), `reservation_name`, `location`, `assignment_names` |

## Send Posture

- **`name`** -- `spec.reservation_name`, defaulting to `metadata.name`.
- **`autoscale`** -- the lifted `max_slots`, sent only when positive.
- **`ignore_idle_slots`, `concurrency`, `edition`, `location`, `reservation_group`, `secondary_location`** -- sent only when set.
- **Assignments** -- keyed by `assignee|job_type|principal`; the assignee composed as `projects/`, `folders/`, or `organizations/`; they share the reservation's project, location, and `deletion_policy`.
- **`max_slots`, `scaling_mode`** -- not offered: google-beta-only at the pinned provider.
- **Labels** -- user labels merged under the attribution labels.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
