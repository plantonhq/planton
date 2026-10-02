# GcpDatastreamStream — Pulumi Implementation

This directory contains the Pulumi implementation for a Datastream stream from the Planton spec: one `gcp.datastream.Stream`. The Datastream API is enabled by the connection profiles the stream requires.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `stream` |
| `module/locals.go` | The defaulted id, display name, and desired state; the dataset and BigQuery-connection conversions; the merged labels |
| `module/stream.go` | Maps the source and destination arms, the backfill mode, and rule sets; exports the outputs |
| `module/object_lists.go` | One builder per source family and place (include, exclude, backfill exclusions) -- the SDK types each place separately |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`stream_id`, `display_name`** -- defaulting to `metadata.name`.
- **`desired_state`** -- always sent, `NOT_STARTED` when unset.
- **Either-or enums** -- `cdc_method`, `large_objects_handling`, and `write_mode` send only the chosen empty block.
- **Marker bools** -- `backfill_none` and `avro_file_format` send Google's empty blocks.
- **`single_target_dataset.dataset_id`** -- a `GcpBigQueryDataset` self link trimmed to `projects/{p}/datasets/{d}`.
- **`blmt_config.connection_name`** -- a connection's full name converted to `{project}.{location}.{connection_id}`.
- **Object lists** -- empty names, zero positions, and false flags omitted; empty lists omitted.
- **Counts and optional strings** -- zero and empty omitted, so Google's defaults apply.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
