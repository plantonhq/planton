# AwsBedrockCustomModel

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBedrockCustomModelSpec defines the desired configuration for an Amazon
Bedrock custom model - a foundation model customized with your training
data through a Bedrock model-customization job (fine-tuning, continued
pre-training, or distillation). Deploying this component STARTS the
customization job; the job runs asynchronously in AWS (minutes to many
hours depending on data size and model) and produces the custom model
when it completes. Training incurs real cost billed per token processed.

The custom model's name is taken from `metadata.name` (1-63 characters).

EVERY spec field is create-time-immutable: a customization job cannot be
altered once started, so any change destroys the job record and custom
model and starts a new job. Job names must be unique per account for all
time (AWS never reuses them) - a recreate with the same derived job name
fails, so set `job_name` explicitly when re-running a customization (the
name-history class).

To serve real traffic, a custom model needs Provisioned Throughput
(AwsBedrockProvisionedThroughput referencing this model's output ARN) or
on-demand custom-model deployment where AWS supports it.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockCustomModel
metadata:
  name: test-support-titan-ft
  id: test-support-titan-ft
  org: test-org
  env: dev
spec:
  region: us-east-1
  baseModelArn: arn:aws:bedrock:us-east-1::foundation-model/amazon.titan-text-lite-v1
  customizationType: FINE_TUNING
  jobName: test-support-titan-ft-001
  hyperparameters:
    epochCount: "1"
    batchSize: "1"
    learningRate: "0.00001"
  roleArn:
    value: arn:aws:iam::123456789012:role/bedrock-customization
  customModelKmsKeyArn:
    value: arn:aws:kms:us-east-1:123456789012:key/abc-123
  trainingDataS3Uri: s3://test-training-bucket/data/train.jsonl
  outputDataS3Uri: s3://test-training-bucket/output/
  validationDataS3Uris:
    - s3://test-training-bucket/data/validate.jsonl
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.baseModelArn` | `string` |  |  |  |
| `spec.customizationType` | `string` |  |  |  |
| `spec.jobName` | `string` |  |  |  |
| `spec.hyperparameters` | `map<string, string>` | yes |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.customModelKmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.trainingDataS3Uri` | `string` |  |  |  |
| `spec.outputDataS3Uri` | `string` |  |  |  |
| `spec.validationDataS3Uris` | `[]string` |  |  |  |
| `spec.vpcConfig` | `AwsBedrockCustomModelVpcConfig` |  |  |  |
| `spec.vpcConfig.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.vpcConfig.securityGroupIds` | `[]string \| valueFrom` | yes |  | AwsSecurityGroup (`status.outputs.security_group_id`) |

## Field Details

### spec.region

`string` · required

The AWS region where the customization job runs and the custom model
is created. Model customization is supported in a subset of Bedrock
regions (us-east-1 has the widest base-model coverage).

- rule: {"string":{"minLen":"1"}}

### spec.baseModelArn

`string`

ARN of the base foundation model to customize (format:
arn:aws:bedrock:<region>::foundation-model/<model-id>). The model must
support the chosen customization type - see the Bedrock console's
"Custom models" page for the eligible list per region.

- rule: {"string":{"pattern":"^arn:aws[a-z-]*:bedrock:[a-z0-9-]+::foundation-model/.+$"}}

### spec.customizationType

`string`

The customization method. FINE_TUNING (the default when omitted)
trains on labeled prompt/completion pairs; CONTINUED_PRE_TRAINING
trains on unlabeled domain text; DISTILLATION transfers capability
from a larger teacher model; REINFORCEMENT_FINE_TUNING trains against
a reward signal. (IMPORTED custom models arrive through import, not
through a customization job, and are out of this kind's scope.)

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["FINE_TUNING","CONTINUED_PRE_TRAINING","DISTILLATION","REINFORCEMENT_FINE_TUNING"]}}

### spec.jobName

`string`

Name for the customization job (1-63 characters; alphanumeric plus
+ - . starting alphanumeric). Defaults to metadata.name. Job names are
unique per account FOREVER - AWS rejects reuse even after the job is
deleted - so a destroy/recreate of this component needs a fresh
job_name.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-zA-Z0-9](-*[a-zA-Z0-9+\\-.])*$"}}

### spec.hyperparameters

`map<string, string>` · required

Training hyperparameters as key/value strings - the keys and legal
values are defined per base model in the Bedrock documentation
(typical keys: "epochCount", "batchSize", "learningRate",
"learningRateWarmupSteps"). At least one entry is required by AWS.
Example: {"epochCount": "1", "batchSize": "1", "learningRate": "0.00001"}

- rule: {"map":{"minPairs":"1"}}

### spec.roleArn

`string | valueFrom` · required

IAM role Bedrock assumes to run the job. The role must trust
bedrock.amazonaws.com and grant read access to the training/validation
S3 locations and write access to the output location (plus KMS
permissions when the buckets or the model are customer-key encrypted).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.customModelKmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for encrypting the resulting custom
model. Without it, AWS encrypts with a Bedrock-managed key.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.trainingDataS3Uri

`string`

S3 URI of the training dataset (s3://bucket/key). Format requirements
(JSONL with prompt/completion fields for fine-tuning; plain text for
continued pre-training) are per customization type in the Bedrock
docs.

- rule: {"string":{"pattern":"^s3://.+$"}}

### spec.outputDataS3Uri

`string`

S3 URI (s3://bucket/prefix) where Bedrock writes job outputs
(metrics, logs).

- rule: {"string":{"pattern":"^s3://.+$"}}

### spec.validationDataS3Uris

`[]string`

S3 URIs of up to 10 validation datasets - Bedrock reports per-dataset
validation metrics on the finished job.

- rule: {"repeated":{"maxItems":"10","items":{"string":{"pattern":"^s3://.+$"}}}}

### spec.vpcConfig

`AwsBedrockCustomModelVpcConfig`

Run the job inside your VPC (both members required together) - for
training data reachable only through private networking. Omit to run
with Bedrock's managed networking.

### spec.vpcConfig.subnetIds

`[]string | valueFrom` · required

Subnets the job's network interfaces attach to (at least one).

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.vpcConfig.securityGroupIds

`[]string | valueFrom` · required

Security groups applied to those interfaces (at least one).

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockCustomModel, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.custom_model_arn` | `string` | ARN of the resulting custom model - the value AwsBedrockProvisionedThroughput references to buy serving capacity for this model. |
| `status.outputs.custom_model_name` | `string` | The custom model's name. Matches metadata.name. |
| `status.outputs.job_arn` | `string` | ARN of the customization job that produced (or is producing) the model - the stable identifier the job record is tracked by. |
| `status.outputs.job_status` | `string` | Status of the customization job at the end of the deploy: InProgress, Completed, Failed, Stopping, or Stopped. Training continues asynchronously after the deploy returns - poll the job (or re-read this component) for completion. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.customModelKmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.vpcConfig.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.vpcConfig.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockProvisionedThroughput | `spec.modelArn` | `status.outputs.custom_model_arn` |

## See Also

- [Overview](../README.md)
