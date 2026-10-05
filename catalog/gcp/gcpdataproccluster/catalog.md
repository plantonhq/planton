# GCP Dataproc Cluster

Deploys a Dataproc cluster for Apache Spark, Hadoop, and related data processing frameworks — either the standard Compute Engine arm (configurable master and worker nodes, preemptible/spot secondary workers, software image versioning, optional components like Jupyter/Flink/Trino, in-cluster security, lifecycle management for cost control, and CMEK encryption) or the Dataproc-on-GKE virtual arm, where Spark workloads run as pods on an existing GKE cluster. Most configuration is immutable after creation: the in-place exceptions on the GCE arm are worker counts, the autoscaling policy attachment, the lifecycle TTLs, and labels — the virtual arm is fully immutable.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Dataproc API enablement** -- `dataproc.googleapis.com` is enabled in the target project (never disabled on destroy, so tearing down one cluster cannot break the rest of the project)
- **Dataproc Cluster** -- a managed cluster resource in the specified GCP project and region, configured with the chosen Dataproc image version, master/worker topology, and software components
- **Master Nodes** -- 1 instance for standard mode or 3 instances for high-availability mode, with configurable machine type, boot disk, local SSDs, and accelerators
- **Primary Worker Nodes** -- on-demand VMs for persistent compute capacity, with configurable machine type, disk, accelerators, and autoscaling minimum
- **Secondary Worker Nodes** -- created only when `clusterConfig.secondaryWorkerConfig` is set; preemptible or spot VMs providing cost-optimized burst capacity
- **Software Configuration** -- Dataproc image version, optional components (Jupyter, Flink, Presto, Trino), and framework property overrides for Spark, Hadoop, YARN, and HDFS
- **Initialization Actions** -- created only when `clusterConfig.initializationActions` is set; startup scripts that run on all nodes during cluster creation
- **Component Gateway** -- created only when `clusterConfig.endpointConfig.enableHttpPortAccess` is true; provides authenticated HTTPS access to Spark UI, YARN, HDFS, and optional component web interfaces
- **Lifecycle Configuration** -- created only when `clusterConfig.lifecycleConfig` is set; automatic cluster deletion or stop after idle timeout or at a scheduled time
- **CMEK Encryption** -- created only when `clusterConfig.encryptionKmsKeyName` is provided; encrypts persistent disks with a customer-managed Cloud KMS key
- **In-Cluster Security** -- created only when `clusterConfig.securityConfig` is set; exactly one of Kerberos (Hadoop Secure Mode) or personal-cluster identity mapping
- **Autoscaling Attachment** -- created only when `clusterConfig.autoscalingPolicyUri` references a GcpDataprocAutoscalingPolicy; attaching, swapping, or detaching updates in place
- **Virtual Cluster (Dataproc-on-GKE)** -- created only when `virtualClusterConfig` is set instead of `clusterConfig`; registers Spark workloads as pods on an existing GKE cluster with node-pool role mapping and shared metastore/Spark History Server attachments
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance; user labels from `spec.labels` propagate to cluster VMs (not supported by the API on virtual clusters)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the cluster will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The module enables the Dataproc API itself; the **Compute Engine API** must already be enabled for the GCE arm's VMs.
- **A VPC network or subnetwork** for cluster node placement. Using a subnetwork is recommended for production clusters with controlled IP ranges. Provide directly or reference GcpVpcNetwork/GcpSubnetwork Infra Components via ValueFromRef.
- **A custom service account** (recommended for production) with minimal permissions for cluster VMs. The default Compute Engine service account works for development.
- **Cloud NAT or Private Google Access** (only with `internalIpOnly: true`) so private nodes can reach the internet for container image pulls and package installs.
- **An existing GKE cluster** (only for the virtual arm) that Dataproc registers Spark workloads onto — reference a GcpGkeCluster Infra Component.

## Deploy

### Console

