# DigitalOceanDatabaseUser

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanDatabaseUserSpec models the full digitalocean_database_user
resource surface: an additional user on a DigitalOcean managed database
cluster, with the MySQL authentication plugin choice and the Kafka /
OpenSearch access-control lists.

DigitalOcean generates the user's password and role server-side; both are
exported as stack outputs, never configured here. The API serializes user
creation and deletion per cluster, so composing many users on one cluster
is safe but inherently sequential.

## Example

```yaml
# Reference manifests for DigitalOceanDatabaseUser -- protovalidate-valid,
# embedded as the reference page's Example block, and the documents the
# offline tofu plans render. Two documents: a plain PostgreSQL/MySQL-style
# user, and a Kafka user exercising the full ACL surface.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseUser
metadata:
  name: orders-service-user
spec:
  # Literal cluster UUID; use valueFrom to reference a
  # DigitalOceanDatabaseCluster resource instead.
  cluster:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  userName: orders-service
  # PostgreSQL users declare an empty settings block (DigitalOcean stores a
  # settings object for every PostgreSQL user); MySQL users leave it out.
  settings: {}
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseUser
metadata:
  name: events-pipeline-user
spec:
  cluster:
    value: bbbbbbbb-cccc-dddd-eeee-ffffffffffff
  userName: events-pipeline
  settings:
    kafkaAcls:
      - topic: events-*
        permission: produceconsume
      - topic: audit-log
        permission: produce
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.cluster` | `string \| valueFrom` | yes |  | DigitalOceanDatabaseCluster (`status.outputs.cluster_id`) |
| `spec.userName` | `string` | yes |  |  |
| `spec.mysqlAuthPlugin` | `string` |  |  |  |
| `spec.settings` | `DigitalOceanDatabaseUserSettings` |  |  |  |
| `spec.settings.kafkaAcls` | `[]DigitalOceanDatabaseUserKafkaAcl` |  |  |  |
| `spec.settings.kafkaAcls[].topic` | `string` | yes |  |  |
| `spec.settings.kafkaAcls[].permission` | `string` | yes |  |  |
| `spec.settings.opensearchAcls` | `[]DigitalOceanDatabaseUserOpenSearchAcl` |  |  |  |
| `spec.settings.opensearchAcls[].index` | `string` | yes |  |  |
| `spec.settings.opensearchAcls[].permission` | `string` | yes |  |  |

## Field Details

### spec.cluster

`string | valueFrom` · required

The database cluster this user belongs to. Use a literal cluster UUID or
a reference to a DigitalOceanDatabaseCluster resource. Changing it
replaces the user (the new user gets a NEW server-generated password).

- references: DigitalOceanDatabaseCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDatabaseCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.userName

`string` · required

Name of the database user. Unique within the cluster; the name IS the
user's API identity. Changing it replaces the user.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.mysqlAuthPlugin

`string`

(Optional) MySQL authentication plugin for this user. Applies only to
MySQL clusters (DigitalOcean rejects it on other engines at request
time). When unset, DigitalOcean uses caching_sha2_password; clearing a
previously set value resets the user back to that default. Updates apply
in place through a password-preserving auth reset.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["mysql_native_password","caching_sha2_password"]}}

### spec.settings

`DigitalOceanDatabaseUserSettings`

(Optional) Engine-specific access control for this user (Kafka topic
ACLs and OpenSearch index ACLs). Both provisioners record these only
from the create response and never refresh them from the API, so what
is configured here is the source of truth; the live ACL state is not
observable through Planton afterward. ACL changes apply in place.

SET IT BY ENGINE (measured 2026-09-17). PostgreSQL: declare
`settings: {}` even with no ACLs -- every PostgreSQL user comes back
with a settings object (`pg_allow_replication: false`) that the
provisioners store at create, and a manifest without the block plans
its removal on the first re-plan (a one-time server-side no-op, but a
change). MySQL: leave it unset -- a MySQL user never carries a settings
object and the API refuses a settings update on a MySQL cluster
(`422 operation is not supported for this cluster type`), so even
`settings: {}` would fail every apply after the first. Kafka and
OpenSearch: declare the ACLs. The provisioners send the block exactly
when this field is present; they cannot infer the engine from a cluster
UUID, so the manifest carries that knowledge.

### spec.settings.kafkaAcls

`[]DigitalOceanDatabaseUserKafkaAcl`

Kafka topic ACLs granting this user permissions on topics. Topic
accepts a literal name or a wildcard pattern (e.g. "events-*").

### spec.settings.kafkaAcls[].topic

`string` · required

The Kafka topic (or wildcard topic pattern) the permission applies to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.settings.kafkaAcls[].permission

`string` · required

Permission on the topic.

- rule: {"required":true,"string":{"in":["admin","consume","produce","produceconsume"]}}

### spec.settings.opensearchAcls

`[]DigitalOceanDatabaseUserOpenSearchAcl`

OpenSearch index ACLs granting this user permissions on indexes. Index
accepts a literal name or a wildcard pattern (e.g. "logs-*").

### spec.settings.opensearchAcls[].index

`string` · required

The OpenSearch index (or wildcard index pattern) the permission applies
to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.settings.opensearchAcls[].permission

`string` · required

Permission on the index.

- rule: {"required":true,"string":{"in":["deny","admin","read","write","readwrite"]}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanDatabaseUser, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_id` | `string` | UUID of the database cluster the user belongs to. |
| `status.outputs.user_name` | `string` | Name of the database user (its API identity within the cluster). |
| `status.outputs.role` | `string` | Role DigitalOcean assigned to the user (normally "normal"; the cluster's built-in default user is "primary"). |
| `status.outputs.password` | `string` | Server-generated password for the user. Secret. MongoDB clusters return it only at creation time. |
| `status.outputs.access_cert` | `string` | Kafka clusters only: PEM access certificate for mutual-TLS authentication. Secret. Empty on other engines. |
| `status.outputs.access_key` | `string` | Kafka clusters only: PEM access key paired with access_cert. Secret. Empty on other engines. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cluster` | DigitalOceanDatabaseCluster | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
