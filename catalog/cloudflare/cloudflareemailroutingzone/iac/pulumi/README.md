# CloudflareEmailRoutingZone Pulumi Module

Pulumi IaC module for enabling Cloudflare Email Routing on a zone, with the folded per-zone catch-all rule and optional managed-DNS locking.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go               — Resources() orchestrator
        ├── locals.go             — Locals struct and initialization
        ├── outputs.go            — output key constants
        └── email_routing_zone.go — settings + catch-all + dns creation
```

## How It Works

1. `main.go` loads the `CloudflareEmailRoutingZoneIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and provisions the zone's routing.
3. `email_routing_zone.go` creates `EmailRoutingSettings` (the enable/disable toggle), then conditionally the catch-all (mapping each typed action — forward/worker/drop — onto the provider's generic `{type, values[]}`) and the managed-DNS resource (`lock_dns_records`, with `dns_name` for subdomain routing).
4. Outputs are exported matching `CloudflareEmailRoutingZoneOutputs`.

## Resource semantics worth knowing

The catch-all resource's provider Delete is a genuine no-op — destroying it drops it from state while the zone keeps its last catch-all configuration; the zone-level disable (destroying the settings resource) is what retires the behavior. See the inline notes in `email_routing_zone.go`.

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
