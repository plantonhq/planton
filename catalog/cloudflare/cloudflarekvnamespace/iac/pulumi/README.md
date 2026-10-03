# CloudflareKvNamespace Pulumi Module

Pulumi IaC module for provisioning a Cloudflare Workers KV namespace — the account-scoped, low-latency key-value container Workers read at the edge.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go         — Resources() orchestrator
        ├── locals.go       — Locals struct and initialization
        ├── outputs.go      — output key constants
        └── kv_namespace.go — namespace creation
```

## How It Works

1. `main.go` loads the `CloudflareKvNamespaceIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and provisions the namespace.
3. Namespace titles are unique within an account; entries are seeded separately as `CloudflareWorkersKvPair` resources.
4. Outputs are exported matching `CloudflareKvNamespaceOutputs`.

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
