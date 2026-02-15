# GcpDataprocCluster

Provision and manage standard (GCE-based) Google Cloud Dataproc clusters for running Apache Spark, Hadoop, Hive, Pig, and related open-source data processing frameworks.

## Overview

GcpDataprocCluster creates a managed Dataproc cluster on Google Compute Engine VMs. Clusters consist of master nodes (HDFS NameNode, YARN ResourceManager), primary worker nodes (DataNodes, NodeManagers), and optional secondary workers (preemptible or Spot VMs for burst capacity).

Common use cases:
- **Batch ETL**: Large-scale data transformations with Spark
- **Interactive analysis**: Jupyter notebooks for data exploration
- **ML training**: Distributed machine learning with Spark MLlib and GPU accelerators
- **Stream processing**: Spark Structured Streaming for real-time data pipelines

## Key Configuration

### Cluster Topology

- **Master nodes**: 1 for standard mode, 3 for high availability (HA)
- **Primary workers**: On-demand VMs for baseline compute capacity
- **Secondary workers**: Preemptible or Spot VMs for cost-optimized burst capacity

### Software

- **Image version**: Controls the versions of Spark, Hadoop, and other frameworks
- **Optional components**: JUPYTER, DOCKER, PRESTO, ZEPPELIN, FLINK, TRINO
- **Properties**: Override any Hadoop, Spark, YARN, or Hive configuration

### Networking

- **VPC/Subnetwork**: Place the cluster in a specific network
- **Internal IP only**: Restrict nodes to private IP addresses
- **Tags**: Apply GCE network tags for firewall rule targeting

### Cost Management

- **Lifecycle config**: Auto-delete clusters after idle timeout or at a scheduled time
- **Spot secondary workers**: Up to 80% cost savings for fault-tolerant workloads
- **Graceful decommission**: Allow running jobs to complete before scaling down

### Security

- **CMEK encryption**: Customer-managed keys for persistent disk encryption
- **Service account**: Custom IAM identity with least-privilege access
- **Internal IP only**: No public internet exposure

## Outputs

After deployment, the following outputs are available:

| Output | Description |
|---|---|
| `cluster_id` | Fully qualified cluster resource name |
| `cluster_name` | Short name of the cluster |
| `cluster_uuid` | Server-generated unique identifier |
| `staging_bucket` | GCS bucket used for staging job dependencies |

## Related Components

- **GcpGcsBucket**: Staging and temp bucket for job artifacts
- **GcpVpc / GcpSubnetwork**: Network placement for cluster nodes
- **GcpServiceAccount**: Custom IAM identity for cluster VMs
- **GcpKmsKey**: Customer-managed encryption keys for disk encryption
- **GcpDataprocVirtualCluster**: Dataproc-on-GKE for Kubernetes-native Spark (separate component)

## Deliberate Exclusions

The following features are excluded from v1 to keep the API focused on the 90% use case:

- **Virtual cluster config** (Dataproc-on-GKE): See GcpDataprocVirtualCluster
- **Kerberos security config**: Enterprise Hadoop security (15+ fields)
- **Auxiliary node groups**: Driver node separation
- **Metastore config**: External Hive Metastore service
- **Dataproc metric config**: Custom monitoring metrics
- **Shielded instance config**: Secure boot, vTPM, integrity monitoring
- **Reservation affinity**: Capacity reservation targeting
- **Confidential compute**: Confidential VMs
