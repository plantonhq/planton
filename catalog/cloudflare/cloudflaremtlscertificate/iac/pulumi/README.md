# CloudflareMtlsCertificate Pulumi Module

Pulumi (Go) IaC module for a certificate uploaded to the account-level mTLS certificate store.

## Architecture

```
main.go                       — Entrypoint loading the IaC input
module/main.go                — Resources(): provider setup, resource, outputs
module/locals.go              — Locals initialization
module/mtls_certificate.go    — cloudflare.MtlsCertificate
module/outputs.go             — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: create-only semantics (any change replaces the upload), the optional private key sent only when present, and the `certificate_id` / `expires_on` / `serial_number` outputs.

## Outputs

| Name | Description |
|------|-------------|
| `certificate_id` | The uploaded certificate's ID -- what consumers reference |
| `expires_on` | Expiry timestamp (RFC3339) |
| `serial_number` | The certificate's serial number |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
