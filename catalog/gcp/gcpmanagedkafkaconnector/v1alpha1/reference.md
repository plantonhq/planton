# GcpManagedKafkaConnector

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpManagedKafkaConnectorSpec defines one connector on a Managed Service
for Apache Kafka Connect cluster (`google_managed_kafka_connector`) -- a
single pipeline moving data between Kafka and another system: a
MirrorMaker 2 source replicating another Kafka cluster, a Pub/Sub source
or sink, a BigQuery sink, a Cloud Storage sink. Connectors are their own
block so the team whose data a pipeline moves owns it in its own
manifest.

Immutable: project_id, location, connect_cluster, connector_id (a change
replaces the connector). configs and the restart policy update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaConnector
metadata:
  name: orders-to-bigquery
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  connectCluster:
    value: events-connect
  configs:
    connector.class: com.wepay.kafka.connect.bigquery.BigQuerySinkConnector
    tasks.max: "3"
    topics: orders.v1
    project: my-gcp-project
    defaultDataset: orders
    value.converter: org.apache.kafka.connect.json.JsonConverter
    value.converter.schemas.enable: "false"
  taskRestartPolicy:
    minimumBackoff: 60s
    maximumBackoff: 1800s
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.connectCluster` | `string \| valueFrom` | yes |  | GcpManagedKafkaConnectCluster (`status.outputs.name`) |
| `spec.connectorId` | `string` |  |  |  |
| `spec.configs` | `map<string, string>` |  |  |  |
| `spec.taskRestartPolicy` | `GcpManagedKafkaConnectorTaskRestartPolicy` |  |  |  |
| `spec.taskRestartPolicy.minimumBackoff` | `string` |  |  |  |
| `spec.taskRestartPolicy.maximumBackoff` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the Connect cluster lives in: a literal project ID or
a GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Connect cluster's region, e.g. "us-central1". Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.connectCluster

`string | valueFrom` · required

The Connect cluster that runs the connector. A
GcpManagedKafkaConnectCluster reference resolves to its full resource
path (name output); a literal takes the full path or the bare Connect
cluster ID. The modules derive the bare ID Google's resource expects.
Immutable.

- references: GcpManagedKafkaConnectCluster (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpManagedKafkaConnectCluster, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.connectorId

`string`

The connector's ID, the name Kafka Connect shows for it. Defaults to
metadata.name. Immutable.

### spec.configs

`map<string, string>`

The connector's Kafka Connect configuration, keyed by property:
"connector.class" picks the plugin (for example
com.google.pubsub.kafka.sink.CloudPubSubSinkConnector,
com.wepay.kafka.connect.bigquery.BigQuerySinkConnector,
io.aiven.kafka.connect.gcs.GcsSinkConnector,
org.apache.kafka.connect.mirror.MirrorSourceConnector), "tasks.max"
its parallelism, "topics" what a sink reads, plus the plugin's own
keys and converters. The values are stored in plain text on the
Connect cluster; keep credentials out of them where the plugin offers
Google IAM authentication instead.

### spec.taskRestartPolicy

`GcpManagedKafkaConnectorTaskRestartPolicy`

Automatic restarts for failed tasks. Omit to leave failed tasks
stopped.

### spec.taskRestartPolicy.minimumBackoff

`string`

The shortest wait before retrying a failed task -- a duration in
seconds with up to nine fractional digits and an "s" suffix, e.g.
"60s". Empty leaves Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]{1,9})?s$"}}

### spec.taskRestartPolicy.maximumBackoff

`string`

The longest wait between retries -- the backoff's upper bound, same
format, e.g. "1800s". Empty leaves Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]{1,9})?s$"}}

### spec.deletionPolicy

`string`

What happens to the connector when this resource is destroyed:
  "" / "DELETE" -- deleted (the pipeline stops)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- it leaves management and keeps running

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpManagedKafkaConnector, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/connectClusters/{connect_cluster}/connectors/{connector_id}. |
| `status.outputs.connector_id` | `string` | The connector's ID (the last segment of name). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.connectCluster` | GcpManagedKafkaConnectCluster | `status.outputs.name` |

## See Also

- [Overview](../README.md)
