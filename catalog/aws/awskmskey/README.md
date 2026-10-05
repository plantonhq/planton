# AwsKmsKey

AWS Key Management Service (KMS) customer-managed keys provide encryption, signing, key agreement, and MAC operations with fine-grained access control. This resource creates a KMS key, optional aliases, and optional grants; consumers reference `status.outputs.key_arn` or an alias name.

## Spec fields (summary)

- `region` — AWS region (required)
- `description` — console description (max 8192 characters)
- `key_spec` — provider string (`SYMMETRIC_DEFAULT` default; RSA, NIST ECC, `ECC_NIST_EDWARDS25519`, `ECC_SECG_P256K1`, post-quantum `ML_DSA_44/65/87`, HMAC, or `SM2`); create-time immutable
- `key_usage` — `ENCRYPT_DECRYPT` (default), `SIGN_VERIFY`, `KEY_AGREEMENT` (NIST ECC + SM2 only), or `GENERATE_VERIFY_MAC`; create-time immutable; the spec validates AWS's full spec/usage compatibility matrix
- `policy` — key policy JSON (optional; empty keeps AWS default)
- `bypass_policy_lockout_safety_check` — skip lockout safety check (dangerous)
- `disabled` — create or pause in disabled state
- `enable_key_rotation` — automatic rotation (SYMMETRIC_DEFAULT only; never in custom key stores)
- `rotation_period_in_days` — 90–2560 (default 365); requires `enable_key_rotation`
- `multi_region` — multi-Region primary key (create-time immutable)
- `deletion_window_days` — 7–30 (default 30)
- `aliases` — repeated `alias/...` strings (not a single `alias_name`)
- `custom_key_store_id` — create the key in a CloudHSM or external key store (literal `cks-...` id; symmetric-only by AWS contract)
- `xks_key_id` — the existing key in the external key manager (external key stores only)
- `grants` — scoped, revocable per-principal key access without key-policy edits: grantee/retiring principals (reference an `AwsIamRole` or pass a literal IAM principal ARN — never a bare service principal, which AWS rejects on this parameter), operations, optional encryption-context constraints, retire-vs-revoke teardown

## Outputs

- `key_id` — generated key ID
- `key_arn` — join key for encryption-at-rest fields
- `alias_names` — attached alias names in spec order
- `grant_ids` — AWS-generated grant IDs keyed by the grant's position in `spec.grants` (the handles `RetireGrant`/`RevokeGrant` and state import take)

## How it works

The Planton CLI validates the manifest, generates IaC inputs, and invokes IaC backends:

- Pulumi (Go modules under `iac/pulumi`)
- Terraform (modules under `iac/tf`)

`metadata.name` drives the Name identity tag. Key shape (`key_spec`, `key_usage`) cannot change in place — a different shape replaces the key.

## References

- [AWS KMS](https://docs.aws.amazon.com/kms/latest/developerguide/overview.html)
- [Key types](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#master_keys)
- [Key rotation](https://docs.aws.amazon.com/kms/latest/developerguide/rotate-keys.html)
- [Catalog page](./catalog.md)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
