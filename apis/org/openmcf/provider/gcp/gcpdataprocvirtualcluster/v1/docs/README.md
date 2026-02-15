# GcpDataprocVirtualCluster - Research and Design Documentation

## What is Dataproc on GKE?

Dataproc on GKE (also called "Dataproc Virtual Clusters") is a deployment model where Google Cloud Dataproc schedules Apache Spark workloads as Kubernetes pods on an existing GKE cluster, rather than provisioning dedicated Compute Engine VMs. The GKE cluster acts as a shared compute substrate, and each virtual cluster is a lightweight Dataproc abstraction that maps to a Kubernetes namespace.

This is fundamentally different from standard Dataproc clusters, which own and manage their own fleet of Compute Engine VMs (master nodes, primary workers, secondary workers).

### Key Differentiators

1. **No dedicated VMs**: Spark drivers and executors run as Kubernetes pods, not as processes on Dataproc-managed VMs
2. **Shared infrastructure**: Multiple virtual clusters share the same GKE cluster
3. **Kubernetes-native**: Resource isolation via namespaces, RBAC, and resource quotas
4. **GKE autoscaling**: Node pools scale based on pod demand, not Dataproc autoscaling policies
5. **Spark only**: Virtual clusters support Spark, PySpark, and SparkR — not the broader Hadoop ecosystem

## Architecture: How Spark Runs on GKE

When a Dataproc virtual cluster is created, the following happens:

1. **Namespace creation**: Dataproc creates a Kubernetes namespace (or uses the specified one) on the target GKE cluster
2. **Agent deployment**: A Dataproc agent pod is deployed in the namespace to manage Spark sessions
3. **Job submission**: When a Spark job is submitted, the agent creates driver and executor pods according to the node pool role assignments
4. **Pod scheduling**: Kubernetes schedules pods onto nodes in the assigned node pools
5. **Completion**: After the job finishes, executor pods are cleaned up; the driver pod retains logs

### Data Flow

```
User → Dataproc API → Dataproc Agent Pod → Spark Driver Pod → Spark Executor Pods
                                                ↕                     ↕
                                          GCS (staging)         GCS / BigQuery (data)
```

Spark reads and writes data through GCS, BigQuery, or other GCP services — not through HDFS. There is no HDFS layer in the virtual cluster model.

### Resource ID Anatomy

The GKE cluster target uses a fully qualified resource ID:

```
projects/{project}/locations/{location}/clusters/{cluster_name}
```

Note: This uses `locations` (not `regions`), matching the GKE API convention where the location can be a region or a zone.

## Node Pool Roles

Each GKE node pool assigned to a virtual cluster receives one or more Dataproc roles that determine which workload types are scheduled on its nodes.

### Role Definitions

| Role | Purpose | Typical Configuration |
|---|---|---|
| `DEFAULT` | Catch-all for any workload not assigned to a specific role. Required — every virtual cluster must have exactly one DEFAULT pool. | On-demand VMs, moderate size |
| `CONTROLLER` | Dataproc agent and cluster controller pods. Lightweight but must be stable. | Small on-demand VMs (e2-standard-2 or e2-standard-4) |
| `SPARK_DRIVER` | Spark driver pods. One per Spark session/job. Must be stable — if the driver dies, the entire job fails. | On-demand VMs, moderate memory |
| `SPARK_EXECUTOR` | Spark executor pods. Many per job; scale with data size. Tolerant of preemption (Spark retries failed tasks). | Spot/preemptible VMs, high CPU/memory |

### Role Assignment Rules

- At least one node pool target must have the `DEFAULT` role
- Each role can only be assigned to one node pool target
- A single node pool target can have multiple roles (e.g., DEFAULT + CONTROLLER)
- Preemptible/Spot VMs cannot be used for `CONTROLLER` or `DEFAULT` (when CONTROLLER is not separately assigned)

### Recommended Topology

**Development / small-scale**:
- One pool with `DEFAULT` (handles everything)

**Production**:
- Pool 1: `DEFAULT` + `CONTROLLER` — small, stable on-demand nodes
- Pool 2: `SPARK_DRIVER` — medium on-demand nodes
- Pool 3: `SPARK_EXECUTOR` — large Spot nodes with autoscaling

## IAM Requirements

Dataproc on GKE uses Workload Identity to authenticate Spark pods as GCP service accounts. This requires several IAM bindings:

### Dataproc Service Agent

The Dataproc service agent for the project (`service-<PROJECT_NUMBER>@dataproc-accounts.iam.gserviceaccount.com`) must have:

- `roles/container.developer` on the GKE cluster — to create pods, services, and other Kubernetes resources

This is typically configured automatically when the Dataproc API is enabled, but may need manual setup in projects with restrictive IAM policies.

### Workload Identity Bindings

Spark pods authenticate using Kubernetes service accounts that are bound to GCP service accounts via Workload Identity. The required bindings:

