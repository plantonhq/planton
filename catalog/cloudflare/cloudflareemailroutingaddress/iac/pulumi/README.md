# CloudflareEmailRoutingAddress Pulumi Module

Pulumi IaC module for provisioning a Cloudflare Email Routing destination address — the account-scoped, verified mailbox that routing rules forward to.

## Architecture

```
main.go (entrypoint)
  └── module/
        ├── main.go                  — Resources() orchestrator
        ├── locals.go                — Locals struct and initialization
        ├── outputs.go               — output key constants
        └── email_routing_address.go — address creation
```

## How It Works

1. `main.go` loads the `CloudflareEmailRoutingAddressIacInput` (the manifest as `target`, plus the Cloudflare `provider_config`) from the Pulumi config key `planton:iac-input`, the `IAC_INPUT_YAML` environment variable (YAML content), or `IAC_INPUT_YAML_FILE` (a path to that YAML).
2. `module.Resources()` initializes locals, creates a Cloudflare provider, and provisions the address.
3. Creating the address sends a verification email; the `verified` output stays empty until the owner clicks the link.
4. Outputs are exported matching `CloudflareEmailRoutingAddressOutputs`.

## Engine parity note

The spec's `status` field (explicit verification-state override) is provisioned by the Terraform module but NOT by this module: the pulumi-cloudflare SDK (v6.17.0) `EmailRoutingAddressArgs` carries only `AccountId` and `Email` while the terraform provider at v5.23.0 has `status`. See the `PARITY-EXCEPTION` note in `module/email_routing_address.go`; wire the field and remove the note when a newer Pulumi SDK adds it. Every other field is at full tofu↔Pulumi parity.

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
