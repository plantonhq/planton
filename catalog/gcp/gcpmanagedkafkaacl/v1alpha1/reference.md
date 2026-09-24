# GcpManagedKafkaAcl

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpManagedKafkaAclSpec defines the access rules for ONE resource pattern
on a Managed Service for Apache Kafka cluster (`google_managed_kafka_acl`)
-- every principal allowed or denied on, say, one topic or one consumer
group. ACLs are their own block so the team that owns a topic or a
consumer group grants access to it in its own manifest, without editing
the cluster.

Kafka ACLs are enforced for clients that authenticate to the cluster;
Google IAM still decides who may connect at all (roles/managedkafka.client).

Immutable: project_id, location, cluster, acl_id (a change replaces the
ACL). The entries update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaAcl
metadata:
  name: orders-consumer-group
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  cluster:
    value: events
  # The consumer group the orders service reads with.
  aclId: consumerGroup/orders-service
  aclEntries:
    - principal: User:orders-service@my-gcp-project.iam.gserviceaccount.com
      operation: READ
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.cluster` | `string \| valueFrom` | yes |  | GcpManagedKafkaCluster (`status.outputs.name`) |
| `spec.aclId` | `string` | yes |  |  |
| `spec.aclEntries` | `[]GcpManagedKafkaAclEntry` | yes |  |  |
| `spec.aclEntries[].principal` | `string` | yes |  |  |
| `spec.aclEntries[].operation` | `string` | yes |  |  |
| `spec.aclEntries[].permissionType` | `string` |  |  |  |
| `spec.aclEntries[].host` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the cluster lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The cluster's region, e.g. "us-central1". Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.cluster

`string | valueFrom` · required

The cluster the ACL applies to. A GcpManagedKafkaCluster reference
resolves to its full resource path (name output); a literal takes the
full path or the bare cluster ID. The modules derive the bare ID
Google's resource expects. Immutable.

- references: GcpManagedKafkaCluster (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpManagedKafkaCluster, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.aclId

`string` · required

The resource pattern, which is also the ACL's ID:
  cluster                          -- the cluster itself
  topic/{name}                     -- one topic
  consumerGroup/{name}             -- one consumer group
  transactionalId/{name}           -- one transactional ID
  topicPrefixed/{prefix}           -- every topic starting with prefix
  consumerGroupPrefixed/{prefix}   -- every consumer group with prefix
  transactionalIdPrefixed/{prefix} -- every transactional ID with prefix
{name} may be the literal wildcard "*" (every resource of the type).
One ACL exists per pattern on a cluster, so two manifests must never
declare the same one. Immutable.

- rule: {"required":true,"string":{"pattern":"^(cluster|(topic|consumerGroup|transactionalId)(Prefixed)?/.+)$"}}

### spec.aclEntries

`[]GcpManagedKafkaAclEntry` · required

The entries for this pattern -- at least one, at most 100.

- rule: {"repeated":{"minItems":"1","maxItems":"100"}}

### spec.aclEntries[].principal

`string` · required

Who the entry applies to, in Kafka's StandardAuthorizer form: "User:"
followed by a Google account, e.g.
"User:orders-api@my-project.iam.gserviceaccount.com", or "User:*" for
everyone. With mTLS, the principal is the certificate's mapped name
(see the cluster's ssl_principal_mapping_rules).

- rule: {"required":true,"string":{"pattern":"^User:.+$"}}

### spec.aclEntries[].operation

`string` · required

The Kafka operation:
  ALL, READ, WRITE, CREATE, DELETE, ALTER, DESCRIBE, CLUSTER_ACTION,
  DESCRIBE_CONFIGS, ALTER_CONFIGS, IDEMPOTENT_WRITE
Which operations are meaningful depends on the resource type (Kafka's
operations table): a producer needs WRITE (and DESCRIBE) on its topic;
a consumer needs READ on the topic and READ on its consumer group.
Google accepts any letter case; the spec takes the canonical upper
case.

- rule: {"required":true,"string":{"in":["ALL","READ","WRITE","CREATE","DELETE","ALTER","DESCRIBE","CLUSTER_ACTION","DESCRIBE_CONFIGS","ALTER_CONFIGS","IDEMPOTENT_WRITE"]}}

### spec.aclEntries[].permissionType

`string`

ALLOW (Google's default when empty) or DENY. A DENY wins over any
ALLOW for the same principal and operation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ALLOW","DENY"]}}

### spec.aclEntries[].host

`string`

The client host the entry applies to. Managed Service for Apache Kafka
accepts only "*" (every host), which is also the default when empty.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["*"]}}

### spec.deletionPolicy

`string`

What happens to the ACL when this resource is destroyed:
  "" / "DELETE" -- deleted (the access it granted ends)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the ACL leaves management and stays on the cluster

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpManagedKafkaAcl, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/clusters/{cluster}/acls/{acl_id}. |
| `status.outputs.resource_type` | `string` | The resource type Google derived from acl_id: CLUSTER, TOPIC, GROUP, or TRANSACTIONAL_ID. |
| `status.outputs.resource_name` | `string` | The resource name Google derived from acl_id ("kafka-cluster" for the cluster pattern; may be the wildcard "*"). |
| `status.outputs.pattern_type` | `string` | LITERAL or PREFIXED. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.cluster` | GcpManagedKafkaCluster | `status.outputs.name` |

## See Also

- [Overview](../README.md)
