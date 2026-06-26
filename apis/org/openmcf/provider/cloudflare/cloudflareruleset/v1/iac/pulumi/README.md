# CloudflareRuleset Pulumi Module

Pulumi IaC module for provisioning Cloudflare Rulesets.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go      — Resources() orchestrator
        ├── locals.go     — Locals struct and initialization
        ├── outputs.go    — Stack output key constants
        └── ruleset.go    — Ruleset creation and rule mapping
```

## How It Works

1. `main.go` loads the `CloudflareRulesetStackInput` from the `STACK_INPUT` environment variable (base64-encoded YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and provisions the ruleset.
3. `ruleset.go` maps proto `CloudflareRulesetRule` messages to Pulumi `cloudflare.RulesetRuleArgs` — including all action parameter sub-types (origin, response, uri, headers, from_value, overrides, cache settings).
4. Stack outputs are exported matching `CloudflareRulesetStackOutputs`.

## Engine parity note

One `action_parameters` field — `vary` (variant caching keyed on response headers) — is modeled in the
proto and provisioned by the Terraform module, but the pulumi-cloudflare SDK (v6.17.0) does not expose a
`vary` field on `RulesetRuleActionParametersArgs`, so this module omits it (see the inline note in
`module/ruleset.go` and `pkg/iac/MODULE_PARITY.md`). When a newer Pulumi SDK adds
`RulesetRuleActionParameters.Vary`, map it and remove the note. Every other ruleset field is at full
tofu↔Pulumi parity on provider v5.

## Local Development

```bash
# Build the binary
make build

# Preview with test manifest
make test

# Or use debug.sh for a specific manifest
./debug.sh ../hack/manifest.yaml
```

## Dependencies

- `github.com/pulumi/pulumi-cloudflare/sdk/v6` — Cloudflare Pulumi provider
- `github.com/pulumi/pulumi/sdk/v3` — Pulumi SDK
- `github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule` — Shared stack input loading and provider wiring
