# AWS SSM Parameter

Deploys one AWS Systems Manager Parameter Store entry: a named configuration value applications read at runtime. Plain config lives as a String, lists as a StringList, and secrets as a KMS-encrypted SecureString whose value is supplied as a managed-secret reference and resolved just-in-time at deploy — plaintext never lives in the control plane or plan output. Hierarchical names like `/prod/db/url` organize configuration and enable by-path reads.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SSM Parameter** — the named value with its type, tier, optional description, and optional write-validation pattern. The name is an explicit spec field (slashes cannot live in `metadata.name`), and changing it forces replacement.
- **KMS Encryption Binding** — configured only when `keyId` is set on a SecureString parameter; without it, SecureString encrypts under the account's AWS-managed `aws/ssm` key.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with permissions to manage SSM parameters (plus KMS permissions when using a customer-managed key). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **An org secret** (only for `secureValue`) — the managed secret holding the parameter's value, referenced as `$secret/<slug>` in the manifest.

### AWS Account

- Nothing — parameters stand alone. For SecureString under your own key, a KMS key the deployment role can use.

## Deploy

### Console

Open the deployment store, find **AWS SSM Parameter**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields covering the name, type, and value arms. Start from the **Plain App Config** preset in the [Presets](#presets) tab for readable configuration, or the **Secure Secret** preset for a KMS-encrypted SecureString.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSsmParameter
metadata:
  name: orders-db-password
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  parameterName: /prod/orders/db-password
  type: SecureString
  secureValue: $secret/orders-db-password
  description: Orders database password, resolved just-in-time at deploy
  keyId:
    valueFrom:
      kind: AwsKmsKey
      name: app-secrets
      fieldPath: status.outputs.key_arn
```

```shell
planton apply -f ssm-parameter.yaml
```

This creates a SecureString at `/prod/orders/db-password` encrypted under the referenced KMS key, its value pulled from the org secret at deploy time. An Infra Job tracks the provisioning in real time.

### InfraChart

When the parameter deploys alongside its encryption key in one chart, wire the key reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  parameterName: /prod/orders/db-password
  type: SecureString
  secureValue: $secret/orders-db-password
  keyId:
    valueFrom:
      kind: AwsKmsKey
      name: app-secrets
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the KMS key first, then creates the parameter encrypted under it.

## Key Configuration

These are the most important decisions when configuring a parameter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Name shape is a contract** — A name containing `/` must be fully qualified with a leading slash (`/prod/db/url`, never `prod/db/url` — validated before AWS sees it), and AWS rejects names beginning with the reserved `aws` or `ssm` prefixes server-side. Pick the hierarchy deliberately: by-path reads (`GetParametersByPath`) make `/prod/orders/*` one call, and renaming later is a replacement, not an edit.

**Two value arms with different visibility** — `value` is plain configuration text, deliberately readable in plans and state; `secureValue` is a sensitive field carrying a `$secret/<slug>` reference resolved just-in-time at deploy. Exactly one is set, and SecureString parameters must use `secureValue` — the spec rejects the plain arm for them, so a secret cannot silently leak into plan output. `secureValue` is also legal on String parameters whose value merely shouldn't appear in plans.

**Tier is the cost lever and a one-way door** — Standard covers values up to 4KB with no per-parameter billing; Advanced unlocks 8KB values and parameter policies but bills per parameter, and downgrading Advanced back to Standard forces replacement — AWS forbids it in place. Intelligent-Tiering never persists: AWS resolves it to Standard or Advanced per write, and the `tier` output reports the resolved answer.

**`allowedPattern` guards future writes only** — AWS validates every subsequent write against the regex; it never validates the value already stored. It is a guardrail against bad writes from any writer, not just this deployment.

**`dataType: aws:ec2:image` turns the parameter into a validated AMI pointer** — AWS verifies the value is a real AMI ID in your account and region on every write, so a wrong-region AMI fails the write rather than the launch that reads it. Changing the data type forces replacement.

**StringList has no escaping** — it is one comma-separated string. A value that itself contains commas needs String, not StringList.

**`overwrite` is deploy-side adoption, not an AWS attribute** — Unset, the first apply fails if a parameter of the same name already exists outside this deployment; set true, the first create adopts and overwrites it. Updates to this deployment's own parameter always overwrite regardless.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** | `keyId` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `parameter_name` | The parameter's name (also the provider's import ID) | Application configuration naming what to read via GetParameter; SSM document parameters |
| `parameter_arn` | The parameter's ARN | IAM policy statements granting read access to specific parameters or paths |

`version` and `tier` are also exposed — the version number that increments on every value write, and the tier AWS resolved. Both are observability echoes for auditing what deployed rather than composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Path-organized app config** — String parameters under one hierarchy (`/prod/orders/log-level`, `/prod/orders/feature-x`) so a service reads its whole subtree in one by-path call, with `allowedPattern` rejecting invalid writes at the AWS API. Values stay readable in plans — the point for non-secret config. Start from the **Plain App Config** preset.

**Secrets under your own key** — SecureString with `secureValue` and a customer-managed KMS key: readers need both SSM and KMS access, giving two independent revocation levers. Rotation is the writer's job — every write publishes a new version and readers pick it up on the next fetch; when rotation must be automated and audited, that is the neighboring Secrets Manager component's territory. Start from the **Secure Secret** preset.

**Validated AMI pointer** — a `dataType: aws:ec2:image` parameter as the handoff point between an image-baking pipeline and the infrastructure that launches from it: the pipeline writes the new AMI ID, AWS validates it exists, and launch templates read the parameter instead of hard-coding AMIs.

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — encrypts SecureString values, wired via the `keyId` reference
- [**AWS Secrets Manager Secret**](/infra-catalog/aws-secrets-manager-secret) — the alternative store when a secret needs managed rotation or cross-region replication rather than a plain encrypted value
