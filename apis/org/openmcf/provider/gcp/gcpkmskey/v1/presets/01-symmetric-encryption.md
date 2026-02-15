# Preset: Symmetric Encryption Key (CMEK)

## When to Use

Use this preset when you need a standard customer-managed encryption key (CMEK)
for encrypting data in GCP services like BigQuery, Spanner, CloudSQL, GCS, GKE,
PubSub, or AlloyDB.

This is the **most common** KMS key pattern. It creates a symmetric encryption key
with 90-day automatic rotation using software-level protection.

## What It Creates

- A KMS key with purpose `ENCRYPT_DECRYPT` (default)
- Algorithm: `GOOGLE_SYMMETRIC_ENCRYPTION` (default)
- Protection level: `SOFTWARE` (default)
- Automatic rotation every 90 days

## Configuration

| Field | Value | Notes |
|-------|-------|-------|
| Purpose | ENCRYPT_DECRYPT | Default, not explicitly set |
| Algorithm | GOOGLE_SYMMETRIC_ENCRYPTION | Default, not explicitly set |
| Protection Level | SOFTWARE | Default, not explicitly set |
| Rotation | 90 days | Creates new primary version automatically |

## How to Use

1. Replace `<key-ring-id>` with your fully qualified key ring path
2. Replace `<your-key-name>` with a descriptive name (e.g., `prod-data-cmek`)
3. Adjust `rotationPeriod` if 90 days is not appropriate for your security policy

## Downstream Usage

Reference this key's `key_id` output from any CMEK-enabled GCP resource:

```yaml
kmsKeyId:
  valueFrom:
    kind: GcpKmsKey
    name: my-cmek-key
    fieldPath: status.outputs.key_id
```
