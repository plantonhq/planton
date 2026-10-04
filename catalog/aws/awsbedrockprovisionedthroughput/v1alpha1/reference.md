# AwsBedrockProvisionedThroughput

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBedrockProvisionedThroughputSpec defines the desired configuration
for Amazon Bedrock Provisioned Throughput - dedicated, guaranteed model
serving capacity purchased in model units. Provisioned Throughput is the
REQUIRED serving path for fine-tuned custom models and an option for
high-volume foundation-model workloads.

The provisioned model's name is taken from `metadata.name` (1-63
characters).

COST WARNING: this resource bills from the moment it is created -
per-hour without a commitment, or for the full term with a 1-month or
6-month commitment (committed purchases CANNOT be canceled early; the
provider will not even destroy them until the term lapses). Model units
for large models run to thousands of dollars per month - size the
purchase deliberately.

EVERY spec field is create-time-immutable: changing one destroys and
recreates the purchase (a committed purchase blocks that destroy until
its term ends).

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockProvisionedThroughput
metadata:
  name: test-support-model-capacity
  id: test-support-model-capacity
  org: test-org
  env: dev
spec:
  region: us-east-1
  modelArn:
    value: arn:aws:bedrock:us-east-1:123456789012:custom-model/amazon.titan-text-lite-v1/abc123def456
  modelUnits: 1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.modelArn` | `string \| valueFrom` | yes |  | AwsBedrockCustomModel (`status.outputs.custom_model_arn`) |
| `spec.modelUnits` | `int32` |  |  |  |
| `spec.commitmentDuration` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region of the provisioned capacity.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.modelArn

`string | valueFrom` · required

ARN of the model to provision capacity for - typically a custom
model's output ARN (the default reference wiring), or a
foundation-model ARN for models AWS allows provisioning directly.

- references: AwsBedrockCustomModel (`status.outputs.custom_model_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsBedrockCustomModel, name: <that resource's name>, fieldPath: status.outputs.custom_model_arn}} -- a bare string does not parse

### spec.modelUnits

`int32`

Number of model units to purchase (at least 1). A model unit's
throughput (tokens per minute) is model-specific - AWS quotes it per
model in the console/docs. Account quotas for no-commitment model
units default LOW (often 0-2); larger purchases need a quota increase
or a commitment.

- rule: {"int32":{"gte":1}}

### spec.commitmentDuration

`string`

Commitment term: omit for NO commitment (hourly billing, delete any
time), or "OneMonth"/"SixMonths" for committed terms with discounted
rates. Committed purchases bill for the full term and CANNOT be
deleted until the term ends.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["OneMonth","SixMonths"]}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockProvisionedThroughput, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.provisioned_model_arn` | `string` | ARN of the provisioned model - the modelId applications pass to InvokeModel/Converse to consume the dedicated capacity. |
| `status.outputs.provisioned_model_name` | `string` | The provisioned model's name. Matches metadata.name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.modelArn` | AwsBedrockCustomModel | `status.outputs.custom_model_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockAgent | `spec.aliases[].routing.provisionedThroughput` | `status.outputs.provisioned_model_arn` |

## See Also

- [Overview](../README.md)
