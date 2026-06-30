# KubernetesGhaRunnerScaleSet Deployment Component

**Date**: January 7, 2026
**Type**: Feature
**Components**: API Definitions, Kubernetes Provider, Pulumi Module, Terraform Module, Documentation

## Summary

Added a new deployment component `KubernetesGhaRunnerScaleSet` that enables declarative deployment of GitHub Actions Runner Scale Sets on Kubernetes. This component creates self-hosted runners that automatically scale based on workflow demand, with support for persistent volumes for caching build dependencies.

## Problem Statement / Motivation

With the GitHub Actions Runner Scale Set Controller deployed (see previous changelog), teams need a way to create and manage runner pools that connect to GitHub repositories, organizations, or enterprises.

### Pain Points

- **Manual Helm Configuration**: Installing runner scale sets required manually crafting Helm values with complex nested structures
- **No PVC Management**: Persistent caching for npm, gradle, maven, or Docker layers required separate PVC provisioning
- **Authentication Complexity**: GitHub PAT tokens and GitHub App credentials had to be managed outside the deployment
- **Inconsistent Interface**: Runner scale sets didn't follow the Planton declarative pattern

## Solution / What's New

Created a complete deployment component following the Planton standard structure:

### Component Structure

```
kubernetesgharunnerscaleset/v1/
├── api.proto              # KRM envelope
├── spec.proto             # Configuration schema with 4 container modes
├── stack_input.proto      # IaC inputs
├── stack_outputs.proto    # Deployment outputs
├── spec_test.go           # 43 validation tests
├── README.md              # User docs
├── examples.md            # 7 deployment scenarios
├── docs/README.md         # Technical docs
└── iac/
    ├── pulumi/module/     # Go-based Helm + PVC deployment
    └── tf/                # Terraform feature parity
```

### Key Features

**Container Modes**:

```protobuf
enum ContainerModeType {
  container_mode_type_unspecified = 0;
  DIND = 1;                    // Docker-in-Docker
  KUBERNETES = 2;              // Each step as K8s pod
  KUBERNETES_NO_VOLUME = 3;    // K8s mode without ephemeral volumes
  DEFAULT = 4;                 // Direct execution
}
```

**Persistent Volumes** (the key new feature):

```protobuf
message KubernetesGhaRunnerScaleSetPersistentVolume {
  string name = 1;           // Volume name
  string storage_class = 2;  // Optional storage class
  string size = 3;           // Size (e.g., "20Gi")
  repeated string access_modes = 4;
  string mount_path = 5;     // Container mount path
  bool read_only = 6;
}
```

**Authentication Options**:

- PAT Token: Simple personal access token
- GitHub App: Recommended for organizations
- Existing Secret: Pre-provisioned credentials

## Implementation Details

### Proto Schema Highlights

Comprehensive validation with CEL expressions:

```protobuf
string config_url = 1 [
  (buf.validate.field).required = true,
  (buf.validate.field).cel = {
    id: "spec.github.config_url"
    message: "Config URL must be a valid GitHub URL"
    expression: "this.startsWith('https://github.com/') || this.startsWith('https://github.')"
  }
];
```

### Pulumi Module

The module handles PVC creation before Helm deployment:

```go
// createPersistentVolumes creates PVCs for persistent storage.
func createPersistentVolumes(ctx *pulumi.Context, locals *Locals, ...) {
    for _, pv := range locals.PersistentVolumes {
        pvcName := fmt.Sprintf("%s-%s", locals.ReleaseName, pv.Name)
        spec := &corev1.PersistentVolumeClaimSpecArgs{
            AccessModes: pulumi.ToStringArray(accessModes),
            Resources: &corev1.VolumeResourceRequirementsArgs{...},
        }
        if pv.StorageClass != "" {
            spec.StorageClassName = pulumi.String(pv.StorageClass)
        }
        // Create PVC...
    }
}
```

Volume mounts are then added to the runner pod template spec:

```go
// Add persistent volume mounts
for _, pv := range locals.PersistentVolumes {
    mount := pulumi.Map{
        "name":      pulumi.String(pv.Name),
        "mountPath": pulumi.String(pv.MountPath),
    }
    mountArray = append(mountArray, mount)
}
```

### Terraform Module

Feature parity with Pulumi:

```hcl
resource "kubernetes_persistent_volume_claim" "this" {
  for_each = { for pv in var.spec.persistent_volumes : pv.name => pv }

  metadata {
    name      = "${local.release_name}-${each.value.name}"
    namespace = local.namespace
  }

  spec {
    access_modes       = each.value.access_modes
    storage_class_name = each.value.storage_class != "" ? each.value.storage_class : null
    resources {
      requests = { storage = each.value.size }
    }
  }
}
```

### Registry Entry

```protobuf
KubernetesGhaRunnerScaleSet = 844 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "k8sgharss"
}];
```

## Benefits

### For Users

- **Declarative Runners**: Define runner pools in YAML, deploy with CLI
- **Persistent Caching**: npm, gradle, maven, Docker layer caching out-of-the-box
- **Authentication Flexibility**: PAT, GitHub App, or bring-your-own-secret
- **Scaling Control**: Scale-to-zero for cost savings, warm pools for speed

### For Platform Teams

- **Multi-IaC**: Choose Pulumi or Terraform
- **Validation**: 43 tests catch configuration errors early
- **Documentation**: 7 examples cover common scenarios

## Impact

| Area              | Impact                                              |
| ----------------- | --------------------------------------------------- |
| Users             | Can deploy self-hosted GitHub runners declaratively |
| CI/CD             | Faster builds with persistent dependency caching    |
| Costs             | Scale-to-zero reduces idle resource waste           |
| Component Catalog | +1 Kubernetes addon (844th cloud resource kind)     |

## Usage Example

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGhaRunnerScaleSet
metadata:
  name: build-runners
spec:
  namespace:
    value: gha-runners
  createNamespace: true

  github:
    configUrl: https://github.com/myorg/myrepo
    patToken:
      token: ghp_xxxxxxxxxxxx

  containerMode:
    type: DIND

  scaling:
    minRunners: 0
    maxRunners: 10

  persistentVolumes:
    - name: npm-cache
      size: '20Gi'
      mountPath: /home/runner/.npm
    - name: gradle-cache
      size: '50Gi'
      mountPath: /home/runner/.gradle
```

Deploy:

```bash
planton pulumi up --manifest runners.yaml --stack org/project/env
```

Use in workflow:

```yaml
jobs:
  build:
    runs-on: [self-hosted, build-runners]
```

## Testing

- ✅ 43 validation tests covering all spec fields
- ✅ `go vet` passes
- ✅ `buf build` validates proto files
- ✅ `make build` compiles successfully
- ✅ Terraform validates

## Related Work

- **KubernetesGhaRunnerScaleSetController** (previous changelog): Deploys the controller that manages these scale sets
- Similar pattern to other Kubernetes addon components (Tekton, ArgoCD, etc.)

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation
