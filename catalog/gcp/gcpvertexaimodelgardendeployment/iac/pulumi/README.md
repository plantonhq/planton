# GcpVertexAiModelGardenDeployment — Pulumi Implementation

This directory contains the Pulumi implementation for a one-step Model
Garden deployment from the Planton spec: one `gcp.projects.Service` (API
enablement) and one `gcp.vertex.AiEndpointWithModelGardenDeployment`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `deployment` |
| `module/locals.go` | Carries the stack input (Google's deployment resource carries no labels of its own) |
| `module/deployment.go` | Enables the API; maps the model source, model config and serving container (three probe adapters), deploy config, and endpoint config; exports the outputs |
| `module/outputs.go` | Output key constants (`endpoint_id`, `endpoint_name`, `deployed_model_id`, `deployed_model_display_name`, `location`) |

## Send Posture (parity with Terraform)

- **Model source** -- exactly one of `PublisherModelName` / `HuggingFaceModelId` (proto-enforced), sent when set.
- **Optional strings and booleans** -- sent only when set; `AcceptEula` and `Spot` only when true.
- **`DedicatedResources`** -- `MinReplicaCount` always; `MaxReplicaCount` and `RequiredReplicaCount` (Optional+Computed) only when set; `MachineSpec` always emitted (the API requires the block) with each field only when set.
- **`SharedMemorySizeMb`** -- the spec's int64 rendered as the decimal string Google's API carries.
- **`PscAutomationConfigs`** -- the spec's single `psc_automation_config` (the provider caps the block at one and the SDK types it as one object).
- **`endpoint_id`** -- rebuilt as `projects/{project}/locations/{location}/endpoints/{endpoint}` from the resource's computed fields, the shape `GcpVertexAiEndpoint` exports.
- **`DeletionPolicy`** -- sent only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
