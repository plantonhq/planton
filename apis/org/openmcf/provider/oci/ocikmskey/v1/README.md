# OciKmsKey

## Overview

OciKmsKey is an OpenMCF component that deploys an OCI KMS encryption key inside a KMS vault. It provides a single declarative manifest to create a cryptographic key with a specific algorithm, key length, protection mode, and optional automatic rotation schedule.

## Purpose

KMS keys are the actual cryptographic material used by OCI services for data-at-rest encryption. Block Volumes, Object Storage buckets, databases, and other resources reference a KMS key OCID to enable customer-managed encryption. This component provisions the key and exports its OCID for composability with those downstream resources.

## Key Features

- **Three algorithms** — AES (symmetric encryption), RSA (asymmetric encryption/signing), and ECDSA (elliptic curve signing).
- **Three protection modes** — HSM (FIPS 140-2 Level 3), software (lower cost), and external (BYOK/EKMS).
- **Automatic rotation** — configurable rotation interval and start time for automated key version rotation.
- **Vault binding** — ties the key to a specific vault via the `managementEndpoint` reference.
- **Foreign key references** — `compartmentId` and `managementEndpoint` support `valueFrom` to reference OpenMCF-managed OciCompartment and OciKmsVault resources.

## Constraints

- `keyShape` (algorithm, length, curveId) is immutable after creation.
- `protectionMode` is immutable after creation.
- `curveId` is required for ECDSA and must not be set for AES or RSA.
- `externalKeyReference` is required when `protectionMode` is `external` and must not be set otherwise.
- `autoKeyRotationDetails` can only be set when `isAutoRotationEnabled` is `true`.
- AES key lengths: 16 (128-bit), 24 (192-bit), 32 (256-bit) bytes.
- RSA key lengths: 256 (2048-bit), 384 (3072-bit), 512 (4096-bit) bytes.
- ECDSA key lengths: 32 (P-256), 48 (P-384), 66 (P-521) bytes.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Standard data-at-rest encryption | AES-256 key with HSM protection |
| Digital signature verification | RSA-4096 or ECDSA P-384 with HSM |
| Cost-sensitive non-regulatory | AES-256 with software protection |
| Regulatory BYOK | AES-256 with external protection |
| Automated compliance rotation | Auto-rotation every 90 days |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels`, including `resource_kind`, `resource_id`, `organization`, and `environment`.
- **HSM isolation** — FIPS 140-2 Level 3 hardware security module for production cryptographic operations.
- **Auto-rotation** — scheduled key version rotation to meet compliance rotation requirements without manual intervention.
