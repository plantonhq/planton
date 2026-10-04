# AwsBedrockInferenceProfile

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBedrockInferenceProfileSpec defines the desired configuration for an
Amazon Bedrock APPLICATION inference profile - a named handle over a
foundation model (or an AWS system-defined cross-region profile) that
carries its own ARN and tags, so per-application/per-tenant model usage
can be tracked, cost-allocated, and IAM-scoped independently. The
profile itself is free; invocations through it bill at the underlying
model's rates.

The profile's name is taken from `metadata.name` (1-64 characters).

EVERY spec field is create-time-immutable (AWS provides no update for
application inference profiles beyond tags) - changing one destroys and
recreates the profile, which changes its ARN; consumers pinning the ARN
must be updated.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockInferenceProfile
metadata:
  name: test-checkout-nova
  id: test-checkout-nova
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Cost tracking for the checkout service
  sourceArn: arn:aws:bedrock:us-west-2:123456789012:inference-profile/us.amazon.nova-micro-v1:0
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.sourceArn` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the inference profile is created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description (1-200 characters when set).

- rule: {"string":{"maxLen":"200"}}

### spec.sourceArn

`string`

ARN of the model source this profile routes to: a foundation model
(arn:...::foundation-model/<model-id>) for single-region use, or an
AWS system-defined inference profile
(arn:...:inference-profile/<geo>.<model-id>, e.g.
"us.amazon.nova-micro-v1:0") to inherit cross-region routing. AWS
never echoes the source back, so the modules pin it in state - it
shows as unchanged as long as the spec value stays the same.
The foundation-model form works ONLY for models that support
ON_DEMAND inference (live-verified 2026-08-13: CreateInferenceProfile
rejects others with "The provided foundation model does not support
On Demand inference"). Models whose inferenceTypesSupported is
INFERENCE_PROFILE only - the Nova family and most 2025+ releases -
must be referenced through their system-defined inference-profile ARN
form instead. Check with:
`aws bedrock list-foundation-models --by-inference-type ON_DEMAND`.

- rule: {"string":{"pattern":"^arn:aws[a-z-]*:bedrock:[a-z0-9-]*:[0-9]*:(foundation-model|inference-profile)/.+$"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockInferenceProfile, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.inference_profile_arn` | `string` | ARN of the inference profile - the modelId applications pass to InvokeModel/Converse and the resource IAM policies scope to. |
| `status.outputs.inference_profile_id` | `string` | The inference profile's unique identifier. |
| `status.outputs.status` | `string` | Profile status as reported by AWS (ACTIVE when usable). |
| `status.outputs.type` | `string` | Profile type as reported by AWS - always APPLICATION for profiles this component creates (SYSTEM_DEFINED profiles are AWS-owned). |

## See Also

- [Overview](../README.md)
