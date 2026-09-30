# AwsEventBridgeBus

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsEventBridgeBusSpec defines the desired configuration for an AWS EventBridge
custom event bus.

EventBridge is a serverless event bus that connects applications using events.
Every AWS account includes a "default" event bus that receives events from AWS
services automatically. This component creates **custom** event buses for
application-defined events and partner integrations.

Custom buses isolate event traffic from the default bus, enabling fine-grained
access control, dedicated encryption, and independent dead-letter queue routing.
EventBridge rules (a separate resource) attach to a bus and route matching
events to targets such as Lambda, SQS, SNS, and Step Functions.

Notes:
- The bus name is derived from `metadata.name`. You cannot create a bus named
  "default" (it already exists in every account).
- For partner event buses, set `event_source_name` to the partner source name.
  The bus name (`metadata.name`) must match the event source name.
- Credentials, region, and deployment workflow live outside this spec in stack
  inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEventBridgeBus
metadata:
  name: test-bus
  org: test-org
  env: dev
  id: test-bus-dev
  annotations:
    planton.dev/provisioner: pulumi
spec:
  region: us-east-1
  description: Test event bus for local development
  # Event archives recorded from this bus -- replay-able event history
  # (EventBridge StartReplay). An empty archive costs nothing; storage bills
  # per GB of archived events.
  archives:
    # Pattern-scoped archive with every knob: retention window, event
    # pattern, and a customer-managed key for the archived events at rest.
    - name: orders-archive
      description: Replay source for the orders domain
      retentionDays: 30
      eventPattern:
        source:
          - orders.service
      kmsKeyIdentifier:
        value: arn:aws:kms:us-east-1:123456789012:key/00000000-0000-0000-0000-000000000000
    # Minimal archive: every event on the bus, retained indefinitely,
    # AWS-owned key.
    - name: full-history
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.kmsKeyIdentifier` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.eventSourceName` | `string` |  |  |  |
| `spec.deadLetterConfig` | `AwsEventBridgeBusDeadLetterConfig` |  |  |  |
| `spec.deadLetterConfig.arn` | `string \| valueFrom` | yes |  | AwsSqsQueue (`status.outputs.queue_arn`) |
| `spec.logConfig` | `AwsEventBridgeBusLogConfig` |  |  |  |
| `spec.logConfig.level` | `string` | yes |  |  |
| `spec.logConfig.includeDetail` | `string` |  |  |  |
| `spec.resourcePolicy` | `object` |  |  |  |
| `spec.archives` | `[]AwsEventBridgeBusArchive` |  |  |  |
| `spec.archives[].name` | `string` | yes |  |  |
| `spec.archives[].description` | `string` |  |  |  |
| `spec.archives[].retentionDays` | `int32` |  |  |  |
| `spec.archives[].eventPattern` | `object` |  |  |  |
| `spec.archives[].kmsKeyIdentifier` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description of the event bus.
Maximum length is 512 characters.

- rule: {"string":{"maxLen":"512"}}

### spec.kmsKeyIdentifier

`string | valueFrom`

KMS key identifier for encrypting events on this bus. Accepts a KMS key
ARN, key ID, key alias, or key alias ARN. When omitted, events are
encrypted with an AWS-owned key at no additional cost.

Accepts a direct value or a reference to an AwsKmsKey resource.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.eventSourceName

`string`

Partner event source name. Set this only when creating a bus for a SaaS
partner integration (e.g., Datadog, Zendesk, PagerDuty).

The value must match the pattern: aws.partner/{partner}/{...} and the bus
name (`metadata.name`) must match this value exactly.

This field is immutable — changing it forces bus replacement.

### spec.deadLetterConfig

`AwsEventBridgeBusDeadLetterConfig`

Dead letter queue configuration for the event bus. When set, events that
fail delivery to any rule target on this bus are routed to the specified
SQS queue for investigation and reprocessing.

This is the bus-level DLQ — it catches events that cannot be delivered
to ANY target on any rule attached to this bus. Individual rules can also
have their own DLQ configuration (configured on AwsEventBridgeRule).

### spec.deadLetterConfig.arn

`string | valueFrom` · required

ARN of the SQS queue to use as the dead letter queue. The queue must
exist in the same AWS account and region as the event bus.

Accepts a direct ARN or a reference to an AwsSqsQueue resource.

