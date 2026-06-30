# KubernetesDeployment: ConfigMap and Volume Mount Support

**Date**: December 21, 2025
**Type**: Enhancement
**Components**: API Definitions, Kubernetes Provider, Pulumi CLI Integration, Terraform Module

## Summary

Enhanced the KubernetesDeployment component to support creating ConfigMaps from inline content and mounting them (along with Secrets, HostPaths, EmptyDirs, and PVCs) as volumes into containers. This enables deploying applications that require configuration files without baking them into container images—a common production pattern for Kubernetes workloads.

## Problem Statement / Motivation

Deploying applications like the Tekton CloudEvents Router requires configuration files to be mounted into containers at runtime. Previously, KubernetesDeployment only supported environment variables and secrets, forcing users to either:
- Pre-create ConfigMaps outside the deployment workflow
- Bake configuration into container images (violating 12-factor app principles)
- Use KubernetesManifest with raw YAML (losing type safety)

### Pain Points

- **No declarative ConfigMap creation**: Users couldn't define configuration file content alongside their deployment
- **No volume mount support**: Even if ConfigMaps existed, there was no way to mount them
- **Command/args override missing**: Couldn't customize container entrypoints without modifying images
- **Inconsistency across IaC modules**: Pulumi and Terraform modules had different capabilities

## Solution / What's New

Added comprehensive volume mount support to KubernetesDeployment, including inline ConfigMap creation. The solution leverages the existing shared `volume_mount.proto` definitions for consistency across workload components.

### New Proto Fields

```protobuf
// In KubernetesDeploymentSpec
map<string, string> config_maps = 8;  // Key=name, Value=content

// In KubernetesDeploymentContainerApp
repeated VolumeMount volume_mounts = 8;
repeated string command = 9;
repeated string args = 10;
```

### Supported Volume Types

| Volume Type | Use Case |
|-------------|----------|
| ConfigMap | Application configuration files |
| Secret | TLS certificates, credentials |
| HostPath | Node-level resources (logs, sockets) |
| EmptyDir | Temporary storage, inter-container sharing |
| PVC | Persistent data storage |

## Implementation Details

### Proto Schema Changes

**File**: `apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1/spec.proto`

Added import for shared volume mount definitions and three new fields:
- `config_maps` on `KubernetesDeploymentSpec` for declarative ConfigMap creation
- `volume_mounts`, `command`, `args` on `KubernetesDeploymentContainerApp`

### Pulumi Module (Go)

**New file**: `iac/pulumi/module/configmap.go`
```go
func configMaps(ctx *pulumi.Context, locals *Locals, 
    kubernetesProvider pulumi.ProviderResource) (map[string]*kubernetescorev1.ConfigMap, error)
```

**Modified**: `iac/pulumi/module/deployment.go`
- Added `buildVolumeMountsAndVolumes()` function processing all 5 volume types
- Container now includes `VolumeMounts`, `Command`, `Args`
- Pod spec now includes `Volumes` array

**Modified**: `iac/pulumi/module/main.go`
- ConfigMaps created before deployment with proper dependency ordering

### Terraform Module

**New file**: `iac/tf/configmap.tf`
```hcl
resource "kubernetes_config_map" "this" {
  for_each = try(var.spec.config_maps, {})
  # ...
}
```

**Modified**: `iac/tf/variables.tf`
- Added `config_maps`, `volume_mounts`, `command`, `args` variable definitions

**Modified**: `iac/tf/deployment.tf`
- Dynamic `volume_mount` blocks for container
- Dynamic `volume` blocks for each volume type (ConfigMap, Secret, HostPath, EmptyDir, PVC)
- `command` and `args` support
- Proper dependency on `kubernetes_config_map.this`

## Benefits

### For Application Developers

- **Single manifest deployment**: ConfigMaps and volumes defined alongside the workload
- **Type-safe configuration**: Proto validation catches errors before deployment
- **Multi-IaC parity**: Same capabilities in Pulumi and Terraform

### For Platform Teams

- **Consistent patterns**: Volume mounts use shared proto definitions across KubernetesDeployment, KubernetesDaemonSet, KubernetesStatefulSet, KubernetesCronJob
- **Reduced maintenance**: Single source of truth for volume mount types

### Concrete Example

Before (required external ConfigMap):
```bash
kubectl create configmap router-config --from-file=config.yaml
# Then deploy separately
planton pulumi up --manifest deployment.yaml
```

After (all-in-one):
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: tekton-cloudevents-router
spec:
  namespace:
    value: tekton-pipelines
  version: main
  configMaps:
    router-config: |
      routes:
        - namespace_prefix: planton-dev-
          target: http://servicehub.planton-dev.svc/webhook
  container:
    app:
      image:
        repo: ghcr.io/plantoncloud/tekton-cloud-event-router
        tag: v0.1.0
      volumeMounts:
        - name: router-config
          mountPath: /etc/router/config.yaml
          configMap:
            name: router-config
            key: router-config
      ports:
        - name: http
          containerPort: 8080
          networkProtocol: TCP
          appProtocol: http
          servicePort: 80
```

## Impact

### Who Is Affected

- **KubernetesDeployment users**: New capabilities for configuration management
- **Other workload components**: Pattern established for adding volume support to KubernetesDaemonSet, KubernetesCronJob, etc.

### Backward Compatibility

✅ **Fully backward compatible**: All new fields are optional. Existing deployments without `config_maps` or `volume_mounts` continue to work unchanged.

## Files Changed

| File | Change Type |
|------|-------------|
| `spec.proto` | Modified - Added 4 fields |
| `configmap.go` | Created - Pulumi ConfigMap creation |
| `deployment.go` | Modified - Volume mount handling |
| `main.go` | Modified - ConfigMap orchestration |
| `configmap.tf` | Created - Terraform ConfigMap resource |
| `variables.tf` | Modified - New variable definitions |
| `deployment.tf` | Modified - Volume blocks |
| `examples.md` | Modified - 4 new examples |
| `iac/tf/examples.md` | Modified - 3 new Terraform examples |

## Validation

All validation steps passed:

```bash
✅ make protos        # Proto generation
✅ make build         # Full build 
✅ go test ./apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1/...
✅ make test          # Full test suite
```

## Related Work

- **Shared Volume Mount Proto**: `volume_mount.proto` was already present with all required message types, enabling this enhancement
- **Future Enhancement**: Apply same pattern to KubernetesDaemonSet (see `kubernetesdaemonset/v1/_cursor/requirement.md`)

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation

