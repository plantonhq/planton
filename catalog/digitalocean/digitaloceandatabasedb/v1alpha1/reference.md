# DigitalOceanDatabaseDb

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

DigitalOceanDatabaseDbSpec models the full digitalocean_database_db
resource surface: an additional logical database inside a DigitalOcean
managed database cluster.

The resource is deliberately minimal upstream: both fields are
create-only (any change replaces the logical database, which DROPS its
data -- treat renames as migrations, never edits), and DigitalOcean's
read is a bare existence check. Connection credentials live on the
cluster and its users, not here.

## Example

```yaml
# Reference manifest for DigitalOceanDatabaseDb -- protovalidate-valid,
# embedded as the reference page's Example block, and the document the
# offline tofu plan renders.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseDb
metadata:
  name: orders-database
spec:
  # Literal cluster UUID; use valueFrom to reference a
  # DigitalOceanDatabaseCluster resource instead.
  cluster:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  databaseName: orders
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.cluster` | `string \| valueFrom` | yes |  | DigitalOceanDatabaseCluster (`status.outputs.cluster_id`) |
| `spec.databaseName` | `string` | yes |  |  |

## Field Details

### spec.cluster

`string | valueFrom` · required

The database cluster to create the logical database in. Use a literal
cluster UUID or a reference to a DigitalOceanDatabaseCluster resource.
Changing it replaces the logical database.

- references: DigitalOceanDatabaseCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDatabaseCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.databaseName

`string` · required

Name of the logical database. Unique within the cluster; the name IS
the database's API identity. Changing it replaces the database and
drops the old one's data.

- rule: {"required":true,"string":{"minLen":"1"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanDatabaseDb, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_id` | `string` | UUID of the database cluster the logical database lives in. |
| `status.outputs.database_name` | `string` | Name of the logical database (its API identity within the cluster). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cluster` | DigitalOceanDatabaseCluster | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
