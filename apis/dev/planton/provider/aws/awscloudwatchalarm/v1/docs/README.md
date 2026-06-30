# AwsCloudwatchAlarm — Research Documentation

## Overview

Amazon CloudWatch Alarms are the core alerting primitive in AWS. A metric alarm continuously evaluates a single metric (or the result of a metric math expression) against a threshold, transitioning between three states — OK, ALARM, and INSUFFICIENT_DATA — and executing configured actions on each transition. This document covers the underlying evaluation model, design trade-offs, and operational guidance for the AwsCloudwatchAlarm deployment component.

## Alarm Evaluation Model

CloudWatch evaluates metric alarms on a fixed cadence determined by the alarm's `period`. At each evaluation cycle, CloudWatch:

1. **Retrieves data points** for the configured metric (or executes the metric math expression) covering the most recent `evaluation_periods × period` seconds.
2. **Applies the statistic** (Average, Sum, p99, etc.) to each period's raw data points, producing one aggregated value per period.
3. **Compares each aggregated value** against the threshold using the `comparison_operator`.
4. **Counts breaching data points** within the evaluation window.
5. **Applies the M-of-N rule**: if at least `datapoints_to_alarm` of the `evaluation_periods` data points breach, the alarm transitions to ALARM. If zero breach, it transitions to OK. States in between are governed by the `treat_missing_data` setting.

The evaluation is **eventually consistent** — there can be a delay of 1–2 minutes between a metric value being published and the alarm evaluating it. High-resolution alarms (10/20/30-second periods) reduce this lag but do not eliminate it entirely.

### State Machine

```
                  breach condition met
     ┌──────────────────────────────────────┐
     │                                      ▼
  ┌──────┐     metric data arrives     ┌─────────┐
  │  OK  │◄────────────────────────────│  ALARM  │
  └──────┘     breach condition cleared└─────────┘
     ▲                                      │
     │        no data / insufficient        │
     │     ┌───────────────────────────┐    │
     └─────│   INSUFFICIENT_DATA       │◄───┘
           └───────────────────────────┘
```

All three states can transition to any other state in a single evaluation cycle. The INSUFFICIENT_DATA state is the initial state for a newly created alarm and is re-entered when the metric stops publishing.

## M-of-N Evaluation

The M-of-N pattern (`datapoints_to_alarm` of `evaluation_periods`) is the single most important feature for reducing false positives in production alarms.

### How It Works

When `evaluation_periods = 5` and `datapoints_to_alarm = 3`:

- CloudWatch looks at the 5 most recent period windows.
- If 3 or more of those 5 data points breach the threshold, the alarm transitions to ALARM.
- If fewer than 3 breach (but more than 0), the alarm remains in its current state.
- If 0 breach, the alarm transitions to OK.

### Choosing M and N

| Scenario | Recommended M-of-N | Rationale |
|----------|-------------------|-----------|
| Critical infrastructure (RDS connections, disk space) | 3 of 3 | Every period must breach — no tolerance for false alarms on resources near hard limits |
| Application error rate | 2 of 3 or 3 of 5 | Tolerate transient spikes from deployments or retry storms |
| Batch job latency | 1 of 1 | Single-period evaluation acceptable because batch jobs have inherent jitter |
| Auto-scaling trigger | 3 of 5 | Smooth out traffic micro-bursts that don't warrant scaling |

### Interaction with Missing Data

Missing data points within the M-of-N window are resolved using `treat_missing_data` **before** the M-of-N count is applied. For example, with `treat_missing_data = "notBreaching"` and a window of `[breach, missing, breach, breach, missing]`, the resolved window is `[breach, OK, breach, breach, OK]` = 3 breaching data points.

## Missing Data Treatment

The `treat_missing_data` field controls how CloudWatch handles evaluation periods where no metric data exists. This is one of the most misunderstood alarm settings and a frequent source of unexpected alarm behavior.

### The Four Options

**`missing` (default)** — Missing data points are treated as missing. The alarm maintains its current state. This sounds safe but can cause **delayed alarm transitions**: if 3 of 5 periods are missing and 2 are breaching, CloudWatch cannot reach a conclusive decision and the alarm stays in its current state.

