# CloudflarePagesProject Pulumi Module

Pulumi IaC module for provisioning a Cloudflare Pages project — build config, optional git source, per-environment deployment configuration, and attached custom domains.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go               — Resources() orchestrator
        ├── locals.go             — Locals struct and initialization
        ├── outputs.go            — output key constants
        ├── deployment_config.go  — preview/production config transform
        └── project.go            — project + custom domains
```

## How It Works

1. `main.go` loads the `CloudflarePagesProjectIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and provisions the project. When only one of preview/production deployment configs is supplied, it is mirrored to both — Cloudflare rejects inconsistently configured environments.
3. Each hostname in `spec.domains` becomes a `cloudflare_pages_domain` companion.
4. Outputs are exported matching `CloudflarePagesProjectOutputs`.

The project name is the identity. Secret env vars inside `deployment_configs` do not survive import.

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
