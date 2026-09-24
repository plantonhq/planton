# DigitalOceanDatabaseCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanDatabaseClusterSpec models the full digitalocean_database_cluster
resource surface: engine/version/size/region/node topology, private
networking, custom storage with autoscale, maintenance windows,
backup-restore provisioning, engine-conditional tuning (sql_mode,
eviction_policy), project placement, and tags.

Per-engine tuning parameters (PostgreSQL/MySQL/Redis/Kafka/OpenSearch/
Valkey/MongoDB config settings) are applied through DigitalOcean's
separate per-engine config APIs, not through the cluster resource, and
are therefore not part of this spec. Database users, logical databases,
connection pools, replicas, and firewall rules are likewise separate
DigitalOcean resources.

## Example

```yaml
# Example DigitalOceanDatabaseCluster manifests.
#
# Deploy with: planton apply -f manifest.yaml
#
# The first document is the smallest real cluster (single-node PostgreSQL
# with defaults). The second exercises the full surface on MySQL: VPC
# placement, custom storage, a maintenance window, storage autoscale,
# sql_mode, project placement, and tags.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseCluster
metadata:
  name: example-postgres
spec:
  clusterName: example-postgres
  engine: pg
  engineVersion: "16"
  region: nyc3
  sizeSlug: db-s-1vcpu-1gb
  nodeCount: 1
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseCluster
metadata:
  name: example-mysql
spec:
  clusterName: example-mysql
  engine: mysql
  engineVersion: "8.4"
  region: nyc3
  sizeSlug: db-s-2vcpu-4gb
  nodeCount: 2
  vpc:
    value: b5648f9e-a28a-4760-bb87-b2fad07ae295
  storageGib: 100
  maintenanceWindow:
    day: sunday
    hour: "02:00"
  storageAutoscale:
    enabled: true
    thresholdPercent: 80
    incrementGib: 50
  sqlMode: "ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"
  projectId:
    value: 3f2a9c6e-8d41-4b7a-9f0e-1c5d7b2a4e68
  tags:
    - env:example
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.clusterName` | `string` | yes |  |  |
| `spec.engine` | `enum` | yes |  |  |
| `spec.engineVersion` | `string` | yes |  |  |
| `spec.region` | `enum` | yes |  |  |
| `spec.sizeSlug` | `string` | yes |  |  |
| `spec.nodeCount` | `uint32` | yes |  |  |
| `spec.vpc` | `string \| valueFrom` |  |  | DigitalOceanVpc (`status.outputs.vpc_id`) |
| `spec.storageGib` | `uint32` |  |  |  |
| `spec.maintenanceWindow` | `DigitalOceanDatabaseClusterMaintenanceWindow` |  |  |  |
| `spec.maintenanceWindow.day` | `string` | yes |  |  |
| `spec.maintenanceWindow.hour` | `string` | yes |  |  |
| `spec.backupRestore` | `DigitalOceanDatabaseClusterBackupRestore` |  |  |  |
| `spec.backupRestore.databaseName` | `string` | yes |  |  |
| `spec.backupRestore.backupCreatedAt` | `string` |  |  |  |
| `spec.storageAutoscale` | `DigitalOceanDatabaseClusterStorageAutoscale` |  |  |  |
| `spec.storageAutoscale.enabled` | `bool` | yes |  |  |
| `spec.storageAutoscale.thresholdPercent` | `uint32` |  |  |  |
| `spec.storageAutoscale.incrementGib` | `uint32` |  |  |  |
| `spec.evictionPolicy` | `string` |  |  |  |
| `spec.sqlMode` | `string` |  |  |  |
| `spec.projectId` | `string \| valueFrom` |  |  | DigitalOceanProject (`status.outputs.project_id`) |
| `spec.tags` | `[]string` |  |  |  |

## Field Details

### spec.clusterName

`string` · required

A human-readable name for the database cluster.
This name is the cluster's identifier in DigitalOcean.

- rule: {"required":true,"string":{"maxLen":"64"}}

### spec.engine

`enum` · required

