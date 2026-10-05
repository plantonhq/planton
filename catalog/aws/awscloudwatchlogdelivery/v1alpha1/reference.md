# AwsCloudwatchLogDelivery

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsCloudwatchLogDeliverySpec defines the two ways logs leave
CloudWatch, as two independently deployable arms:

The VENDED arm is the modern delivery framework: a delivery SOURCE
wraps one AWS resource whose service vends logs (CloudFront access
logs, Bedrock knowledge-base logs, SES mail events, ...), owned
delivery DESTINATIONS wrap where logs land (S3, CloudWatch Logs,
Firehose, X-Ray), and DELIVERIES join this source to destinations -
owned ones by name or destinations that exist elsewhere by ARN. One
instance owns at most one source; a destination shared by many
sources is owned by one instance and referenced by ARN from the
rest.

The CROSS-ACCOUNT arm is the legacy subscription destination: a
named Kinesis-backed receiving endpoint (with its access policy)
that OTHER accounts' subscription filters target. Its access policy
PERSISTS at AWS after destroy - deleting the destination is real,
deleting the policy alone is a provider no-op.

## Example

```yaml
# Canonical AwsCloudwatchLogDelivery example (hack/dev manifest and
# refgen Example source): a vended pipeline (a Bedrock knowledge base's
# application logs delivered to an owned S3 destination with a
# cross-account grant) plus the legacy cross-account Kinesis
# destination arm. Literal values stand in for composed references so
# the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchLogDelivery
metadata:
  name: kb-log-pipeline
  id: kb-log-pipeline
  org: test-org
  env: dev
spec:
  region: us-west-2
  vended:
    source:
      name: kb-application-logs
      logType: APPLICATION_LOGS
      resourceArn:
        value: arn:aws:bedrock:us-west-2:123456789012:knowledge-base/EXAMPLEKB
    destinations:
      - name: central-log-archive
        destinationResourceArn:
          value: arn:aws:s3:::central-log-archive
        outputFormat: json
        policy:
          Version: "2012-10-17"
          Statement:
            - Sid: AllowProducerAccountDeliveries
              Effect: Allow
              Principal:
                AWS: arn:aws:iam::210987654321:root
              Action: logs:CreateDelivery
              Resource: "*"
    deliveries:
      - name: to-s3
        destinationName: central-log-archive
        s3Configuration:
          enableHiveCompatiblePath: true
          suffixPath: kb-logs
  crossAccountDestination:
    name: org-log-sink
    roleArn:
      value: arn:aws:iam::123456789012:role/logs-to-kinesis
    targetArn:
      value: arn:aws:kinesis:us-west-2:123456789012:stream/org-logs
    accessPolicy:
      Version: "2012-10-17"
      Statement:
        - Effect: Allow
          Principal:
            AWS: arn:aws:iam::210987654321:root
          Action: logs:PutSubscriptionFilter
          Resource: "*"
    forceUpdate: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.vended` | `AwsVendedLogDelivery` |  |  |  |
| `spec.vended.source` | `AwsVendedLogSource` |  |  |  |
| `spec.vended.source.name` | `string` | yes |  |  |
| `spec.vended.source.logType` | `string` | yes |  |  |
| `spec.vended.source.resourceArn` | `string \| valueFrom` | yes |  |  |
| `spec.vended.destinations` | `[]AwsVendedLogDestination` |  |  |  |
| `spec.vended.destinations[].name` | `string` | yes |  |  |
| `spec.vended.destinations[].destinationResourceArn` | `string \| valueFrom` |  |  | AwsS3Bucket (`status.outputs.bucket_arn`), AwsCloudwatchLogGroup (`status.outputs.log_group_arn`), AwsKinesisFirehose (`status.outputs.delivery_stream_arn`) |
| `spec.vended.destinations[].deliveryDestinationType` | `string` |  |  |  |
| `spec.vended.destinations[].outputFormat` | `string` |  |  |  |
| `spec.vended.destinations[].policy` | `object` |  |  |  |
| `spec.vended.deliveries` | `[]AwsVendedLogDeliveryEntry` |  |  |  |
| `spec.vended.deliveries[].name` | `string` | yes |  |  |
| `spec.vended.deliveries[].destinationName` | `string` |  |  |  |
| `spec.vended.deliveries[].destinationArn` | `string \| valueFrom` |  |  |  |
| `spec.vended.deliveries[].recordFields` | `[]string` |  |  |  |
| `spec.vended.deliveries[].fieldDelimiter` | `string` |  |  |  |
| `spec.vended.deliveries[].s3Configuration` | `AwsVendedLogS3DeliveryConfiguration` |  |  |  |
| `spec.vended.deliveries[].s3Configuration.enableHiveCompatiblePath` | `bool` |  |  |  |
| `spec.vended.deliveries[].s3Configuration.suffixPath` | `string` |  |  |  |
| `spec.crossAccountDestination` | `AwsCrossAccountLogDestination` |  |  |  |
| `spec.crossAccountDestination.name` | `string` | yes |  |  |
| `spec.crossAccountDestination.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.crossAccountDestination.targetArn` | `string \| valueFrom` | yes |  | AwsKinesisStream (`status.outputs.stream_arn`) |
| `spec.crossAccountDestination.accessPolicy` | `object` | yes |  |  |
| `spec.crossAccountDestination.forceUpdate` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the delivery objects live in. Example:
"us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.vended

