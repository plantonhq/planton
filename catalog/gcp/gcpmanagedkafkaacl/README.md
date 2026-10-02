# GCP Managed Kafka ACL

The Kafka access rules for one resource pattern on a Managed Service for Apache Kafka cluster -- one topic, one consumer group, a prefix of either, a transactional id, or the cluster itself -- as up to 100 ALLOW or DENY entries for principals. ACLs are their own block so the team that owns a topic or a consumer group grants access to it in its own manifest, without editing the cluster.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Kafka ACL** -- a `managed_kafka_acl` for the pattern named by `aclId`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Managed Service for Apache Kafka admin permissions (`roles/managedkafka.admin`) on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpManagedKafkaCluster`** -- the cluster the ACL applies to (`cluster`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpManagedKafkaAcl
metadata:
  name: orders-topic
spec:
  location: us-central1
  cluster:
    valueFrom:
      kind: GcpManagedKafkaCluster
      name: events
      fieldPath: status.outputs.name
  aclId: topic/orders
  aclEntries:
    - principal: User:orders-api@my-project.iam.gserviceaccount.com
      operation: WRITE
    - principal: User:billing-worker@my-project.iam.gserviceaccount.com
      operation: READ
```

```shell
planton apply -f managed-kafka-acl.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The cluster's region. Immutable. |
| `cluster` | `StringValueOrRef` | The cluster -- a `GcpManagedKafkaCluster` reference or its full path or bare id. Immutable. |
| `aclId` | `string` | The resource pattern: `cluster`, `topic/{name}`, `consumerGroup/{name}`, `transactionalId/{name}`, or a `*Prefixed/{prefix}` form. Immutable. |
| `aclEntries` | `list` | 1-100 entries: `principal` (`User:` + Google account, or `User:*`), `operation`, optional `permissionType` and `host`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The cluster's project. |
| `aclEntries[].permissionType` | `string` | `ALLOW` | `ALLOW` or `DENY`; a DENY wins. |
| `aclEntries[].host` | `string` | `*` | Managed Kafka accepts only `*`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `aclId` follows Google's resource-pattern grammar.
- Principals start with `User:`; operations and permission types take only Google's values; `host` is `*`.
- 1-100 entries.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/clusters/{cluster}/acls/{acl_id}` |
| `resource_type` | `string` | `CLUSTER`, `TOPIC`, `GROUP`, or `TRANSACTIONAL_ID` |
| `resource_name` | `string` | The resource named by the pattern (`kafka-cluster` for the cluster) |
| `pattern_type` | `string` | `LITERAL` or `PREFIXED` |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **One ACL per pattern.** A cluster holds one ACL per resource pattern, so two manifests must never declare the same `aclId` -- give each pattern one owner.
- **IAM first.** Kafka ACLs decide what an authenticated client may do; connecting at all still needs `roles/managedkafka.client`.
- **Consumers need two ACLs.** READ on `topic/{name}` and READ on `consumerGroup/{group}`.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpManagedKafkaCluster** -- the cluster the ACL applies to
- **GcpManagedKafkaTopic** -- the topics a `topic/` pattern names

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
