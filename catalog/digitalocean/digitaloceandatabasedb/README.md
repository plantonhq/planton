# DigitalOcean Database Db

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_database_db` resource at the pinned provider version.

## What this kind models

An additional logical database inside a DigitalOcean managed database cluster. The resource is deliberately minimal upstream -- two create-only arguments -- and this kind models both:

- `cluster` -- the owning cluster, by literal UUID or by reference to a `DigitalOceanDatabaseCluster` (create-only)
- `database_name` -- the logical database's name and API identity (create-only; a rename REPLACES the database and DROPS its data)

Connection credentials are not part of this resource: they live on the cluster (default user) and on `DigitalOceanDatabaseUser` resources.

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseDb
metadata:
  name: orders-database
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  databaseName: orders
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `cluster_id` | UUID of the owning cluster (half of the database's API identity) |
| `database_name` | The logical database's name (the other half of its identity) |

## Behavior worth knowing

- **Renames are replacements.** Both arguments are create-only; changing `database_name` drops the old database -- and its data -- and creates an empty new one. Treat renames as migrations.
- **Reads are existence checks.** DigitalOcean's API reports only that the database exists; there is no drift to detect beyond presence.
- **Compose by name.** A `DigitalOceanDatabaseConnectionPool`'s `db_name` addresses this database by its name -- deploy the database first (reference ordering handles this in charts).

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
