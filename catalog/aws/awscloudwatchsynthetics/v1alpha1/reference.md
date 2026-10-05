# AwsCloudwatchSynthetics

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsCloudwatchSyntheticsSpec defines CloudWatch Synthetics resources:
a CANARY (a scheduled scripted probe - a Synthetics-managed Lambda
that runs your script from an S3-staged code bundle and writes run
artifacts to S3) and/or owned GROUPS (name-and-tags containers the
console uses to aggregate canary results).

Two independently deployable arms: a canary instance monitors an
endpoint (optionally joining groups by name); a groups-only instance
manages shared groups for many canaries. Group joins reference the
group NAME, so shared groups are referenced, never fought over.

The canary's name is metadata.name (lowercase letters, digits,
hyphens, underscores - the canary charset). Renaming replaces the
canary. AWS deletes and recreates a canary that lands in
CREATE_FAILED - reprovisioning is the only repair AWS offers.

## Example

```yaml
# Canonical AwsCloudwatchSynthetics example (hack/dev manifest and
# refgen Example source): a heartbeat canary with an owned group and
# its join, run config, and retention. Literal values stand in for
# composed references so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCloudwatchSynthetics
metadata:
  name: checkout-heartbeat
  id: checkout-heartbeat
  org: test-org
  env: dev
spec:
  region: us-west-2
  canary:
    artifactBucket:
      value: my-canary-artifacts
    artifactPrefix: canary/checkout
    executionRoleArn:
      value: arn:aws:iam::123456789012:role/canary-exec
    handler: heartbeat.handler
    runtimeVersion: syn-nodejs-puppeteer-9.1
    code:
      s3Bucket:
        value: my-canary-code
      s3Key: e2e/canary-heartbeat.zip
    schedule:
      expression: rate(5 minutes)
      maxRetries: 1
    runConfig:
      memoryInMb: 960
      timeoutInSeconds: 60
      environmentVariables:
        TARGET_URL: https://checkout.example.com/health
    failureRetentionPeriod: 31
    successRetentionPeriod: 7
    startCanary: false
    deleteLambda: true
  groups:
    - name: checkout-canaries
  groupNames:
    - checkout-canaries
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.canary` | `AwsSyntheticsCanary` |  |  |  |
| `spec.canary.artifactBucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.canary.artifactPrefix` | `string` |  |  |  |
| `spec.canary.executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.canary.handler` | `string` | yes |  |  |
| `spec.canary.runtimeVersion` | `string` | yes |  |  |
| `spec.canary.code` | `AwsSyntheticsCanaryCode` | yes |  |  |
| `spec.canary.code.s3Bucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.canary.code.s3Key` | `string` | yes |  |  |
| `spec.canary.code.s3Version` | `string` |  |  |  |
| `spec.canary.schedule` | `AwsSyntheticsCanarySchedule` | yes |  |  |
| `spec.canary.schedule.expression` | `string` | yes |  |  |
| `spec.canary.schedule.durationInSeconds` | `int64` |  |  |  |
| `spec.canary.schedule.maxRetries` | `int32` |  |  |  |
| `spec.canary.runConfig` | `AwsSyntheticsCanaryRunConfig` |  |  |  |
| `spec.canary.runConfig.activeTracing` | `bool` |  |  |  |
| `spec.canary.runConfig.environmentVariables` | `map<string, string>` |  |  |  |
| `spec.canary.runConfig.ephemeralStorage` | `int32` |  |  |  |
| `spec.canary.runConfig.memoryInMb` | `int32` |  |  |  |
| `spec.canary.runConfig.timeoutInSeconds` | `int32` |  |  |  |
| `spec.canary.vpcConfig` | `AwsSyntheticsCanaryVpcConfig` |  |  |  |
| `spec.canary.vpcConfig.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.canary.vpcConfig.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.canary.vpcConfig.ipv6AllowedForDualStack` | `bool` |  |  |  |
| `spec.canary.artifactEncryptionMode` | `string` |  |  |  |
| `spec.canary.artifactEncryptionKmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.canary.failureRetentionPeriod` | `int32` |  |  |  |
| `spec.canary.successRetentionPeriod` | `int32` |  |  |  |
| `spec.canary.startCanary` | `bool` |  |  |  |
| `spec.canary.deleteLambda` | `bool` |  |  |  |
| `spec.groups` | `[]AwsSyntheticsGroup` |  |  |  |
| `spec.groups[].name` | `string` | yes |  |  |
| `spec.groupNames` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the canary and groups live in. Example:
"us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.canary

