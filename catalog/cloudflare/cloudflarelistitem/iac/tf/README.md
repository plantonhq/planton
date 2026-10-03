# CloudflareListItem Terraform Module

Terraform IaC module for writing a single entry into a Cloudflare List — an IP/CIDR, ASN, hostname, or redirect, matching the parent list's kind.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareListItemSpec
locals.tf     — Resource naming
main.tf       — cloudflare_list_item resource
outputs.tf    — outputs (item_id, list_id)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareListItem YAML manifest. For standalone use:

```hcl
module "list_item" {
  source = "./path/to/module"

  metadata = {
    name = "block-testnet"
  }

  spec = {
    account_id = "your-account-id"
    list_id    = "the-parent-list-id"
    ip         = "192.0.2.0/24"
  }
}
```

Exactly one of `ip`, `asn`, `hostname`, or `redirect` must be set, matching the parent list's kind. Item values are immutable in the provider: changing an entry replaces it. Do not also declare inline `items` on the parent CloudflareList — the two writers fight.

## Outputs

| Name | Description |
|------|-------------|
| `item_id` | Cloudflare-assigned item ID |
| `list_id` | The list the entry was written to |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
