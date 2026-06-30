# KubernetesCronJob: ConfigMap and Volume Mount Support

**Date**: December 21, 2025
**Type**: Enhancement
**Components**: API Definitions, Kubernetes Provider, Pulumi Module, Terraform Module

## Summary

Enhanced the KubernetesCronJob component to support creating ConfigMaps from inline content and mounting them (along with Secrets, HostPaths, EmptyDirs, and PVCs) as volumes into containers. This enables deploying CronJobs that require configuration files—a common pattern for batch jobs like database backups, data processing scripts, and scheduled tasks.

## Problem Statement / Motivation

Deploying CronJobs that need configuration files (backup scripts, config files, etc.) required either:
- Pre-creating ConfigMaps outside the deployment workflow
- Baking configuration into container images (violating 12-factor app principles)
- Using KubernetesManifest with raw YAML (losing type safety)

### Pain Points

- **No declarative ConfigMap creation**: Users couldn't define configuration file content alongside their CronJob
- **No volume mount support**: Even if ConfigMaps existed, there was no way to mount them
- **Inconsistency with KubernetesDeployment**: The KubernetesDeployment component already had this support

## Solution / What's New

Added comprehensive volume mount support to KubernetesCronJob, mirroring the implementation already present in KubernetesDeployment. The solution leverages the existing shared `volume_mount.proto` definitions for consistency.

### New Proto Fields

```protobuf
// In KubernetesCronJobSpec
map<string, string> config_maps = 17;  // Key=name, Value=content
repeated VolumeMount volume_mounts = 18;
```

### Supported Volume Types

| Volume Type | Use Case |
|-------------|----------|
| ConfigMap | Scripts, configuration files |
| Secret | TLS certificates, credentials |
| HostPath | Node-level resources (logs, sockets) |
| EmptyDir | Temporary storage, scratch space |
| PVC | Persistent data storage |

## Implementation Details

### Proto Schema Changes

**File**: `apis/dev/planton/provider/kubernetes/kubernetescronjob/v1/spec.proto`

Added import for shared volume mount definitions and two new fields:
- `config_maps` for declarative ConfigMap creation
- `volume_mounts` for mounting volumes into the container

### Pulumi Module (Go)

**New file**: `iac/pulumi/module/configmap.go`
- Creates ConfigMap resources from the `spec.config_maps` map

**New file**: `iac/pulumi/module/volumes.go`
- `buildVolumeMountsAndVolumes()` function processing all 5 volume types

**Modified**: `iac/pulumi/module/cron_job.go`
- Container now includes `VolumeMounts`
- Pod spec now includes `Volumes` array

**Modified**: `iac/pulumi/module/main.go`
- ConfigMaps created before CronJob with proper dependency ordering

### Terraform Module

**New file**: `iac/tf/configmap.tf`
- Creates ConfigMap resources using `for_each`

**Modified**: `iac/tf/variables.tf`
- Added `config_maps` and `volume_mounts` variable definitions

**Modified**: `iac/tf/cron_job.tf`
- Dynamic `volume_mount` blocks for container
- Dynamic `volume` blocks for each volume type
- Proper dependency on `kubernetes_config_map.this`

## Benefits

### For DevOps Engineers

- **Single manifest deployment**: ConfigMaps and volumes defined alongside the CronJob
- **Type-safe configuration**: Proto validation catches errors before deployment
- **Multi-IaC parity**: Same capabilities in Pulumi and Terraform

### For Platform Teams

- **Consistent patterns**: Volume mounts use shared proto definitions across KubernetesDeployment, KubernetesStatefulSet, KubernetesDaemonSet, and now KubernetesCronJob
- **Reduced maintenance**: Single source of truth for volume mount types

### Concrete Example

Before (required external ConfigMap):
```bash
kubectl create configmap backup-script --from-file=backup.sh
# Then deploy CronJob separately
planton pulumi up --manifest cronjob.yaml
```

After (all-in-one):
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesCronJob
metadata:
  name: db-backup
spec:
  namespace:
    value: database
  schedule: "0 2 * * *"
  configMaps:
    backup-script: |
      #!/bin/bash
      pg_dump -h $DB_HOST -U $DB_USER $DB_NAME > /backup/dump.sql
  image:
    repo: postgres
    tag: "15"
  command: ["/bin/bash", "/scripts/backup.sh"]
  volumeMounts:
    - name: backup-script
      mountPath: /scripts/backup.sh
      configMap:
        name: backup-script
        key: backup-script
        path: backup.sh
        defaultMode: 493  # 0755 - executable
    - name: backup-data
      mountPath: /backup
      emptyDir:
        sizeLimit: 1Gi
```

## Impact

### Who Is Affected

- **KubernetesCronJob users**: New capabilities for configuration management
- **Consistency across components**: KubernetesCronJob now has feature parity with KubernetesDeployment for volume mounts

### Backward Compatibility

✅ **Fully backward compatible**: All new fields are optional. Existing CronJobs without `config_maps` or `volume_mounts` continue to work unchanged.

## Files Changed

| File | Change Type |
|------|-------------|
| `spec.proto` | Modified - Added 2 fields |
| `configmap.go` | Created - Pulumi ConfigMap creation |
| `volumes.go` | Created - Volume building logic |
| `cron_job.go` | Modified - Volume mount handling |
| `main.go` | Modified - ConfigMap orchestration |
| `configmap.tf` | Created - Terraform ConfigMap resource |
| `variables.tf` | Modified - New variable definitions |
| `cron_job.tf` | Modified - Volume blocks |
| `examples.md` | Modified - 3 new examples |
| `iac/tf/examples.md` | Modified - 2 new Terraform examples |

## Validation

All validation steps passed:

```bash
✅ make protos        # Proto generation
✅ go build ./...     # Build verification
✅ go test ./...      # Test suite
```

## Related Work

- **KubernetesDeployment Enhancement**: Same pattern implemented on 2025-12-21 (see `2025-12-21-042723-kubernetes-deployment-configmap-volume-mount-support.md`)
- **Shared Volume Mount Proto**: `volume_mount.proto` contains all required message types

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation

