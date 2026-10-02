# GcpManagedKafkaTopic

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpManagedKafkaTopicSpec defines one topic on a Managed Service for
Apache Kafka cluster (`google_managed_kafka_topic`). Topics are their own
block so the team that produces to a topic owns it in its own manifest,
without touching the cluster's.

Immutable: project_id, location, cluster, topic_id, replication_factor
(a change replaces the topic and DELETES its messages). partition_count
grows in place but never shrinks; configs update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaTopic
metadata:
  name: orders
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  cluster:
    value: projects/my-gcp-project/locations/us-central1/clusters/events
  topicId: orders.v1
  partitionCount: 12
  replicationFactor: 3
  configs:
    retention.ms: "604800000"
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.cluster` | `string \| valueFrom` | yes |  | GcpManagedKafkaCluster (`status.outputs.name`) |
| `spec.topicId` | `string` |  |  |  |
| `spec.partitionCount` | `int32` |  |  |  |
| `spec.replicationFactor` | `int32` |  |  |  |
| `spec.configs` | `map<string, string>` |  |  |  |
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

The cluster the topic lives on. A GcpManagedKafkaCluster reference
resolves to its full resource path (name output); a literal takes the
full path or the bare cluster ID. The modules derive the bare ID
Google's resource expects. Immutable.

- references: GcpManagedKafkaCluster (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpManagedKafkaCluster, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.topicId

`string`

The topic's name as producers and consumers see it -- Kafka's own
rule: letters, digits, dots, underscores, hyphens, at most 249
characters (avoid mixing dots and underscores; Kafka treats them as
colliding in metric names). Defaults to metadata.name. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-zA-Z0-9._-]{1,249}$"}}

### spec.partitionCount

`int32`

How many partitions the topic has -- the unit of consumer parallelism.
Can be raised in place, never lowered; raising it changes which
partition a keyed message lands on, so per-key ordering is only
guaranteed for messages written after the change. Empty leaves
Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"gte":1}}

### spec.replicationFactor

`int32`

How many copies of each partition the cluster keeps. Google
recommends 3 for high availability (the brokers span three zones).
Immutable.

- rule: {"int32":{"gte":1}}

### spec.configs

`map<string, string>`

Topic-level overrides of the cluster's defaults, keyed by Kafka topic
property, e.g. "cleanup.policy" = "compact", "retention.ms" =
"604800000", "compression.type" = "producer". Keys absent here follow
the cluster default.

### spec.deletionPolicy

`string`

What happens to the topic when this resource is destroyed:
  "" / "DELETE" -- deleted, with its messages
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the topic leaves management and stays on the cluster

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpManagedKafkaTopic, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/clusters/{cluster}/topics/{topic_id}. |
| `status.outputs.topic_id` | `string` | The topic's name as Kafka clients use it (the last segment of name). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.cluster` | GcpManagedKafkaCluster | `status.outputs.name` |

## See Also

- [Overview](../README.md)
