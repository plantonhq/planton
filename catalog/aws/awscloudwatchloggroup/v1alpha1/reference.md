# AwsCloudwatchLogGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsCloudwatchLogGroupSpec defines the desired configuration for an AWS
CloudWatch Logs log group.

A CloudWatch Logs log group is a container for log streams that share the same
retention, monitoring, and access control settings. Log groups are the primary
destination for application logs, service logs, and operational data across
AWS services.

CloudWatch Log Groups are referenced by many AWS services:
- Step Functions (execution logging)
- API Gateway (access logging)
- OpenSearch (slow logs, application logs, audit logs)
- Lambda (function execution logs — auto-created, but can be pre-created for policy control)
- ECS (container logs via awslogs driver)
- Route 53 (public hosted zone query logging)
- WAF (web request logging — one of several destination options)
- ElastiCache (slow logs, engine logs — one of several destination options)

Beyond the container itself, the spec folds in the log-group-scoped
satellites that share the group's lifecycle:
- `metric_filters` — turn matching log events into CloudWatch metrics
  (the raw material for alarms on error counts, latencies parsed from logs, etc.).
- `subscription_filters` — stream matching events in real time to a Kinesis
  stream, Firehose delivery stream, or Lambda function.
- `data_protection_policy` — mask/audit sensitive data (PII) in log events.
- `field_index_policy` — index selected log fields to accelerate and cheapen
  Logs Insights queries.

For the common use case, only `retention_in_days` is needed. KMS encryption
and log group class are optional settings for compliance and cost optimization.

Credentials, region, and deployment workflow live outside this spec in IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogGroup
metadata:
  name: test-log-group
  org: test-org
  env: dev
  id: test-log-group-dev
  annotations:
    planton.dev/provisioner: pulumi
