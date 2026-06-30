# OCI Log Group Deployment Component

**Date**: February 21, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Implemented the OciLogGroup deployment component -- OCI's Logging service organizational container for related logs, bundled with individual log sub-resources supporting both service logs (auto-collected from OCI services) and custom logs (pushed via the Logging Ingestion API). Flattened the provider's nested `configuration > source` structure for clean YAML UX while reconstructing the nesting in IaC modules. Second and final resource of Phase 9 (Monitoring and Logging), completing the phase.

## Problem Statement / Motivation

The Planton Oracle Cloud provider needs logging infrastructure to enable centralized log collection for OCI resources. Log groups are the organizational container for OCI's Logging service -- without them, users cannot declaratively configure log collection from VCN flow logs, Object Storage audit trails, API Gateway access logs, or any other OCI service.

### Pain Points

- No logging component existed in the OCI provider catalog
- Teams running OCI workloads had no declarative way to provision log collection
- Phase 9 (Monitoring and Logging) was half-complete with only OciAlarm

## Solution / What's New

A complete deployment component (`OciLogGroup`) with proto API definitions, Pulumi module (Go), and Terraform module (HCL), registered as CloudResourceKind 3381. The component bundles `oci_logging_log_group` with zero or more `oci_logging_log` sub-resources.

### Key Design Decisions

**Flattened source configuration**: The OCI provider nests log source configuration as `configuration > source { service, resource, category, source_type, parameters }` with an additional `configuration.compartment_id` for source lookup override. We flatten this by removing the `source` wrapper (the only meaningful child of `configuration`) and hardcoding `source_type = "OCISERVICE"` (the only valid value) in IaC modules. This gives users a significantly cleaner YAML experience:

```yaml
logs:
  - displayName: vcn-flow-logs
    logType: service
    configuration:
      service: flowlogs
      resource:
        value: "ocid1.vcn.oc1..example"
      category: all
```

**LogType enum with CEL validation**: `unspecified = 0`, `custom = 1`, `service = 2`. The zero-value is rejected by a CEL rule (`log_type_required`), requiring explicit choice. A second CEL rule (`service_log_requires_configuration`) at the Log message level enforces that service logs must include a configuration block.

**Polymorphic resource reference**: The `resource` field in `ServiceLogConfiguration` uses `StringValueOrRef` without `default_kind` because the source resource could be any OCI component (VCN for flow logs, bucket for Object Storage, API gateway for access logs, etc.). The `valueFrom` mechanism still enables composability -- users specify the kind explicitly in their reference.

**Retention validation**: Optional `retention_duration` validated via CEL (`this >= 30 && this <= 180 && this % 30 == 0`) to enforce OCI's 30-day increment constraint at schema level rather than deferring to API errors.

## Implementation Details

### Proto API (4 files)

- **spec.proto**: 3 top-level fields, 1 embedded enum (LogType), 2 nested messages (Log, ServiceLogConfiguration), 3 CEL rules (log_type_required, service_log_requires_configuration, retention_30_day_increments)
- **api.proto**: Standard KRM wiring (OciLogGroup, OciLogGroupStatus)
- **stack_input.proto**: OciLogGroupStackInput with target + provider config
- **stack_outputs.proto**: 1 output (`log_group_id`)

### Spec Fields

**OciLogGroupSpec**:

| Field | Type | Notes |
|-------|------|-------|
| `compartment_id` | StringValueOrRef (required) | default_kind: OciCompartment |
| `description` | string | Optional log group description |
| `logs` | repeated Log | Bundled log sub-resources |

**Log** (nested):

| Field | Type | Notes |
|-------|------|-------|
| `display_name` | string (required) | for_each key, unique within group |
| `log_type` | LogType enum (required) | CEL: != unspecified |
| `is_enabled` | optional bool | nil = OCI default (true) |
| `retention_duration` | optional int32 | CEL: 30-day increments, 30-180 |
| `configuration` | ServiceLogConfiguration | Required for service logs |

**ServiceLogConfiguration** (nested inside Log):

| Field | Type | Notes |
|-------|------|-------|
| `service` | string (required) | OCI service name |
| `resource` | StringValueOrRef (required) | Source resource OCID (polymorphic) |
| `category` | string (required) | Log category |
| `parameters` | map<string, string> | Pass-through to OCI |
| `compartment_id` | StringValueOrRef | Override for source lookup |

### Validation Tests

26 Ginkgo/Gomega tests (13 valid, 13 invalid scenarios) covering minimal configuration, description, custom logs, service logs with full configuration, parameters, compartment override, is_enabled, all valid retention durations (30-180), mixed log types, valueFrom refs, full configuration, and all required-field and constraint validations.

### Pulumi Module (6 files)

- `iac/pulumi/main.go`: Entry point with stack input loading
- `module/main.go`: Resources orchestrator (`logGroupResource` -> `logResources`)
- `module/locals.go`: Locals struct with freeform tags from metadata labels
- `module/outputs.go`: Output constant (`log_group_id`)
- `module/log_group.go`: `logGroupResource()` creating `logging.NewLogGroup()` with conditional description
- `module/log.go`: `logResources()` looping over `spec.Logs`, creating `logging.NewLog()` with `DependsOn` log group; `buildLogConfiguration()` reconstructing provider's nested `LogConfigurationArgs > LogConfigurationSourceArgs` from flattened spec, hardcoding `SourceType: "OCISERVICE"`

### Terraform Module (6 files)

- `main.tf`: `oci_logging_log_group.this` with compartment_id, display_name, description, freeform_tags
- `log.tf`: `oci_logging_log.this` with `for_each` keyed by display_name, dynamic `configuration` block with nested `source` block, hardcoded `source_type = "OCISERVICE"`
- `locals.tf`: Freeform tags + `log_type_map` for enum conversion (custom -> CUSTOM, service -> SERVICE)
- `outputs.tf`: `log_group_id`
- `variables.tf`: Metadata and spec type definitions with nested optional configuration object
- `provider.tf`: OCI provider requirement (>= 5.0)

### Kind Registration

`OciLogGroup = 3381` registered under "Monitoring and Logging" section in `cloud_resource_kind.proto`, `kind_map_gen.go` regenerated.

## Benefits

- Enables declarative log collection from any OCI service via service logs
- Bundles log group + logs for atomic deployment of logging infrastructure
- Flattened source configuration reduces YAML verbosity while maintaining full functionality
- StringValueOrRef on resource field enables composability with any Planton OCI component
- Completes Phase 9 (Monitoring and Logging), providing both alerting (OciAlarm) and logging (OciLogGroup)

## Impact

- **Users**: Can now define log groups with service and custom logs through a single YAML manifest, enabling centralized log collection for VCN flow logs, Object Storage auditing, API Gateway access logs, and more
- **Platform**: Phase 9 (Monitoring and Logging) now 100% complete -- both resources done (OciAlarm, OciLogGroup)
- **Infra Charts**: All 5 planned OCI infra charts can now incorporate logging (e.g., serverless-stack chart references OciLogGroup)

## Related Work

- **OciAlarm** (R32): Monitoring counterpart, first resource of Phase 9
- **OciCompartment** (R04): Compartment referenced via `compartment_id`
- **OciVcn** (R01): VCN flow logs are a primary service log source
- **OciApiGateway** (R29): API Gateway access logs are another common source
- **OciObjectStorageBucket** (R21): Object Storage write/read audit logs

---

**Status**: Production Ready
