# AwsCloudTrail

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsCloudTrailSpec defines the desired configuration for an AWS
CloudTrail trail - the account's API audit log.

A trail records AWS API calls (management events, and optionally
data and Insights events) and delivers them as log files to an S3
bucket, with optional fan-out to CloudWatch Logs and SNS. The trail
name is metadata.name (AWS requires 3-128 characters: letters,
digits, '.', '-', '_'; it must start and end with a letter or
digit).

The delivery bucket must carry a bucket POLICY granting
"cloudtrail.amazonaws.com" s3:PutObject on the trail's prefix and
s3:GetBucketAcl on the bucket - AWS rejects trail creation without
it ("Incorrect S3 bucket policy is detected"). The policy lives on
the bucket (AwsS3Bucket spec.policy), not on this component.

Destroying this component deletes the trail (a real delete); the
delivered log files stay in the bucket.

## Example

```yaml
# Canonical AwsCloudTrail example (hack/dev manifest and refgen Example
# source): a multi-region audit trail with log-file validation, advanced
# event selectors, Insights, and CloudWatch Logs mirroring. Literal
# ARNs/names stand in for composed references so the offline `tofu plan`
# renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudTrail
metadata:
  name: org-audit-trail
  id: org-audit-trail
  org: test-org
  env: dev
spec:
  region: us-west-2
  s3BucketName:
    value: my-audit-trail-logs
  s3KeyPrefix: audit
  isMultiRegionTrail: true
  enableLogFileValidation: true
  kmsKeyId:
    value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
  snsTopicName:
    value: audit-trail-delivery
  cloudwatchLogs:
    logGroupArn:
      value: arn:aws:logs:us-west-2:123456789012:log-group:org-audit-trail
    roleArn:
      value: arn:aws:iam::123456789012:role/cloudtrail-to-cloudwatch
  advancedEventSelectors:
    - name: Management events
      fieldSelectors:
        - field: eventCategory
          equals: ["Management"]
    - name: S3 object writes in the data lake
      fieldSelectors:
        - field: eventCategory
          equals: ["Data"]
        - field: resources.type
          equals: ["AWS::S3::Object"]
        - field: resources.ARN
          startsWith: ["arn:aws:s3:::my-data-lake/"]
        - field: readOnly
          equals: ["false"]
  insightTypes:
    - ApiCallRateInsight
    - ApiErrorRateInsight
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.s3BucketName` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.s3KeyPrefix` | `string` |  |  |  |
| `spec.isMultiRegionTrail` | `bool` |  |  |  |
| `spec.includeGlobalServiceEvents` | `bool` |  |  |  |
| `spec.isOrganizationTrail` | `bool` |  |  |  |
| `spec.enableLogging` | `bool` |  |  |  |
| `spec.enableLogFileValidation` | `bool` |  |  |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.snsTopicName` | `string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_name`) |
| `spec.cloudwatchLogs` | `AwsCloudTrailCloudwatchLogs` |  |  |  |
| `spec.cloudwatchLogs.logGroupArn` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.cloudwatchLogs.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.eventSelectors` | `[]AwsCloudTrailEventSelector` |  |  |  |
| `spec.eventSelectors[].readWriteType` | `string` |  |  |  |
| `spec.eventSelectors[].includeManagementEvents` | `bool` |  |  |  |
| `spec.eventSelectors[].excludeManagementEventSources` | `[]string` |  |  |  |
| `spec.eventSelectors[].dataResources` | `[]AwsCloudTrailDataResource` |  |  |  |
| `spec.eventSelectors[].dataResources[].type` | `string` |  |  |  |
| `spec.eventSelectors[].dataResources[].values` | `[]string` | yes |  |  |
| `spec.advancedEventSelectors` | `[]AwsCloudTrailAdvancedEventSelector` |  |  |  |
| `spec.advancedEventSelectors[].name` | `string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors` | `[]AwsCloudTrailFieldSelector` | yes |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].field` | `string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].equals` | `[]string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].notEquals` | `[]string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].startsWith` | `[]string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].notStartsWith` | `[]string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].endsWith` | `[]string` |  |  |  |
| `spec.advancedEventSelectors[].fieldSelectors[].notEndsWith` | `[]string` |  |  |  |
| `spec.insightTypes` | `[]string` |  |  |  |
| `spec.organizationDelegatedAdminAccountId` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the trail lives in (its "home region"). A
multi-region trail still has exactly one home region - manage it
from there.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.s3BucketName

