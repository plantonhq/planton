# GcpDeliveryPipeline — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud Deploy delivery pipeline from the Planton spec: `gcp.projects.Service`, one `gcp.clouddeploy.DeliveryPipeline`, and the folded `gcp.clouddeploy.Automation` resources.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `deliveryPipeline` |
| `module/locals.go` | IaC input, attribution labels, the label merge |
| `module/delivery_pipeline.go` | API enablement, the pipeline, its stages and strategies, the send-when-set helpers, the outputs |
| `module/automation.go` | The automations, keyed by `automation_id` |
| `module/outputs.go` | Output key constants (`name`, `delivery_pipeline_id`, `uid`) |

## Send Posture (parity with Terraform)

- **Optional fields** -- set only when declared (strings non-empty, lists and maps non-empty, booleans true); a custom canary phase's `Percentage` always.
- **Per-path types** -- the SDK gives every job's container its own type; `readContainer` reads the spec once and each path shapes it. The canary paths carry `Actions` only, as the provider declares no tasks there.
- **Names** -- the pipeline's `Name` from `deliveryPipelineId` (default `metadata.name`); each automation's from `automationId`, logical name `<name>-automation-<id>`.
- **Automations** -- the spec's project, the pipeline's region and deletion policy, and the created pipeline's `Name`.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
