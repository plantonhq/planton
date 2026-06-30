# OciVaultSecret

## Overview

OciVaultSecret is an Planton component that deploys an OCI Vault secret. It provides a single declarative manifest to store sensitive data (credentials, certificates, API keys) in a KMS vault with encryption, lifecycle rules, and optional scheduled rotation.

## Purpose

Secrets management is a critical part of any infrastructure platform. OCI Vault secrets provide encrypted storage for sensitive data with versioning, expiry policies, and automated rotation. This component wraps the secret creation, content provisioning (explicit or auto-generated), lifecycle rules, and rotation configuration into a single resource.

## Key Features

- **Two content modes** — explicit base64 content or OCI-managed auto-generation (bytes, passphrase, SSH key).
- **Lifecycle rules** — configurable expiry intervals, absolute expiry timestamps, and content reuse prevention.
- **Scheduled rotation** — automatic credential rotation against Autonomous Database or custom OCI Functions targets.
- **Version tracking** — content updates create new secret versions; the `currentVersionNumber` output tracks the active version.
- **Foreign key references** — `compartmentId`, `vaultId`, `keyId`, and rotation target `adbId` support `valueFrom` for composability.

## Constraints

- `secretName`, `vaultId`, and `keyId` are immutable after creation.
- `secretContent` and `enableAutoGeneration` are mutually exclusive.
- `secretGenerationContext` is required when `enableAutoGeneration` is `true` and must not be set otherwise.
- `passphraseLength` must be > 0 when `generationType` is `passphrase`.
- `secretVersionExpiryInterval` range: 1-90 days (ISO 8601).
- `timeOfAbsoluteExpiry` range: 1-365 days from creation (RFC 3339).
- `rotationInterval` range: 1-360 days (ISO 8601).
- The encryption key must be a symmetric key within the specified vault.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Store a database password | Explicit base64 content |
| Generate a random passphrase | Auto-generation with `passphrase` type |
| Generate SSH key pairs | Auto-generation with `ssh_key` type |
| Enforce secret expiry | Expiry rule with 30-day interval |
| Prevent secret reuse | Reuse rule enforced on deleted versions |
| Auto-rotate ADB credentials | Scheduled rotation with ADB target |
| Custom rotation logic | Scheduled rotation with Functions target |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels`, including `resource_kind`, `resource_id`, `organization`, and `environment`.
- **Encryption at rest** — all secret content is encrypted by the specified KMS master key.
- **Automatic rotation** — scheduled rotation updates credentials in target systems without manual intervention.
- **Version history** — content updates create new versions, enabling rollback and audit trails.
