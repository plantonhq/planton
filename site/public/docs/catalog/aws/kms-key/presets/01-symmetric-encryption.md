---
title: "Symmetric Encryption Key"
description: "This preset creates a customer-managed symmetric KMS key with automatic annual rotation enabled and the maximum 30-day deletion window. Symmetric keys are the most common KMS key type, used for..."
type: "preset"
rank: "01"
presetSlug: "01-symmetric-encryption"
componentSlug: "kms-key"
componentTitle: "KMS Key"
provider: "aws"
icon: "package"
order: 1
---

# Symmetric Encryption Key

This preset creates a customer-managed symmetric KMS key with automatic annual rotation enabled and the maximum 30-day deletion window. Symmetric keys are the most common KMS key type, used for envelope encryption of data at rest across AWS services and for generating data keys.

## When to Use

- Encrypting data at rest across AWS services (S3, EBS, RDS, DynamoDB, Secrets Manager)
- Customer-managed key requirement for compliance (HIPAA, PCI-DSS, SOC2)
- Envelope encryption where AWS generates and manages data keys on your behalf

## Key Configuration Choices

- **Symmetric key** (`keySpec: symmetric`) -- AWS-recommended default for general-purpose encryption; supports encrypt, decrypt, and generate data key operations
- **Rotation enabled** (`disableKeyRotation: false`) -- Automatic annual key rotation for compliance; AWS retains all previous key material for decryption
- **30-day deletion window** (`deletionWindowDays: 30`) -- Maximum protection against accidental deletion of encryption keys

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<key-description>` | Human-readable description of the key's purpose (e.g., "Production database encryption key") | Your team's naming conventions |
| `<key-alias>` | Short alias for the key (e.g., `myapp/data-encryption`) | Your team's naming conventions; must match `alias/[A-Za-z0-9/_-]+` |