1. **Agent service account**: The Dataproc agent KSA must be bound to the Dataproc service agent GSA
2. **Spark service account**: Spark driver/executor KSAs must be bound to a GSA with permissions for the data sources (GCS, BigQuery, etc.)

In OpenMCF, use the `GcpGkeWorkloadIdentityBinding` component to manage these bindings declaratively.

### Minimum GCP APIs

The following APIs must be enabled on the project:

- `dataproc.googleapis.com`
- `container.googleapis.com`
- `storage.googleapis.com`

## Cost Considerations

### Sharing GKE Infrastructure

The primary cost advantage of Dataproc on GKE is infrastructure sharing:

- **No idle VM costs**: Unlike standard Dataproc clusters that maintain always-on master and worker VMs, virtual clusters only consume resources when Spark jobs are running
- **Multi-tenant**: Multiple virtual clusters (for different teams, environments, or workloads) share the same GKE node pools
- **Bin packing**: Kubernetes schedules pods efficiently across nodes, reducing waste

### Spot/Preemptible Executors

Spark executors are ideal candidates for Spot VMs:

- Spark automatically retries tasks when an executor is preempted
- Spot VMs cost 60-91% less than on-demand
- Autoscaling from 0 means zero cost when no jobs are running

### Cost Comparison: Virtual vs Standard

| Dimension | GcpDataprocCluster (Standard) | GcpDataprocVirtualCluster |
|---|---|---|
| Idle cost | Master + worker VMs always running | Zero (pods only during jobs) |
| Spark infrastructure | Managed by Dataproc | Managed by GKE |
| Multi-tenant | One cluster per team/environment | Multiple virtual clusters on one GKE |
| Autoscaling | Dataproc autoscaling (policy-based) | GKE node autoscaling (pod-based) |
| Spot/preemptible | Secondary workers only | Executor node pools |

### Dataproc Pricing

Dataproc charges a per-vCPU-hour premium on top of the underlying compute costs. For virtual clusters, this premium applies to the vCPUs used by Dataproc-managed pods (driver, executor, agent) for the duration they run.

## Spark Configuration

### Component Versions

The `componentVersion` map is mandatory and must include at least the `SPARK` key. The version string follows the `{major}.{minor}-dataproc-{patch}` format:

| Version String | Spark Version | Notes |
|---|---|---|
| `3.5-dataproc-17` | 3.5.x | Latest stable (recommended) |
| `3.4-dataproc-15` | 3.4.x | Previous stable |
| `3.3-dataproc-12` | 3.3.x | Maintenance |

### Properties

The `properties` map allows fine-grained Spark configuration using the `prefix:property` format:

```yaml
properties:
  # Memory configuration
  "spark:spark.executor.memory": "8g"
  "spark:spark.driver.memory": "4g"
  "spark:spark.executor.memoryOverhead": "2g"

  # Dynamic allocation
  "spark:spark.dynamicAllocation.enabled": "true"
  "spark:spark.dynamicAllocation.minExecutors": "0"
  "spark:spark.dynamicAllocation.maxExecutors": "100"

  # Hive catalog
  "spark:spark.sql.catalogImplementation": "hive"

  # Custom container image
  "spark:spark.kubernetes.container.image": "gcr.io/my-project/custom-spark:latest"
```

Common prefixes:

| Prefix | Daemon / Component |
|---|---|
| `spark` | Spark configuration |
| `mapred` | MapReduce configuration |
| `dataproc` | Dataproc-specific settings |

## Auxiliary Services

### Hive Metastore (Dataproc Metastore)

A Dataproc Metastore service provides a managed Hive Metastore for schema management. When configured, Spark jobs can access Hive tables without provisioning a standalone metastore:

```yaml
auxiliaryServicesConfig:
  metastoreService: "projects/{project}/locations/{location}/services/{service}"
```

The metastore must be in the same project and region as the virtual cluster. It uses the fully qualified resource name.

### Spark History Server

A Spark History Server provides a web UI for viewing completed Spark job logs. It runs as a standard Dataproc cluster (not a virtual cluster) and is referenced by its resource name:

```yaml
auxiliaryServicesConfig:
  sparkHistoryServerCluster: "projects/{project}/regions/{region}/clusters/{cluster}"
```

The history server reads Spark event logs from GCS. Configure Spark to write event logs by setting:

```yaml
properties:
  "spark:spark.eventLog.enabled": "true"
  "spark:spark.eventLog.dir": "gs://my-bucket/spark-events"
```

## 80/20 Scoping Rationale

### What Was Included

The OpenMCF GcpDataprocVirtualCluster spec exposes ~25 fields across 6 message types. These cover:

| Area | Fields | Rationale |
|---|---|---|
| Cluster identity | project_id, region, cluster_name | Fundamental resource identification |
| GKE integration | gke_cluster_target, kubernetes_namespace | Core of the virtual cluster model |
| Software | component_version, properties | Essential for any Spark workload |
| Node pools | node_pool, roles, node_pool_config, autoscaling | Primary mechanism for workload placement |
| Cost optimization | spot, preemptible on node pool config | Critical for production economics |
| Auxiliary services | metastore_service, spark_history_server_cluster | Common production integrations |
| Staging | staging_bucket | Job artifact management |

