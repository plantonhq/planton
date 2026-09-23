# GcpVertexAiSearchDataConnector — Terraform Implementation

This directory contains the Terraform implementation for a Vertex AI Search
data connector from the Planton spec: one `google_project_service` (API
enablement) and one `google_discovery_engine_data_connector`, which creates
the collection and one data store per `spec.entities[]` entry.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the collection id and display name defaulted from `metadata.name`, null-for-empty optionals |
| `main.tf` | `google_project_service`, `google_discovery_engine_data_connector` with its entities, destinations, action config, and BAP config |
| `outputs.tf` | `name`, `collection_id`, `location`, `state`, `entity_data_stores`, `static_ip_addresses`, `private_connectivity_project_id` |

## Send Posture

- **`collection_id` / `collection_display_name`** -- `spec.collection_id` / `spec.collection_display_name`, each defaulting to `metadata.name` -- PARITY with the Pulumi module.
- **`params` XOR `json_params`** -- `null` when empty, so exactly the one the spec set reaches the provider.
- **Optional strings, lists, and maps** become `null` when empty (`incremental_refresh_interval`, `sync_mode`, `connector_modes`, `kms_key_name`, entity params and mappings, destination keys and hosts, action params, BAP lists); blocks are `dynamic` on presence.
- **Bools** are sent as declared.
- **`entity_data_stores`** -- a comprehension over the connector's `entities` in manifest order, skipping entries Google has not filled yet.
- **No labels** -- Discovery Engine resources carry none.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
