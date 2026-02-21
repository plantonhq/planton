# AliCloudSaeApplication Component Added

**Date**: 2026-02-20
**Component**: AliCloudSaeApplication
**Enum**: 3111
**ID Prefix**: acsae

## Summary

Added the AliCloudSaeApplication deployment component -- manages Serverless App Engine (SAE) applications in Alibaba Cloud. SAE is a container-based serverless platform that supports deploying applications as container images, JAR/WAR packages, or Python/PHP ZIP archives. The component covers the core application lifecycle including VPC placement, health checks, rolling deployments, and environment variable management.

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudsaeapplication/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AliCloudSaeApplication = 3111` in `CloudResourceKind` enum under the Serverless category
- 8 protobuf message types: spec, health check (with HttpGet/TcpSocket/Exec sub-types), custom host alias, update strategy (with batch update config)

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and a single `sae.Application` resource. Converts `envs` map to the JSON array format SAE expects. Maps all optional fields to v2 API variants (CommandArgsV2s, LivenessV2, ReadinessV2, CustomHostAliasV2s, UpdateStrategyV2). StringValueOrRef support for VPC/VSwitch/SecurityGroup cross-component wiring.
- **Terraform** (HCL): Single `alicloud_sae_application` resource with dynamic blocks for liveness_v2, readiness_v2, custom_host_alias_v2, and update_strategy_v2. Environment variables converted from map to JSON array in locals. Input validations for package_type, cpu, and memory tiers.

### Tests
- Ginkgo/Gomega spec validation tests: 27 specs covering valid inputs (minimal Image, FatJar with JDK, War, PythonZip, VPC config, HTTP/TCP/Exec health checks, environment variables, custom host aliases, update strategy, all CPU tiers, all memory tiers, full production config) and invalid inputs (missing required fields, name too long, invalid package_type, zero replicas, invalid cpu/memory tiers, wrong api_version/kind, missing metadata, out-of-range termination grace period, invalid programming_language, invalid update strategy type, invalid release type)

### Documentation
- README.md with package type matrix, stack outputs, and related components
- examples.md with 4 YAML examples (minimal Image, Java FatJar with VPC/health checks, Python microservice, ACR EE container with full production config)
- catalog-page.md with complete configuration reference tables for all fields including health checks and update strategy

## Design Decisions (Deviations from T02)

- **VPC/VSwitch optional**: T02 marked vpc_id and vswitch_id as required. Made them optional StringValueOrRef since SAE supports managed networking without VPC.
- **Namespace not bundled**: T02 suggested auto-creating namespaces. Simplified to plain optional string `namespace_id` -- default namespace used when omitted.
- **Expanded memory tiers**: Added 12288, 24576, 131072 MB tiers that T02 did not include but the actual provider supports.
- **Expanded package types**: Added PythonZip and PhpZip beyond T02's Image/FatJar/War.
- **command_args as list**: T02 had single string; implemented as `repeated string` mapping to provider's v2 `command_args_v2` field.
- **envs as map with JSON conversion**: T02's map<string,string> approach maintained; IaC modules handle the JSON array conversion that the SAE API requires.
- **Health checks with proper sub-types**: Modeled after the provider's v2 health check structure with separate HttpGet, TcpSocket, and Exec action types.
- **Update strategy with batch config**: Exposed the full batch update configuration (batch count, wait time, release type) rather than the simplified version in the plan.
- **Skipped niche fields**: PHP-specific, WAR-specific, OSS mounts, ConfigMap mounts, NAS configs, Kafka configs, grey tag routing, AHAS, micro-registration, PVTZ discovery, lifecycle hooks (post_start/pre_stop), Tomcat config. All can be added in v2.

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (27/27 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
