# CloudflareCertificatePack Pulumi Module

Pulumi IaC module for ordering an advanced edge certificate pack on a Cloudflare zone — a CA-issued certificate covering the zone apex and its hostnames.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go             — Resources() orchestrator
        ├── locals.go           — Locals struct and initialization
        ├── outputs.go          — output key constants
        └── certificate_pack.go — pack creation
```

## How It Works

1. `main.go` loads the `CloudflareCertificatePackIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and orders the pack. The `type` default (`advanced`) is coalesced here so a standalone run matches the Terraform module.
3. Outputs are exported matching `CloudflareCertificatePackOutputs`, including `zone_id` because a pack's API identity is (zone_id, certificate_pack_id).

A pack is an order, not an editable object: changing hosts, CA, validation method, or validity days replaces the pack.

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
