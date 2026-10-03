# CloudflareEmailRoutingAddress Terraform Module

Terraform IaC module for provisioning a Cloudflare Email Routing destination address — the account-scoped, verified mailbox that routing rules forward to.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareEmailRoutingAddressSpec
locals.tf     — Resource naming
main.tf       — cloudflare_email_routing_address resource
outputs.tf    — outputs (address_id, email, verified, created)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareEmailRoutingAddress YAML manifest. For standalone use:

```hcl
module "destination_address" {
  source = "./path/to/module"

  metadata = {
    name = "ops-destination"
  }

  spec = {
    account_id = "your-account-id"
    email      = "ops@example.com"
  }
}
```

Creating the address sends a verification email to the mailbox; the address is usable as a forwarding target only after its owner clicks the link (the `verified` output stays empty until then). The optional `status` field is an explicit verification-state override — Cloudflare permits non-admin callers only to flip a verified address back to `"unverified"`.

## Outputs

| Name | Description |
|------|-------------|
| `address_id` | Cloudflare-assigned destination-address ID |
| `email` | The destination email (echoed; referenced by routing rules) |
| `verified` | RFC3339 verification timestamp, empty until the owner verifies |
| `created` | RFC3339 creation timestamp |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
