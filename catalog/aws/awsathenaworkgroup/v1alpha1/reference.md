# AwsAthenaWorkgroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsAthenaWorkgroupSpec defines the desired configuration for an Amazon Athena
workgroup.

An Athena workgroup is the resource that isolates query execution, enforces
cost controls, and manages result storage for teams or applications running
interactive SQL (or Apache Spark) analytics against data in Amazon S3, the
AWS Glue Data Catalog, and federated data sources.

Workgroups provide:
- Query result isolation: each workgroup directs results either to a
  dedicated S3 location with its own encryption settings, or to
  AWS-managed storage (managed_query_results) that needs no bucket at all.
- Cost controls: enforce per-query data scan limits to prevent runaway costs.
- Configuration enforcement: lock settings so individual queries cannot
  override the workgroup defaults.
- Engine selection: pin to a specific Athena engine version or use AUTO to
  always run the latest.
- Observability: publish CloudWatch metrics and deliver query/Spark logs to
  CloudWatch Logs, S3, or Athena-managed storage.

For the common use case (SQL analytics on S3 data), only result_configuration
with an output_location is needed. Cost controls, encryption, engine version,
identity integration, and monitoring are optional settings for production
governance.

Notes:
- The workgroup name (from metadata.name) cannot be changed after creation
  (ForceNew). Naming constraints: 1-128 characters, alphanumeric, periods,
  underscores, and hyphens only.
- Credentials, region, and deployment workflow live outside this spec in stack
  inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAthenaWorkgroup
metadata:
  name: test-athena-wg
  org: test-org
  env: dev
  id: test-athena-wg-dev