The database engine for the cluster. Enum value names are exactly the
DigitalOcean engine slugs (pg, mysql, redis, mongodb, kafka,
opensearch, valkey). DigitalOcean no longer creates Redis clusters --
`redis` only adopts one that already exists; new caches are `valkey`.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_database_engine_unspecified`
- `pg`
- `mysql`
- `redis`
- `mongodb`
- `kafka`
- `opensearch`
- `valkey`

### spec.engineVersion

`string` · required

The engine version for the cluster, exactly as DigitalOcean lists it
for the engine in `GET /v2/databases/options` -- a major for some
engines ("16" for PostgreSQL, "8" for Valkey), major.minor for others
("8.4" for MySQL, "4.2" for Kafka, "2.19" for OpenSearch, "8.0" for
MongoDB, as offered on 2026-09-16). The offer list moves: DigitalOcean
retires versions on a published schedule and rejects any value not on
it at create (422 "invalid cluster engine version" -- a bare "8" for
MySQL fails today). Changing the version on an existing cluster
performs an in-place major version upgrade; DigitalOcean does not
support downgrades.

- rule: {"required":true,"string":{"pattern":"^[0-9]+(\\.[0-9]+)?$"}}

### spec.region

`enum` · required

The DigitalOcean region where the cluster will be created.
Changing the region on an existing cluster performs a live migration.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_region_unspecified` -- 0: default / unspecified region
- `nyc3` -- new york 3
- `sfo3` -- san francisco 3
- `fra1` -- frankfurt 1
- `sgp1` -- singapore 1
- `lon1` -- london 1
- `tor1` -- toronto 1
- `blr1` -- bangalore 1
- `ams3` -- amsterdam 3
- `nyc1` -- new york 1
- `nyc2` -- new york 2
- `sfo2` -- san francisco 2
- `syd1` -- sydney 1
- `atl1` -- atlanta 1

### spec.sizeSlug

`string` · required

The slug identifier for the cluster's node size (e.g. "db-s-2vcpu-4gb").
Defines the CPU/memory resources for each node; changing it resizes the
cluster in place.

- rule: {"required":true}

### spec.nodeCount

`uint32` · required

The number of nodes in the cluster. Valid counts are engine-specific
and enforced by the DigitalOcean API: PostgreSQL/MySQL/Redis/Valkey/
MongoDB accept 1-3 (1 primary plus up to 2 standbys), Kafka requires
at least 3, OpenSearch accepts 1-15.

- rule: {"required":true,"uint32":{"gte":1}}

### spec.vpc

`string | valueFrom`

(Optional) Reference to a DigitalOcean VPC for the database cluster.
If provided, the cluster is created within the specified private
network. Use a literal VPC UUID or a reference to a DigitalOceanVpc
resource. Cannot be changed after creation.

