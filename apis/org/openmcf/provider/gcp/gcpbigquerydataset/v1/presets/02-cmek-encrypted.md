# Preset: CMEK-Encrypted Dataset

## When to Use

Use this preset when your dataset contains sensitive or regulated data that
requires customer-managed encryption keys (CMEK). Common compliance scenarios:

- **PCI DSS** -- payment card data
- **HIPAA** -- healthcare data
- **SOX** -- financial reporting data
- **GDPR** -- EU personal data with encryption requirements

## What It Creates

- A BigQuery dataset in a specific region (for data residency)
- Customer-managed encryption via Cloud KMS
- Physical storage billing (cost-optimized for large datasets)
- No explicit access entries (uses project-level defaults)

## Configuration

| Field | Value | Notes |
|-------|-------|-------|
| Location | Regional | Specific region for data residency compliance |
| Encryption | CMEK | Customer-managed via Cloud KMS |
| Billing Model | PHYSICAL | Cost savings for compressible data |
| Time Travel | 168 hours | Default maximum for recovery |

## Prerequisites

- A Cloud KMS key ring and key in the **same region** as the dataset
- The BigQuery service account must have `roles/cloudkms.cryptoKeyEncrypterDecrypter`
  on the KMS key

## How to Use

1. Replace `<project-id>` with your GCP project ID
2. Replace `<your_dataset_id>` with a descriptive name
3. Replace `<kms-key-name>` with the fully qualified KMS key path
4. Update `location` to match your KMS key's region
5. Add `description` to document the dataset's purpose and compliance scope

## Cost Considerations

- PHYSICAL billing typically reduces storage costs 60-80% for compressible data
- CMEK adds per-operation costs for encryption/decryption
- Regional storage is generally cheaper than multi-regional
