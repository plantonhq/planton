# OCI Functions Application Deployment Component

**Date**: 2026-02-20
**Type**: New Feature
**Components**: `apis/org/openmcf/provider/oci/ocifunctionsapplication/v1/`

## Summary

Added the OciFunctionsApplication deployment component -- OCI's organizational container for serverless functions with subnet networking, processor architecture selection (x86/ARM/multi-arch), image signature verification via KMS keys, and APM distributed tracing integration. First resource of Phase 7 (Serverless and Functions). Resource R28 in the OCI provider expansion.

## Problem Statement / Motivation

OpenMCF's OCI provider had comprehensive infrastructure coverage across networking, compute, containers, databases, storage, and security (Phases 1-6, 27 resources), but no serverless function support. OCI Functions is Oracle's managed serverless platform (similar to AWS Lambda) where functions run in user-specified subnets with optional NSG controls. Without a declarative Functions Application component, platform teams cannot provision the execution environment that functions are deployed into, blocking the Serverless Stack infra chart.

## Solution / What's New

A complete OciFunctionsApplication deployment component with both Pulumi (Go) and Terraform (HCL) modules.

### Proto API

- **spec.proto**: 9 fields, 1 enum (Shape), 3 nested messages (ImagePolicyConfig, ImagePolicyKeyDetail, TraceConfig), 1 CEL validation rule
- **Key fields**: compartment_id (StringValueOrRef), subnet_ids (repeated StringValueOrRef, min_items=1), display_name, shape (enum), config (map<string,string>), network_security_group_ids (repeated StringValueOrRef), syslog_url, image_policy_config, trace_config
- **CEL rule**: When image_policy_config.is_policy_enabled is true, key_details must be non-empty
- **api.proto**: Standard KRM wrapper with const-validated api_version and kind
- **stack_outputs.proto**: 1 output (application_id)

### Design Decisions

- **Shape as proto enum**: Three values (generic_x86, generic_arm, generic_x86_arm) mapping to GENERIC_X86/GENERIC_ARM/GENERIC_X86_ARM via IaC maps. Enum provides type safety and clean YAML UX (`shape: generic_arm`). ForceNew in provider.
- **subnet_ids as repeated StringValueOrRef**: Follows OciApplicationLoadBalancer pattern for infra-chart composability. At least one subnet required (min_items=1). ForceNew in provider.
- **image_policy_config with CEL enforcement**: Enabling verification without any KMS keys is a config error. CEL catches it at validation time rather than at deployment time.
- **trace_config.domain_id as plain string**: APM domains are not modeled as OpenMCF components, so StringValueOrRef with default_kind would be misleading.
- **config map pass-through**: OCI enforces the 4KB limit and key format constraints server-side. No proto-level validation needed for key/value format.
- **Single output (application_id)**: The application OCID is the primary composability value for downstream `fn deploy` and CI/CD pipelines.
- **Directory name**: `ocifunctionsapplication` (per WA02 -- lowercased kind name, not id_prefix `ocifnapp`).

### Bundled Resources

1. **Functions Application** -- the organizational container with shared networking, config, and execution architecture for all functions deployed within it

### Pulumi Module (Go)

5 files across the module package:
- `main.go` -- orchestrator calling applicationResource() with OCI provider setup, shapeMap enum conversion
- `locals.go` -- Locals struct with freeform tags and display name fallback
- `application.go` -- functions.NewApplication() with conditional shape, config map, NSG IDs, image policy config builder, and trace config builder
- `outputs.go` -- 1 output constant (application_id)

### Terraform Module (HCL)

5 files:
- `main.tf` -- oci_functions_application.this with dynamic image_policy_config (nested dynamic key_details) and dynamic trace_config blocks
- `locals.tf` -- display_name fallback, freeform tags, shape_map enum conversion
- `variables.tf`, `outputs.tf`, `provider.tf`

### Validation Tests

26 Ginkgo/Gomega tests (16 valid, 10 invalid scenarios) covering:
- Minimal application (compartment_id + 1 subnet)
- With display_name, each shape value, config map, NSG IDs, syslog_url
- Trace config enabled/disabled, image policy config enabled with keys, disabled without keys
- Multiple subnets, full configuration with all fields
- StringValueOrRef with literal and valueFrom patterns
- Required field validation (compartment_id, subnet_ids non-empty, metadata, spec)
- CEL validation (image policy enabled with empty/nil key_details)
- Missing kms_key_id in key_details

### Kind Registration

- **Enum**: OciFunctionsApplication = 3360
- **ID Prefix**: ocifnapp
- **Section**: Serverless and Functions (new section in cloud_resource_kind.proto)

## Benefits

- Enables declarative provisioning of OCI serverless function environments
- Unblocks the Serverless Stack infra chart (Chart 4)
- Full attribute coverage: networking, architecture selection, image verification, config, tracing
- Infra-chart composable via StringValueOrRef on subnet_ids, NSG IDs, and KMS key references

## Impact

- **Platform teams**: Can now declaratively create function application environments with proper networking isolation and security controls
- **OCI provider coverage**: 28/37 resources complete (75.7%), Phase 7 started
- **Infra charts**: Serverless Stack chart prerequisites advancing (needs R29 OciApiGateway to complete)

## Validation Results

- `go build` -- clean
- `go vet` -- clean
- `go test` -- 26/26 passed
- `terraform validate` -- success

---

**Status**: Production Ready
