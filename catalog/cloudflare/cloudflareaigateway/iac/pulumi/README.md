# CloudflareAiGateway Pulumi Module

Pulumi (Go) IaC module for one AI Gateway and its dynamic routes.

## Architecture

```
main.go                     — Entrypoint loading the IaC input
module/main.go              — Resources(): provider setup, gateway then routes
module/locals.go            — Locals initialization
module/ai_gateway.go        — cloudflare.AiGateway + nested-tree builders
module/dynamic_routes.go    — cloudflare.AiGatewayDynamicRouting per route row
module/outputs.go           — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: the spec's grouped retry{} / log_management{} fan out to the provider's flat arguments; the spec's on_true/on_false route edges map to the wire's true/false; spend-limit rule ids are always sent explicitly (the provider default is a leaked example value that collapses rules); otel and stripe authorization values are kept secret in Pulumi state (`pulumi.ToSecret`) though the provider leaves them unmarked. Routes attach through the created gateway's id, so they create after it and die with it.

The module always sends `authentication`, `logpush`, `zdr`, and the log-management pair even when the manifest omits them: Cloudflare echoes false / 100000 / DELETE_OLDEST for the unset forms on every read (live-measured 2026-08-27), so an omitted send would drift on any refresh. The sent defaults ARE Cloudflare's own -- semantics are unchanged.

**Engine note (live-measured 2026-08-27)**: the API is write-only for `guardrails`, `spend_limits`, `otel`, and route graphs. Pulumi lifecycles verified clean against the live API (previews do not refresh), but the same surfaces make Terraform plans never settle at provider v5.23.0-v5.24.0 -- see the Terraform module README and the GUIDE before choosing an engine for gateways carrying those fields.

## Outputs

| Name | Description |
|------|-------------|
| `gateway_id` | The gateway's URL slug |
| `dynamic_route_ids` | Each managed route's id, keyed by route name |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
