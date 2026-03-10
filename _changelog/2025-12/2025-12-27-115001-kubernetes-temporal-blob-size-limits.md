# KubernetesTemporal: Blob/Payload Size Limits for Large Workflows

**Date**: December 27, 2025
**Type**: Enhancement
**Components**: KubernetesTemporal, Protobuf Schema, Pulumi Module, Terraform Module, Documentation

## Summary

Added blob/payload size limit configuration to the KubernetesTemporal component, enabling workflows to handle large payloads (markers, signals, activity I/O) up to 10MB+. This addresses `BadRecordMarkerAttributes.Details exceeds size limit` errors when Temporal workflows process large data like IaC diffs.

## Problem Statement / Motivation

Temporal workflows sending large payloads (e.g., Pulumi diffs from the iac-runner) were failing with:

```
BadRecordMarkerAttributes: RecordMarkerCommandAttributes.Details exceeds size limit
```

### Pain Points

- **Default 2MB limit too restrictive**: Temporal's default blob size limit of 2MB is insufficient for workflows handling large IaC diffs, API responses, or file contents
- **No existing configuration path**: The KubernetesTemporal spec only exposed history size limits, not individual payload/blob size limits
- **Confusion with history limits**: The `historySizeLimitError` controls total workflow history size, not individual payload sizes—a common misconception
- **Production failures**: Large Pulumi preview diffs were causing workflow terminations in production

## Solution / What's New

Extended the KubernetesTemporal dynamic configuration with two new fields:

| Field | Purpose | Default | Recommended |
|-------|---------|---------|-------------|
| `blobSizeLimitError` | Maximum bytes for a single payload before rejection | 2MB | 10MB for IaC workflows |
| `blobSizeLimitWarn` | Warning threshold before hitting the error limit | 512KB | 5MB |

### Dynamic Configuration Mapping

The new fields map to Temporal's dynamic configuration:

```yaml
# Proto field → Helm chart value
blobSizeLimitError → limit.blobSize.error
blobSizeLimitWarn  → limit.blobSize.warn
```

## Implementation Details

### Proto Schema Changes

Added two new optional fields to `KubernetesTemporalDynamicConfig`:

```protobuf
message KubernetesTemporalDynamicConfig {
  // ... existing history limits ...
  
  /**
   * Maximum size in bytes for a single blob/payload (marker details, signal data, activity I/O).
   * When a payload exceeds this limit, Temporal rejects it with "exceeds size limit" error.
   * Default: 2097152 (2 MB). Increase for workflows that send large payloads like IaC diffs.
   */
  optional int64 blob_size_limit_error = 5 [(buf.validate.field).int64 = {gte: 1048576}];

  /**
   * Warning threshold for blob/payload size in bytes.
   * Temporal logs warnings when payloads approach this limit.
   * Default: 524288 (512 KB).
   */
  optional int64 blob_size_limit_warn = 6 [(buf.validate.field).int64 = {gte: 262144}];
}
```

**Validation rules**:
- `blobSizeLimitError`: Minimum 1MB (1048576 bytes)
- `blobSizeLimitWarn`: Minimum 256KB (262144 bytes)

### Pulumi Module Changes

Updated `helm_chart.go` to map the new fields to Temporal Helm chart values:

```go
if dynamicConfig.BlobSizeLimitError != nil {
    dynamicConfigValues["limit.blobSize.error"] = pulumi.Array{
        pulumi.Map{"value": pulumi.Int(*dynamicConfig.BlobSizeLimitError)},
    }
}

if dynamicConfig.BlobSizeLimitWarn != nil {
    dynamicConfigValues["limit.blobSize.warn"] = pulumi.Array{
        pulumi.Map{"value": pulumi.Int(*dynamicConfig.BlobSizeLimitWarn)},
    }
}
```

### Terraform Module Changes

Added corresponding variables to `variables.tf`:

```hcl
dynamic_config = optional(object({
  // ... existing fields ...
  blob_size_limit_error = optional(number)
  blob_size_limit_warn  = optional(number)
}))
```

### Files Changed

| File | Change |
|------|--------|
| `spec.proto` | Added `blob_size_limit_error` and `blob_size_limit_warn` fields |
| `spec.pb.go` | Generated Go stubs |
| `spec_pb.ts` | Generated TypeScript stubs |
| `helm_chart.go` | Map new fields to Helm chart dynamic config |
| `variables.tf` | Add Terraform variable definitions |
| `examples.md` | Add usage examples for blob size limits |

## Benefits

- **Unblocks large workflows**: IaC diffs, API responses, and file contents can now exceed 2MB
- **Configurable per deployment**: Each Temporal instance can have appropriate limits for its workload
- **Warning before failure**: The warn threshold provides early visibility before hitting hard limits
- **Consistent with existing pattern**: Follows the same structure as history size limits

## Impact

### Who's Affected

- **IaC workflows**: Pulumi preview/up operations with large diffs
- **Data processing workflows**: Workflows handling large API responses or file contents
- **Signal-heavy workflows**: Workflows using large signal payloads

### Upgrade Path

Existing deployments continue to work unchanged. To increase limits:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesTemporal
metadata:
  name: temporal-prod
spec:
  dynamicConfig:
    blobSizeLimitError: 10485760   # 10 MB
    blobSizeLimitWarn: 5242880     # 5 MB
```

## Usage Example

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesTemporal
metadata:
  name: temporal-large-payloads
spec:
  target_cluster:
    cluster_name: prod-gke-cluster
  namespace:
    value: temporal-prod
  create_namespace: true
  database:
    backend: postgresql
    external_database:
      host: postgres-db.example.com
      port: 5432
      username: temporaluser
      password:
        secretRef:
          name: temporal-db-credentials
          key: password
  dynamic_config:
    # Increase blob size limit to 10 MB (default is 2 MB)
    blob_size_limit_error: 10485760
    # Set warning threshold at 5 MB
    blob_size_limit_warn: 5242880
```

## Related Work

- **Preceding changelog**: `planton/_changelog/2025-12/2025-12-26-132601-kubernetes-temporal-web-console-dynamic-config-services.md` - Added initial history size limit support
- **Web Console**: Updated separately to expose these fields in forms and details pages
- **Production manifest**: `planton/ops/.../temporal.yaml` updated with 10MB limit

## Best Practices

While increasing blob size limits is supported, Temporal is not optimized for extremely large payloads. For optimal performance:

1. **Store large data externally**: Use S3, GCS, or a database for large payloads
2. **Pass references**: Store the data and pass a reference/URL in the workflow
3. **Compress when possible**: Reduce payload size before sending to Temporal
4. **Monitor warning threshold**: Set `blobSizeLimitWarn` to get early visibility

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation

