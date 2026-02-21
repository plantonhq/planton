# OCI KMS Vault Deployment Component

**Date**: 2026-02-20
**Type**: New Feature
**Components**: `apis/org/openmcf/provider/oci/ocikmsvault/v1/`

## Summary

Added the OciKmsVault deployment component -- OCI's Key Management Service vault that provides an HSM-backed container for encryption keys. Supports three vault types: shared HSM (DEFAULT), dedicated HSM (VIRTUAL_PRIVATE), and external key manager integration (EXTERNAL/BYOK/EKMS). This is the first resource of Phase 6 (Security and Secrets). Resource R24 in the OCI provider expansion.

## Problem Statement / Motivation

OpenMCF's OCI provider had no key management component. OCI KMS Vaults are the foundational resource for encryption key management -- every service that uses customer-managed encryption keys (Compute, Block Volume, Object Storage, Database, etc.) requires a vault. Without a declarative vault component, users could not provision encryption infrastructure as part of their infra charts, and downstream components like OciKmsKey and OciVaultSecret cannot be implemented.

## Solution / What's New

A complete OciKmsVault deployment component with both Pulumi (Go) and Terraform (HCL) modules.

### Proto API

- **spec.proto**: 4 top-level fields, 1 embedded enum (VaultType), 2 nested messages (ExternalKeyManagerMetadata, OAuthMetadata)
- **CEL validations**: 2 rules -- external_key_manager_metadata required when vault_type is external; external_key_manager_metadata forbidden when vault_type is not external
- **buf.validate**: compartment_id required, vault_type defined_only + not_in [0], all ExternalKeyManagerMetadata sub-fields min_len validated, oauth_metadata required within ExternalKeyManagerMetadata
- **VaultType enum**: default_vault (shared HSM), virtual_private (dedicated HSM), external (BYOK/EKMS)
- **api.proto**: Standard wrapper with const-validated api_version and kind
- **stack_outputs.proto**: 3 outputs (vault_id, crypto_endpoint, management_endpoint)

### Bundled Resources

1. **Vault** -- the HSM-backed key container with optional external key manager metadata

### Pulumi Module (Go)

5 files across the module package:
- `main.go` -- orchestrator calling vault() with vault_type enum map
- `locals.go` -- Locals struct with freeform tags and display name
- `vault.go` -- kms.NewVault with conditional buildExternalKeyManagerMetadata()
- `outputs.go` -- 3 output constants (vault_id, crypto_endpoint, management_endpoint)

### Terraform Module (HCL)

5 files:
- `main.tf` -- oci_kms_vault.this with dynamic external_key_manager_metadata block
- `locals.tf` -- 1 enum conversion map (vault_type), freeform tags
- `variables.tf`, `outputs.tf`, `provider.tf`

### Validation Tests

22 Ginkgo/Gomega tests (7 valid, 15 invalid scenarios) covering:
- Minimal default_vault and virtual_private configurations
- External vault with full ExternalKeyManagerMetadata and OAuthMetadata
- Display name, compartment_id via valueFrom ref
- Required field validation (compartment_id, vault_type)
- CEL enforcement (external requires metadata, non-external forbids metadata)
- Nested field validation (empty endpoint URL, missing oauth_metadata, empty private_endpoint_id, empty OAuth sub-fields)

### Kind Registration

`OciKmsVault = 3350` under new "Security and Secrets" section in CloudResourceKind enum.

## Implementation Details

### Design Decisions

- **`default_vault` enum value**: Uses `default_vault` instead of `default` because `default` is a reserved keyword in Go, Java, C++, and other target languages.
- **ExternalKeyManagerMetadata included in v1**: Despite being niche, including it ensures the VaultType enum is complete and functional. The conditional CEL pattern prevents misconfiguration.
- **client_app_secret as plain string**: Follows the admin_password precedent from OciAutonomousDatabase, OciDbSystem, and OciPostgresqlDbSystem.
- **private_endpoint_id as plain string**: Not StringValueOrRef because the referenced resource (oci_kms_ekms_private_endpoint) is not modeled as an OpenMCF component.
- **3 outputs**: vault_id for composability, crypto_endpoint and management_endpoint for downstream OciKmsKey consumption.
- **No vault replication**: oci_kms_vault_replication is a separate resource with an independent lifecycle.
- **Directory name**: `ocikmsvault` (per WA02 -- lowercased kind name, not id_prefix `ocivlt`).

### Excluded

- `restore_from_file` / `restore_from_object_store` / `restore_trigger` -- operational restore
- `time_of_deletion` -- deletion scheduling
- `defined_tags`, `system_tags` -- managed by platform
- `freeform_tags` -- auto-populated from metadata labels
- `oci_kms_vault_replication` -- separate resource with independent lifecycle

## Impact

- Adds 1 new CloudResourceKind to the OCI provider (R24 of 37)
- Opens Phase 6 (Security and Secrets) -- first of 4 resources
- Unblocks R25 OciKmsKey (depends on vault_id and management_endpoint outputs)
- Next: R25 OciKmsKey (enum 3351, Phase 6: Security and Secrets)