`string | valueFrom` · required

The S3 bucket that receives the trail's log files. The bucket
needs the CloudTrail service-principal bucket policy (see the
message comment) BEFORE the trail is created.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.s3KeyPrefix

`string`

Key prefix for the delivered log files inside the bucket
(logs land under "<prefix>/AWSLogs/<account-id>/...").

- rule: {"string":{"maxLen":"2000"}}

### spec.isMultiRegionTrail

`bool`

Record API activity in ALL regions, not just the home region.
The posture AWS recommends for audit trails - a single-region
trail misses activity everywhere else.

### spec.includeGlobalServiceEvents

`bool` · optional (explicit presence)

Record global-service events (IAM, STS, CloudFront). Unset =
enabled (the AWS default). Disable only when another trail
already captures them - duplicate global events double storage.

### spec.isOrganizationTrail

`bool`

Capture activity from EVERY account in the AWS Organization into
this one trail. Requires running in the organization's management
account (or its delegated CloudTrail administrator - see
organization_delegated_admin_account_id) with all-features
organizations enabled.

### spec.enableLogging

`bool` · optional (explicit presence)

Start delivering events as soon as the trail exists. Unset =
logging on (the AWS default). Set false to create the trail
stopped (e.g. staging a trail before cutover).

### spec.enableLogFileValidation

`bool`

Write a digest file every hour so log tampering is detectable
(SHA-256 chain over delivered files). Audit-grade trails want
this on.

### spec.kmsKeyId

`string | valueFrom`

KMS key that encrypts the delivered log files (SSE-KMS instead of
SSE-S3). The key policy must allow CloudTrail to encrypt with it
("Allow CloudTrail to encrypt logs" grant for
cloudtrail.amazonaws.com). Unset = SSE-S3.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.snsTopicName

`string | valueFrom`

SNS topic notified on every log-file delivery (the topic policy
must allow cloudtrail.amazonaws.com to publish). Unset = no
notifications.

