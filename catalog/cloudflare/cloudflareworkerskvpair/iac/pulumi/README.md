# CloudflareWorkersKvPair Pulumi Module

Pulumi IaC module for provisioning a single Workers KV entry — a versioned, infrastructure-seeded key-value pair inside a KV namespace.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go    — Resources() orchestrator
        ├── locals.go  — Locals struct and initialization
        ├── outputs.go — output key constants
        └── kv_pair.go — entry creation
```

## How It Works

1. `main.go` loads the `CloudflareWorkersKvPairIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and writes the entry.
3. Account, namespace, and key all force replacement when changed — an entry's identity is the full `{account}/{namespace}/{key}` triple.
4. Outputs are exported matching `CloudflareWorkersKvPairOutputs`.

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
