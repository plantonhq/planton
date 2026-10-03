# DigitalOcean VPC -- Terraform Module

Deploys a `digitalocean_vpc` from a `DigitalOceanVpc` spec: the VPC's name from `metadata.name`, the region, an optional description, and an optional immutable IP range. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanVpc`). Do not hand-edit it. The API token lives in `credentials.tf`.

## Prerequisites

- OpenTofu or Terraform 1.5+
- DigitalOcean API token (`digitalocean_token`)

## Usage

```hcl
module "vpc" {
  source = "./path/to/module"

  metadata = {
    name = "prod-network"
  }

  spec = {
    description   = "Production network for the web tier"
    region        = "nyc3"
    ip_range_cidr = "10.10.0.0/16"
  }

  digitalocean_token = var.digitalocean_token
}
```

## Behavior notes

- An unset `ip_range_cidr` is sent as null: DigitalOcean assigns a non-conflicting range, and because the provider attribute is Optional+Computed the assigned value lands in state without a perpetual diff.
- The range is immutable (ForceNew): an edit to a set range plans a full VPC REPLACEMENT, surfaced honestly -- there is no `ignore_changes` hiding it.
- `name` and `description` update in place; `region` is create-only.
- The provider serializes VPC creation per region internally (an IP-range assignment race guard), so parallel creates in one region queue rather than fail.

## Outputs

Exactly the kind's output contract, identical to the Pulumi module: `vpc_id`, `ip_range`, `urn`.
