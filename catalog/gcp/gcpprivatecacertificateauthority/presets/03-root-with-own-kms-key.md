# Root With Own KMS Key

## Use Case

A root whose signing key lives in a Cloud KMS key you own, so every signature shows up in your KMS audit trail and the key follows your own IAM and rotation policy.

## When to Use

- Compliance regimes that require customer-controlled CA keys
- Enterprise pools only (Google's tier rule)

## What This Creates

- An enabled root in `root-pool` signing with the primary version of the `root-ca-signing` KMS key (an asymmetric-sign key; CA Service's service agent needs signerVerifier and viewer on it)

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `keySpec.cloudKmsKeyVersion` | the key's primary version | Pin a specific version instead of following the primary. |
| `lifetime` | `630720000s` (20 years) | Match your root rotation policy. |
