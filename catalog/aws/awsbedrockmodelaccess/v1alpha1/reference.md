# AwsBedrockModelAccess

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBedrockModelAccessSpec defines the desired configuration for Amazon
Bedrock model access - accepting the marketplace agreement that entitles
this AWS account (in this region) to invoke a specific foundation model.
This is the declarative form of the console's "Model access" page: one
component instance per model.

Deploying accepts the model's PUBLIC offer (the modules look the offer
token up at deploy time, so the spec needs only the model identifier).
Most offers carry no subscription charge - invocations bill
per-token regardless. Destroying the component CANCELS the agreement,
removing the account's access to the model in that region.

Anthropic model agreements additionally require the account's use-case
form to be on file - supply `use_case_form` (or record it once in the
console) before the agreement will activate.

EVERY spec field is create-time-immutable: changing the model (or the
form) replaces the agreement.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockModelAccess
metadata:
  name: test-command-r-access
  id: test-command-r-access
  org: test-org
  env: dev
spec:
  region: us-west-2
  modelId: cohere.command-r-v1:0
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.modelId` | `string` | yes |  |  |
| `spec.useCaseForm` | `object` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region whose model access this agreement grants (agreements
are regional; the use-case form is account-global).
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.modelId

`string` · required

The foundation model identifier to enable access to, as listed by
Bedrock (ListFoundationModels / the console's Model access page).
Example: "anthropic.claude-3-5-haiku-20241022-v1:0"

- rule: {"string":{"minLen":"1"}}

### spec.useCaseForm

`object`

The account's use-case-for-model-access form, as the JSON object the
Bedrock console submits (keys like companyName, companyWebsite,
intendedUsers, industryOption, useCases). ACCOUNT-GLOBAL and
write-once per account: AWS keeps exactly one form; deleting this
component never removes it (AWS provides no delete), and two
instances with DIFFERENT forms conflict - keep one instance the owner
of the form and omit it everywhere else. Required (once per account)
for Anthropic models; other vendors' models need no form.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockModelAccess, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.model_id` | `string` | The foundation model identifier the agreement covers. Matches spec.model_id - exported so charts can order model-consuming components (agents, provisioned throughput) after access is granted. |

## See Also

- [Overview](../README.md)
