---
title: "Preset: Error Rate Metric Math Alarm"
description: "**Use case:** Alert when the 5xx error rate exceeds a percentage threshold using CloudWatch Metric Math."
type: "preset"
rank: "02"
presetSlug: "02-error-rate-metric-math"
componentSlug: "cloudwatch-alarm"
componentTitle: "CloudWatch Alarm"
provider: "aws"
icon: "package"
order: 2
---

# Preset: Error Rate Metric Math Alarm

**Use case:** Alert when the 5xx error rate exceeds a percentage threshold using CloudWatch Metric Math.

This pattern uses three metric queries to compute a derived metric: `error_rate = errors / requests * 100`. Only the computed expression has `returnData: true`, which means the alarm evaluates the percentage — not the raw counts. This is the standard pattern for ratio-based alerting.

## What You Get

- A CloudWatch metric math alarm with 3 queries:
  - `errors` — raw 5xx error count from ALB
  - `requests` — total request count from ALB
  - `error_rate` — computed percentage (the alarm target)
- 2-of-3 evaluation (must breach 2 of the last 3 periods)
- 5-minute evaluation period (300 seconds)
- Threshold: 5% error rate
- Missing data treated as not breaching (safe for low-traffic periods)
- SNS notification on ALARM state

## The 3-Query Pattern

CloudWatch metric math alarms follow a consistent structure:

1. **Input metrics** — raw CloudWatch metrics with `returnData` omitted (defaults to false)
2. **Expression** — a math expression referencing the input metric IDs, with `returnData: true`

The alarm only evaluates the query where `returnData: true`. Input metrics are fetched but not directly compared to the threshold.

## When to Use

- ALB/NLB error rate monitoring
- API Gateway 4xx/5xx rate alerting
- Any ratio-based metric (cache hit rate, success rate, utilization percentage)
- When raw counts are misleading (10 errors out of 10 requests vs. 10 out of 1 million)

## What to Customize

| Field | Why |
|-------|-----|
| `threshold` | Adjust percentage — 1% for critical APIs, 10% for internal services |
| `dimensions.LoadBalancer` | Replace with your ALB's dimension value |
| `metricName` / `namespace` | Adapt for other services (API Gateway, CloudFront, etc.) |
| `expression` | Change the formula for different ratios |
| `alarmActions` | Replace with your SNS topic ARN |

## Cost

- **CloudWatch alarms**: $0.30/alarm/month (metric math alarms)
- **SNS**: first 1M notifications/month free
