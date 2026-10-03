# CloudflareHealthcheck Pulumi Module

Pulumi (Go) IaC module for a standalone origin health check -- scheduled probes with healthy/unhealthy status, no load balancer required.

## Architecture

```
main.go                    — Entrypoint loading the IaC input
module/main.go             — Resources(): provider setup, resource, outputs
module/locals.go           — Locals initialization
module/healthcheck.go      — cloudflare.Healthcheck
module/outputs.go          — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: type/config pairing, headers wrapper unwrapped to `header`, `healthcheck_id` / `zone_id` outputs.

## Outputs

| Name | Description |
|------|-------------|
| `healthcheck_id` | The created health check's ID |
| `zone_id` | The zone the health check belongs to |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
