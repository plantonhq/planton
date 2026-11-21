# Team Topic Prefix

## Use Case

A team namespace: one principal may do everything on every topic whose name starts with the team's prefix, so new topics under the prefix need no new grant.

## When to Use

- Teams that create their own topics under a naming convention
- Platform services that manage a family of topics

## What This Creates

- An ACL on `topicPrefixed/payments.` granting ALL to the payments platform service account

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `projectId`, `cluster`, principals | `my-gcp-project`, `events`, the service accounts | Stand-ins: name the cluster's project, reference the GcpManagedKafkaCluster (`valueFrom`), and name your own service accounts. |
| `aclId` | `topicPrefixed/payments.` | The prefix your topic naming convention uses. |
| `aclEntries[].operation` | `ALL` | Narrow to READ/WRITE/DESCRIBE when the principal should not create or delete topics. |
