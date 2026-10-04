# AwsSagemakerModel

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSagemakerModelSpec defines the desired configuration for an Amazon
SageMaker AI model - the immutable serving definition (container
image, model artifacts, execution role, networking) that endpoints,
batch transform jobs, and inference components deploy. The model's
AWS name derives from metadata.name.

A model is either a SINGLE container (`primary_container`) or an
inference PIPELINE of 2-15 containers executed in sequence or called
directly (`containers` + `inference_execution_mode`) - exactly one of
the two forms. Every field below is create-time only: SageMaker
models are immutable, so any change replaces the model (AWS's own
contract - roll a new model and repoint the endpoint).

## Example

```yaml
# Canonical AwsSagemakerModel example (hack/dev manifest and refgen
# Example source): an inference pipeline exercising every arm - two
# containers with environments and hostnames, uncompressed S3 model
# data with EULA acceptance, an additional adapter channel, a private
# registry image config, multi-model caching, VPC attachment, and
# network isolation. Literal ARNs stand in for composed references so
# the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerModel
metadata:
  name: churn-scoring-pipeline
  id: churn-scoring-pipeline
  org: test-org
  env: dev
spec:
  region: us-west-2
  executionRoleArn:
    value: arn:aws:iam::123456789012:role/sagemaker-execution
  enableNetworkIsolation: true
  inferenceExecutionMode: Serial
  containers:
    - image: 123456789012.dkr.ecr.us-west-2.amazonaws.com/churn-preprocess:2.1
      containerHostname: preprocess
      environment:
        LOG_LEVEL: info
      modelDataSource:
        s3Uri: s3://my-models/churn/preprocess/
        s3DataType: S3Prefix
        compressionType: None
      imageConfig:
        repositoryAccessMode: Vpc
        repositoryCredentialsProviderArn: arn:aws:lambda:us-west-2:123456789012:function:registry-creds
    - image: 123456789012.dkr.ecr.us-west-2.amazonaws.com/churn-score:2.1
      containerHostname: score
      mode: MultiModel
      multiModelCache: Disabled
      modelDataUrl: s3://my-models/churn/scoring/
      additionalModelDataSources:
        - channelName: adapters
          source:
            s3Uri: s3://my-models/churn/adapters/
            s3DataType: S3Prefix
            compressionType: None
            acceptEula: true
  vpcConfig:
    subnetIds:
      - value: subnet-0a1b2c3d4e5f60001
      - value: subnet-0a1b2c3d4e5f60002
    securityGroupIds:
      - value: sg-0a1b2c3d4e5f60001
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.enableNetworkIsolation` | `bool` |  |  |  |
| `spec.primaryContainer` | `AwsSagemakerModelContainer` |  |  |  |
| `spec.primaryContainer.image` | `string` |  |  |  |
| `spec.primaryContainer.modelPackageArn` | `string` |  |  |  |
| `spec.primaryContainer.containerHostname` | `string` |  |  |  |
| `spec.primaryContainer.environment` | `map<string, string>` |  |  |  |
| `spec.primaryContainer.mode` | `string` |  |  |  |
| `spec.primaryContainer.modelDataUrl` | `string` |  |  |  |
| `spec.primaryContainer.modelDataSource` | `AwsSagemakerModelS3DataSource` |  |  |  |
| `spec.primaryContainer.modelDataSource.s3Uri` | `string` | yes |  |  |
| `spec.primaryContainer.modelDataSource.s3DataType` | `string` |  |  |  |
| `spec.primaryContainer.modelDataSource.compressionType` | `string` |  |  |  |
| `spec.primaryContainer.modelDataSource.acceptEula` | `bool` |  |  |  |
| `spec.primaryContainer.additionalModelDataSources` | `[]AwsSagemakerModelAdditionalDataSource` |  |  |  |
| `spec.primaryContainer.additionalModelDataSources[].channelName` | `string` | yes |  |  |
| `spec.primaryContainer.additionalModelDataSources[].source` | `AwsSagemakerModelS3DataSource` | yes |  |  |
| `spec.primaryContainer.additionalModelDataSources[].source.s3Uri` | `string` | yes |  |  |
| `spec.primaryContainer.additionalModelDataSources[].source.s3DataType` | `string` |  |  |  |
| `spec.primaryContainer.additionalModelDataSources[].source.compressionType` | `string` |  |  |  |
| `spec.primaryContainer.additionalModelDataSources[].source.acceptEula` | `bool` |  |  |  |
| `spec.primaryContainer.inferenceSpecificationName` | `string` |  |  |  |
| `spec.primaryContainer.multiModelCache` | `string` |  |  |  |
| `spec.primaryContainer.imageConfig` | `AwsSagemakerModelImageConfig` |  |  |  |
| `spec.primaryContainer.imageConfig.repositoryAccessMode` | `string` |  |  |  |
| `spec.primaryContainer.imageConfig.repositoryCredentialsProviderArn` | `string` |  |  |  |
| `spec.containers` | `[]AwsSagemakerModelContainer` |  |  |  |
| `spec.containers[].image` | `string` |  |  |  |
| `spec.containers[].modelPackageArn` | `string` |  |  |  |
| `spec.containers[].containerHostname` | `string` |  |  |  |
| `spec.containers[].environment` | `map<string, string>` |  |  |  |
| `spec.containers[].mode` | `string` |  |  |  |
| `spec.containers[].modelDataUrl` | `string` |  |  |  |
| `spec.containers[].modelDataSource` | `AwsSagemakerModelS3DataSource` |  |  |  |
| `spec.containers[].modelDataSource.s3Uri` | `string` | yes |  |  |
| `spec.containers[].modelDataSource.s3DataType` | `string` |  |  |  |
| `spec.containers[].modelDataSource.compressionType` | `string` |  |  |  |
| `spec.containers[].modelDataSource.acceptEula` | `bool` |  |  |  |
| `spec.containers[].additionalModelDataSources` | `[]AwsSagemakerModelAdditionalDataSource` |  |  |  |
| `spec.containers[].additionalModelDataSources[].channelName` | `string` | yes |  |  |
| `spec.containers[].additionalModelDataSources[].source` | `AwsSagemakerModelS3DataSource` | yes |  |  |
| `spec.containers[].additionalModelDataSources[].source.s3Uri` | `string` | yes |  |  |
| `spec.containers[].additionalModelDataSources[].source.s3DataType` | `string` |  |  |  |
| `spec.containers[].additionalModelDataSources[].source.compressionType` | `string` |  |  |  |
| `spec.containers[].additionalModelDataSources[].source.acceptEula` | `bool` |  |  |  |
| `spec.containers[].inferenceSpecificationName` | `string` |  |  |  |
| `spec.containers[].multiModelCache` | `string` |  |  |  |
| `spec.containers[].imageConfig` | `AwsSagemakerModelImageConfig` |  |  |  |
| `spec.containers[].imageConfig.repositoryAccessMode` | `string` |  |  |  |
| `spec.containers[].imageConfig.repositoryCredentialsProviderArn` | `string` |  |  |  |
| `spec.inferenceExecutionMode` | `string` |  |  |  |
| `spec.vpcConfig` | `AwsSagemakerModelVpcConfig` |  |  |  |
| `spec.vpcConfig.subnetIds` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.vpcConfig.securityGroupIds` | `[]string \| valueFrom` | yes |  | AwsSecurityGroup (`status.outputs.security_group_id`) |

## Field Details

### spec.region

`string` · required

The AWS region where the model will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.executionRoleArn

`string | valueFrom` · required

IAM role SageMaker AI assumes to pull the container image from ECR
and read model artifacts from S3. The role must trust
sagemaker.amazonaws.com.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.enableNetworkIsolation

`bool`

Isolate the model container: no inbound or outbound network calls
(not even to AWS services). Model artifacts and images still load
normally; combine with `vpc_config` for private serving.

### spec.primaryContainer

`AwsSagemakerModelContainer`

The single serving container - the common form. Exactly one of
`primary_container` and `containers` must be set.

- rule: at least one of image and model_package_arn must be set
- rule: at most one of model_data_url and model_data_source may be set
- rule: environment keys must match ^[A-Za-z_][0-9A-Za-z_]*$ (max 1024 chars)
- rule: multi_model_cache requires mode MultiModel

### spec.primaryContainer.image

`string`

ECR registry path of the inference image (max 255 characters, no
whitespace). Example:
"246618743249.dkr.ecr.us-west-2.amazonaws.com/sagemaker-scikit-learn:1.2-1"
(AWS-owned per-region registries host the built-in framework
images). NOTE the registry ACCOUNT differs per region for AWS's
prebuilt images (us-west-2 is 246618743249 for the
scikit-learn/XGBoost library; us-west-1 is 746614075791) - a
wrong-region account fails CreateModel with a MISLEADING 400
claiming the repository "does not grant ... to
sagemaker.amazonaws.com service principal" (live-caught
2026-08-25). Derive accounts from the sagemaker-python-sdk
image_uri_config files, never from another region's example. At
least one of `image` and `model_package_arn` is required.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255","pattern":"^\\S+$"}}

### spec.primaryContainer.modelPackageArn

`string`

Deploy from a versioned model package in the model registry instead
of a raw image (an "arn:aws:sagemaker:...:model-package/..." ARN).
At least one of `image` and `model_package_arn` is required.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.primaryContainer.containerHostname

`string`

DNS hostname for this container (1-63 characters; letters, digits,
hyphens). Pipelines use it with Direct invocation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[0-9A-Za-z-]+$"}}

### spec.primaryContainer.environment

`map<string, string>`

Environment variables for the container (keys: letters, digits,
underscore, not starting with a digit; key and value each max 1024
characters).

### spec.primaryContainer.mode

`string`

"SingleModel" (default) or "MultiModel" (one endpoint serves many
models loaded on demand from the artifact prefix). Omitted = AWS
default (SingleModel).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SingleModel","MultiModel"]}}

### spec.primaryContainer.modelDataUrl

`string`

S3 URL of the model artifacts (a .tar.gz for SingleModel, a prefix
for MultiModel). The classic compressed form - for uncompressed
artifacts use `model_data_source` instead (at most one of the two).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"1024","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.primaryContainer.modelDataSource

`AwsSagemakerModelS3DataSource`

Model artifacts as an S3 data source (supports uncompressed
deployment and gated-model EULA acceptance). At most one of
`model_data_url` and `model_data_source`.

### spec.primaryContainer.modelDataSource.s3Uri

`string` · required

The S3 path of the artifacts. Example:
"s3://my-models/llm-7b/" (S3Prefix) or
"s3://my-models/model.tar.gz" (S3Object).

- rule: {"string":{"minLen":"1","maxLen":"1024","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.primaryContainer.modelDataSource.s3DataType

`string`

"S3Prefix" (everything under the prefix) or "S3Object" (one file).

- rule: {"string":{"in":["S3Prefix","S3Object"]}}

### spec.primaryContainer.modelDataSource.compressionType

`string`

"None" (uncompressed - required for large models deployed as
prefixes) or "Gzip" (a .tar.gz artifact).

- rule: {"string":{"in":["None","Gzip"]}}

### spec.primaryContainer.modelDataSource.acceptEula

`bool`

Accept the end-user license agreement of a gated model (must be
true for EULA-gated artifacts; AWS rejects false).

### spec.primaryContainer.additionalModelDataSources

`[]AwsSagemakerModelAdditionalDataSource`

Additional artifact channels mounted under
/opt/ml/additional-model-data-sources/<channel_name>/ (adapters,
draft models for speculative decoding).

### spec.primaryContainer.additionalModelDataSources[].channelName

`string` · required

Channel name - artifacts mount under
/opt/ml/additional-model-data-sources/<channel_name>/.

- rule: {"string":{"minLen":"1"}}

### spec.primaryContainer.additionalModelDataSources[].source

`AwsSagemakerModelS3DataSource` · required

Where the channel's artifacts live.

- rule: {"required":true}

### spec.primaryContainer.additionalModelDataSources[].source.s3Uri

`string` · required

The S3 path of the artifacts. Example:
"s3://my-models/llm-7b/" (S3Prefix) or
"s3://my-models/model.tar.gz" (S3Object).

- rule: {"string":{"minLen":"1","maxLen":"1024","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.primaryContainer.additionalModelDataSources[].source.s3DataType

`string`

"S3Prefix" (everything under the prefix) or "S3Object" (one file).

- rule: {"string":{"in":["S3Prefix","S3Object"]}}

### spec.primaryContainer.additionalModelDataSources[].source.compressionType

`string`

"None" (uncompressed - required for large models deployed as
prefixes) or "Gzip" (a .tar.gz artifact).

- rule: {"string":{"in":["None","Gzip"]}}

### spec.primaryContainer.additionalModelDataSources[].source.acceptEula

`bool`

Accept the end-user license agreement of a gated model (must be
true for EULA-gated artifacts; AWS rejects false).

### spec.primaryContainer.inferenceSpecificationName

`string`

Inference specification to use from the model package (only with
`model_package_arn`).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[0-9A-Za-z-]+$"}}

### spec.primaryContainer.multiModelCache

`string`

MultiModel caching: "Enabled" or "Disabled". Omitted = AWS default
(Enabled). Only meaningful when mode is MultiModel.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Enabled","Disabled"]}}

### spec.primaryContainer.imageConfig

`AwsSagemakerModelImageConfig`

Pull the image from a private Docker registry in your VPC instead
of ECR.

- rule: repository_credentials_provider_arn requires repository_access_mode Vpc

### spec.primaryContainer.imageConfig.repositoryAccessMode

`string`

"Platform" (ECR - the default when this block is omitted) or "Vpc"
(a private registry reachable from the model's VPC).

- rule: {"string":{"in":["Platform","Vpc"]}}

### spec.primaryContainer.imageConfig.repositoryCredentialsProviderArn

`string`

Lambda function that returns credentials for the private registry
(only with Vpc access mode; omit for registries that need no
authentication).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.containers

`[]AwsSagemakerModelContainer`

An inference pipeline of 2-15 containers. Exactly one of
`primary_container` and `containers` must be set.

- rule: {"repeated":{"maxItems":"15"}}
- rule: at least one of image and model_package_arn must be set
- rule: at most one of model_data_url and model_data_source may be set
- rule: environment keys must match ^[A-Za-z_][0-9A-Za-z_]*$ (max 1024 chars)
- rule: multi_model_cache requires mode MultiModel

### spec.containers[].image

`string`

ECR registry path of the inference image (max 255 characters, no
whitespace). Example:
"246618743249.dkr.ecr.us-west-2.amazonaws.com/sagemaker-scikit-learn:1.2-1"
(AWS-owned per-region registries host the built-in framework
images). NOTE the registry ACCOUNT differs per region for AWS's
prebuilt images (us-west-2 is 246618743249 for the
scikit-learn/XGBoost library; us-west-1 is 746614075791) - a
wrong-region account fails CreateModel with a MISLEADING 400
claiming the repository "does not grant ... to
sagemaker.amazonaws.com service principal" (live-caught
2026-08-25). Derive accounts from the sagemaker-python-sdk
image_uri_config files, never from another region's example. At
least one of `image` and `model_package_arn` is required.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255","pattern":"^\\S+$"}}

### spec.containers[].modelPackageArn

`string`

Deploy from a versioned model package in the model registry instead
of a raw image (an "arn:aws:sagemaker:...:model-package/..." ARN).
At least one of `image` and `model_package_arn` is required.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.containers[].containerHostname

`string`

DNS hostname for this container (1-63 characters; letters, digits,
hyphens). Pipelines use it with Direct invocation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[0-9A-Za-z-]+$"}}

### spec.containers[].environment

`map<string, string>`

Environment variables for the container (keys: letters, digits,
underscore, not starting with a digit; key and value each max 1024
characters).

### spec.containers[].mode

`string`

"SingleModel" (default) or "MultiModel" (one endpoint serves many
models loaded on demand from the artifact prefix). Omitted = AWS
default (SingleModel).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SingleModel","MultiModel"]}}

### spec.containers[].modelDataUrl

`string`

S3 URL of the model artifacts (a .tar.gz for SingleModel, a prefix
for MultiModel). The classic compressed form - for uncompressed
artifacts use `model_data_source` instead (at most one of the two).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"1024","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.containers[].modelDataSource

`AwsSagemakerModelS3DataSource`

Model artifacts as an S3 data source (supports uncompressed
deployment and gated-model EULA acceptance). At most one of
`model_data_url` and `model_data_source`.

### spec.containers[].modelDataSource.s3Uri

`string` · required

The S3 path of the artifacts. Example:
"s3://my-models/llm-7b/" (S3Prefix) or
"s3://my-models/model.tar.gz" (S3Object).

- rule: {"string":{"minLen":"1","maxLen":"1024","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.containers[].modelDataSource.s3DataType

`string`

"S3Prefix" (everything under the prefix) or "S3Object" (one file).

- rule: {"string":{"in":["S3Prefix","S3Object"]}}

### spec.containers[].modelDataSource.compressionType

`string`

"None" (uncompressed - required for large models deployed as
prefixes) or "Gzip" (a .tar.gz artifact).

- rule: {"string":{"in":["None","Gzip"]}}

### spec.containers[].modelDataSource.acceptEula

`bool`

Accept the end-user license agreement of a gated model (must be
true for EULA-gated artifacts; AWS rejects false).

### spec.containers[].additionalModelDataSources

`[]AwsSagemakerModelAdditionalDataSource`

Additional artifact channels mounted under
/opt/ml/additional-model-data-sources/<channel_name>/ (adapters,
draft models for speculative decoding).

### spec.containers[].additionalModelDataSources[].channelName

`string` · required

Channel name - artifacts mount under
/opt/ml/additional-model-data-sources/<channel_name>/.

- rule: {"string":{"minLen":"1"}}

### spec.containers[].additionalModelDataSources[].source

`AwsSagemakerModelS3DataSource` · required

Where the channel's artifacts live.

- rule: {"required":true}

### spec.containers[].additionalModelDataSources[].source.s3Uri

`string` · required

The S3 path of the artifacts. Example:
"s3://my-models/llm-7b/" (S3Prefix) or
"s3://my-models/model.tar.gz" (S3Object).

- rule: {"string":{"minLen":"1","maxLen":"1024","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.containers[].additionalModelDataSources[].source.s3DataType

`string`

"S3Prefix" (everything under the prefix) or "S3Object" (one file).

- rule: {"string":{"in":["S3Prefix","S3Object"]}}

### spec.containers[].additionalModelDataSources[].source.compressionType

`string`

"None" (uncompressed - required for large models deployed as
prefixes) or "Gzip" (a .tar.gz artifact).

- rule: {"string":{"in":["None","Gzip"]}}

### spec.containers[].additionalModelDataSources[].source.acceptEula

`bool`

Accept the end-user license agreement of a gated model (must be
true for EULA-gated artifacts; AWS rejects false).

### spec.containers[].inferenceSpecificationName

`string`

Inference specification to use from the model package (only with
`model_package_arn`).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[0-9A-Za-z-]+$"}}

### spec.containers[].multiModelCache

`string`

MultiModel caching: "Enabled" or "Disabled". Omitted = AWS default
(Enabled). Only meaningful when mode is MultiModel.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Enabled","Disabled"]}}

### spec.containers[].imageConfig

`AwsSagemakerModelImageConfig`

Pull the image from a private Docker registry in your VPC instead
of ECR.

- rule: repository_credentials_provider_arn requires repository_access_mode Vpc

### spec.containers[].imageConfig.repositoryAccessMode

`string`

"Platform" (ECR - the default when this block is omitted) or "Vpc"
(a private registry reachable from the model's VPC).

- rule: {"string":{"in":["Platform","Vpc"]}}

### spec.containers[].imageConfig.repositoryCredentialsProviderArn

`string`

Lambda function that returns credentials for the private registry
(only with Vpc access mode; omit for registries that need no
authentication).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.inferenceExecutionMode

`string`

How pipeline containers are invoked: "Serial" (each container's
output feeds the next) or "Direct" (callers address a container by
hostname). Only meaningful with `containers`; omitted = AWS default
(Serial).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Serial","Direct"]}}

### spec.vpcConfig

`AwsSagemakerModelVpcConfig`

Attach the model containers to your VPC (private serving,
Model Monitor over private data). Changing it replaces the model.

### spec.vpcConfig.subnetIds

`[]string | valueFrom` · required

Subnets the model ENIs are placed in (1-16).

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"1","maxItems":"16"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.vpcConfig.securityGroupIds

`[]string | valueFrom` · required

Security groups applied to the model ENIs (1-5).

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"minItems":"1","maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

## Validation Rules

- `container_form_exactly_one`: exactly one of primary_container and containers must be set
- `execution_mode_requires_pipeline`: inference_execution_mode requires containers (an inference pipeline)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSagemakerModel, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.model_name` | `string` | The model name (the AWS identity - endpoint production variants reference it). |
| `status.outputs.model_arn` | `string` | The Amazon Resource Name of the model. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.vpcConfig.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.vpcConfig.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsSagemakerEndpoint | `spec.productionVariants[].model` | `status.outputs.model_name` |
| AwsSagemakerEndpoint | `spec.shadowVariants[].model` | `status.outputs.model_name` |

## See Also

- [Overview](../README.md)
