# AliCloudKmsKey

Manages an Alibaba Cloud Key Management Service (KMS) customer-managed key (CMK).

## Overview

A KMS key is used to encrypt and decrypt data across Alibaba Cloud services. It serves as the root of trust for encryption at rest -- RDS Transparent Data Encryption, OSS Server-Side Encryption, ECS disk encryption, and PolarDB all delegate key management to KMS. Keys can also perform digital signing and verification when configured with an asymmetric key spec.

### What Gets Created

- **KMS Key** -- a customer-managed key (CMK) in the specified region with configurable algorithm, usage type, and rotation policy

This is a standalone component. Downstream components (RDS, OSS, ECS) reference the key via its `key_id` output.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`, `cn-shanghai`, `us-west-1`) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | `""` | Human-readable description of the key's purpose |
| `keySpec` | string | `"Aliyun_AES_256"` | Cryptographic algorithm (see below) |
| `keyUsage` | string | `"ENCRYPT/DECRYPT"` | Usage type: `ENCRYPT/DECRYPT` or `SIGN/VERIFY` |
| `protectionLevel` | string | `"SOFTWARE"` | Protection level: `SOFTWARE` or `HSM` |
| `automaticRotation` | bool | `false` | Enable automatic key rotation (symmetric keys only) |
| `rotationInterval` | string | `""` | Rotation period (e.g., `365d`, `8760h`). Required when `automaticRotation` is true |
| `pendingWindowInDays` | int32 | `30` | Deletion grace period in days (7-366) |
| `deletionProtection` | bool | `false` | Prevent accidental deletion. Recommended for production |
| `deletionProtectionDescription` | string | `""` | Reason for enabling deletion protection |
| `tags` | map | `{}` | Key-value tags applied to the key |

### Key Spec Values

| Value | Type | Usage |
|-------|------|-------|
| `Aliyun_AES_256` | Symmetric | ENCRYPT/DECRYPT (default) |
| `Aliyun_AES_128` | Symmetric | ENCRYPT/DECRYPT (Dedicated KMS only) |
| `Aliyun_AES_192` | Symmetric | ENCRYPT/DECRYPT (Dedicated KMS only) |
| `Aliyun_SM4` | Symmetric | ENCRYPT/DECRYPT (Chinese national standard) |
| `RSA_2048` | Asymmetric | SIGN/VERIFY |
| `RSA_3072` | Asymmetric | SIGN/VERIFY |
| `EC_P256` | Asymmetric | SIGN/VERIFY (NIST P-256) |
| `EC_P256K` | Asymmetric | SIGN/VERIFY (secp256k1) |
| `EC_SM2` | Asymmetric | SIGN/VERIFY (Chinese national standard) |

### Immutability

The following fields cannot be changed after creation (ForceNew): `keySpec`, `keyUsage`, `protectionLevel`. Changing any of these requires replacing the key.

### Deletion Behavior

KMS keys are never deleted immediately. When a deletion is requested, the key enters `PendingDeletion` state for the configured `pendingWindowInDays`. During this window, the deletion can be cancelled. After the window elapses, the key is permanently destroyed and all data encrypted with it becomes irrecoverable.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `key_id` | The KMS key ID, referenced by downstream components for encryption configuration |
| `arn` | The key ARN (`acs:kms:{region}:{account-id}:key/{key-id}`), used in RAM policies |

## Related Components

- **AliCloudRdsInstance** -- uses a KMS key for Transparent Data Encryption (TDE)
- **AliCloudStorageBucket** -- uses a KMS key for Server-Side Encryption (SSE-KMS)
- **AliCloudEcsInstance** -- uses a KMS key for disk encryption