- references: AwsSqsQueue (`status.outputs.queue_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSqsQueue, name: <that resource's name>, fieldPath: status.outputs.queue_arn}} -- a bare string does not parse

### spec.logConfig

`AwsEventBridgeBusLogConfig`

Logging configuration for the event bus. When set, EventBridge writes
event delivery logs to CloudWatch Logs. Useful for debugging event
routing, monitoring delivery failures, and auditing event traffic.

- rule: level must be one of: OFF, ERROR, INFO, TRACE
- rule: include_detail must be one of: NONE, FULL

### spec.logConfig.level

`string` · required

Logging level. Controls which events are logged.

Valid values:
- "OFF"   — no logging
- "ERROR" — log only delivery failures
- "INFO"  — log delivery successes and failures
- "TRACE" — log all events including matched/unmatched (most verbose)

When the log_config block is provided, this field is required.

- rule: {"required":true}

### spec.logConfig.includeDetail

`string`

Whether to include the full event detail in log entries.

Valid values:
- "NONE" — exclude event detail from logs (smaller log volume)
- "FULL" — include complete event detail in each log entry

Default behavior when omitted: "NONE".

### spec.resourcePolicy

`object`

Resource-based policy for the event bus. Controls which AWS principals
(accounts, organizations, or roles) may put events onto this bus —
the mechanism behind cross-account event ingestion. Expressed as a
standard IAM policy document structure; one policy per bus (statements
express per-account/per-org grants). When unset, only the owning
account can put events.

### spec.archives

`[]AwsEventBridgeBusArchive`

Event archives on this bus. An archive continuously records events
delivered to the bus (all events, or the subset matching its event
pattern) so they can be replayed later — the disaster-recovery and
backfill lever for event-driven systems. Replay itself is an on-demand
operation (EventBridge StartReplay), not declarative configuration; the
archive defines what is recorded and for how long.

Each entry materializes one archive sourced from THIS bus. Archives are
bus-scoped children: they cannot outlive the bus and cannot record from
any other source. An empty archive costs nothing — storage is billed
per GB of archived events.

### spec.archives[].name

`string` · required

Name of the archive. Unique per AWS account and region; at most 48
characters of letters, digits, dots, hyphens, and underscores.

- rule: {"required":true,"string":{"maxLen":"48","pattern":"^[.\\-_A-Za-z0-9]+$"}}

### spec.archives[].description

`string`

Human-readable description of the archive.
Maximum length is 512 characters.

- rule: {"string":{"maxLen":"512"}}

### spec.archives[].retentionDays

`int32`

Maximum number of days to retain archived events. Leave at 0 to retain
events indefinitely (the AWS default).

- rule: {"int32":{"gte":0}}

### spec.archives[].eventPattern

`object`

Event pattern selecting which of the bus's events are archived,
expressed as the standard EventBridge pattern structure. When omitted,
EVERY event delivered to the bus is archived.

Example:
  source: ["orders.service"]
  detail-type: ["OrderPlaced"]

### spec.archives[].kmsKeyIdentifier

`string | valueFrom`

KMS key for encrypting this archive. Accepts a KMS key ARN, key ID, key
alias, or key alias ARN. When omitted, the archive is encrypted with an
AWS-owned key at no additional cost. Note: this is the ARCHIVE's key —
it is independent of the bus's own kms_key_identifier, and AWS applies
it only to events at rest in the archive.

Accepts a direct value or a reference to an AwsKmsKey resource.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

## Validation Rules

- `event_source_name_pattern`: event_source_name must match the pattern aws.partner/{partner}/{...} (e.g., aws.partner/example.com/tenant/event-source)
- `archive_names_unique`: archive names must be unique within the bus

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEventBridgeBus, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.bus_name` | `string` | The name of the event bus. This is the primary identifier used in EventBridge API calls and in rule configurations that target this bus. |
| `status.outputs.bus_arn` | `string` | The Amazon Resource Name (ARN) of the event bus. Used for IAM policies, cross-account event delivery, and as a reference in other resources. |
| `status.outputs.archives` | `[]AwsEventBridgeBusArchiveOutput` | The archives declared in spec.archives — one entry per archive, exposing the identifiers replay operations (EventBridge StartReplay) and IAM policies need. |
| `status.outputs.archives[].name` | `string` | The archive's spec.archives entry name. |
| `status.outputs.archives[].arn` | `string` | The Amazon Resource Name (ARN) of the archive. StartReplay calls and IAM policies reference the archive by this ARN. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsKeyIdentifier` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.deadLetterConfig.arn` | AwsSqsQueue | `status.outputs.queue_arn` |
| `spec.archives[].kmsKeyIdentifier` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAppSyncApi | `spec.datasources[].eventbridge.eventBusArn` | `status.outputs.bus_arn` |
| AwsEventBridgeRule | `spec.eventBusName` | `status.outputs.bus_name` |
| AwsEventBridgeScheduler | `spec.target.arn` | `status.outputs.bus_arn` |
| AwsSesConfigurationSet | `spec.eventDestinations[].eventBus` | `status.outputs.bus_arn` |

## See Also

- [Overview](../README.md)
