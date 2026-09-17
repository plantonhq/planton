# DigitalOcean Database Replica

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_database_replica` resource at the pinned provider version.

## What this component models

A single-node read-only replica of a DigitalOcean managed database cluster (PostgreSQL and MySQL primaries support replicas), in the primary's region or a different one. The component covers the provider's full argument surface:

- `cluster` -- the primary, by literal UUID or by reference to a `DigitalOceanDatabaseCluster` (create-only)
- `replica_name` -- the replica's API identity within the cluster (create-only)
- `region` -- the replica's region; same as the primary for a local read endpoint, different for a cross-region replica (create-only)
- `size` -- the replica's node slug; must be at least the primary's size; grows in place, never shrinks
- `vpc` -- optional private-network placement in the REPLICA's region, by UUID or `DigitalOceanVpc` reference (create-only)
- `storage_size_mib` -- optional custom disk; grows in place with size; must stay at least the primary's storage
- `tags` -- CREATE-ONLY upstream: a retag REPLACES the replica. DigitalOcean caps the combined tags (joined by commas) at 255 characters, the same rule as the primary, and the six Planton label tags carrying `metadata.name`/`metadata.id` count toward it; both provisioners fail fast with the arithmetic when the budget is crossed

`region` and `size` are REQUIRED here although the upstream provider marks them optional: the provider reads both back but never computes them, so omitted values drift on the next apply -- and region's drift schedules a full replica REPLACEMENT. Explicit values make that failure class unrepresentable; "inherit from the primary" is expressed by writing the primary's values.

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseReplica
metadata:
  name: orders-read-replica
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  replicaName: orders-read-replica
  region: nyc3
  size: db-s-1vcpu-1gb
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `replica_id` | UUID of the replica itself |
| `cluster_id` | UUID of the primary cluster |
| `replica_name` | The replica's name (its API identity within the cluster) |
| `host` / `private_host` | Public / private-network hostnames of the replica endpoint |
| `port` | Port the replica listens on |
| `database` / `user` | Default database and username served by the replica |
| `password` | Default user's password (secret) |
| `uri` / `private_uri` | Full connection URIs (secrets; include credentials) |

## Behavior worth knowing

- **A replica is a second bill.** It is a full single-node database of its own slug, billed hourly from creation -- see `cost.yaml`.
- **Only size and storage change in place.** Everything else -- name, region, VPC, and TAGS -- replaces the replica (a fresh seed from the primary; no primary data is ever at risk).
- **Creation waits on the primary's first backup.** DigitalOcean retries through 412 responses while it completes; expect create times comparable to the primary's own provisioning.
- **Read-only, primary's credentials.** The replica serves the primary's users and databases; it has no user management of its own.

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
