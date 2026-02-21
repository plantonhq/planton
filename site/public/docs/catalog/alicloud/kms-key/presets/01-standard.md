---
title: "Standard Encryption Key"
description: "This preset creates a KMS key with all defaults: AES-256 symmetric encryption, software-based protection, no automatic rotation, and no deletion protection. This is the simplest configuration,..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "kms-key"
componentTitle: "KMS Key"
provider: "alicloud"
icon: "package"
order: 1
---

# Standard Encryption Key

This preset creates a KMS key with all defaults: AES-256 symmetric encryption, software-based protection, no automatic rotation, and no deletion protection. This is the simplest configuration, suitable for development and staging environments where key lifecycle management is not critical.

## When to Use

- Development and staging environments
- Short-lived encryption keys for testing
- Scenarios where automatic rotation is unnecessary (e.g., ephemeral data)
- Proof-of-concept deployments

## Key Configuration Choices

- **Aliyun_AES_256** (default) -- The strongest symmetric encryption algorithm available in shared KMS. Industry standard for data-at-rest encryption.
- **ENCRYPT/DECRYPT** (default) -- Symmetric encryption/decryption usage.
- **SOFTWARE** (default) -- Key material protected by software cryptographic modules. Sufficient for the vast majority of workloads.
- **No rotation** -- Rotation is disabled by default. Enable for long-lived production keys.
- **No deletion protection** -- Not enabled. For production, use the `02-production-with-rotation` preset instead.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy |
| `<key-description>` | Description of the key's purpose (e.g., `Dev RDS encryption key`) | Your key inventory |
| `<your-team>` | Team or business unit that owns this key | Your organizational structure |

## Related Presets

- **02-production-with-rotation** -- Use instead for production keys with automatic rotation and deletion protection
- **03-asymmetric-signing** -- Use instead for digital signature operations
