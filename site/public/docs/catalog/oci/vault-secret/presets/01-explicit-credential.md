---
title: "Explicit Credential"
description: "This preset creates an OCI Vault Secret with user-provided base64-encoded content and a 90-day version expiry rule. This is the standard pattern for storing application credentials, API keys,..."
type: "preset"
rank: "01"
presetSlug: "01-explicit-credential"
componentSlug: "vault-secret"
componentTitle: "Vault Secret"
provider: "oci"
icon: "package"
order: 1
---

# Explicit Credential

This preset creates an OCI Vault Secret with user-provided base64-encoded content and a 90-day version expiry rule. This is the standard pattern for storing application credentials, API keys, database passwords, TLS certificates, or any sensitive data that the user provides directly. The expiry rule enforces rotation awareness by blocking retrieval of stale secret versions.

## When to Use

- Storing application API keys, database passwords, or service account credentials that are provisioned outside OCI
- TLS certificates and private keys that need centralized, encrypted storage with access auditing
- Configuration secrets (OAuth client secrets, webhook signing keys) consumed by applications via OCI SDK or CLI
- Any secret where the content is known at deployment time and provided by the user or CI/CD pipeline

## Key Configuration Choices

- **Explicit content** (`secretContent`) -- the user provides base64-encoded secret data directly. This is the most common pattern. Encode your secret value with `echo -n 'my-secret-value' | base64` before placing it in the `content` field. Updating the content field creates a new secret version automatically.
- **CURRENT stage** (`stage: CURRENT`) -- marks this content version as the active version that consumers retrieve by default. The alternative `PENDING` stage is used during manual rotation workflows where a new version is staged before being promoted.
- **90-day version expiry** (`secretRules` with `secret_expiry_rule`, `P90D`) -- each secret version expires 90 days after creation. After expiry, retrieval is blocked (`isSecretContentRetrievalBlockedOnExpiry: true`), forcing consumers to use a newer version. This prevents indefinite use of stale credentials. Adjust the interval to match your rotation policy (e.g., `P30D` for 30 days, `P365D` for annual rotation).
- **Vault and key references** (`vaultId`, `keyId`) -- the secret is stored in the specified vault and encrypted with the specified AES master key. Both are immutable after creation. Use an `OciKmsVault` and `OciKmsKey` created from the companion presets.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the secret will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<secret-name>` | Unique name for the secret within the vault (e.g., `myapp-db-password`) | Choose a name following your naming convention |
| `<vault-ocid>` | OCID of the vault that will contain this secret | `OciKmsVault` status outputs (`vaultId`), or OCI Console > Identity & Security > Vault |
| `<encryption-key-ocid>` | OCID of the AES master encryption key within the vault | `OciKmsKey` status outputs (`keyId`), or OCI Console > Identity & Security > Vault > Keys |
| `<base64-encoded-secret-value>` | Base64-encoded secret data | `echo -n 'your-secret' \| base64` |

## Related Presets

- **02-auto-generated-passphrase** -- Use instead when OCI should generate and rotate the secret automatically (e.g., database credentials)
