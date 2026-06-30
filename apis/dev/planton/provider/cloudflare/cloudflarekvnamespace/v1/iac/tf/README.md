# Cloudflare KV Namespace - OpenTofu Module

Provisions a Cloudflare Workers KV namespace (`cloudflare_workers_kv_namespace`) and
exports its identifier for binding to Workers.

## Inputs

| Variable | Description |
|----------|-------------|
| `metadata` | Resource metadata (name, labels, ...). |
| `spec.namespace_name` | Human-readable title for the KV namespace (max 64 chars). |
| `spec.account_id` | Cloudflare account ID (32 hex characters) that owns the namespace. |
| `spec.ttl_seconds` | Spec-level default TTL hint; not represented on the KV namespace resource. |
| `spec.description` | Spec-level description; not represented on the KV namespace resource. |

## Outputs

| Output | Description |
|--------|-------------|
| `namespace_id` | The unique identifier of the created KV namespace. |

## Provider

```hcl
terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
  }
}

provider "cloudflare" {
  # Automatically uses CLOUDFLARE_API_TOKEN environment variable
}
```
