# KubernetesStatefulSet: ConfigMap and Volume Mount Support

**Date**: December 21, 2025
**Type**: Enhancement
**Components**: API Definitions, Kubernetes Provider, Pulumi CLI Integration, Terraform Module

## Summary

Enhanced the KubernetesStatefulSet component to support creating ConfigMaps from inline content and mounting them (along with Secrets, HostPaths, EmptyDirs, and PVCs) as volumes into containers. This enables deploying stateful applications like databases that require both persistent storage and configuration files—a common production pattern for StatefulSets.

## Problem Statement / Motivation

Deploying stateful applications like PostgreSQL or MongoDB often requires configuration files mounted alongside persistent volumes. Previously, KubernetesStatefulSet had limited volume mount support through a component-specific type that only handled basic PVC mounts from volumeClaimTemplates.

### Pain Points

- **No declarative ConfigMap creation**: Users couldn't define configuration file content alongside their StatefulSet
- **Limited volume mount support**: Only basic name/mountPath/readOnly were available
- **Component-specific type**: `KubernetesStatefulSetContainerVolumeMount` was isolated from the shared volume mount ecosystem
- **Inconsistency with Deployments**: KubernetesDeployment had richer volume support after its recent enhancement
- **No support for Secrets, HostPaths, or EmptyDirs**: Missing common volume types needed for stateful workloads

## Solution / What's New

Added comprehensive volume mount support to KubernetesStatefulSet, including inline ConfigMap creation. The solution replaces the component-specific `KubernetesStatefulSetContainerVolumeMount` with the shared `VolumeMount` type from `volume_mount.proto`, ensuring consistency across all Kubernetes workload components.

### New Proto Fields

```protobuf
// In KubernetesStatefulSetSpec
map<string, string> config_maps = 9;  // Key=name, Value=content

// In KubernetesStatefulSetContainerApp (field 5 updated)
repeated dev.planton.provider.kubernetes.VolumeMount volume_mounts = 5;
```

### Supported Volume Types

| Volume Type | Use Case |
|-------------|----------|
| ConfigMap | Application configuration files (e.g., postgresql.conf) |
| Secret | TLS certificates, database credentials |
| HostPath | Node-level resources (logs, sockets) |
| EmptyDir | Temporary storage, inter-container sharing |
| PVC | Persistent data storage (integrated with volumeClaimTemplates) |

### StatefulSet-Specific Handling

StatefulSets have special volume handling for `volumeClaimTemplates`. When a PVC mount references a volumeClaimTemplate name, the StatefulSet controller manages the volume binding automatically. The implementation correctly detects this case and skips creating a separate volume definition.

## Implementation Details

### Proto Schema Changes

**File**: `apis/dev/planton/provider/kubernetes/kubernetesstatefulset/v1/spec.proto`

- Added import for `dev/planton/provider/kubernetes/volume_mount.proto`
- Added `config_maps` field (field 9) to `KubernetesStatefulSetSpec`
- Replaced `KubernetesStatefulSetContainerVolumeMount` with shared `VolumeMount` type
- Deleted the component-specific `KubernetesStatefulSetContainerVolumeMount` message

### Pulumi Module (Go)

**New file**: `iac/pulumi/module/configmap.go`
```go
func configMaps(ctx *pulumi.Context, locals *Locals, 
    kubernetesProvider pulumi.ProviderResource) (map[string]*kubernetescorev1.ConfigMap, error)
```

**Modified**: `iac/pulumi/module/statefulset.go`
- Added `buildVolumeMountsAndVolumes()` function processing all 5 volume types
- Added `isVolumeClaimTemplate()` helper to detect volumeClaimTemplate PVCs
- Container now includes proper `VolumeMounts` from shared type
- Pod spec now includes `Volumes` array (excluding volumeClaimTemplate PVCs)

**Modified**: `iac/pulumi/module/main.go`
- ConfigMaps created before StatefulSet with proper dependency ordering

### Terraform Module

