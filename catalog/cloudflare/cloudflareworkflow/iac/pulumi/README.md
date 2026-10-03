# CloudflareWorkflow Pulumi Module

Pulumi (Go) IaC module for a Workflow registration: the binding of a durable-execution class exported by a deployed Worker script to a named workflow.

## Architecture

```
main.go              — Entrypoint loading the IaC input
module/main.go       — Resources(): provider setup, resource, outputs
module/locals.go     — Locals initialization
module/workflow.go   — cloudflare.Workflow
module/outputs.go    — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: retention values pass through verbatim (the API accepts milliseconds or duration expressions), absent retention/limits/schedules trees keep Cloudflare's defaults, and the `workflow_name` / `version_id` outputs. Create is a PUT at the API (name-as-upsert); account_id and workflow_name force replacement.

## Outputs

| Name | Description |
|------|-------------|
| `workflow_name` | The workflow's identity -- what Worker workflow bindings reference |
| `version_id` | The workflow version the registration produced |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
