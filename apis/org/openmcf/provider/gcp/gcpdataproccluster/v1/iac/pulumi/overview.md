# GcpDataprocCluster - Pulumi Module Architecture

## Overview

The Pulumi module creates a standard (GCE-based) Google Cloud Dataproc cluster using the `dataproc.NewCluster` resource from the `pulumi-gcp` provider.

## Resource Graph

```
dataproc.Cluster ("dataproc-cluster")
├── ClusterConfig
│   ├── GceClusterConfig (networking, service account, zone)
│   ├── MasterConfig (machine type, disk, accelerators)
│   ├── WorkerConfig (machine type, disk, accelerators)
│   ├── PreemptibleWorkerConfig (spot/preemptible)
│   ├── SoftwareConfig (image version, components, properties)
│   ├── InitializationActions (startup scripts)
│   ├── AutoscalingConfig (external policy reference)
│   ├── EncryptionConfig (CMEK)
│   ├── EndpointConfig (Component Gateway)
│   └── LifecycleConfig (auto-delete, idle shutdown)
└── Labels (framework labels)
```

## Module Files

| File | Purpose |
|---|---|
| `module/main.go` | Entry point; initializes locals, GCP provider, and calls resource creation |
| `module/locals.go` | Builds the `Locals` struct with labels and extracted configuration |
| `module/dataproc_cluster.go` | Creates the `dataproc.Cluster` resource with all nested configurations |
| `module/outputs.go` | Defines output key constants for stack exports |

## Key Implementation Details

### Single Resource

Unlike components like GcpAlloydbCluster (cluster + primary instance), GcpDataprocCluster creates a single `dataproc.Cluster` resource. Master, worker, and secondary worker configurations are all nested within the cluster's `ClusterConfig`.

### Conditional Nested Blocks

Every nested configuration block (GCE config, master config, worker config, etc.) is conditionally set only when the corresponding spec field is non-nil. This allows users to rely on GCP defaults for any unspecified configuration.

### StringValueOrRef Resolution

All cross-resource references (project, network, subnetwork, service account, GCS buckets, KMS key) use `.GetValue()` to resolve `StringValueOrRef` fields.

### Framework Labels

GCP labels are constructed from metadata (organization, environment, resource ID, resource kind) and applied to the cluster. Dataproc supports GCP labels on the cluster resource.

## Outputs

| Output Key | Source | Description |
|---|---|---|
| `cluster_id` | `createdCluster.Name` | Fully qualified cluster resource name |
| `cluster_name` | `spec.ClusterName` | Short cluster name |
| `cluster_uuid` | Computed | Server-generated UUID (if available) |
| `staging_bucket` | `ClusterConfig.Bucket()` | Staging bucket (user-supplied or auto-created) |
