# AwsSagemakerImage

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSagemakerImageSpec defines the desired configuration for an Amazon
SageMaker AI image - the named registry entry that makes YOUR
container images (custom kernels, training environments) selectable
in Studio and notebook surfaces - together with its folded VERSIONS
(each pointing at a concrete ECR image). The image's AWS name derives
from metadata.name.

Versions are AWS-numbered sequentially (1, 2, 3, ...) as entries are
created - entry POSITION in `versions` is each version's stable
identity here (the modules key satellites by index), so append new
versions at the END and never reorder existing entries.

## Example

```yaml
# Canonical AwsSagemakerImage example (hack/dev manifest and refgen
# Example source): a registry entry with one fully-annotated version.
# Literal ECR paths stand in for real images so the offline `tofu plan`
# renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerImage
metadata:
  name: team-pytorch-kernels
  id: team-pytorch-kernels
  org: test-org
  env: dev
spec:
  region: us-west-2
  roleArn:
    value: arn:aws:iam::123456789012:role/sagemaker-execution
  displayName: PyTorch Kernels
  description: Team PyTorch kernel images for Studio
  versions:
    - baseImage: 123456789012.dkr.ecr.us-west-2.amazonaws.com/team-kernels:pytorch-2.4
      aliases:
        - latest
        - stable
      horovod: false
      jobType: NOTEBOOK_KERNEL
      mlFramework: PyTorch 2.4
      processor: GPU
      programmingLang: python 3.12
      releaseNotes: CUDA 12.4 base image
      vendorGuidance: STABLE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.versions` | `[]AwsSagemakerImageVersion` |  |  |  |
| `spec.versions[].baseImage` | `string` | yes |  |  |
| `spec.versions[].aliases` | `[]string` |  |  |  |
| `spec.versions[].horovod` | `bool` |  |  |  |
| `spec.versions[].jobType` | `string` |  |  |  |
| `spec.versions[].mlFramework` | `string` |  |  |  |
| `spec.versions[].processor` | `string` |  |  |  |
| `spec.versions[].programmingLang` | `string` |  |  |  |
| `spec.versions[].releaseNotes` | `string` |  |  |  |
| `spec.versions[].vendorGuidance` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the image will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.roleArn

`string | valueFrom` · required

IAM role SageMaker assumes to pull the version images from ECR.
Must trust sagemaker.amazonaws.com. Changing it replaces the image.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.displayName

`string`

Display name shown in Studio (1-128 characters; unique within a
domain when the image is attached to one). Updates in place.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128"}}

### spec.description

`string`

Free-form description (max 512 characters). Updates in place.

- rule: {"string":{"maxLen":"512"}}

### spec.versions

`[]AwsSagemakerImageVersion`

The image's versions - each points at a concrete ECR image in the
SAME account and region. Append-only by position (see the message
comment above).

### spec.versions[].baseImage

`string` · required

ECR registry path of the container image (max 255 characters, no
whitespace; must live in the same account and region). Example:
"123456789012.dkr.ecr.us-west-2.amazonaws.com/my-kernels:pytorch-2.4"

- rule: {"string":{"minLen":"1","maxLen":"255","pattern":"^\\S+$"}}

### spec.versions[].aliases

`[]string`

Stable names for this version (e.g. "latest", "stable") - aliases
move freely between versions and update in place.

- rule: {"repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.versions[].horovod

`bool`

The image bundles Horovod (distributed training).

### spec.versions[].jobType

`string`

SageMaker job type the version is compatible with: "TRAINING",
"INFERENCE", or "NOTEBOOK_KERNEL".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["TRAINING","INFERENCE","NOTEBOOK_KERNEL"]}}

### spec.versions[].mlFramework

`string`

ML framework vended in the image, as "<name> <major.minor[.patch]>"
(e.g. "PyTorch 2.4", "TensorFlow 2.16.1").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-zA-Z]+ ?\\d+\\.\\d+(\\.\\d+)?$"}}

### spec.versions[].processor

`string`

Compute the image targets: "CPU" or "GPU".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CPU","GPU"]}}

### spec.versions[].programmingLang

`string`

Programming language vended in the image, as
"<name> <major.minor[.patch]>" (e.g. "python 3.12").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-zA-Z]+ ?\\d+\\.\\d+(\\.\\d+)?$"}}

### spec.versions[].releaseNotes

`string`

Maintainer release notes (max 255 characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.versions[].vendorGuidance

`string`

Maintainer stability signal: "NOT_PROVIDED", "STABLE",
"TO_BE_ARCHIVED", or "ARCHIVED".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NOT_PROVIDED","STABLE","TO_BE_ARCHIVED","ARCHIVED"]}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSagemakerImage, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.image_name` | `string` | The image name (the AWS identity Studio configurations reference). |
| `status.outputs.image_arn` | `string` | The Amazon Resource Name of the image. |
| `status.outputs.version_numbers` | `map<string, string>` | AWS-assigned version numbers keyed by each `versions` entry's POSITION (as a string: "0", "1", ...) - the modules' for_each keys. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)
