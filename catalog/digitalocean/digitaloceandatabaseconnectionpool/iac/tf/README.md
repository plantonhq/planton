# Terraform Module: DigitalOcean Database Connection Pool

Provisions a PgBouncer connection pool on a DigitalOcean managed PostgreSQL cluster -- the complete `digitalocean_database_connection_pool` resource surface.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_database_connection_pool.pool` | The pool: name, mode, size, target database, optional user |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanDatabaseConnectionPoolSpec` proto: `cluster` (resolved reference -- arrives as the literal cluster UUID), `pool_name`, `mode`, `size`, `db_name`, optional `user`. Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanDatabaseConnectionPoolOutputs` contract: `cluster_id`, `pool_name`, `host`, `private_host`, `port`, and the secrets `uri`, `private_uri`, `password`.

## Behavior notes

- EVERY argument is create-only (the provider registers no update path); any change replaces the pool and drops its live connections.
- `user` left null creates DigitalOcean's inbound-user pool; the empty read-back is drift-stable.
- The provider assembles `uri`/`private_uri` from live details plus state credentials -- host/port are the cross-engine contract, never URI bytes.
- Import: `terraform import ... <cluster_id>,<pool_name>` (see `iac/import-map.yaml`).
