# AliCloudRamPolicy Component Added

**Date**: 2026-02-19
**Component**: AliCloudRamPolicy
**Enum**: 3011
**ID Prefix**: acramp

## Summary

Added the AliCloudRamPolicy component for managing Alibaba Cloud RAM custom IAM policies.

Custom policies fill the gap when Alibaba Cloud's system-managed policies don't provide the exact permission boundaries you need. Once created, a custom policy can be attached to RAM roles, users, or groups via their respective components.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudrampolicy/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AliCloudRamPolicy = 3011` in `CloudResourceKind` enum

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and RAM policy resource, exports policy_name and policy_type
- **Terraform** (HCL): Same resource with tag computation in locals

### Tests
- 13 Ginkgo/Gomega spec validation tests covering:
  - Valid: minimal fields, all optional fields, rotate_strategy options, boundary policy_name length
  - Invalid: missing required fields (region, policy_name, policy_document), policy_name exceeding 128 chars, description exceeding 1024 chars, invalid rotate_strategy, wrong api_version/kind, missing metadata

### Documentation
- README.md with configuration reference, policy document structure, and related components
- examples.md with minimal, scoped-bucket, and multi-service CI/CD YAML examples

## Corrections from T02 Spec

- **`name` -> `policy_name`**: Follows provider-authentic naming consistent with AliCloudRamRole's `role_name` pattern
- **Added `rotate_strategy`**: Not in T02 but important for production policies hitting the 5-version limit
- **Added `tags`**: Consistent with all other components in the catalog
- **Output `policy_type`**: Closes the contract with AliCloudRamRole's policy_attachments (needs both name and type)

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test -v ./...` -- PASS (13/13 specs)
- `terraform init` -- PASS
- `terraform validate` -- PASS
