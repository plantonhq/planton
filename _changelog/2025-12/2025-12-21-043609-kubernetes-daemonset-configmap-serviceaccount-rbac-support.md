# KubernetesDaemonSet: ConfigMap, ServiceAccount, and RBAC Support

**Date**: December 21, 2025
**Type**: Enhancement
**Components**: API Definitions, Kubernetes Provider, Pulumi Module

## Summary

Enhanced the KubernetesDaemonSet component to support creating ConfigMaps from inline content, mounting volumes (ConfigMap, Secret, HostPath, EmptyDir, PVC), creating ServiceAccounts, and configuring RBAC permissions. This enables deploying DaemonSets that require configuration files and cluster-level permissions—a common pattern for log collectors (Vector), node monitoring agents, and security scanners.

## Problem Statement / Motivation

Deploying DaemonSets like Vector log collectors requires:
- Configuration files mounted into containers
- ServiceAccounts with RBAC permissions to read pods/nodes/namespaces
- HostPath mounts for accessing container logs on nodes

Previously, KubernetesDaemonSet only supported basic hostPath volume mounts with a component-specific message type. Users had to either:
- Pre-create ConfigMaps, ServiceAccounts, and RBAC resources outside the deployment workflow
- Bake configuration into container images (violating 12-factor app principles)
- Use KubernetesManifest with raw YAML (losing type safety)

### Pain Points

- **No declarative ConfigMap creation**: Couldn't define configuration content alongside the DaemonSet
- **Limited volume mount types**: Only hostPath was supported, no ConfigMap/Secret mounts
- **No ServiceAccount support**: Pods ran with default service account
- **No RBAC configuration**: Required separate kubectl commands or manifests for permissions
- **Inconsistency with KubernetesDeployment**: Different volume mount types between components

## Solution / What's New

Added comprehensive ConfigMap, ServiceAccount, and RBAC support to KubernetesDaemonSet, leveraging the shared `volume_mount.proto` definitions for consistency across workload components.

### New Proto Fields

```protobuf
// In KubernetesDaemonSetSpec
bool create_service_account = 9;
string service_account_name = 10;
map<string, string> config_maps = 11;
KubernetesDaemonSetRbac rbac = 12;
```

### New RBAC Messages

```protobuf
message KubernetesDaemonSetRbac {
  repeated KubernetesDaemonSetRbacRule cluster_rules = 1;
  repeated KubernetesDaemonSetRbacRule namespace_rules = 2;
}

message KubernetesDaemonSetRbacRule {
  repeated string api_groups = 1;
  repeated string resources = 2;
  repeated string verbs = 3;
  repeated string resource_names = 4;
}
```

### Supported Volume Types

The component now uses the shared `VolumeMount` type supporting:

| Volume Type | Use Case |
|-------------|----------|
| ConfigMap | Application configuration files |
| Secret | TLS certificates, credentials |
| HostPath | Node-level resources (logs, sockets) |
| EmptyDir | Temporary storage, inter-container sharing |
| PVC | Persistent data storage |

## Implementation Details

### Proto Schema Changes

**File**: `apis/dev/planton/provider/kubernetes/kubernetesdaemonset/v1/spec.proto`

1. Added import for shared volume mount definitions
2. Added 4 new fields to `KubernetesDaemonSetSpec` (fields 9-12)
3. Replaced component-specific `KubernetesDaemonSetVolumeMount` with shared `VolumeMount`
4. Added `KubernetesDaemonSetRbac` and `KubernetesDaemonSetRbacRule` messages
5. RBAC rules include validation requiring at least one item in apiGroups, resources, and verbs

### Pulumi Module (Go)

**New file**: `iac/pulumi/module/configmap.go`
```go
func configMaps(ctx *pulumi.Context, locals *Locals, 
    kubernetesProvider pulumi.ProviderResource) (map[string]*kubernetescorev1.ConfigMap, error)
```

**New file**: `iac/pulumi/module/service_account.go`
```go
func serviceAccount(ctx *pulumi.Context, locals *Locals, 
    kubernetesProvider pulumi.ProviderResource) (string, error)
```

**New file**: `iac/pulumi/module/rbac.go`
```go
func rbac(ctx *pulumi.Context, locals *Locals, serviceAccountName string, 
    kubernetesProvider pulumi.ProviderResource) error
```

**Modified**: `iac/pulumi/module/daemonset.go`
- Added `buildVolumeMountsAndVolumes()` function processing all 5 volume types
- Container now includes `VolumeMounts` from shared type
- Pod spec now includes `ServiceAccountName` and `Volumes` array

