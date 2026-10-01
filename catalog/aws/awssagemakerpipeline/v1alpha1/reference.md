# AwsSagemakerPipeline

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsSagemakerPipelineSpec defines the desired configuration for an
Amazon SageMaker AI pipeline - the ML workflow DAG (processing,
training, evaluation, registration steps) that executions run
against. The pipeline's AWS name derives from metadata.name.

The definition itself is SageMaker's pipeline-definition JSON schema
(normally produced by the SageMaker Python SDK's pipeline.definition()).
Provide it inline (`definition`) or point at an S3 object
(`definition_s3_location`) - exactly one. Creating a pipeline is
free; only executions bill.

## Example

```yaml
# Canonical AwsSagemakerPipeline example (hack/dev manifest and refgen
# Example source): an S3-located definition with parallelism and
# display metadata. Literal ARNs stand in for composed references so
# the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerPipeline
metadata:
  name: churn-training
  id: churn-training
  org: test-org
  env: dev
spec:
  region: us-west-2
  displayName: churn-training-nightly
  description: Nightly churn-model retraining
  roleArn:
    value: arn:aws:iam::123456789012:role/sagemaker-pipeline-execution
  definitionS3Location:
    bucket:
      value: my-pipeline-definitions
    objectKey: pipelines/churn-training.json
    versionId: 3sL4kqtJlcpXroDTDmJ+rmSpXd3dIbrHY+MTRCxf3vjVBH40Nr8X8gdRQBpUMLUo
  parallelismMaxSteps: 4
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.definition` | `object` |  |  |  |
| `spec.definitionS3Location` | `AwsSagemakerPipelineDefinitionS3Location` |  |  |  |
| `spec.definitionS3Location.bucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.definitionS3Location.objectKey` | `string` | yes |  |  |
| `spec.definitionS3Location.versionId` | `string` |  |  |  |
| `spec.parallelismMaxSteps` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the pipeline will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.displayName

`string`

Display name shown in Studio (1-256 characters; letters, digits,
hyphens). Omitted = the modules reuse the pipeline name.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"256","pattern":"^[0-9A-Za-z]([0-9A-Za-z-])*$"}}

### spec.description

`string`

Free-form description (max 3072 characters).

- rule: {"string":{"maxLen":"3072"}}

### spec.roleArn

`string | valueFrom` · required

IAM role pipeline executions assume to run their steps (training
jobs, processing jobs, model registration). Must trust
sagemaker.amazonaws.com.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.definition

`object`

The pipeline-definition JSON inline, as structured data (the
SageMaker SDK's pipeline.definition() output). Exactly one of
`definition` and `definition_s3_location`.

### spec.definitionS3Location

`AwsSagemakerPipelineDefinitionS3Location`

Read the definition JSON from S3 instead of inlining it. Exactly
one of `definition` and `definition_s3_location`. NOTE: AWS's
describe API returns only the RESOLVED definition, never this
location - drift on the S3 object is invisible to refresh.

### spec.definitionS3Location.bucket

`string | valueFrom` · required

The bucket holding the definition object.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.definitionS3Location.objectKey

`string` · required

The object key of the definition JSON.

- rule: {"string":{"minLen":"1"}}

### spec.definitionS3Location.versionId

`string`

Pin a specific object version. Omitted = latest.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.parallelismMaxSteps

`int32` · optional (explicit presence)

Default cap on steps executed in parallel across this pipeline's
executions (>= 1). Omitted = no pipeline-level cap.

- rule: {"int32":{"gte":1}}

## Validation Rules

- `definition_exactly_one`: exactly one of definition and definition_s3_location must be set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSagemakerPipeline, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.pipeline_name` | `string` | The pipeline name (the AWS identity executions start against). |
| `status.outputs.pipeline_arn` | `string` | The Amazon Resource Name of the pipeline. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.definitionS3Location.bucket` | AwsS3Bucket | `status.outputs.bucket_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsEventBridgeScheduler | `spec.target.arn` | `status.outputs.pipeline_arn` |

## See Also

- [Overview](../README.md)