- references: DigitalOceanVpc (`status.outputs.vpc_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.storageGib

`uint32`

(Optional) Custom storage size in GiB for the cluster. If not set, the
default storage for the chosen size_slug is used. Storage can only be
increased, never decreased, and growing size_slug with this unset
adopts the new size's default base storage.

### spec.maintenanceWindow

`DigitalOceanDatabaseClusterMaintenanceWindow`

(Optional) Weekly maintenance window for automatic updates.

### spec.maintenanceWindow.day

`string` · required

Day of the week, e.g. "monday". Case-insensitive.

- rule: {"required":true,"string":{"pattern":"^(?i)(monday|tuesday|wednesday|thursday|friday|saturday|sunday)$"}}

### spec.maintenanceWindow.hour

`string` · required

Start hour of the window in UTC, "HH:MM" (seconds optional), for
example "02:00".

- rule: {"required":true,"string":{"pattern":"^[0-9]{2}:[0-9]{2}(:[0-9]{2})?$"}}

### spec.backupRestore

`DigitalOceanDatabaseClusterBackupRestore`

(Optional) Provision this cluster by restoring a backup of an existing
cluster. Consumed only at creation time; DigitalOcean never reports it
back afterward.

### spec.backupRestore.databaseName

`string` · required

Name of the existing cluster whose backup to restore from.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.backupRestore.backupCreatedAt

`string`

(Optional) ISO8601 timestamp of the backup to restore. When unset, the
most recent backup is used.

### spec.storageAutoscale

`DigitalOceanDatabaseClusterStorageAutoscale`

(Optional) Automatic storage growth when the disk approaches capacity.

### spec.storageAutoscale.enabled

`bool` · required

Whether automatic storage growth is enabled.

- rule: {"required":true}

### spec.storageAutoscale.thresholdPercent

`uint32`

(Optional) Disk usage percentage that triggers growth, 15-95. When
unset, DigitalOcean applies its default threshold.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","uint32":{"lte":95,"gte":15}}

### spec.storageAutoscale.incrementGib

`uint32`

(Optional) Growth step size in GiB, minimum 10, rounded to the nearest
10 GiB. When unset, DigitalOcean auto-calculates 25% of current size
(min 50 GiB, max 1024 GiB).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","uint32":{"gte":10}}

### spec.evictionPolicy

`string`

(Optional) Key eviction policy for Redis or Valkey clusters. Known
values: noeviction, allkeys_lru, allkeys_random, volatile_lru,
volatile_random, volatile_ttl. Removing a previously set policy resets
the cluster to noeviction.

### spec.sqlMode

`string`

(Optional) Comma-separated SQL modes for MySQL clusters, for example
"ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION".

### spec.projectId

`string | valueFrom`

(Optional) The project the cluster is created in. Reference a
DigitalOceanProject resource (the default wiring resolves its
project_id output) or pass a literal project UUID. Cannot be changed
after creation.

- references: DigitalOceanProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.tags

`[]string`

(Optional) Tags applied to the cluster in DigitalOcean, in addition to
the standard Planton labels both provisioners always apply
(`planton-ai_resource:true`, `planton-ai_name:<metadata.name>`,
`planton-ai_kind:DigitalOceanDatabaseCluster`,
`planton-ai_organization:<org>`, `planton-ai_environment:<env>`,
`planton-ai_id:<metadata.id>`). DigitalOcean caps a database cluster's
COMBINED tags -- every tag joined by commas -- at 255 characters
(measured 2026-09-17: the API answers `422 combined tags cannot exceed
255 characters`). The label tags alone cost about 133 characters plus
the length of metadata.name and metadata.id, so a long resource name
leaves little room here; both provisioners check the budget before
creating anything and fail with the exact arithmetic.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

## Validation Rules

- `sql_mode_mysql_only`: sql_mode applies only to MySQL clusters
- `eviction_policy_caching_only`: eviction_policy applies only to Redis or Valkey clusters

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanDatabaseCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_id` | `string` | The unique identifier (UUID) of the created database cluster. |
| `status.outputs.connection_uri` | `string` | The full public connection URI for the database cluster (including credentials and database name). |
| `status.outputs.host` | `string` | The public hostname at which the database cluster is accessible. |
| `status.outputs.port` | `uint32` | The network port that the database cluster is listening on. |
| `status.outputs.database_user` | `string` | The username for the cluster's default database user. |
| `status.outputs.database_password` | `string` | The password for the cluster's default database user. |
| `status.outputs.private_host` | `string` | The private-network hostname, reachable from resources in the same VPC. |
| `status.outputs.private_uri` | `string` | The full private-network connection URI (including credentials and database name), reachable from resources in the same VPC. |
| `status.outputs.database_name` | `string` | The name of the cluster's default database. |
| `status.outputs.ui_host` | `string` | OpenSearch only: hostname of the OpenSearch Dashboards endpoint. |
| `status.outputs.ui_port` | `uint32` | OpenSearch only: port of the OpenSearch Dashboards endpoint. |
| `status.outputs.ui_uri` | `string` | OpenSearch only: full connection URI for OpenSearch Dashboards (including credentials). |
| `status.outputs.ui_database` | `string` | OpenSearch only: default database of the OpenSearch Dashboards connection. |
| `status.outputs.ui_user` | `string` | OpenSearch only: username for OpenSearch Dashboards. |
| `status.outputs.ui_password` | `string` | OpenSearch only: password for OpenSearch Dashboards. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpc` | DigitalOceanVpc | `status.outputs.vpc_id` |
| `spec.projectId` | DigitalOceanProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanApp | `spec.databases[].clusterName` | `spec.cluster_name` |
| DigitalOceanDatabaseConnectionPool | `spec.cluster` | `status.outputs.cluster_id` |
| DigitalOceanDatabaseDb | `spec.cluster` | `status.outputs.cluster_id` |
| DigitalOceanDatabaseFirewall | `spec.cluster` | `status.outputs.cluster_id` |
| DigitalOceanDatabaseKafkaSchema | `spec.cluster` | `status.outputs.cluster_id` |
| DigitalOceanDatabaseKafkaTopic | `spec.cluster` | `status.outputs.cluster_id` |
| DigitalOceanDatabaseReplica | `spec.cluster` | `status.outputs.cluster_id` |
| DigitalOceanDatabaseUser | `spec.cluster` | `status.outputs.cluster_id` |
| DigitalOceanMonitorAlert | `spec.databaseClusterIds` | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
