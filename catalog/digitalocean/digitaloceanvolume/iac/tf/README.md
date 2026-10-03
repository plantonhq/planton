# DigitalOcean Volume -- Terraform Module

Deploys a `digitalocean_volume` from a `DigitalOceanVolume` spec: the full provider argument surface -- name, region, size, description, one-time filesystem formatting with an optional label, snapshot source, and tags. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanVolume`). Do not hand-edit it. The API token lives in `credentials.tf`.

## Prerequisites

- OpenTofu or Terraform 1.5+
- DigitalOcean API token (`digitalocean_token`)

## Usage

```hcl
module "volume" {
  source = "./path/to/module"

  metadata = {
    name = "postgres-data"
  }

  spec = {
    volume_name              = "postgres-data"
    region                   = "nyc3"
    size_gib                 = 500
    filesystem_type          = "xfs"
    initial_filesystem_label = "pgdata"
    tags                     = ["env:prod"]
  }

  digitalocean_token = var.digitalocean_token
}
```

## Behavior notes

- The spec enum's value names ARE the provider's filesystem strings (`ext4`/`xfs`); `unformatted` (or an omitted value) stays null -- no lookup table. The label is sent only when a filesystem type is being formatted.
- Size can only be EXPANDED: the provider rejects a shrink at plan time. Description is create-only at this pin -- a change plans a REPLACEMENT.
- `snapshot_id`, `initial_filesystem_type`, and `initial_filesystem_label` act at creation and are never read back by the API (recorded config-only import tolerances).
- Tags are the union of `spec.tags` and the standard Planton labels rendered as `key:value` -- the exact set the Pulumi module applies.

## Outputs

Exactly the kind's output contract, identical to the Pulumi module: `volume_id`, `urn`.
