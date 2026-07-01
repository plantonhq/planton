# Data Platform InfraChart

This chart provisions an **analytics and data lake platform on Oracle Cloud Infrastructure**:

* Private VCN with NAT and service gateways (no internet-facing data services)
* Autonomous Data Warehouse (ADW) with auto-scaling for analytics workloads
* Object Storage bucket as the data lake with versioning and auto-tiering
* Kafka-compatible streaming (OCI Streaming) for real-time data ingestion
* Centralized log group for platform-wide observability

## Resources Created

| Resource | Kind | Condition |
|----------|------|-----------|
| Virtual Cloud Network | `OciVcn` | Always |
| Private Subnet | `OciSubnet` | Always |
| Autonomous Data Warehouse | `OciAutonomousDatabase` | Always |
| Data Lake Bucket | `OciObjectStorageBucket` | Always |
| Stream Pool + Stream | `OciStreamPool` | Always |
| Log Group | `OciLogGroup` | Always |

## Parameters

| Name | Description | Default |
|------|-------------|---------|
| `compartment_ocid` | OCI compartment OCID | — |
| `vcn_cidr` | VCN CIDR block | `10.0.0.0/16` |
| `subnet_cidr` | Private subnet CIDR | `10.0.1.0/24` |
| `db_name` | Data warehouse name | `datawarehouse` |
| `compute_count` | ECPU count | `4` |
| `storage_in_tbs` | Storage in TB | `1` |
| `admin_password` | ADMIN password | — |
| `bucket_name` | Data lake bucket | `data-lake` |
| `stream_pool_name` | Stream pool name | `data-ingestion` |
| `stream_partitions` | Partitions per stream | `1` |
| `stream_retention_hours` | Message retention (hours) | `24` |
| `log_retention_days` | Log retention (days) | `30` |

## Architecture

```
┌────────────────────────────────────────────────────────┐
│  Private VCN                                           │
│                                                        │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Private Subnet                                  │  │
│  │                                                  │  │
│  │  ┌─────────────────┐  ┌──────────────────────┐  │  │
│  │  │  Autonomous DW   │  │  Object Storage      │  │  │
│  │  │  (ADW - ECPU)    │  │  (Data Lake)         │  │  │
│  │  └─────────────────┘  └──────────────────────┘  │  │
│  │                                                  │  │
│  │  ┌─────────────────┐  ┌──────────────────────┐  │  │
│  │  │  Stream Pool     │  │  Log Group           │  │  │
│  │  │  (Kafka ingest)  │  │  (Observability)     │  │  │
│  │  └─────────────────┘  └──────────────────────┘  │  │
│  └──────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────┘
```

## Data Flow

1. **Ingest**: Producers write events to the Kafka-compatible stream pool
2. **Store**: Raw data lands in the Object Storage data lake bucket
3. **Analyze**: ADW queries data directly from Object Storage using external tables
4. **Monitor**: All operations logged to the centralized log group