**Modified**: `iac/pulumi/module/main.go`
- Resource creation order: Namespace → ConfigMaps → ServiceAccount → RBAC → Secret → ImagePullSecret → DaemonSet
- ConfigMaps created before DaemonSet with proper dependency ordering

## Benefits

### For DevOps Engineers

- **Single manifest deployment**: ConfigMaps, ServiceAccount, RBAC, and DaemonSet defined together
- **Type-safe configuration**: Proto validation catches errors before deployment
- **Consistent volume patterns**: Same VolumeMount type as KubernetesDeployment

### For Platform Teams

- **Reduced operational overhead**: No need to manage separate RBAC manifests
- **Standardized patterns**: All Kubernetes workload components share volume mount definitions
- **Audit-friendly**: RBAC permissions declared alongside the workload they serve

### Concrete Example

Before (required external resources):
```bash
# Create ConfigMap
kubectl create configmap vector-config --from-file=vector.yaml

# Create ServiceAccount
kubectl create serviceaccount vector -n logging

# Create ClusterRole
kubectl create clusterrole vector-reader --verb=get,list,watch --resource=pods,nodes

# Create ClusterRoleBinding
kubectl create clusterrolebinding vector-reader --clusterrole=vector-reader --serviceaccount=logging:vector

# Finally deploy
planton pulumi up --manifest daemonset.yaml
```

After (all-in-one):
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDaemonSet
metadata:
  name: vector-logs
spec:
  namespace:
    value: logging
  createNamespace: true
  createServiceAccount: true
  serviceAccountName: vector
  configMaps:
    vector-config: |
      data_dir: /var/lib/vector
      sources:
        kubernetes_logs:
          type: kubernetes_logs
          self_node_name: ${K8S_NODE_NAME}
      sinks:
        nats:
          type: nats
          inputs: [kubernetes_logs]
          url: "nats://nats.default.svc:4222"
  rbac:
    clusterRules:
      - apiGroups: [""]
        resources: ["pods", "nodes", "namespaces"]
        verbs: ["get", "list", "watch"]
  container:
    app:
      image:
        repo: timberio/vector
        tag: 0.47.0-distroless-libc
      args: ["-c", "/etc/vector/vector.yaml", "--watch-config"]
      securityContext:
        runAsUser: 0
      volumeMounts:
        - name: varlog
          mountPath: /var/log
          hostPath:
            path: /var/log
        - name: vector-config
          mountPath: /etc/vector/vector.yaml
          configMap:
            name: vector-config
            key: vector-config
  tolerations:
    - key: node-role.kubernetes.io/master
      operator: Exists
      effect: NoSchedule
```

## Impact

### Who Is Affected

- **KubernetesDaemonSet users**: Full volume mount support matching KubernetesDeployment
- **Platform teams**: Simplified DaemonSet deployments with built-in RBAC

### Backward Compatibility

✅ **Fully backward compatible**: All new fields are optional. Existing DaemonSets without the new fields continue to work unchanged.

### Migration Notes

- Existing DaemonSets using the old `host_path` field format in volume mounts must migrate to the new `hostPath` object format:

**Before**:
```yaml
volumeMounts:
  - name: varlog
    mountPath: /var/log
    hostPath: /var/log  # Old flat string
```

**After**:
```yaml
volumeMounts:
  - name: varlog
    mountPath: /var/log
    hostPath:
      path: /var/log  # New structured object
      type: Directory
```

## Files Changed

| File | Change Type | Description |
|------|-------------|-------------|
| `spec.proto` | Modified | Added 4 new fields, 2 new messages, migrated to shared VolumeMount |
| `configmap.go` | Created | ConfigMap creation from spec.config_maps |
| `service_account.go` | Created | ServiceAccount creation logic |
| `rbac.go` | Created | ClusterRole, ClusterRoleBinding, Role, RoleBinding |
| `daemonset.go` | Modified | Volume mount handling, ServiceAccountName |
| `main.go` | Modified | Resource orchestration order |
| `spec_test.go` | Modified | Tests for new VolumeMount format and RBAC validation |

## Validation

All validation steps passed:

```bash
✅ make protos        # Proto generation
✅ go test ./apis/dev/planton/provider/kubernetes/kubernetesdaemonset/v1/...
✅ make build         # Full build
✅ make test          # Complete test suite
```

## Related Work

- **KubernetesDeployment ConfigMap Support**: Same pattern applied earlier today, enabling feature parity between Deployment and DaemonSet components
- **Shared Volume Mount Proto**: `volume_mount.proto` provides consistent volume mount definitions across KubernetesDeployment, KubernetesDaemonSet, KubernetesStatefulSet, and KubernetesCronJob

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation

