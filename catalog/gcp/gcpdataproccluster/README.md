# GCP Dataproc Cluster

Deploys a Google Cloud Dataproc cluster (`google_dataproc_cluster`) for running Apache Spark, Hadoop, Hive, and related open-source data processing frameworks — either on dedicated Compute Engine VMs (the standard arm) or as Kubernetes pods on an existing GKE cluster (the Dataproc-on-GKE virtual arm).

## Overview

Dataproc is Google's managed Spark/Hadoop service. A cluster consists of master nodes (HDFS NameNode, YARN ResourceManager), primary workers (DataNodes, NodeManagers), and optional secondary workers (Spot/preemptible VMs for burst capacity). Clusters are frequently ephemeral — spun up for a batch job, torn down after — so lifecycle TTLs and autoscaling are first-class cost levers.

The spec mirrors the API's own shape with two mutually exclusive deployment arms:

- **`clusterConfig`** (Compute Engine) — dedicated VMs fully managed by Dataproc: node sizing and disks, networking and hardening, software and init actions, security (Kerberos or personal-cluster auth), CMEK, metastore attachment, metric collection, dedicated driver groups, and lifecycle TTLs.
- **`virtualClusterConfig`** (Dataproc-on-GKE) — Spark runs as pods on an existing GKE cluster, composing against `GcpGkeCluster` and `GcpGkeNodePool` by reference.

Omitting both arms is valid: GCP creates a default GCE-based cluster (2 workers, default machine types).

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDataprocCluster
metadata:
  name: my-spark-cluster
spec:
  region: us-central1
  clusterName: my-spark-cluster
  clusterConfig:
    masterConfig:
      machineType: n2-standard-4
    workerConfig:
      numInstances: 2
      machineType: n2-standard-4
    softwareConfig:
      imageVersion: "2.2-debian12"
    endpointConfig:
      enableHttpPortAccess: true
    lifecycleConfig:
      idleDeleteTtl: "1800s"
