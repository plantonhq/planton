# Examples

## Minimal Configuration

Creates a KMS key with defaults: Aliyun_AES_256, ENCRYPT/DECRYPT, SOFTWARE protection, no rotation. Suitable for development or testing.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudKmsKey
metadata:
  name: my-key
spec:
  region: cn-hangzhou
```

## Production Encryption Key with Rotation

A production-grade encryption key with annual automatic rotation and deletion protection enabled to prevent accidental data loss.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudKmsKey
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

## Asymmetric Signing Key

An RSA-2048 key for digital signature generation and verification. Useful for signing JWTs, certificates, or API payloads.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudKmsKey
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