spec:
  region: us-west-2
  retentionInDays: 30
  metricFilters:
    - name: error-count
      pattern: "ERROR"
      applyOnTransformedLogs: true
      transformation:
        metricName: ErrorCount
        metricNamespace: TestApp/Errors
        metricValue: "1"
        defaultValue: 0
  subscriptionFilters:
    - name: central-delivery
      destinationArn:
        value: arn:aws:kinesis:us-west-2:123456789012:stream/central-logs
      roleArn:
        value: arn:aws:iam::123456789012:role/cwl-to-kinesis-delivery
      filterPattern: ""
      emitSystemFields:
        - "@aws.account"
        - "@aws.region"
        - "@source.log"
  logStreams:
    - agent-primary
  transformer:
    processors:
      - parseJson: {}
      - renameKeys:
          entries:
            - key: msg
              renameTo: message
      - typeConverter:
          entries:
            - key: statusCode
              type: integer
  fieldIndexPolicy:
    Fields:
      - requestId
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.retentionInDays` | `int32` |  | `30` |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.logGroupClass` | `string` |  |  |  |
| `spec.deletionProtectionEnabled` | `bool` |  |  |  |
| `spec.metricFilters` | `[]AwsCloudwatchLogGroupMetricFilter` |  |  |  |
| `spec.metricFilters[].name` | `string` | yes |  |  |
| `spec.metricFilters[].pattern` | `string` |  |  |  |
| `spec.metricFilters[].applyOnTransformedLogs` | `bool` |  |  |  |
| `spec.metricFilters[].transformation` | `AwsCloudwatchLogGroupMetricTransformation` | yes |  |  |
| `spec.metricFilters[].transformation.metricName` | `string` | yes |  |  |
| `spec.metricFilters[].transformation.metricNamespace` | `string` | yes |  |  |
| `spec.metricFilters[].transformation.metricValue` | `string` | yes |  |  |
| `spec.metricFilters[].transformation.defaultValue` | `double` |  |  |  |
| `spec.metricFilters[].transformation.dimensions` | `map<string, string>` |  |  |  |
| `spec.metricFilters[].transformation.unit` | `string` |  |  |  |
| `spec.subscriptionFilters` | `[]AwsCloudwatchLogGroupSubscriptionFilter` |  |  |  |
| `spec.subscriptionFilters[].name` | `string` | yes |  |  |
| `spec.subscriptionFilters[].destinationArn` | `string \| valueFrom` | yes |  |  |
| `spec.subscriptionFilters[].filterPattern` | `string` |  |  |  |
| `spec.subscriptionFilters[].roleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.subscriptionFilters[].distribution` | `string` |  |  |  |
| `spec.subscriptionFilters[].emitSystemFields` | `[]string` |  |  |  |
| `spec.subscriptionFilters[].applyOnTransformedLogs` | `bool` |  |  |  |
| `spec.dataProtectionPolicy` | `object` |  |  |  |
| `spec.fieldIndexPolicy` | `object` |  |  |  |
| `spec.logStreams` | `[]string` |  |  |  |
| `spec.transformer` | `AwsCloudwatchLogGroupTransformer` |  |  |  |
| `spec.transformer.processors` | `[]AwsCloudwatchLogGroupTransformerProcessor` | yes |  |  |
| `spec.transformer.processors[].addKeys` | `AwsCloudwatchLogGroupTransformerAddKeys` |  |  |  |
| `spec.transformer.processors[].addKeys.entries` | `[]AwsCloudwatchLogGroupTransformerAddKeysEntry` | yes |  |  |
| `spec.transformer.processors[].addKeys.entries[].key` | `string` | yes |  |  |
| `spec.transformer.processors[].addKeys.entries[].value` | `string` | yes |  |  |
| `spec.transformer.processors[].addKeys.entries[].overwriteIfExists` | `bool` |  |  |  |
| `spec.transformer.processors[].copyValue` | `AwsCloudwatchLogGroupTransformerCopyValue` |  |  |  |
| `spec.transformer.processors[].copyValue.entries` | `[]AwsCloudwatchLogGroupTransformerCopyValueEntry` | yes |  |  |
| `spec.transformer.processors[].copyValue.entries[].source` | `string` | yes |  |  |
| `spec.transformer.processors[].copyValue.entries[].target` | `string` | yes |  |  |
| `spec.transformer.processors[].copyValue.entries[].overwriteIfExists` | `bool` |  |  |  |
| `spec.transformer.processors[].csv` | `AwsCloudwatchLogGroupTransformerCsv` |  |  |  |
| `spec.transformer.processors[].csv.columns` | `[]string` |  |  |  |
| `spec.transformer.processors[].csv.delimiter` | `string` |  |  |  |
| `spec.transformer.processors[].csv.quoteCharacter` | `string` |  |  |  |
| `spec.transformer.processors[].csv.source` | `string` |  |  |  |
| `spec.transformer.processors[].dateTimeConverter` | `AwsCloudwatchLogGroupTransformerDateTimeConverter` |  |  |  |
| `spec.transformer.processors[].dateTimeConverter.source` | `string` | yes |  |  |
| `spec.transformer.processors[].dateTimeConverter.target` | `string` | yes |  |  |
| `spec.transformer.processors[].dateTimeConverter.matchPatterns` | `[]string` | yes |  |  |
| `spec.transformer.processors[].dateTimeConverter.locale` | `string` |  |  |  |
| `spec.transformer.processors[].dateTimeConverter.sourceTimezone` | `string` |  |  |  |
| `spec.transformer.processors[].dateTimeConverter.targetFormat` | `string` |  |  |  |
| `spec.transformer.processors[].dateTimeConverter.targetTimezone` | `string` |  |  |  |
| `spec.transformer.processors[].deleteKeys` | `AwsCloudwatchLogGroupTransformerDeleteKeys` |  |  |  |
| `spec.transformer.processors[].deleteKeys.withKeys` | `[]string` | yes |  |  |
| `spec.transformer.processors[].grok` | `AwsCloudwatchLogGroupTransformerGrok` |  |  |  |
| `spec.transformer.processors[].grok.match` | `string` | yes |  |  |
| `spec.transformer.processors[].grok.source` | `string` |  |  |  |
| `spec.transformer.processors[].listToMap` | `AwsCloudwatchLogGroupTransformerListToMap` |  |  |  |
| `spec.transformer.processors[].listToMap.source` | `string` | yes |  |  |
| `spec.transformer.processors[].listToMap.key` | `string` | yes |  |  |
| `spec.transformer.processors[].listToMap.valueKey` | `string` |  |  |  |
| `spec.transformer.processors[].listToMap.target` | `string` |  |  |  |
| `spec.transformer.processors[].listToMap.flatten` | `bool` |  |  |  |
| `spec.transformer.processors[].listToMap.flattenedElement` | `string` |  |  |  |
| `spec.transformer.processors[].lowerCaseString` | `AwsCloudwatchLogGroupTransformerWithKeys` |  |  |  |
| `spec.transformer.processors[].lowerCaseString.withKeys` | `[]string` | yes |  |  |
| `spec.transformer.processors[].moveKeys` | `AwsCloudwatchLogGroupTransformerMoveKeys` |  |  |  |
| `spec.transformer.processors[].moveKeys.entries` | `[]AwsCloudwatchLogGroupTransformerMoveKeysEntry` | yes |  |  |
| `spec.transformer.processors[].moveKeys.entries[].source` | `string` | yes |  |  |
| `spec.transformer.processors[].moveKeys.entries[].target` | `string` | yes |  |  |
| `spec.transformer.processors[].moveKeys.entries[].overwriteIfExists` | `bool` |  |  |  |
| `spec.transformer.processors[].parseCloudfront` | `AwsCloudwatchLogGroupTransformerVendedLogParser` |  |  |  |
| `spec.transformer.processors[].parseCloudfront.source` | `string` |  |  |  |
| `spec.transformer.processors[].parseJson` | `AwsCloudwatchLogGroupTransformerParseJson` |  |  |  |
| `spec.transformer.processors[].parseJson.source` | `string` |  |  |  |
| `spec.transformer.processors[].parseJson.destination` | `string` |  |  |  |
| `spec.transformer.processors[].parseKeyValue` | `AwsCloudwatchLogGroupTransformerParseKeyValue` |  |  |  |
| `spec.transformer.processors[].parseKeyValue.source` | `string` |  |  |  |
| `spec.transformer.processors[].parseKeyValue.destination` | `string` |  |  |  |
| `spec.transformer.processors[].parseKeyValue.fieldDelimiter` | `string` |  |  |  |
| `spec.transformer.processors[].parseKeyValue.keyValueDelimiter` | `string` |  |  |  |
| `spec.transformer.processors[].parseKeyValue.keyPrefix` | `string` |  |  |  |
| `spec.transformer.processors[].parseKeyValue.nonMatchValue` | `string` |  |  |  |
| `spec.transformer.processors[].parseKeyValue.overwriteIfExists` | `bool` |  |  |  |
| `spec.transformer.processors[].parsePostgres` | `AwsCloudwatchLogGroupTransformerVendedLogParser` |  |  |  |
| `spec.transformer.processors[].parsePostgres.source` | `string` |  |  |  |
| `spec.transformer.processors[].parseRoute53` | `AwsCloudwatchLogGroupTransformerVendedLogParser` |  |  |  |
| `spec.transformer.processors[].parseRoute53.source` | `string` |  |  |  |
| `spec.transformer.processors[].parseToOcsf` | `AwsCloudwatchLogGroupTransformerParseToOcsf` |  |  |  |
| `spec.transformer.processors[].parseToOcsf.eventSource` | `string` | yes |  |  |
| `spec.transformer.processors[].parseToOcsf.ocsfVersion` | `string` | yes |  |  |
| `spec.transformer.processors[].parseToOcsf.source` | `string` |  |  |  |
| `spec.transformer.processors[].parseVpc` | `AwsCloudwatchLogGroupTransformerVendedLogParser` |  |  |  |
| `spec.transformer.processors[].parseVpc.source` | `string` |  |  |  |
| `spec.transformer.processors[].parseWaf` | `AwsCloudwatchLogGroupTransformerVendedLogParser` |  |  |  |
| `spec.transformer.processors[].parseWaf.source` | `string` |  |  |  |
| `spec.transformer.processors[].renameKeys` | `AwsCloudwatchLogGroupTransformerRenameKeys` |  |  |  |
| `spec.transformer.processors[].renameKeys.entries` | `[]AwsCloudwatchLogGroupTransformerRenameKeysEntry` | yes |  |  |
| `spec.transformer.processors[].renameKeys.entries[].key` | `string` | yes |  |  |
| `spec.transformer.processors[].renameKeys.entries[].renameTo` | `string` | yes |  |  |
| `spec.transformer.processors[].renameKeys.entries[].overwriteIfExists` | `bool` |  |  |  |
| `spec.transformer.processors[].splitString` | `AwsCloudwatchLogGroupTransformerSplitString` |  |  |  |
| `spec.transformer.processors[].splitString.entries` | `[]AwsCloudwatchLogGroupTransformerSplitStringEntry` | yes |  |  |
| `spec.transformer.processors[].splitString.entries[].source` | `string` | yes |  |  |
| `spec.transformer.processors[].splitString.entries[].delimiter` | `string` | yes |  |  |
| `spec.transformer.processors[].substituteString` | `AwsCloudwatchLogGroupTransformerSubstituteString` |  |  |  |
| `spec.transformer.processors[].substituteString.entries` | `[]AwsCloudwatchLogGroupTransformerSubstituteStringEntry` | yes |  |  |
| `spec.transformer.processors[].substituteString.entries[].source` | `string` | yes |  |  |
| `spec.transformer.processors[].substituteString.entries[].from` | `string` | yes |  |  |
| `spec.transformer.processors[].substituteString.entries[].to` | `string` | yes |  |  |
| `spec.transformer.processors[].trimString` | `AwsCloudwatchLogGroupTransformerWithKeys` |  |  |  |
| `spec.transformer.processors[].trimString.withKeys` | `[]string` | yes |  |  |
| `spec.transformer.processors[].typeConverter` | `AwsCloudwatchLogGroupTransformerTypeConverter` |  |  |  |
| `spec.transformer.processors[].typeConverter.entries` | `[]AwsCloudwatchLogGroupTransformerTypeConverterEntry` | yes |  |  |
| `spec.transformer.processors[].typeConverter.entries[].key` | `string` | yes |  |  |
| `spec.transformer.processors[].typeConverter.entries[].type` | `string` | yes |  |  |
| `spec.transformer.processors[].upperCaseString` | `AwsCloudwatchLogGroupTransformerWithKeys` |  |  |  |
| `spec.transformer.processors[].upperCaseString.withKeys` | `[]string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.retentionInDays

`int32`

Number of days to retain log events. After this period, log events are
automatically deleted. Set to 0 (the default) to retain log events
indefinitely — they will never expire.

AWS only accepts specific values for this field. Any value not in the
allowed set will be rejected by the AWS API.

Allowed values: 0 (never expire), 1, 3, 5, 7, 14, 30, 60, 90, 120, 150,
180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653.

Recommended: Set an explicit retention for cost control. Indefinite retention
(0) accumulates storage costs over time.

- default: `30`

### spec.kmsKeyId

`string | valueFrom`

ARN of the KMS key to use for encrypting log data at rest. When omitted,
CloudWatch Logs uses its default server-side encryption (SSE-CWL).

Customer-managed KMS keys provide:
- Key rotation control
- Cross-account access via key policy
- CloudTrail audit trail of log data access
- Compliance with regulations requiring customer-controlled encryption keys

The KMS key must be in the same region as the log group, and its key policy
must allow the CloudWatch Logs service principal
(`logs.<region>.amazonaws.com`) to use the key. Associating or
disassociating the key updates the log group in place (no replacement).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.logGroupClass

`string`

The class of the log group, which determines pricing, feature availability,
and data durability characteristics.

Valid values:
- "STANDARD" — Full feature set including metric filters, subscription
  filters, Logs Insights, and Contributor Insights. Default when omitted.
- "INFREQUENT_ACCESS" — Reduced cost (~50% cheaper storage) with a subset
  of features. Supports Logs Insights and managed ingestion to S3, but does
  NOT support metric filters, subscription filters, or Contributor Insights.
  Best for high-volume logs accessed infrequently (VPC flow logs, CDN access
  logs, compliance archives).
- "DELIVERY" — Purpose-built for AWS service log delivery (VPC Flow Logs,
  CloudTrail, Route53 Resolver). Lowest cost. Retention is managed by AWS
  and retention_in_days must not be set.

This field is ForceNew: changing it requires replacing the log group.

### spec.deletionProtectionEnabled

`bool` · optional (explicit presence)

When true, the log group is protected from deletion. Any attempt to delete
the log group (including via IaC destroy) will fail until this flag is set
to false. Useful for protecting production log groups from accidental
deletion.

Three states, and the difference matters for turning protection OFF: the
provider attribute is Optional+Computed, so an OMITTED value keeps whatever
the log group already has — only an EXPLICIT false disables protection.
- unset — new groups get AWS's default (unprotected); existing groups keep
  their current protection state.
- true  — protection enabled.
- false — protection explicitly disabled (the only way back off).

### spec.metricFilters

`[]AwsCloudwatchLogGroupMetricFilter`

Metric filters that extract CloudWatch metrics from log events flowing
through this group. Each filter matches events against a pattern and
publishes a metric value for every match — the standard way to alarm on
"ERROR" counts, parsed latencies, or any signal that lives only in logs.

Filters are keyed by `name` within the group; names must be unique.
Not supported on INFREQUENT_ACCESS log groups (AWS rejects the call).

### spec.metricFilters[].name

`string` · required

Name of the metric filter, unique within the log group. Changing the
name replaces the filter. Must not contain ':' or '*' (the
PutMetricFilter contract).

- rule: {"required":true,"string":{"maxLen":"512","pattern":"^[^:*]+$"}}

### spec.metricFilters[].pattern

`string`

Filter pattern that selects which log events produce metric values.
An empty pattern matches ALL log events.

Pattern syntax supports plain terms ("ERROR"), JSON field matching
({ $.statusCode = 500 }), and space-delimited column matching
([ip, user, ts, request, status=5*, size]).
Maximum 1024 characters.

- rule: {"string":{"maxLen":"1024"}}

### spec.metricFilters[].applyOnTransformedLogs

`bool` · optional (explicit presence)

When true, this filter matches against log events AFTER they pass through
the log group's transformer (or an account-level transformer managed
outside this resource). When false or unset, the filter matches raw
ingested events.

The provider attribute is Optional+Computed, so an omitted value keeps the
filter's existing setting — only an explicit false switches an existing
filter back to matching raw events.

### spec.metricFilters[].transformation

`AwsCloudwatchLogGroupMetricTransformation` · required

The metric to publish for each matching log event. Required — a filter
without a transformation does nothing.

- rule: {"required":true}
- rule: default_value cannot be set together with dimensions — AWS does not support a default value on dimensional metric filters
- rule: a metric filter supports at most 3 dimensions
- rule: unit must be a valid CloudWatch unit (e.g. Seconds, Milliseconds, Bytes, Percent, Count, Bytes/Second, Count/Second, None) when set

### spec.metricFilters[].transformation.metricName

`string` · required

Name of the CloudWatch metric to publish (e.g. "ErrorCount").
Must not contain ':', '*', or '$' (the MetricTransformation contract).

- rule: {"required":true,"string":{"maxLen":"255","pattern":"^[^:*$]*$"}}

### spec.metricFilters[].transformation.metricNamespace

`string` · required

Namespace for the metric (e.g. "MyApp/Errors"). Custom namespaces keep
log-derived metrics separate from AWS service namespaces.
Must not contain ':', '*', or '$' (the MetricTransformation contract).

- rule: {"required":true,"string":{"maxLen":"255","pattern":"^[^:*$]*$"}}

### spec.metricFilters[].transformation.metricValue

`string` · required

Value to publish for each matching event. Either a literal number
("1" to count occurrences) or a field reference that extracts a numeric
value from the matched event (e.g. "$.latencyMs" for JSON patterns, or
"$size" for a named column in space-delimited patterns).
Maximum 100 characters.

- rule: {"required":true,"string":{"maxLen":"100"}}

### spec.metricFilters[].transformation.defaultValue

`double` · optional (explicit presence)

Value to publish for periods when NO log events match the pattern.
Typically "0" so count metrics report zero instead of missing data —
which lets alarms use standard missing-data handling. When unset, no
value is published for non-matching periods.

AWS does not allow a default value on filters that publish dimensions.

### spec.metricFilters[].transformation.dimensions

`map<string, string>`

Dimensions to publish with the metric, mapping dimension names to field
references in the matched event (e.g. {"ErrorCode": "$.errorCode"}).
Maximum 3 dimensions per AWS limit.

Every unique dimension-value combination creates a distinct custom
metric (billed separately) — keep dimension cardinality low.

### spec.metricFilters[].transformation.unit

`string`

Unit for the metric (e.g. "Count", "Milliseconds", "Bytes"). Defaults to
"None" when unset. Must be a valid CloudWatch StandardUnit.

### spec.subscriptionFilters

`[]AwsCloudwatchLogGroupSubscriptionFilter`

Real-time subscription filters that stream matching log events to a
destination: a Kinesis data stream, a Kinesis Data Firehose delivery
stream, or a Lambda function. Use these to fan logs out to analytics
pipelines, SIEM tooling, or custom processing.

AWS allows at most TWO subscription filters per log group. Names must be
unique within the group. Not supported on INFREQUENT_ACCESS log groups.

- rule: distribution must be 'ByLogStream' or 'Random' when set
- rule: emit_system_fields entries must be '@aws.account', '@aws.region', or '@source.log'

### spec.subscriptionFilters[].name

`string` · required

Name of the subscription filter, unique within the log group. Changing
the name replaces the filter. Must not contain ':' or '*' (the
PutSubscriptionFilter contract; the provider validates length only —
the character rule is AWS's own).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"512","pattern":"^[^:*]+$"}}

### spec.subscriptionFilters[].destinationArn

`string | valueFrom` · required

ARN of the destination that receives matching log events. Supported
destinations:
- Kinesis data stream (same-account) — reference an AwsKinesisStream
- Kinesis Data Firehose delivery stream — reference an AwsKinesisFirehose
- Lambda function — reference an AwsLambda
- Cross-account CloudWatch Logs destination (by literal ARN)

No default kind is set because all destination types are equally common —
reference the specific resource kind's ARN output, or provide a literal ARN.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.subscriptionFilters[].filterPattern

`string`

Filter pattern that selects which log events are delivered. An empty
pattern delivers ALL log events. Same syntax as metric filter patterns.
Maximum 1024 characters.

- rule: {"string":{"maxLen":"1024"}}

### spec.subscriptionFilters[].roleArn

`string | valueFrom`

ARN of the IAM role that grants CloudWatch Logs permission to put records
to the destination. REQUIRED for Kinesis stream and Firehose destinations
(the role must trust `logs.amazonaws.com` and allow `kinesis:PutRecord` /
`firehose:PutRecord` on the destination). NOT used for Lambda destinations —
those authorize via a Lambda resource-based permission instead.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.subscriptionFilters[].distribution

`string`

How log data is distributed to Kinesis stream destinations:
- "ByLogStream" — (default) events from the same log stream go to the
  same shard, preserving per-stream ordering.
- "Random" — events are spread across shards for maximum throughput,
  at the cost of ordering.

Only meaningful for Kinesis stream destinations.

### spec.subscriptionFilters[].emitSystemFields

`[]string`

System fields to include with each delivered log event. Supported values:
"@aws.account" (the source account ID), "@aws.region" (the source
region), and "@source.log" (the source log group and stream). Useful when
a central destination aggregates logs from many accounts/regions/groups.

### spec.subscriptionFilters[].applyOnTransformedLogs

`bool` · optional (explicit presence)

When true, this filter matches and delivers log events AFTER they pass
through the log group's transformer (or an account-level transformer
managed outside this resource). When false or unset, the filter operates
on raw ingested events.

The provider attribute is Optional+Computed, so an omitted value keeps the
filter's existing setting — only an explicit false switches an existing
filter back to raw events.

### spec.dataProtectionPolicy

`object`

CloudWatch Logs data protection policy for this log group, as a JSON
policy document. Data protection audits and masks sensitive data (PII
such as email addresses, credit card numbers, or custom identifiers) in
log events as they are ingested.

The document follows the CloudWatch Logs data protection policy schema:
a `Name`, `Version: "2021-06-01"`, and a `Statement` list carrying one
"Audit" operation statement and one "Deidentify" (masking) statement over
the same data identifiers. Audit findings destinations (CloudWatch Logs,
S3, Firehose) are optional — masking works without them.

One policy per log group. Masked data is visible only to principals with
the `logs:Unmask` permission.

### spec.fieldIndexPolicy

`object`

CloudWatch Logs field index policy for this log group, as a JSON policy
document with a `Fields` array of log-event field names to index
(e.g. {"Fields": ["requestId", "userId"]}).

Field indexes make Logs Insights queries that filter on the indexed
fields faster and cheaper: matching scans skip events where the indexed
value cannot match. Index up to 20 fields per policy. One policy per log
group; an account-level policy (managed outside this resource) can also
apply — the two merge at query time.

### spec.logStreams

`[]string`

Names of log streams to pre-create inside this log group. Most log streams
are created at runtime by the writing agent or AWS service and should NOT
be listed here — pre-create a stream only when something depends on the
stream existing before the first write (an agent configured with a fixed
stream name, IAM policies scoped to specific stream ARNs, or tooling that
writes with `sequenceToken` semantics).

Stream names must be 1-512 characters and must not contain ':' or '*'
(the CreateLogStream contract). Streams are deleted with the group.

### spec.transformer

`AwsCloudwatchLogGroupTransformer`

Log transformer for this log group: an ordered pipeline of 1-20 processors
that parses and reshapes every log event at ingestion time. The transformed
form is what Logs Insights queries, metric filters, and subscription
filters see (the latter two only when their `apply_on_transformed_logs` is
set). One transformer per log group; supported ONLY on STANDARD class log
groups (the PutTransformer contract).

The first processor must be a parser (parse_json, grok, csv,
parse_key_value, or one of the vended-log parsers). An account-level
transformer (managed outside this resource) can also exist; when both
apply, the log-group-level transformer wins and the account-level one is
ignored.

- rule: the first processor must be a parser: parse_json, grok, csv, parse_key_value, parse_cloudfront, parse_postgres, parse_route53, parse_to_ocsf, parse_vpc, or parse_waf
- rule: add_keys, copy_value, and grok may each appear at most once per transformer
- rule: parse_json, parse_key_value, and csv may each appear at most 5 times per transformer
- rule: each vended-log parser (parse_cloudfront, parse_postgres, parse_route53, parse_to_ocsf, parse_vpc, parse_waf) may appear at most once per transformer
- rule: a vended-log parser (parse_cloudfront, parse_postgres, parse_route53, parse_to_ocsf, parse_vpc, parse_waf) must be the first processor in the pipeline

### spec.transformer.processors

`[]AwsCloudwatchLogGroupTransformerProcessor` · required

The processor pipeline, applied in order. Each entry configures exactly
one processor.

- rule: {"repeated":{"minItems":"1","maxItems":"20"}}
- rule: exactly one processor type must be set per pipeline entry

### spec.transformer.processors[].addKeys

`AwsCloudwatchLogGroupTransformerAddKeys`

Adds new key-value pairs to the log event. Single-use per transformer.

### spec.transformer.processors[].addKeys.entries

`[]AwsCloudwatchLogGroupTransformerAddKeysEntry` · required

Keys to add. 1-5 entries.

- rule: {"repeated":{"minItems":"1","maxItems":"5"}}

### spec.transformer.processors[].addKeys.entries[].key

`string` · required

Key of the new entry. 1-128 characters. Use dot notation for nested
fields (e.g. "metadata.environment").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].addKeys.entries[].value

`string` · required

Value of the new entry. 1-256 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"256"}}

### spec.transformer.processors[].addKeys.entries[].overwriteIfExists

`bool`

When true, overwrites the value if the key already exists in the log
event. Defaults to false (existing values win).

### spec.transformer.processors[].copyValue

`AwsCloudwatchLogGroupTransformerCopyValue`

Copies values from existing keys to new keys. Single-use per transformer.

### spec.transformer.processors[].copyValue.entries

`[]AwsCloudwatchLogGroupTransformerCopyValueEntry` · required

Values to copy. 1-5 entries.

- rule: {"repeated":{"minItems":"1","maxItems":"5"}}

### spec.transformer.processors[].copyValue.entries[].source

`string` · required

Key to copy from. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].copyValue.entries[].target

`string` · required

Key to copy the value to. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].copyValue.entries[].overwriteIfExists

`bool`

When true, overwrites the value if the target key already exists.
Defaults to false.

### spec.transformer.processors[].csv

`AwsCloudwatchLogGroupTransformerCsv`

Parses comma-separated values into columns. Counts as a parser (may be
the pipeline's first processor). Up to 5 per transformer.

- rule: each csv column name must be 1-128 characters

### spec.transformer.processors[].csv.columns

`[]string`

Names for the parsed columns. When omitted, default names
(column_1, column_2, ...) are used. Up to 100 names, each 1-128
characters.

- rule: {"repeated":{"maxItems":"100"}}

### spec.transformer.processors[].csv.delimiter

`string`

Character separating columns in the source value. Defaults to ",".
1-2 characters.

- rule: {"string":{"maxLen":"2"}}

### spec.transformer.processors[].csv.quoteCharacter

`string`

Character used as a text qualifier for a single column of data.
Defaults to '"'. Exactly 1 character.

- rule: {"string":{"maxLen":"1"}}

### spec.transformer.processors[].csv.source

`string`

Path to the field to parse. When omitted, the whole @message is
processed. 1-128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.transformer.processors[].dateTimeConverter

`AwsCloudwatchLogGroupTransformerDateTimeConverter`

Converts a datetime string into a target format.

- rule: each match_patterns entry must be non-empty

### spec.transformer.processors[].dateTimeConverter.source

`string` · required

Key holding the datetime string to convert. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].dateTimeConverter.target

`string` · required

Key to store the converted result in. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].dateTimeConverter.matchPatterns

`[]string` · required

Patterns to match against the source value (Java DateTimeFormatter
syntax, e.g. "dd/MMM/yyyy:HH:mm:ss" — or "epoch" for epoch timestamps).
1-5 patterns.

- rule: {"repeated":{"minItems":"1","maxItems":"5"}}

### spec.transformer.processors[].dateTimeConverter.locale

`string`

Locale of the source field (e.g. "en-US"). Defaults to locale.ROOT.

### spec.transformer.processors[].dateTimeConverter.sourceTimezone

`string`

Time zone of the source field (e.g. "America/Los_Angeles"). Defaults
to UTC.

### spec.transformer.processors[].dateTimeConverter.targetFormat

`string`

Datetime format for the converted value in the target field. Defaults to
"yyyy-MM-dd'T'HH:mm:ss.SSS'Z". 1-64 characters.

- rule: {"string":{"maxLen":"64"}}

### spec.transformer.processors[].dateTimeConverter.targetTimezone

`string`

Time zone of the target field. Defaults to UTC.

### spec.transformer.processors[].deleteKeys

`AwsCloudwatchLogGroupTransformerDeleteKeys`

Deletes keys from the log event.

- rule: each with_keys entry must be non-empty

### spec.transformer.processors[].deleteKeys.withKeys

`[]string` · required

Keys to delete. 1-5 keys, each non-empty.

- rule: {"repeated":{"minItems":"1","maxItems":"5"}}

### spec.transformer.processors[].grok

`AwsCloudwatchLogGroupTransformerGrok`

Parses unstructured text with a grok pattern. Counts as a parser.
Single-use per transformer.

### spec.transformer.processors[].grok.match

`string` · required

Grok pattern to match against the log event (e.g.
"%{COMMONAPACHELOG}" or a composition of named grok patterns).
1-512 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"512"}}

### spec.transformer.processors[].grok.source

`string`

Path to the field to parse. When omitted, the whole @message is
processed. 1-128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.transformer.processors[].listToMap

`AwsCloudwatchLogGroupTransformerListToMap`

Converts a list of key-value objects into a map.

- rule: flattened_element must be 'first' or 'last' when set
- rule: flattened_element is required when flatten is true

### spec.transformer.processors[].listToMap.source

`string` · required

Key of the field holding the list of objects to convert.
1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].listToMap.key

`string` · required

Field within each source object whose value becomes a key in the
generated map. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].listToMap.valueKey

`string`

Field within each source object whose value is placed into the map's
values. When omitted, the whole source object becomes the value.
1-128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.transformer.processors[].listToMap.target

`string`

Key of the field that will hold the generated map. When omitted, the
map is placed under the root node. 1-128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.transformer.processors[].listToMap.flatten

`bool`

When true, lists of values in the generated map are flattened into
single items using flattened_element.

### spec.transformer.processors[].listToMap.flattenedElement

`string`

Which element to keep when flattening: "first" or "last". Required when
flatten is true.

### spec.transformer.processors[].lowerCaseString

`AwsCloudwatchLogGroupTransformerWithKeys`

Converts the values of the given keys to lowercase.

- rule: each with_keys entry must be non-empty

### spec.transformer.processors[].lowerCaseString.withKeys

`[]string` · required

Keys whose values the operation applies to. 1-10 keys, each non-empty.

- rule: {"repeated":{"minItems":"1","maxItems":"10"}}

### spec.transformer.processors[].moveKeys

`AwsCloudwatchLogGroupTransformerMoveKeys`

Moves values from one key to another.

### spec.transformer.processors[].moveKeys.entries

`[]AwsCloudwatchLogGroupTransformerMoveKeysEntry` · required

Keys to move. 1-5 entries.

- rule: {"repeated":{"minItems":"1","maxItems":"5"}}

### spec.transformer.processors[].moveKeys.entries[].source

`string` · required

Key to move. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].moveKeys.entries[].target

`string` · required

Key to move the value to. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].moveKeys.entries[].overwriteIfExists

`bool`

When true, overwrites the value if the target key already exists.
Defaults to false.

### spec.transformer.processors[].parseCloudfront

`AwsCloudwatchLogGroupTransformerVendedLogParser`

Parses CloudFront vended access logs into JSON fields. Must be the first
processor when present; single-use per transformer.

- rule: source must be '@message' when set — vended-log parsers can only parse the raw log message

### spec.transformer.processors[].parseCloudfront.source

`string`

Source field to parse. The only allowed value is "@message"; when
omitted, the whole log message is processed.

### spec.transformer.processors[].parseJson

`AwsCloudwatchLogGroupTransformerParseJson`

Parses JSON-format log events. Counts as a parser. Up to 5 per
transformer.

### spec.transformer.processors[].parseJson.source

`string`

Path to the field to parse. Defaults to "@message". 1-128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.transformer.processors[].parseJson.destination

`string`

Location to put the parsed key-value pairs. When omitted, they are
placed under the root node. 1-128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.transformer.processors[].parseKeyValue

`AwsCloudwatchLogGroupTransformerParseKeyValue`

Parses a field into key-value pairs. Counts as a parser. Up to 5 per
transformer.

### spec.transformer.processors[].parseKeyValue.source

`string`

Path to the field to parse. Defaults to "@message". 1-128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.transformer.processors[].parseKeyValue.destination

`string`

Destination field for the extracted key-value pairs. 1-128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.transformer.processors[].parseKeyValue.fieldDelimiter

`string`

Delimiter between key-value pairs in the source (e.g. ";"). Defaults
to "&". 1-128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.transformer.processors[].parseKeyValue.keyValueDelimiter

`string`

Delimiter between the key and the value within a pair (e.g. ":").
Defaults to "=". 1-128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.transformer.processors[].parseKeyValue.keyPrefix

`string`

Prefix added to all transformed keys. 1-128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.transformer.processors[].parseKeyValue.nonMatchValue

`string`

Value inserted when a pair cannot be split successfully.
1-128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.transformer.processors[].parseKeyValue.overwriteIfExists

`bool`

When true, overwrites the value if the destination key already exists.
Defaults to false.

### spec.transformer.processors[].parsePostgres

`AwsCloudwatchLogGroupTransformerVendedLogParser`

Parses RDS for PostgreSQL vended logs into JSON fields. Must be the first
processor when present; single-use per transformer.

- rule: source must be '@message' when set — vended-log parsers can only parse the raw log message

### spec.transformer.processors[].parsePostgres.source

`string`

Source field to parse. The only allowed value is "@message"; when
omitted, the whole log message is processed.

### spec.transformer.processors[].parseRoute53

`AwsCloudwatchLogGroupTransformerVendedLogParser`

Parses Route 53 vended logs into JSON fields. Must be the first processor
when present; single-use per transformer.

- rule: source must be '@message' when set — vended-log parsers can only parse the raw log message

### spec.transformer.processors[].parseRoute53.source

`string`

Source field to parse. The only allowed value is "@message"; when
omitted, the whole log message is processed.

### spec.transformer.processors[].parseToOcsf

`AwsCloudwatchLogGroupTransformerParseToOcsf`

Converts log events into Open Cybersecurity Schema Framework (OCSF)
events. Must be the first processor when present; single-use per
transformer.

- rule: event_source must be one of: CloudTrail, Route53Resolver, VPCFlow, EKSAudit, AWSWAF
- rule: ocsf_version must be 'V1.1' or 'V1.5'
- rule: source must be '@message' when set — OCSF conversion can only parse the raw log message

### spec.transformer.processors[].parseToOcsf.eventSource

`string` · required

Service or process producing the log events. Valid values:
"CloudTrail", "Route53Resolver", "VPCFlow", "EKSAudit", "AWSWAF".

- rule: {"required":true}

### spec.transformer.processors[].parseToOcsf.ocsfVersion

`string` · required

OCSF schema version for the transformed events. Valid values: "V1.1",
"V1.5".

- rule: {"required":true}

### spec.transformer.processors[].parseToOcsf.source

`string`

Source field to parse. The only allowed value is "@message"; when
omitted, the whole log message is processed.

### spec.transformer.processors[].parseVpc

`AwsCloudwatchLogGroupTransformerVendedLogParser`

Parses Amazon VPC flow logs into JSON fields. Must be the first processor
when present; single-use per transformer.

- rule: source must be '@message' when set — vended-log parsers can only parse the raw log message

### spec.transformer.processors[].parseVpc.source

`string`

Source field to parse. The only allowed value is "@message"; when
omitted, the whole log message is processed.

### spec.transformer.processors[].parseWaf

`AwsCloudwatchLogGroupTransformerVendedLogParser`

Parses AWS WAF vended logs into JSON fields. Must be the first processor
when present; single-use per transformer.

- rule: source must be '@message' when set — vended-log parsers can only parse the raw log message

### spec.transformer.processors[].parseWaf.source

`string`

Source field to parse. The only allowed value is "@message"; when
omitted, the whole log message is processed.

### spec.transformer.processors[].renameKeys

`AwsCloudwatchLogGroupTransformerRenameKeys`

Renames keys in the log event.

### spec.transformer.processors[].renameKeys.entries

`[]AwsCloudwatchLogGroupTransformerRenameKeysEntry` · required

Keys to rename. 1-5 entries.

- rule: {"repeated":{"minItems":"1","maxItems":"5"}}

### spec.transformer.processors[].renameKeys.entries[].key

`string` · required

Key to rename. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].renameKeys.entries[].renameTo

`string` · required

New name for the key. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].renameKeys.entries[].overwriteIfExists

`bool`

When true, overwrites the value if the new key name already exists.
Defaults to false.

### spec.transformer.processors[].splitString

`AwsCloudwatchLogGroupTransformerSplitString`

Splits field values into arrays using a delimiter.

### spec.transformer.processors[].splitString.entries

`[]AwsCloudwatchLogGroupTransformerSplitStringEntry` · required

Fields to split. 1-10 entries.

- rule: {"repeated":{"minItems":"1","maxItems":"10"}}

### spec.transformer.processors[].splitString.entries[].source

`string` · required

Key of the field to split. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].splitString.entries[].delimiter

`string` · required

Separator characters to split the string on. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].substituteString

`AwsCloudwatchLogGroupTransformerSubstituteString`

Replaces regex matches in field values with a replacement string.

### spec.transformer.processors[].substituteString.entries

`[]AwsCloudwatchLogGroupTransformerSubstituteStringEntry` · required

Fields to substitute. 1-10 entries.

- rule: {"repeated":{"minItems":"1","maxItems":"10"}}

### spec.transformer.processors[].substituteString.entries[].source

`string` · required

Key of the field to modify. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].substituteString.entries[].from

`string` · required

Regular expression whose matches are replaced. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].substituteString.entries[].to

`string` · required

Replacement string for each match. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].trimString

`AwsCloudwatchLogGroupTransformerWithKeys`

Trims leading and trailing whitespace from the given keys' values.

- rule: each with_keys entry must be non-empty

### spec.transformer.processors[].trimString.withKeys

`[]string` · required

Keys whose values the operation applies to. 1-10 keys, each non-empty.

- rule: {"repeated":{"minItems":"1","maxItems":"10"}}

### spec.transformer.processors[].typeConverter

`AwsCloudwatchLogGroupTransformerTypeConverter`

Converts field values to a different data type.

### spec.transformer.processors[].typeConverter.entries

`[]AwsCloudwatchLogGroupTransformerTypeConverterEntry` · required

Fields to convert. 1-5 entries.

- rule: {"repeated":{"minItems":"1","maxItems":"5"}}
- rule: type must be one of: boolean, integer, double, string

### spec.transformer.processors[].typeConverter.entries[].key

`string` · required

Key whose value is converted. 1-128 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.transformer.processors[].typeConverter.entries[].type

`string` · required

Type to convert the value to. Valid values: "boolean", "integer",
"double", "string".

- rule: {"required":true}

### spec.transformer.processors[].upperCaseString

`AwsCloudwatchLogGroupTransformerWithKeys`

Converts the values of the given keys to uppercase.

- rule: each with_keys entry must be non-empty

### spec.transformer.processors[].upperCaseString.withKeys

`[]string` · required

Keys whose values the operation applies to. 1-10 keys, each non-empty.

- rule: {"repeated":{"minItems":"1","maxItems":"10"}}

## Validation Rules

- `retention_in_days_valid_values`: retention_in_days must be one of: 0, 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653
- `log_group_class_valid_values`: log_group_class must be 'STANDARD', 'INFREQUENT_ACCESS', or 'DELIVERY' when set
- `delivery_class_no_retention`: retention_in_days must not be set (must be 0) when log_group_class is 'DELIVERY' — AWS manages retention for Delivery log groups
- `subscription_filters_max_2`: a log group supports at most 2 subscription filters — AWS enforces this limit per log group
- `metric_filter_names_unique`: each metric filter must have a unique name within the log group
- `subscription_filter_names_unique`: each subscription filter must have a unique name within the log group
- `infrequent_access_no_filters`: metric_filters and subscription_filters are not supported on INFREQUENT_ACCESS log groups — use a STANDARD class log group for filtering features
- `transformer_requires_standard_class`: transformer is only supported on STANDARD class log groups — AWS rejects PutTransformer for INFREQUENT_ACCESS and DELIVERY classes
- `log_stream_names_unique`: each log stream name must be unique within the log group
- `log_stream_names_format`: log stream names must be 1-512 characters and must not contain ':' or '*'

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCloudwatchLogGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.log_group_arn` | `string` | The Amazon Resource Name (ARN) of the log group. This is the primary identifier used to reference the log group in other AWS resources such as Step Functions logging, API Gateway access logs, and OpenSearch log publishing. |
| `status.outputs.log_group_name` | `string` | The name of the log group. Some AWS services (e.g., ElastiCache log delivery, ECS awslogs driver) reference log groups by name rather than ARN. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.subscriptionFilters[].roleArn` | AwsIamRole | `status.outputs.role_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockAgentCoreEvaluation | `spec.onlineEvaluationConfigs[].dataSource.logGroupNames` | `status.outputs.log_group_name` |
| AwsBedrockInvocationLogging | `spec.cloudwatch.logGroupName` | `status.outputs.log_group_name` |
| AwsClientVpn | `spec.connectionLog.cloudwatchLogGroup` | `status.outputs.log_group_name` |
| AwsCloudTrail | `spec.cloudwatchLogs.logGroupArn` | `status.outputs.log_group_arn` |
| AwsCloudwatchLogAnomalyDetector | `spec.logGroupArns` | `status.outputs.log_group_arn` |
| AwsCloudwatchLogDelivery | `spec.vended.destinations[].destinationResourceArn` | `status.outputs.log_group_arn` |
| AwsCloudwatchLogResourcePolicy | `spec.resourceArn` | `status.outputs.log_group_arn` |
| AwsCodeBuildProject | `spec.logsConfig.cloudwatchLogs.groupName` | `status.outputs.log_group_name` |
| AwsCognitoUserPool | `spec.logConfigurations[].cloudwatchLogGroupArn` | `status.outputs.log_group_arn` |
| AwsEcsTaskDefinition | `spec.logging.logGroup` | `status.outputs.log_group_name` |
| AwsEventBridgePipe | `spec.logConfiguration.cloudwatchLogs.logGroupArn` | `status.outputs.log_group_arn` |
| AwsFsxLustreFileSystem | `spec.logConfiguration.destination` | `status.outputs.log_group_arn` |
| AwsFsxWindowsFileSystem | `spec.auditLogConfiguration.auditLogDestination` | `status.outputs.log_group_arn` |
| AwsHttpApiGateway | `spec.stage.accessLog.destinationArn` | `status.outputs.log_group_arn` |
| AwsLambda | `spec.loggingConfig.logGroup` | `status.outputs.log_group_name` |
| AwsManagedPrometheus | `spec.logging.logGroupArn` | `status.outputs.log_group_arn` |
| AwsManagedPrometheus | `spec.queryLogging.destinations[].logGroupArn` | `status.outputs.log_group_arn` |
| AwsManagedPrometheusScraper | `spec.logging.logGroupArn` | `status.outputs.log_group_arn` |
| AwsMskCluster | `spec.logging.cloudwatchLogs.logGroup` | `status.outputs.log_group_name` |
| AwsOpenSearchDomain | `spec.logPublishingOptions[].cloudwatchLogGroupArn` | `status.outputs.log_group_arn` |
| AwsRestApiGateway | `spec.stage.accessLog.destinationArn` | `status.outputs.log_group_arn` |
| AwsRoute53ResolverQueryLog | `spec.destinationArn` | `status.outputs.log_group_arn` |
| AwsRoute53Zone | `spec.queryLogging.cloudwatchLogGroupArn` | `status.outputs.log_group_arn` |
| AwsStepFunction | `spec.logging.logDestination` | `status.outputs.log_group_arn` |
| AwsWafWebAcl | `spec.logging.destinationArn` | `status.outputs.log_group_arn` |

## See Also

- [Overview](../README.md)
