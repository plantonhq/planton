---
title: "Preset: Production Multi-Action Alarm"
description: "**Use case:** Production-grade alarm with separate notification channels for ALARM, OK, and INSUFFICIENT_DATA state transitions."
type: "preset"
rank: "03"
presetSlug: "03-production-multi-action"
componentSlug: "cloudwatch-alarm"
componentTitle: "CloudWatch Alarm"
provider: "aws"
icon: "package"
order: 3
---

# Preset: Production Multi-Action Alarm

**Use case:** Production-grade alarm with separate notification channels for ALARM, OK, and INSUFFICIENT_DATA state transitions.

This pattern monitors SQS queue depth with a 3-of-5 M-of-N evaluation window, 1-minute periods for fast detection, and three separate action lists. This is the recommended production pattern — you know when something breaks, when it recovers, and when monitoring itself has gaps.

## What You Get

- A CloudWatch metric alarm on `AWS/SQS > ApproximateNumberOfMessagesVisible`
- 3-of-5 M-of-N evaluation (must breach 3 of the last 5 periods)
- 1-minute evaluation period (60 seconds) for fast detection
- Threshold: 1000 messages (Maximum statistic)
- Missing data treated as not breaching
- Three action channels:
  - `alarmActions` → critical alerts (e.g., PagerDuty, Slack #incidents)
  - `okActions` → recovery notifications (e.g., Slack #ops-resolved)
  - `insufficientDataActions` → monitoring gap warnings (e.g., Slack #ops-warnings)

## M-of-N Evaluation

CloudWatch evaluates the last N periods and triggers only if M of them breach the threshold:

- **evaluationPeriods** (N) = 5 — the sliding window size
- **datapointsToAlarm** (M) = 3 — how many must breach

This reduces false positives from transient spikes while still detecting sustained issues quickly. With 1-minute periods, the alarm can fire as fast as 3 minutes into a sustained issue.

## Multi-Action Pattern

Production alarms should use all three action types:

| Action Type | When It Fires | Typical Target |
|-------------|---------------|----------------|
| `alarmActions` | Metric breaches threshold | PagerDuty, Slack #critical |
| `okActions` | Alarm returns to OK state | Slack #resolved, auto-close ticket |
| `insufficientDataActions` | Metric data stops arriving | Slack #warnings, ops dashboard |

Using `okActions` prevents alert fatigue — teams see that an issue resolved without manual checking. Using `insufficientDataActions` catches silent failures where the monitored resource stops emitting metrics entirely.

## When to Use

- Any production workload requiring full lifecycle alerting
- SQS consumer lag monitoring
- Queue-based architectures (order processing, event pipelines)
- Workloads where recovery notification is as important as alert notification

## What to Customize

| Field | Why |
|-------|-----|
| `threshold` | Adjust based on normal queue depth and consumer throughput |
| `period` | Use 300s if 1-minute granularity is too noisy |
| `datapointsToAlarm` / `evaluationPeriods` | Tune the M-of-N window for your tolerance |
| `alarmActions` | Point to your critical alert SNS topic |
| `okActions` | Point to your resolution notification topic |
| `insufficientDataActions` | Point to your monitoring health topic |

## Cost

- **CloudWatch alarms**: $0.10/alarm/month (standard resolution)
- **SNS**: first 1M notifications/month free
