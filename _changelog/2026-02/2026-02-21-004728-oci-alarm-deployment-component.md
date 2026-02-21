# OCI Alarm Deployment Component

**Date**: February 21, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Implemented the OciAlarm deployment component -- OCI's Monitoring alarm that evaluates metrics via Monitoring Query Language (MQL) expressions and triggers notifications to ONS topics or Streaming endpoints when thresholds are breached. Supports multi-threshold evaluation via overrides, configurable severity levels, notification formatting, and cross-compartment metric monitoring. First resource of Phase 9 (Monitoring and Logging).

## Problem Statement / Motivation

The OpenMCF Oracle Cloud provider needs observability infrastructure to enable proactive monitoring of OCI resources. Monitoring alarms are the foundational building block for alerting -- without them, users have no declarative way to define threshold-based notifications for their OCI workloads.

### Pain Points

- No monitoring or alerting component existed in the OCI provider catalog
- Teams running OCI workloads had no declarative way to provision metric-based alarms
- Phase 9 (Monitoring and Logging) was entirely unstarted

## Solution / What's New

A complete deployment component (`OciAlarm`) with proto API definitions, Pulumi module (Go), and Terraform module (HCL), registered as CloudResourceKind 3380.

### Key Design Decisions

**Severity and MessageFormat as embedded proto enums**: Lowercase enum values (`critical`, `error`, `warning`, `info` and `raw`, `pretty_json`, `ons_optimized`) per proto convention, converted to uppercase in IaC modules via `strings.ToUpper()`. The Severity zero-value `unspecified` is rejected by a CEL rule, making severity explicitly required.

**Overrides for multi-threshold alarms**: The `overrides` repeated message enables defining multiple evaluation rules (e.g., warn at 80%, critical at 95%) within a single alarm. Each override can customize the query, severity, notification body, and pending duration independently. Overrides are evaluated in order before the base rule.

**Suppression excluded**: The suppression block requires hardcoded RFC3339 timestamps (`time_suppress_from`, `time_suppress_until`). Hardcoded datetime values in declarative IaC become stale immediately after the window passes. Suppression is an operational control managed via OCI Console or CLI.

**`is_enabled` as plain bool**: Proto3 default of `false` means alarms start disabled unless explicitly set to `true`. This is the safe default -- users opt-in to active alarms rather than accidentally creating firing alarms.

**`destinations` as plain strings**: ONS topics (the primary alarm destination) are not OpenMCF components. Using `repeated string` instead of `repeated StringValueOrRef` keeps the spec honest about composability boundaries.

## Implementation Details

### Proto API (4 files)

- **spec.proto**: 20 fields, 2 embedded enums (Severity, MessageFormat), 1 nested message (AlarmOverride), 1 CEL rule (severity != unspecified)
- **api.proto**: Standard KRM wiring (OciAlarm, OciAlarmStatus)
- **stack_input.proto**: OciAlarmStackInput with target + provider config
- **stack_outputs.proto**: 1 output (`alarm_id`)

### Spec Fields

| Field | Type | Notes |
|-------|------|-------|
| `compartment_id` | StringValueOrRef (required) | default_kind: OciCompartment |
| `metric_compartment_id` | StringValueOrRef (required) | default_kind: OciCompartment |
| `namespace` | string (required) | Source service (e.g., oci_computeagent) |
| `query` | string (required) | MQL expression |
| `severity` | Severity enum (required) | CEL: != unspecified |
| `destinations` | repeated string (min 1) | ONS topic or Stream OCIDs |
| `is_enabled` | bool | Default false (safe) |
| `body` | string | Notification body with dynamic variables |
| `alarm_summary` | string | Customizable summary |
| `notification_title` | string | Email subject / Slack title |
| `pending_duration` | string | ISO 8601 (PT1M-PT1H) |
| `evaluation_slack_duration` | string | ISO 8601 (PT3M-PT2H) |
| `repeat_notification_duration` | string | ISO 8601 (PT1M-P30D) |
| `message_format` | MessageFormat enum | RAW, PRETTY_JSON, ONS_OPTIMIZED |
| `metric_compartment_id_in_subtree` | optional bool | Cross-compartment monitoring |
| `is_notifications_per_metric_dimension_enabled` | optional bool | Per-stream splitting |
| `resource_group` | string | Metric resource group filter |
| `notification_version` | string | e.g., "1.X" |
| `rule_name` | string | Base rule ID (default "BASE") |
| `overrides` | repeated AlarmOverride | Multi-threshold evaluation |

### Validation Tests

32 Ginkgo/Gomega tests (19 valid, 13 invalid scenarios) covering minimal configuration, all severity levels, notification fields, durations, message formats, subtree monitoring, per-dimension notifications, resource groups, overrides, multiple destinations, valueFrom refs, full configuration, and all required-field validation.

### Pulumi Module (5 files)

- `main.go`: Entry point with stack input loading
- `module/main.go`: Resources orchestrator with OCI provider setup
- `module/locals.go`: Locals struct with freeform tags from metadata labels
- `module/outputs.go`: Output constant (`alarm_id`)
- `module/alarm.go`: `alarmResource()` creating `monitoring.NewAlarm()` with conditional field assignment for all optional fields; `buildOverrides()` helper converting proto AlarmOverride messages to Pulumi `AlarmOverrideArgs` with severity uppercasing

### Terraform Module (5 files)

- `main.tf`: `oci_monitoring_alarm.this` with dynamic `overrides` block
- `locals.tf`: Freeform tags + severity_map and message_format_map for enum conversion
- `outputs.tf`: `alarm_id`
- `variables.tf`: Metadata and spec type definitions with optional fields
- `provider.tf`: OCI provider requirement (>= 5.0)

### Kind Registration

`OciAlarm = 3380` registered under new "Monitoring and Logging" section in `cloud_resource_kind.proto`, `kind_map_gen.go` regenerated.

## Benefits

- Enables declarative provisioning of OCI monitoring alarms with full MQL query support
- Multi-threshold evaluation via overrides provides sophisticated alerting patterns in a single resource
- Two OCID fields use StringValueOrRef for infra-chart composability (compartment, metric compartment)
- Starts Phase 9 (Monitoring and Logging), unlocking observability for all OCI workloads

## Impact

- **Users**: Can now define metric-based alarms with threshold evaluation, multi-level severity, and notification routing through a single YAML manifest
- **Platform**: Phase 9 (Monitoring and Logging) started -- 1/2 resources done
- **Infra Charts**: All 5 planned OCI infra charts can now incorporate monitoring alarms for their provisioned resources

## Related Work

- **OciLogGroup** (R33): Logging counterpart, second resource of Phase 9
- **OciCompartment** (R04): Compartment referenced via `compartment_id` and `metric_compartment_id`
- **OciStreamPool** (R30): Streaming endpoints can serve as alarm destinations

---

**Status**: Production Ready
