# GcpKmsKey

A GcpKmsKey provisions a Cloud KMS cryptographic key within an existing key ring.
Keys perform the actual cryptographic operations -- symmetric encryption/decryption
(CMEK), asymmetric signing, asymmetric decryption, or MAC generation.

## When to Use

Use GcpKmsKey when you need:

- **Customer-managed encryption keys (CMEK)** for BigQuery, Spanner, GKE, CloudSQL,
  GCS, PubSub, AlloyDB, or any GCP service that supports encryption with your own keys
- **Asymmetric signing keys** for CI/CD artifact signing, JWT signing, or code
  integrity verification
- **HSM-protected keys** to meet FIPS 140-2 Level 3 compliance requirements
- **Automatic key rotation** for symmetric encryption keys

## Prerequisites

- A GCP project with the Cloud KMS API enabled
- An existing KMS key ring (see [GcpKmsKeyRing](../gcpkmskeyring/v1/))
- Appropriate IAM permissions (`roles/cloudkms.admin` or `roles/cloudkms.cryptoKeyEncrypterDecrypter`)

## Quick Start

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

This creates a symmetric encryption key with 90-day automatic rotation -- the most
common configuration for CMEK.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `keyRingId` | StringValueOrRef | Yes | Fully qualified key ring path |
| `keyName` | string | Yes | Key name (1-63 chars: `[a-zA-Z0-9_-]`) |
| `purpose` | string | No | Key purpose (default: `ENCRYPT_DECRYPT`) |
| `rotationPeriod` | string | No | Auto-rotation period (e.g., `"7776000s"`) |
| `destroyScheduledDuration` | string | No | Destroy delay (default: 30 days) |
| `versionTemplate` | object | No | Algorithm and protection level |
| `skipInitialVersionCreation` | bool | No | Skip initial key version |

### Purpose Values

| Value | Description |
|-------|-------------|
| `ENCRYPT_DECRYPT` | Symmetric encryption for CMEK (default) |
| `ASYMMETRIC_SIGN` | Digital signatures |
| `ASYMMETRIC_DECRYPT` | Asymmetric decryption |
| `MAC` | Message authentication codes |
| `RAW_ENCRYPT_DECRYPT` | Authenticated encryption for small payloads |

### Version Template

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `algorithm` | string | Yes (within template) | Encryption algorithm |
| `protectionLevel` | string | No | `SOFTWARE` (default) or `HSM` |

## Important Notes

**Keys cannot be deleted from GCP.** When you destroy a GcpKmsKey resource,
all CryptoKeyVersions are destroyed and automatic rotation is disabled, but the
key itself remains permanently in the key ring. Plan key names carefully.

**Most fields are immutable.** The following cannot be changed after creation:
key name, key ring, purpose, destroy scheduled duration, and protection level.
Only rotation period and algorithm can be updated.

## Related Components

- [GcpKmsKeyRing](../gcpkmskeyring/v1/) -- Parent key ring (required dependency)
- [GcpBigQueryDataset](../gcpbigquerydataset/v1/) -- Uses `key_id` for CMEK
- [GcpCloudSql](../gcpcloudsql/v1/) -- Uses `key_id` for database encryption
- [GcpGkeCluster](../gcpgkecluster/v1/) -- Uses `key_id` for boot disk encryption
- [GcpGcsBucket](../gcpgcsbucket/v1/) -- Uses `key_id` for bucket encryption
