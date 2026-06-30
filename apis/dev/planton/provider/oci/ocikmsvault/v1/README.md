# OciKmsVault

## Overview

OciKmsVault is an Planton component that deploys an OCI Key Management Service vault. It provides a single declarative manifest to create an HSM-backed container for encryption keys with configurable isolation levels — shared, dedicated, or external key manager.

## Purpose

KMS vaults are the foundation of OCI's encryption infrastructure. Every encryption key belongs to a vault, and every OCI service that supports customer-managed encryption (Block Volume, Object Storage, Database, etc.) requires a KMS key from a vault. This component provisions the vault and exports the management and crypto endpoints that downstream OciKmsKey resources need for key creation and cryptographic operations.

## Key Features

- **Three vault types** — shared HSM (`default_vault`), dedicated HSM (`virtual_private`), and external key manager (`external`) for BYOK/EKMS.
- **Endpoint export** — outputs `managementEndpoint` and `cryptoEndpoint` for composability with OciKmsKey resources.
- **External key manager support** — IDCS OAuth integration with third-party HSMs via a KMS private endpoint.
- **Foreign key references** — `compartmentId` supports `valueFrom` to reference an Planton-managed OciCompartment.

## Constraints

- `vaultType` is immutable after creation — changing it forces recreation.
- `externalKeyManagerMetadata` must be set when `vaultType` is `external` and must not be set otherwise.
- All fields in `externalKeyManagerMetadata` are immutable after creation.
- `clientAppSecret` is sensitive and not returned by the API after creation.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Standard encryption for dev/test | `default_vault` — shared HSM, lower cost |
| High-throughput production encryption | `virtual_private` — dedicated HSM partition |
| Regulatory BYOK/EKMS requirement | `external` — keys managed by a third-party HSM |
| Multi-key organization | One vault containing multiple OciKmsKey resources |

## Production Features

- **Freeform tags** — automatically populated from `metadata.labels`, including `resource_kind`, `resource_id`, `organization`, and `environment`.
- **Dedicated HSM isolation** — `virtual_private` vaults provide cryptographic isolation and higher throughput limits for production workloads.
- **BYOK/EKMS** — `external` vaults enable compliance scenarios where keys must never leave a customer-controlled HSM.
