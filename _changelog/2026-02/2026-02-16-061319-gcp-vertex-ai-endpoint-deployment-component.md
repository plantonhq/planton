# GcpVertexAiEndpoint Deployment Component

**Date**: February 16, 2026
**Type**: Feature
**Components**: API Definitions, GCP Provider, Pulumi CLI Integration, Provider Framework

## Summary

Added a complete GcpVertexAiEndpoint deployment component for provisioning Vertex AI model serving endpoints with three networking modes (public, VPC-peered, Private Service Connect), optional CMEK encryption, optional dedicated DNS, and an auto-generated numeric endpoint identifier. This is resource R20 in the GCP expansion project, bringing the total GCP resource count to 40.

## Problem Statement / Motivation

ML teams need infrastructure-as-code management for Vertex AI Endpoints -- the serving surfaces where trained models are deployed for prediction. Without this component, endpoint creation is either manual (console/gcloud), one-off Terraform, or custom Pulumi code -- none of which integrate with Planton's composable infrastructure model.

### Pain Points

- No way to compose Vertex AI endpoints with other Planton-managed resources (VPCs, KMS keys, projects) using `valueFrom` references
- Manual endpoint creation doesn't benefit from framework labels, consistent naming, or infra-chart composition
- Three networking modes (public, VPC-peered, PSC) need to be properly abstracted with mutual exclusion validation
- Vertex AI's numeric-only endpoint name requirement is unusual and needs to be handled transparently

## Solution / What's New

A full deployment component following the established Planton forge pattern: 4 proto files, dual IaC modules (Pulumi Go + Terraform HCL), 31 validation tests, documentation, catalog page, and 3 presets.

### Key Design Choices

- **Flattened `encryption_spec`** to a top-level `kms_key_name` field (single-field wrapper adds no value)
- **PSC as a sub-message** (not a bare boolean) to include `project_allowlist` for access control
- **`endpoint_name` exposed as optional** with CEL validation for numeric-only format; IaC modules auto-generate when omitted
- **`dedicated_endpoint_enabled` added** (not in original plan) for production-grade dedicated DNS endpoints
- **Excluded operational concerns**: `traffic_split`, `predict_request_response_logging_config`, `model_deployment_monitoring_job` -- model deployment is a separate step

## Implementation Details

### Proto API

```
apis/dev/planton/provider/gcp/gcpvertexaiendpoint/v1/
├── spec.proto          # 9 fields, 1 sub-message, 3 CEL validations
├── stack_outputs.proto # 4 outputs
├── api.proto           # KRM envelope
└── stack_input.proto   # target + provider config
```

3 StringValueOrRef fields enable infra-chart composition:
- `project_id` → GcpProject
- `network` → GcpVpc
- `kms_key_name` → GcpKmsKey

3 CEL validations enforce mutual exclusion:
- `network` vs `private_service_connect_config`
- `dedicated_endpoint_enabled` vs `private_service_connect_config`
- `endpoint_name` numeric-only regex

### Pulumi Module

```
iac/pulumi/module/
├── main.go          # Resources() entry point
├── locals.go        # Labels, display name initialization
├── ai_endpoint.go   # vertex.NewAiEndpoint with conditional blocks
└── outputs.go       # 4 output constants
```

### Terraform Module

```
iac/tf/
├── provider.tf   # google ~> 6.0, random ~> 3.0
├── variables.tf  # Typed variable blocks matching spec
├── locals.tf     # Framework labels, endpoint name resolution
├── main.tf       # google_vertex_ai_endpoint + random_integer
└── outputs.tf    # 4 outputs
```

Notable: The Terraform module uses `random_integer` with keepers to auto-generate a stable 10-digit numeric endpoint name when `endpoint_name` is not provided. Vertex AI is the only GCP resource that requires numeric-only names.

### Validation Tests

31 tests covering all spec fields and validation rules:
- 16 positive cases (minimal, all networking modes, CMEK, PSC with allowlist, endpoint name patterns, boundary testing)
- 15 negative cases (missing required fields, mutual exclusions, invalid endpoint names, wrong api_version/kind)

## Benefits

- ML teams can provision Vertex AI endpoints through the same Planton workflow as all other infrastructure
- Three networking modes properly abstracted with proto-level validation preventing invalid combinations
- `valueFrom` composition enables endpoint creation in infra charts alongside VPCs, KMS keys, and notebooks
- Numeric endpoint name is auto-generated -- users never need to think about it unless importing existing endpoints

## Impact

- **GCP resource count**: 40 (19 original + 21 new from expansion project)
- **R20 of 23** in the GCP expansion project
- **Files**: 41 files, 2,837 lines added
- **Tests**: 31 passing
- **Build**: `go build`, `go test`, `terraform validate` all pass

## Related Work

- Part of project 20260215.01.sp.gcp-resource-expansion (R20 of 23)
- Companion to GcpVertexAiNotebook (R19) for ML infrastructure
- Builds on KMS (R03/R04) and VPC (existing) for encryption and networking composition

---

**Status**: Production Ready
