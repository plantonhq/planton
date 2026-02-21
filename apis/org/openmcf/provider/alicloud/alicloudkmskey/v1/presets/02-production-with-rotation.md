# Production Encryption Key with Rotation

This preset creates a production-grade KMS key with annual automatic rotation and deletion protection enabled. It is the recommended starting point for any KMS key that protects production data -- RDS TDE, OSS SSE-KMS, ECS disk encryption, or PolarDB encryption.

## When to Use

- Production environments with long-lived encryption keys
- Compliance-driven deployments (PCI-DSS, HIPAA, SOC 2) requiring key rotation
- Any scenario where accidental key deletion would cause irrecoverable data loss
- Master encryption keys referenced by multiple downstream services

## Key Configuration Choices

- **Aliyun_AES_256** (default) -- Standard symmetric encryption.
- **Automatic rotation** enabled with **365-day interval** -- Annual key rotation creates new key material while preserving decryption capability for data encrypted with previous versions. Limits blast radius of a compromised key version.
- **Deletion protection** enabled -- Prevents accidental deletion. Must be explicitly disabled before the key can be deleted. Critical for production.
- **30-day pending window** -- If deletion protection is disabled and deletion is requested, there is a 30-day grace period before permanent destruction.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<key-name>` | Unique key name (e.g., `prod-rds-encryption-key`) | Your naming convention |
| `<your-org>` | Organization identifier | Your OpenMCF org |
| `<alibaba-cloud-region>` | Alibaba Cloud region code | Your deployment region strategy |
| `<key-description>` | Description (e.g., `Production RDS TDE master key`) | Your key inventory |
| `<reason-for-protection>` | Why deletion is protected (e.g., `Protects prod RDS and OSS data`) | Security rationale |
| `<your-team>` | Owning team | Your organizational structure |
| `<compliance-framework>` | Compliance requirement (e.g., `pci-dss`, `hipaa`, `soc2`) | Your compliance needs |

## Related Presets

- **01-standard** -- Simpler configuration without rotation or deletion protection, for dev/test
- **03-asymmetric-signing** -- Use instead for digital signature operations
