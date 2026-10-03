# CloudflareHyperdriveConfig Pulumi Module

Pulumi IaC module for provisioning a Cloudflare Hyperdrive config — the connection pooler and global cache a Worker binds to for low-latency access to a regional SQL database.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go              — Resources() orchestrator
        ├── locals.go            — Locals struct and initialization
        ├── outputs.go           — output key constants
        └── hyperdrive_config.go — config creation
```

## How It Works

1. `main.go` loads the `CloudflareHyperdriveConfigIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and provisions the config.
3. Cloudflare validates the origin connection at CREATE — an unreachable host, wrong credentials, or a blocked port fail the deploy, not the first query.
4. Outputs are exported matching `CloudflareHyperdriveConfigOutputs`.

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
