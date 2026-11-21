# GcpBigQueryCapacityCommitment — Pulumi Implementation

This directory contains the Pulumi implementation for a BigQuery capacity commitment from the Planton spec: one `gcp.projects.Service` and one `gcp.bigquery.CapacityCommitment`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `capacityCommitment` |
| `module/locals.go` | The defaulted id |
| `module/capacity_commitment.go` | Enables the API; creates the commitment; exports the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`capacity_commitment_id`** -- defaulting to `metadata.name` rather than a Google-generated id.
- **`enforce_single_admin_project_per_org`** -- the spec's bool passed as the provider's string `"true"` when set, omitted otherwise. The provider never sends it to Google (Google takes it as a create-call query parameter the provider leaves out), so it guards nothing; changing it still replaces the commitment.
- **`renewal_plan`, `edition`, `location`** -- sent only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
