---
title: "Production Multi-Pool"
description: "A production-grade Dataproc on GKE virtual cluster with role-separated node pools, Spot executor autoscaling, and dedicated staging bucket."
type: "preset"
rank: "02"
presetSlug: "02-production-multi-pool"
componentSlug: "gcpdataprocvirtualcluster-research-and-design-documentation"
componentTitle: "GcpDataprocVirtualCluster - Research and Design Documentation"
provider: "gcp"
icon: "package"
order: 2
---

# Production Multi-Pool

A production-grade Dataproc on GKE virtual cluster with role-separated node pools, Spot executor autoscaling, and dedicated staging bucket.

## When to Use

- Production Spark workloads on shared GKE infrastructure
- Workloads requiring separate scaling for drivers and executors
- Cost-optimized batch processing with Spot VMs for executors
- Multi-team environments where namespace isolation is needed

## Key Configuration Choices

- **Three node pools with role separation**:
  - DEFAULT + CONTROLLER on stable on-demand nodes (e2-standard-4)
  - SPARK_DRIVER on dedicated on-demand nodes (n2-standard-4) — drivers must not be preempted
  - SPARK_EXECUTOR on Spot nodes (n2-standard-8) — executors tolerate preemption gracefully
- **Spot executor autoscaling (0-30)**: Scale to zero when no jobs are running; burst to 30 nodes under load
- **Spark dynamic allocation**: Automatically adjusts executor count per job based on workload
- **Explicit namespace**: Isolates Spark pods from other GKE workloads
- **Dedicated staging bucket**: Controlled artifact storage instead of auto-created bucket
- **Spark 3.5**: Latest stable version

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gcp-project-id>` | GCP project ID | GCP Console > Project Settings |
| `<your-gke-cluster-resource-id>` | Fully qualified GKE cluster ID | GCP Console > Kubernetes Engine > Clusters |
| `<your-kubernetes-namespace>` | Kubernetes namespace for Spark pods | Choose a name or use an existing namespace |
| `<your-staging-bucket-name>` | GCS bucket for staging artifacts | GCP Console > Cloud Storage > Buckets |
| `<your-default-pool-name>` | GKE node pool for controller workloads | GCP Console > Kubernetes Engine > Node Pools |
| `<your-driver-pool-name>` | GKE node pool for Spark drivers | GCP Console > Kubernetes Engine > Node Pools |
| `<your-executor-pool-name>` | GKE node pool for Spark executors | GCP Console > Kubernetes Engine > Node Pools |

## Important Notes

- Spot VMs may be preempted at any time. Spark's task retry mechanism handles this for executors. Never use Spot for drivers or controllers.
- The executor pool scales to 0 when idle — no cost when no Spark jobs are running.
- Adjust `maxNodeCount` based on your cluster's node quota and expected peak parallelism.
- The driver pool stays at min 1 node to avoid cold-start delays for job submissions.

## Related Presets

- **01-basic-spark-on-gke**: Minimal single-pool setup for development
- **03-metastore-integrated**: Adds Hive Metastore and Spark History Server
