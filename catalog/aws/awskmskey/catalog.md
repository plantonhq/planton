# AWS KMS Key

Deploys a customer-managed KMS key with configurable cryptographic shape (key spec and usage), a custom key policy, automatic rotation, multi-Region designation, a scheduled deletion window, aliases, and scoped grants. KMS keys have no name in AWS -- identity is the generated key ID and ARN -- so downstream Infra Components compose with this key by referencing its `key_arn` output, the value encryption-at-rest fields across databases, queues, buckets, and functions all take. The cryptographic shape is create-time immutable: changing `keySpec` or `keyUsage` replaces the key, and old ciphertext stays decryptable only by the old key.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **KMS Key** -- a customer-managed key with the specified key spec and usage, description, key policy, rotation configuration, multi-Region designation, deletion window, and (optionally) a custom key store home
- **KMS Grants** -- created only when `grants` is set; one grant resource per entry, giving a principal scoped, revocable use of the key without key-policy edits
- **KMS Aliases** -- one alias resource per entry in `aliases`; friendly names (e.g., `alias/data-encryption`) that reference the key, added and removed in place as the list changes
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the key

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **KMS permissions** -- the credentials used by the Provider Connection must have `kms:CreateKey`, `kms:CreateAlias`, `kms:EnableKeyRotation`, `kms:PutKeyPolicy`, and related KMS permissions.
- **Key policy** -- leaving `policy` empty keeps the AWS default policy, which grants the account root full access and enables IAM-policy delegation. Provide a custom policy document for cross-account grants or restricted administration -- and always keep an administrator principal in it, or the key can become unmanageable.
- **Region selection** -- KMS keys are regional resources. Create keys in the same region as the resources that will use them (S3 buckets, RDS instances, EKS clusters). For cross-region decryption with the same material, enable `multiRegion` and create replica keys from this primary.

## Deploy

### Console

