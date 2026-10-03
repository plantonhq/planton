# AwsCloudwatchLogGroup

A **CloudWatch Logs log group** is a container for log streams that share the same retention, monitoring, and access control settings. It is the primary destination for application logs, service logs, and operational data across AWS.

Beyond the container itself, this component folds in the log-group-scoped satellites that share the group's lifecycle: **metric filters** (turn matching log events into CloudWatch metrics), **subscription filters** (stream matching events to Kinesis, Firehose, or Lambda in real time), a **data protection policy** (audit and mask PII in log events), a **field index policy** (accelerate Logs Insights queries on selected fields), a **transformer** (an ingest-time processor pipeline that parses and reshapes every event), and **pre-created log streams** (for fixed-name agent configs and stream-scoped IAM).

## When to Use

- **Centralized application logging** — Pre-create log groups with retention policies before deploying Lambda, ECS, or EKS workloads.
- **Compliance and audit** — Enforce KMS encryption, deletion protection, specific retention periods, and PII masking to meet regulatory requirements (HIPAA, SOC2, PCI-DSS).
- **Log-derived alerting** — Attach metric filters that count errors or extract latencies from log events, then alarm on the resulting metrics with `AwsCloudwatchAlarm`.
- **Real-time log fan-out** — Attach subscription filters that deliver matching events to analytics pipelines (Kinesis), archival (Firehose), or custom processing (Lambda).
- **Cross-resource log destinations** — Create log groups that Step Functions, API Gateway, OpenSearch, Route 53 query logging, and other services reference for their logging configuration.
- **Cost optimization** — Use INFREQUENT_ACCESS class for high-volume logs (VPC flow logs, CDN access logs) that are rarely queried.

## When NOT to Use

- If you only need logs from a single Lambda function and defaults are fine — Lambda auto-creates log groups (though pre-creating gives you retention, encryption, and policy control).
- For account-wide log governance (account-level data protection or subscription policies) — those are account-scoped settings, not per-group resources.

## Prerequisites

- An AWS account and region configured in your Planton IaC input.
- (Optional) A KMS key if you need customer-managed encryption — its key policy must allow the `logs.<region>.amazonaws.com` service principal.
- (Optional) For subscription filters to Kinesis/Firehose: an IAM role trusting `logs.amazonaws.com` with put permissions on the destination.

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | — | AWS region for the log group. |
| `retentionInDays` | int | No | 0 (never expire) | Days to retain log events. Must be one of: 0, 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653. |
| `kmsKeyId` | StringValueOrRef | No | — | KMS key ARN for encrypting log data at rest. Can reference AwsKmsKey via `valueFrom`. |
| `logGroupClass` | string | No | STANDARD | Log group class: `STANDARD`, `INFREQUENT_ACCESS`, or `DELIVERY`. ForceNew. |
| `deletionProtectionEnabled` | bool (tri-state) | No | unset | Blocks every delete (including IaC destroy) until cleared. Unset keeps the group's current state; **only an explicit `false` disables protection once enabled**. |
| `metricFilters[]` | list | No | — | Metric filters extracting CloudWatch metrics from matching log events. Unique names per group; STANDARD class only. |
| `subscriptionFilters[]` | list | No | — | Real-time delivery of matching events to Kinesis/Firehose/Lambda. Maximum 2 per group (AWS limit); STANDARD class only. |
| `dataProtectionPolicy` | object | No | — | CloudWatch Logs data protection policy document (PII audit + masking). One per group. |
| `fieldIndexPolicy` | object | No | — | Field index policy document (`{"Fields": [...]}`) accelerating Logs Insights queries. One per group. |
| `logStreams[]` | list | No | — | Names of log streams to pre-create (fixed-name agent configs, stream-scoped IAM). Most streams are runtime-created — leave empty unless something depends on the stream existing first. |
| `transformer` | object | No | — | Ingest-time processor pipeline (1–20 processors; the first must be a parser). One per group; STANDARD class only. |

**ForceNew warning:** `logGroupClass` triggers log group replacement when changed. Choose carefully at creation time.

**DELIVERY class note:** When `logGroupClass` is `DELIVERY`, `retentionInDays` must not be set (AWS manages retention for Delivery log groups).

**INFREQUENT_ACCESS note:** Metric filters, subscription filters, and transformers are not supported on INFREQUENT_ACCESS log groups — the spec rejects the combinations at validation time.

