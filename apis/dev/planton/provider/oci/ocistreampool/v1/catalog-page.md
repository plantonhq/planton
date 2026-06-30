# OCI Stream Pool

Deploys an Oracle Cloud Infrastructure Streaming stream pool with bundled streams. The stream pool provides a Kafka-compatible managed event-streaming endpoint with configurable Kafka settings, optional KMS encryption, and optional private networking. Streams are declared inline and inherit the pool's configuration.

## What Gets Created

When you deploy an OciStreamPool resource, Planton provisions:

- **Stream Pool** — a `streaming.StreamPool` resource in the specified compartment with configurable Kafka compatibility settings (auto-create topics, log retention, default partitions), optional KMS encryption, and optional private endpoint.
- **Streams** — one `streaming.Stream` per entry in the `streams` list. Each stream is created within the pool with a specified partition count and optional retention period. Streams depend on the pool for creation ordering.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the stream pool will be created — either a literal value or a reference to an OciCompartment resource
- **A subnet OCID** (for private pools only) — the subnet where the private endpoint will be placed
- **A KMS key OCID** (optional) — if using customer-managed encryption

## Quick Start

Create a file `stream-pool.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciStreamPool
metadata:
  name: my-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciStreamPool.my-pool
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  streams:
    - name: "events"
      partitions: 1
```

Deploy:

```shell
planton apply -f stream-pool.yaml
```

This creates a stream pool with Oracle-managed encryption and one stream with a single partition and 24-hour default retention. The pool OCID, endpoint FQDN, and Kafka bootstrap servers are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the stream pool will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `kafkaSettings` | `KafkaSettings` | — | Kafka compatibility layer settings. All sub-fields are optional and updatable. |
| `kmsKeyId` | `StringValueOrRef` | — | OCID of a KMS master encryption key. When unset, Oracle-managed encryption is used. Updatable. Can reference an OciKmsKey resource via `valueFrom`. |
| `privateEndpointSettings` | `PrivateEndpointSettings` | — | Private networking configuration. All sub-fields are ForceNew (changes force pool recreation). |
| `streams` | `Stream[]` | — | Streams within the pool. Each stream is a sub-resource. |

### KafkaSettings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `autoCreateTopicsEnable` | `bool` | — | Auto-create topics when a Kafka producer publishes to a non-existent topic. |
| `logRetentionHours` | `int32` | `24` | Default hours to retain log data (24-168). |
| `numPartitions` | `int32` | — | Default partition count for auto-created topics. |

### PrivateEndpointSettings

| Field | Type | Description |
|-------|------|-------------|
| `subnetId` | `StringValueOrRef` | OCID of the subnet for the private endpoint. Required. Can reference an OciSubnet resource via `valueFrom`. |
| `nsgIds` | `StringValueOrRef[]` | NSGs for the private endpoint. Can reference OciSecurityGroup resources via `valueFrom`. |
| `privateEndpointIp` | `string` | Specific IP within the subnet CIDR. When omitted, OCI auto-assigns. |

### Stream

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | — | Stream name. Used as the Kafka topic name. ForceNew. |
| `partitions` | `int32` | — | Number of partitions. ForceNew. >= 1. |
| `retentionInHours` | `int32` | `24` | Retention period in hours (24-168). ForceNew. |

## Examples

### Minimal Public Pool with One Stream

A stream pool with a single stream for development:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciStreamPool
metadata:
  name: dev-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciStreamPool.dev-pool
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  streams:
    - name: "events"
      partitions: 1
```

### Multi-Stream Pool with Kafka Settings

A pool with Kafka auto-topic creation enabled, 48-hour retention, and multiple streams:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciStreamPool
metadata:
  name: event-hub
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciStreamPool.event-hub
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  kafkaSettings:
    autoCreateTopicsEnable: true
    logRetentionHours: 48
    numPartitions: 3
  streams:
    - name: "orders"
      partitions: 5
      retentionInHours: 168
    - name: "notifications"
      partitions: 3
    - name: "audit-events"
      partitions: 3
      retentionInHours: 168
```

### Private Pool with KMS Encryption

A pool accessible only from a private subnet, with customer-managed encryption:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciStreamPool
metadata:
  name: secure-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciStreamPool.secure-pool
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  kmsKeyId:
    valueFrom:
      kind: OciKmsKey
      name: streaming-key
      fieldPath: status.outputs.keyId
  privateEndpointSettings:
    subnetId:
      valueFrom:
        kind: OciSubnet
        name: private-subnet
        fieldPath: status.outputs.subnetId
    nsgIds:
      - valueFrom:
          kind: OciSecurityGroup
          name: streaming-nsg
          fieldPath: status.outputs.networkSecurityGroupId
  streams:
    - name: "sensitive-data"
      partitions: 5
      retentionInHours: 168
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `stream_pool_id` | `string` | OCID of the stream pool |
| `endpoint_fqdn` | `string` | FQDN for accessing streams. For private pools, resolves only within the associated subnet. |
| `kafka_bootstrap_servers` | `string` | Kafka-compatible bootstrap server string for producing and consuming messages |

## Related Components

- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciKmsKey](/docs/catalog/oci/ocikmskey) — provides the encryption key referenced by `kmsKeyId` via `valueFrom`
- [OciSubnet](/docs/catalog/oci/ocisubnet) — provides the subnet for private endpoints via `valueFrom`
- [OciSecurityGroup](/docs/catalog/oci/ocisecuritygroup) — provides NSGs for private endpoints via `valueFrom`
