# KubernetesTemporal: Dynamic Configuration, History Shards, and Per-Service Resources

**Date**: December 26, 2025
**Type**: Enhancement
**Components**: API Definitions, Pulumi CLI Integration, Terraform IaC, Kubernetes Provider

## Summary

Expanded the `KubernetesTemporalSpec` protobuf schema to support dynamic configuration values, history shards, and per-service resource allocation. This enables fine-grained control over Temporal cluster behavior and resource management without modifying the underlying Helm chart.

## Problem Statement / Motivation

The existing `KubernetesTemporalSpec` provided basic deployment configuration but lacked controls for:

### Pain Points

- **Workflow history limits**: Workflows with many activities or large payloads hit default history size/count limits, causing failures without clear mitigation paths
- **Scalability tuning**: No way to configure history shards—a critical Day 0 decision that determines cluster parallelism
- **Resource allocation**: All Temporal services (frontend, history, matching, worker) used default resources with no way to right-size per service
- **Production readiness**: Teams had to fork Helm values files rather than configure through the declarative API

## Solution / What's New

Added three new configuration areas to `KubernetesTemporalSpec`:

### 1. Dynamic Configuration (`dynamic_config`)

Runtime settings that control Temporal server behavior:

```protobuf
message KubernetesTemporalDynamicConfig {
  optional int64 history_size_limit_error = 1;   // Max history size (bytes)
  optional int64 history_count_limit_error = 2;  // Max event count
  optional int64 history_size_limit_warn = 3;    // Warning threshold (size)
  optional int64 history_count_limit_warn = 4;   // Warning threshold (count)
}
```

### 2. History Shards (`num_history_shards`)

Immutable scalability setting with validation (1-16384):

```protobuf
optional int32 num_history_shards = 13;  // Default: 512
```

### 3. Per-Service Resources (`services`)

Independent configuration for each Temporal service:

```protobuf
message KubernetesTemporalServices {
  KubernetesTemporalServiceConfig frontend = 1;
  KubernetesTemporalServiceConfig history = 2;
  KubernetesTemporalServiceConfig matching = 3;
  KubernetesTemporalServiceConfig worker = 4;
}

message KubernetesTemporalServiceConfig {
  optional int32 replicas = 1;
  ContainerResources resources = 2;  // Reuses existing type
}
```

## Implementation Details

### Proto Schema Updates

**File**: `apis/dev/planton/provider/kubernetes/kubernetestemporal/v1/spec.proto`

- Added `KubernetesTemporalDynamicConfig` message with buf.validate rules (minimum values > 0)
- Added `KubernetesTemporalServiceConfig` message reusing `ContainerResources` from `kubernetes.proto`
- Added `KubernetesTemporalServices` message for per-service configuration
- Extended `KubernetesTemporalSpec` with fields 12-14 for dynamic_config, num_history_shards, and services

### Pulumi IaC Updates

**Files**:
- `helm_chart.go` - Maps new spec fields to Helm chart values
- `variables.go` - Added default constants

Key implementation patterns:

```go
// Dynamic config mapping
if dynamicConfig := input.KubernetesTemporal.Spec.DynamicConfig; dynamicConfig != nil {
    if dynamicConfig.HistorySizeLimitError != nil {
        helmValues["server"].(pulumi.Map)["config"].(pulumi.Map)["dynamicConfigValues"] = pulumi.Map{
            "limit.historySize.error": pulumi.Array{...}
        }
    }
}

// Service resources mapping with helper function
func buildResourcesMap(resources *kubernetes.ContainerResources) pulumi.Map
```

### Terraform IaC Updates

**Files**:
- `variables.tf` - Added nested object types for new configuration
- `locals.tf` - Extracted values with proper null handling
- `main.tf` - Added dynamic set blocks for Helm values

Pattern for dynamic configuration:

```hcl
dynamic "set" {
  for_each = local.dynamic_config_history_size_limit_error != null ? [1] : []
  content {
    name  = "server.config.dynamicConfigValues.limit\\.historySize\\.error[0].value"
    value = local.dynamic_config_history_size_limit_error
  }
}
```

### Validation Tests

**File**: `apis/dev/planton/provider/kubernetes/kubernetestemporal/v1/spec_test.go`

Added comprehensive test coverage:
- Dynamic config validation (positive values required)
- History shards validation (range 1-16384)
- Service configuration validation (positive replicas)

## Benefits

- **Production readiness**: Teams can deploy production-grade Temporal without forking Helm values
- **Self-service scalability**: Control history shards and service resources declaratively
- **Workflow resilience**: Increase history limits for complex workflows without code changes
- **Cost optimization**: Right-size each service based on actual workload characteristics
- **Consistent UX**: Configuration follows the same patterns as other deployment components

## Impact

### Who Benefits

- **Platform teams**: Configure Temporal clusters declaratively through the standard API
- **Application developers**: Increase history limits for complex workflow patterns
- **SREs**: Fine-tune resource allocation per service for cost and performance optimization

### Files Changed

| Area | Files |
|------|-------|
| Proto | `spec.proto`, `spec.pb.go`, `spec_pb.ts` |
| Pulumi | `helm_chart.go`, `variables.go` |
| Terraform | `variables.tf`, `locals.tf`, `main.tf` |
| Tests | `spec_test.go` |
| Docs | `examples.md` (4 new examples) |

## Usage Examples

### Increase History Limits

```yaml
spec:
  dynamic_config:
    history_size_limit_error: 104857600   # 100 MB
    history_count_limit_error: 102400     # 100K events
```

### Configure History Shards

```yaml
spec:
  num_history_shards: 1024  # Higher parallelism
```

### Per-Service Resources

```yaml
spec:
  services:
    frontend:
      replicas: 2
    history:
      replicas: 3
      resources:
        limits:
          cpu: "4000m"
          memory: "8Gi"
```

## Related Work

- **Issue**: `_issues/2025-12-26-104033.deployment-component.feat.expand-temporal-spec-dynamic-config.md`
- **Audit**: `docs/audit/2025-12-26-113019.md`

---

**Status**: ✅ Production Ready
**Timeline**: Single session