`AwsVendedLogDelivery`

The vended-log delivery arm.

- rule: configure a source, at least one owned destination, or both
- rule: deliveries join this instance's source to destinations - configure the source or drop the deliveries
- rule: destinations entries must have unique names
- rule: deliveries entries must have unique names
- rule: each delivery's destination_name must match a destinations entry

### spec.vended.source

`AwsVendedLogSource`

The delivery source: the AWS resource whose vended logs this
pipeline carries.

### spec.vended.source.name

`string` · required

The source's name in AWS (its identity - renaming replaces it).

- rule: {"string":{"minLen":"1","maxLen":"60"}}

### spec.vended.source.logType

`string` · required

The log type this source vends, as the producing service names it
(e.g. "ACCESS_LOGS" for CloudFront, "APPLICATION_LOGS" for
Bedrock knowledge bases - each service's vended-logs docs list its
types). Changing it replaces the source.

- rule: {"string":{"minLen":"1","maxLen":"255"}}

### spec.vended.source.resourceArn

`string | valueFrom` · required

ARN of the resource whose logs this source vends (a CloudFront
distribution, a Bedrock knowledge base, ...). Any vended-logs
producer works - reference the producing kind's ARN output or
pass a literal ARN. Changing it replaces the source.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.vended.destinations

`[]AwsVendedLogDestination`

Owned delivery destinations, keyed by name. A destination shared
by many pipelines lives in ONE owning instance; other instances
reference it by ARN in their deliveries.

- rule: destination_resource_arn is required unless delivery_destination_type is XRAY (which forbids it)

### spec.vended.destinations[].name

`string` · required

The destination's name in AWS (its identity - renaming replaces
it). Also the key deliveries join on.

- rule: {"string":{"minLen":"1","maxLen":"60"}}

### spec.vended.destinations[].destinationResourceArn

`string | valueFrom`

ARN of the receiving resource: an S3 bucket, a CloudWatch log
group, or a Firehose delivery stream. AWS infers the destination
type from the ARN's service. Forbidden for XRAY destinations.

