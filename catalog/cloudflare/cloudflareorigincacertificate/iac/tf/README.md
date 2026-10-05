# CloudflareOriginCaCertificate Terraform Module

Terraform IaC module for issuing a Cloudflare Origin CA certificate — the certificate an origin presents to Cloudflare so the edge can validate TLS to the origin without a public CA.

## Architecture

```
provider.tf   — Cloudflare provider configuration (plus hashicorp/tls for the generated-key path)
variables.tf  — Input variables mirroring CloudflareOriginCaCertificateSpec
locals.tf     — request_type / requested_validity defaults; generate_key flag
main.tf       — optional tls_private_key + tls_cert_request, then cloudflare_origin_ca_certificate
outputs.tf    — outputs (certificate_id, certificate, private_key, expires_on)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareOriginCaCertificate YAML manifest. For standalone use:

```hcl
module "origin_ca" {
  source = "./path/to/module"

  metadata = {
    name = "origin-cert"
  }

  spec = {
    hostnames    = ["example.com", "*.example.com"]
    request_type = "origin-rsa"
  }
}
```

When `spec.csr` is omitted the module mints the private key and CSR (the one-click path) and exports the key as a sensitive output. When a CSR is supplied, the user's key never leaves their control and `private_key` is empty. `csr` is write-only (the API never returns it); `requested_validity` is not returned after creation. Revoke is not delete — a just-revoked certificate may still answer GET for a window.

## Outputs

| Name | Description |
|------|-------------|
| `certificate_id` | Origin CA certificate identifier |
| `certificate` | Issued certificate in PEM (public material) |
| `private_key` | Generated private key in PEM (sensitive; empty when a CSR was supplied) |
| `expires_on` | RFC3339 expiry timestamp |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23` and `hashicorp/tls` for the generated-key path.
