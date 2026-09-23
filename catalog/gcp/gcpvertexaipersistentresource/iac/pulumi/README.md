# GcpVertexAiPersistentResource — Pulumi Implementation

This directory contains the Pulumi implementation for a Vertex AI persistent resource from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.vertex.AiPersistentResource`, with one guarded `organizations.LookupProject` when the peered network's path carries a project ID.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `persistentResource` |
| `module/locals.go` | The id defaulted from `metadata.name`; the `planton-ai_*` attribution labels |
| `module/persistent_resource.go` | Enables the API; resolves the network; maps the pools, networking, runtime spec, and key; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `persistent_resource_id`, `location`, `state`) |

## Send Posture (parity with Terraform)

- **`Name`** -- `spec.persistent_resource_id`, defaulting to `metadata.name`.
- **`Network`** -- the self-link prefix stripped; a non-numeric project segment resolved to its number by one project lookup.
- **Pools** -- `Id` only when set; `ReplicaCount` and the autoscaling bounds as decimal strings; machine fields only when set; `DiskSpec` only when the spec shapes it.
- **`ResourceRuntimeSpec`** -- emitted only when `enable_custom_service_account` is true.
- **`EncryptionSpec`**, **`PscInterfaceConfig`**, **`ReservedIpRanges`** -- emitted only when set.
- **`name` output** -- the resource ID (the SDK's `Name` is the short id).

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