*When to use:* Metrics that publish intermittently but where the absence of data is not meaningful (e.g., Lambda invocation duration — no invocations means no data, which is fine).

**`notBreaching`** — Missing data is treated as within the threshold. The alarm will transition to OK if all periods are missing, and missing periods do not count toward the M-of-N breach count.

*When to use:* Error count metrics on low-traffic services. If there are no requests, there are no errors — this is healthy. Without `notBreaching`, the alarm would go to INSUFFICIENT_DATA during off-hours, potentially triggering `insufficient_data_actions`.

**`breaching`** — Missing data is treated as exceeding the threshold. The alarm will transition to ALARM if enough periods are missing.

*When to use:* Heartbeat metrics, health checks, or any metric where the absence of data signals a problem. For example, a custom metric published by an agent — if the agent stops publishing, the data is missing, and that itself is the failure.

**`ignore`** — Missing data points are entirely ignored. Only periods with actual data are evaluated. If all periods are missing, the alarm stays in its current state.

*When to use:* Metrics with known, regular gaps (e.g., a metric published every 5 minutes but the alarm period is 1 minute). The alarm only evaluates when data exists.

### Common Mistakes

| Mistake | Consequence | Fix |
|---------|-------------|-----|
| Using `missing` on an error-count metric for a low-traffic service | Alarm goes to INSUFFICIENT_DATA at night, triggering `insufficient_data_actions` | Use `notBreaching` |
| Using `notBreaching` on a heartbeat metric | Agent crash goes undetected because missing data is treated as healthy | Use `breaching` |
| Using `breaching` on a Lambda error metric | Alarm triggers during periods with zero invocations (no data = no errors = OK) | Use `notBreaching` |

## Simple Metric Mode vs. Metric Query Mode

The spec supports two mutually exclusive modes for defining what the alarm evaluates. The CEL validation `simple_metric_or_metric_queries` enforces this exclusivity.

### Simple Metric Mode

Set `metric_name`, `namespace`, `period`, and one of `statistic`/`extended_statistic`. Optionally add `dimensions` and `unit`.

**Use when:**
- Alarming on a single, pre-aggregated metric (CPUUtilization, MemoryUtilization, RequestCount)
- The metric's native statistic (Average, Sum, etc.) is exactly what you need
- No computation across multiple metrics is required

**Advantages:**
- Simpler YAML — fewer fields to configure
- Easier to read and maintain
- Lower cost (no metric math charges)

### Metric Query Mode

Set `metric_queries` with one or more `AwsCloudwatchAlarmMetricQuery` entries.

**Use when:**
- Computing a derived value: error rate (errors/total), latency ratio, custom KPIs
- Using anomaly detection (requires the `ANOMALY_DETECTION_BAND` function)
- Combining metrics from different namespaces or dimensions
- Applying cross-account monitoring
- Needing a statistic that simple mode doesn't support in the alarm context

**Advantages:**
- Full metric math expression language
- Anomaly detection support
- Multi-metric correlation
- Cross-account monitoring

### Decision Flowchart

```
Is the alarm on a single metric with a standard statistic?
├── YES → Use simple metric mode
└── NO
    ├── Do you need to compute a ratio, rate, or composite value?
    │   └── YES → Use metric query mode with metric math
    ├── Do you need anomaly detection?
    │   └── YES → Use metric query mode with ANOMALY_DETECTION_BAND
    └── Do you need cross-account monitoring?
        └── YES → Use metric query mode with account_id
```

## Metric Math

Metric math allows alarms to evaluate expressions that combine, transform, or aggregate multiple metrics. Expressions reference other queries by their `id` field.

### Common Patterns

**Error rate percentage:**
```
m1 = HTTPCode_Target_5XX_Count (Sum)
m2 = RequestCount (Sum)
e1 = m1/m2*100  ← alarm on this
```

**Latency percentile delta:**
```
m1 = TargetResponseTime (p99)
m2 = TargetResponseTime (p50)
e1 = m1 - m2  ← alarm when tail latency diverges from median
```

