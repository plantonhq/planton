# GcpVertexAiSearchDataConnector — Pulumi Implementation

This directory contains the Pulumi implementation for a Vertex AI Search
data connector from the Planton spec: one `gcp.projects.Service` (API
enablement) and one `gcp.discoveryengine.DataConnector`, which creates the
collection and one data store per `spec.entities[]` entry.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `dataConnector` |
| `module/locals.go` | The collection id and display name defaulted from `metadata.name` |
| `module/data_connector.go` | Enables the API; maps the connector, its entities, destinations, and action side; exports the outputs (the created stores read from `Entities[].DataStore`) |
| `module/outputs.go` | Output key constants (`name`, `collection_id`, `location`, `state`, `entity_data_stores`, `static_ip_addresses`, `private_connectivity_project_id`) |

## Send Posture (parity with Terraform)

- **`CollectionId` / `CollectionDisplayName`** -- `spec.collection_id` / `spec.collection_display_name`, each defaulting to `metadata.name`.
- **`Params` XOR `JsonParams`** -- whichever the spec sets (exactly one, a spec rule); the maps carry Secret Manager resource names, never secret material.
- **Optional strings, lists, and blocks** (`DataSourceVersion`, `IncrementalRefreshInterval`, `SyncMode`, `ConnectorModes`, `KmsKeyName`, entity params and mappings, destinations, `ActionConfig`, `BapConfig`) are sent only when set so Google's defaults stay in charge.
- **Bools** (`IncrementalSyncDisabled`, `AutoRunDisabled`, `StaticIpEnabled`, `CreateBapConnection`) are sent as declared.
- **`entity_data_stores`** -- the created stores in manifest order, read back from the connector's entities.
- **No labels** -- Discovery Engine resources carry none.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
