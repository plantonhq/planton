# GCP Managed Kafka Connect Cluster

A Managed Service for Apache Kafka Connect cluster -- Google-operated Kafka Connect workers attached to one Kafka cluster, running the connectors that move data between Kafka and other systems (Pub/Sub, BigQuery, Cloud Storage, another Kafka cluster through MirrorMaker 2). The connectors are `GcpManagedKafkaConnector` blocks, owned by the teams whose data they move.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `managedkafka.googleapis.com` on the project (never disabled on destroy)
- **Connect cluster** -- a `managed_kafka_connect_cluster` with its capacity and worker networks, carrying the platform attribution labels

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Managed Service for Apache Kafka admin permissions (`roles/managedkafka.admin`) on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpManagedKafkaCluster`** -- the Kafka cluster the workers attach to (`kafkaCluster`, its full path).
- **`GcpSubnetwork`** -- a private (RFC 1918) subnet in the Connect cluster's region for the workers' interface (`networkConfigs[].primarySubnet`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaConnectCluster
metadata:
  name: events-connect
spec:
  location: us-central1
  kafkaCluster:
    valueFrom:
      kind: GcpManagedKafkaCluster
      name: events
      fieldPath: status.outputs.name
  capacityConfig:
    vcpuCount: 3
    memoryBytes: 3221225472
  networkConfigs:
    - primarySubnet:
        valueFrom:
          kind: GcpSubnetwork
          name: kafka-subnet
          fieldPath: status.outputs.subnetwork_self_link
```

```shell
planton apply -f managed-kafka-connect-cluster.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The region the workers run in. Immutable. |
| `kafkaCluster` | `StringValueOrRef` | The Kafka cluster -- a `GcpManagedKafkaCluster` reference or a full path. |
| `capacityConfig.vcpuCount` | `int64` | At least 3. |
| `capacityConfig.memoryBytes` | `int64` | At least 3 GiB, at a 1:1 to 1:8 vCPU:GiB ratio. |
| `networkConfigs[].primarySubnet` | `StringValueOrRef` | 1-10 subnets for the workers' interface. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. |
| `connectClusterId` | `string` | `metadata.name` | Immutable. |
| `networkConfigs[].dnsDomainNames` | `[]string` | none | Extra DNS domains the workers resolve -- a MirrorMaker 2 source cluster's domain. |
| `labels` | `map<string,string>` | `{}` | Labels; attribution labels win. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `vcpuCount` at least 3; `memoryBytes` at least 3 GiB and 1-8 GiB per vCPU.
- A literal `kafkaCluster` is the full path `projects/{p}/locations/{l}/clusters/{c}`.
- 1-10 network configs.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/connectClusters/{id}` -- what connectors reference |
| `connect_cluster_id` | `string` | The id |
| `location` | `string` | The region |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Connectors run as the Managed Kafka service agent.** Grant it what each connector reaches -- Pub/Sub publisher for a Pub/Sub sink, BigQuery data editor for a BigQuery sink.
- **MirrorMaker 2 needs DNS.** To replicate from another Managed Kafka cluster, add its DNS domain to `dnsDomainNames` (the bootstrap address without its `bootstrap.` label and port).
- **Destroy removes the connectors.** Deleting the Connect cluster deletes every connector running on it.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpManagedKafkaCluster** -- the Kafka cluster the workers attach to
- **GcpManagedKafkaConnector** -- the pipelines that run here
- **GcpSubnetwork** -- the workers' subnet

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
