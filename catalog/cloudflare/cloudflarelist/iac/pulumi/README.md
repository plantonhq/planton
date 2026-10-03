# CloudflareList Pulumi Module

Pulumi IaC module for provisioning an account-scoped Cloudflare List — a named collection referenced from rule expressions (WAF, custom rules, Bulk Redirect).

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go    — Resources() orchestrator
        ├── locals.go  — Locals struct and initialization
        ├── outputs.go — output key constants
        └── list.go    — list creation
```

## How It Works

1. `main.go` loads the `CloudflareListIacInput` from the `IAC_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and provisions the list. Inline `items` are never sent — entries are a separate kind (CloudflareListItem).
3. Outputs are exported matching `CloudflareListOutputs`.

`kind` and `name` are immutable. List names must match `^[a-zA-Z][a-zA-Z0-9_]*$` (no hyphens) because they appear in rule expressions as `$name`.

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
