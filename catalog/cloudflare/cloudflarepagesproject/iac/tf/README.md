# CloudflarePagesProject Terraform Module

Terraform IaC module for provisioning a Cloudflare Pages project — build config, optional git source, per-environment deployment configuration, and attached custom domains.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflarePagesProjectSpec
locals.tf     — build/source/deployment_configs transforms; domains_map for_each
main.tf       — cloudflare_pages_project + cloudflare_pages_domain resources
outputs.tf    — outputs (project_name, subdomain, domains, created_on)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflarePagesProject YAML manifest. For standalone use:

```hcl
module "pages_project" {
  source = "./path/to/module"

  metadata = {
    name = "marketing-site"
  }

  spec = {
    account_id         = "your-account-id"
    name               = "marketing-site"
    production_branch  = "main"
  }
}
```

The project name is the identity (Pages has no separate id). When only one of preview/production deployment configs is supplied, the module mirrors it to both — Cloudflare rejects a project whose environments are configured inconsistently. Secret env vars inside `deployment_configs` do not survive import: the API never returns `secret_text` values. Custom domains are a companion resource (`cloudflare_pages_domain`) keyed by hostname.

## Outputs

| Name | Description |
|------|-------------|
| `project_name` | The project name (downstream resources reference this) |
| `subdomain` | The project's `*.pages.dev` subdomain |
| `domains` | Custom domains attached to the project |
| `created_on` | RFC3339 creation timestamp |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
