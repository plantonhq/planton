# OciStreamPool Examples

## Minimal Pool with One Stream

A stream pool with a single stream for development and testing:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciStreamPool
metadata:
  name: dev-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciStreamPool.dev-pool
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  streams:
    - name: "events"
      partitions: 1
```

## Multi-Stream Event Hub with Kafka Settings

A pool with Kafka auto-topic creation, 48-hour default retention, and multiple streams for an event-driven architecture:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciStreamPool
metadata:
  name: event-hub
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciStreamPool.event-hub
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

## Private Pool with KMS Encryption

A pool accessible only from a private subnet, encrypted with a customer-managed KMS key. Uses `valueFrom` for all infrastructure references:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciStreamPool
metadata:
  name: secure-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciStreamPool.secure-pool
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

## Pool Without Streams (Kafka Auto-Create)

A pool that relies on Kafka auto-topic creation — streams are created on-demand when producers publish:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciStreamPool
metadata:
  name: dynamic-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciStreamPool.dynamic-pool
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  kafkaSettings:
    autoCreateTopicsEnable: true
    logRetentionHours: 72
    numPartitions: 3
```

## Common Operations

### Connect with a Kafka client

After deploying the pool, use the `kafkaBootstrapServers` output as the bootstrap server in your Kafka client configuration. OCI Streaming uses SASL/PLAIN authentication with your OCI credentials.

### Add a new stream

Append a new entry to the `streams` list and re-apply. Existing streams are not affected.

### Increase Kafka log retention

Update `kafkaSettings.logRetentionHours` and re-apply. This is updatable without pool recreation. Note: this sets the default for auto-created topics; existing stream retention is controlled per-stream.

### Switch to customer-managed encryption

Set `kmsKeyId` to the OCID of a KMS key and re-apply. The pool is re-encrypted with the new key. This is updatable without recreation.

## Best Practices

1. **Declare streams explicitly** — avoid relying solely on auto-topic creation in production; explicit streams provide version control and visibility.
2. **Set 7-day retention for audit streams** — `retentionInHours: 168` provides a full week of replay capability.
3. **Use private endpoints for sensitive data** — restricts streaming traffic to within the VCN.
4. **Size partitions for throughput** — each partition provides a unit of parallelism; more partitions = higher throughput.
5. **Use `valueFrom` references** for `compartmentId` and `kmsKeyId` — avoids hardcoding OCIDs and maintains dependency ordering.
