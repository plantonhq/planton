# CloudflareListItem Pulumi Module

Pulumi IaC module for writing a single entry into a Cloudflare List — an IP/CIDR, ASN, hostname, or redirect, matching the parent list's kind.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go      — Resources() orchestrator
        ├── locals.go    — Locals struct and initialization
        ├── outputs.go   — output key constants
        └── list_item.go — item creation
```

## How It Works

1. `main.go` loads the `CloudflareListItemIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and writes the entry. Exactly one of `ip` / `asn` / `hostname` / `redirect` is sent.
3. Outputs are exported matching `CloudflareListItemOutputs`.

Item values are immutable in the provider: changing an entry replaces it. Do not also declare inline `items` on the parent CloudflareList.

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
