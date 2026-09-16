# DigitalOcean Database Cluster -- Pulumi Module

Deploys a `digitalocean:index/databaseCluster:DatabaseCluster` from a `DigitalOceanDatabaseCluster` stack input: every engine DigitalOcean offers, node topology, VPC placement, custom storage (the provider's bare-MiB string, converted from the spec's GiB), maintenance window, backup-restore provisioning, engine-conditional tuning, project placement, and tags. Bridge SDK pin is `pulumi-digitalocean/sdk/v4 v4.53.0`.

Users, logical databases, connection pools, replicas, firewall rules, and per-engine config parameters are separate DigitalOcean resources, not part of this module.

## Module structure

- `main.go` -- Pulumi program entry point reading the stack input
- `module/main.go` -- `Resources()`: locals, provider, cluster
- `module/locals.go` -- stack-input references and the standard Planton label map
- `module/database_cluster.go` -- the cluster resource and stack-output exports
- `module/outputs.go` -- output key constants (the kind's outputs.proto contract)

## Outputs

Exactly the kind's stack-output contract, identical to the Terraform module: `cluster_id`, `connection_uri`, `host`, `port`, `database_user`, `database_password`, `private_host`, `private_uri`, `database_name`, and the OpenSearch-only `ui_host` / `ui_port` / `ui_uri` / `ui_database` / `ui_user` / `ui_password`. The URI and password outputs are Pulumi secret outputs.

## Behavior notes

- **PARITY-EXCEPTION**: `spec.storage_autoscale` is modeled and the Terraform module wires it, but the pinned Pulumi DigitalOcean SDK (v4.53.0, re-verified against `DatabaseClusterArgs`) has no `storage_autoscale` field on DatabaseCluster — this module fails loudly when it is set rather than silently dropping it. Re-evaluate when the SDK exposes storage_autoscale.
- Tags are the user's `spec.tags` plus the standard Planton labels rendered as `key:value` strings — the identical set the Terraform module applies.
- `sql_mode` and `eviction_policy` are passed only when set; the spec's validation rules enforce the engine pairing before any deploy.
- Changing `engine_version` performs an in-place major upgrade; changing `region` performs a live migration. See the kind [GUIDE](../../GUIDE.md).
