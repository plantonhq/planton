# AwsBedrockAgentCoreTokenVault

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBedrockAgentCoreTokenVaultSpec defines the encryption posture of
an AgentCore token vault - the store AgentCore Identity keeps OAuth
tokens and API keys in.

This is a SETTINGS SINGLETON: AWS provisions ONE default token
vault per account+region and this component sets which KMS key
encrypts it. Deploy at most one instance per vault (in practice:
per region); two instances targeting the same vault fight over one
setting. metadata.name never reaches AWS - it is Planton-side
identity only.

Destroying this component is a NO-OP at AWS (the provider has no
delete API for the vault's key setting): whatever key type was last
applied REMAINS in effect after destroy. To return to AWS-managed
encryption, apply an instance with key_type "ServiceManagedKey"
BEFORE destroying - destroy alone does not revert it.

## Example

```yaml
# Canonical AwsBedrockAgentCoreTokenVault example (hack/dev manifest
# and refgen Example source): the region's default token vault
# encrypted with a customer-managed KMS key. Literal ARN stands in for
# the composed AwsKmsKey reference so the offline `tofu plan` renders
# the resource.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreTokenVault
metadata:
  name: agentcore-vault-us-west-2
  id: agentcore-vault-us-west-2
  org: test-org
  env: dev
spec:
  region: us-west-2
  keyType: CustomerManagedKey
  kmsKeyArn:
    value: arn:aws:kms:us-west-2:123456789012:key/11111111-2222-3333-4444-555555555555
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.tokenVaultId` | `string` |  |  |  |
| `spec.keyType` | `string` | yes |  |  |
| `spec.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region whose default token vault this instance manages.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.tokenVaultId

`string`

The token vault to configure. AWS provisions exactly one vault
per account+region named "default"; leave unset to target it
(the modules send "default"). Changing this value replaces the
configuration onto the other vault. This is also the provider's
import ID.

- rule: {"string":{"maxLen":"128"}}

### spec.keyType

`string` · required

Who owns the vault's encryption key:
  - "CustomerManagedKey": encrypt with YOUR KMS key (named in
    kms_key_arn). You control rotation, policy, and revocation -
    revoking the key's grant locks AgentCore out of every stored
    credential.
  - "ServiceManagedKey": AWS owns and rotates the key (the
    default posture every new account starts in).

- rule: {"required":true,"string":{"in":["CustomerManagedKey","ServiceManagedKey"]}}

### spec.kmsKeyArn

`string | valueFrom`

The customer-managed KMS key that encrypts the vault. Required
with key_type "CustomerManagedKey", forbidden with
"ServiceManagedKey". The key must be symmetric and live in the
same region; AgentCore's grants on it are created by AWS when the
setting is applied. KEEP THIS KEY ENABLED until after you revert:
AWS validates the PREVIOUS key's state on every vault write, so
once this key is disabled or scheduled for deletion the vault
cannot even be reverted to ServiceManagedKey ("Old KMS Key
validation failed ... expected KeyState:ENABLED" - live-caught).
Revert first, retire the key second.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

## Validation Rules

- `spec.customer_key_requires_arn`: kms_key_arn is required when key_type is CustomerManagedKey - name the KMS key the vault should use
- `spec.service_key_forbids_arn`: kms_key_arn must be unset when key_type is ServiceManagedKey - AWS owns and rotates the service-managed key

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockAgentCoreTokenVault, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.token_vault_id` | `string` | The vault the setting was applied to ("default" unless the spec targeted another vault) - also the provider's import ID. |
| `status.outputs.key_type` | `string` | The key ownership in effect: "CustomerManagedKey" or "ServiceManagedKey". |
| `status.outputs.kms_key_arn` | `string` | The customer-managed KMS key ARN in effect (empty under ServiceManagedKey). Charts read this to align other AgentCore components onto the same key. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
