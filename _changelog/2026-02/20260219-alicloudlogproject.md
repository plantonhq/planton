# AliCloudLogProject Component Added

**Date**: 2026-02-19
**Component**: AliCloudLogProject
**Enum**: 3000
**ID Prefix**: aclog

## Summary

Added the first Alibaba Cloud deployment component: AliCloudLogProject.

This component manages an Alibaba Cloud Simple Log Service (SLS) project with optional bundled log stores and full-text search indexes.

## What Was Created

### API Definition
- `apis/dev/planton/provider/alicloud/alicloudlogproject/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AliCloudLogProject = 3000` in `CloudResourceKind` enum

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider, SLS project, log stores (iterated), and conditional full-text indexes
- **Terraform** (HCL): Same resources using `for_each` for stores and conditional indexes

### Tests
- Ginkgo/Gomega spec validation tests covering valid inputs, missing required fields, out-of-range values, and wrong api_version/kind

### Documentation
- README.md with configuration reference and related components
- examples.md with minimal, development, and production YAML examples

## Design Decisions

- **Index config**: Boolean `enable_index` per log store (default: true) with sensible full-text index defaults. Follows the 80/20 principle.
- **Tags**: Included `map<string,string> tags` for consistency with all existing provider components.
- **Default handling**: Proto optional fields with explicit default application in Go code (not relying on proto zero values).
- **Provider setup**: Region-only explicit config; credentials injected via environment variables by the runner.

## Dependencies Added

- `github.com/pulumi/pulumi-alicloud/sdk/v3 v3.95.0` added to go.mod

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (all spec validation tests)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
