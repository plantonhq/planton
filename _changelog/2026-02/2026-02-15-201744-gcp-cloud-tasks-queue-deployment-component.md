# GCP Cloud Tasks Queue Deployment Component

**Date**: February 15, 2026
**Type**: Feature
**Components**: GCP Provider, API Definitions, Pulumi Module, Terraform Module, Documentation

## Summary

Added the GcpCloudTasksQueue deployment component to OpenMCF, enabling provisioning and management of GCP Cloud Tasks queues with configurable dispatch rate limits, retry policies, and queue-level HTTP target configuration with OIDC or OAuth authentication. This is the 18th GCP resource added in the cloud provider expansion initiative (R17 in the queue).

## Problem Statement / Motivation

Cloud Tasks is a core GCP service for asynchronous task dispatch, used extensively in microservices architectures for background processing, rate-limited API calls, and reliable HTTP delivery. Without this component, OpenMCF users managing GCP infrastructure had to configure task queues outside the framework, breaking the declarative, composable infrastructure model.

### Pain Points

- No way to declare Cloud Tasks queues alongside other GCP infrastructure in OpenMCF
- Queue-level OIDC/OAuth authentication for Cloud Run dispatch required manual configuration
- Rate limits and retry policies could not be versioned and composed in infra charts
- No cross-resource dependency wiring (e.g., queue -> service account -> Cloud Run service)

## Solution / What's New

A complete deployment component following the OpenMCF forge standard:

### Proto API (4 proto files, 10 message types)

- `spec.proto` with 8 top-level fields and 8 sub-messages covering HTTP target, rate limits, retry config, and logging
- 3 `StringValueOrRef` fields: `project_id` (GcpProject), `oauth_token.service_account_email` (GcpServiceAccount), `oidc_token.service_account_email` (GcpServiceAccount)
- 7 CEL validations: queue_name pattern, desired_state enum, http_method enum, scheme enum, enforce_mode enum, sampling_ratio range, oauth/oidc mutual exclusion
- `stack_outputs.proto` with queue_id, queue_name, state

### Pulumi Module (4 Go files)

- Creates `cloudtasks.NewQueue` with conditional blocks for all optional configurations
- Maps flattened URI path/query fields back to the SDK's nested structure
- Exports queue_id, queue_name, and state

### Terraform Module (6 files)

- Provider `~> 6.0`, dynamic blocks for http_target, rate_limits, retry_config, logging
- Nested dynamic blocks within http_target for oauth_token, oidc_token, uri_override, header_overrides
- Feature parity with Pulumi (except `state` output not available in Terraform)

### Validation Tests (48 total)

- 26 positive cases covering all field values, all HTTP methods, auth modes, URI overrides, full-featured spec
- 22 negative cases covering missing required fields, invalid enums, mutual exclusion, name pattern violations, boundary values

### Documentation

- README with feature overview and configuration guide
- 7 YAML examples from minimal to full-featured with valueFrom references
- Research documentation with design decisions, comparisons, and integration patterns
- Catalog page following the AWS ALB exemplar standard
- 3 presets: basic-queue, rate-limited-processing, secure-cloud-run-target

## Implementation Details

### Key Design Decisions (Corrections to T01 Plan)

1. **Added `queue_name` field** -- Consistent with R01-R16 naming pattern. GCP requires an explicit name.
2. **Added `http_target` block** -- Major feature NOT in original plan. Queue-level HTTP target is the modern Cloud Tasks pattern for Cloud Run/Functions integration.
3. **Added `desired_state`** -- RUNNING/PAUSED operational control for maintenance scenarios.
4. **Corrected `max_burst_size`** -- Plan listed it as an input; it's computed-only by GCP. Excluded from spec.
5. **Excluded `app_engine_routing_override`** -- Legacy App Engine routing in decline. Reduces spec complexity for a niche use case.
6. **Flattened nested structures** -- `uri_override.path_override.path` -> `uri_override.path`, `query_override.query_params` -> `query_params`, header wrapper removed.
7. **No GCP labels** -- Cloud Tasks queues do not support GCP labels (GCP API limitation).

### StringValueOrRef for Infra-Chart Composability

Three StringValueOrRef fields enable infra-chart composition:
- `project_id` -> GcpProject for project wiring
- `oauth_token.service_account_email` -> GcpServiceAccount for OAuth auth
- `oidc_token.service_account_email` -> GcpServiceAccount for OIDC auth

This enables the common infra-chart pattern: dedicated service account -> Cloud Tasks queue with auth -> Cloud Run service.

## Benefits

- **Declarative queue management** -- Cloud Tasks queues defined as code alongside other GCP infrastructure
- **Composable authentication** -- Service account emails wired via StringValueOrRef for dependency-aware infra charts
- **Modern HTTP pattern** -- Queue-level OIDC auth for Cloud Run is the recommended GCP architecture
- **Production-ready defaults** -- GCP-managed defaults when rate_limits and retry_config are omitted
- **Dual IaC** -- Both Pulumi and Terraform implementations with feature parity

## Impact

- GCP provider coverage expanded from 37 to 38 resource kinds
- Messaging category now covers Pub/Sub (topic + subscription) and Cloud Tasks
- 5 remaining resources in the GCP expansion queue: GcpCloudSchedulerJob, GcpVertexAiNotebook, GcpVertexAiEndpoint, GcpCloudArmorPolicy

## Related Work

- Part of the GCP Resource Expansion initiative (20260215.01.sp.gcp-resource-expansion)
- Follows R16 GcpFilestoreInstance
- Next: R18 GcpCloudSchedulerJob (which can target Cloud Tasks queues)

---

**Status**: Production Ready
