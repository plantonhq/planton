# CloudflareTurnstileWidget Terraform Module

Terraform IaC module for provisioning a Cloudflare Turnstile widget — a privacy-preserving CAPTCHA alternative that yields a public site key and a server-side secret.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareTurnstileWidgetSpec
locals.tf     — Resource naming
main.tf       — cloudflare_turnstile_widget resource
outputs.tf    — outputs (sitekey, secret, created_on, modified_on)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareTurnstileWidget YAML manifest. For standalone use:

```hcl
module "turnstile_widget" {
  source = "./path/to/module"

  metadata = {
    name = "login-widget"
  }

  spec = {
    account_id = "your-account-id"
    name       = "login-widget"
    domains    = ["example.com"]
    mode       = "managed"
  }
}
```

The sitekey is the API identity (not a separate id). The secret is exported as a sensitive output for `/siteverify`. Optional Enterprise flags (`bot_fight_mode`, `ephemeral_id`, `offlabel`) and `clearance_level` / `region` are sent only when set, so the provider applies its own defaults.

## Outputs

| Name | Description |
|------|-------------|
| `sitekey` | Public site key embedded in the page frontend |
| `secret` | Secret key used server-side for `/siteverify` (sensitive) |
| `created_on` | RFC3339 creation timestamp |
| `modified_on` | RFC3339 last-modified timestamp |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
