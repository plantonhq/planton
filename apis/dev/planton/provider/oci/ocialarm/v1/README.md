# OciAlarm

## Overview

OciAlarm is an Planton component that deploys an OCI Monitoring Alarm. It provides a single declarative manifest to create a metric evaluation rule that triggers notifications when thresholds are breached, with support for multi-threshold overrides and configurable notification behavior.

## Purpose

OCI Monitoring evaluates metrics emitted by OCI services and custom applications. Alarms define MQL (Monitoring Query Language) expressions that OCI evaluates continuously. When the expression evaluates to true for the configured pending duration, the alarm transitions from OK to FIRING and sends notifications to ONS topics or Streaming streams. This component provisions the alarm rule; the metrics and notification infrastructure must exist separately.

## Key Features

- **MQL-based evaluation** — flexible metric queries using OCI's Monitoring Query Language with statistics, intervals, and group-by dimensions.
- **Multi-threshold overrides** — alternative evaluation parameters (query, severity, body, pending duration) for tiered alerting (e.g., warn at 70%, critical at 90%).
- **Configurable notification delivery** — notification body, title, and summary support dynamic variables (`{{severity}}`, `{{query}}`, `{{metricValues}}`, `{{resourceId}}`, `{{timestamp}}`).
- **Three notification formats** — `raw` (all destinations), `pretty_json` (human-readable JSON), `ons_optimized` (compact for email).
- **Per-metric-dimension alerting** — optional split notifications per metric stream dimension.
- **Sub-compartment evaluation** — optional evaluation of metrics across compartment hierarchies.
- **Foreign key references** — `compartmentId` and `metricCompartmentId` support `valueFrom`.

## Constraints

- All fields are updatable after creation (no ForceNew attributes).
- `severity` must be explicitly set — alarms without severity cannot fire.
- `destinations` must contain at least one OCID.
- Duration fields use ISO 8601 format (e.g., `PT5M` for 5 minutes, `PT1H` for 1 hour).
- `metricCompartmentIdInSubtree` can only be true when `metricCompartmentId` is a tenancy OCID.
- `pretty_json` and `ons_optimized` message formats only work with Notifications destinations, not Streaming.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| CPU utilization alert | `oci_computeagent` namespace, `CpuUtilization[5m].mean() > 80` |
| Database storage warning | `oci_autonomous_database` namespace, `StorageUtilization[1h].max() > 85` |
| Queue depth monitoring | `oci_queue` namespace, `QueueDepth[5m].max() > 1000` |
| VCN security list drops | `oci_vcn` namespace, `VnicEgressDropsSecurityList[5m].sum() > 100` |
| Tiered alerting | Overrides with warning at lower threshold, critical at higher |
| Cross-compartment monitoring | `metricCompartmentIdInSubtree: true` for tenancy-wide visibility |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels`.
- **Repeat notifications** — configurable re-notification frequency while alarm remains in FIRING state.
- **Dynamic variables** — notification body, summary, and title support runtime variable substitution for contextual alerts.
