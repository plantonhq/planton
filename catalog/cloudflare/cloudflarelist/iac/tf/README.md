# CloudflareList Terraform Module

Terraform IaC module for provisioning an account-scoped Cloudflare List — a named collection referenced from rule expressions (WAF, custom rules, Bulk Redirect).

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareListSpec
locals.tf     — Resource naming
main.tf       — cloudflare_list resource (never declares inline items)
outputs.tf    — outputs (list_id, name, kind)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareList YAML manifest. For standalone use:

```hcl
module "list" {
  source = "./path/to/module"

  metadata = {
    name = "blocked-cidrs"
  }

  spec = {
    account_id = "your-account-id"
    kind       = "ip"
    name       = "blocked_cidrs"
  }
}
```

The list is the container; entries are managed as CloudflareListItem resources. This module never declares inline `items` — mixing inline items with CloudflareListItem makes the provider treat them as competing writers. `kind` and `name` are immutable (changing either replaces the list). List names must match `^[a-zA-Z][a-zA-Z0-9_]*$` (no hyphens) because they appear in rule expressions as `$name`.

## Outputs

| Name | Description |
|------|-------------|
| `list_id` | Cloudflare-assigned list ID (what a CloudflareListItem references) |
| `name` | The list name (the identifier used in rule expressions) |
| `kind` | The list kind (`ip`, `redirect`, `hostname`, or `asn`) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