- references: AwsS3Bucket (`status.outputs.bucket_arn`), AwsCloudwatchLogGroup (`status.outputs.log_group_arn`), AwsKinesisFirehose (`status.outputs.delivery_stream_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.vended.destinations[].deliveryDestinationType

`string`

The destination's type. Usually unset - AWS derives it from
destination_resource_arn. XRAY must be declared explicitly (it
has no ARN to infer from).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["S3","CWL","FH","XRAY"]}}

### spec.vended.destinations[].outputFormat

`string`

The wire format written to the destination. Unset keeps AWS's
default for the destination type. Changing it replaces the
destination.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["json","plain","w3c","raw","parquet"]}}

### spec.vended.destinations[].policy

`object`

The destination's resource policy - the cross-account grant
letting OTHER accounts' delivery sources deliver here. Statements
allow logs:CreateDelivery to the producer accounts. Unset for
same-account pipelines.

### spec.vended.deliveries

`[]AwsVendedLogDeliveryEntry`

The deliveries joining this source to destinations. AWS allows at
most one delivery per (source, destination-type) pair - e.g. one
to S3 plus one to Firehose, never two to S3.

- rule: set exactly one of destination_name (an owned destination) and destination_arn (a destination that exists elsewhere)

### spec.vended.deliveries[].name

`string` · required

The delivery's name within this instance - the key the modules
use for the for_each entry and the outputs map. Planton-side
only; AWS identifies deliveries by a generated ID.

- rule: {"string":{"minLen":"1","maxLen":"63"}}

### spec.vended.deliveries[].destinationName

`string`

An owned destination's name (a destinations entry in this spec).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.vended.deliveries[].destinationArn

`string | valueFrom`

ARN of a delivery destination that exists elsewhere - another
instance's owned destination (reference its
destination_arns output) or any pre-existing one.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.vended.deliveries[].recordFields

`[]string`

The record fields written per log event, in order. Unset delivers
the source's default field set (each log type documents its
fields).

- rule: {"repeated":{"maxItems":"128","items":{"string":{"minLen":"1","maxLen":"64"}}}}

### spec.vended.deliveries[].fieldDelimiter

`string`

The field delimiter for plain/w3c text formats (up to 5
characters). Unset keeps the format's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"5"}}

### spec.vended.deliveries[].s3Configuration

`AwsVendedLogS3DeliveryConfiguration`

S3-destination layout settings. Ignored for non-S3 destinations.

### spec.vended.deliveries[].s3Configuration.enableHiveCompatiblePath

`bool`

Write Hive-compatible key paths (year=.../month=... partitions),
so Athena/Glue partition discovery works out of the box.

### spec.vended.deliveries[].s3Configuration.suffixPath

`string`

The key suffix path under the destination bucket. For CloudFront
sources AWS silently prepends "AWSLogs/{account-id}/CloudFront/";
both engines store the suffix WITHOUT that AWS-added prefix, so
configure only your own path segment here.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.crossAccountDestination

`AwsCrossAccountLogDestination`

The legacy cross-account subscription destination arm.

### spec.crossAccountDestination.name

`string` · required

The destination's name (its identity - renaming replaces it). No
colons or asterisks, up to 512 characters.

- rule: {"string":{"minLen":"1","maxLen":"512","pattern":"^[^:*]*$"}}

### spec.crossAccountDestination.roleArn

`string | valueFrom` · required

The role CloudWatch Logs assumes to write into the Kinesis
stream. Its trust policy must allow logs.amazonaws.com; IAM
propagation makes the first create eventually consistent (both
engines retry for up to two minutes). Reference an AwsIamRole
role_arn output or pass a literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.crossAccountDestination.targetArn

`string | valueFrom` · required

The Kinesis stream that receives the forwarded log events.
Reference an AwsKinesisStream stream_arn output or pass a literal
ARN.

- references: AwsKinesisStream (`status.outputs.stream_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKinesisStream, name: <that resource's name>, fieldPath: status.outputs.stream_arn}} -- a bare string does not parse

### spec.crossAccountDestination.accessPolicy

`object` · required

The destination's access policy: which accounts/principals may
logs:PutSubscriptionFilter against this destination. Write AWS
principals as BARE ACCOUNT IDs ("AWS": "123456789012") - the
service rejects the ARN form with "Principal section of policy
contains ARN instead of account ID: arn:aws:iam::...:root"
(server contract; the usual IAM root-ARN spelling is invalid
here). In YAML, QUOTE the id: an unquoted account id parses as a
number and IAM's grammar rejects a numeric principal ("Error
occurred while parsing accessPolicy"). NOTE the AWS contract:
destroying the policy alone never removes it - only destroying
the whole destination does.

- rule: {"required":true}

### spec.crossAccountDestination.forceUpdate

`bool` · optional (explicit presence)

Update the policy even while it is actively in use by
subscription filters. Unset keeps AWS's default behavior.

## Validation Rules

- `spec.at_least_one_arm`: configure the vended arm, the cross_account_destination arm, or both - an empty spec manages nothing

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCloudwatchLogDelivery, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.source_arn` | `string` | The vended source's ARN. Empty when the vended arm has no source. |
| `status.outputs.source_name` | `string` | The vended source's name. |
| `status.outputs.source_service` | `string` | The service AWS recorded as the source's producer (derived from resource_arn). |
| `status.outputs.destination_arns` | `map<string, string>` | Owned destination ARNs keyed by destination name - what other instances' deliveries reference. |
| `status.outputs.delivery_ids` | `map<string, string>` | AWS-generated delivery IDs keyed by delivery name (each delivery imports by this ID). |
| `status.outputs.delivery_arns` | `map<string, string>` | Delivery ARNs keyed by delivery name. |
| `status.outputs.cross_account_destination_arn` | `string` | The cross-account destination's ARN - what other accounts' subscription filters target. Empty without that arm. |
| `status.outputs.cross_account_destination_name` | `string` | The cross-account destination's name (the provider's import ID for it). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vended.destinations[].destinationResourceArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.vended.destinations[].destinationResourceArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |
| `spec.vended.destinations[].destinationResourceArn` | AwsKinesisFirehose | `status.outputs.delivery_stream_arn` |
| `spec.crossAccountDestination.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.crossAccountDestination.targetArn` | AwsKinesisStream | `status.outputs.stream_arn` |

## See Also

- [Overview](../README.md)
