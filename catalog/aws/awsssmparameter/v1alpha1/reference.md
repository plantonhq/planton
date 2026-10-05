# AwsSsmParameter

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSsmParameterSpec defines the desired configuration for one AWS
Systems Manager Parameter Store entry: a named configuration value
(String, StringList, or SecureString) applications read at runtime.

The parameter's name is an explicit field rather than metadata.name:
parameter names are hierarchical paths ("/prod/db/url") and slashes
cannot live in metadata.name. Changing the name forces replacement.

The VALUE is set through exactly one of `value` (plain configuration
text - the common String/StringList case, visible in plans and
state) or `secure_value` (a sensitive field: the platform stores it
as a managed-secret reference and resolves it just-in-time at
deploy, so plaintext never lives in the control plane). SecureString
parameters MUST use `secure_value`.

## Example

```yaml
# Canonical AwsSsmParameter example (hack/dev manifest and refgen
# Example source): a hierarchical SecureString sourced from the secret
# arm with an explicit KMS key. Literal values stand in for composed
# references so the offline `tofu plan` renders the secret path.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSsmParameter
metadata:
  name: prod-db-password
  id: prod-db-password
  org: test-org
  env: dev
spec:
  region: us-west-2
  parameterName: /prod/payments/db-password
  type: SecureString
  secureValue: example-secret-value
  description: Payments database password, rotated by the platform
  keyId:
    value: alias/prod-secrets
  tier: Standard
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.parameterName` | `string` | yes |  |  |
| `spec.type` | `string` |  |  |  |
| `spec.value` | `string` |  |  |  |
| `spec.secureValue` | `string` (sensitive) |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.allowedPattern` | `string` |  |  |  |
| `spec.tier` | `string` |  |  |  |
| `spec.keyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.dataType` | `string` |  |  |  |
| `spec.overwrite` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the parameter lives in.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.parameterName

`string` · required

The parameter's name in AWS: letters, digits, underscores,
periods, hyphens, and slashes, up to 2048 characters. Hierarchical
paths ("/prod/db/url") organize parameters and enable by-path
reads; a name containing slashes must start with one. AWS rejects
names beginning with the reserved "aws" or "ssm" prefixes
(case-insensitive) - a server-side rule with no client-side
mirror. Changing the name forces replacement.

- rule: {"string":{"minLen":"1","maxLen":"2048","pattern":"^[a-zA-Z0-9_.\\-/]+$"}}

### spec.type

`string`

The parameter type. String holds one value, StringList a
comma-separated list, SecureString a KMS-encrypted secret.
The type cannot move away from SecureString in place.

- rule: {"string":{"in":["String","StringList","SecureString"]}}

### spec.value

`string`

The parameter value as plain configuration text (for StringList: a
comma-separated list). Visible in engine plans and state - use
secure_value for anything secret. Mutually exclusive with
secure_value; forbidden for SecureString parameters.

### spec.secureValue

`string` · sensitive

The parameter value as a secret. Sensitive: supply a
managed-secret reference; the platform resolves it just-in-time at
deploy, so plaintext never lives in the control plane. Required
for SecureString parameters and legal for any type whose value
should stay out of plan output. Mutually exclusive with value.

### spec.description

`string`

Human description of the parameter, up to 1024 characters.

- rule: {"string":{"maxLen":"1024"}}

### spec.allowedPattern

`string`

A regular expression AWS validates VALUES against on every write
(e.g. "^\\d+$" for numeric-only). Applies to future writes, not
the current value.

- rule: {"string":{"maxLen":"1024"}}

### spec.tier

`string`

The parameter tier. Unset = Standard (free, 4KB values, no
policies). Advanced allows 8KB values and parameter policies but
bills per parameter, and DOWNGRADING Advanced back to Standard
forces replacement (AWS forbids it in place). Intelligent-Tiering
lets AWS pick per write - AWS resolves it to a concrete tier
server-side, so reads report Standard or Advanced, never
Intelligent-Tiering.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Standard","Advanced","Intelligent-Tiering"]}}

### spec.keyId

`string | valueFrom`

KMS key for SecureString encryption (key ID, alias, or ARN).
Unset = the account's AWS-managed aws/ssm key. Only meaningful on
SecureString parameters - AWS ignores it for other types.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.dataType

`string`

The value's data type. Unset = "text". "aws:ec2:image" makes AWS
validate the value as an AMI ID on every write; "aws:ssm:integration"
is for SSM integrations. Changing the data type forces
replacement.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["text","aws:ec2:image","aws:ssm:integration"]}}

### spec.overwrite

`bool`

Overwrite a parameter of the same name that already exists OUTSIDE
this deployment on first create (deploy-side behavior, never an
AWS attribute: unset, the first apply fails on a pre-existing
name; updates of this deployment's own parameter always
overwrite). Never read back from AWS - a freshly imported
parameter always shows it false.

## Validation Rules

- `spec.parameter_name_hierarchy`: a parameter name containing '/' must be fully qualified with a leading '/'
- `spec.value_arms_exactly_one`: exactly one of value and secure_value must be set
- `spec.secure_string_uses_secure_value`: SecureString parameters must carry their value in secure_value, never in value

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSsmParameter, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.parameter_name` | `string` | The parameter's name (also the provider's import ID; ARNs import too). |
| `status.outputs.parameter_arn` | `string` | The parameter's ARN. |
| `status.outputs.version` | `string` | The parameter's version number (increments on every value write). |
| `status.outputs.tier` | `string` | The tier AWS resolved for the parameter (Standard or Advanced - Intelligent-Tiering never persists; AWS resolves it per write). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.keyId` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
