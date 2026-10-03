# AWS Secrets Manager Secret

Deploys an AWS Secrets Manager secret — a named, versioned, KMS-encrypted container for credential material with optional automatic rotation and cross-region replication. The secret value is supplied as a managed-secret reference and resolved just-in-time at deploy, so plaintext never lives in the control plane. The secret's AWS name defaults to `metadata.name`, and `spec.secretName` overrides it when the name needs shapes a resource name cannot carry — hierarchical paths like `prod/payments/db` or service-required prefixes like `ecr-pullthroughcache/dockerhub`; either way the name is create-only, while the value, key, policy, replicas, and rotation all update in place.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Secret** — the named, KMS-encrypted container; up to 512 characters of alphanumeric plus `/_+=.@-`, so path-style names work.
- **Secret Version** — created when a value arm (`stringValue` or `binaryValue`) is set, staged `AWSCURRENT` with optional custom staging labels riding alongside. Omitting both arms creates a shell secret an application or rotation function fills.
- **Resource Policy** — created only when `policy` is declared, rendered through the standalone policy resource so `blockPublicPolicy` (default on) rejects policies that grant anonymous access.
- **Rotation Configuration** — created only when `rotation` is declared: exactly one of a self-managed rotation Lambda or a partner-managed external rotation role.
- **AWS Tags** — resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A managed secret holding the value** — create it in Planton's secrets store and reference it as `$secret/<slug>` in `stringValue`; the platform rejects plaintext on sensitive fields.

### AWS Account

- **A KMS key** (only for customer-managed encryption) — required in practice when other AWS accounts must read the secret: the AWS-managed `aws/secretsmanager` key cannot be granted cross-account. Provide the ARN directly or reference an AwsKmsKey Infra Component.
- **A rotation Lambda** (only for self-managed rotation) — the function must implement the four rotation steps and grant Secrets Manager invoke permission (principal `secretsmanager.amazonaws.com`).

## Deploy

### Console

