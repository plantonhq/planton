# OciLogGroup

## Overview

OciLogGroup is an Planton component that deploys an OCI Log Group with bundled logs. It provides a single declarative manifest to create the organizational container for OCI Logging service logs along with its constituent log definitions.

## Purpose

OCI Logging provides centralized log management for both OCI service logs and custom application logs. The log group is the top-level container — logs cannot exist outside a group. Service logs auto-collect from OCI services (VCN flow logs, Object Storage access logs, API Gateway logs, Functions invocation logs, etc.). Custom logs accept entries pushed via the Logging Ingestion API. This component bundles the group and its logs into a single manifest.

## Key Features

- **Two log types** — custom logs for application ingestion via the Logging Ingestion API, and service logs for auto-collection from OCI services.
- **Bundled logs** — logs are declared inline within the group, reflecting the OCI model where logs belong to exactly one group.
- **Service log configuration** — polymorphic source configuration referencing any OCI service, resource, and log category.
- **Configurable retention** — per-log retention in 30-day increments from 30 to 180 days.
- **Foreign key references** — `compartmentId` and service log `resource` support `valueFrom` for composability with any Planton component.

## Constraints

- `logType` and the entire `configuration` block are ForceNew — changing them forces log recreation.
- `isEnabled` and `retentionDuration` are updatable.
- `displayName` must be unique within the log group.
- `retentionDuration` must be a 30-day increment: 30, 60, 90, 120, 150, or 180.
- Service logs require `configuration`; custom logs ignore it.
- `configuration.source_type` is hardcoded to `"OCISERVICE"` — the only valid value.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Application log ingestion | Custom log type, push via Ingestion API |
| VCN flow log collection | Service log with `flowlogs` service and subnet resource |
| Object Storage audit trail | Service log with `objectstorage` service, `write` category |
| API Gateway access logs | Service log with `apigateway` service, `access` category |
| Functions invocation logs | Service log with `functionsInvoke` service, `invoke` category |
| Mixed observability group | Both custom and service logs in one group |

## Production Features

- **Freeform tags** — automatically populated on both the log group and all logs from `metadata.labels`.
- **Per-log retention** — each log can have its own retention period, enabling cost optimization for different log types.
- **Service log auto-collection** — zero-code log collection from supported OCI services.
