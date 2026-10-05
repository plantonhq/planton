# CloudflareAuthenticatedOriginPullsCertificate Pulumi Module

Pulumi (Go) IaC module for an Authenticated Origin Pulls client-certificate upload (zone-wide or hostname-scoped).

## Architecture

```
main.go                   — Entrypoint loading the IaC input
module/main.go            — Resources(): provider setup, resource, outputs
module/locals.go          — Locals initialization
module/certificate.go     — cloudflare.AuthenticatedOriginPullsCertificate
                            XOR cloudflare.AuthenticatedOriginPullsHostnameCertificate
module/outputs.go         — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: the `scope` selects which resource is created (exactly one), rotation is replacement (never key-only -- the zone surface silently ignores it at v5.23.0), and the `certificate_id` / `zone_id` / `expires_on` outputs.

## Outputs

| Name | Description |
|------|-------------|
| `certificate_id` | The uploaded certificate's ID -- what associations reference |
| `zone_id` | The zone the certificate belongs to |
| `expires_on` | Expiry timestamp (RFC3339) |

Deployment status is deliberately not an output: it transitions asynchronously (pending_deployment to active seconds after create), so a point-in-time phase would flip on the first refresh and re-plan forever.

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
