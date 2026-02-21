# AliCloudRamRole Component Added

**Date**: 2026-02-19
**Component**: AliCloudRamRole
**Enum**: 3010
**ID Prefix**: acramr

## Summary

Added the AliCloudRamRole component for managing Alibaba Cloud RAM roles with bundled policy attachments.

RAM roles are the identity foundation for Alibaba Cloud -- ACK clusters, FC functions, ECS instances, and SAE applications all use RAM roles for service authentication via STS (Security Token Service).

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudramrole/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AliCloudRamRole = 3010` in `CloudResourceKind` enum

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider, RAM role, and iterates policy attachments as child resources
- **Terraform** (HCL): Same resources using `for_each` for policy attachments

### Tests
- 14 Ginkgo/Gomega spec validation tests covering:
  - Valid: minimal fields, policy attachments, all optional fields, boundary values
  - Invalid: missing required fields, role_name length, max_session_duration range, invalid policy_type, wrong api_version/kind, missing policy_name

### Documentation
- README.md with configuration reference, trust policy patterns, and related components
- examples.md with minimal, ECS service role, and cross-account YAML examples

## Spec Design Decisions

- **`role_name` not `name`**: Follows provider-authentic naming (`name` is deprecated in TF since v1.252.0)
- **`max_session_duration`**: Added beyond T02 spec (range 3600-43200s, default 3600) -- important for CI/CD and cross-account workflows
- **`tags`**: Added for consistency with LogProject and all existing provider components
- **`force`**: Added for clean teardown support (force-detach policies before deletion)
- **`policy_type` validation**: Uses `buf.validate string.in` with "System" and "Custom" values

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test -v ./...` -- PASS (14/14 specs)
- `terraform init` -- PASS
- `terraform validate` -- PASS
