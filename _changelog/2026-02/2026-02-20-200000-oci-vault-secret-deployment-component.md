# OCI Vault Secret Deployment Component

**Date**: 2026-02-20
**Type**: New Feature
**Components**: `apis/dev/planton/provider/oci/ocivaultsecret/v1/`

## Summary

Added the OciVaultSecret deployment component -- OCI's managed secret stored in a KMS Vault, encrypted by a master encryption key, supporting explicit base64 content, automatic secret generation (bytes, passphrase, SSH key), lifecycle rules (expiry and reuse), and scheduled rotation against Autonomous Database or OCI Functions targets. This is the third resource of Phase 6 (Security and Secrets). Resource R26 in the OCI provider expansion.

## Problem Statement / Motivation

Planton's OCI provider had KMS Vault (R24) and KMS Key (R25) components for encryption infrastructure, but no way to store secrets (credentials, API keys, certificates) within those vaults. OCI Vault Secrets are the primary mechanism for managing sensitive data across OCI services -- without them, platform teams cannot declaratively manage secrets for database credentials, API tokens, or SSH keys with proper lifecycle controls, rotation, and encryption.

## Solution / What's New

A complete OciVaultSecret deployment component with both Pulumi (Go) and Terraform (HCL) modules.

### Proto API

- **spec.proto**: 11 top-level fields, 4 nested messages (SecretContent, SecretGenerationContext, SecretRule, RotationConfig with nested TargetSystemDetails), 3 enums (GenerationType, RuleType, TargetSystemType)
- **CEL validations**: 4 rules total -- 3 at spec level (secret_content mutually exclusive with auto-generation, auto-generation requires context, context requires auto-generation), 1 at SecretGenerationContext level (passphrase requires length > 0)
- **buf.validate**: compartment_id required, secret_name min_len 1, vault_id required, key_id required, generation_type/rule_type/target_system_type defined_only + not_in [0], generation_template min_len 1, target_system_details required in rotation_config, stage in-list validation
- **api.proto**: Standard wrapper with const-validated api_version and kind
- **stack_outputs.proto**: 2 outputs (secret_id, current_version_number)

### Bundled Resources

1. **Secret** -- the vault secret with configurable content (explicit or auto-generated), lifecycle rules, rotation, and metadata

### Pulumi Module (Go)

5 files across the module package:
- `main.go` -- orchestrator calling secret() with 3 enum maps (generationType, ruleType, targetSystemType)
- `locals.go` -- Locals struct with freeform tags and display name (uses secret_name)
- `secret.go` -- vault.NewSecret with buildSecretContent(), buildSecretGenerationContext(), buildSecretRules(), buildRotationConfig() + buildTargetSystemDetails()
- `outputs.go` -- 2 output constants (secret_id, current_version_number)

### Terraform Module (HCL)

5 files:
- `main.tf` -- oci_vault_secret.this with 4 dynamic blocks (secret_content, secret_generation_context, secret_rules, rotation_config with nested target_system_details)
- `locals.tf` -- 3 enum conversion maps (generation_type, rule_type, target_system_type), freeform tags
- `variables.tf`, `outputs.tf`, `provider.tf`

### Validation Tests

37 Ginkgo/Gomega tests (20 valid, 17 invalid scenarios) covering:
- Minimal secret (no content, no auto-generation)
- Explicit base64 content with stage (CURRENT, PENDING) and version name
- Auto-generation: bytes, passphrase (with length), ssh_key, with secret_template
- Description and secret_metadata fields
- Expiry rules, reuse rules, and combined rules
- Rotation config with ADB target, Function target, and scheduling disabled
- StringValueOrRef with literal and valueFrom patterns
- Fully populated secret with all features
- Required field validation (compartment_id, secret_name, vault_id, key_id)
- Mutual exclusivity: secret_content + auto-generation, context without auto-generation
- Enum validation: unspecified generation_type, empty template, zero passphrase length
- Rule type and target system type validation
- Invalid stage value

### Kind Registration

`OciVaultSecret = 3352` under "Security and Secrets" section in CloudResourceKind enum.

## Implementation Details

### Design Decisions

- **content_type hardcoded to BASE64**: Only valid value today. Hardcoded in both IaC modules; not exposed in spec to avoid false optionality.
- **stage as plain string with in-list validation**: "CURRENT" and "PENDING" are the only valid user-set values. Empty string allowed (defaults to CURRENT). Plain string avoids enum overhead for a 2-value field.
- **secret_metadata field name**: Named `secret_metadata` to avoid collision with Planton's `CloudResourceMetadata` on the parent message. Maps to OCI's `metadata` attribute.
- **Rotation target adb_id as StringValueOrRef**: default_kind OciAutonomousDatabase enables infra-chart composability. function_id uses StringValueOrRef without default_kind (individual functions are code artifacts, not Planton components).
- **Display name = secret_name**: Unlike other components that use display_name or metadata.name, secrets are identified by secret_name. The Locals.DisplayName is set to secret_name for consistency in Pulumi resource naming.
- **Secret versions not bundled**: Secret versions are managed implicitly -- updating secret_content creates a new version automatically. current_version_number is a computed output.
- **Directory name**: `ocivaultsecret` (per WA02 -- lowercased kind name, not id_prefix `ocisec`).

### Excluded

- `replication_config` -- cross-region replication is a separate infrastructure concern (consistent with OciKmsVault excluding vault replication)
- `defined_tags`, `system_tags` -- managed by platform
- `freeform_tags` -- auto-populated from metadata labels

## Impact

- Adds 1 new CloudResourceKind to the OCI provider (R26 of 37)
- Third resource of Phase 6 (Security and Secrets) -- 3 of 4 resources complete
- Completes the vault-key-secret chain: OciKmsVault -> OciKmsKey -> OciVaultSecret
- Enables declarative secrets management for database credentials, API keys, certificates, and SSH keys
- Next: R27 OciBastion (enum 3353, Phase 6: Security and Secrets)
