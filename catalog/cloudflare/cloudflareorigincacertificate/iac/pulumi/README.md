# CloudflareOriginCaCertificate Pulumi Module

Pulumi IaC module for issuing a Cloudflare Origin CA certificate — the certificate an origin presents to Cloudflare so the edge can validate TLS to the origin without a public CA.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go                     — Resources() orchestrator
        ├── locals.go                   — Locals struct and initialization
        ├── outputs.go                  — output key constants
        └── origin_ca_certificate.go    — optional key+CSR generation, then the certificate
```

## How It Works

1. `main.go` loads the `CloudflareOriginCaCertificateIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals and creates a Cloudflare provider. When `spec.csr` is omitted it generates a private key (RSA or ECDSA keyed to `request_type`) and a CSR for the requested hostnames; when a CSR is supplied, the user's key never leaves their control.
3. Outputs are exported matching `CloudflareOriginCaCertificateOutputs`. `private_key` is a secret and is empty on the BYO-CSR path.

`csr` is write-only. Revoke is not delete — a just-revoked certificate may still answer GET for a window.

## Local Development

```bash
# Build the binary
make build

# Preview with test manifest
make test

# Or use debug.sh for a specific manifest
./debug.sh ../../e2e/manifest.yaml
```

## Dependencies

- `github.com/pulumi/pulumi-cloudflare/sdk/v6` — Cloudflare Pulumi provider
- `github.com/pulumi/pulumi-tls/sdk/v4` — TLS helpers for the generated-key path
- `github.com/pulumi/pulumi/sdk/v3` — Pulumi SDK
- `github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule` — Shared IaC input loading and provider wiring
