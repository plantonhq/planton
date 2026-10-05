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

1. `main.go` loads the `CloudflareQueueIacInput` (the manifest as `target`, plus the Cloudflare `provider_config`) from the Pulumi config key `planton:iac-input`, the `IAC_INPUT_YAML` environment variable (YAML content), or `IAC_INPUT_YAML_FILE` (a path to that YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and provisions the queue, then the consumer when the spec carries one.
3. The consumer is its own provider resource depending on the queue, so editing consumer settings never recreates the queue.
4. Outputs are exported matching `CloudflareQueueOutputs`.

## Local Development

Run the module from this directory with the planton CLI. It builds the IaC input from the manifest and the provider config file, and hands it to the module through `IAC_INPUT_YAML_FILE`:

```bash
# Create the stack once
planton pulumi init --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>

# Preview with the test manifest
planton pulumi preview --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> -p cloudflare-provider-config.yaml

# Deploy, then tear down
planton pulumi update --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> -p cloudflare-provider-config.yaml
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> -p cloudflare-provider-config.yaml
```

Without `-p`, the Cloudflare provider reads `CLOUDFLARE_API_TOKEN` from the environment.

## Dependencies

- `github.com/pulumi/pulumi-cloudflare/sdk/v6` — Cloudflare Pulumi provider
- `github.com/pulumi/pulumi/sdk/v3` — Pulumi SDK
- `github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule` — Shared IaC input loading and provider wiring