Open the deployment store, find **AWS KMS Key**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Symmetric Encryption Key** preset in the [Presets](#presets) tab to pre-populate a standard encryption key configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsKmsKey
metadata:
  name: data-encryption-key
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  description: "Customer-managed key for data encryption"
  keySpec: SYMMETRIC_DEFAULT
  enableKeyRotation: true
  deletionWindowDays: 30
  aliases:
    - "alias/acme-data-encryption"
```

```shell
planton apply -f kms-key.yaml
```

This creates a symmetric encryption key with automatic rotation enabled, a 30-day deletion window, and one alias. SYMMETRIC_DEFAULT is suitable for use with S3 SSE-KMS, EBS volume encryption, RDS storage encryption, and EKS secrets encryption. An Infra Job tracks the provisioning in real time.

### InfraChart

When the key deploys alongside the workload role that uses it, wire the grant's principal via ValueFromRef:

```yaml
spec:
  region: us-west-2
  keySpec: SYMMETRIC_DEFAULT
  enableKeyRotation: true
  grants:
    - name: orders-service-encrypt
      granteePrincipal:
        valueFrom:
          kind: AwsIamRole
          name: orders-service-role
          fieldPath: status.outputs.role_arn
      operations:
        - Encrypt
        - Decrypt
        - GenerateDataKey
```

The InfraPipeline resolves the dependency graph, deploys the role first, then creates the key with the grant in place -- "this workload may use this key" wired as a first-class dependency.

## Key Configuration

These are the most important decisions when configuring a KMS key. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Key spec and usage** -- The `keySpec` field defaults to `SYMMETRIC_DEFAULT`, the shape every AWS service integration requires. Choose an RSA (`RSA_2048`/`RSA_3072`/`RSA_4096`) or ECC (`ECC_NIST_P256`/`ECC_NIST_P384`/`ECC_NIST_P521`/`ECC_NIST_EDWARDS25519`/`ECC_SECG_P256K1`) spec for signing or public-key workflows, a post-quantum ML-DSA spec (`ML_DSA_44`/`ML_DSA_65`/`ML_DSA_87`) for quantum-resistant signing, an HMAC spec (`HMAC_224`/`HMAC_256`/`HMAC_384`/`HMAC_512`) for token authentication, or `SM2` in China regions. `keyUsage` is coupled to the spec's family: symmetric keys encrypt/decrypt, HMAC keys generate/verify MACs, the NIST ECC curves sign/verify or derive shared secrets (`KEY_AGREEMENT`, ECDH), Ed25519/secp256k1/ML-DSA keys sign only, and RSA/SM2 keys choose between encryption and signing -- the spec validates the full AWS compatibility matrix before deployment. Both fields are create-time immutable -- changing them replaces the key.

**Grants** -- The `grants` list delegates scoped, revocable key access without editing the key policy: each entry names an IAM principal (reference an `AwsIamRole`'s `role_arn` output, or pass a literal IAM principal ARN -- AWS rejects bare service principals on this parameter; service-principal grants ride a separate API surface the provider does not expose), the allowed operations, optional encryption-context constraints, and whether teardown retires (graceful) or revokes (immediate) the grant. Grants are the chart-native way to wire "this workload may use this key."

**Custom key stores** -- `customKeyStoreId` creates the key in a CloudHSM key store (your hardware) or, with `xksKeyId`, in an external key store whose material lives outside AWS entirely. Custom key store keys are symmetric-only, never rotate automatically, and cannot be multi-Region -- AWS's contract, enforced at validate time.

**Key rotation** -- Set `enableKeyRotation: true` to rotate the material automatically (symmetric keys only). Rotation is transparent to callers: the key ID, ARN, and aliases never change, and old ciphertext keeps decrypting. `rotationPeriodInDays` (90-2560) tunes the cadence; leaving it unset keeps AWS's 365-day default.

**Key policy** -- The `policy` field takes a JSON policy document that is the root of access control on the key. Empty keeps the AWS default policy (account root full access, IAM delegation enabled) -- the right choice for most keys. `bypassPolicyLockoutSafetyCheck` skips AWS's protection against lockout policies; leave it false unless deliberately constructing a lockout.

**Multi-Region** -- `multiRegion: true` creates a multi-Region PRIMARY key whose replicas in other regions share its material, so ciphertext encrypted in one region decrypts in another. Create-time immutable.

**Deletion window** -- The `deletionWindowDays` field (7-30 days, default 30) defines the waiting period before a scheduled key deletion becomes permanent. During this window, the deletion can be cancelled. Use 30 days for production keys to allow time for discovery of dependent resources.

**Aliases** -- Each entry in `aliases` must start with `alias/` (the `alias/aws/` prefix is reserved for AWS-managed keys). Aliases are how humans and SDK callers address the key without its generated ID; re-pointing an alias to a new key is the manual-rotation idiom for specs AWS cannot rotate automatically.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `grants[].granteePrincipal` | `status.outputs.role_arn` |
| **AwsIamRole** | `grants[].retiringPrincipal` | `status.outputs.role_arn` |

Both fields also accept literal IAM principal ARNs -- how users, account roots, and cross-account principals are named.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `key_id` | Generated ID of the KMS key (`mrk-…` for multi-Region keys) | Direct key references in AWS service configurations |
| `key_arn` | Amazon Resource Name of the KMS key | EKS secrets encryption, S3 SSE-KMS, RDS storage encryption, EBS volume encryption |
| `alias_names` | Alias names attached to the key, in spec order | Human-readable key references in application configuration |
| `grant_ids` | AWS-generated grant IDs, keyed by the grant's position in `spec.grants` | Grant retirement/revocation tooling, state import |

The `key_arn` output is the primary value consumed by downstream resources. S3 buckets, EKS clusters, RDS instances, and EBS volumes reference it to enable customer-managed encryption instead of AWS-managed default keys.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Symmetric encryption key** -- A SYMMETRIC_DEFAULT key with rotation enabled and a 30-day deletion window. The standard pattern for encrypting data at rest across AWS services (S3, EBS, RDS, EKS secrets, DynamoDB). Start from the **Symmetric Encryption Key** preset.

**Shared key with grants** -- One key serving several workloads, each granted only the operations it needs (encrypt-only for producers, decrypt for consumers), with encryption-context constraints scoping what each grant covers. Keeps the key policy small and auditable while access stays revocable per workload. Start from the **Shared Key with Grants** preset.

**External key store key** -- A key whose material lives in your own key manager outside AWS, addressed through `customKeyStoreId` and `xksKeyId`. The sovereignty shape for regulated workloads; accept that these keys are symmetric-only, never rotate automatically, and cannot be multi-Region. Start from the **External Key Store Key** preset.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the grantee and retiring principals grants reference by `role_arn`
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- consumes `key_arn` for SSE-KMS default encryption
- [**AWS EKS Cluster**](/infra-catalog/aws-eks-cluster) -- consumes `key_arn` for secrets envelope encryption
- [**AWS RDS Instance**](/infra-catalog/aws-rds-instance) -- consumes `key_arn` for storage encryption
