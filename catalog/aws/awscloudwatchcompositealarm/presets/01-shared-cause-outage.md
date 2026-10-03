# Shared-Cause Outage Page

**Use case:** Page on-call once for an outage with many symptoms.

When a database failure takes down CPU, latency, and error-rate alarms at
once, per-symptom paging creates an alert storm. This composite fires only
when BOTH constituent alarms are in ALARM — one page, one incident.

## What You Get

- A composite alarm over two existing metric alarms, combined with `AND`
- Critical-channel notification on ALARM, resolution notice on OK
- Constituent alarms keep evaluating and recording state; only the paging
  moves to the composite
- Outputs: `alarm_arn`, `alarm_name`

## When to Use

- Multiple symptom alarms that share a root cause
- Reducing pager noise without deleting the underlying alarms
- Building an incident signal from independent service-level alarms

## Key Configuration Choices

- **`AND` semantics** — both alarms must breach simultaneously; use `OR` for
  "any of these" paging, and parentheses with `NOT` for dependency-aware
  rules
- **Name-based composition** — the rule references alarms by their CloudWatch
  names; compose them from each `AwsCloudwatchAlarm`'s exported `alarm_name`
  output
- **Silence the constituents** — consider setting `actionsEnabled: false` on
  the underlying alarms so only the composite pages

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | Region of the composite and its constituent alarms | Your deployment region |
| `<cpu-alarm-name>` | Name of the first constituent alarm | The alarm's `alarm_name` output |
| `<error-rate-alarm-name>` | Name of the second constituent alarm | The alarm's `alarm_name` output |
