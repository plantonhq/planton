# GcpDataprocVirtualCluster - Pulumi Implementation

## Overview

The Pulumi module provisions a Dataproc on GKE virtual cluster using the `dataproc.NewCluster` resource from the `pulumi-gcp` provider (v9). Instead of a `ClusterConfig` with GCE-based master/worker nodes, it uses a `VirtualClusterConfig` that targets an existing GKE cluster and schedules Spark workloads as Kubernetes pods.

## Prerequisites

Before deploying a virtual cluster, ensure:

1. **GKE cluster exists**: The target GKE cluster must be running in the same project and region
2. **Node pools exist**: At least one GKE node pool must be available for the DEFAULT role
3. **IAM bindings**: The Dataproc service agent must have `roles/container.developer` on the GKE cluster
4. **Workload Identity**: Kubernetes service accounts must be bound to GCP service accounts for data access
5. **APIs enabled**: `dataproc.googleapis.com`, `container.googleapis.com`, `storage.googleapis.com`

## File Structure

| File | Purpose |
|---|---|
| `main.go` | Pulumi program entry point; deserializes stack input and calls module |
| `Makefile` | Build, preview, apply, and destroy targets |
| `Pulumi.yaml` | Pulumi project configuration |
| `module/main.go` | Module entry point; initializes locals, GCP provider, and calls resource creation |
| `module/locals.go` | Builds the `Locals` struct with GCP labels and extracted configuration |
| `module/dataproc_virtual_cluster.go` | Creates the `dataproc.Cluster` resource with `VirtualClusterConfig` |
| `module/outputs.go` | Defines output key constants for stack exports |

## Building

```bash
cd iac/pulumi
make build
```

## Testing

```bash
# Preview changes
make preview

# Apply changes
make up

# Destroy resources
make destroy
```

## Debug

Set the stack input using the hack manifest:

```bash
pulumi config set --path 'target' "$(cat iac/hack/manifest.yaml | yq -o json)"
```

Then run a preview to verify the resource graph:

```bash
make preview
```

To inspect the resolved Pulumi arguments, add debug logging in `module/dataproc_virtual_cluster.go` and rebuild.
