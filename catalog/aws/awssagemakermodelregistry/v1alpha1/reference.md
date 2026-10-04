# AwsSagemakerModelRegistry

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSagemakerModelRegistrySpec defines the desired configuration for
an Amazon SageMaker model package group - the model registry's unit
of organization, holding the versioned model packages a team
registers, approves, and deploys from. The group's AWS name derives
from metadata.name.

The group itself is a named shell: model package VERSIONS are
registered into it by training pipelines and SDK calls, never
declaratively. Everything here except the resource policy is
create-time only (even the description replaces the group -
provider-enforced).

## Example

```yaml
# Canonical AwsSagemakerModelRegistry example (hack/dev manifest and
# refgen Example source): a described group with a cross-account
# sharing policy. Literal account IDs stand in for real principals so
# the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerModelRegistry
metadata:
  name: churn-models
  id: churn-models
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Registered churn-scoring model versions
  resourcePolicy:
    Version: "2012-10-17"
    Statement:
      - Sid: AllowCrossAccountDescribe
        Effect: Allow
        Principal:
          AWS: arn:aws:iam::210987654321:root
        Action:
          - sagemaker:DescribeModelPackage
          - sagemaker:ListModelPackages
        Resource: "*"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.resourcePolicy` | `object` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the model package group will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Free-form description (1-1024 characters when set). Changing it
REPLACES the group (provider-enforced) - write it once, well.

- rule: {"string":{"maxLen":"1024"}}

### spec.resourcePolicy

`object`

IAM resource policy attached to the group (cross-account model
sharing - grant other accounts DescribeModelPackage /
CreateModelPackage on the group). The policy document as structured
JSON; the folded policy resource updates in place.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSagemakerModelRegistry, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.model_package_group_name` | `string` | The model package group name (the AWS identity training pipelines register packages into). |
| `status.outputs.model_package_group_arn` | `string` | The Amazon Resource Name of the model package group. |

## See Also

- [Overview](../README.md)
