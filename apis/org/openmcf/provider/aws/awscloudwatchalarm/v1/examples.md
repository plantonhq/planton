# AwsCloudwatchAlarm Examples

## 1. Minimal CPU Alarm (EC2 Instance)

The simplest alarm: alert when an EC2 instance's average CPU exceeds 80% for 3 consecutive 5-minute periods.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchAlarm
metadata:
  name: ec2-high-cpu
  org: acme
  env: prod
  id: ec2-high-cpu-prod
spec:
  comparisonOperator: GreaterThanThreshold
  evaluationPeriods: 3
  threshold: 80.0
  metricName: CPUUtilization
  namespace: AWS/EC2
  period: 300
  statistic: Average
  dimensions:
    InstanceId: i-0abc123def456789a
  treatMissingData: breaching
  alarmDescription: "EC2 instance i-0abc123def456789a CPU > 80% for 15 minutes"
  alarmActions:
    - value: arn:aws:sns:us-east-1:123456789012:ops-alerts
```

**What this creates:** A static threshold alarm on the `CPUUtilization` metric for a single EC2 instance. Missing data is treated as breaching because a stopped instance that isn't reporting metrics should trigger investigation. The alarm evaluates every 5 minutes (period=300) and requires all 3 periods to breach.

---

## 2. ECS Service Memory Alarm with M-of-N

Alert when an ECS service's memory utilization exceeds 90% in at least 3 of the last 5 evaluation periods. The M-of-N pattern avoids false alarms from brief memory spikes during deployments.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchAlarm
metadata:
  name: ecs-memory-high
  org: acme
  env: prod
  id: ecs-memory-high-prod
spec:
  comparisonOperator: GreaterThanOrEqualToThreshold
  evaluationPeriods: 5
  datapointsToAlarm: 3
  threshold: 90.0
  metricName: MemoryUtilization
  namespace: AWS/ECS
  period: 60
  statistic: Average
  dimensions:
    ClusterName: prod-cluster
    ServiceName: api-service
  treatMissingData: notBreaching
  alarmDescription: "ECS api-service memory >= 90% in 3 of last 5 minutes. Check for memory leaks or scale out."
  alarmActions:
    - valueFrom:
        kind: AwsSnsTopic
        name: infra-alerts
        fieldPath: status.outputs.topic_arn
```

**What this creates:** A 3-of-5 M-of-N alarm on ECS memory. With 60-second periods and 5 evaluation periods, the alarm looks at a 5-minute sliding window. Missing data is treated as not-breaching because ECS services may briefly stop reporting during rolling deployments. The SNS topic is referenced via `valueFrom` rather than a hardcoded ARN.

---

## 3. ALB 5xx Error Rate Using Metric Math

Compute the 5xx error rate as a percentage and alarm when it exceeds 5%. This uses two raw metric queries combined by a metric math expression.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchAlarm
metadata:
  name: alb-5xx-rate
  org: acme
  env: prod
  id: alb-5xx-rate-prod
spec:
  comparisonOperator: GreaterThanThreshold
  evaluationPeriods: 3
  datapointsToAlarm: 2
  threshold: 5.0
  treatMissingData: notBreaching
  alarmDescription: "ALB 5xx error rate > 5% for 2 of 3 periods. Investigate backend health."
  metricQueries:
    - id: errors
      metric:
        metricName: HTTPCode_Target_5XX_Count
        namespace: AWS/ApplicationELB
        period: 300
        stat: Sum
        dimensions:
          LoadBalancer: app/prod-alb/a1b2c3d4e5f6g7h8
      returnData: false
    - id: requests
      metric:
        metricName: RequestCount
        namespace: AWS/ApplicationELB
        period: 300
        stat: Sum
        dimensions:
          LoadBalancer: app/prod-alb/a1b2c3d4e5f6g7h8
      returnData: false
    - id: error_rate
      expression: "errors/requests*100"
      label: "5xx Error Rate (%)"
      returnData: true
  alarmActions:
    - value: arn:aws:sns:us-east-1:123456789012:ops-critical
```

**What this creates:** A metric math alarm with three queries. `errors` retrieves the 5xx count, `requests` retrieves the total request count, and `error_rate` computes the percentage. Only `error_rate` has `returnData: true`, making it the evaluation signal. Missing data is not-breaching so the alarm stays quiet during zero-traffic windows (division by zero is treated as missing by CloudWatch).

---

## 4. SQS Queue Depth Alarm with SNS valueFrom

Alert when an SQS queue accumulates more than 1000 visible messages, indicating consumers are falling behind.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchAlarm
metadata:
  name: orders-queue-depth
  org: acme
  env: prod
  id: orders-queue-depth-prod
spec:
  comparisonOperator: GreaterThanOrEqualToThreshold
  evaluationPeriods: 2
  threshold: 1000.0
  metricName: ApproximateNumberOfMessagesVisible
  namespace: AWS/SQS
  period: 300
  statistic: Average
  dimensions:
    QueueName: prod-orders-queue
  treatMissingData: missing
  alarmDescription: "Orders queue depth >= 1000 messages for 10 minutes. Scale consumers or investigate dead letters."
  alarmActions:
    - valueFrom:
        kind: AwsSnsTopic
        name: ops-alerts
        fieldPath: status.outputs.topic_arn
  okActions:
    - valueFrom:
        kind: AwsSnsTopic
        name: ops-alerts
        fieldPath: status.outputs.topic_arn
```

