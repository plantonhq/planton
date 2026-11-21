# Consumer Group Read

## Use Case

The second ACL every consumer needs: READ on the consumer group it joins, so it can commit offsets.

## When to Use

- Any consumer that reads with a group id
- Pairing with a topic READ grant

## What This Creates

- An ACL on `consumerGroup/billing` letting the billing worker join the group

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `projectId`, `cluster`, principals | `my-gcp-project`, `events`, the service accounts | Stand-ins: name the cluster's project, reference the GcpManagedKafkaCluster (`valueFrom`), and name your own service accounts. |
| `aclId` | `consumerGroup/billing` | Use the group id your consumer configures. |
