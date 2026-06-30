# Auth0Action — Terraform Module

Terraform/OpenTofu module that creates and manages Auth0 Actions with optional trigger binding.

## What It Creates

- `auth0_action` — The action resource with code, trigger, runtime, dependencies, and secrets.
- `auth0_trigger_action` (conditional) — Binds the action to its trigger when `trigger_binding` is set.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials (domain, client_id, client_secret)

## Usage

```hcl
module "auth0_action" {
  source = "."

  auth0_credential = {
    domain        = "your-tenant.auth0.com"
    client_id     = "your-client-id"
    client_secret = "your-client-secret"
  }

  metadata = {
    name = "enrich-token-claims"
  }

  spec = {
    supported_trigger = {
      id      = "post-login"
      version = "v3"
    }
    code = <<-EOT
      exports.onExecutePostLogin = async (event, api) => {
        api.idToken.setCustomClaim('https://myapp/email', event.user.email);
      };
    EOT
    deploy = true
    trigger_binding = {
      display_name = "Enrich Token Claims"
    }
  }
}
```

## Inputs

| Variable | Type | Required | Description |
|---|---|---|---|
| `auth0_credential` | object | Yes | Auth0 M2M credentials |
| `metadata` | object | Yes | Resource metadata (name, org, env) |
| `spec` | object | Yes | Action specification |

## Outputs

| Output | Description |
|---|---|
| `id` | Auth0 action identifier |
| `name` | Action name |
| `version_id` | Deployed version ID |
| `runtime` | Resolved Node.js runtime |
