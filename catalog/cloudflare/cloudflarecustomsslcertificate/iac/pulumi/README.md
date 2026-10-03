# CloudflareCustomSslCertificate Pulumi Module

Pulumi (Go) IaC module for a bring-your-own TLS certificate uploaded to a Cloudflare zone.

## Architecture

```
main.go                              — Entrypoint loading the IaC input
module/main.go                       — Resources(): provider setup, resource, outputs
module/locals.go                     — Locals initialization
module/custom_ssl_certificate.go     — cloudflare.CustomSsl
module/outputs.go                    — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: replacement-on-rotation semantics, empty-string drops for `policy`/`custom_csr_id`, the nested geo restriction sent only when the label is present, and the `certificate_id` / `zone_id` / `expires_on` outputs. `priority` is deliberately absent (read-only at provider v5.23.0).

## Outputs

| Name | Description |
|------|-------------|
| `certificate_id` | The uploaded certificate's ID |
| `zone_id` | The zone the certificate belongs to |
| `expires_on` | Expiry timestamp (RFC3339) |

Deployment status is deliberately not an output: it transitions asynchronously (pending before active), so a point-in-time phase would flip on the first refresh and re-plan forever (the class was measured live on the sibling AOP certificate).

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
