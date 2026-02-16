---
title: "Preset: Asymmetric Signing Key"
description: "Use this preset when you need a key for digital signatures -- signing build artifacts, container images, JWTs, or any data that requires integrity verification."
type: "preset"
rank: "03"
presetSlug: "03-asymmetric-signing"
componentSlug: "kms-key"
componentTitle: "KMS Key"
provider: "gcp"
icon: "package"
order: 3
---

# Preset: Asymmetric Signing Key

## When to Use

Use this preset when you need a key for digital signatures -- signing build
artifacts, container images, JWTs, or any data that requires integrity verification.

Common scenarios:
- **CI/CD artifact signing** -- sign build outputs so consumers can verify authenticity
- **Container image signing** -- used with Binary Authorization for GKE
- **JWT signing** -- sign authentication tokens with a private key held in KMS
- **Code signing** -- verify code integrity in deployment pipelines

## What It Creates

- A KMS key with purpose `ASYMMETRIC_SIGN`
- Algorithm: `EC_SIGN_P256_SHA256` (ECDSA with P-256 curve)
- Protection level: `SOFTWARE` (default)
- No automatic rotation (asymmetric keys are rotated manually)

## Configuration

| Field | Value | Notes |
|-------|-------|-------|
| Purpose | ASYMMETRIC_SIGN | Digital signatures |
| Algorithm | EC_SIGN_P256_SHA256 | Good balance of security and performance |
| Protection Level | SOFTWARE | Default; set to HSM for higher assurance |
| Rotation | Manual | Asymmetric keys don't support auto-rotation |

## Algorithm Options

If `EC_SIGN_P256_SHA256` is not suitable for your use case, consider:

| Algorithm | Key Size | Use Case |
|-----------|----------|----------|
| `EC_SIGN_P256_SHA256` | 256-bit | General purpose (recommended) |
| `EC_SIGN_P384_SHA384` | 384-bit | Higher security, slightly slower |
| `RSA_SIGN_PSS_2048_SHA256` | 2048-bit | RSA compatibility |
| `RSA_SIGN_PSS_4096_SHA512` | 4096-bit | Maximum RSA security |
| `RSA_SIGN_PKCS1_2048_SHA256` | 2048-bit | Legacy PKCS#1 v1.5 compatibility |

## Key Version Management

Unlike symmetric keys, asymmetric keys are **not automatically rotated**. To rotate:

1. Create a new key version in the GCP console or API
2. Distribute the new version's public key to verifiers
3. Update signing operations to use the new version
4. Optionally disable the old version (existing signatures remain valid)

## How to Use

1. Replace `<key-ring-id>` with your fully qualified key ring path
2. Replace `<your-key-name>` with a descriptive name (e.g., `cicd-artifact-signer`)
3. Choose an appropriate algorithm for your signing needs
4. The public key can be retrieved via `gcloud kms keys versions get-public-key`