### Metric filter fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Filter name, unique within the group. |
| `pattern` | string | No | Filter pattern; empty matches ALL events. Supports terms, JSON matching, and column matching. |
| `applyOnTransformedLogs` | bool | No | Apply after the group's transformer (if configured). |
| `transformation.metricName` | string | Yes | Metric to publish (e.g. `ErrorCount`). |
| `transformation.metricNamespace` | string | Yes | Custom namespace (e.g. `MyApp/Errors`). |
| `transformation.metricValue` | string | Yes | Literal (`"1"`) or field reference (`"$.latencyMs"`). |
| `transformation.defaultValue` | double | No | Published when no events match; mutually exclusive with dimensions. |
| `transformation.dimensions` | map | No | Dimension name → field reference; maximum 3. |
| `transformation.unit` | string | No | CloudWatch StandardUnit; defaults to `None`. |

### Subscription filter fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Filter name, unique within the group. |
| `destinationArn` | StringValueOrRef | Yes | Kinesis stream, Firehose delivery stream, or Lambda function ARN. |
| `filterPattern` | string | No | Empty delivers ALL events. |
| `roleArn` | StringValueOrRef | No | IAM role CloudWatch Logs assumes for Kinesis/Firehose delivery (required for those destinations; not used for Lambda). |
| `distribution` | string | No | `ByLogStream` (default, ordered) or `Random` (throughput) — Kinesis streams only. |
| `emitSystemFields[]` | list | No | `@aws.account` / `@aws.region` / `@source.log` enrichment for centralized destinations. |
| `applyOnTransformedLogs` | bool (tri-state) | No | Match transformed events instead of raw ones. Unset keeps an existing filter's setting; only an explicit `false` switches it back. |

### Transformer fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `processors[]` | list | Yes | 1–20 pipeline steps, applied in order; each entry carries exactly ONE processor. |
| `processors[].parseJson` / `grok` / `csv` / `parseKeyValue` | object | — | Configurable parsers — one of these (or a vended parser) must be the FIRST processor. `parseJson`/`parseKeyValue`/`csv` up to 5 each; `grok` once. |
| `processors[].parseCloudfront` / `parsePostgres` / `parseRoute53` / `parseToOcsf` / `parseVpc` / `parseWaf` | object | — | Vended-log parsers — at most once each, and always first. |
| `processors[].addKeys` / `copyValue` / `moveKeys` / `renameKeys` / `deleteKeys` | object | — | Key surgery (add/copy/move/rename/delete); `addKeys` and `copyValue` once per transformer. |
| `processors[].splitString` / `substituteString` / `trimString` / `lowerCaseString` / `upperCaseString` | object | — | String operations on named fields. |
| `processors[].dateTimeConverter` / `typeConverter` / `listToMap` | object | — | Datetime normalization, type conversion (`integer`/`double`/`string`/`boolean`), and list-to-map reshaping. |

## Outputs

| Output | Description |
|--------|-------------|
| `log_group_arn` | Log group ARN (without the `:*` suffix). Primary reference for Step Functions, API Gateway, OpenSearch, and other services. |
| `log_group_name` | Log group name. Used by services that reference log groups by name (ElastiCache, ECS awslogs driver). |

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogGroup
metadata:
  name: app-logs
spec:
  region: us-west-2
  retentionInDays: 30
```

## Log-Derived Alerting Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogGroup
metadata:
  name: app-logs
spec:
  region: us-west-2
  retentionInDays: 30
  metricFilters:
    - name: error-count
      pattern: "ERROR"
      transformation:
        metricName: ErrorCount
        metricNamespace: MyApp/Errors
        metricValue: "1"
        defaultValue: 0
```

Then alarm on the published metric with an `AwsCloudwatchAlarm` on `MyApp/Errors` / `ErrorCount`.

## Production Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogGroup
metadata:
  name: prod-app-logs
spec:
  region: us-west-2
  retentionInDays: 90
  deletionProtectionEnabled: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: log-encryption-key
      fieldPath: status.outputs.key_arn
  subscriptionFilters:
    - name: to-analytics
      destinationArn:
        valueFrom:
          kind: AwsKinesisStream
          name: log-analytics
          fieldPath: status.outputs.stream_arn
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: cwl-to-kinesis
          fieldPath: status.outputs.role_arn
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

## What Is Deliberately Omitted

- **Log anomaly detector** — One detector trains over a LIST of log groups; multi-group scope makes it a standalone resource, never a single group's satellite.
- **Account-level storage tier policy** — A regional account singleton, not per-group configuration.
- **Delivery source/destination/delivery** — The vended-log delivery plane (`aws_cloudwatch_log_delivery_*`), a separate surface.
- **Cross-account destinations and destination policies** — Cross-account log aggregation plumbing; subscription filters compose to them by literal ARN today.
- **Query definitions and account policies** — Account-scoped operational tooling, not group shape.
- **`skip_destroy`** — Retaining a resource on destroy contradicts honest lifecycle management; protection needs are served by `deletionProtectionEnabled`.
- **`name_prefix`** — Planton derives the name from `metadata.name`.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
