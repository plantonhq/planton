# Alibaba Cloud KMS Key

Deploys an Alibaba Cloud Key Management Service (KMS) customer-managed key (CMK). The component provisions a cryptographic key that can be used for data encryption across Alibaba Cloud services (RDS, OSS, ECS, PolarDB) or for digital signing and verification with asymmetric key types.

## What Gets Created

When you deploy an AlicloudKmsKey resource, OpenMCF provisions:

- **KMS Key** -- an `alicloud_kms_key` resource in the specified region with configurable algorithm, rotation policy, and deletion protection

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config

## Quick Start

Create a file `kms-key.yaml`:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKmsKey
metadata:
  name: my-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AlicloudKmsKey.my-key
spec:
  region: cn-hangzhou
  description: Encryption key for development resources
```

Deploy:

```shell
openmcf apply -f kms-key.yaml
```

This creates an AES-256 symmetric encryption key with software-based protection.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region for the key (e.g., `cn-hangzhou`, `us-west-1`). | Required; non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable key description. |
| `keySpec` | `string` | `"Aliyun_AES_256"` | Cryptographic algorithm. Symmetric: `Aliyun_AES_256`, `Aliyun_AES_128`, `Aliyun_AES_192`, `Aliyun_SM4`. Asymmetric: `RSA_2048`, `RSA_3072`, `EC_P256`, `EC_P256K`, `EC_SM2`. Immutable after creation. |
| `keyUsage` | `string` | `"ENCRYPT/DECRYPT"` | Usage type. `ENCRYPT/DECRYPT` for symmetric encryption. `SIGN/VERIFY` for asymmetric signing. Immutable after creation. |
| `protectionLevel` | `string` | `"SOFTWARE"` | Protection level. `SOFTWARE` or `HSM`. Immutable after creation. |
| `automaticRotation` | `bool` | `false` | Enable automatic key rotation. Only for symmetric keys. |
| `rotationInterval` | `string` | `""` | Rotation period (e.g., `365d`). Required when `automaticRotation` is true. |
| `pendingWindowInDays` | `int32` | `30` | Deletion grace period in days (7-366). |
| `deletionProtection` | `bool` | `false` | Prevent accidental key deletion. Recommended for production. |
| `deletionProtectionDescription` | `string` | `""` | Reason for deletion protection. |
| `tags` | `map<string, string>` | `{}` | Tags applied to the key. Merged with standard OpenMCF tags. |

## Examples

### Minimal KMS Key

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKmsKey
metadata:
  name: my-key
spec:
  region: cn-hangzhou
```

### Production Encryption Key with Rotation

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKmsKey
metadata:
  name: prod-encryption-key
  org: my-org
  env: production
spec:
  region: cn-shanghai
  description: Production master encryption key for RDS TDE and OSS SSE
  keySpec: Aliyun_AES_256
  automaticRotation: true
  rotationInterval: "365d"
  deletionProtection: true
  deletionProtectionDescription: Protects production database and storage encryption keys
  pendingWindowInDays: 30
  tags:
    team: security
    compliance: pci-dss
```

### Asymmetric Signing Key

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKmsKey
metadata:
  name: signing-key
spec:
  region: cn-hangzhou
  description: RSA signing key for API payload verification
  keySpec: RSA_2048
  keyUsage: SIGN/VERIFY
  tags:
    purpose: signing
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `key_id` | `string` | The KMS key ID assigned by Alibaba Cloud |
| `arn` | `string` | The key ARN for use in RAM policies |

## Related Components

- [AlicloudRdsInstance](/docs/catalog/alicloud/alicloudrdsinsstance) -- uses a KMS key for Transparent Data Encryption
- [AlicloudOssBucket](/docs/catalog/alicloud/alicloudossbucket) -- uses a KMS key for Server-Side Encryption
- [AlicloudEcsInstance](/docs/catalog/alicloud/alicloudecsinstance) -- uses a KMS key for disk encryption
