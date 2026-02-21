---
title: "AES-256 HSM Key with Auto-Rotation"
description: "This preset creates an AES-256 symmetric encryption key stored in an HSM (FIPS 140-2 Level 3) with automatic 90-day key rotation. This is the standard key for encrypting data at rest across OCI..."
type: "preset"
rank: "01"
presetSlug: "01-aes-256-hsm-auto-rotation"
componentSlug: "kms-key"
componentTitle: "KMS Key"
provider: "oci"
icon: "package"
order: 1
---

# AES-256 HSM Key with Auto-Rotation

This preset creates an AES-256 symmetric encryption key stored in an HSM (FIPS 140-2 Level 3) with automatic 90-day key rotation. This is the standard key for encrypting data at rest across OCI services -- Block Volume, Object Storage, Database, File Storage, and Streaming all accept a KMS key OCID for customer-managed encryption.

## When to Use

- Encrypting Block Volumes, Object Storage buckets, databases, or file systems with customer-managed keys
- Meeting compliance requirements that mandate HSM-backed encryption with periodic key rotation (SOC 2, ISO 27001, HIPAA, PCI-DSS)
- Any data-at-rest encryption use case where Oracle-managed keys are insufficient for your security posture
- Centralized encryption key for a project or environment shared across multiple OCI resources

## Key Configuration Choices

- **AES-256** (`algorithm: aes`, `length: 32`) -- 256-bit AES is the industry standard for symmetric encryption. All OCI services that support customer-managed encryption use AES keys for data-at-rest encryption. The key shape is immutable after creation.
- **HSM protection** (`protectionMode: hsm`) -- key material never leaves the hardware security module. All encrypt/decrypt operations are performed inside the HSM. This provides the highest level of key protection and is required by most compliance frameworks. The alternative `software` mode stores keys in software and is cheaper but offers lower isolation.
- **90-day auto-rotation** (`isAutoRotationEnabled: true`, `rotationIntervalInDays: 90`) -- OCI automatically creates a new key version every 90 days. Existing data encrypted with older versions remains readable (OCI tracks which version encrypted each block). New encryption operations use the latest version. 90 days balances security (limiting key exposure window) with operational simplicity. Adjust to 30 or 365 days based on your compliance policy.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the key will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<vault-management-endpoint>` | Management endpoint URL of the vault containing this key | `OciKmsVault` status outputs (`managementEndpoint`), or OCI Console > Identity & Security > Vault > Vault Details |

## Related Presets

- **02-rsa-4096-hsm-signing** -- Use instead for asymmetric signing use cases (container image verification, document signing)
