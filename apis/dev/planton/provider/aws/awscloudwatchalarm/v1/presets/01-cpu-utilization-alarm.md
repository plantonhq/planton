# Preset: CPU Utilization Alarm

**Use case:** Alert when EC2 instance CPU exceeds a threshold, indicating high compute load.

This is the most common CloudWatch alarm pattern — a single-metric alarm that watches `CPUUtilization` in the `AWS/EC2` namespace. It uses 2-of-3 M-of-N evaluation to reduce noise from transient spikes, and treats missing data as breaching so you're alerted if the instance stops reporting.

## What You Get

- A CloudWatch metric alarm on `AWS/EC2 > CPUUtilization`
- 2-of-3 evaluation (must breach 2 of the last 3 periods)
- 5-minute evaluation period (300 seconds)
- Threshold: 80% (Average statistic)
- Missing data treated as breaching
- SNS notification on ALARM state

## When to Use

- Monitoring EC2 instance compute utilization
- Auto Scaling trigger thresholds
- Baseline CPU monitoring for any EC2 workload
- Detecting runaway processes or under-provisioned instances

## What to Customize

| Field | Why |
|-------|-----|
| `threshold` | Adjust based on workload — 70% for latency-sensitive, 90% for batch |
| `period` | Shorten to 60s for faster detection, lengthen for smoother average |
| `statistic` | Use `Maximum` if you care about peak spikes |
| `treatMissingData` | Change to `notBreaching` if instance stop/start is expected |
| `alarmActions` | Replace with your SNS topic ARN |

## Cost

- **CloudWatch alarms**: $0.10/alarm/month (standard resolution)
- **SNS**: first 1M notifications/month free
