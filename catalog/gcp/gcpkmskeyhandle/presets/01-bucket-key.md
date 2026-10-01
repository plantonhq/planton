# Bucket Key

## Use Case

Get a customer-managed key for a Cloud Storage bucket from Autokey, so the bucket's objects are encrypted with a key you own.

## When to Use

- A regional bucket that must use CMEK, in a project with Autokey on
- Replacing hand-made bucket keys with Autokey's

## What This Creates

- The Cloud KMS API on the project
- A key handle for `storage.googleapis.com/Bucket` in `us-central1`, and through it an HSM key; reference `status.outputs.kms_key` from the bucket's `kmsKeyName`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | Match the bucket's location (`us` for a multi-region bucket). |
| `projectId` | `my-gcp-project` | Reference the bucket's `GcpProject`. |
