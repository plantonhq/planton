# CloudflareEmailRoutingRule Pulumi Module

Pulumi IaC module for provisioning a single Cloudflare Email Routing rule: match incoming mail and drop it, forward it, and/or hand it to an Email Worker.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go               — Resources() orchestrator
        ├── locals.go             — Locals struct and initialization
        ├── outputs.go            — output key constants
        └── email_routing_rule.go — rule creation and typed-action mapping
```

## How It Works

1. `main.go` loads the `CloudflareEmailRoutingRuleIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and provisions the rule.
3. `email_routing_rule.go` maps the matchers 1:1 and each typed action (forward/worker/drop — a rule carries a LIST of actions, matching the Cloudflare API) onto the provider's generic `{type, values[]}`.
4. Outputs are exported matching `CloudflareEmailRoutingRuleOutputs`.

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
