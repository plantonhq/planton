# DigitalOcean Database Cluster

Managed databases on DigitalOcean: one Planton component models the full `digitalocean_database_cluster` resource — every engine DigitalOcean offers (PostgreSQL, MySQL, Valkey, MongoDB, Kafka, OpenSearch — plus Redis for adopting existing clusters), node topology and sizing, VPC-private networking, custom storage with automatic growth, weekly maintenance windows, restore-from-backup provisioning, engine-specific tuning, project placement, and tags.

## What this component models

The spec maps one-to-one onto DigitalOcean's managed database cluster:

| Spec field | What it controls |
|---|---|
| `clusterName` | The cluster's name in DigitalOcean (up to 64 characters) |
| `engine` | `pg`, `mysql`, `valkey`, `mongodb`, `kafka`, or `opensearch` for new clusters; `redis` only adopts an existing cluster (DigitalOcean no longer creates Redis) |
| `engineVersion` | A version DigitalOcean currently offers for the engine, exactly as `GET /v2/databases/options` lists it (`"16"` for PostgreSQL, `"8.4"` for MySQL, `"8"` for Valkey, `"4.2"` for Kafka, `"2.19"` for OpenSearch, `"8.0"` for MongoDB as of 2026-09-16); a version not on that list is rejected at create. Changing it performs an in-place major upgrade — DigitalOcean never downgrades |
| `region` | Data-center region; changing it live-migrates the cluster |
| `sizeSlug` | Per-node CPU/memory (`db-s-1vcpu-1gb`, `db-s-2vcpu-4gb`, ...); changing it resizes in place |
| `nodeCount` | Engine-specific: 1–3 for most engines, 3+ for Kafka, up to 15 for OpenSearch |
| `vpc` | Optional private-network placement — a literal UUID or a reference to a `DigitalOceanVpc`; create-only |
| `storageGib` | Optional custom disk beyond the slug's default; increase-only |
| `storageAutoscale` | Optional automatic disk growth (threshold percent, increment) |
| `maintenanceWindow` | Optional weekly slot (`day`, `hour` in UTC) for automatic updates |
| `backupRestore` | Optional provision-from-backup of an existing cluster (write-once, create-time only) |
| `evictionPolicy` | Redis/Valkey only: key eviction under memory pressure |
| `sqlMode` | MySQL only: comma-separated SQL modes |
| `projectId` | Optional DigitalOcean project placement (UUID); create-only |
| `tags` | Your tags, applied alongside the six standard Planton labels; DigitalOcean caps the comma-joined total at 255 characters and both provisioners fail fast with the arithmetic when a long `metadata.name` or `metadata.id` spends it |

Engine pairing is validated at manifest time: `sqlMode` with anything but MySQL, or `evictionPolicy` with anything but Redis/Valkey, is rejected before any provisioner runs — the same rules DigitalOcean enforces server-side.

Database users, logical databases, connection pools, read replicas, firewall rules, and per-engine config parameters are separate DigitalOcean resources with their own lifecycles, not part of the cluster resource.

## Quick start

A production PostgreSQL cluster with failover, private networking, and managed disk growth:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseCluster
metadata:
  name: app-db
  org: acme-corp
  env: prod
spec:
  clusterName: app-db
  engine: pg
  engineVersion: "16"
  region: nyc3
  sizeSlug: db-s-2vcpu-4gb
  nodeCount: 3
  vpc:
    valueFrom:
      kind: DigitalOceanVpc
      name: app-network
      fieldPath: status.outputs.vpc_id
  storageAutoscale:
    enabled: true
    thresholdPercent: 80
  maintenanceWindow:
    day: sunday
    hour: "02:00"
```

```shell
planton apply -f app-db.yaml
```

A single-node Valkey cache with an LRU eviction policy:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseCluster
metadata:
  name: session-cache
spec:
  clusterName: session-cache
  engine: valkey
  engineVersion: "8"
  region: nyc3
  sizeSlug: db-s-1vcpu-1gb
  nodeCount: 1
  evictionPolicy: allkeys_lru
```

## Outputs

Both provisioners export the identical output set:

| Output | Description |
|---|---|
| `cluster_id` | The cluster's UUID |
| `connection_uri` | Full public connection URI (credentials included) — sensitive |
| `host` / `port` | Public connection endpoint |
| `database_user` / `database_password` | Default user credentials — password is sensitive |
| `private_host` / `private_uri` | Private-network endpoint, reachable from the same VPC |
| `database_name` | The default database's name |
| `ui_host`, `ui_port`, `ui_uri`, `ui_database`, `ui_user`, `ui_password` | OpenSearch Dashboards connection details (populated for OpenSearch clusters only) |

## Behavior worth knowing

- **Storage never shrinks.** `storageGib` can only grow. Growing `sizeSlug` with `storageGib` unset adopts the new size's default base storage.
- **Version changes upgrade in place.** Setting a higher `engineVersion` runs a major version upgrade on the live cluster. There is no downgrade path.
- **Region changes migrate live.** The cluster stays up while DigitalOcean moves it; plan for elevated latency during the move.
- **Removing `evictionPolicy` resets to `noeviction`** rather than leaving the last policy in place.
- **`backupRestore` acts only at creation.** DigitalOcean never reports it back; changing it on an existing cluster does nothing.
- **`storageAutoscale` deploys on both provisioners.** Keep `incrementGib` at or below the size slug's maximum plan storage (30 GiB on `db-s-1vcpu-1gb`) — the API refuses a larger step at create time. See the [GUIDE](GUIDE.md).

See `GUIDE.md` for operational judgment (sizing, engine selection, upgrade practice) and `catalog.md` for the deployment-store page.

## Module layout

- `v1alpha1/` — the versioned contract: `spec.proto`, `outputs.proto`, validation tests, generated reference
- `iac/tf/` and `iac/pulumi/` — the two provisioner modules implementing the same contract with identical outputs
- `iac/provider-parity.yaml` — the recorded mapping judgment against the pinned provider
- `iac/import-map.yaml` — how an existing cluster's identity derives for import
- `presets/` — ready-to-deploy starting points (PostgreSQL HA/dev, Valkey, Kafka, OpenSearch)
- `e2e/` — test profile, canonical manifests, and live-lane scenarios

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
