# OciQueue — Design Notes

## Design Rationale

OciQueue provisions a single queue resource. The component flattens the OCI provider's capability model into intuitive spec fields.

### Why flatten capabilities into spec fields?

The OCI Queue API models optional features (large messages, consumer groups) as a `capabilities` list with discriminated entries. This is an API design pattern that works for extensibility but makes YAML manifests harder to read. Flattening to `isLargeMessagesEnabled` (bool) and `consumerGroupConfig` (message) makes the intent clear at a glance. The IaC module reconstructs the capabilities list from these fields.

### Why is retentionInSeconds ForceNew?

The OCI Queue service does not support updating the retention period after creation. Changing retention requires destroying and recreating the queue. Marking it as ForceNew in the spec documentation ensures operators understand this constraint before setting the value.

### Why include channelConsumptionLimit?

Channel consumption limits prevent a single consumer channel from monopolizing queue resources in multi-tenant scenarios. While not commonly used, omitting it would prevent operators who need fine-grained throughput control from using Planton for their queue deployments.

### Why is deadLetterQueueDeliveryCount zero-disabling?

A value of 0 explicitly disables the DLQ, matching the OCI API behavior. This gives operators a clear way to opt out of dead-letter handling when they want messages to be retried indefinitely or handle failures at the application level.

### Why model consumerGroupConfig as a nested message?

Consumer group configuration involves multiple related fields (enable flag, display name, DLQ override). Grouping them in a nested message makes it clear that all fields are part of the same capability. Setting `consumerGroupConfig` implicitly enables the `CONSUMER_GROUPS` capability — no separate enable flag needed.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Flatten capabilities | Readable YAML; clear intent | Deviates from raw provider schema |
| retentionInSeconds ForceNew | Matches OCI API; no surprise recreation | Must plan retention at creation time |
| Capabilities are additive | Simple add-only model | Cannot disable large messages or consumer groups once enabled |
| Zero-disabling DLQ | Clear opt-out mechanism | Zero is a valid value with special meaning |
| consumerGroupConfig as message | Grouped related fields; implicit enable | One level of nesting |

## Resource Graph

```
OciQueue
└── oci_queue_queue (always)
    ├── custom_encryption_key_id (if set)
    ├── dead_letter_queue_delivery_count (if set)
    ├── retention_in_seconds (ForceNew)
    ├── timeout_in_seconds (if set)
    ├── visibility_in_seconds (if set)
    ├── channel_consumption_limit (if set)
    ├── capabilities (computed from isLargeMessagesEnabled + consumerGroupConfig)
    │   ├── LARGE_MESSAGES (if isLargeMessagesEnabled)
    │   └── CONSUMER_GROUPS (if consumerGroupConfig set)
    └── outputs: queue_id, messages_endpoint
```

## Deferred from v1

- **purge_trigger / purge_type** — imperative operational controls for queue purging; not declarative infrastructure. Manage via OCI Console or CLI.
- **primary_consumer_group_filter** — always empty for the primary group; no user-facing value.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.

## Freeform Tags

The module automatically populates freeform tags from metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciQueue` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
