---
title: "Asymmetric Signing Key"
description: "This preset creates an RSA-2048 asymmetric key for digital signature generation and verification. The private key never leaves KMS; only the public key can be exported for external verification."
type: "preset"
rank: "03"
presetSlug: "03-asymmetric-signing"
componentSlug: "kms-key"
componentTitle: "KMS Key"
provider: "alicloud"
icon: "package"
order: 3
---

# Asymmetric Signing Key

This preset creates an RSA-2048 asymmetric key for digital signature generation and verification. The private key never leaves KMS; only the public key can be exported for external verification.

## When to Use

- Signing JWTs or API tokens where the verifier only needs the public key
- Code signing and artifact verification pipelines
- Certificate signing (internal PKI)
- Blockchain or cryptocurrency-related signature operations (use `EC_P256K` key spec for secp256k1)

## Key Configuration Choices

- **RSA_2048** -- Widely supported 2048-bit RSA key. Use `RSA_3072` for stronger security or `EC_P256` for ECDSA if your verification ecosystem supports it.
- **SIGN/VERIFY** -- Asymmetric keys cannot be used for ENCRYPT/DECRYPT. This usage type restricts the key to signing operations only.
- **SOFTWARE** (default) -- Software-based protection. Use `HSM` if regulatory requirements mandate hardware key storage.
- **No rotation** -- Automatic rotation is not supported for asymmetric keys. Key rotation for asymmetric keys must be managed manually by creating a new key and updating references.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<key-name>` | Unique key name (e.g., `api-signing-key`, `jwt-signing-key`) | Your naming convention |
| `<alibaba-cloud-region>` | Alibaba Cloud region code | Your deployment region strategy |
| `<key-description>` | Description (e.g., `RSA key for JWT signing`) | Your key inventory |
| `<your-team>` | Owning team | Your organizational structure |

## Alternative Key Specs

| Key Spec | Algorithm | Use Case |
|----------|-----------|----------|
| `RSA_2048` | RSA 2048-bit (this preset) | General-purpose signing, widest compatibility |
| `RSA_3072` | RSA 3072-bit | Stronger RSA key for long-term use |
| `EC_P256` | NIST P-256 ECDSA | Compact signatures, modern TLS/JWT |
| `EC_P256K` | secp256k1 ECDSA | Blockchain, cryptocurrency |
| `EC_SM2` | Chinese SM2 | Chinese national standard compliance |

## Related Presets

- **01-standard** -- Use instead for symmetric encryption keys
- **02-production-with-rotation** -- Use instead for production encryption keys with rotation
