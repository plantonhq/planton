# Multi-Threshold Escalation Alarm

This preset creates a single monitoring alarm with tiered alerting using the override mechanism. The base rule fires a WARNING when CPU exceeds 70% for 5 minutes, while the override escalates to CRITICAL when CPU exceeds 90% for 3 minutes. This eliminates the need for separate alarms per severity level and ensures consistent notification routing through a single ONS topic. Overrides are evaluated in order before the base rule, so the critical threshold is checked first.

## When to Use

- Operations teams that need tiered alerting (warn early, escalate if conditions worsen)
- Environments where creating multiple alarms for the same metric creates management overhead
- Runbooks that define different response procedures based on severity (e.g., warning = monitor, critical = page on-call)
- Any scenario where a single metric needs multiple evaluation thresholds with different urgency levels

## Key Configuration Choices

- **Base rule at WARNING 70%** (`query: CpuUtilization[5m].mean() > 70`, `severity: warning`) -- the base rule fires first, providing early warning that CPU utilization is trending high. This gives teams time to investigate before conditions become critical.
- **Override at CRITICAL 90%** (`overrides[0].query: CpuUtilization[5m].mean() > 90`, `severity: critical`) -- the override escalates severity when CPU is critically high. Overrides are evaluated before the base rule in list order, so this check takes precedence when both conditions are true.
- **Shorter pending duration for critical** (`overrides[0].pendingDuration: PT3M` vs base `PT5M`) -- critical conditions trigger faster (3 minutes vs 5 minutes) because at 90% CPU, waiting longer risks service degradation.
- **Explicit rule names** (`ruleName: BASE`, `overrides[0].ruleName: critical-threshold`) -- rule names must be unique across the alarm. They appear in notification payloads and API responses, making it clear which threshold triggered the alarm.
- **Custom body per threshold** -- the override carries its own notification body with urgency-appropriate messaging, so responders immediately know the severity without parsing the query.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment containing both the alarm and the monitored instances | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<ons-topic-ocid>` | OCID of the ONS notification topic for alarm delivery | OCI Console > Developer Services > Notifications > Topics |

## Related Presets

- **01-cpu-utilization-critical** -- use instead when a single threshold is sufficient and the override pattern is not needed
