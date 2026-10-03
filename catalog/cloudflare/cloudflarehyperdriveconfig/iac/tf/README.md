# CloudflareHyperdriveConfig Terraform Module

Terraform IaC module for provisioning a Cloudflare Hyperdrive config — the connection pooler and global cache a Worker binds to for low-latency access to a regional SQL database.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareHyperdriveConfigSpec
locals.tf     — Resource naming
main.tf       — cloudflare_hyperdrive_config resource
outputs.tf    — outputs (hyperdrive_id, name)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareHyperdriveConfig YAML manifest. For standalone use:

```hcl
module "hyperdrive" {
  source = "./path/to/module"

  metadata = {
    name = "app-prod-pg"
  }

  spec = {
    account_id = "your-account-id"
    name       = "app-prod-pg"
    origin = {
      database = "app_production"
      scheme   = "postgres"
      user     = "hyperdrive_user"
      host     = "db.example.com"
      port     = 5432
      password = { value = "resolved-just-in-time" }
    }
  }
}
```

Cloudflare validates the origin connection at CREATE — an unreachable host, wrong credentials, or a blocked port fail the apply, not the first query. The origin password is write-only: the API never returns it, so an imported config carries an empty password in state until the configuration re-asserts it.

## Outputs

| Name | Description |
|------|-------------|
| `hyperdrive_id` | Cloudflare-assigned Hyperdrive config ID (the Worker binding's target) |
| `name` | The config name (echoed) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
