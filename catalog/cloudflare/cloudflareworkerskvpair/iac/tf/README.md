# CloudflareWorkersKvPair Terraform Module

Terraform IaC module for provisioning a single Workers KV entry — a versioned, infrastructure-seeded key-value pair inside a KV namespace.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareWorkersKvPairSpec
locals.tf     — Resource naming
main.tf       — cloudflare_workers_kv resource
outputs.tf    — outputs (key_name, namespace_id)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareWorkersKvPair YAML manifest. For standalone use:

```hcl
module "kv_entry" {
  source = "./path/to/module"

  metadata = {
    name = "app-config-entry"
  }

  spec = {
    account_id   = "your-account-id"
    namespace_id = "your-namespace-id"
    key_name     = "feature-flags"
    value        = "{\"beta\":true}"
    metadata     = "{\"managed-by\":\"planton\"}"
  }
}
```

Account, namespace, and key all force replacement when changed — an entry's identity is the full `{account}/{namespace}/{key}` triple. Keep keys free of `/`: the terraform import ID is slash-delimited, so a key containing one cannot be imported.

## Outputs

| Name | Description |
|------|-------------|
| `key_name` | The entry's key (echoed; a Worker reads it via its KV binding) |
| `namespace_id` | The namespace holding the entry (echoed from the spec reference) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
