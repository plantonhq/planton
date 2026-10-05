# DigitalOceanDatabaseConnectionPool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

DigitalOceanDatabaseConnectionPoolSpec models the full
digitalocean_database_connection_pool resource surface: a PgBouncer
connection pool on a DigitalOcean managed PostgreSQL cluster.

EVERY field here is create-only: the DigitalOcean Terraform provider
registers no update path for pools, so any change replaces the pool
(dropping its live connections). DigitalOcean's own API can update pools
in place; mirroring the provider is the recorded parity decision, to be
revisited if the provider ever grows an update path.

## Example

```yaml
# Reference manifests for DigitalOceanDatabaseConnectionPool --
# protovalidate-valid, embedded as the reference page's Example block, and
# the documents the offline tofu plans render. Two documents: a
# dedicated-user pool and an inbound-user pool.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseConnectionPool
metadata:
  name: orders-pool
spec:
  # Literal cluster UUID; use valueFrom to reference a
  # DigitalOceanDatabaseCluster resource instead.
  cluster:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  poolName: orders-pool
  mode: transaction
  size: 20
  dbName: orders
  user: orders-service
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseConnectionPool
metadata:
  name: shared-inbound-pool
spec:
  cluster:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  poolName: shared-inbound-pool
  mode: session
  size: 10
  dbName: defaultdb
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.cluster` | `string \| valueFrom` | yes |  | DigitalOceanDatabaseCluster (`status.outputs.cluster_id`) |
| `spec.poolName` | `string` | yes |  |  |
| `spec.mode` | `string` | yes |  |  |
| `spec.size` | `int32` | yes |  |  |
| `spec.dbName` | `string` | yes |  |  |
| `spec.user` | `string` |  |  |  |

## Field Details

### spec.cluster

`string | valueFrom` · required

The PostgreSQL cluster to create the pool on. Use a literal cluster
UUID or a reference to a DigitalOceanDatabaseCluster resource.
Connection pools exist only on PostgreSQL clusters (API-enforced).

- references: DigitalOceanDatabaseCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDatabaseCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.poolName

`string` · required

Name of the connection pool. Unique within the cluster; the name IS
the pool's API identity, and clients connect to it as if it were a
database name.

- rule: {"required":true,"string":{"minLen":"3","maxLen":"63"}}

### spec.mode

`string` · required

PgBouncer pooling mode. "transaction" is what most applications want
(a server connection per transaction); "session" holds a server
connection for the client's whole session (required for session-state
features like LISTEN/NOTIFY or prepared statements); "statement" is
the most aggressive and forbids multi-statement transactions.

- rule: {"required":true,"string":{"in":["session","transaction","statement"]}}

### spec.size

`int32` · required

Size of the pool: how many backend server connections it may hold
open. Bounded by the cluster's connection limit (size-slug dependent,
API-enforced) minus DigitalOcean's reserved connections.

- rule: {"required":true,"int32":{"gte":1}}

### spec.dbName

`string` · required

Name of the logical database the pool connects to. A plain name --
compose it with a DigitalOceanDatabaseDb resource by using the same
name, or use the cluster's default database ("defaultdb").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.user

`string`

(Optional) Name of the cluster user the pool authenticates as. When
omitted, DigitalOcean creates an "inbound user" pool: clients
authenticate with their own database credentials and the pool proxies
them, which is the safer default for shared pools. Reads echo the
empty value back, so omission is stable.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanDatabaseConnectionPool, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_id` | `string` | UUID of the PostgreSQL cluster the pool runs on. |
| `status.outputs.pool_name` | `string` | Name of the connection pool (its API identity within the cluster, and the database name clients connect to). |
| `status.outputs.host` | `string` | Public hostname of the pool endpoint. |
| `status.outputs.private_host` | `string` | Private-network hostname of the pool endpoint, reachable from resources in the same VPC. |
| `status.outputs.port` | `uint32` | Port the pool listens on (distinct from the cluster's own port). |
| `status.outputs.uri` | `string` (sensitive) | Full public connection URI for the pool, including credentials. Secret. |
| `status.outputs.private_uri` | `string` (sensitive) | Full private-network connection URI for the pool, including credentials. Secret. |
| `status.outputs.password` | `string` (sensitive) | Password of the pool's user. Secret. Empty for inbound-user pools (no dedicated user; clients bring their own credentials). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cluster` | DigitalOceanDatabaseCluster | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
