# AlicloudFunction Component Added

**Date**: 2026-02-20
**Component**: AlicloudFunction
**Enum**: 3110
**ID Prefix**: acfc

## Summary

Added the AlicloudFunction deployment component -- manages Function Compute v3 functions in Alibaba Cloud. FC v3 uses a service-less model where functions are standalone top-level resources. The component supports all major runtime families (Python, Node.js, Java, Go, PHP, .NET), custom runtimes, custom container images, and GPU-accelerated workloads.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudfunction/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AlicloudFunction = 3110` in `CloudResourceKind` enum under a new Serverless category
- 12 protobuf message types covering the function spec and all nested configurations (code, VPC, logging, custom container, custom runtime, health check, lifecycle hooks, NAS, GPU)

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and a single `fc.V3Function` resource with full field mapping for all nested config blocks (code, vpc_config, log_config, custom_container_config, custom_runtime_config, instance_lifecycle_config, nas_config, gpu_config)
- **Terraform** (HCL): Single `alicloud_fcv3_function` resource with dynamic blocks for all optional nested configurations, runtime validation, and tag merging

### Tests
- Ginkgo/Gomega spec validation tests: 33 specs covering valid inputs (minimal, compute sizing, OSS code, VPC config, log config, custom container with health check, custom runtime, lifecycle hooks, NAS mount, GPU config, all 18 runtimes, layers with role, boundary values) and invalid inputs (missing required fields, invalid runtime, wrong api_version/kind, missing metadata, out-of-range compute values, invalid gpu_type, invalid log_begin_rule, empty container image, invalid health check values, empty NAS mount_dir)

### Documentation
- README.md with configuration reference tables, runtime matrix, and related components
- examples.md with 4 YAML examples (minimal Python, production API with VPC/logging, custom container with health check, GPU-accelerated AI inference)
- catalog-page.md with full catalog documentation including quick start, nested config block reference, and 3 deployment examples

## Design Decisions (Deviations from T02)

- **Renamed to AlicloudFunction**: T02 used `AlicloudFcFunction`; simplified to `AlicloudFunction` since FC is an internal service abbreviation. ID prefix `acfc` retained.
- **FC v3 confirmed**: DD04 fallback to v2 was not needed -- both TF (`alicloud_fcv3_function`) and Pulumi (`fc.V3Function`) fully support v3.
- **Significantly expanded spec vs T02**: T02 had ~10 fields; actual spec has 24 top-level fields and 12 message types covering FC v3's full capability surface (custom containers, custom runtimes, GPU, NAS, lifecycle hooks).
- **Removed options.proto import**: FC v3 defaults are provider-computed (dynamic based on memory/cpu ratio), not static -- no `(org.openmcf.shared.options.default)` annotations.
- **Triggers not bundled**: Per DD07, triggers have independent lifecycles and varied event sources. They are not bundled with the function.
- **Excluded niche fields**: `oss_mount_config`, `custom_dns`, `session_affinity`, `idle_timeout`, `invocation_restriction`, `tracing_config`, `instance_isolation_mode` -- can be added later without breaking changes.
- **Shared AlicloudFunctionHealthCheckConfig**: Both custom container and custom runtime use the same health check proto message; Pulumi SDK has separate types but the fields are identical.

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (33/33 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
- `go build ./pkg/crkreflect/...` -- PASS (kind map regenerated)
