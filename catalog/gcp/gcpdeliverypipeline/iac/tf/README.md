# GcpDeliveryPipeline — Terraform Implementation

This directory contains the Terraform implementation for a Cloud Deploy delivery pipeline from the Planton spec: `google_project_service` for the Cloud Deploy API, one `google_clouddeploy_delivery_pipeline`, and the folded `google_clouddeploy_automation` resources.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Project, pipeline ID default, attribution labels, stages, automations keyed by their IDs |
| `main.tf` | `google_project_service`, the pipeline with its stages and strategies, the automations |
| `outputs.tf` | `name`, `delivery_pipeline_id`, `uid` |

## Send Posture

- **Optional fields** -- sent only when set (strings non-empty, lists and maps non-empty, booleans true), so Cloud Deploy's defaults apply otherwise; a custom canary phase's `percentage` is always sent.
- **Strategies** -- every nested block is a `dynamic` block rendered only when declared; the canary paths' predeploy and postdeploy carry `actions` only, because the provider declares no `tasks` there.
- **Automations** -- keyed by `automation_id`; each takes the pipeline's project, region, deletion policy, and the created pipeline's ID -- PARITY with the Pulumi module.
- **Labels** -- attribution labels merged into the pipeline's and every automation's `labels`.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
