# AWS Bedrock AgentCore Token Vault

Sets the encryption posture of a region's AgentCore token vault — the store AgentCore Identity keeps OAuth tokens and API keys in. This is a settings singleton: AWS provisions exactly one default vault per account and region, and this component decides whether it encrypts under your KMS key (`CustomerManagedKey`) or AWS's (`ServiceManagedKey`). Deploy at most one instance per region — two instances targeting the same vault fight over one setting.

## What Gets Created

This component creates nothing at AWS. The vault already exists — AWS provisions one default token vault per account and region — and the IaC module adopts that existing account object and configures it:

- **Token Vault Key Setting** — the vault's key ownership: your symmetric, same-region KMS key under `CustomerManagedKey`, or AWS's owned-and-rotated key under `ServiceManagedKey`. When a customer-managed key is applied, AWS creates its own grants on the key and re-encrypts stored credentials under it

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AgentCore and KMS permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- A symmetric KMS key in the same region, wired as `kmsKeyArn` (only for `keyType: CustomerManagedKey` — cross-region and asymmetric keys are rejected)

## Deploy

### Console

Open the deployment store, find **AWS Bedrock AgentCore Token Vault**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the key ownership choice. Start from the **Customer-Managed Key** preset in the [Presets](#presets) tab to put the vault under your key, or the **Service-Managed Key (Revert)** preset to return it to AWS-managed encryption.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockAgentCoreTokenVault
metadata:
  name: agentcore-vault-uswest2
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  keyType: CustomerManagedKey
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: agentcore-vault-key
      fieldPath: status.outputs.key_arn
```

```shell
planton apply -f agentcore-token-vault.yaml
```

This points the region's default token vault at your KMS key — every credential AgentCore Identity stores in the region is encrypted under it from this apply onward. An Infra Job tracks the provisioning in real time.

### InfraChart

When the vault setting deploys alongside the key it adopts in one chart, wire the key reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  keyType: CustomerManagedKey
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: agentcore-vault-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, creates the KMS key first, then applies the vault setting onto it.

## Key Configuration

These are the most important decisions when configuring the token vault. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Destroy does not revert the setting** — the provider's delete is a no-op at AWS, so the vault keeps the last-applied key forever. The revert is an apply with `keyType: ServiceManagedKey`. Concretely: never schedule the KMS key's deletion while the vault still points at it, or every stored agent credential becomes unreadable.

**The revert itself needs the old key alive** — AWS validates the previous key's state on every vault write. A vault pointing at a disabled or pending-deletion key refuses even the revert to `ServiceManagedKey` ("Old KMS Key validation failed ... expected KeyState:ENABLED"). Always revert first, retire the key second.

**Wedged anyway? The deletion window is the way out** — while the key is still in its 7–30 day deletion window: cancel the key deletion, re-enable the key, apply `ServiceManagedKey`, and only then re-schedule the deletion. Once the key is truly gone, only AWS Support can help.

**Key revocation is an agent outage** — AgentCore reads the vault on every credential fetch. Revoking the key's grants or disabling the key locks every agent in the region out of its OAuth tokens and API keys at once. Treat key policy changes as outage-grade changes.

**The key pairing is enforced at manifest time** — `CustomerManagedKey` requires `kmsKeyArn`; `ServiceManagedKey` forbids it. AWS would accept a stray ARN silently; the manifest validation refuses it.

**Adopt the key before storing credentials** — create the key, apply this component, then start storing credentials. AWS re-encrypts existing credentials on the switch, but switching before first use keeps the audit story clean.

**tokenVaultId defaults to the only vault that exists** — every account has exactly one vault per region, named "default", and the modules send that when the field is unset. The field exists for AWS's forward surface; manifests never need to set it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** | `kmsKeyArn` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `kms_key_arn` | The customer-managed key ARN in effect (empty under `ServiceManagedKey`) | Charts aligning other AgentCore components' encryption onto the same key |

`token_vault_id` and `key_type` are also exported; they echo the applied configuration back for audit and import purposes rather than feeding composition.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Bring your own key** — a symmetric, same-region KMS key you control for rotation, policy, and revocation, applied before the region's agents start storing credentials. This is the compliance posture: your key policy is the access boundary around every vaulted agent credential. Start from the **Customer-Managed Key** preset.

**Deliberate revert before key retirement** — an apply with `ServiceManagedKey` that hands encryption back to AWS, run before disabling or scheduling deletion of the old key. Because destroy never reverts, this apply is a mandatory step in any key-retirement runbook, not an optional cleanup. Start from the **Service-Managed Key (Revert)** preset.

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — the customer-managed key the vault encrypts under
- [**AWS Bedrock AgentCore Identity**](/infra-catalog/aws-bedrock-agent-core-identity) — the credential providers whose vaulted secrets this component's key setting protects
