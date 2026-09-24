# GcpDatastreamStream — Terraform Implementation

This directory contains the Terraform implementation for a Datastream stream from the Planton spec: one `google_datastream_stream`. The Datastream API is enabled by the connection profiles the stream requires.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the id and display name defaulted from `metadata.name`, `desired_state` defaulted to `NOT_STARTED`, the dataset and BigQuery-connection conversions, the merged labels |
| `main.tf` | `google_datastream_stream` with one dynamic block per source and destination arm, the object lists level for level, the backfill mode, and rule sets |
| `outputs.tf` | `name`, `stream_id` |

## Send Posture

- **`stream_id`, `display_name`** -- defaulting to `metadata.name`.
- **`desired_state`** -- always sent, `NOT_STARTED` when unset.
- **Either-or enums** -- `cdc_method`, `large_objects_handling`, and `write_mode` emit only the chosen empty block.
- **Marker bools** -- `backfill_none` and `avro_file_format` emit Google's empty blocks.
- **`single_target_dataset.dataset_id`** -- a `GcpBigQueryDataset` self link trimmed to `projects/{p}/datasets/{d}`.
- **`blmt_config.connection_name`** -- a connection's full name converted to `{project}.{location}.{connection_id}`.
- **Object lists** -- empty names, zero positions, and false flags sent as null; empty lists emit no blocks.
- **Counts and optional strings** -- zero and empty sent as null, so Google's defaults apply.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
