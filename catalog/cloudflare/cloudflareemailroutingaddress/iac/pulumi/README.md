# CloudflareEmailRoutingAddress Pulumi Module

Pulumi IaC module for provisioning a Cloudflare Email Routing destination address — the account-scoped, verified mailbox that routing rules forward to.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go                  — Resources() orchestrator
        ├── locals.go                — Locals struct and initialization
        ├── outputs.go               — output key constants
        └── email_routing_address.go — address creation
```

## How It Works

1. `main.go` loads the `CloudflareEmailRoutingAddressIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and provisions the address.
3. Creating the address sends a verification email; the `verified` output stays empty until the owner clicks the link.
4. Outputs are exported matching `CloudflareEmailRoutingAddressOutputs`.

## Engine parity note

The spec's `status` field (explicit verification-state override) is provisioned by the Terraform module but NOT by this module: the pulumi-cloudflare SDK (v6.17.0) `EmailRoutingAddressArgs` carries only `AccountId` and `Email` while the terraform provider at v5.23.0 has `status`. See the `PARITY-EXCEPTION` note in `module/email_routing_address.go`; wire the field and remove the note when a newer Pulumi SDK adds it. Every other field is at full tofu↔Pulumi parity.

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
- `github.com/pulumi/pulumi/sdk/v3` — Pulumi SDK
- `github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule` — Shared IaC input loading and provider wiring