spec:
  region: us-west-2
  description: Test Athena workgroup exercising the full configuration surface
  state: ENABLED
  resultConfiguration:
    outputLocation: s3://test-athena-results/queries/
    encryptionOption: SSE_KMS
    kmsKeyArn:
      value: arn:aws:kms:us-west-2:123456789012:key/00000000-0000-0000-0000-000000000000
    expectedBucketOwner: "123456789012"
    s3AclOption: BUCKET_OWNER_FULL_CONTROL
  bytesScannedCutoffPerQuery: 10737418240
  enforceWorkgroupConfiguration: true
  publishCloudwatchMetricsEnabled: true
  requesterPaysEnabled: false
  enableMinimumEncryptionConfiguration: true
  selectedEngineVersion: AUTO
  executionRole:
    value: arn:aws:iam::123456789012:role/athena-spark-execution
  customerContentEncryptionKmsKey:
    value: arn:aws:kms:us-west-2:123456789012:key/00000000-0000-0000-0000-000000000000
  identityCenter:
    enableIdentityCenter: true
    identityCenterInstanceArn: arn:aws:sso:::instance/ssoins-0000000000000000
  s3AccessGrants:
    enableS3AccessGrants: true
    authenticationType: DIRECTORY_IDENTITY
    createUserLevelPrefix: true
  monitoring:
    cloudWatchLogging:
      logGroup: /athena/test-workgroup
      logStreamNamePrefix: test
      logTypes:
        - key: SPARK_DRIVER
          values:
            - STDOUT
            - STDERR
    managedLogging:
      kmsKey:
        value: arn:aws:kms:us-west-2:123456789012:key/00000000-0000-0000-0000-000000000000
    s3Logging:
      logLocation: s3://test-athena-logs/archive/
  forceDestroy: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.state` | `string` |  | `ENABLED` |  |
| `spec.resultConfiguration` | `AwsAthenaWorkgroupResultConfig` |  |  |  |
| `spec.resultConfiguration.outputLocation` | `string` |  |  |  |
| `spec.resultConfiguration.encryptionOption` | `string` |  |  |  |
| `spec.resultConfiguration.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.resultConfiguration.expectedBucketOwner` | `string` |  |  |  |
| `spec.resultConfiguration.s3AclOption` | `string` |  |  |  |
| `spec.managedQueryResults` | `AwsAthenaWorkgroupManagedQueryResults` |  |  |  |
| `spec.managedQueryResults.kmsKey` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.managedQueryResults.enabled` | `bool` |  | `true` |  |
| `spec.bytesScannedCutoffPerQuery` | `int64` |  |  |  |
| `spec.enforceWorkgroupConfiguration` | `bool` |  | `true` |  |
| `spec.publishCloudwatchMetricsEnabled` | `bool` |  | `true` |  |
| `spec.requesterPaysEnabled` | `bool` |  |  |  |
| `spec.enableMinimumEncryptionConfiguration` | `bool` |  |  |  |
| `spec.selectedEngineVersion` | `string` |  |  |  |
| `spec.executionRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.customerContentEncryptionKmsKey` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.identityCenter` | `AwsAthenaWorkgroupIdentityCenterConfig` |  |  |  |
| `spec.identityCenter.enableIdentityCenter` | `bool` |  |  |  |
| `spec.identityCenter.identityCenterInstanceArn` | `string` |  |  |  |
| `spec.s3AccessGrants` | `AwsAthenaWorkgroupS3AccessGrantsConfig` |  |  |  |
| `spec.s3AccessGrants.enableS3AccessGrants` | `bool` |  |  |  |
| `spec.s3AccessGrants.authenticationType` | `string` |  |  |  |
| `spec.s3AccessGrants.createUserLevelPrefix` | `bool` |  |  |  |
| `spec.monitoring` | `AwsAthenaWorkgroupMonitoringConfig` |  |  |  |
| `spec.monitoring.cloudWatchLogging` | `AwsAthenaWorkgroupCloudWatchLoggingConfig` |  |  |  |
| `spec.monitoring.cloudWatchLogging.logGroup` | `string` |  |  |  |
| `spec.monitoring.cloudWatchLogging.logStreamNamePrefix` | `string` |  |  |  |
| `spec.monitoring.cloudWatchLogging.logTypes` | `[]AwsAthenaWorkgroupLogTypeEntry` |  |  |  |
| `spec.monitoring.cloudWatchLogging.logTypes[].key` | `string` | yes |  |  |
| `spec.monitoring.cloudWatchLogging.logTypes[].values` | `[]string` | yes |  |  |
| `spec.monitoring.cloudWatchLogging.enabled` | `bool` |  | `true` |  |
| `spec.monitoring.managedLogging` | `AwsAthenaWorkgroupManagedLoggingConfig` |  |  |  |
| `spec.monitoring.managedLogging.kmsKey` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.monitoring.managedLogging.enabled` | `bool` |  | `true` |  |
| `spec.monitoring.s3Logging` | `AwsAthenaWorkgroupS3LoggingConfig` |  |  |  |
| `spec.monitoring.s3Logging.logLocation` | `string` |  |  |  |
| `spec.monitoring.s3Logging.kmsKey` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.monitoring.s3Logging.enabled` | `bool` |  | `true` |  |
| `spec.forceDestroy` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the Athena workgroup will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description of the workgroup shown in the Athena console.
Helps teams understand which application or team owns the workgroup.
Maximum 1024 characters (enforced by the AWS API).

- rule: {"string":{"maxLen":"1024"}}

### spec.state

`string` · optional (explicit presence)

Operational state of the workgroup. A DISABLED workgroup rejects new query
submissions while keeping its configuration, history, and saved queries
intact -- the safe way to pause a team's spend without deleting anything.
Defaults to ENABLED.

- default: `ENABLED`
- rule: {"string":{"in":["ENABLED","DISABLED"]}}

### spec.resultConfiguration

`AwsAthenaWorkgroupResultConfig`

Configuration for customer-managed query result storage in S3: location,
encryption, and cross-account access control. When omitted (and
managed_query_results is not set), queries must specify their own result
location or fall back to the AWS account-level Athena settings.

Mutually exclusive with managed_query_results.output_location by AWS's own
rule: a workgroup stores results either in your bucket or in AWS-managed
storage, never both.

### spec.resultConfiguration.outputLocation

`string`

S3 URI where query results are stored, e.g., "s3://my-bucket/athena-results/".
Include the trailing slash. When omitted, queries must specify their own
result location.

This is a plain string (not StringValueOrRef) because it's an S3 URI with a
user-defined path prefix, not a direct resource identifier.

### spec.resultConfiguration.encryptionOption

`string`

Server-side encryption mode for query results written to S3.

Valid values:
- "SSE_S3"   — Amazon S3-managed encryption keys (no additional cost).
- "SSE_KMS"  — AWS KMS-managed key. Requires kms_key_arn. Provides key
               rotation control and CloudTrail audit of key usage.
- "CSE_KMS"  — Client-side encryption with AWS KMS key. Requires kms_key_arn.
               Data is encrypted before leaving the Athena service.

When omitted, query results are not encrypted (unless
enable_minimum_encryption_configuration is true at the workgroup level,
which enforces at least SSE_S3).

### spec.resultConfiguration.kmsKeyArn

`string | valueFrom`

KMS key ARN for encrypting query results. Required when encryption_option is
SSE_KMS or CSE_KMS. Must not be set for SSE_S3 or when encryption is
disabled.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.resultConfiguration.expectedBucketOwner

`string`

AWS account ID that owns the S3 bucket used for query results. Set this when
the output_location bucket belongs to a different AWS account than the one
running the Athena workgroup. Ensures Athena applies the correct bucket
ownership controls.

### spec.resultConfiguration.s3AclOption

`string`

S3 ACL option applied to query result objects. The only valid value is
"BUCKET_OWNER_FULL_CONTROL", which grants the bucket owner full control of
result objects. Useful for cross-account scenarios where the query executor
writes to a bucket owned by another account.

When omitted, S3 default ACL behavior applies.

### spec.managedQueryResults

`AwsAthenaWorkgroupManagedQueryResults`

AWS-managed query result storage. When this block is present, Athena stores
query results in storage that AWS owns and operates -- no S3 bucket to
create, secure, or lifecycle. Results are retained for 24 hours and are
retrievable only through Athena APIs (GetQueryResults), which is exactly
what most programmatic and BI-driven workloads need.

Choose this over result_configuration when nothing downstream reads the
result files directly from S3. The two are mutually exclusive: AWS rejects
a workgroup that sets both managed results and an S3 output_location.

- rule: managed_query_results.kms_key literal value must be a full KMS key ARN (arn:...) -- aliases are not accepted here

### spec.managedQueryResults.kmsKey

`string | valueFrom`

KMS key that encrypts results in AWS-managed storage. When omitted, AWS
encrypts managed results with an AWS-owned key -- already encrypted at
rest; supply a key only when your compliance posture requires
customer-controlled key rotation and CloudTrail audit of key usage.

Unlike the other three KMS fields on this workgroup, this one accepts a
full key ARN ONLY -- the provider rejects "alias/..." forms here at plan
time.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.managedQueryResults.enabled

`bool` · optional (explicit presence)

Whether AWS-managed result storage is on. Unset means on: declaring the
block has always meant enabling it, and this switch lets a manifest say
the opposite out loud while keeping the block's settings in place.

- default: `true`

### spec.bytesScannedCutoffPerQuery

`int64`

Maximum number of bytes a single query is allowed to scan. Queries that
exceed this limit are cancelled automatically. This is the primary cost
control mechanism for Athena workgroups.

Must be 0 (no limit, the default) or at least 10485760 (10 MB). AWS
enforces this minimum to prevent trivially small limits that would break
most queries.

Recommended for production: set to a reasonable ceiling based on your
dataset sizes (e.g., 10737418240 for 10 GB).

### spec.enforceWorkgroupConfiguration

`bool` · optional (explicit presence)

When true (the default), workgroup settings override client-side settings
for result location, encryption, and other configuration. Individual queries
cannot override these values.

When false, queries can override workgroup settings. Useful for development
workgroups where engineers need flexibility, but not recommended for
production where consistent encryption and result locations are required.

- default: `true`

### spec.publishCloudwatchMetricsEnabled

`bool` · optional (explicit presence)

When true, Athena publishes query execution metrics (data scanned, execution
time, etc.) to CloudWatch. Useful for monitoring query performance and cost
trends across the workgroup.

- default: `true`

### spec.requesterPaysEnabled

`bool`

When true, the requester pays for data access charges when querying data in
requester-pays S3 buckets. Default is false (the bucket owner pays).

### spec.enableMinimumEncryptionConfiguration

`bool`

When true, enforces a minimum encryption level (at least SSE_S3) for all
query results written by this workgroup. Queries that do not specify
encryption will default to SSE_S3.

Useful as a compliance guardrail to ensure no query results are ever written
unencrypted, even if result_configuration.encryption_option is not set.

### spec.selectedEngineVersion

`string`

The Athena engine version to use for queries in this workgroup. Leave empty
or set to "AUTO" (the default) to use the latest available engine version.

Pinning to a specific version (e.g., "Athena engine version 3") is useful
when you need to control query behavior across engine upgrades. The actual
engine version in use is available in the effective_engine_version output.

Valid values depend on the AWS region and change over time as AWS releases
new engine versions. AWS validates at apply time.

### spec.executionRole

`string | valueFrom`

IAM role ARN assumed by the workgroup for Apache Spark workloads and for
IAM Identity Center-enabled workgroups. Standard Athena SQL workgroups do
not need this field.

For Spark, the role must have permissions to read from S3 data sources,
write results, and access the AWS Glue Data Catalog.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.customerContentEncryptionKmsKey

`string | valueFrom`

KMS key that encrypts customer content stored by Athena for this workgroup:
Spark notebook cells, session data, and saved calculation results. Relevant
only for Spark-enabled workgroups; SQL query RESULTS are encrypted through
result_configuration / managed_query_results instead.

Accepts a KMS key ARN (the default reference path) or a key alias.
When omitted, Athena encrypts notebook content with an AWS-owned key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.identityCenter

`AwsAthenaWorkgroupIdentityCenterConfig`

AWS IAM Identity Center integration. Enables trusted identity propagation
so queries run as the workforce identity of the console/IDE user instead of
a shared IAM role -- per-user auditing in CloudTrail and per-user S3 Access
Grants become possible. Create-time setting on the workgroup.

- rule: identity_center_instance_arn must be an ARN (arn:...) when set

### spec.identityCenter.enableIdentityCenter

`bool`

Enables IAM Identity Center integration for this workgroup. Queries then
execute under the propagated workforce identity of the signed-in user,
making per-user auditing and per-user data grants possible.

### spec.identityCenter.identityCenterInstanceArn

`string`

ARN of the IAM Identity Center instance to integrate with, e.g.
"arn:aws:sso:::instance/ssoins-...". Find it in the Identity Center
console or via `aws sso-admin list-instances`.

### spec.s3AccessGrants

`AwsAthenaWorkgroupS3AccessGrantsConfig`

Amazon S3 Access Grants integration for query results. When enabled, Athena
obtains result-bucket credentials from S3 Access Grants (scoped to the
calling identity) instead of the workgroup role's static S3 permissions --
the fine-grained-access companion to the identity_center block.

### spec.s3AccessGrants.enableS3AccessGrants

`bool`

Turns S3 Access Grants on for query results in this workgroup.

### spec.s3AccessGrants.authenticationType

`string`

Authentication mode used when requesting credentials from S3 Access
Grants. The only value AWS currently supports is "DIRECTORY_IDENTITY"
(the propagated IAM Identity Center identity), which is why this block
pairs with identity_center.

- rule: {"string":{"in":["DIRECTORY_IDENTITY"]}}

### spec.s3AccessGrants.createUserLevelPrefix

`bool`

When true, Athena creates a per-user prefix under the result location so
each user's results are isolated by grant, e.g.
s3://bucket/results/${user}/. Recommended with DIRECTORY_IDENTITY so
users can only read their own query results.

### spec.monitoring

`AwsAthenaWorkgroupMonitoringConfig`

Log delivery for queries and Spark sessions executed in this workgroup.
Each destination arm is enabled by its presence: CloudWatch Logs for
searchable operational logs, S3 for cheap long-term archive, and
Athena-managed storage for zero-setup retention.

### spec.monitoring.cloudWatchLogging

`AwsAthenaWorkgroupCloudWatchLoggingConfig`

Delivers logs to a CloudWatch Logs log group -- the destination to pick
when logs must be searchable, alarmable, or subscription-filtered.

### spec.monitoring.cloudWatchLogging.logGroup

`string`

Name of the CloudWatch Logs log group to publish to. When omitted, Athena
publishes to its service-default log group. 1-512 characters:
alphanumeric, periods, underscores, hyphens, and slashes.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512","pattern":"^[a-zA-Z0-9._/-]+$"}}

### spec.monitoring.cloudWatchLogging.logStreamNamePrefix

`string`

Prefix for the log stream names Athena creates inside the log group.
Useful when several workgroups share one log group. Same character rules
as log_group.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512","pattern":"^[a-zA-Z0-9._/-]+$"}}

### spec.monitoring.cloudWatchLogging.logTypes

`[]AwsAthenaWorkgroupLogTypeEntry`

Which logs to publish, keyed by worker type. Keys are Spark worker types
such as "SPARK_DRIVER" and "SPARK_EXECUTOR"; values are the log streams
to deliver for that worker, such as "STDOUT" and "STDERR". When omitted,
Athena publishes its default set.

### spec.monitoring.cloudWatchLogging.logTypes[].key

`string` · required

Worker type key, e.g. "SPARK_DRIVER" or "SPARK_EXECUTOR".

- rule: {"string":{"minLen":"1"}}

### spec.monitoring.cloudWatchLogging.logTypes[].values

`[]string` · required

Log streams to publish for the worker, e.g. "STDOUT", "STDERR".

- rule: {"repeated":{"minItems":"1"}}

### spec.monitoring.cloudWatchLogging.enabled

`bool` · optional (explicit presence)

Whether this destination is on. Unset means on: declaring the block has
always meant enabling it, and this switch lets a manifest say the opposite
out loud while keeping the block's settings in place.

- default: `true`

### spec.monitoring.managedLogging

`AwsAthenaWorkgroupManagedLoggingConfig`

Delivers logs to Athena-managed storage: zero infrastructure to create,
retrievable through the Athena console/APIs.

### spec.monitoring.managedLogging.kmsKey

`string | valueFrom`

KMS key that encrypts logs in managed storage. Accepts a key ARN (the
default reference path) or an alias. When omitted, AWS encrypts the logs
with an AWS-owned key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.monitoring.managedLogging.enabled

`bool` · optional (explicit presence)

Whether this destination is on. Unset means on: declaring the block has
always meant enabling it, and this switch lets a manifest say the opposite
out loud while keeping the block's settings in place.

- default: `true`

### spec.monitoring.s3Logging

`AwsAthenaWorkgroupS3LoggingConfig`

Delivers logs to an S3 location you own -- the cheap long-term archive
destination.

- rule: s3_logging.log_location must be an s3:// URI with a valid lowercase bucket name when set

### spec.monitoring.s3Logging.logLocation

`string`

S3 URI where logs are delivered, e.g. "s3://my-log-archive/athena/".
Plain string (not a reference) because it is a URI with a user-defined
path prefix, not a direct resource identifier.

- rule: {"string":{"maxLen":"1024"}}

### spec.monitoring.s3Logging.kmsKey

`string | valueFrom`

KMS key that encrypts delivered log objects. Accepts a key ARN (the
default reference path) or an alias. When omitted, objects use the
bucket's default encryption.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.monitoring.s3Logging.enabled

`bool` · optional (explicit presence)

Whether this destination is on. Unset means on: declaring the block has
always meant enabling it, and this switch lets a manifest say the opposite
out loud while keeping the block's settings in place.

- default: `true`

### spec.forceDestroy

`bool`

When true, all named queries and prepared statements associated with the
workgroup are deleted when the workgroup is destroyed. When false (the
default), destroying a workgroup that contains named queries or prepared
statements will fail.

## Validation Rules

- `bytes_scanned_cutoff_range`: bytes_scanned_cutoff_per_query must be 0 (no limit) or at least 10485760 (10 MB)
- `result_encryption_option_valid`: result_configuration.encryption_option must be 'SSE_S3', 'SSE_KMS', or 'CSE_KMS' when set
- `s3_acl_option_valid`: result_configuration.s3_acl_option must be 'BUCKET_OWNER_FULL_CONTROL' when set
- `managed_results_excludes_output_location`: managed_query_results cannot be combined with result_configuration.output_location; results are stored either in AWS-managed storage or in your S3 bucket, not both
- `execution_role_arn_format`: execution_role literal value must be an IAM role ARN (arn:...)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsAthenaWorkgroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.workgroup_arn` | `string` | The Amazon Resource Name (ARN) of the Athena workgroup. Used for IAM policies, cross-service permissions, and as a reference identifier. |
| `status.outputs.workgroup_name` | `string` | The name of the Athena workgroup. Used in Athena API calls (StartQueryExecution, etc.) and as a human-readable identifier. The workgroup name is unique within an AWS account and region. |
| `status.outputs.effective_engine_version` | `string` | The actual engine version in use by the workgroup. When selected_engine_version is "AUTO" or empty, this reflects the latest engine version AWS has assigned. When a specific version is pinned, this confirms the version is active. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resultConfiguration.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.managedQueryResults.kmsKey` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.executionRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.customerContentEncryptionKmsKey` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.monitoring.managedLogging.kmsKey` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.monitoring.s3Logging.kmsKey` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
