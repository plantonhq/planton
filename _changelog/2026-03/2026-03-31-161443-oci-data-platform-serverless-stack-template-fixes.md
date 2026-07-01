# Fix OCI data-platform and serverless-stack Chart Templates

**Date**: March 31, 2026
**Type**: Fix
**Provider**: OCI
**Chart(s)**: oci/data-platform, oci/serverless-stack

## Summary

Fixed field name and enum value mismatches in OCI `data-platform` and `serverless-stack` chart templates that caused `planton chart build` validation failures. The templates used boolean fields and bare enum values that don't exist in the protobuf spec; replaced them with the correct enum field names and qualified enum values.

## Problem Statement / Motivation

After the initial creation of 5 OCI InfraCharts, `planton chart build` passed for 3 charts (autonomous-db-stack, compute-environment, oke-environment) but failed for 2 charts (data-platform, serverless-stack) with protobuf validation errors.

### Pain Points

- `OciObjectStorageBucket` templates used `isVersioningEnabled: true` — a boolean field that doesn't exist; the spec uses a `versioning` enum (`enabled`, `disabled`, `suspended`)
- `OciObjectStorageBucket` templates used `isAutoTieringEnabled: true` — a boolean field that doesn't exist; the spec uses an `autoTiering` enum (`infrequent_access`, `auto_tiering_disabled`)
- `OciStreamPool` templates included a top-level `displayName` field that doesn't exist in `OciStreamPoolSpec` (the resource uses `metadata.name`)
- `OciStreamPool` templates used `autoCreateTopicsEnabled` instead of the correct field name `autoCreateTopicsEnable`
- `OciLogGroup` templates included a top-level `displayName` field that doesn't exist in `OciLogGroupSpec`
- `OciApiGateway` templates used `endpointType: public` — a bare value; the spec requires the fully-qualified `endpoint_type_public`

## Solution / What's New

Cross-referenced each failing resource template against its protobuf spec definition in the planton repo and corrected the field names and enum values.

### Files Changed

| Chart | Template | Fix |
|-------|----------|-----|
| oci/data-platform | `templates/storage.yaml` | `isVersioningEnabled: true` → `versioning: enabled`; `isAutoTieringEnabled: true` → `autoTiering: infrequent_access` |
| oci/data-platform | `templates/streaming.yaml` | Removed `displayName`; `autoCreateTopicsEnabled` → `autoCreateTopicsEnable` |
| oci/data-platform | `templates/monitoring.yaml` | Removed `displayName` from `OciLogGroup` |
| oci/serverless-stack | `templates/storage.yaml` | `isVersioningEnabled: true` → `versioning: enabled` |
| oci/serverless-stack | `templates/compute.yaml` | `endpointType: public` → `endpointType: endpoint_type_public` |
| oci/serverless-stack | `templates/monitoring.yaml` | Removed `displayName` from `OciLogGroup` |

## Implementation Details

### OciObjectStorageBucket — versioning and auto-tiering

The spec defines `Versioning` and `AutoTiering` as enums, not booleans:

```proto
Versioning versioning = 6;    // enabled | disabled | suspended
AutoTiering auto_tiering = 7; // infrequent_access | auto_tiering_disabled
```

Corrected YAML:

```yaml
versioning: enabled
autoTiering: infrequent_access
```

### OciStreamPool — no displayName, field name typo

`OciStreamPoolSpec` fields are: `compartmentId`, `kafkaSettings`, `kmsKeyId`, `privateEndpointSettings`, `streams`. There is no `displayName` — the resource name comes from `metadata.name`. Additionally, the Kafka setting field is `autoCreateTopicsEnable` (no trailing "d").

### OciLogGroup — no displayName

`OciLogGroupSpec` fields are: `compartmentId`, `description`, `logs`. There is no `displayName`.

### OciApiGateway — qualified enum value

The `EndpointType` enum requires fully-qualified values:

```proto
enum EndpointType {
  endpoint_type_unspecified = 0;
  endpoint_type_public = 1;
  endpoint_type_private = 2;
}
```

## Benefits

- All 5 OCI InfraCharts now pass `planton chart build` validation
- Templates are consistent with the protobuf spec definitions in planton

## Impact

Users provisioning OCI data-platform or serverless-stack environments can now build and deploy these charts without validation errors.

---

**Status**: ✅ Production Ready
