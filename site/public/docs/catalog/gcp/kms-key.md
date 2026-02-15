---
title: "KMS Key"
description: "Deploy GCP Cloud KMS cryptographic keys using OpenMCF"
---

# GcpKmsKey

Provision and manage Cloud KMS cryptographic keys for customer-managed encryption,
digital signing, and message authentication across GCP services.

## Overview

GcpKmsKey creates a Cloud KMS cryptographic key within an existing key ring.
Keys are the foundation of customer-managed encryption (CMEK) in GCP -- they
protect data in BigQuery, Spanner, GKE, CloudSQL, GCS, PubSub, AlloyDB, and
dozens of other GCP services with encryption keys that you control.

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

## Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `keyRingId` | StringValueOrRef | Yes | Fully qualified key ring path |
| `keyName` | string | Yes | Key name (1-63 chars) |
| `purpose` | string | No | ENCRYPT_DECRYPT (default), ASYMMETRIC_SIGN, ASYMMETRIC_DECRYPT, MAC, RAW_ENCRYPT_DECRYPT |
| `rotationPeriod` | string | No | Auto-rotation period (e.g., "7776000s") |
| `destroyScheduledDuration` | string | No | Destroy delay (default: 30 days) |
| `versionTemplate.algorithm` | string | Conditional | Encryption algorithm |
| `versionTemplate.protectionLevel` | string | No | SOFTWARE (default) or HSM |
| `skipInitialVersionCreation` | bool | No | Skip initial key version |

## Outputs

| Output | Description |
|--------|-------------|
| `key_id` | Fully qualified path for CMEK references |
| `key_name` | Short name of the key |

## Important

**Keys cannot be deleted from GCP.** On destroy, key versions are destroyed and rotation
is disabled, but the key itself remains permanently in the key ring.

## Related

- [GcpKmsKeyRing](/docs/catalog/gcp/kms-key-ring) -- Parent container (required)
