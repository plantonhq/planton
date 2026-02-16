---
title: "Autoscaling Production"
description: "This preset provisions a production Cloud Spanner instance with autoscaling enabled. Spanner automatically adjusts compute capacity between 1 and 3 nodes based on CPU and storage utilization targets...."
type: "preset"
rank: "03"
presetSlug: "03-autoscaling-production"
componentSlug: "spanner-instance"
componentTitle: "Spanner Instance"
provider: "gcp"
icon: "package"
order: 3
---

# Autoscaling Production

This preset provisions a production Cloud Spanner instance with autoscaling enabled. Spanner automatically adjusts compute capacity between 1 and 3 nodes based on CPU and storage utilization targets. Ideal for workloads with variable or unpredictable traffic patterns.

## When to Use

- Production workloads with variable traffic (e.g., e-commerce with peak hours, event-driven spikes)
- Applications where over-provisioning is too costly and under-provisioning risks latency
- Workloads that grow over time and need capacity to adjust without manual intervention
- Teams that want operational simplicity without capacity planning

## Key Configuration

- **Autoscaling** -- Spanner adjusts between 1 and 3 nodes automatically
- **CPU target: 65%** -- Scale up when high-priority CPU exceeds 65%. Google-recommended default for balanced performance/cost.
- **Storage target: 80%** -- Scale up when storage exceeds 80%. Leaves headroom for growth.
- **ENTERPRISE edition** -- required for granular autoscaling control
- **AUTOMATIC backup schedule** -- GCP creates backup schedules for new databases

## Scaling Behavior

- **Scale up**: When CPU or storage utilization exceeds targets, Spanner adds nodes (up to maxNodes)
- **Scale down**: When utilization drops below targets, Spanner removes nodes (down to minNodes)
- **Scale time**: Typically 5-10 minutes for capacity changes to take effect
- **Cost control**: maxNodes caps the maximum spend; minNodes ensures baseline performance

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the instance will be created | GCP Console or `GcpProject` outputs |
| `<instance-name>` | Name for this Spanner instance (6-30 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `prod-spanner-as`) |
| `<instance-config>` | Instance configuration (e.g., `regional-us-central1`) | [Spanner configurations](https://cloud.google.com/spanner/docs/instance-configurations) |
| `<display-name>` | Human-readable display name (4-30 chars) | Choose a descriptive name (e.g., `Autoscaling Spanner`) |

## Related Presets

- **01-free-instance** -- Zero-cost instance for development/testing
- **02-regional-production** -- Fixed capacity for predictable workloads