```

```shell
planton apply -f dataproc.yaml
```

This creates a cluster with 1 master, 2 workers, Spark 3.5, Component Gateway enabled, and auto-delete after 30 minutes idle.

## Configuration Options

### The GCE arm (`clusterConfig`)

| Category | Options |
|----------|---------|
| **Buckets** | `stagingBucket` / `tempBucket` — GCS buckets for job dependencies and shuffle/spill; auto-created when omitted; reference `GcpGcsBucket` |
| **Tier** | `clusterTier` — `CLUSTER_TIER_STANDARD` (default) or `CLUSTER_TIER_PREMIUM`; immutable |
| **Networking** | `gceConfig.network` XOR `gceConfig.subnetwork` (references resolve to self-links); `internalIpOnly`, `zone`, `tags`, `metadata`, `serviceAccountScopes` |
| **Identity** | `gceConfig.serviceAccount` — custom node identity; reference `GcpServiceAccount` |
| **Hardening** | `shieldedInstanceConfig` (secure boot, vTPM, integrity monitoring), `confidentialInstanceConfig` (N2D-only data-in-use encryption) |
| **Placement** | `reservationAffinity` (SPECIFIC_RESERVATION needs key + values), `nodeGroupAffinity` (sole-tenant node group) |
| **Masters** | `masterConfig` — 1 (standard) or 3 (HA) instances, machine type, disk config, accelerators, `minCpuPlatform`, `imageUri` |
| **Workers** | `workerConfig` — `numInstances` and `minNumInstances` update in place (manual scaling + autoscaler floor) |
| **Secondaries** | `secondaryWorkerConfig` — `preemptibility` (`SPOT` recommended), disk config, `instanceFlexibilityPolicy` (ranked machine types + standard/spot capacity mix) |
| **Disks** | per group: `bootDiskSizeGb` (>=10), `bootDiskType` (`pd-standard`/`pd-ssd`/`pd-balanced`), `numLocalSsds`, `localSsdInterface` (`scsi`/`nvme`) |
| **Software** | `softwareConfig` — `imageVersion`, `optionalComponents` (JUPYTER, DOCKER, TRINO, …), `properties` (`"prefix:property"` overrides) |
| **Init actions** | `initializationActions` — GCS scripts with per-script timeout |
| **Autoscaling** | `autoscalingPolicyUri` — references a `GcpDataprocAutoscalingPolicy` (resolves to the policy's full resource name); attach/swap/detach updates in place |
| **Encryption** | `encryptionKmsKeyName` — CMEK for all persistent disks; reference `GcpKmsKey` |
| **Security** | `securityConfig` — exactly one of `kerberosConfig` (Hadoop Secure Mode; secret fields are GCS URIs of KMS-encrypted files, never inline secrets) or `identityConfig` (personal-cluster user-to-service-account mapping) |
| **Web UIs** | `endpointConfig.enableHttpPortAccess` — Component Gateway for Spark UI, YARN, Jupyter |
| **Lifecycle** | `lifecycleConfig.idleDeleteTtl` / `autoDeleteTime` — both update in place |
| **Metastore** | `metastoreConfig.dataprocMetastoreService` — attach a persistent Hive metastore by resource name |
| **Metrics** | `dataprocMetricConfig.metrics` — sources: `MONITORING_AGENT_DEFAULTS`, `HDFS`, `SPARK`, `YARN`, `SPARK_HISTORY_SERVER`, `HIVESERVER2`; optional per-source overrides |
| **Driver groups** | `auxiliaryNodeGroups` — dedicated `DRIVER`-role capacity so heavy Spark drivers don't compete with the master |

### The virtual arm (`virtualClusterConfig`)

| Category | Options |
|----------|---------|
| **Staging** | `stagingBucket` — auto-created when omitted |
| **Target GKE** | `kubernetesClusterConfig.gkeClusterConfig.gkeClusterTarget` — reference `GcpGkeCluster` (required) |
| **Node pools** | `nodePoolTarget[]` — pool reference (`GcpGkeNodePool`) + roles (`DEFAULT`, `CONTROLLER`, `SPARK_DRIVER`, `SPARK_EXECUTOR`); optional `nodePoolConfig` sizes pools Dataproc creates (locations, autoscaling bounds, machine type, preemptible XOR spot) |
| **Namespace** | `kubernetesNamespace` — where Spark pods land; derived from the cluster name when omitted |
| **Software** | `kubernetesSoftwareConfig.componentVersion` (SPARK required, e.g. `"3.5-dataproc-17"`) + `properties` |
| **Shared services** | `auxiliaryServicesConfig` — persistent metastore and a Spark History Server hosted on another Dataproc cluster (reference by `cluster_id` output) |

### Spec-level fields

| Field | Notes |
|-------|-------|
| `projectId` | Optional; omitted rides the provider's default project; reference `GcpProject` |
| `region` | Required; pattern accepts multi-digit regions (e.g. `europe-west12`); immutable |
| `clusterName` | Required; 2-55 chars, lowercase/digits/hyphens; immutable |
| `gracefulDecommissionTimeout` | YARN drain window on scale-down (e.g. `"3600s"`); GCE arm only |
| `labels` | User labels merged beneath platform attribution labels; **rejected on the virtual arm** (API limitation, validated pre-deploy) |

## Mutability

| Surface | Mutability |
|---|---|
| `labels` | In place |
| `workerConfig.numInstances`, `secondaryWorkerConfig.numInstances` | In place (manual scaling) |
| `workerConfig.minNumInstances` | In place |
| `autoscalingPolicyUri` (attach/swap/detach) | In place |
| `lifecycleConfig.idleDeleteTtl` / `autoDeleteTime` | In place |
| Everything else on the GCE arm | Recreates the cluster |
| The entire virtual arm | Immutable — any change replaces the virtual cluster (the underlying GKE cluster and pools are untouched) |

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | string | Fully qualified resource name (`projects/{p}/regions/{r}/clusters/{c}`) — the composition handle, including another cluster's `sparkHistoryServerConfig` |
| `cluster_name` | string | Short cluster name |
| `staging_bucket` | string | The staging bucket in use (user-supplied or auto-created), on either arm |

## Important Notes

- **Exactly one arm (or none)**: `clusterConfig` and `virtualClusterConfig` are mutually exclusive; omitting both yields a default GCE cluster.
- **Labels and the virtual arm**: the Dataproc API rejects user labels on virtual clusters — the spec validation catches this before deploy.
- **Ephemeral by design**: set `lifecycleConfig.idleDeleteTtl` on anything that isn't a permanent shared cluster; both TTLs tune in place.
- **Scaling without recreation**: worker counts, the autoscaler floor, and the autoscaling-policy attachment are the in-place levers — plan production scaling around them.
- **Kerberos secrets are paths**: every `*Uri` field in `kerberosConfig` is a GCS URI of a KMS-encrypted file; no literal secret material enters the manifest.
- **Metastore accepts literals**: `metastoreConfig` takes a full Dataproc Metastore service resource name today; references attach when a metastore-service kind lands in the catalog.

### Deliberately not modeled (recorded reasons)

Everything else on `google_dataproc_cluster` at the pinned provider is representable — including `clusterType` (SINGLE_NODE / ZERO_SCALE), the Lightning `engine`, stop-lifecycle TTLs, resource manager tags, boot-disk provisioned IOPS/throughput and additional attached disks (`diskConfig.attachedDisks`) on every node role, master and primary-worker `instanceFlexibilityPolicy` with per-selection disk shapes, the Confidential Compute technology (`confidentialInstanceType`: SEV, SEV_SNP, TDX), and `deletionPolicy`. The recorded exclusions:

- **`enable_confidential_compute`** — the provider-deprecated boolean predecessor of `confidentialInstanceType`; the spec keeps it so existing manifests keep working, and sends it only when set.
- **Attached disks on auxiliary (driver) node groups** — the provider defines additional disks on master, worker, and secondary worker nodes only; the spec rejects them on auxiliary groups.
- **Dataproc IAM member/binding/policy trios** — resource-scoped IAM deferred.
- **`google_dataproc_job` / `batch` / `workflow_template` / `session_template`** — workloads, not infrastructure; Serverless Batches is a future kind candidate.
- **Dataproc Metastore service** — future kind candidate; `metastoreConfig` accepts literal resource names today.

## Related Components

- **GcpDataprocAutoscalingPolicy** — the shared autoscaling policy `autoscalingPolicyUri` references
- **GcpGcsBucket** — staging and temp buckets for job artifacts
- **GcpVpcNetwork / GcpSubnetwork** — network placement for cluster nodes
- **GcpServiceAccount** — custom IAM identity for cluster VMs
- **GcpKmsKey** — CMEK for disk encryption and Kerberos file decryption
- **GcpGkeCluster / GcpGkeNodePool** — the compute substrate of the virtual arm

## Additional Resources

- [Dataproc Documentation](https://cloud.google.com/dataproc/docs)
- [Dataproc image versions](https://cloud.google.com/dataproc/docs/concepts/versioning/dataproc-versions)
- [Dataproc on GKE](https://cloud.google.com/dataproc/docs/guides/dpgke/dataproc-gke-overview)
- [Autoscaling clusters](https://cloud.google.com/dataproc/docs/concepts/configuring-clusters/autoscaling)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
