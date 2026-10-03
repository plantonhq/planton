# Terraform Module: DigitalOcean Database Db

Provisions an additional logical database inside a DigitalOcean managed database cluster -- the complete `digitalocean_database_db` resource surface.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_database_db.database` | The logical database (cluster + name, both create-only) |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanDatabaseDbSpec` proto: `cluster` (resolved reference -- arrives as the literal cluster UUID) and `database_name`. Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanDatabaseDbOutputs` contract: `cluster_id` and `database_name` -- the (cluster, name) pair IS the API identity.

## Behavior notes

- Both arguments are create-only: any change replaces the logical database and drops its data.
- Import: `terraform import ... <cluster_id>,<database_name>` (see `iac/import-map.yaml`).
