---
title: "Critical CPU Utilization Alarm"
description: "This preset creates a monitoring alarm that fires when average CPU utilization on compute instances exceeds 80% for 5 consecutive minutes. The alarm evaluates the `CpuUtilization` metric from the..."
type: "preset"
rank: "01"
presetSlug: "01-cpu-utilization-critical"
componentSlug: "alarm"
componentTitle: "Alarm"
provider: "oci"
icon: "package"
order: 1
---

# Critical CPU Utilization Alarm

This preset creates a monitoring alarm that fires when average CPU utilization on compute instances exceeds 80% for 5 consecutive minutes. The alarm evaluates the `CpuUtilization` metric from the `oci_computeagent` namespace using a 5-minute sliding window and delivers notifications to an ONS topic in human-readable JSON format. This is the most common starting point for OCI monitoring -- high CPU is the universal indicator of capacity pressure.

## When to Use

- Monitoring compute instances for sustained CPU pressure that may impact application performance
- Alerting operations teams when autoscaling thresholds are approaching or exceeded
- Establishing baseline monitoring for any new compute deployment
- Triggering automated remediation workflows (scaling, restart) via ONS topic subscribers

## Key Configuration Choices

- **oci_computeagent namespace** (`namespace: oci_computeagent`) -- targets the built-in compute agent metrics. The agent runs on all OCI compute instances and emits CpuUtilization, MemoryUtilization, and DiskUtilization without additional configuration.
- **5-minute mean threshold** (`query: CpuUtilization[5m].mean() > 80`) -- evaluates the average CPU over a 5-minute window. The `mean()` aggregation smooths transient spikes; the 80% threshold catches sustained pressure before it becomes critical. Adjust the threshold and window for your workload profile.
- **Critical severity** (`severity: critical`) -- signals immediate attention required. Use `warning` for early indicators and `critical` for conditions requiring prompt response.
- **5-minute pending duration** (`pendingDuration: PT5M`) -- the alarm must remain in the FIRING state for 5 minutes before notifications are sent. This prevents alert fatigue from brief CPU bursts that self-resolve.
- **Pretty JSON format** (`messageFormat: pretty_json`) -- delivers a structured, human-readable JSON notification body. Easier to parse for automated tooling and more readable in email notifications than the default RAW format.
- **Enabled on creation** (`isEnabled: true`) -- the alarm starts evaluating metrics immediately. Set to `false` to create the alarm definition without activating it.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment containing both the alarm and the monitored instances | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<ons-topic-ocid>` | OCID of the ONS notification topic for alarm delivery | OCI Console > Developer Services > Notifications > Topics |

## Related Presets

- **02-multi-threshold-escalation** -- use instead when you need tiered alerting (warning at 70%, critical at 90%) using a single alarm with overrides
