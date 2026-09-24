# GCP Managed Kafka Connector

One connector on a Managed Service for Apache Kafka Connect cluster -- a single pipeline moving data between Kafka and another system: a Pub/Sub source or sink, a BigQuery sink, a Cloud Storage sink, or a MirrorMaker 2 source replicating another Kafka cluster. Connectors are their own block so the team whose data a pipeline moves owns it in its own manifest.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Connector** -- a `managed_kafka_connector` on the referenced Connect cluster

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Managed Service for Apache Kafka admin permissions (`roles/managedkafka.admin`) on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpManagedKafkaConnectCluster`** -- the Connect cluster that runs the connector (`connectCluster`).

### Optional Dependencies

- **`GcpManagedKafkaTopic`** -- the topics a sink reads or a source writes, named in `configs`.
- **The destination or source** -- a Pub/Sub topic, BigQuery dataset, or bucket, with the Managed Kafka service agent granted access to it.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaConnector
metadata:
  name: orders-to-pubsub
spec:
  location: us-central1
  connectCluster:
    valueFrom:
      kind: GcpManagedKafkaConnectCluster
      name: events-connect
      fieldPath: status.outputs.name
  configs:
    connector.class: com.google.pubsub.kafka.sink.CloudPubSubSinkConnector
    tasks.max: "3"
    topics: orders
    cps.project: my-project
    cps.topic: orders-events
```

```shell
planton apply -f managed-kafka-connector.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Connect cluster's region. Immutable. |
| `connectCluster` | `StringValueOrRef` | The Connect cluster -- a reference or its full path or bare id. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. |
| `connectorId` | `string` | `metadata.name` | Immutable. |
| `configs` | `map<string,string>` | none | The Kafka Connect configuration: `connector.class`, `tasks.max`, `topics`, and the plugin's keys. |
| `taskRestartPolicy` | `object` | no restarts | `minimumBackoff` and `maximumBackoff` durations (`60s`). |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- Backoffs are durations in seconds with an `s` suffix, up to nine fractional digits.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/connectClusters/{cc}/connectors/{id}` |
| `connector_id` | `string` | The id |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Set a restart policy.** Without `taskRestartPolicy`, a failed task stays failed until someone restarts it.
- **Configs are plain text.** Connector configuration is stored as written; prefer plugins that authenticate with Google IAM over credentials in `configs`.
- **The service agent needs access.** Connectors act as the Managed Kafka service agent -- grant it publisher on a Pub/Sub topic, data editor on a BigQuery dataset, object creator on a bucket.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpManagedKafkaConnectCluster** -- the workers that run the connector
- **GcpManagedKafkaTopic** -- the topics it reads or writes

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