Open the deployment store, find **GCP Dataproc Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev Jupyter** preset in the [Presets](#presets) tab to pre-populate a development cluster with Jupyter notebooks.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDataprocCluster
metadata:
  name: analytics-spark
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  region: us-central1
  clusterName: analytics-spark
```

```shell
planton apply -f dataproc-cluster.yaml
```

This creates a standard cluster with GCP defaults: 1 master and 2 workers on default machine types, 500 GB pd-standard disks, the latest stable Dataproc image, and no lifecycle management — add `lifecycleConfig` before leaving a cluster like this unattended. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the cluster to VPC infrastructure deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  clusterConfig:
    gceConfig:
      subnetwork:
        valueFrom:
          kind: GcpSubnetwork
          name: dataproc-subnet
          fieldPath: status.outputs.subnetwork_self_link
      serviceAccount:
        valueFrom:
          kind: GcpServiceAccount
          name: dataproc-sa
          fieldPath: status.outputs.email
    stagingBucket:
      valueFrom:
        kind: GcpGcsBucket
        name: spark-staging
        fieldPath: status.outputs.bucket_id
    encryptionKmsKeyName:
      valueFrom:
        kind: GcpKmsKey
        name: dataproc-key
        fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project, subnet, service account, bucket, and KMS key first, then provisions the Dataproc cluster with all resolved values.

## Key Configuration

These are the most important decisions when configuring a Dataproc cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Cluster topology** -- Set `clusterConfig.masterConfig.numInstances` to 1 for standard mode or 3 for high-availability mode. HA mode tolerates a single master failure but triples master costs. Worker count is set via `clusterConfig.workerConfig.numInstances` (default 2) and is one of the few fields updatable in place.

**Machine type XOR flexibility policy** -- `machineType` and `instanceFlexibilityPolicy` are alternative provisioning contracts, never complements: the API silently drops a paired machine type from the stored config, and because machine type is create-only the pairing re-plans as a whole-cluster REPLACEMENT on every subsequent apply. Rank machine types inside the flexibility policy's `instanceSelectionList` instead of setting both.

**Secondary workers for cost optimization** -- Configure `clusterConfig.secondaryWorkerConfig` with `preemptibility: SPOT` for cost-optimized burst capacity (the API's default is legacy PREEMPTIBLE, and the choice is immutable). Spot VMs can be preempted at any time, so use them for fault-tolerant batch workloads with Spark's dynamic allocation enabled; `provisioningModelMix` protects a baseline of on-demand capacity under the spot burst.

**Lifecycle management** -- Set `clusterConfig.lifecycleConfig.idleDeleteTtl` (e.g., `"1800s"` for 30 minutes) to auto-delete idle clusters. Critical for ephemeral batch clusters to avoid runaway costs. Use `autoDeleteTime` for time-boxed clusters with a known end date. The stop variants — `idleStopTtl` / `autoStopTime` — shut the VMs down instead of deleting, keeping the cluster restartable; and `deletionPolicy: ABANDON` releases a cluster from IaC management without destroying it.

**Structural type and engine** -- `clusterConfig.clusterType: SINGLE_NODE` runs everything on one VM (the modern form of the zero-workers property); `ZERO_SCALE` keeps only the control plane warm and provisions workers on demand. `engine: LIGHTNING` enables the Lightning Engine, Google's accelerated Spark runtime (premium tier). Both are immutable.

**Software and components** -- Set `clusterConfig.softwareConfig.imageVersion` (e.g., `"2.2-debian12"`) and add optional components like `JUPYTER`, `FLINK`, or `PRESTO` via `optionalComponents`. Override framework properties via the `properties` map (e.g., `"spark:spark.executor.memory": "12g"`).

**Network isolation** -- Set `clusterConfig.gceConfig.internalIpOnly: true` for private nodes with no external IPs. Requires Cloud NAT or Private Google Access for internet connectivity. Combine with `tags` for firewall rule targeting.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (optional) | `clusterConfig.gceConfig.network` | `status.outputs.network_self_link` |
| **GcpSubnetwork** (optional) | `clusterConfig.gceConfig.subnetwork` | `status.outputs.subnetwork_self_link` |
| **GcpServiceAccount** (optional) | `clusterConfig.gceConfig.serviceAccount` | `status.outputs.email` |
| **GcpGcsBucket** (optional) | `clusterConfig.stagingBucket` | `status.outputs.bucket_id` |
| **GcpGcsBucket** (optional) | `clusterConfig.tempBucket` | `status.outputs.bucket_id` |
| **GcpKmsKey** (optional) | `clusterConfig.encryptionKmsKeyName` | `status.outputs.key_id` |
| **GcpDataprocAutoscalingPolicy** (optional) | `clusterConfig.autoscalingPolicyUri` | `status.outputs.name` |
| **GcpKmsKey** (Kerberos only) | `clusterConfig.securityConfig.kerberosConfig.kmsKeyUri` | `status.outputs.key_id` |
| **GcpGkeCluster** (virtual arm) | `virtualClusterConfig.kubernetesClusterConfig.gkeClusterConfig.gkeClusterTarget` | `status.outputs.cluster_id` |
| **GcpGkeNodePool** (virtual arm) | `virtualClusterConfig...nodePoolTarget[].nodePool` | `status.outputs.node_pool_id` |
| **KubernetesNamespace** (virtual arm, optional) | `virtualClusterConfig.kubernetesClusterConfig.kubernetesNamespace` | `spec.name` |
| **GcpGcsBucket** (virtual arm) | `virtualClusterConfig.stagingBucket` | `status.outputs.bucket_id` |
| **GcpDataprocCluster** (virtual arm) | `virtualClusterConfig.auxiliaryServicesConfig.sparkHistoryServerConfig.dataprocCluster` | `status.outputs.cluster_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cluster_id` | Fully qualified cluster resource name (`projects/{p}/regions/{r}/clusters/{c}`) | Dataproc job submissions, workflow template references, another cluster's Spark History Server attachment |
| `cluster_name` | Short cluster name | Display, logging, job targeting |
| `staging_bucket` | GCS bucket used for staging job dependencies (user-supplied or GCP auto-created) | Job dependency uploads, output inspection |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev with Jupyter** -- Single master, 2 workers, Jupyter notebooks enabled, Component Gateway for web UI access, and 30-minute idle auto-delete. Optimized for interactive data exploration and prototyping. Start from the **Dev Jupyter** preset.

**HA production** -- 3 masters for high availability, workers with SSD disks and NVMe local SSDs, Shielded VMs, internal-only networking, CMEK encryption, custom service account, and OSS metrics into Cloud Monitoring. Suitable for production ETL pipelines and long-running Spark applications. Start from the **HA Production** preset.

**Cost-optimized batch** -- A small on-demand base with a spot secondary group using machine-type flexibility and a standard/spot capacity mix, an attached autoscaling policy, and aggressive idle auto-delete. Suitable for ephemeral batch ETL jobs where cost efficiency outweighs preemption risk. Start from the **Cost-Optimized Batch** preset.

**Spark on GKE** -- A Dataproc-on-GKE virtual cluster: Spark workloads run as pods on an existing GKE cluster referenced by ValueFromRef, sharing its capacity, autoscaling, and operational tooling. Start from the **Spark on GKE** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the cluster is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network for cluster node placement
- [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) -- provides the subnet with controlled IP ranges for cluster nodes
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- provides the identity for cluster VMs
- [**GCP GCS Bucket**](/infra-catalog/gcp-gcs-bucket) -- provides staging and temp buckets for job dependencies and shuffle data
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides the CMEK encryption key for cluster persistent disks
- [**GCP Dataproc Autoscaling Policy**](/infra-catalog/gcp-dataproc-autoscaling-policy) -- provides the reusable worker-scaling contract attached via `autoscalingPolicyUri`
- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) -- hosts the virtual arm's Spark pods
- [**GCP GKE Node Pool**](/infra-catalog/gcp-gke-node-pool) -- provides the node pools the virtual arm schedules Dataproc roles onto