Open the deployment store, find **AWS Secrets Manager Secret**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Application Credentials** preset in the [Presets](#presets) tab for the shape most secrets start as.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSecretsManagerSecret
metadata:
  name: payments-db-credentials
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  secretName: prod/payments/db
  description: Payments database credentials
  stringValue: $secret/payments-db-credentials
```

```shell
planton apply -f secret.yaml
```

This creates a secret named `prod/payments/db` (the hierarchical AWS name lives in `secretName`; omit it to name the secret after the resource) encrypted under the AWS-managed key, its JSON credential document pulled from the org secret at deploy time and staged `AWSCURRENT`. An Infra Job tracks the provisioning in real time.

### InfraChart

When the secret deploys alongside its encryption key and rotation function in one chart, wire both references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  stringValue: $secret/payments-db-credentials
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: app-secrets
      fieldPath: status.outputs.key_arn
  rotation:
    rotationLambdaArn:
      valueFrom:
        kind: AwsLambda
        name: db-rotation-fn
        fieldPath: status.outputs.function_arn
    automaticallyAfterDays: 30
```

The InfraPipeline resolves the dependency graph — key and function first, then the secret encrypted under one and rotated by the other.

## Key Configuration

These are the most important decisions when configuring a secret. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Wire consumers through the ARN output, never a derived ARN** — AWS appends a random 6-character suffix to secret ARNs (`arn:...:secret:name-AbCdEf`), so the ARN is not derivable from the name. Use the `secret_arn` output for IAM policies and injection wiring.

**The recovery window is a name lease** — deletion is soft by default: recoverable for `recoveryWindowInDays` (default 30), during which the NAME stays reserved and a same-named create stalls against AWS's "scheduled for deletion" error. Use `0` for ephemeral and test secrets that must be recreatable immediately — but see the replication caveat below before using `0` on anything replicated.

**Cross-account reads need BOTH sides** — a resource policy statement granting the reader AND a customer-managed KMS key whose key policy grants the reader `kms:Decrypt`. A policy-only grant fails at GetSecretValue with a KMS error that reads like a permissions bug on the wrong service. Keep `blockPublicPolicy` on — a public secret policy is almost always a mistake.

**Replication has a stranding trap** — replicas are read-only regional copies AWS keeps in sync, each encrypted under a key in ITS OWN region. AWS deletes replicas asynchronously after replication is removed, and `recoveryWindowInDays: 0` lets the primary's force-delete outrun that — stranding a live replica that rejects direct deletion ("Operation not permitted on a replica secret"). Keep a non-zero window on replicated secrets; recover a stranded replica with stop-replication-to-replica in its region, then delete.

**`forceOverwriteReplicaSecret` fails silently on stranded ex-replicas** — it overwrites an ordinary same-named secret in the replica region, but NOT one that is itself a stranded ex-replica; replication then fails without failing the apply (neither engine waits on replication status). After enabling replication into a region with name history, verify `ReplicationStatus` reaches `InSync`.

**Rotation rotates immediately by default** — `rotateImmediately` defaults true, so the rotation mechanism runs as soon as it is configured and must be able to read a current value: configuring rotation on a valueless shell secret fails the first rotation. The Lambda must grant `secretsmanager.amazonaws.com` invoke permission — the deploy retries through IAM propagation, but a genuinely missing permission fails after the retry budget. Pick exactly one cadence (`automaticallyAfterDays` or a `scheduleExpression`) and one mechanism (your Lambda, or a partner's external rotation role paired with `type`).

**Staging labels ride with AWSCURRENT** — custom `versionStages` labels attach alongside the automatic `AWSCURRENT` (the module guarantees the concat), supporting blue/green-style consumption. A label lives on one version at a time.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** | `kmsKeyId` (and per-replica `replicaRegions[].kmsKeyId`) | `status.outputs.key_arn` |
| **AwsLambda** | `rotation.rotationLambdaArn` | `status.outputs.function_arn` |
| **AwsIamRole** | `rotation.externalRotationRoleArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `secret_arn` | The secret's ARN (with AWS's random suffix) | IAM policy statements, ECS/Lambda secret injection, cross-service references |
| `secret_name` | The secret's AWS name (`spec.secretName`, else `metadata.name`) | Consumers resolving secrets by name via SDK GetSecretValue calls |

`version_id` is also exposed — the version this deployment manages (the `AWSCURRENT` version when a value arm is set; empty for a shell secret). It is useful for auditing which version a deploy published rather than as a composition input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Application credentials by path** — a JSON key/value document under a hierarchical name (`prod/payments/db`), AWS-managed encryption, a short recovery window. Path-style names keep IAM policies simple (`GetSecretValue` on `arn:...:secret:prod/*`), and this is the shape most secrets start as — rotation and replication bolt on later without replacement. Start from the **Application Credentials** preset.

**Rotated database password** — a customer-managed KMS key plus a rotation Lambda on a 30-day cadence, with the bootstrap value replaced by a rotation-issued credential on first deploy. The dedicated key buys rotation-independent audit and revocation, and it is what makes cross-account reads possible at all. Start from the **Rotated Database Password** preset.

**Multi-region replication** — write to the primary, read locally everywhere by the same name (each region gets its own ARN), with a resource policy restricting reads to the application role. The right posture for DR standbys and regional consumers; keep the recovery window non-zero so teardown can't strand replicas. Start from the **Multi-Region Replicated** preset.

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption for the secret and its replicas, wired via `kmsKeyId`
- [**AWS Lambda**](/infra-catalog/aws-lambda) — the self-managed rotation function, wired via `rotation.rotationLambdaArn`
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the role partner-managed external rotation assumes
- [**AWS SSM Parameter**](/infra-catalog/aws-ssm-parameter) — the lighter alternative for encrypted values that need neither managed rotation nor replication
