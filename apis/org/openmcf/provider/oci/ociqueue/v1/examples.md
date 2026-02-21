# OciQueue Examples

## Minimal Queue

A queue with default settings for development and testing:

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

## Production Queue with DLQ and Encryption

A queue with dead-letter handling, KMS encryption, and custom timeouts:

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

## Large Messages with Consumer Groups

A queue supporting large payloads with partitioned consumption:

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

## Queue with Channel Throttling

A queue with per-channel consumption limits to prevent a single consumer from monopolizing resources:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciQueue
metadata:
  name: shared-queue
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciQueue.shared-queue
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  channelConsumptionLimit: 50
  visibilityInSeconds: 60
  timeoutInSeconds: 10
```

## Common Operations

### Adjust visibility timeout

Update `visibilityInSeconds` and re-apply. This changes how long a consumed message remains invisible to other consumers. Increase for long-running processing; decrease for faster retries.

### Enable large messages

Set `isLargeMessagesEnabled: true` and re-apply. This adds the `LARGE_MESSAGES` capability, which cannot be removed once enabled.

### Add consumer groups

Add a `consumerGroupConfig` block and re-apply. The `CONSUMER_GROUPS` capability is enabled automatically and cannot be removed once added.

### Monitor queue depth

Use OciAlarm with the MQL query `QueueDepth[5m].max() > 1000` in the `oci_queue` namespace to alert when messages accumulate.

## Best Practices

1. **Set `deadLetterQueueDeliveryCount` in production** — prevents poison messages from blocking the queue indefinitely. A value of 3-5 is typical.
2. **Size `visibilityInSeconds` to your processing time** — too short causes duplicate processing; too long delays retries after failures.
3. **Use KMS encryption for sensitive data** — `customEncryptionKeyId` encrypts message content at rest with a customer-managed key.
4. **Be aware that `retentionInSeconds` is ForceNew** — changing the retention period destroys and recreates the queue. Plan retention at creation time.
5. **Use `valueFrom` references** for `compartmentId` and `customEncryptionKeyId` — avoids hardcoding OCIDs and maintains dependency ordering.