**What this creates:** A queue depth alarm with both ALARM and OK actions pointing to the same SNS topic. Teams receive a notification when the queue backs up and a recovery notification when it drains. The SNS topic is referenced via `valueFrom`, creating an infrastructure dependency edge.

---

## 5. Anomaly Detection Alarm (ML-Based Threshold)

Use CloudWatch's machine-learning anomaly detection to alert when a metric deviates from its expected band. Ideal for metrics with daily/weekly patterns where a static threshold is too rigid.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchAlarm
metadata:
  name: api-latency-anomaly
  org: acme
  env: prod
  id: api-latency-anomaly-prod
spec:
  comparisonOperator: LessThanLowerOrGreaterThanUpperThreshold
  evaluationPeriods: 3
  thresholdMetricId: ad1
  treatMissingData: missing
  alarmDescription: "API latency anomaly detected — p99 latency outside expected band for 3 periods."
  metricQueries:
    - id: m1
      metric:
        metricName: TargetResponseTime
        namespace: AWS/ApplicationELB
        period: 300
        stat: p99
        dimensions:
          LoadBalancer: app/prod-alb/a1b2c3d4e5f6g7h8
      returnData: true
    - id: ad1
      expression: "ANOMALY_DETECTION_BAND(m1, 2)"
      label: "Expected p99 Latency Band"
      returnData: false
  alarmActions:
    - valueFrom:
        kind: AwsSnsTopic
        name: ops-alerts
        fieldPath: status.outputs.topic_arn
```

**What this creates:** An anomaly detection alarm with two metric queries. `m1` retrieves the p99 latency, and `ad1` computes an anomaly detection band with 2 standard deviations. The `thresholdMetricId` points to `ad1`, telling CloudWatch to use the band as a dynamic threshold. The comparison operator `LessThanLowerOrGreaterThanUpperThreshold` triggers when the actual value is above the upper band or below the lower band.

---

## 6. High-Resolution Alarm (10-Second Period) for Lambda Errors

Monitor Lambda function errors at 10-second resolution for near-real-time alerting. High-resolution alarms are useful for latency-sensitive workloads where 5-minute evaluation is too slow.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchAlarm
metadata:
  name: lambda-errors-hires
  org: acme
  env: prod
  id: lambda-errors-hires-prod
spec:
  comparisonOperator: GreaterThanThreshold
  evaluationPeriods: 6
  datapointsToAlarm: 4
  threshold: 0.0
  metricName: Errors
  namespace: AWS/Lambda
  period: 10
  statistic: Sum
  dimensions:
    FunctionName: prod-payment-processor
  treatMissingData: notBreaching
  alarmDescription: "Lambda payment-processor errors detected at 10s resolution. 4 of 6 periods must breach."
  alarmActions:
    - value: arn:aws:sns:us-east-1:123456789012:pager-critical
```

**What this creates:** A high-resolution alarm that evaluates every 10 seconds with a 4-of-6 M-of-N window (60-second effective window). The threshold is 0 with `GreaterThanThreshold`, so any error triggers the alarm. Missing data is not-breaching because periods with no invocations (and therefore no errors) are healthy. Note: high-resolution alarms cost more than standard alarms.

---

## 7. Production-Ready Multi-Action Alarm with OK + Insufficient Data Actions

A comprehensive production alarm that notifies different teams for different state transitions. Uses separate SNS topics for critical alerts vs. informational recovery notices.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchAlarm
metadata:
  name: rds-connections-critical
  org: acme
  env: prod
  id: rds-connections-critical-prod
spec:
  comparisonOperator: GreaterThanOrEqualToThreshold
  evaluationPeriods: 5
  datapointsToAlarm: 3
  threshold: 450.0
  metricName: DatabaseConnections
  namespace: AWS/RDS
  period: 60
  statistic: Maximum
  dimensions:
    DBInstanceIdentifier: prod-primary-db
  treatMissingData: breaching
  alarmDescription: >
    RDS prod-primary-db connections >= 450 (of 500 max) in 3 of 5 minutes.
    Immediate action: check for connection leaks, idle connections, or
    misconfigured connection pools. Escalate if sustained.
  alarmActions:
    - valueFrom:
        kind: AwsSnsTopic
        name: db-critical-alerts
        fieldPath: status.outputs.topic_arn
    - value: arn:aws:sns:us-east-1:123456789012:pagerduty-oncall
  okActions:
    - valueFrom:
        kind: AwsSnsTopic
        name: db-recovery-notices
        fieldPath: status.outputs.topic_arn
  insufficientDataActions:
    - valueFrom:
        kind: AwsSnsTopic
        name: infra-monitoring-meta
        fieldPath: status.outputs.topic_arn
```

**What this creates:** A production alarm with three distinct action targets:
- **ALARM** — Sends to both `db-critical-alerts` (Slack/email for the DBA team) and a PagerDuty SNS integration (pager for on-call).
- **OK** — Sends a recovery notice to `db-recovery-notices` so the team knows the issue resolved.
- **INSUFFICIENT_DATA** — Sends to `infra-monitoring-meta` to alert the monitoring team that the metric stream itself may be broken (RDS instance down, CloudWatch agent failure, etc.).

Missing data is treated as breaching because an RDS instance that stops reporting connection metrics is itself an emergency. The 3-of-5 M-of-N pattern absorbs brief connection count spikes during connection pool rebalancing.
