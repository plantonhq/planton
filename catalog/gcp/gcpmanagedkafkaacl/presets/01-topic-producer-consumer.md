# Topic Producer and Consumer

## Use Case

The grants one topic needs: its producer may write and describe it, and its consumer may read it.

## When to Use

- A service that owns a topic and one or more readers
- Least-privilege access per topic

## What This Creates

- An ACL on `topic/orders` with WRITE and DESCRIBE for the producer and READ for the consumer

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `projectId`, `cluster`, principals | `my-gcp-project`, `events`, the service accounts | Stand-ins: name the cluster's project, reference the GcpManagedKafkaCluster (`valueFrom`), and name your own service accounts. |
| `aclId` | `topic/orders` | The topic this ACL governs. |
| `aclEntries` | producer + consumer | One entry per principal and operation. |
