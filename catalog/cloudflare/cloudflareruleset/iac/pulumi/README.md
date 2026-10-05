# CloudflareRuleset Pulumi Module

Pulumi IaC module for provisioning Cloudflare Rulesets.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go      — Resources() orchestrator
        ├── locals.go     — Locals struct and initialization
        ├── outputs.go    — output key constants
        └── ruleset.go    — Ruleset creation and rule mapping
```

## How It Works

1. `main.go` loads the `CloudflareRulesetIacInput` (the manifest as `target`, plus the Cloudflare `provider_config`) from the Pulumi config key `planton:iac-input`, the `IAC_INPUT_YAML` environment variable (YAML content), or `IAC_INPUT_YAML_FILE` (a path to that YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and provisions the ruleset.
3. `ruleset.go` maps proto `CloudflareRulesetRule` messages to Pulumi `cloudflare.RulesetRuleArgs` — including all action parameter sub-types (origin, response, uri, headers, from_value, overrides, cache settings).
4. Outputs are exported matching `CloudflareRulesetOutputs`.

## Engine parity note

One `action_parameters` field — `vary` (variant caching keyed on response headers) — is modeled in the
proto and provisioned by the Terraform module, but the pulumi-cloudflare SDK (v6.17.0) does not expose a
`vary` field on `RulesetRuleActionParametersArgs`, so this module omits it (see the inline note in
`module/ruleset.go` and `pkg/iac/MODULE_PARITY.md`). When a newer Pulumi SDK adds
`RulesetRuleActionParameters.Vary`, map it and remove the note. Every other ruleset field is at full
tofu↔Pulumi parity on provider v5.

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
