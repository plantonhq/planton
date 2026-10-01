# CloudWatch Alarm Health Check

This preset creates a health check that mirrors a CloudWatch alarm's state instead of probing an endpoint. It is the pattern for PRIVATE resources the Route 53 checker fleet cannot reach — internal load balancers, databases, anything without a public address — and for gating DNS on application-level signals like 5xx rates or queue depth.

## When to Use

- Failover routing for internal/private endpoints (split-horizon DNS)
- Gating DNS answers on application metrics rather than raw reachability
- Any resource where opening the firewall to the checker fleet is unacceptable

## Key Configuration Choices

- **Alarm mirroring** (`checkType: CLOUDWATCH_METRIC`) -- no probing; the check is healthy exactly when the alarm is in OK state
- **Regional alarm, global check** (`cloudwatchAlarmRegion`) -- CloudWatch alarms are regional objects even though the health check is global; both coordinates are required
- **Insufficient-data posture** (`insufficientDataHealthStatus: LastKnownStatus`) -- keeps the last state during metric gaps instead of flapping; use `Unhealthy` for fail-safe posture or `Healthy` for fail-open

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | Region for provider API calls | Your deployment region |
| `<alarm-name>` | The CloudWatch alarm to mirror | CloudWatch console or your AwsCloudwatchAlarm resource |
| `us-east-1` (`cloudwatchAlarmRegion`) | The region the alarm lives in | Where the alarm was created |

## Related Presets

- **01-https-endpoint** -- Use instead for publicly reachable endpoints