### What Was Excluded

| Feature | GCP API Field | Rationale |
|---|---|---|
| Custom container images | `image_uri` on node pool config | Advanced feature for custom Spark images. Properties-based override is sufficient for most cases. |
| Boot disk KMS key | `boot_disk_kms_key` on node pool config | CMEK for GKE node boot disks. Security hardening, can add in v2. |
| GPU accelerators | `accelerators` on node pool config | GPU attachment for ML workloads. Niche for Dataproc-on-GKE. |
| Dataproc Serverless | Separate API | Completely different deployment model — no cluster to manage. |
| Node affinity | `node_affinity` on node pool config | Kubernetes scheduling constraints. Advanced optimization. |

### Design Philosophy

The virtual cluster API is intentionally simpler than the standard Dataproc cluster API because:

1. **Infrastructure management is delegated to GKE**: Node pool sizing, networking, and machine types are managed by the GKE cluster and node pool resources
2. **Fewer knobs**: No HDFS, YARN, initialization actions, lifecycle config, or Component Gateway — those are standard Dataproc concepts
3. **Composition over configuration**: Complex setups compose virtual clusters with GKE node pools, Metastore services, and Workload Identity bindings as separate resources

## Comparison: GcpDataprocCluster vs GcpDataprocVirtualCluster

| Dimension | GcpDataprocCluster | GcpDataprocVirtualCluster |
|---|---|---|
| **Deployment target** | Dedicated Compute Engine VMs | Existing GKE cluster (Kubernetes pods) |
| **Infrastructure ownership** | Dataproc manages all VMs | GKE manages nodes; Dataproc schedules pods |
| **Supported frameworks** | Spark, Hadoop, Hive, Pig, Presto, Flink | Spark, PySpark, SparkR only |
| **Storage layer** | HDFS on local disks + GCS | GCS only (no HDFS) |
| **Optional components** | Jupyter, Zeppelin, Docker, Presto, Flink | None (Spark only) |
| **Networking** | VPC/subnetwork, internal IP only, tags | Inherited from GKE cluster |
| **Autoscaling** | Dataproc autoscaling policy (YARN-based) | GKE cluster autoscaler (pod-based) |
| **Multi-tenancy** | One cluster per workload/team | Multiple virtual clusters on shared GKE |
| **Idle cost** | Master + worker VMs always running | Zero (pods only during jobs) |
| **Startup time** | 60-120 seconds (VM provisioning) | 10-30 seconds (pod scheduling) |
| **Security model** | GCE service account, CMEK, internal IP | Workload Identity, K8s RBAC, namespaces |
| **Init actions** | Startup scripts on VMs | Not supported (use custom images) |
| **Lifecycle management** | idle_delete_ttl, auto_delete_time | Managed by GKE; virtual cluster is persistent |
| **Best for** | Full Hadoop ecosystem, GPU workloads, notebooks | Kubernetes-native Spark, multi-tenant, cost optimization |

## Terraform Provider Notes

- Provider version: `~> 6.0` (Google provider v6.50.0 tested)
- Uses the same `google_dataproc_cluster` resource as standard Dataproc, but with `virtual_cluster_config` instead of `cluster_config`
- `cluster_config` and `virtual_cluster_config` are mutually exclusive (`ExactlyOneOf`)
- The `virtual_cluster_config` block is `ForceNew` — changes require cluster recreation
- Node pool targets can reference existing GKE node pools or define new ones inline

## Pulumi SDK Notes

- SDK version: `pulumi-gcp/sdk/v9`
- Uses `dataproc.NewCluster` with `VirtualClusterConfig` instead of `ClusterConfig`
- Nested type names are deeply qualified (e.g., `ClusterVirtualClusterConfigKubernetesClusterConfigGkeClusterConfigNodePoolTargetArgs`)
- `cluster_uuid` is not directly exposed by the provider; downstream consumers should use `cluster_id`

## References

- [Dataproc on GKE overview](https://cloud.google.com/dataproc/docs/concepts/jobs/dataproc-gke)
- [Create a Dataproc on GKE cluster](https://cloud.google.com/dataproc/docs/guides/dpgke/dataproc-gke-quickstart)
- [Node pool targets and roles](https://cloud.google.com/dataproc/docs/guides/dpgke/dataproc-gke-node-pools)
- [Workload Identity for Dataproc on GKE](https://cloud.google.com/dataproc/docs/concepts/iam/dataproc-principals)
- [Terraform google_dataproc_cluster](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/dataproc_cluster)
- [Pulumi gcp.dataproc.Cluster](https://www.pulumi.com/registry/packages/gcp/api-docs/dataproc/cluster/)
