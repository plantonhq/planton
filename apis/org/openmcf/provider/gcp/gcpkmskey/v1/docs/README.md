# GcpKmsKey — Research & Design Documentation

## Architecture Decision: Key vs CryptoKey Naming

The GCP KMS API uses the term "CryptoKey" (`projects.locations.keyRings.cryptoKeys`),
and the Terraform resource is `google_kms_crypto_key`. OpenMCF uses the shortened
name **GcpKmsKey** because:

1. "Crypto" is redundant in the KMS (Key Management Service) context
2. Users naturally say "KMS key" in conversation
3. OpenMCF already abbreviates (GcpGcsBucket, GcpGkeCluster)
4. `kind: GcpKmsKey` is cleaner in YAML manifests

The Terraform and Pulumi implementations still use their native resource names
(`google_kms_crypto_key` and `kms.CryptoKey`).

## GCP KMS Key Hierarchy

```
GCP Project
  └── KMS Key Ring (permanent container, scoped to a location)
        └── Crypto Key (permanent key with immutable purpose)
              └── Crypto Key Version (actual key material, rotatable)
```

OpenMCF models the first two levels: GcpKmsKeyRing and GcpKmsKey. Key versions
are operational concerns managed by rotation policies or the GCP console, not IaC.

## Field Analysis

### Immutable Fields (ForceNew)

These fields cannot be changed after creation. Any change destroys and recreates
the resource (which creates a new key, since GCP keys are permanent):

- `key_ring_id` -- the parent key ring
- `key_name` -- the GCP resource name
- `purpose` -- determines supported operations
- `destroy_scheduled_duration` -- version destruction delay
- `version_template.protection_level` -- SOFTWARE vs HSM

### Mutable Fields

- `rotation_period` -- can be added, changed, or removed
- `version_template.algorithm` -- can be updated for new versions
- `labels` -- managed by OpenMCF framework

### Labels Support

Unlike GcpKmsKeyRing (which does not support GCP labels), CryptoKeys do support
labels. The OpenMCF framework applies standard labels automatically:

- `openmcf-resource: true`
- `openmcf-resource-name: <key_name>`
- `openmcf-resource-kind: gcpkmskey`
- `openmcf-organization: <metadata.org>` (if set)
- `openmcf-environment: <metadata.env>` (if set)
- `openmcf-resource-id: <metadata.id>` (if set)

## Purpose and Algorithm Compatibility

The `purpose` field determines which algorithms are valid in `version_template`:

| Purpose | Valid Algorithms |
|---------|-----------------|
| ENCRYPT_DECRYPT | GOOGLE_SYMMETRIC_ENCRYPTION (default), EXTERNAL_SYMMETRIC_ENCRYPTION |
| ASYMMETRIC_SIGN | EC_SIGN_P256_SHA256, EC_SIGN_P384_SHA384, EC_SIGN_SECP256K1_SHA256, RSA_SIGN_PSS_*, RSA_SIGN_PKCS1_* |
| ASYMMETRIC_DECRYPT | RSA_DECRYPT_OAEP_2048_SHA256, RSA_DECRYPT_OAEP_3072_SHA256, RSA_DECRYPT_OAEP_4096_SHA256, etc. |
| MAC | HMAC_SHA1, HMAC_SHA224, HMAC_SHA256, HMAC_SHA384, HMAC_SHA512 |
| RAW_ENCRYPT_DECRYPT | AES_128_GCM, AES_256_GCM, AES_128_CBC, AES_256_CBC, AES_128_CTR, AES_256_CTR |

The full matrix is complex (20+ valid combinations), so OpenMCF delegates
purpose-to-algorithm validation to the GCP API, which returns clear error messages.

## Deletion Behavior

**Keys cannot be deleted from GCP.** This is a fundamental property of the KMS
service, designed to prevent accidental loss of encryption capability.

When a GcpKmsKey is destroyed (via OpenMCF, Terraform, or Pulumi):

1. All CryptoKeyVersions are scheduled for destruction
2. Automatic rotation is disabled
3. The key is removed from the IaC state
4. **The key remains in GCP** as a permanent resource in the key ring

This means:
- Key names cannot be reused in the same key ring
- Plan key names carefully before creation
- Consider using naming conventions that include purpose and date

## Rotation Strategy

### Symmetric Keys (ENCRYPT_DECRYPT)

Set `rotation_period` to automatically create new key versions:

- **Recommended**: 90 days (`7776000s`) for most use cases
- **Minimum**: 1 day (`86400s`)
- **Maximum**: No maximum, but longer periods increase exposure risk

When rotation occurs:
1. A new CryptoKeyVersion is created
2. The new version becomes the primary (used for new encrypt operations)
3. Previous versions remain active (can still decrypt data encrypted with them)

### Asymmetric Keys

Asymmetric keys do NOT support automatic rotation. Version management is manual:
1. Create a new key version in the GCP console or via API
2. Update consumers to use the new version's public key
3. Optionally disable the old version

## Infra-Chart Composability

GcpKmsKey is a **Layer 1-2** resource in infra chart topology:

```
Layer 0: GcpProject
Layer 0-1: GcpKmsKeyRing (references Project)
Layer 1-2: GcpKmsKey (references KeyRing) ← this resource
Layer 2+: BigQuery, Spanner, GKE, CloudSQL, etc. (reference Key for CMEK)
```

The `key_id` output is the most-referenced output in the CMEK ecosystem. It
provides the fully qualified path that all downstream CMEK consumers expect.

## Deliberate Exclusions

These Terraform/Pulumi fields are deliberately excluded from the OpenMCF spec:

| Field | Reason |
|-------|--------|
| `import_only` | BYOK via import jobs -- requires infrastructure we don't model |
| `crypto_key_backend` | EXTERNAL_VPC -- requires EKM connections we don't model |
| `key_access_justifications_policy` | Beta feature, subject to change |

These can be added in future versions if demand materializes.

## Best Practices

1. **Use descriptive key names** -- keys are permanent, so names should convey purpose
2. **Set rotation for symmetric keys** -- 90 days is a reasonable default
3. **Use HSM only when required** -- significantly more expensive than SOFTWARE
4. **Separate key rings by environment** -- prod keys in prod key ring, dev in dev
5. **Use CMEK for sensitive data** -- BigQuery, Spanner, CloudSQL with PII/financial data
6. **Document key purposes** -- use metadata labels and naming conventions
