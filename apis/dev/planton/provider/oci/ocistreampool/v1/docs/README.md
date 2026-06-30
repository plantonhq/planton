# OciStreamPool — Design Notes

## Design Rationale

OciStreamPool bundles the stream pool and its streams into a single component. This matches how operators think about streaming infrastructure: a pool and its topics are one logical unit.

### Why bundle streams with the pool?

Streams belong to exactly one pool and inherit the pool's Kafka settings, encryption, and networking. Managing them separately would require explicit pool references on every stream and would not add meaningful composability — a stream cannot be moved between pools. Bundling keeps the streaming topology in one manifest, making it easy to understand the full picture.

### Why use the stream `name` as the resource key?

OCI Streaming streams are identified by name within a pool. The name is also the Kafka topic name used by producers and consumers. Using it as the IaC resource key provides a stable, human-readable identifier that survives re-applies without accidental recreation.

### Why is `privateEndpointSettings` entirely ForceNew?

OCI does not support updating a stream pool's private endpoint configuration after creation. Changing the subnet, NSGs, or IP requires pool recreation. Marking the entire block as ForceNew reflects this API behavior and prevents confusing errors during updates.

### Why support both Kafka auto-topic creation and explicit streams?

Auto-topic creation is useful for migration scenarios where Kafka producers dynamically create topics. Explicit streams provide version-controlled, auditable infrastructure. Supporting both gives operators flexibility based on their maturity level and use case.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Streams bundled with pool | Single manifest; clear topology | Adding/removing one stream re-applies the entire stack |
| Stream name as resource key | Human-readable; stable across re-applies | Names must be unique within the pool |
| ForceNew on private endpoint | Matches OCI API behavior exactly | Changing networking requires full pool recreation |
| KMS key updatable | Can re-encrypt without recreation | Brief encryption transition period during update |
| Kafka bootstrap servers as output | Ready-to-use connection string for clients | Computed via ApplyT; not available until pool creation completes |

## Resource Graph

```
OciStreamPool
├── oci_streaming_stream_pool (always)
│   ├── kafka_settings (optional)
│   ├── custom_encryption_key (if kmsKeyId is set)
│   └── private_endpoint_settings (if privateEndpointSettings is set)
└── oci_streaming_stream (0..N, one per entry in streams)
    └── DependsOn: stream_pool
```

Each stream declares `DependsOn` the pool to ensure correct creation order.

## Deferred from v1

- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.
- **security_attributes** — Oracle ZPR (Zero-Trust Packet Routing) attributes; very low adoption.

## Freeform Tags

The module automatically populates freeform tags on both the pool and all streams:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciStreamPool` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