- references: AwsSnsTopic (`status.outputs.topic_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_name}} -- a bare string does not parse

### spec.cloudwatchLogs

`AwsCloudTrailCloudwatchLogs`

Mirror events to a CloudWatch Logs group for live querying and
metric filters (S3 delivery continues regardless). AWS requires
the group and the delivery role TOGETHER - presence of this
message wires both.

### spec.cloudwatchLogs.logGroupArn

`string | valueFrom` · required

The CloudWatch Logs group ARN. AWS expects the ":*" suffix form
(e.g. "arn:aws:logs:us-west-2:123456789012:log-group:my-trail:*");
both engines append ":*" when the referenced value lacks it.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.cloudwatchLogs.roleArn

`string | valueFrom` · required

The IAM role CloudTrail assumes to write into the group.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.eventSelectors

`[]AwsCloudTrailEventSelector`

Classic event selectors: management-event scope plus coarse data
events for S3 objects, Lambda functions, and DynamoDB tables. At
most 5. Mutually exclusive with advanced_event_selectors - AWS
keeps exactly one selector style per trail.

- rule: {"repeated":{"maxItems":"5"}}

### spec.eventSelectors[].readWriteType

`string`

Which API calls to record. Unset = "All".

- rule: {"string":{"in":["","ReadOnly","WriteOnly","All"]}}

### spec.eventSelectors[].includeManagementEvents

`bool` · optional (explicit presence)

Record management events (control-plane calls). Unset = enabled
(the AWS default); disabling leaves only the data events below.

### spec.eventSelectors[].excludeManagementEventSources

`[]string`

Management-event sources to EXCLUDE (only "kms.amazonaws.com" and
"rdsdata.amazonaws.com" are accepted by AWS) - high-volume
sources that drown audit logs.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["kms.amazonaws.com","rdsdata.amazonaws.com"]}}}}

### spec.eventSelectors[].dataResources

`[]AwsCloudTrailDataResource`

Data-event scopes: which objects/functions/tables to record
data-plane activity for. Data events bill per event - scope
deliberately.

### spec.eventSelectors[].dataResources[].type

`string`

The data-resource type.

- rule: {"string":{"in":["AWS::DynamoDB::Table","AWS::Lambda::Function","AWS::S3::Object"]}}

### spec.eventSelectors[].dataResources[].values

`[]string` · required

ARN scopes: e.g. "arn:aws:s3:::my-bucket/" records every object
in the bucket; "arn:aws:s3" records ALL buckets;
"arn:aws:lambda" records all functions.

- rule: {"repeated":{"minItems":"1","maxItems":"250"}}

### spec.advancedEventSelectors

`[]AwsCloudTrailAdvancedEventSelector`

Advanced event selectors: fine-grained field matching (eventName,
resources.ARN prefixes, eventCategory, ...) over management, data,
and network-activity events. The style AWS recommends for new
trails. Mutually exclusive with event_selectors.

### spec.advancedEventSelectors[].name

`string`

Display name for the selector (shown in the CloudTrail console).

- rule: {"string":{"maxLen":"1000"}}

### spec.advancedEventSelectors[].fieldSelectors

`[]AwsCloudTrailFieldSelector` · required

The AND-set of field conditions. Every advanced selector needs at
least a field selector on "eventCategory" (AWS rejects selectors
without one).

- rule: {"repeated":{"minItems":"1"}}
- rule: set at least one of equals, not_equals, starts_with, not_starts_with, ends_with, not_ends_with

### spec.advancedEventSelectors[].fieldSelectors[].field

`string`

The event field to match.

- rule: {"string":{"in":["errorCode","eventCategory","eventName","eventSource","eventType","readOnly","resources.ARN","resources.type","sessionCredentialFromConsole","userIdentity.arn","vpcEndpointId"]}}

### spec.advancedEventSelectors[].fieldSelectors[].equals

`[]string`

Exact-match values (OR within the list).

### spec.advancedEventSelectors[].fieldSelectors[].notEquals

`[]string`

Exact-mismatch values.

### spec.advancedEventSelectors[].fieldSelectors[].startsWith

`[]string`

Prefix-match values.

### spec.advancedEventSelectors[].fieldSelectors[].notStartsWith

`[]string`

Prefix-mismatch values.

### spec.advancedEventSelectors[].fieldSelectors[].endsWith

`[]string`

Suffix-match values.

### spec.advancedEventSelectors[].fieldSelectors[].notEndsWith

`[]string`

Suffix-mismatch values.

### spec.insightTypes

`[]string`

CloudTrail Insights engines to run on this trail: anomaly
detection over call rates and error rates. Insights events bill
separately.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["ApiCallRateInsight","ApiErrorRateInsight"]}}}}

### spec.organizationDelegatedAdminAccountId

`string`

Register this AWS account ID as the organization's delegated
CloudTrail administrator - an ACCOUNT-GLOBAL act (one delegation
per organization, performed from the management account), kept
here because it exists to run organization trails from a member
account. Most deployments leave this unset. The delegation is
deregistered on destroy.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]{12}$"}}

## Validation Rules

- `one_selector_style`: event_selectors and advanced_event_selectors are mutually exclusive - pick one style
- `delegated_admin_requires_org_trail`: organization_delegated_admin_account_id requires is_organization_trail = true

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCloudTrail, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.trail_arn` | `string` | The trail's ARN (also the provider's import ID). |
| `status.outputs.home_region` | `string` | The trail's home region - the region that manages a multi-region trail. |
| `status.outputs.sns_topic_arn` | `string` | The ARN of the SNS topic notified on log delivery (set only when spec.sns_topic_name is configured). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.s3BucketName` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.snsTopicName` | AwsSnsTopic | `status.outputs.topic_name` |
| `spec.cloudwatchLogs.logGroupArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |
| `spec.cloudwatchLogs.roleArn` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)
