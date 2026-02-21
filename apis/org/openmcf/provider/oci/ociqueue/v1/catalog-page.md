# OCI Queue

Deploys an Oracle Cloud Infrastructure Queue — a fully managed, serverless message queue for asynchronous communication between decoupled services. Supports configurable message retention, visibility timeouts, dead-letter queues, optional KMS encryption, large message support (up to 512 KB), and consumer groups for partitioned consumption.

## What Gets Created

When you deploy an OciQueue resource, OpenMCF provisions:

- **Queue** — a `queue.Queue` resource in the specified compartment with configurable retention, visibility timeout, polling timeout, optional KMS encryption, optional dead-letter queue, and optional capabilities (large messages, consumer groups).

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the queue will be created — either a literal value or a reference to an OciCompartment resource
- **A KMS key OCID** (optional) — if using customer-managed encryption

## Quick Start

Create a file `queue.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciQueue
metadata:
  name: my-queue
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciQueue.my-queue
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
```

Deploy:

```shell
openmcf apply -f queue.yaml
```

This creates a queue with Oracle-managed encryption, 7-day default retention, and standard message size limits. The queue OCID and messages endpoint are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the queue will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `customEncryptionKeyId` | `StringValueOrRef` | — | OCID of a KMS master encryption key for encrypting message content. When omitted, Oracle-managed encryption is used. Can reference an OciKmsKey resource via `valueFrom`. |
| `deadLetterQueueDeliveryCount` | `int32` | OCI default | Number of delivery attempts before a message is moved to the dead-letter queue. A value of 0 disables the DLQ. |
| `retentionInSeconds` | `int32` | `604800` (7 days) | Retention period for messages in seconds. ForceNew — changing this value forces queue recreation. |
| `timeoutInSeconds` | `int32` | OCI default | Default polling timeout for GetMessages calls, in seconds. |
| `visibilityInSeconds` | `int32` | OCI default | Default visibility timeout for consumed messages, in seconds. A consumed message is invisible to other consumers for this duration. |
| `channelConsumptionLimit` | `int32` | OCI default | Percentage of allocated queue resources that can be consumed by a single channel. |
| `isLargeMessagesEnabled` | `bool` | `false` | Enable large message support (up to 512 KB per message). Maps to the `LARGE_MESSAGES` capability. |
| `consumerGroupConfig` | `ConsumerGroupConfig` | — | Consumer group configuration. When provided, enables the `CONSUMER_GROUPS` capability for partitioned consumption. |

### ConsumerGroupConfig

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `isPrimaryEnabled` | `bool` | — | Whether to enable the primary consumer group after adding the capability. |
| `primaryDeadLetterQueueDeliveryCount` | `int32` | — | DLQ delivery count for the primary consumer group. Overrides the queue-level `deadLetterQueueDeliveryCount`. A value of 0 disables the DLQ for this group. |
| `primaryDisplayName` | `string` | `"Primary Consumer Group"` | Display name for the primary consumer group. |

## Examples

### Minimal Queue

A queue with default settings for development:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciQueue
metadata:
  name: dev-queue
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciQueue.dev-queue
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
```

### Queue with DLQ and Custom Timeouts

A production queue with dead-letter queue enabled, extended visibility timeout, and KMS encryption:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciQueue
metadata:
  name: order-processing
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciQueue.order-processing
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  customEncryptionKeyId:
    valueFrom:
      kind: OciKmsKey
      name: queue-key
      fieldPath: status.outputs.keyId
  deadLetterQueueDeliveryCount: 5
  retentionInSeconds: 1209600
  visibilityInSeconds: 120
  timeoutInSeconds: 20
```

### Large Messages with Consumer Groups

A queue supporting 512 KB messages with partitioned consumption via consumer groups:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciQueue
metadata:
  name: event-bus
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciQueue.event-bus
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  isLargeMessagesEnabled: true
  deadLetterQueueDeliveryCount: 3
  consumerGroupConfig:
    isPrimaryEnabled: true
    primaryDisplayName: "main-consumers"
    primaryDeadLetterQueueDeliveryCount: 5
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `queue_id` | `string` | OCID of the queue |
| `messages_endpoint` | `string` | Endpoint URL for consuming or publishing messages |

## Related Components

- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciKmsKey](/docs/catalog/oci/ocikmskey) — provides the encryption key referenced by `customEncryptionKeyId` via `valueFrom`
- [OciAlarm](/docs/catalog/oci/ocialarm) — monitors queue metrics (queue depth, message age) via MQL
- [OciLogGroup](/docs/catalog/oci/ociloggroup) — collects queue service logs
