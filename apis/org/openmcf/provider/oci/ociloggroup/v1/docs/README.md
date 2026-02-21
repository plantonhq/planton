# OciLogGroup — Design Notes

## Design Rationale

OciLogGroup bundles a log group with its constituent logs. The component supports both service logs (auto-collected from OCI services) and custom logs (pushed via the Ingestion API).

### Why bundle logs with the log group?

Logs belong to exactly one log group and cannot exist independently. Managing them separately would require explicit group references on every log without adding meaningful composability. Bundling keeps the logging topology in one manifest — operators see the full picture of what is being collected.

### Why use displayName as the resource key?

Log display names are unique within a group and serve as human-readable identifiers in the OCI Console and APIs. Using them as IaC resource keys provides stable, descriptive identifiers that survive re-applies without accidental recreation.

### Why flatten the configuration source?

The OCI provider nests service log configuration as `configuration > source`, where source is the only meaningful child (the sibling `compartment_id` is an optional override). Flattening removes the source wrapper, making the YAML more readable. The `source_type` field is hardcoded to `"OCISERVICE"` — it is the only valid value and offers no user choice.

### Why is the resource field a generic StringValueOrRef?

Service logs can collect from many different OCI resource types: VCNs, subnets, buckets, API gateways, functions, load balancers, etc. Using a generic `StringValueOrRef` without a fixed `default_kind` allows the field to reference any OpenMCF component via `valueFrom`, providing maximum composability across the OCI ecosystem.

### Why validate retentionDuration in 30-day increments?

OCI Logging only accepts retention periods in 30-day increments (30, 60, 90, 120, 150, 180). Validating this at the proto level with a CEL expression prevents deployment failures from invalid values.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Logs bundled with group | Single manifest; clear topology | Adding/removing one log re-applies the entire stack |
| displayName as resource key | Human-readable; stable across re-applies | Names must be unique within the group |
| Flatten configuration source | Simpler YAML; one level less nesting | Deviates from raw provider schema |
| Generic resource StringValueOrRef | Works with any OCI resource type | No default_kind for auto-completion |
| source_type hardcoded to OCISERVICE | No user confusion from a single-value field | Cannot support future source types without spec change |

## Resource Graph

```
OciLogGroup
├── oci_logging_log_group (always)
│   └── outputs: log_group_id
└── oci_logging_log (0..N, one per entry in logs)
    ├── log_type (CUSTOM or SERVICE)
    ├── is_enabled (optional)
    ├── retention_duration (optional, 30-day increments)
    ├── configuration (for SERVICE logs only)
    │   └── source: service, resource, category, parameters
    └── DependsOn: log_group
```

Each log declares `DependsOn` the log group to ensure correct creation order.

## Deferred from v1

- **configuration.source.source_type** — hardcoded to `"OCISERVICE"` (the only valid value). Not exposed in the spec.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.

## Freeform Tags

The module automatically populates freeform tags on both the log group and all logs:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciLogGroup` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
