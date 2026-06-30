# OciAlarm — Design Notes

## Design Rationale

OciAlarm provisions a single monitoring alarm resource. The component exposes the full richness of OCI's alarm API including multi-threshold overrides, notification customization, and dimension-level alerting.

### Why require metricCompartmentId separately from compartmentId?

The alarm resource lives in one compartment, but the metrics it evaluates may live in a different compartment. This is common in centralized monitoring setups where a monitoring compartment holds alarms that watch metrics across the organization. Keeping them as separate fields reflects the OCI API model and enables cross-compartment monitoring patterns.

### Why use lowercase enum values for severity and messageFormat?

Users type these values in YAML manifests. Lowercase (`critical`, `warning`, `raw`, `pretty_json`) is more natural to read and write. The IaC module maps them to uppercase strings expected by the OCI API.

### Why exclude suppression?

Suppression requires hardcoded RFC 3339 timestamps (`start_time`, `end_time`). Embedding specific timestamps in declarative IaC manifests is an anti-pattern — they go stale and require constant updates. Suppression is an operational action best managed via the OCI Console or CLI.

### Why exclude resolution?

The `resolution` parameter only accepts a single value (`"1m"`). Exposing a field with no meaningful choice adds complexity without value. The IaC module lets OCI apply the default.

### Why is isEnabled defaulting to false?

Proto3 booleans default to false. Since alarms that fire unexpectedly can cause alert fatigue, starting disabled is the safer default — operators explicitly enable alarms when they are ready. The Quick Start and Best Practices sections call this out.

### Why model overrides as a repeated message?

Overrides enable multi-threshold evaluation within a single alarm (e.g., warn at 70%, critical at 90%). Modeling them as a list matches the OCI API structure and allows operators to define any number of evaluation tiers. Each override can customize query, severity, body, and pending duration independently.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Separate compartmentId and metricCompartmentId | Enables cross-compartment monitoring | Two compartment fields to configure |
| Lowercase enum values | Natural YAML; better readability | IaC module maps to uppercase |
| Exclude suppression | No stale timestamps in IaC | Must manage suppression via Console/CLI |
| Exclude resolution | No single-value field noise | Cannot override if OCI adds values later |
| isEnabled defaults false | Safe default; no accidental alerting | Must explicitly enable |
| Overrides as repeated message | Full multi-threshold capability | More complex spec for simple alarms |

## Resource Graph

```
OciAlarm
└── oci_monitoring_alarm (always)
    ├── compartment_id (alarm location)
    ├── metric_compartment_id (metric source)
    ├── namespace, query, severity
    ├── destinations (1..N ONS topics or streams)
    ├── is_enabled
    ├── body, alarm_summary, notification_title
    ├── pending_duration, evaluation_slack_duration
    ├── repeat_notification_duration
    ├── message_format
    ├── metric_compartment_id_in_subtree
    ├── is_notifications_per_metric_dimension_enabled
    ├── resource_group, notification_version, rule_name
    ├── overrides (0..N)
    │   └── rule_name, query, severity, body, pending_duration
    └── outputs: alarm_id
```

## Deferred from v1

- **suppression** — requires hardcoded RFC 3339 timestamps; operational concern, not declarative IaC.
- **resolution** — only supported value is `"1m"`; no user choice.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.

## Freeform Tags

The module automatically populates freeform tags from metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciAlarm` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
