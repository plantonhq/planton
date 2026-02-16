---
title: "Metastore Integrated"
description: "A Dataproc on GKE virtual cluster with Hive Metastore integration and Spark History Server for catalog-based Spark SQL workloads and job history browsing."
type: "preset"
rank: "03"
presetSlug: "03-metastore-integrated"
componentSlug: "dataproc-virtual-cluster"
componentTitle: "Dataproc Virtual Cluster"
provider: "gcp"
icon: "package"
order: 3
---

# Metastore Integrated

A Dataproc on GKE virtual cluster with Hive Metastore integration and Spark History Server for catalog-based Spark SQL workloads and job history browsing.

## When to Use

- Spark SQL workloads that query Hive-managed tables
- Data lake architectures with shared schema catalog
- Environments where job history and debugging visibility is important
- Teams migrating from standard Dataproc clusters that used Hive Metastore

## Key Configuration Choices

- **Dataproc Metastore integration**: Spark jobs can access Hive tables without configuring a standalone metastore. Schema definitions are shared across all virtual clusters using the same metastore service.
- **Spark History Server**: Completed job logs are viewable through a web UI hosted on a dedicated Dataproc cluster. Event logging must be enabled via Spark properties.
- **Hive catalog enabled**: `spark.sql.catalogImplementation` set to `hive` so Spark SQL uses the metastore for table resolution
- **Event logging to GCS**: Spark writes event logs to a GCS path for the History Server to read
- **Two node pools**: DEFAULT + CONTROLLER on stable nodes; combined SPARK_DRIVER + SPARK_EXECUTOR on a larger pool
- **Autoscaling executor pool (1-20)**: Scales with workload demand
- **Spark 3.5**: Latest stable version

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-gcp-project-id>` | GCP project ID | GCP Console > Project Settings |
| `<your-gke-cluster-resource-id>` | Fully qualified GKE cluster ID | GCP Console > Kubernetes Engine > Clusters |
| `<your-kubernetes-namespace>` | Kubernetes namespace for Spark pods | Choose a name or use an existing namespace |
| `<your-staging-bucket-name>` | GCS bucket for staging artifacts | GCP Console > Cloud Storage > Buckets |
| `<your-default-pool-name>` | GKE node pool for controller workloads | GCP Console > Kubernetes Engine > Node Pools |
| `<your-executor-pool-name>` | GKE node pool for Spark drivers and executors | GCP Console > Kubernetes Engine > Node Pools |
| `<your-spark-event-log-gcs-path>` | GCS path for Spark event logs (e.g., `gs://my-bucket/spark-events`) | GCP Console > Cloud Storage |
| `<your-dataproc-metastore-service-resource-name>` | Fully qualified Dataproc Metastore service name (`projects/{project}/locations/{location}/services/{service}`) | GCP Console > Dataproc > Metastore |
| `<your-spark-history-server-cluster-resource-name>` | Fully qualified Dataproc cluster name running the History Server (`projects/{project}/regions/{region}/clusters/{cluster}`) | GCP Console > Dataproc > Clusters |

## Important Notes

- The Dataproc Metastore service must be in the same project and region as the virtual cluster.
- The Spark History Server cluster must be a standard Dataproc cluster (not a virtual cluster) with the Spark History Server component enabled.
- Event logs are written to GCS. Ensure the Spark pods' service account has write access to the event log GCS path.
- The metastore provides schema management only — actual table data resides in GCS or BigQuery.

## Related Presets

- **01-basic-spark-on-gke**: Minimal single-pool setup without auxiliary services
- **02-production-multi-pool**: Production setup with role-separated pools (no metastore)
