---
title: "KMS Key"
description: "KMS Key deployment documentation"
icon: "package"
order: 100
componentName: "gcpkmskey"
---

# GCP KMS Key

Deploys a Cloud KMS cryptographic key within an existing key ring for customer-managed encryption (CMEK), digital signing, asymmetric decryption, or MAC generation. Downstream GCP services — BigQuery, Spanner, GKE, Cloud SQL, GCS, Pub/Sub — reference this key for encryption at rest with keys you control. Keys are permanent GCP resources and cannot be deleted; on destroy, all key versions are destroyed and automatic rotation is disabled.

## What Gets Created

When you deploy a GcpKmsKey resource, Planton provisions:

- **CryptoKey** — a `google_kms_crypto_key` resource within the specified key ring, configured with the requested purpose, rotation policy, version template, and framework labels (`planton-resource`, `planton-resource-name`, `planton-resource-kind`, plus org/env/id when set)

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing KMS key ring** — referenced via `keyRingId` (fully qualified path from a GcpKmsKeyRing resource or a direct string)
- **Cloud KMS API enabled** (`cloudkms.googleapis.com`) on the target project
- **IAM permissions** — `roles/cloudkms.admin` on the key ring or project

## Quick Start

Create a file `kms-key.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpKmsKey
metadata:
  name: my-cmek-key
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpKmsKey.my-cmek-key
spec:
  keyRingId:
    value: "projects/my-project/locations/us-central1/keyRings/prod-encryption"
  keyName: cmek-encrypt-key
  rotationPeriod: "7776000s"
```

Deploy:

```shell
planton apply -f kms-key.yaml
```

This creates a symmetric encryption key with automatic 90-day rotation, suitable for CMEK use across GCP services.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `keyRingId` | `StringValueOrRef` | Fully qualified key ring path (`projects/{p}/locations/{l}/keyRings/{name}`). Can reference GcpKmsKeyRing via `valueFrom`. Immutable after creation. | Required |
| `keyName` | `string` | Name of the key in GCP. Distinct from the Planton `metadata.name`. Immutable after creation. | 1-63 chars; pattern: `^[a-zA-Z0-9_-]{1,63}$` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `purpose` | `string` | `ENCRYPT_DECRYPT` | Cryptographic purpose. Valid values: `ENCRYPT_DECRYPT`, `ASYMMETRIC_SIGN`, `ASYMMETRIC_DECRYPT`, `MAC`, `RAW_ENCRYPT_DECRYPT`. Immutable after creation. |
| `rotationPeriod` | `string` | — | Auto-rotation interval in seconds with suffix `s` (e.g., `7776000s` for 90 days). Minimum `86400s` (24 hours). Only meaningful for `ENCRYPT_DECRYPT` keys. |
| `destroyScheduledDuration` | `string` | `2592000s` (30 days) | How long key versions remain in `DESTROY_SCHEDULED` state before permanent destruction. Minimum `86400s`. Format: seconds with suffix `s`. Immutable after creation. |
| `versionTemplate.algorithm` | `string` | `GOOGLE_SYMMETRIC_ENCRYPTION` | Encryption algorithm for new key versions. Required when `versionTemplate` is specified. Common values: `GOOGLE_SYMMETRIC_ENCRYPTION`, `EC_SIGN_P256_SHA256`, `RSA_SIGN_PSS_2048_SHA256`, `RSA_DECRYPT_OAEP_2048_SHA256`, `HMAC_SHA256`. |
| `versionTemplate.protectionLevel` | `string` | `SOFTWARE` | Protection level for key versions: `SOFTWARE` (standard) or `HSM` (Cloud HSM, FIPS 140-2 Level 3). Immutable after creation. |
| `skipInitialVersionCreation` | `bool` | `false` | When `true`, the key is created without an initial key version. Versions must be created manually afterward. |

## Examples

### HSM-Protected Key for Compliance

A Cloud HSM-backed symmetric key for regulated workloads requiring FIPS 140-2 Level 3 hardware protection:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpKmsKey
metadata:
  name: compliance-cmek
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpKmsKey.compliance-cmek
spec:
  keyRingId:
    value: "projects/my-project/locations/us-central1/keyRings/compliance-keys"
  keyName: hsm-cmek-key
  rotationPeriod: "7776000s"
  destroyScheduledDuration: "2592000s"
  versionTemplate:
    algorithm: GOOGLE_SYMMETRIC_ENCRYPTION
    protectionLevel: HSM
```

### Asymmetric Signing Key

An elliptic curve signing key for verifying build artifacts, container images, or JWTs:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpKmsKey
metadata:
  name: artifact-signer
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpKmsKey.artifact-signer
spec:
  keyRingId:
    value: "projects/my-project/locations/us-central1/keyRings/signing-keys"
  keyName: artifact-signing-key
  purpose: ASYMMETRIC_SIGN
  versionTemplate:
    algorithm: EC_SIGN_P256_SHA256
```

### Using Foreign Key References

Reference a key ring from a GcpKmsKeyRing resource instead of hardcoding the path:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpKmsKey
metadata:
  name: composed-cmek
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpKmsKey.composed-cmek
spec:
  keyRingId:
    valueFrom:
      kind: GcpKmsKeyRing
      name: prod-encryption
      field: status.outputs.key_ring_id
  keyName: composed-cmek-key
  rotationPeriod: "7776000s"
  skipInitialVersionCreation: false
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `key_id` | `string` | Fully qualified crypto key path (`projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{name}`). This is the CMEK reference used by downstream resources for customer-managed encryption. |
| `key_name` | `string` | Short name of the key as it exists in GCP (same as the `keyName` input). |

## Related Components

- [GcpKmsKeyRing](/docs/catalog/gcp/kms-key-ring) — parent container for this key (required)
- [GcpBigQueryDataset](/docs/catalog/gcp/bigquery-dataset) — references `key_id` for dataset-level CMEK encryption
- [GcpCloudSql](/docs/catalog/gcp/cloud-sql) — references `key_id` for database CMEK encryption
- [GcpGcsBucket](/docs/catalog/gcp/gcs-bucket) — references `key_id` for bucket encryption
- [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — references `key_id` for boot disk and secrets encryption
