# CloudflareZoneTlsSettings Pulumi Module

Pulumi (Go) IaC module for managing a Cloudflare zone's edge TLS posture — Universal SSL issuance, Total TLS, automatic origin TLS key exchange, origin TLS compliance modes, per-hostname TLS overrides, and certificate-authority hostname associations.

## Architecture

```
main.go             — Entrypoint loading the IaC input
module/main.go      — Resources(): provider setup, settings, outputs
module/locals.go    — Locals initialization
module/tls_settings.go — Singletons + the per-(setting, hostname) override fan-out
module/outputs.go   — output keys (zone_id)
```

## Behavior

An unset spec field is NOT MANAGED: the module never sends it. Universal SSL, Total TLS, auto origin key exchange, and CA hostname associations have no delete at Cloudflare — destroy drops state and abandons the live values. Per-hostname overrides and origin TLS compliance modes have real deletes. Disabling Universal SSL can make proxied hostnames unreachable over HTTPS unless other certificates cover them — only disable it deliberately.

Each hostname row's set attributes become separate `HostnameTlsSetting` resources keyed by (setting id, hostname), and each CA association row manages either the zone's managed-CA list (no certificate id) or one mTLS certificate's hostname list. The module mirrors the Terraform module's contract exactly, including the `zone_id` output.

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | The zone the TLS settings belong to (the singleton's identity) |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
