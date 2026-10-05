# DigitalOcean Database Connection Pool

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_database_connection_pool` resource at the pinned provider version.

## What this kind models

A PgBouncer connection pool on a DigitalOcean managed PostgreSQL cluster -- the endpoint applications should connect to when their connection count would otherwise exhaust the cluster's limit. The kind covers the provider's full argument surface:

- `cluster` -- the owning PostgreSQL cluster, by literal UUID or by reference to a `DigitalOceanDatabaseCluster`
- `pool_name` -- the pool's API identity and the "database name" clients connect to (3-63 chars)
- `mode` -- PgBouncer pooling mode: `transaction` (most apps), `session` (session-state features), `statement` (most aggressive)
- `size` -- backend server connections the pool may hold open (bounded by the cluster's size-dependent connection limit)
- `db_name` -- the logical database the pool serves (compose with `DigitalOceanDatabaseDb` by name, or `defaultdb`)
- `user` -- optional; the user the pool authenticates as. Omitted creates DigitalOcean's inbound-user pool (clients bring their own credentials)

**Everything is create-only**: the provider registers no update path for pools, so any change replaces the pool and drops its live connections.

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseConnectionPool
metadata:
  name: orders-pool
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  poolName: orders-pool
  mode: transaction
  size: 20
  dbName: orders
  user: orders-service
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `cluster_id` | UUID of the owning cluster (half of the pool's API identity) |
| `pool_name` | The pool's name (the other half; also the client-facing database name) |
| `host` | Public hostname of the pool endpoint |
| `private_host` | Private-network hostname (same-VPC access) |
| `port` | Pool port (distinct from the cluster's own port) |
| `uri` | Full public connection URI (secret; includes credentials) |
| `private_uri` | Full private-network connection URI (secret) |
| `password` | Pool user's password (secret; empty for inbound-user pools) |

## Behavior worth knowing

- **PostgreSQL only.** DigitalOcean pools exist only on PostgreSQL clusters (API-enforced).
- **Replacement semantics.** Every field is create-only upstream; a size change, mode change, or retarget replaces the pool. Schedule such changes like brief connection outages.
- **URIs are assembled, not fetched.** The provider builds `uri`/`private_uri` from live connection details plus state credentials -- treat host/port as the cross-engine contract, not URI bytes.
- **Mode choice matters.** `transaction` breaks session-state features (LISTEN/NOTIFY, session prepared statements); pick `session` for those and size for concurrent clients instead of transactions.

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
