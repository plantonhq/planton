# DigitalOcean Bucket -- Terraform Module

Deploys a `digitalocean_spaces_bucket` plus the per-bucket settings satellites (`digitalocean_spaces_bucket_cors_configuration`, `digitalocean_spaces_bucket_policy`, `digitalocean_spaces_bucket_logging`) from a `DigitalOceanBucket` spec: region and canned ACL, versioning, lifecycle rules, CORS, a JSON policy, access logging, and force-destroy. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanBucket`). Do not hand-edit it. The API token and Spaces key pair live in `credentials.tf`.

## Prerequisites

- OpenTofu or Terraform 1.5+
- DigitalOcean API token (`digitalocean_token`)
- Spaces key pair (`spaces_access_id` / `spaces_secret_key`, or the provider's env defaults `SPACES_ACCESS_KEY_ID` / `SPACES_SECRET_ACCESS_KEY`)

## Usage

```hcl
module "bucket" {
  source = "./path/to/module"

  metadata = {
    name = "app-data"
  }

  spec = {
    bucket_name         = "app-data"
    region              = "nyc3"
    access_control      = "PRIVATE"
    versioning_enabled  = true
    force_destroy       = false
    lifecycle_rules = [{
      id      = "cleanup"
      enabled = true
      abort_incomplete_multipart_upload_days = 7
      noncurrent_version_expiration = {
        days = 90
      }
    }]
  }

  digitalocean_token = var.digitalocean_token
  spaces_access_id   = var.spaces_access_id
  spaces_secret_key  = var.spaces_secret_key
}
```

## Outputs

Exactly the kind's output contract, identical to the Pulumi module:

| Output | Description |
|--------|-------------|
| `bucket_id` | The bucket's name (Spaces buckets have no UUID; import is `<region>,<name>`) |
| `endpoint` | Region-level host (`<region>.digitaloceanspaces.com`) |
| `region` | The region slug the bucket landed in |
| `bucket_domain_name` | Virtual-host FQDN (`<bucket>.<region>.digitaloceanspaces.com`) |
| `urn` | `do:space:<name>` |

## Behavior notes

- `region` is omitted when unset — the provider applies its own default (`nyc3`). The unspecified enum name is never sent as a slug, and there is no silent fallback in the module.
- CORS, policy, and logging resources are created only when configured. Their `bucket` argument is wired to the bucket this module creates; their `region` is the spec's region (the spec requires it whenever a satellite is set).
- CORS uses the standalone `digitalocean_spaces_bucket_cors_configuration` resource. The bucket's inline `cors_rule` is deprecated at the pin and is never written.
- `force_destroy` empties the bucket (every object and every version) before destroy when true.
- See the kind [GUIDE](../../GUIDE.md).
