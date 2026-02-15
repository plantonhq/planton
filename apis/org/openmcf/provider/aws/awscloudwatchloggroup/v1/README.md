# AwsCloudwatchLogGroup

A **CloudWatch Logs log group** is a container for log streams that share the same retention, monitoring, and access control settings. It is the primary destination for application logs, service logs, and operational data across AWS.

## When to Use

- **Centralized application logging** — Pre-create log groups with retention policies before deploying Lambda, ECS, or EKS workloads.
- **Compliance and audit** — Enforce KMS encryption and specific retention periods to meet regulatory requirements (HIPAA, SOC2, PCI-DSS).
- **Cross-resource log destinations** — Create log groups that Step Functions, API Gateway, OpenSearch, and other services reference for their logging configuration.
- **Cost optimization** — Use INFREQUENT_ACCESS class for high-volume logs (VPC flow logs, CDN access logs) that are rarely queried.

## When NOT to Use

- For application-level log routing or filtering — use CloudWatch Logs subscription filters or metric filters instead.
- If you only need logs from a single Lambda function — Lambda auto-creates log groups (though pre-creating gives you retention and encryption control).

## Prerequisites

- An AWS account and region configured in your OpenMCF stack input.
- (Optional) A KMS key if you need customer-managed encryption.

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `retentionInDays` | int | No | 0 (never expire) | Days to retain log events. Must be one of: 0, 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653. |
| `kmsKeyId` | StringValueOrRef | No | — | KMS key ARN for encrypting log data at rest. Can reference AwsKmsKey via `valueFrom`. |
| `logGroupClass` | string | No | STANDARD | Log group class: `STANDARD`, `INFREQUENT_ACCESS`, or `DELIVERY`. ForceNew. |
| `deletionProtectionEnabled` | bool | No | false | Prevents accidental deletion. Note: not yet implemented in IaC modules (provider version limitation). |

**ForceNew warning:** `logGroupClass` triggers log group replacement when changed. Choose carefully at creation time.

**DELIVERY class note:** When `logGroupClass` is `DELIVERY`, `retentionInDays` must not be set (AWS manages retention for Delivery log groups).

## Outputs

| Output | Description |
|--------|-------------|
| `log_group_arn` | Log group ARN. Primary reference for Step Functions, API Gateway, OpenSearch, and other services. |
| `log_group_name` | Log group name. Used by services that reference log groups by name (ElastiCache, ECS awslogs driver). |

## Minimal Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: app-logs
spec:
  retentionInDays: 30
```

## Production Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: prod-app-logs
spec:
  retentionInDays: 90
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: log-encryption-key
      fieldPath: status.outputs.key_arn
```

Then reference from a Step Function:

```yaml
spec:
  logging:
    level: ERROR
    logDestination:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: prod-app-logs
        fieldPath: status.outputs.log_group_arn
```

## What Is Deliberately Omitted (v1)

- **Log streams** — Created automatically by services writing to the log group.
- **Metric filters** — Separate CloudWatch resource with independent lifecycle.
- **Subscription filters** — Separate CloudWatch resource for log forwarding.
- **Resource policy** — Separate resource for cross-account/service access control.
- **`skip_destroy`** — IaC lifecycle management, not an AWS API property. OpenMCF has its own resource lifecycle.
- **`name_prefix`** — OpenMCF derives the name from `metadata.name`.
