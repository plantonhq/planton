# OCI KMS Key Deployment Component

**Date**: 2026-02-20
**Type**: New Feature
**Components**: `apis/org/openmcf/provider/oci/ocikmskey/v1/`

## Summary

Added the OciKmsKey deployment component -- OCI's managed cryptographic key stored inside a KMS Vault, used to encrypt data at rest across OCI services (Block Volume, Object Storage, Database, File Storage, etc.). Supports three encryption algorithms (AES, RSA, ECDSA) with configurable key lengths, three protection modes (HSM, SOFTWARE, EXTERNAL), and automatic key rotation schedules. This is the second resource of Phase 6 (Security and Secrets). Resource R25 in the OCI provider expansion.

## Problem Statement / Motivation

OpenMCF's OCI provider had a KMS Vault component (R24) but no way to create encryption keys within those vaults. OCI KMS Keys are the resources that actually perform cryptographic operations -- without them, customer-managed encryption for Block Volumes, Object Storage, Autonomous Database, and other services is not possible. This component completes the vault-to-key relationship and enables end-to-end encryption infrastructure as code.

## Solution / What's New

A complete OciKmsKey deployment component with both Pulumi (Go) and Terraform (HCL) modules.

### Proto API

- **spec.proto**: 8 top-level fields, 3 nested messages (KeyShape, AutoKeyRotationDetails, ExternalKeyReference), 3 enums (ProtectionMode, Algorithm, CurveId)
- **CEL validations**: 7 rules total -- 3 at spec level (external_key_reference conditional on protection_mode, auto_key_rotation_details requires is_auto_rotation_enabled), 4 at KeyShape level (algorithm/length matrix for AES/RSA/ECDSA, ECDSA curve_id/length consistency, non-ECDSA forbids curve_id)
- **buf.validate**: compartment_id required, management_endpoint required, key_shape required, algorithm defined_only + not_in [0], length > 0, curve_id defined_only, external_key_id min_len
- **api.proto**: Standard wrapper with const-validated api_version and kind
- **stack_outputs.proto**: 2 outputs (key_id, current_key_version)

### Bundled Resources

1. **Key** -- the cryptographic key with configurable algorithm, length, protection mode, and optional auto-rotation and external key reference

### Pulumi Module (Go)

5 files across the module package:
- `main.go` -- orchestrator calling key() with 3 enum maps (protectionMode, algorithm, curveId)
- `locals.go` -- Locals struct with freeform tags and display name
- `key.go` -- kms.NewKey with buildKeyShape(), conditional buildAutoKeyRotationDetails(), conditional buildExternalKeyReference()
- `outputs.go` -- 2 output constants (key_id, current_key_version)

### Terraform Module (HCL)

5 files:
- `main.tf` -- oci_kms_key.this with key_shape block, dynamic auto_key_rotation_details and external_key_reference blocks
- `locals.tf` -- 3 enum conversion maps (protection_mode, algorithm, curve_id), freeform tags
- `variables.tf`, `outputs.tf`, `provider.tf`

### Validation Tests

37 Ginkgo/Gomega tests (19 valid, 18 invalid scenarios) covering:
- All three algorithms with valid lengths (AES: 16/24/32, RSA: 256/384/512, ECDSA: P-256/P-384/P-521)
- All three protection modes (HSM, SOFTWARE, EXTERNAL)
- Auto-rotation with and without details
- StringValueOrRef with literal and valueFrom patterns
- Required field validation (compartment_id, management_endpoint, key_shape, algorithm, length)
- Algorithm/length matrix enforcement (invalid AES length, invalid RSA length)
- ECDSA curve_id/length consistency (mismatched curve and length)
- Curve_id forbidden for AES and RSA
- External key reference conditional validation (required for EXTERNAL, forbidden for non-EXTERNAL)
- Auto-rotation details conditional (requires is_auto_rotation_enabled)

### Kind Registration

`OciKmsKey = 3351` under "Security and Secrets" section in CloudResourceKind enum.

## Implementation Details

### Design Decisions

- **Key version NOT bundled**: Plan stub suggested `oci_kms_key + key_version`, but key versions are created automatically with the key. `oci_kms_key_version` is create-only (triggers manual rotation) -- an operational concern. Auto-rotation via `auto_key_rotation_details` provides declarative rotation.
- **management_endpoint as StringValueOrRef**: default_kind OciKmsVault, default_kind_field_path "status.outputs.managementEndpoint" -- enables infra-chart composability where vault output feeds directly into key creation.
- **protection_mode optional**: OCI defaults to HSM (most secure). Unlike vault_type (required), protection mode has a sensible secure default.
- **CurveId zero-value prefix**: `curve_unspecified` instead of `unspecified` to avoid C++ scoping collision with Algorithm.unspecified within the same KeyShape message.
- **desired_state excluded**: Operational toggle (ENABLED/DISABLED), keys always create as ENABLED. Disabling is an operational action, not an IaC deployment concern.
- **Directory name**: `ocikmskey` (per WA02 -- lowercased kind name, not id_prefix `ocikey`).

### Excluded

- `oci_kms_key_version` -- operational rotation trigger, not deployment
- `desired_state` -- operational toggle
- `restore_from_file` / `restore_from_object_store` / `restore_trigger` -- operational restore
- `time_of_deletion` -- deletion scheduling
- `defined_tags`, `system_tags` -- managed by platform
- `freeform_tags` -- auto-populated from metadata labels

## Impact

- Adds 1 new CloudResourceKind to the OCI provider (R25 of 37)
- Second resource of Phase 6 (Security and Secrets) -- 2 of 4 resources complete
- Enables customer-managed encryption for OciBlockVolume, OciObjectStorageBucket, OciAutonomousDatabase, and other encryption-capable resources
- Next: R26 OciVaultSecret (enum 3352, Phase 6: Security and Secrets)