`AwsSyntheticsCanary`

The canary arm: one scripted probe.

- rule: artifact_encryption_kms_key_arn requires artifact_encryption_mode SSE_KMS

### spec.canary.artifactBucket

`string | valueFrom` · required

The S3 bucket receiving run artifacts (screenshots, HAR files,
logs). Reference an AwsS3Bucket bucket_id output or pass a literal
bucket name.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.canary.artifactPrefix

`string`

Key prefix inside the artifact bucket ("canary/prod" stores under
s3://bucket/canary/prod/). Unset stores at the bucket root.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.canary.executionRoleArn

`string | valueFrom` · required

The role the canary's Lambda executes under. Needs the canary
permissions AWS documents (s3:PutObject on the artifact bucket,
logs, cloudwatch:PutMetricData, ...). Reference an AwsIamRole
role_arn output or pass a literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.canary.handler

`string` · required

The script's entry point. For Node.js runtimes
"<fileName>.handler" where the file sits at
nodejs/node_modules/<fileName>.js inside the code bundle; for
Python, "<fileName>.<functionName>" under python/.

- rule: {"string":{"minLen":"1"}}

### spec.canary.runtimeVersion

`string` · required

The Synthetics runtime the script runs on (e.g.
"syn-nodejs-puppeteer-9.1", "syn-python-selenium-4.1"). Runtimes
deprecate on AWS's schedule - upgrading is an in-place update.

- rule: {"string":{"minLen":"1","pattern":"^syn-.+$"}}

### spec.canary.code

`AwsSyntheticsCanaryCode` · required

The canary's code bundle: a zip staged in S3, laid out per the
runtime's convention (nodejs/node_modules/<file>.js for Node.js,
python/<file>.py for Python). Local-path uploads are deliberately
not modeled - stage code through a bucket (an AwsS3ObjectSet
carries small bundles inline as base64).

- rule: {"required":true}

### spec.canary.code.s3Bucket

`string | valueFrom` · required

The bucket holding the code zip. Reference an AwsS3Bucket
bucket_id output or pass a literal bucket name.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.canary.code.s3Key

`string` · required

The zip's object key.

- rule: {"string":{"minLen":"1"}}

### spec.canary.code.s3Version

`string`

A specific object version of the zip (versioned buckets). Unset
uses the latest.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.canary.schedule

`AwsSyntheticsCanarySchedule` · required

When the canary runs.

- rule: {"required":true}

### spec.canary.schedule.expression

`string` · required

A rate ("rate(5 minutes)", "rate(1 hour)") or cron
("cron(0 8 * * ? *)") expression. "rate(0 minute)" means run once
per manual start.

- rule: {"string":{"minLen":"1","pattern":"^(rate\\(.+\\)|cron\\(.+\\))$"}}

### spec.canary.schedule.durationInSeconds

`int64`

How long the canary keeps running per schedule activation, in
seconds. Unset runs the script once per activation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.canary.schedule.maxRetries

`int32` · optional (explicit presence)

Automatic retries after a failed run (0-2). Unset keeps AWS's
default (no retries). Presence-typed so an explicit 0 is
expressible.

- rule: {"int32":{"lte":2,"gte":0}}

### spec.canary.runConfig

`AwsSyntheticsCanaryRunConfig`

Per-run sizing and environment. Unset keeps every AWS default.

- rule: memory_in_mb must be a multiple of 64 and at least 960

### spec.canary.runConfig.activeTracing

`bool`

Enable AWS X-Ray tracing for the canary's runs (supported
runtimes only; adds Lambda tracing cost).

### spec.canary.runConfig.environmentVariables

`map<string, string>`

Environment variables the script reads at run time. Keys and
values land in the Synthetics-managed Lambda - AWS does not
return them on reads, so treat them as write-only configuration
and never put secrets here (use Secrets Manager lookups in the
script instead).

### spec.canary.runConfig.ephemeralStorage

`int32` · optional (explicit presence)

Ephemeral /tmp storage per run, in MB (1024-5120). Unset keeps
AWS's default (1024).

- rule: {"int32":{"lte":5120,"gte":1024}}

