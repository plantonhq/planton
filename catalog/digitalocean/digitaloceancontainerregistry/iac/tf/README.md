# DigitalOcean Container Registry -- Terraform Module

Deploys a `digitalocean_container_registry` from a `DigitalOceanContainerRegistry` spec, plus a `digitalocean_container_registry_docker_credentials` when (and only when) the spec's `docker_credentials` block is set. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanContainerRegistry`). Do not hand-edit it. The API token lives in `credentials.tf`.

## Prerequisites

- OpenTofu or Terraform 1.5+
- DigitalOcean API token (`digitalocean_token`)
- No existing registry on the account (DigitalOcean allows exactly one)

## Usage

```hcl
module "registry" {
  source = "./path/to/module"

  metadata = {
    name = "acme-registry"
  }

  spec = {
    name              = "acme-registry"
    subscription_tier = "basic"
    region            = "nyc3"
    docker_credentials = {
      write          = true
      expiry_seconds = 2592000
    }
  }

  digitalocean_token = var.digitalocean_token
}
```

## Behavior notes

- The proto enum value names ARE the DigitalOcean tier slugs (`starter`/`basic`/`professional`) and region slugs -- both pass through unmodified. An unset region is sent as null (never an empty string, which the provider rejects) so DigitalOcean chooses.
- The credentials resource is conditional (`count`) on the spec block: no block, no long-lived token. Its `expiry_seconds` defers to the provider default (the API maximum, ~50 years) when unset. Neither knob is recoverable from the API afterwards, and the resource's importer is DEFECTIVE at this pin -- see `../import-map.yaml`.
- The `docker_credentials` output is marked `sensitive` and is an empty string when the block is unset -- identical to the Pulumi module.

## Outputs

Exactly the kind's output contract, identical to the Pulumi module: `registry_name`, `server_url`, `endpoint`, `region`, `docker_credentials` (sensitive), `credential_expiration_time`.
