# CMEK Image

## Use Case

Encrypt an image with a customer-managed key that Cloud KMS Autokey created for images.

## When to Use

- Regulated workloads that require customer-managed encryption on boot images
- Projects that use Autokey for CMEK

## What This Creates

- The Compute Engine API on the project
- One image encrypted with the key handle's key

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `kmsKey` | a `GcpKmsKeyHandle` reference | A `GcpKmsKey` you manage. The key must be in the image's storage location. |
| `sourceImage` | Debian 12 family | Your source. |
