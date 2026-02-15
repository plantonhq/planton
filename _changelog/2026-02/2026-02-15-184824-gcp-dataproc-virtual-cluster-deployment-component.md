# GcpDataprocVirtualCluster Deployment Component

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, GCP Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added the GcpDataprocVirtualCluster deployment component (R14b) for running Apache Spark workloads on existing GKE clusters via Dataproc on GKE. This component complements GcpDataprocCluster (standard Dataproc) by targeting organizations that want to consolidate Spark processing onto shared Kubernetes infrastructure. Also enhanced GcpGkeCluster and GcpGkeNodePool with fully qualified resource ID outputs to enable cross-resource composition.

## Problem Statement / Motivation

Dataproc on GKE allows scheduling Spark, PySpark, and SparkR jobs as Kubernetes pods on existing GKE node pools, instead of requiring dedicated Compute Engine VMs. The standard GcpDataprocCluster component uses `cluster_config` (Compute Engine-based), while Dataproc on GKE uses an entirely different `virtual_cluster_config` structure with Kubernetes-specific concepts like node pool roles and namespaces.

### Pain Points

- No way to deploy Spark workloads on shared GKE infrastructure through OpenMCF
- GcpGkeCluster did not export its fully qualified cluster ID, preventing downstream resource composition
- GcpGkeNodePool did not export its fully qualified node pool ID for the same reason

## Solution / What's New

### GcpDataprocVirtualCluster Component (47 files)

- **Proto API** with 8 message types, 5 StringValueOrRef fields, and CEL validations
- **Pulumi module** (4 Go files) creating `dataproc.Cluster` with `VirtualClusterConfig`
- **Terraform module** (6 files) with dynamic blocks for node pool targets, autoscaling, and auxiliary services
- **33 validation tests** (19 positive, 14 negative) all passing
- **Documentation**: user-facing README, 7 YAML examples, comprehensive research docs
- **3 presets**: basic-spark-on-gke, production-multi-pool, metastore-integrated

### GcpGkeCluster Output Enhancement

Added `cluster_id` output (fully qualified `projects/{project}/locations/{location}/clusters/{name}`) to stack_outputs.proto, Pulumi module, and Terraform module.

### GcpGkeNodePool Output Enhancement

Added `node_pool_id` output (fully qualified path) to stack_outputs.proto, Pulumi module, and Terraform module.

## Implementation Details

### Flattened Spec Design

The provider's virtual cluster config is deeply nested (6 levels). Since this component is always a virtual cluster and Kubernetes is the only virtual cluster type, the outer wrappers (`virtual_cluster_config`, `kubernetes_cluster_config`) were flattened while preserving meaningful inner groupings (`software_config`, `node_pool_targets`, `auxiliary_services_config`).

### StringValueOrRef Composition (5 fields)

| Field | Default Kind | Field Path |
|-------|-------------|------------|
| `project_id` | GcpProject | `status.outputs.project_id` |
| `gke_cluster_target` | GcpGkeCluster | `status.outputs.cluster_id` |
| `kubernetes_namespace` | KubernetesNamespace | `spec.name` |
| `staging_bucket` | GcpGcsBucket | `status.outputs.bucket_id` |
| `node_pool` (in targets) | GcpGkeNodePool | `status.outputs.node_pool_id` |

### Node Pool Roles

Each node pool target is assigned one or more roles: DEFAULT, CONTROLLER, SPARK_DRIVER, SPARK_EXECUTOR. At least one target must have the DEFAULT role.

## Benefits

- Consolidate Spark processing onto existing GKE infrastructure (cost sharing)
- 5 StringValueOrRef fields enable full cross-resource composition
- GcpGkeCluster and GcpGkeNodePool now export fully qualified IDs usable by any downstream component
- Consistent with GcpDataprocCluster (R14) patterns while optimized for the Kubernetes deployment model

## Impact

- **New resource kind**: GcpDataprocVirtualCluster (enum 652, id_prefix: gcpdvc)
- **Enhanced outputs**: GcpGkeCluster now exports `cluster_id`, GcpGkeNodePool now exports `node_pool_id`
- **Companion component**: GcpGkeWorkloadIdentityBinding handles required IAM bindings

## Related Work

- R14 GcpDataprocCluster: standard Dataproc component (same underlying Terraform/Pulumi resource, different config mode)
- GcpGkeCluster / GcpGkeNodePool: prerequisites for virtual cluster deployment
- GcpGkeWorkloadIdentityBinding: manages Workload Identity IAM bindings required by Dataproc on GKE

---

**Status**: Production Ready