**New file**: `iac/tf/configmap.tf`
```hcl
resource "kubernetes_config_map" "this" {
  for_each = try(var.spec.config_maps, {})
  # ...
}
```

**Modified**: `iac/tf/variables.tf`
- Enhanced `volume_mounts` to support all 5 volume sources
- Added `config_maps` variable

**Modified**: `iac/tf/statefulset.tf`
- Dynamic `volume_mount` blocks for container
- Dynamic `volume` blocks for each volume type (ConfigMap, Secret, HostPath, EmptyDir, PVC)
- Excludes PVC volumes that reference volumeClaimTemplates
- Proper dependency on `kubernetes_config_map.this`

### Test Updates

**Modified**: `spec_test.go`
- Updated test to use shared `kubernetes.VolumeMount` type with proper PVC reference

## Benefits

### For Application Developers

- **Single manifest deployment**: ConfigMaps, volumes, and persistent storage defined alongside the workload
- **Type-safe configuration**: Proto validation catches errors before deployment
- **Multi-IaC parity**: Same capabilities in Pulumi and Terraform

### For Platform Teams

- **Consistent patterns**: Volume mounts now use shared proto definitions across all Kubernetes workload types
- **Reduced maintenance**: Single source of truth for volume mount types
- **Ecosystem alignment**: KubernetesStatefulSet matches KubernetesDeployment capabilities

### Concrete Example

Deploy PostgreSQL with custom configuration:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesStatefulSet
metadata:
  name: postgres
spec:
  namespace:
    value: database
  createNamespace: true
  configMaps:
    postgres-config: |
      max_connections = 200
      shared_buffers = 256MB
      effective_cache_size = 768MB
  container:
    app:
      image:
        repo: postgres
        tag: "15"
      volumeMounts:
        - name: data
          mountPath: /var/lib/postgresql/data
          pvc:
            claimName: data
        - name: postgres-config
          mountPath: /etc/postgresql/postgresql.conf
          configMap:
            name: postgres-config
            key: postgres-config
            path: postgresql.conf
      ports:
        - name: postgres
          containerPort: 5432
          networkProtocol: TCP
          appProtocol: tcp
          servicePort: 5432
  volumeClaimTemplates:
    - name: data
      size: 10Gi
      accessModes: ["ReadWriteOnce"]
```

## Impact

### Who Is Affected

- **KubernetesStatefulSet users**: New capabilities for configuration management
- **Database deployments**: Can now include custom config files declaratively
- **Other workload components**: Pattern now consistent across Deployment, StatefulSet, and future DaemonSet, CronJob

### Backward Compatibility

✅ **Fully backward compatible**: All new fields are optional. Existing StatefulSets without `config_maps` or enhanced `volume_mounts` continue to work unchanged.

## Files Changed

| File | Change Type |
|------|-------------|
| `spec.proto` | Modified - Added import, config_maps field, use shared VolumeMount |
| `configmap.go` | Created - Pulumi ConfigMap creation |
| `statefulset.go` | Modified - Volume mount handling with volumeClaimTemplate detection |
| `main.go` | Modified - ConfigMap orchestration |
| `configmap.tf` | Created - Terraform ConfigMap resource |
| `variables.tf` | Modified - Enhanced volume_mounts and config_maps |
| `statefulset.tf` | Modified - Volume blocks with volumeClaimTemplate filtering |
| `spec_test.go` | Modified - Updated to use shared VolumeMount type |

## Validation

All validation steps passed:

```bash
✅ make protos        # Proto generation
✅ go build ./...     # Component build
✅ go test ./...      # Component tests
```

## Related Work

- **Shared Volume Mount Proto**: Uses `volume_mount.proto` for consistency
- **KubernetesDeployment Enhancement**: Same pattern applied earlier today (see `2025-12-21-042723-kubernetes-deployment-configmap-volume-mount-support.md`)
- **Future Enhancement**: Apply same pattern to KubernetesDaemonSet (see `kubernetesdaemonset/v1/_cursor/requirement.md`)

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation

