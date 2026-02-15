# GcpDataprocVirtualCluster - Pulumi Module Architecture

## Overview

The Pulumi module creates a Dataproc on GKE virtual cluster using the `dataproc.NewCluster` resource with a `VirtualClusterConfig` instead of the standard `ClusterConfig`. This single resource encapsulates the GKE cluster target, Kubernetes software configuration, node pool role assignments, and optional auxiliary services.

## Resource Graph

```
dataproc.Cluster ("dataproc-virtual-cluster")
├── VirtualClusterConfig
│   ├── KubernetesClusterConfig
│   │   ├── GkeClusterConfig
│   │   │   ├── GkeClusterTarget (fully qualified GKE cluster ID)
│   │   │   └── NodePoolTargets[] (per-pool role and config)
│   │   │       ├── NodePool (GKE node pool reference)
│   │   │       ├── Roles[] (DEFAULT, CONTROLLER, SPARK_DRIVER, SPARK_EXECUTOR)
│   │   │       └── NodePoolConfig (optional: autoscaling, machine type, spot)
│   │   ├── KubernetesSoftwareConfig
│   │   │   ├── ComponentVersion (SPARK → version string)
│   │   │   └── Properties (spark:key → value)
│   │   └── KubernetesNamespace (optional)
│   ├── AuxiliaryServicesConfig (optional)
│   │   ├── MetastoreConfig (Dataproc Metastore service)
│   │   └── SparkHistoryServerConfig (Dataproc cluster reference)
│   └── StagingBucket (optional)
└── Labels (framework labels)
```

## Module Files

| File | Purpose |
|---|---|
| `module/main.go` | Entry point; initializes locals, GCP provider, and calls resource creation |
| `module/locals.go` | Builds the `Locals` struct with labels and extracted configuration |
| `module/dataproc_virtual_cluster.go` | Creates the `dataproc.Cluster` resource with all nested configurations |
| `module/outputs.go` | Defines output key constants for stack exports |

## How Spec Fields Map to Pulumi Args

### Top-Level Cluster Args

| Spec Field | Pulumi Arg | Notes |
|---|---|---|
| `spec.clusterName` / `metadata.name` | `Name` | Falls back to metadata name if cluster_name is empty |
| `spec.region` | `Region` | Must match GKE cluster region |
| `spec.projectId.GetValue()` | `Project` | Resolved from StringValueOrRef |
| `locals.GcpLabels` | `Labels` | Framework labels (resource kind, name, org, env) |

### VirtualClusterConfig Structure

The core of the mapping is the deeply nested `VirtualClusterConfig`:

```go
VirtualClusterConfig → ClusterVirtualClusterConfigArgs
  ├── KubernetesClusterConfig → ClusterVirtualClusterConfigKubernetesClusterConfigArgs
  │   ├── GkeClusterConfig → ...GkeClusterConfigArgs
  │   │   ├── GkeClusterTarget ← spec.GkeClusterTarget.GetValue()
  │   │   └── NodePoolTargets ← spec.NodePoolTargets[] (loop)
  │   ├── KubernetesSoftwareConfig → ...KubernetesSoftwareConfigArgs
  │   │   ├── ComponentVersion ← spec.SoftwareConfig.ComponentVersion
  │   │   └── Properties ← spec.SoftwareConfig.Properties
  │   └── KubernetesNamespace ← spec.KubernetesNamespace.GetValue()
  ├── AuxiliaryServicesConfig → ...AuxiliaryServicesConfigArgs
  │   ├── MetastoreConfig ← spec.AuxiliaryServicesConfig.MetastoreService
  │   └── SparkHistoryServerConfig ← spec.AuxiliaryServicesConfig.SparkHistoryServerCluster
  └── StagingBucket ← spec.StagingBucket.GetValue()
```

### Node Pool Target Mapping

Each `spec.NodePoolTargets[]` entry maps to a `NodePoolTargetArgs`:

```go
NodePoolTargetArgs
  ├── NodePool ← npt.NodePool.GetValue()      // StringValueOrRef
  ├── Roles ← npt.Roles                        // []string
  └── NodePoolConfig (optional)
      ├── Locations ← npt.NodePoolConfig.Locations
      ├── Autoscaling
      │   ├── MinNodeCount ← npt.NodePoolConfig.Autoscaling.MinNodeCount
      │   └── MaxNodeCount ← npt.NodePoolConfig.Autoscaling.MaxNodeCount
      └── Config
          ├── MachineType ← npt.NodePoolConfig.MachineType
          ├── LocalSsdCount ← npt.NodePoolConfig.LocalSsdCount
          ├── MinCpuPlatform ← npt.NodePoolConfig.MinCpuPlatform
          ├── Preemptible ← npt.NodePoolConfig.Preemptible
          └── Spot ← npt.NodePoolConfig.Spot
```

### Conditional Nested Blocks

Every optional nested block is only set when the corresponding spec field is non-nil or non-empty. This ensures:

1. Users can rely on GCP defaults for unspecified configuration
2. Terraform/Pulumi don't send empty blocks that may trigger validation errors
3. The resource graph is minimal and readable in `pulumi preview`

## StringValueOrRef Resolution

All cross-resource references use `.GetValue()` to resolve `StringValueOrRef` fields at module execution time:

- `spec.ProjectId.GetValue()` → GCP project ID
- `spec.GkeClusterTarget.GetValue()` → Fully qualified GKE cluster ID
- `spec.KubernetesNamespace.GetValue()` → Kubernetes namespace name
- `spec.StagingBucket.GetValue()` → GCS bucket name
- `npt.NodePool.GetValue()` → GKE node pool name or resource ID

## Labels and Naming

### Cluster Name

The cluster name is determined by:
1. `spec.ClusterName` if explicitly set
2. `metadata.Name` as fallback

### GCP Labels

Framework labels are applied to the Dataproc cluster resource:

| Label Key | Source |
|---|---|
| `openmcf-resource` | Always `"true"` |
| `openmcf-resource-kind` | `"gcpdataprocvirtualcluster"` |
| `openmcf-resource-name` | Resolved cluster name |
| `openmcf-organization` | `metadata.Org` (if set) |
| `openmcf-environment` | `metadata.Env` (if set) |
| `openmcf-resource-id` | `metadata.Id` (if set) |

## Outputs

| Output Key | Source | Description |
|---|---|---|
| `cluster_id` | `createdCluster.ID()` | Fully qualified Dataproc cluster resource name |
| `cluster_name` | `createdCluster.Name` | Short cluster name |
| `cluster_uuid` | `""` (not exposed) | Server-generated UUID — not available from the provider |
