# CloudflareQueue Pulumi Module

Pulumi IaC module for provisioning a Cloudflare Queue and its optional single consumer — a managed, guaranteed-delivery message queue for Workers.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go    — Resources() orchestrator
        ├── locals.go  — Locals struct and initialization
        ├── outputs.go — output key constants
        └── queue.go   — queue + consumer creation
```

## How It Works

1. `main.go` loads the `CloudflareQueueIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and provisions the queue, then the consumer when the spec carries one.
3. The consumer is its own provider resource depending on the queue, so editing consumer settings never recreates the queue.
4. Outputs are exported matching `CloudflareQueueOutputs`.

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
