# GCP Managed Kafka ACL

Grants producers and consumers access to their topics and consumer groups in the manifest of the team that owns them: one ACL per resource pattern, with ALLOW and DENY entries for Google identities or certificate principals.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Kafka ACL** -- a `managed_kafka_acl` for the pattern named by `aclId`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Managed Service for Apache Kafka admin permissions (`roles/managedkafka.admin`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpManagedKafkaCluster`** -- the cluster the ACL applies to (`cluster`).

## Deploy

### Console

Open the deployment store, find **GCP Managed Kafka ACL**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Topic Producer and Consumer** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaAcl
metadata:
  name: orders-consumers
  org: acme-corp
  env: prod
spec:
  location: us-central1
  cluster:
    value: events
  aclId: consumerGroup/billing
  aclEntries:
    - principal: User:billing-worker@acme-prod.iam.gserviceaccount.com
      operation: READ
```

```shell
planton apply -f managed-kafka-acl.yaml
```

This lets the billing worker's service account join the `billing` consumer group. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference the `GcpManagedKafkaCluster` from `cluster`; declare one ACL per topic or consumer group next to the `GcpManagedKafkaTopic` and the service that uses it.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The pattern** -- `aclId` is both the ACL's name and what it governs -- a literal topic or group, a prefix, a transactional id, or the whole cluster.

**Entries** -- each entry is a principal, an operation, and ALLOW (default) or DENY. Producers need WRITE (and DESCRIBE), consumers READ on the topic and on their group.

**Principals** -- `User:` followed by a Google service account or user email, or the certificate principal your cluster's mTLS mapping rules produce; `User:*` is everyone.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpManagedKafkaCluster** | `cluster` | `status.outputs.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The ACL's full resource name | Audit |
| `resource_type` | The resource type Google derived | Review |
| `pattern_type` | LITERAL or PREFIXED | Review |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One topic** -- A producer's WRITE and DESCRIBE and a consumer's READ on one topic. Start from the **Topic Producer and Consumer** preset.

**Consumer group** -- READ on the consumer group a service reads with. Start from the **Consumer Group Read** preset.

**Team namespace** -- ALL on every topic starting with a team's prefix. Start from the **Team Topic Prefix** preset.

## Works With

- [**GCP Managed Kafka Cluster**](/cloud-catalog/gcp-managed-kafka-cluster) -- the cluster the rules apply to
- [**GCP Managed Kafka Topic**](/cloud-catalog/gcp-managed-kafka-topic) -- the topics they govern
