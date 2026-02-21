# AliCloudKmsKey Component Added

**Date**: 2026-02-19
**Component**: AliCloudKmsKey
**Enum**: 3060
**ID Prefix**: ackms

## Summary

Added the AliCloudKmsKey deployment component -- a standalone KMS customer-managed key (CMK) for data encryption and digital signing across Alibaba Cloud services. This key serves as the root of trust for envelope encryption used by RDS (TDE), OSS (SSE-KMS), ECS (disk encryption), and PolarDB.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudkmskey/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AliCloudKmsKey = 3060` in `CloudResourceKind` enum under a new Security category

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and a single `kms.Key` resource with bool-to-string conversion for `automatic_rotation` and `deletion_protection`, default resolution for optional fields
- **Terraform** (HCL): Single `alicloud_kms_key` resource with matching variables, outputs, locals for bool-to-string conversion, and tag merging

### Tests
- Ginkgo/Gomega spec validation tests covering: valid inputs (minimal, full config, HSM protection, all 5 asymmetric key specs, all 4 symmetric key specs, pending_window boundaries, rotation disabled), invalid inputs (missing region, wrong api_version/kind, missing metadata, invalid key_spec, invalid key_usage, invalid protection_level, pending_window below minimum, pending_window above maximum)

### Documentation
- README.md with configuration reference, key spec values table, immutability notes, deletion behavior, and related components
- examples.md with 3 YAML examples (minimal, production with rotation, asymmetric signing)
- catalog-page.md with full configuration reference and examples
- docs/README.md with comprehensive research documentation covering history, deployment methods, design decisions, and best practices
- Pulumi overview.md documenting module architecture, bool conversion, and output keys
- Pulumi/TF README.md and examples.md

### Presets
- 01-standard: AES-256, no rotation, no deletion protection (development/staging)
- 02-production-with-rotation: AES-256, annual rotation, deletion protection enabled (production)
- 03-asymmetric-signing: RSA-2048, SIGN/VERIFY usage (digital signatures)

## Spec Design Decisions

- **Bool for `automatic_rotation` and `deletion_protection`**: Provider uses string "Enabled"/"Disabled", but bool provides a cleaner YAML UX (`automaticRotation: true`). Conversion handled in IaC modules.
- **9 key_spec values (not 4)**: T02 listed 4 values; the actual provider supports 9 (AES-128/192 for Dedicated KMS, RSA-3072, EC_P256K, EC_SM2). All included for completeness.
- **`pending_window_in_days` range 7-366**: Alibaba Cloud allows up to 366 days (unlike AWS which caps at 30). Default 30 provides a safe recovery window.
- **`deletion_protection` added**: Not in T02 spec, but critical for KMS keys where accidental deletion means permanent, irrecoverable data loss. Defaults to false for simplicity; production preset enables it.
- **Fields excluded for v1**: `dkms_instance_id` (Dedicated KMS, enterprise-only), `origin` (BYOK is a niche use case), `policy` (JSON IAM policy adds complexity; users manage via RAM), `status` (operational state, not config).

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
- `go build ./pkg/crkreflect/...` -- PASS (kind map regenerated)
