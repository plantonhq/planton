# DigitalOceanDatabaseReplica

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanDatabaseReplicaSpec models the full
digitalocean_database_replica resource surface: a single-node read-only
replica of a DigitalOcean managed database cluster (PostgreSQL and MySQL
primaries support replicas), in the primary's region or a different one.

region and size are REQUIRED here even though the upstream provider
marks them optional: the provider always reads both back from the API
but never computes them, so a configuration that omits them drifts on
the next apply -- and because region is create-only upstream, that drift
schedules a full REPLICA REPLACEMENT. Requiring explicit values makes
that failure class unrepresentable; "inherit from the primary" is
expressed by simply writing the primary's region and size.

## Example

```yaml
# Reference manifests for DigitalOceanDatabaseReplica -- protovalidate-
# valid, embedded as the reference page's Example block, and the documents
# the offline tofu plans render. Two documents: a same-region hot standby
# and a cross-region replica with VPC placement.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseReplica
metadata:
  name: orders-read-replica
spec:
  # Literal cluster UUID; use valueFrom to reference a
  # DigitalOceanDatabaseCluster resource instead.
  cluster:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  replicaName: orders-read-replica
  region: nyc3
  size: db-s-1vcpu-1gb
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseReplica
metadata:
  name: orders-eu-replica
spec:
  cluster:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  replicaName: orders-eu-replica
  region: ams3
  size: db-s-1vcpu-2gb
  # Literal VPC UUID; use valueFrom to reference a DigitalOceanVpc
  # resource instead.
  vpc:
    value: dddddddd-eeee-ffff-0000-111111111111
  storageSizeMib: 30720
  tags:
    - env:prod
    - read-scaling
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.cluster` | `string \| valueFrom` | yes |  | DigitalOceanDatabaseCluster (`status.outputs.cluster_id`) |
| `spec.replicaName` | `string` | yes |  |  |
| `spec.region` | `enum` | yes |  |  |
| `spec.size` | `string` | yes |  |  |
| `spec.vpc` | `string \| valueFrom` |  |  | DigitalOceanVpc (`status.outputs.vpc_id`) |
| `spec.storageSizeMib` | `uint64` |  |  |  |
| `spec.tags` | `[]string` |  |  |  |

## Field Details

### spec.cluster

`string | valueFrom` · required

The database cluster to replicate (the primary). Use a literal cluster
UUID or a reference to a DigitalOceanDatabaseCluster resource.
Changing it replaces the replica.

Containment-exempt: a read replica is a single-node cluster of its own
that FOLLOWS the primary, in the primary's region or another; it is not
created inside it. On a diagram the replica stands as its own store
with a line to the primary, never inside the primary's room.

- references: DigitalOceanDatabaseCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDatabaseCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.replicaName

`string` · required

Name of the read replica. Unique within the cluster; the name IS the
replica's API identity for reads and deletes (DigitalOcean also mints
a UUID, exported as an output). Changing it replaces the replica.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.region

`enum` · required

The DigitalOcean region for the replica. Same region as the primary
gives a hot standby for read scaling; a different region gives a
cross-region read replica. Changing it replaces the replica.

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

### spec.size

`string` · required

The slug identifier for the replica's node size (e.g. "db-s-1vcpu-1gb").
Must be at least the primary's size (API-enforced). Growing it resizes
the replica in place; shrinking is not supported by DigitalOcean.

- rule: {"required":true}

### spec.vpc

`string | valueFrom`

(Optional) Reference to a DigitalOcean VPC for the replica's private
networking -- useful when the replica lives in a different region from
the primary and must join that region's VPC. When unset, DigitalOcean
places the replica in the target region's default VPC. Cannot be
changed after creation.

- references: DigitalOceanVpc (`status.outputs.vpc_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.storageSizeMib

`uint64`

(Optional) Custom disk size in MiB for the replica. When unset, the
size slug's default storage applies. Grows in place together with
size changes; DigitalOcean rounds and enforces per-slug bounds
server-side.

WARNING -- replica storage must be at least the primary's storage, or
replication can fail when the primary outgrows the replica's disk.

### spec.tags

`[]string`

(Optional) Tags applied to the replica in DigitalOcean, in addition to
the standard Planton labels both provisioners always apply.

WARNING -- replica tags are CREATE-ONLY upstream: changing this list
REPLACES the whole replica (a new replica is seeded from the primary;
replication catch-up applies). Settle tagging before production use.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanDatabaseReplica, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.replica_id` | `string` | UUID of the replica itself. |
| `status.outputs.cluster_id` | `string` | UUID of the primary database cluster this replica follows. |
| `status.outputs.replica_name` | `string` | Name of the replica (its API identity within the cluster). |
| `status.outputs.host` | `string` | Public hostname of the replica endpoint. |
| `status.outputs.private_host` | `string` | Private-network hostname of the replica endpoint, reachable from resources in the same VPC. |
| `status.outputs.port` | `uint32` | Port the replica listens on. |
| `status.outputs.database` | `string` | Name of the default database served by the replica. |
| `status.outputs.user` | `string` | Username of the replica's default user. |
| `status.outputs.password` | `string` | Password of the replica's default user. Secret. |
| `status.outputs.uri` | `string` | Full public connection URI for the replica, including credentials. Secret. |
| `status.outputs.private_uri` | `string` | Full private-network connection URI for the replica, including credentials. Secret. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cluster` | DigitalOceanDatabaseCluster | `status.outputs.cluster_id` |
| `spec.vpc` | DigitalOceanVpc | `status.outputs.vpc_id` |

## See Also

- [Overview](../README.md)