### spec.canary.runConfig.memoryInMb

`int32` · optional (explicit presence)

Memory per run, in MB - a multiple of 64, at least 960. Unset
keeps the runtime's default.

### spec.canary.runConfig.timeoutInSeconds

`int32` · optional (explicit presence)

Per-run timeout, in seconds (3-840). Unset defaults to the
schedule frequency, capped at 840.

- rule: {"int32":{"lte":840,"gte":3}}

### spec.canary.vpcConfig

`AwsSyntheticsCanaryVpcConfig`

VPC placement for canaries probing private endpoints. Unset runs
the canary outside any VPC (public probing).

### spec.canary.vpcConfig.subnetIds

`[]string | valueFrom` · required

The subnets the canary's Lambda attaches to. Reference AwsSubnet
subnet_id outputs or pass literal IDs.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.canary.vpcConfig.securityGroupIds

`[]string | valueFrom`

The security groups on the canary's network interfaces. Reference
AwsSecurityGroup security_group_id outputs or pass literal IDs.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.canary.vpcConfig.ipv6AllowedForDualStack

`bool`

Allow IPv6 egress from dual-stack subnets.

### spec.canary.artifactEncryptionMode

`string`

How run artifacts are encrypted at rest in S3. Unset keeps AWS's
default (SSE_S3-managed keys).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SSE_S3","SSE_KMS"]}}

### spec.canary.artifactEncryptionKmsKeyArn

`string | valueFrom`

The customer-managed key for SSE_KMS artifact encryption.
Reference an AwsKmsKey key_arn output or pass a literal ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.canary.failureRetentionPeriod

`int32` · optional (explicit presence)

Days to retain failing-run artifacts (1-455). Unset keeps AWS's
default (31).

- rule: {"int32":{"lte":455,"gte":1}}

### spec.canary.successRetentionPeriod

`int32` · optional (explicit presence)

Days to retain successful-run artifacts (1-455). Unset keeps
AWS's default (31).

- rule: {"int32":{"lte":455,"gte":1}}

### spec.canary.startCanary

`bool`

Start the canary's schedule after create/update. False creates
the canary READY but not running - runs (and their Lambda/S3
costs) only begin once started.

### spec.canary.deleteLambda

`bool`

On destroy, also delete the Synthetics-managed Lambda behind the
canary. False (AWS's default) leaves the Lambda behind for
forensics - set true for a clean teardown.

### spec.groups

`[]AwsSyntheticsGroup`

Owned groups - each entry creates one Synthetics group this
instance owns. Groups shared by many canaries belong in ONE
owning instance; other instances join them via group_names.

### spec.groups[].name

`string` · required

The group's name in AWS (its identity - renaming replaces the
group). Also the key for the outputs maps.

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.groupNames

`[]string`

Groups THIS canary joins, by group name - owned groups from this
spec or groups that exist elsewhere. Group names are user-chosen
(never cloud-generated), so literal names compose across chart
instances.

- rule: {"repeated":{"unique":true,"items":{"string":{"minLen":"1","maxLen":"64"}}}}

## Validation Rules

- `spec.at_least_one_arm`: configure a canary, at least one owned group, or both - an empty spec manages nothing
- `spec.joins_require_canary`: group_names associates THIS canary with groups - configure the canary arm or drop the joins
- `spec.group_names_unique`: groups entries must have unique names

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCloudwatchSynthetics, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.canary_name` | `string` | The canary's name (the provider's import ID for the canary). Empty on groups-only instances. |
| `status.outputs.canary_arn` | `string` | The canary's ARN. Empty on groups-only instances. |
| `status.outputs.engine_arn` | `string` | ARN of the Synthetics-managed Lambda behind the canary. |
| `status.outputs.source_location_arn` | `string` | ARN of the canary's staged code location. |
| `status.outputs.canary_status` | `string` | The canary's lifecycle status after apply (READY, RUNNING, ...). |
| `status.outputs.group_arns` | `map<string, string>` | Owned group ARNs keyed by group name. |
| `status.outputs.group_ids` | `map<string, string>` | Owned group IDs keyed by group name (with the group name, the console's group identity). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.canary.artifactBucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.canary.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.canary.code.s3Bucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.canary.vpcConfig.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.canary.vpcConfig.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.canary.artifactEncryptionKmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
