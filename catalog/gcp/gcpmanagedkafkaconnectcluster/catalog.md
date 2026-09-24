# GCP Managed Kafka Connect Cluster

Runs Kafka Connect without managing workers: a Google-operated Connect cluster attached to your Managed Kafka cluster, sized by vCPUs and memory, reaching your VPC through a private interface, ready for the connectors that stream data between Kafka and Pub/Sub, BigQuery, Cloud Storage, or another Kafka cluster.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `managedkafka.googleapis.com` on the project (never disabled on destroy)
- **Connect cluster** -- a `managed_kafka_connect_cluster` with its capacity and worker networks, carrying the platform attribution labels

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Managed Service for Apache Kafka admin permissions (`roles/managedkafka.admin`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpManagedKafkaCluster`** -- the Kafka cluster the workers attach to (`kafkaCluster`, its full path).
- **`GcpSubnetwork`** -- a private (RFC 1918) subnet in the Connect cluster's region for the workers' interface (`networkConfigs[].primarySubnet`).

## Deploy

### Console

Open the deployment store, find **GCP Managed Kafka Connect Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Connect Workers** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaConnectCluster
metadata:
  name: events-connect
  org: acme-corp
  env: prod
spec:
  location: us-central1
  kafkaCluster:
    value: projects/acme-prod/locations/us-central1/clusters/events
  capacityConfig:
    vcpuCount: 3
    memoryBytes: 12884901888
  networkConfigs:
    - primarySubnet:
        value: projects/acme-prod/regions/us-central1/subnetworks/kafka
```

```shell
planton apply -f managed-kafka-connect-cluster.yaml
```

This creates a 3-vCPU Connect cluster for the `events` Kafka cluster with its workers on the `kafka` subnet. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference the `GcpManagedKafkaCluster` from `kafkaCluster` and a `GcpSubnetwork` from `networkConfigs[].primarySubnet`; connectors reference this cluster's `name` output.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Capacity** -- `vcpuCount` and `memoryBytes` size the workers every connector shares and bill around the clock; both scale in place.

**Networks** -- each `primarySubnet` gets a Private Service Connect interface for the workers -- the path to the Kafka cluster and to the systems connectors reach.

**DNS** -- `dnsDomainNames` lets workers resolve names in other clusters' private zones -- required for MirrorMaker 2 from another Managed Kafka cluster.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpManagedKafkaCluster** | `kafkaCluster` | `status.outputs.name` |
| **GcpSubnetwork** | `networkConfigs[].primarySubnet` | `status.outputs.subnetwork_self_link` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The Connect cluster's full resource name | A connector's `connectCluster` |
| `connect_cluster_id` | The id | Console links |
| `location` | The region | Co-location |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Connect workers** -- 3 vCPU and 12 GiB of workers on the Kafka cluster's subnet. Start from the **Connect Workers** preset.

**Cluster migration** -- Workers that can resolve a second Kafka cluster's DNS domain for MirrorMaker 2. Start from the **MirrorMaker 2 Migration** preset.

## Works With

- [**GCP Managed Kafka Cluster**](/cloud-catalog/gcp-managed-kafka-cluster) -- the Kafka cluster served
- [**GCP Managed Kafka Connector**](/cloud-catalog/gcp-managed-kafka-connector) -- the pipelines that run here
- [**GCP Subnetwork**](/cloud-catalog/gcp-subnetwork) -- the workers' interface
