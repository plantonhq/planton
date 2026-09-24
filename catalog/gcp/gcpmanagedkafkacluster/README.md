# GCP Managed Kafka Cluster

A Managed Service for Apache Kafka cluster -- Google-operated Apache Kafka brokers in one region, reachable from the VPC subnets you attach. Clients authenticate with Google IAM (SASL OAUTHBEARER) or, with `tlsConfig`, mutual TLS. Topics, access rules, and Kafka Connect are their own blocks (`GcpManagedKafkaTopic`, `GcpManagedKafkaAcl`, `GcpManagedKafkaConnectCluster`) so the teams that own them declare them without editing the cluster.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `managedkafka.googleapis.com` on the project (never disabled on destroy)
- **Kafka cluster** -- a `managed_kafka_cluster` with its capacity, networks, optional CMEK, rebalancing, and mTLS configuration, carrying the platform attribution labels

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Managed Service for Apache Kafka admin permissions (`roles/managedkafka.admin`) on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpSubnetwork`** -- at least one subnet in the cluster's region (`networkConfigs[].subnet`), one per VPC network; Google creates the bootstrap and broker addresses and their DNS entries in it. The subnet's project may differ (a Shared VPC host).

### Optional Dependencies

- **`GcpKmsKey`** -- a key in the cluster's region for CMEK (`kmsKey`); the Managed Kafka service agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on it.
- **Certificate Authority Service CA pools** -- for mTLS (`tlsConfig.caPools`), named by their full resource names.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaCluster
metadata:
  name: events
spec:
  location: us-central1
  capacityConfig:
    vcpuCount: 3
    memoryBytes: 12884901888
  networkConfigs:
    - subnet:
        valueFrom:
          kind: GcpSubnetwork
          name: kafka-subnet
          fieldPath: status.outputs.subnetwork_self_link
```

```shell
planton apply -f managed-kafka-cluster.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The region the brokers run in. Immutable. |
| `capacityConfig.vcpuCount` | `int64` | vCPUs across the cluster, at least 3. |
| `capacityConfig.memoryBytes` | `int64` | Memory in bytes, 1-8 GiB per vCPU. |
| `networkConfigs[].subnet` | `StringValueOrRef` | 1-10 subnets, one per VPC network; a `GcpSubnetwork` reference or a `projects/.../subnetworks/...` literal. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `clusterId` | `string` | `metadata.name` | RFC 1035, 1-63 characters. Immutable. |
| `brokerDiskSizeGib` | `int64` | Google's default | Disk per broker, at least 100 GiB. |
| `kmsKey` | `StringValueOrRef` | Google-managed | CMEK key in the cluster's region (`GcpKmsKey` reference). Immutable. |
| `rebalanceMode` | `string` | `NO_REBALANCE` | `AUTO_REBALANCE_ON_SCALE_UP` moves partitions onto new brokers after a scale-up. |
| `tlsConfig` | `object` | unchanged | mTLS: `sslPrincipalMappingRules` and up to 10 `caPools`. An empty object clears an earlier configuration. |
| `labels` | `map<string,string>` | `{}` | Labels; the platform attribution labels win on a key conflict. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `capacityConfig.vcpuCount` is at least 3, and `memoryBytes` is between 1 GiB and 8 GiB per vCPU.
- `networkConfigs` holds 1-10 entries; `tlsConfig.caPools` at most 10, each a full CA pool name.
- `brokerDiskSizeGib`, when set, is at least 100; `rebalanceMode` takes only Google's two modes.
- `clusterId` follows RFC 1035.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/clusters/{cluster_id}` -- what topics, ACLs, and Connect clusters reference |
| `cluster_id` | `string` | The cluster's id |
| `location` | `string` | The cluster's region |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **The bootstrap address is read, not wired.** Google fixes a cluster's bootstrap address for its lifetime, but its format can differ between clusters, and the pinned Pulumi SDK cannot read it yet -- so it is not an output. Read it once with `gcloud managed-kafka clusters describe CLUSTER --location=LOCATION --format="value(bootstrapAddress)"` (port 9092 for Google IAM over SASL, 9192 for mTLS).
- **Internet access waits for the Pulumi SDK.** Google's public cluster access (`public_cluster_config`, new in provider 8.3) is not yet in the pinned pulumi-gcp SDK, so it is held out of the spec on both engines until pulumi-gcp v10 -- clients connect from an attached VPC.
- **Creation is slow.** Google quotes up to about 30 minutes for a new cluster.
- **The key and the id are forever.** `kmsKey`, `clusterId`, and `location` cannot change -- a change replaces the cluster and every message in it.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpManagedKafkaTopic** -- the topics on this cluster
- **GcpManagedKafkaAcl** -- access rules for topics, consumer groups, and the cluster
- **GcpManagedKafkaConnectCluster** -- Kafka Connect workers attached to this cluster
- **GcpSubnetwork** -- the subnets the cluster is reachable from
- **GcpKmsKey** -- the CMEK key

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
