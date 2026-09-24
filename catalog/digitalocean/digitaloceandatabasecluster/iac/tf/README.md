# DigitalOcean Database Cluster -- Terraform Module

Deploys a `digitalocean_database_cluster` from a `DigitalOceanDatabaseCluster` spec: every engine DigitalOcean offers, node topology, VPC placement, custom storage (the provider's bare-MiB string, converted from the spec's GiB), storage autoscale, maintenance window, backup-restore provisioning, engine-conditional tuning, project placement, and tags. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanDatabaseCluster`). Do not hand-edit it. The API token lives in `credentials.tf`.

Users, logical databases, connection pools, replicas, firewall rules, and per-engine config parameters are separate DigitalOcean resources, not part of this module.

## Prerequisites

- OpenTofu or Terraform 1.5+
- DigitalOcean API token (`digitalocean_token`)

## Usage

```hcl
module "database_cluster" {
  source = "./path/to/module"

  metadata = {
    name = "app-db"
  }

  spec = {
    cluster_name   = "app-db"
    engine         = "pg"
    engine_version = "16"
    region         = "nyc3"
    size_slug      = "db-s-2vcpu-4gb"
    node_count     = 3
    vpc            = "b5648f9e-a28a-4760-bb87-b2fad07ae295"
    storage_gib    = 100
    maintenance_window = {
      day  = "sunday"
      hour = "02:00"
    }
    storage_autoscale = {
      enabled           = true
      threshold_percent = 80
    }
  }

  digitalocean_token = var.digitalocean_token
}

output "host" {
  value = module.database_cluster.host
}
```

## Outputs

Exactly the kind's stack-output contract, identical to the Pulumi module:

| Output | Description |
|--------|-------------|
| `cluster_id` | The cluster UUID (import id for `digitalocean_database_cluster`) |
| `connection_uri` | Full public connection URI (sensitive) |
| `host` / `port` | Public connection endpoint |
| `database_user` / `database_password` | Default user credentials (sensitive) |
| `private_host` / `private_uri` | Private-network endpoint (URI sensitive) |
| `database_name` | Default database name |
| `ui_host` / `ui_port` / `ui_uri` / `ui_database` / `ui_user` / `ui_password` | OpenSearch Dashboards details (OpenSearch only; URI/password sensitive) |

## Behavior notes

- `sql_mode` and `eviction_policy` are passed only when set; the provider rejects them at plan time on engines they don't apply to (the spec's validation rules prevent that pairing earlier).
- `storage_size_mib` is only rendered when `storage_gib` is set, so growing `size_slug` without a custom storage value correctly adopts the new slug's default disk.
- Changing `engine_version` performs an in-place major upgrade; changing `region` performs a live migration. See the kind [GUIDE](../../GUIDE.md).
- Tags are `spec.tags` plus the six standard Planton labels rendered as `key:value` strings (the identical set the Pulumi module applies). DigitalOcean caps the comma-joined tag string at 255 characters (measured 2026-09-17), so the resource carries a `precondition` against `local.tags_combined_budget` that fails the plan with the exact arithmetic before anything is created — the same number and message as the Pulumi module.
