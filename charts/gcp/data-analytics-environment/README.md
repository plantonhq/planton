# GCP Data Analytics Environment

Provisions a complete data engineering platform with a Dataproc Spark/Hadoop cluster, BigQuery dataset for analytics warehousing, Pub/Sub topic for streaming data ingestion, GCS bucket for staging and data lake storage, a dedicated service account, and private VPC networking.

This chart gives a data engineering team everything they need to start building ETL pipelines, running Spark jobs, querying data in BigQuery, and ingesting streaming events -- all deployed in under 20 minutes with proper networking, IAM, and cost controls.

## Architecture

```
                  ┌──────────────────────────────────────────┐
                  │  Networking                               │
                  │                                          │
                  │  ┌──────────┐     ┌───────────────────┐  │
                  │  │  GcpVpc  │────▶│  GcpSubnetwork    │  │
                  │  │          │     │  Private Google    │  │
                  │  └──────────┘     │  Access enabled    │  │
                  │                   └─────────┬──────────┘  │
                  └─────────────────────────────│─────────────┘
                                                │
                                                ▼
                               ┌────────────────────────────┐
                               │   GcpDataprocCluster       │
                               │   (Spark + Hadoop + Jupyter)│
                               │   1 master + N workers     │
                               └──────────┬─────────────────┘
                                          │ reads/writes
                         ┌────────────────┼────────────────┐
                         ▼                ▼                ▼
                  ┌────────────┐  ┌────────────┐  ┌────────────┐
                  │GcpGcsBucket│  │GcpBigQuery │  │GcpPubSub   │
                  │ (data lake)│  │  Dataset   │  │  Topic     │
                  └────────────┘  └────────────┘  └────────────┘

                               ┌────────────────────────────┐
                               │   GcpServiceAccount        │
                               │   (dataproc.worker +       │
                               │    bigquery + storage +    │
                               │    pubsub.subscriber)      │
                               └────────────────────────────┘
```

## Dependency Graph

```
Layer 0 (parallel):  GcpVpc, GcpServiceAccount, GcpGcsBucket, GcpBigQueryDataset, GcpPubSubTopic
Layer 1 (dep VPC):   GcpSubnetwork
Layer 2 (dep all):   GcpDataprocCluster
```

## Included Cloud Resources

| Resource | Kind | Group | Purpose |
|----------|------|-------|---------|
| VPC Network | `GcpVpc` | network | Private networking for the cluster |
| Subnetwork | `GcpSubnetwork` | network | Subnet with Private Google Access |
| Service Account | `GcpServiceAccount` | identity | Cluster node identity with data access roles |
| GCS Bucket | `GcpGcsBucket` | storage | Dataproc staging, data lake, intermediate artifacts |
| BigQuery Dataset | `GcpBigQueryDataset` | storage | Analytics warehouse for processed data |
| Pub/Sub Topic | `GcpPubSubTopic` | messaging | Streaming data ingestion endpoint |
| Dataproc Cluster | `GcpDataprocCluster` | compute | Managed Spark/Hadoop cluster |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| `gcp_project_id` | GCP project ID | `my-gcp-project` | Yes |
| `region` | GCP region | `us-central1` | Yes |
| `vpc_name` | VPC network name | `data-analytics-vpc` | Yes |
| `subnet_cidr` | Subnet CIDR (/20 recommended for Dataproc) | `10.0.0.0/20` | Yes |
| `service_account_id` | Service account ID | `data-analytics-sa` | Yes |
| `bucket_name` | GCS bucket name (globally unique) | `my-project-data-lake` | Yes |
| `dataset_id` | BigQuery dataset ID | `analytics_warehouse` | Yes |
| `topic_name` | Pub/Sub topic name | `analytics-ingest` | Yes |
| `cluster_name` | Dataproc cluster name | `analytics-cluster` | Yes |
| `master_machine_type` | Master node machine type | `n2-standard-4` | Yes |
| `worker_count` | Number of worker nodes | `2` | Yes |
| `worker_machine_type` | Worker node machine type | `n2-standard-4` | Yes |
| `image_version` | Dataproc image version | `2.2-debian12` | Yes |
| `jupyterEnabled` | Install Jupyter on the cluster | `true` | No |
| `idle_delete_ttl` | Auto-delete after idle duration (e.g., `3600s`) | `3600s` | No |

## Service Account Roles

| Role | Purpose |
|------|---------|
| `roles/dataproc.worker` | Required for Dataproc cluster nodes |
| `roles/bigquery.dataEditor` | Read/write BigQuery tables |
| `roles/bigquery.jobUser` | Run BigQuery queries from Spark |
| `roles/storage.objectAdmin` | Read/write GCS bucket objects |
| `roles/pubsub.subscriber` | Read from Pub/Sub topics via Spark Streaming |

## Cost Management

The chart includes built-in cost controls:

- **Auto-delete on idle** (`idle_delete_ttl`): The cluster automatically deletes itself after the specified idle period (default: 1 hour). This prevents forgotten clusters from running up bills.
- **Internal IP only**: Cluster nodes have no external IPs, reducing network costs.
- Set `idle_delete_ttl` to an empty string to disable auto-delete for long-running clusters.

## Typical Data Flow

```
Events → Pub/Sub Topic → Spark Streaming → BigQuery
                       → Spark Batch → GCS (data lake) → BigQuery
```

1. **Ingest**: Publish events to the Pub/Sub topic from applications or IoT devices
2. **Process**: Spark jobs on Dataproc read from Pub/Sub and GCS, transform data
3. **Store**: Write results to BigQuery for analytics and dashboards
4. **Archive**: Raw data stored in GCS bucket for long-term retention

## Important Notes

- `cluster_name` and `region` are **immutable** after creation.
- The GCS bucket serves double duty: Dataproc staging AND data lake storage. For large-scale deployments, consider deploying a separate staging bucket.
- The Dataproc cluster uses internal IPs only. Access the Jupyter UI via the Component Gateway (enabled by default) through the GCP console or `gcloud` SSH tunnel.
- Subnet CIDR `/20` provides 4,094 IPs -- sufficient for clusters up to ~1,000 nodes. Use a larger CIDR for bigger deployments.