**Availability:**
```
m1 = HealthyHostCount (Minimum)
m2 = UnHealthyHostCount (Maximum)
e1 = m1/(m1+m2)*100  ← alarm when availability drops below 99%
```

**Rate of change (per-minute derivative):**
```
m1 = SomeMetric (Sum)
e1 = RATE(m1)  ← alarm on sudden changes
```

**Fill missing data in expressions:**
```
m1 = Errors (Sum)
e1 = FILL(m1, 0)  ← treat missing as zero for math
e2 = e1/m2*100
```

### Limits

- Maximum 20 metric queries per alarm
- Expression strings up to 1024 characters
- Metric query IDs must start with a lowercase letter and contain only `[a-z0-9_]`
- Exactly one query must have `return_data = true`
- Functions available: `METRICS()`, `RATE()`, `FILL()`, `ANOMALY_DETECTION_BAND()`, `PERIOD()`, `IF()`, `CEIL()`, `FLOOR()`, `ABS()`, `LOG()`, `LOG10()`, `STDDEV()`, arithmetic operators (`+`, `-`, `*`, `/`), and more

## Anomaly Detection

CloudWatch anomaly detection uses machine learning to model the expected behavior of a metric and create an "anomaly detection band" — a range of expected values. When the actual metric value falls outside this band, the alarm fires.

### How It Works

1. CloudWatch trains a model using 2 weeks of historical metric data (automatically, no user action required).
2. The model captures daily, weekly, and seasonal patterns.
3. The `ANOMALY_DETECTION_BAND(metricId, stdDevs)` function generates upper and lower bounds based on the model.
4. The `stdDevs` parameter (typically 2 or 3) controls band width — higher values mean wider bands and fewer false alarms.

### Configuration Pattern

An anomaly detection alarm requires:
1. A metric query (`m1`) that retrieves the raw metric with `return_data = true`
2. An expression query (`ad1`) that computes `ANOMALY_DETECTION_BAND(m1, 2)` with `return_data = false`
3. `threshold_metric_id = "ad1"` on the alarm spec
4. An anomaly detection comparison operator:
   - `LessThanLowerOrGreaterThanUpperThreshold` — outside either bound (most common)
   - `LessThanLowerThreshold` — only alert on unusually low values
   - `GreaterThanUpperThreshold` — only alert on unusually high values

### When to Use

- Metrics with **seasonal patterns** (request count varies by time of day, day of week)
- Metrics where the **"normal" range is hard to define** with a static number
- **Latency monitoring** where absolute thresholds miss gradual degradation

### When NOT to Use

- Metrics with well-known, fixed limits (disk space, connection count) — static thresholds are clearer
- Newly created metrics with less than 2 weeks of history — the model hasn't trained yet
- Extremely spiky metrics (batch job completion) — the model may not capture the pattern

### Cost

Anomaly detection alarms cost more than static threshold alarms. Each anomaly detection band counts as 3 custom metrics for pricing purposes.

## Actions

Actions are ARNs that CloudWatch invokes when the alarm transitions between states. The three action fields (`alarm_actions`, `ok_actions`, `insufficient_data_actions`) each accept up to 5 ARNs.

### Supported Action Targets

| Target | ARN Pattern | Use Case |
|--------|-------------|----------|
| **SNS Topic** | `arn:aws:sns:REGION:ACCOUNT:TOPIC` | Notifications (email, Slack, PagerDuty), Lambda fan-out |
| **Auto Scaling Policy** | `arn:aws:autoscaling:REGION:ACCOUNT:scalingPolicy:...` | Scale EC2 Auto Scaling groups up/down |
| **EC2 Action** | `arn:aws:automate:REGION:ec2:stop`, `ec2:terminate`, `ec2:recover`, `ec2:reboot` | Instance lifecycle automation |
| **Lambda Function** | `arn:aws:lambda:REGION:ACCOUNT:function:NAME` | Custom remediation (requires SNS as intermediary for alarm actions) |
| **SSM OpsItem** | `arn:aws:ssm:REGION:ACCOUNT:opsitem:severity#CATEGORY` | Create OpsCenter items for operational tracking |
| **Systems Manager Incident** | `arn:aws:ssm-incidents:REGION:ACCOUNT:response-plan/NAME` | Trigger Incident Manager response plans |

