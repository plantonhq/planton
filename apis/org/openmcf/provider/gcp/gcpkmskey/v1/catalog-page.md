# GcpKmsKey

Provision and manage Cloud KMS cryptographic keys for customer-managed encryption,
digital signing, and message authentication across GCP services.

## Overview

GcpKmsKey creates a Cloud KMS cryptographic key within an existing key ring.
Keys are the foundation of customer-managed encryption (CMEK) in GCP -- they
protect data in BigQuery, Spanner, GKE, CloudSQL, GCS, PubSub, AlloyDB, and
dozens of other GCP services with encryption keys that you control.

Beyond CMEK, keys also support asymmetric signing (for CI/CD artifact verification,
JWT signing), asymmetric decryption, and MAC generation.

## Key Features

- **Multiple key purposes** -- symmetric encryption, asymmetric signing, asymmetric
  decryption, MAC, and raw encryption for small payloads
- **Hardware security module (HSM)** -- optional Cloud HSM protection for FIPS 140-2
  Level 3 compliance
- **Automatic rotation** -- configurable rotation period for symmetric keys creates
  new key versions automatically
- **Cross-resource references** -- use `valueFrom` to reference a key ring from another
  OpenMCF resource, creating dependency-aware provisioning
- **Framework labels** -- automatic labeling with resource kind, organization, and
  environment metadata

## Quick Start

### Standard CMEK Key

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKey
metadata:
  name: my-cmek-key
spec:
  keyRingId:
    value: "projects/my-project/locations/us-central1/keyRings/my-key-ring"
  keyName: cmek-encrypt-key
  rotationPeriod: "7776000s"  # 90 days
```

### HSM-Protected Key

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKey
metadata:
  name: my-hsm-key
spec:
  keyRingId:
    value: "projects/my-project/locations/us-central1/keyRings/compliance-keys"
  keyName: hsm-cmek-key
  rotationPeriod: "7776000s"
  versionTemplate:
    algorithm: GOOGLE_SYMMETRIC_ENCRYPTION
    protectionLevel: HSM
```

### Asymmetric Signing Key

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpKmsKey
metadata:
  name: my-signing-key
spec:
  keyRingId:
    value: "projects/my-project/locations/us-central1/keyRings/signing-keys"
  keyName: artifact-signing-key
  purpose: ASYMMETRIC_SIGN
  versionTemplate:
    algorithm: EC_SIGN_P256_SHA256
```

## Configuration Reference

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `keyRingId` | StringValueOrRef | Yes | — | Fully qualified key ring path |
| `keyName` | string | Yes | — | Key name (`[a-zA-Z0-9_-]{1,63}`) |
| `purpose` | string | No | ENCRYPT_DECRYPT | Cryptographic purpose |
| `rotationPeriod` | string | No | — | Auto-rotation period (min: 86400s) |
| `destroyScheduledDuration` | string | No | 2592000s (30d) | Version destroy delay |
| `versionTemplate.algorithm` | string | Conditional | — | Algorithm (required if versionTemplate set) |
| `versionTemplate.protectionLevel` | string | No | SOFTWARE | SOFTWARE or HSM |
| `skipInitialVersionCreation` | bool | No | false | Create key without initial version |

## Outputs

| Output | Description |
|--------|-------------|
| `key_id` | Fully qualified path: `projects/{p}/locations/{l}/keyRings/{kr}/cryptoKeys/{name}` |
| `key_name` | Short name of the key |

## Important: Permanent Resource

**Keys cannot be deleted from GCP.** When you destroy a GcpKmsKey:

1. All CryptoKeyVersions are scheduled for destruction
2. Automatic rotation is disabled
3. The resource is removed from your IaC state
4. **The key itself remains permanently in the key ring**

Key names cannot be reused. Plan names carefully before creation.

## Use Cases

### Customer-Managed Encryption (CMEK)

The primary use case. Reference the `key_id` output from downstream resources:

```yaml
# BigQuery dataset with CMEK
apiVersion: gcp.openmcf.org/v1
kind: GcpBigQueryDataset
metadata:
  name: analytics-dataset
spec:
  kmsKeyId:
    valueFrom:
      kind: GcpKmsKey
      name: my-cmek-key
      fieldPath: status.outputs.key_id
```

### Compliance (HSM)

For regulated industries requiring hardware-backed key protection, set
`versionTemplate.protectionLevel` to `HSM`. This uses Google Cloud HSM,
which is validated to FIPS 140-2 Level 3.

### Artifact Signing

Use `purpose: ASYMMETRIC_SIGN` with an elliptic curve or RSA algorithm to
sign build artifacts, container images, or JWTs. The public key can be
distributed to verifiers without exposing the private key material.

## Presets

- **[Symmetric Encryption](presets/01-symmetric-encryption.md)** -- Standard CMEK key with 90-day rotation
- **[HSM Symmetric Encryption](presets/02-hsm-symmetric-encryption.md)** -- HSM-protected CMEK for compliance
- **[Asymmetric Signing](presets/03-asymmetric-signing.md)** -- EC P-256 signing key for CI/CD

## Related Components

| Component | Relationship |
|-----------|-------------|
| [GcpKmsKeyRing](../gcpkmskeyring/v1/) | Parent container (required) |
| [GcpBigQueryDataset](../gcpbigquerydataset/v1/) | CMEK consumer |
| [GcpCloudSql](../gcpcloudsql/v1/) | CMEK consumer |
| [GcpGkeCluster](../gcpgkecluster/v1/) | Boot disk encryption consumer |
| [GcpGcsBucket](../gcpgcsbucket/v1/) | Bucket encryption consumer |
| [GcpSpannerInstance](../gcpspannerinstance/v1/) | CMEK consumer |

## Implementation

Both Pulumi (Go) and Terraform modules are provided with full feature parity:

- **Pulumi**: `iac/pulumi/` -- uses `kms.NewCryptoKey` with framework label management
- **Terraform**: `iac/tf/` -- uses `google_kms_crypto_key` with dynamic `version_template` block
