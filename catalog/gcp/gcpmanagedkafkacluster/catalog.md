# GCP Managed Kafka Cluster

Runs Apache Kafka for your event streams without running brokers: Google operates a highly available cluster across three zones in one region, sized by vCPUs and memory, reachable by name from every VPC you attach it to, with clients signing in through Google IAM or mutual TLS. Topics, access rules, and connectors are separate blocks, so application teams add their own without touching the cluster.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `managedkafka.googleapis.com` on the project (never disabled on destroy)
- **Kafka cluster** -- a `managed_kafka_cluster` with its capacity, networks, optional CMEK, rebalancing, and mTLS configuration, carrying the platform attribution labels

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Managed Service for Apache Kafka admin permissions (`roles/managedkafka.admin`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpSubnetwork`** -- at least one subnet in the cluster's region (`networkConfigs[].subnet`), one per VPC network; Google creates the bootstrap and broker addresses and their DNS entries in it. The subnet's project may differ (a Shared VPC host).

### Optional Dependencies

- **`GcpKmsKey`** -- a key in the cluster's region for CMEK (`kmsKey`); the Managed Kafka service agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on it.
- **Certificate Authority Service CA pools** -- for mTLS (`tlsConfig.caPools`), named by their full resource names.

## Deploy

### Console

Open the deployment store, find **GCP Managed Kafka Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Development Cluster** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaCluster
metadata:
  name: events
  org: acme-corp
  env: prod
spec:
  location: us-central1
  capacityConfig:
    vcpuCount: 3
    memoryBytes: 12884901888
  networkConfigs:
    - subnet:
        value: projects/acme-prod/regions/us-central1/subnetworks/kafka
```

```shell
planton apply -f managed-kafka-cluster.yaml
```

This creates a 3-vCPU, 12 GiB Kafka cluster in us-central1 reachable from the `kafka` subnet's network. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpSubnetwork` from `networkConfigs[].subnet` and, for CMEK, a `GcpKmsKey` from `kmsKey`. Topics, ACLs, and a Connect cluster reference this cluster's `name` output.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Capacity** -- `vcpuCount` and `memoryBytes` size the whole cluster and bill around the clock; both scale in place. Google spreads them across brokers in three zones.

**Networks** -- each `networkConfigs` entry makes the cluster reachable in one VPC network, one subnet per network, up to ten. Clients anywhere in those networks resolve the brokers by name.

**Authentication** -- Google IAM (`roles/managedkafka.client` plus Kafka ACLs) is always on; `tlsConfig.caPools` adds mutual TLS with certificates from Certificate Authority Service.

**Rebalancing** -- `AUTO_REBALANCE_ON_SCALE_UP` spreads partitions onto new brokers after a scale-up; the default leaves them where they are.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpSubnetwork** | `networkConfigs[].subnet` | `status.outputs.subnetwork_self_link` |
| **GcpKmsKey** | `kmsKey` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The cluster's full resource name | A topic's or ACL's `cluster`, a Connect cluster's `kafkaCluster` |
| `cluster_id` | The cluster's id | Client configuration, console links |
| `location` | The cluster's region | Co-locating Connect clusters |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Development cluster** -- The smallest cluster Google sells -- 3 vCPU, 3 GiB -- on one subnet. Start from the **Development Cluster** preset.

**Production** -- A larger cluster with your own encryption key, a bigger broker disk, rebalance on scale-up, and destroy protection. Start from the **Production with CMEK** preset.

**Certificate-based clients** -- mTLS against a Certificate Authority Service pool, with principal mapping rules for Kafka ACLs. Start from the **Mutual TLS Clients** preset.

## Works With

- [**GCP Managed Kafka Topic**](/cloud-catalog/gcp-managed-kafka-topic) -- the topics on this cluster
- [**GCP Managed Kafka ACL**](/cloud-catalog/gcp-managed-kafka-acl) -- who may read and write them
- [**GCP Managed Kafka Connect Cluster**](/cloud-catalog/gcp-managed-kafka-connect-cluster) -- connectors to and from other systems
- [**GCP Subnetwork**](/cloud-catalog/gcp-subnetwork) -- where the brokers are reachable
