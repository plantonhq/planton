# GcpDataprocCluster - Research and Design Documentation

## GCP Dataproc Landscape

Google Cloud Dataproc is a fully managed service for running Apache Spark, Apache Hadoop, Apache Hive, Apache Pig, and other open-source data processing frameworks on Google Cloud infrastructure.

### Deployment Models

1. **Standard Clusters (GCE-based)** - This component. Dataproc provisions and manages Compute Engine VMs as cluster nodes. Most common deployment model.

2. **Dataproc-on-GKE (Virtual Clusters)** - Runs Spark workloads on an existing GKE cluster using Kubernetes pods instead of dedicated VMs. Separate component: GcpDataprocVirtualCluster.

3. **Dataproc Serverless** - Fully managed, auto-scaling Spark execution without cluster management. Not a Dataproc "cluster" -- it's a job submission API. Out of scope for infrastructure-as-code.

### Cluster Architecture

A standard Dataproc cluster consists of:

- **Master nodes**: Run HDFS NameNode, YARN ResourceManager, Spark History Server, and Hive Metastore. Standard mode uses 1 master; HA mode uses 3.
- **Primary workers**: Run HDFS DataNodes and YARN NodeManagers. These are persistent, on-demand VMs.
- **Secondary workers**: Optional preemptible or Spot VMs for burst capacity. They participate in YARN but not HDFS (no DataNode), so preemption doesn't cause data loss.

### Image Versions

Dataproc image versions determine the installed software versions:

| Image | Spark | Hadoop | Hive | Python |
|---|---|---|---|---|
| 2.2-debian12 | 3.5.x | 3.3.x | 3.1.x | 3.11 |
| 2.1-debian12 | 3.4.x | 3.3.x | 3.1.x | 3.10 |
| 2.0-debian11 | 3.3.x | 3.2.x | 3.1.x | 3.9 |

### Optional Components

Commonly installed components:
- **JUPYTER**: JupyterLab notebooks for interactive data analysis
- **DOCKER**: Docker daemon on all nodes for containerized workloads
- **PRESTO/TRINO**: Distributed SQL query engine for ad-hoc analytics
- **ZEPPELIN**: Alternative notebook interface
- **FLINK**: Stream processing engine
- **HIVE_WEBHCAT**: REST API for Hive DDL

## Design Decisions

### 80/20 Scoping

The Terraform `google_dataproc_cluster` resource has 150+ fields across 40+ nested types and 5 levels of nesting depth. We expose approximately 40 fields that cover 90%+ of real-world Dataproc deployments.

### Kept `cluster_config` Wrapper

Unlike some Planton components that flatten nested configs, we kept the `cluster_config` wrapper to:
1. Mirror the Terraform/Pulumi structure familiar to engineers
2. Maintain compatibility for a future `GcpDataprocVirtualCluster` component

### Lifecycle Config is Critical

Dataproc clusters are often ephemeral. The `lifecycle_config` block (with `idle_delete_ttl` and `auto_delete_time`) prevents runaway costs from forgotten clusters. This was missing from the original plan and added as a correction.

### Secondary Workers, Not Preemptible Workers

GCP now supports PREEMPTIBLE, SPOT, and NON_PREEMPTIBLE secondary workers. We use "secondary_worker_config" (matching GCP's current terminology) instead of the older "preemptible_worker_config" (still used in Terraform for backward compatibility).

### Autoscaling via External Policy

Dataproc autoscaling uses a separate `google_dataproc_autoscaling_policy` resource referenced by URI. This is different from inline autoscaling (as in AlloyDB or Bigtable). We expose this as a simple `autoscaling_policy_uri` string field rather than modeling the policy inline.

### Accelerators on Master and Worker

GPU/TPU accelerators are included on both master and worker configs for ML-on-Spark workloads. Secondary workers inherit machine configuration from primary workers and don't independently support accelerators.

## Excluded Features (with Rationale)

| Feature | Rationale |
|---|---|
| Virtual cluster config | Fundamentally different deployment model (Kubernetes pods vs GCE VMs). Separate component. |
| Security config / Kerberos | 15+ fields for enterprise Hadoop security. Very niche. |
| Auxiliary node groups | Driver node separation. Advanced optimization. |
| Metastore config | References an external Hive Metastore service. Separate resource. |
| Dataproc metric config | Observability configuration. Operational concern, not infrastructure. |
| Shielded instance config | Secure boot, vTPM, integrity monitoring. Security hardening, can add in v2. |
| Reservation affinity | Capacity reservation targeting. Niche. |
| Node group affinity | Sole-tenant node placement. Niche. |
| Confidential instance config | Confidential computing. Niche. |
| Instance flexibility policy | Advanced fleet management with mixed machine types. Newer feature. |
| Cluster tier | STANDARD/PREMIUM. Newer feature, not widely adopted. |

## Terraform Provider Notes

- Provider version: `~> 6.0` (Google provider v6.50.0 tested)
- Most `cluster_config` fields are `ForceNew` (require cluster recreation on change)
- Exceptions: `worker_config.num_instances`, `worker_config.min_num_instances`, `graceful_decommission_timeout`
- `cluster_config` and `virtual_cluster_config` are mutually exclusive in Terraform (ExactlyOneOf)

## Pulumi SDK Notes

- SDK version: `pulumi-gcp/sdk/v9`
- Disk config types are separate per node group (MasterConfigDiskConfig vs WorkerConfigDiskConfig) despite identical fields
- Accelerator types are similarly separate per node group
- `ClusterConfig.Bucket()` returns the computed staging bucket