### Action Behavior

- Actions execute **once per state transition**, not continuously while in ALARM.
- If the alarm transitions ALARM → OK → ALARM, both `alarm_actions` and `ok_actions` fire (once each).
- Setting `actions_enabled = false` suppresses all actions without removing them. Useful for maintenance windows.
- If an action fails (e.g., SNS topic doesn't exist, IAM permission denied), CloudWatch logs the failure but does not retry.

### SNS Topic Convenience

The action fields use `StringValueOrRef` with `default_kind = AwsSnsTopic` and `default_kind_field_path = "status.outputs.topic_arn"`. This means `valueFrom` references default to resolving an AwsSnsTopic resource's ARN:

```yaml
alarmActions:
  - valueFrom:
      kind: AwsSnsTopic        # default_kind, can be omitted
      name: ops-alerts
      fieldPath: status.outputs.topic_arn  # default_kind_field_path, can be omitted
```

## Alarm States

### OK

The metric is within the threshold. The alarm entered this state because the evaluation found no breaching data points (or fewer than `datapoints_to_alarm` breaching points).

### ALARM

The metric has breached the threshold for the required number of data points. `alarm_actions` execute on the OK→ALARM or INSUFFICIENT_DATA→ALARM transition.

### INSUFFICIENT_DATA

Either the metric has not published enough data points for CloudWatch to evaluate (common for new alarms or stopped instances), or the metric does not exist at all. This is also the **initial state** for every newly created alarm.

### Transition Table

| From | To | Trigger | Actions Executed |
|------|----|---------|-----------------|
| OK | ALARM | Breach count >= `datapoints_to_alarm` | `alarm_actions` |
| OK | INSUFFICIENT_DATA | Not enough data to evaluate | `insufficient_data_actions` |
| ALARM | OK | Breach count = 0 | `ok_actions` |
| ALARM | INSUFFICIENT_DATA | Not enough data to evaluate | `insufficient_data_actions` |
| INSUFFICIENT_DATA | OK | Data arrives, no breach | `ok_actions` |
| INSUFFICIENT_DATA | ALARM | Data arrives, breach count >= M | `alarm_actions` |

## Evaluation Periods and Period Alignment

### Period

The `period` field defines two things:
1. **Aggregation window** — raw data points within each period are aggregated using the statistic.
2. **Evaluation cadence** — the alarm evaluates once per period.

CloudWatch aligns periods to UTC boundaries:
- 60-second periods align to the start of each minute
- 300-second periods align to 5-minute boundaries (00:00, 00:05, 00:10, ...)
- 3600-second periods align to the start of each hour

This alignment means a 5-minute alarm created at 12:03 will first evaluate at 12:05 (the next 5-minute boundary), not at 12:08.

### Evaluation Periods

The `evaluation_periods` field determines how many of these aligned periods constitute the evaluation window. The total evaluation window duration is `evaluation_periods × period` seconds.

| Period | Evaluation Periods | Window Duration | Practical Use |
|--------|-------------------|-----------------|---------------|
| 60s | 5 | 5 minutes | Fast detection of sustained issues |
| 300s | 3 | 15 minutes | Standard production alarm |
| 300s | 12 | 60 minutes | Slow-burn detection (gradual resource exhaustion) |
| 10s | 6 | 60 seconds | Real-time alerting for critical paths |

## High-Resolution Alarms

Standard CloudWatch metrics publish at 1-minute or 5-minute resolution. **High-resolution metrics** (published with `StorageResolution = 1` via the PutMetricData API) support sub-minute periods.

### Supported High-Resolution Periods

- **10 seconds** — Fastest evaluation. Use for critical real-time paths.
- **20 seconds** — Moderate resolution for near-real-time.
- **30 seconds** — Sub-minute with lower cost than 10s.

### How to Use

Set `period: 10` (or `20` or `30`) in the alarm spec. The metric must actually be published at high resolution — if the metric only has 1-minute data points, a 10-second alarm will see missing data in most periods.

### Cost Impact

High-resolution alarms are billed at the same per-alarm rate as standard alarms. However, high-resolution **custom metrics** (the data source) cost $0.30/metric/month vs. $0.30 for standard resolution (same price, but high-resolution data is stored for only 3 hours at full resolution before being aggregated).

### When to Use

- **Payment processing** — Detect errors within 30 seconds
- **Real-time bidding / ad tech** — Latency spikes must be caught in seconds
- **Health checks** — Heartbeat metrics published every 10 seconds

### When NOT to Use

- Standard AWS service metrics (CPUUtilization, RequestCount) — these publish at 1-minute or 5-minute intervals. A 10-second alarm on a 5-minute metric will see data in only 1 of 30 periods.
- Cost-sensitive environments where 1-minute evaluation is sufficient.

## Cost Model

### Per-Alarm Pricing (as of 2025, us-east-1)

| Alarm Type | Monthly Cost |
|-----------|-------------|
| Standard resolution metric alarm | $0.10/alarm |
| High-resolution metric alarm | $0.30/alarm |
| Anomaly detection alarm | $0.30/alarm + 3 custom metric equivalents |
| Composite alarm | $0.50/alarm |

### Metric Math Cost

Metric math queries that reference only standard AWS metrics do not incur additional metric charges. Queries that reference custom metrics are charged for each unique custom metric retrieved.

### Cost Optimization Tips

- **Consolidate where possible** — Use metric math to compute error rates in a single alarm rather than separate alarms for errors and request count.
- **Avoid high-resolution unless needed** — The 3x price difference between standard and high-resolution alarms adds up at scale.
- **Use M-of-N instead of shorter periods** — A 3-of-5 alarm with 5-minute periods is cheaper and often more useful than a 1-of-1 alarm with 1-minute periods.
- **Review INSUFFICIENT_DATA alarms** — Alarms stuck in INSUFFICIENT_DATA are still billed. Delete alarms for decommissioned resources.

## Security Considerations

### IAM Permissions

Creating and modifying alarms requires the following IAM actions:
- `cloudwatch:PutMetricAlarm` — create or update an alarm
- `cloudwatch:DeleteAlarms` — delete alarms
- `cloudwatch:DescribeAlarms` — read alarm configuration and state
- `cloudwatch:SetAlarmState` — manually set alarm state (useful for testing)

### SNS Topic Policies

For alarm actions to publish to an SNS topic, the topic's resource policy must allow `cloudwatch.amazonaws.com` to call `SNS:Publish`. If the SNS topic is in a different account, a cross-account policy is required.

Example SNS topic policy statement:
```json
{
  "Sid": "AllowCloudWatchAlarms",
  "Effect": "Allow",
  "Principal": {"Service": "cloudwatch.amazonaws.com"},
  "Action": "SNS:Publish",
  "Resource": "arn:aws:sns:us-east-1:123456789012:ops-alerts",
  "Condition": {
    "ArnLike": {
      "aws:SourceArn": "arn:aws:cloudwatch:us-east-1:123456789012:alarm:*"
    }
  }
}
```

### Cross-Account Alarms

Metric query mode supports the `account_id` field for cross-account monitoring. This requires:
1. A CloudWatch cross-account sharing role in the source account
2. The alarm account must be configured as a monitoring account in the source account's CloudWatch settings

### Least Privilege

- Grant `PutMetricAlarm` only to deployment pipelines and IaC roles, not application code.
- Use `cloudwatch:DescribeAlarms` for dashboards and read-only monitoring tools.
- Restrict `SetAlarmState` to break-glass roles — it can suppress real alarms.

## Common Patterns

### Error Rate Alarm (Most Common Pattern)

Compute error rate as a percentage and alarm when it exceeds a threshold. Use `notBreaching` for `treat_missing_data` because zero traffic means zero errors.

```
errors (Sum) / requests (Sum) * 100 → alarm at > 5%
```

### Queue Depth / Consumer Lag

Alarm on `ApproximateNumberOfMessagesVisible` (SQS) or `IteratorAge` (Kinesis). Use `missing` for `treat_missing_data` to avoid false alarms when the queue is empty.

### Heartbeat / Liveness

Publish a custom metric every N seconds. Alarm with `treat_missing_data = "breaching"` so that a stopped publisher triggers ALARM.

### Capacity Threshold

Alarm when a resource approaches its hard limit (RDS connections, DynamoDB consumed capacity, EC2 credit balance). Use static thresholds with `treat_missing_data = "breaching"`.

### Cost Anomaly

Use anomaly detection on the `EstimatedCharges` metric in the `AWS/Billing` namespace to detect unexpected cost spikes.

## Common Anti-Patterns

### Too-Sensitive Alarms

Setting `evaluation_periods = 1` with `period = 60` creates a hair-trigger alarm that fires on any 1-minute spike. Use M-of-N evaluation or longer periods to absorb transient noise.

### Wrong treat_missing_data

Using the default `missing` for all alarms is the most common configuration error. Every alarm should have an intentional `treat_missing_data` value based on what "no data" means for that specific metric.

### Alarm on Auto Scaling Metrics with Static Threshold

Metrics like CPU and memory on auto-scaling services naturally fluctuate as instances scale. Anomaly detection or higher thresholds prevent constant alarm flapping.

### Orphaned Alarms

Alarms pointing to deleted or renamed SNS topics silently fail to deliver notifications. Audit alarm action targets periodically.

### INSUFFICIENT_DATA Storms

Creating alarms before the monitored resource exists causes all alarms to enter INSUFFICIENT_DATA simultaneously, potentially flooding `insufficient_data_actions` targets.

## Relationship to Other CloudWatch Features

### Composite Alarms

Composite alarms evaluate the states of other alarms using boolean logic (AND, OR, NOT). They are the recommended way to:
- Reduce alarm noise by requiring multiple conditions
- Create hierarchical alerting (page only if both the error rate alarm AND latency alarm fire)
- Implement "at least N of these M alarms must be in ALARM" patterns

Composite alarms are a separate Terraform resource (`aws_cloudwatch_composite_alarm`) and will be a separate Planton component.

### CloudWatch Dashboards

Dashboards can display alarm widgets showing the current state, alarm history, and the metric graph with threshold line. Reference alarms by ARN from `status.outputs.alarm_arn`.

### Contributor Insights

Contributor Insights identifies the top-N contributors to a metric (e.g., the top IP addresses contributing to 5xx errors). It creates rules, not alarms — but you can alarm on the Contributor Insights metric output.

### Metric Filters

Metric filters extract numeric values from CloudWatch Logs and publish them as custom metrics. These custom metrics can then be alarmed on. The typical flow is:

```
Logs → Metric Filter → Custom Metric → Metric Alarm → SNS Topic → Notification
```

### CloudWatch Logs Insights

Logs Insights is an interactive query engine, not an alerting tool. For log-based alerting, use Metric Filters (above) or CloudWatch Logs Anomaly Detection.

### EventBridge Integration

CloudWatch Alarms automatically emit events to EventBridge on state transitions. The event detail type is `CloudWatch Alarm State Change`. This allows routing alarm events to targets beyond what the native action ARNs support (e.g., Step Functions, custom EventBridge rules with filtering).

## Terraform Provider Reference

The primary Terraform resource is `aws_cloudwatch_metric_alarm`. Key attributes:

- `alarm_name` (ForceNew) — alarm name is immutable; changing it recreates the alarm
- `comparison_operator` — required, 7 valid values
- `evaluation_periods` — required, >= 1
- `metric_name` / `namespace` / `period` / `statistic` — simple metric mode fields
- `metric_query` — dynamic block for metric math mode
- `threshold` vs. `threshold_metric_id` — mutually exclusive
- `treat_missing_data` — defaults to `"missing"` if not set
- `actions_enabled` — defaults to `true`

## Pulumi Resource Reference

The Pulumi resource is `cloudwatch.MetricAlarm` from `pulumi-aws/sdk/v7/go/aws/cloudwatch`. Input properties map directly to Terraform attributes with camelCase naming.

Key outputs:
- `Arn` — the alarm ARN
- `Name` — the alarm name (equal to the logical resource name)

The alarm resource name is derived from `metadata.name`, ensuring stable identity across updates.
