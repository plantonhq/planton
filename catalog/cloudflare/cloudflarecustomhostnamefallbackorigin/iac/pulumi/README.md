# CloudflareCustomHostnameFallbackOrigin Pulumi Module

Pulumi IaC module for setting a zone's fallback origin — the default backend every custom hostname on that zone routes to when no more-specific origin is configured.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go            — Resources() orchestrator
        ├── locals.go          — Locals struct and initialization
        ├── outputs.go         — output key constants
        └── fallback_origin.go — fallback origin create/update
```

## How It Works

1. `main.go` loads the `CloudflareCustomHostnameFallbackOriginIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and writes the fallback origin. The write path is PUT — create equals update.
3. Outputs are exported matching `CloudflareCustomHostnameFallbackOriginOutputs`, including `zone_id` because this singleton has no resource id of its own.

This is a zone singleton: one fallback origin per zone, and its API identity IS the zone.

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
