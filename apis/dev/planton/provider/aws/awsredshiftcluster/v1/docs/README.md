# AWS Redshift Cluster — Architecture and Design

## Overview

Amazon Redshift is a petabyte-scale, fully managed columnar data warehouse
service optimized for online analytical processing (OLAP). It uses massively
parallel processing (MPP) to execute complex SQL queries across large datasets
in seconds, making it the backbone of many enterprise analytics platforms.

Redshift stores data in a columnar format, which dramatically reduces I/O for
analytical queries that typically touch a small number of columns across
billions of rows. Combined with zone maps, sort keys, and automatic compression
encoding, Redshift delivers query performance that is orders of magnitude faster
than row-oriented databases for analytical workloads.

## The Evolution of Cloud Data Warehousing

### Traditional Data Warehouses (Pre-Cloud)

Before cloud data warehousing, organizations ran analytical workloads on
expensive on-premises appliances like Teradata, Oracle Exadata, or Netezza.
These systems required:

- **Large upfront capital expenditure** — Hardware procurement cycles of 3-6 months.
- **Fixed capacity** — Over-provisioning was common to handle peak loads.
- **Manual tuning** — DBAs spent significant time on distribution keys, sort
  keys, compression, and vacuuming.
- **Complex ETL pipelines** — Data movement from OLTP systems to the warehouse
  was fragile and slow.

### First-Generation Cloud Warehouses (2012-2018)

Amazon Redshift launched in 2013 as one of the first cloud data warehouses. It
democratized analytical databases by offering:

- **Pay-per-hour** pricing instead of million-dollar appliance purchases.
- **Elastic resize** — Add or remove nodes (with brief downtime).
- **Managed infrastructure** — No hardware procurement, patching, or backups.
- **SQL compatibility** — PostgreSQL wire protocol for tool ecosystem compatibility.

### Modern Architecture: Decoupled Compute and Storage (2019+)

The introduction of RA3 nodes in late 2019 marked Redshift's evolution to a
modern architecture with decoupled compute and storage:

- **Managed storage** — Data automatically tiers between local NVMe SSDs and S3.
- **Independent scaling** — Scale compute (node count/type) without moving data.
- **Data sharing** — Share live data between clusters without copying.
- **Redshift Spectrum** — Query S3 data lakes directly from the warehouse.
- **AQUA (Advanced Query Accelerator)** — Hardware-accelerated caching on RA3 nodes.

This evolution makes RA3 the recommended node family for nearly all new
deployments, which is reflected in this component's design.

## Deployment Methods Landscape

### AWS Console

The Redshift console provides a wizard for cluster creation. While suitable for
one-off exploration, it lacks reproducibility, version control, and integration
with CI/CD pipelines. Configuration drift is inevitable.

### AWS CLI

The `aws redshift create-cluster` command exposes ~60 parameters. It's
scriptable but requires manual orchestration of dependent resources (subnet
groups, security groups, parameter groups, IAM roles) and has no built-in state
management.

### CloudFormation

`AWS::Redshift::Cluster` supports the full API surface. However, CloudFormation's
JSON/YAML is verbose, the drift detection is slow, and cross-stack references are
awkward. Module packaging and reuse are limited.

### Terraform

The `aws_redshift_cluster` resource in the AWS provider is mature and widely
used. Terraform's state management, plan/apply workflow, and module ecosystem
make it the most popular IaC tool for Redshift. However, Terraform modules still
expose the full complexity of the underlying API.

### Pulumi

Pulumi's `aws.redshift.Cluster` resource mirrors the Terraform provider (both
use the AWS SDK). Pulumi adds general-purpose programming languages, enabling
richer abstractions and type safety. The tradeoff is a smaller ecosystem of
reusable modules compared to Terraform.

### The Planton Approach

Planton wraps both Terraform and Pulumi behind a declarative YAML manifest with
a strongly-typed protobuf schema. The goals:

1. **80/20 surface area** — Expose the 20% of fields that cover 80% of real-world
   use cases. Advanced knobs are omitted from v1 and can be added based on demand.
2. **Bundled resources** — A single manifest creates the cluster plus its
   networking, security, parameter, and logging resources. No need to wire
   five separate resources together.
3. **Cross-resource references** — `StringValueOrRef` fields allow referencing
   outputs from other Planton resources (e.g., VPC subnet IDs, KMS key ARNs)
   without hardcoding values.
4. **Validation at authoring time** — Protobuf CEL expressions catch
   misconfigurations (e.g., missing `final_snapshot_identifier` when
   `skip_final_snapshot` is false) before any cloud API call.
5. **IaC-engine agnostic** — The same manifest deploys via Pulumi or Terraform.

## What's Included in v1

### Core Cluster Configuration

| Feature | Why Included |
|---------|-------------|
| Node type selection | Fundamental — determines compute/storage profile |
| Multi-node clusters | Required for production scale-out |
| Database name + admin user | Every cluster needs these |
| Managed password (Secrets Manager) | Best practice for production, eliminates static secrets |
| Port configuration | Needed for non-standard network environments |

### Networking

| Feature | Why Included |
|---------|-------------|
| Subnet group (auto-created from subnet IDs) | ~90% of deployments need a custom subnet group |
| Managed security group (from SG IDs or CIDRs) | Most common networking pattern |
| Associate existing security groups | Covers the "bring your own SG" pattern |
| Enhanced VPC routing | Required for compliance (PCI, HIPAA, SOC2) |
| Multi-AZ | Production HA requirement for RA3 clusters |
| Public accessibility toggle | Needed for external BI tool access |

### Encryption

| Feature | Why Included |
|---------|-------------|
| At-rest encryption (service key) | AWS default, always recommended |
| Customer-managed KMS key | Required for regulated workloads |
| KMS-encrypted Secrets Manager secret | End-to-end encryption for credentials |

### IAM Integration

| Feature | Why Included |
|---------|-------------|
| IAM role attachment (up to 10) | Required for COPY, UNLOAD, Spectrum |
| Default IAM role | Simplifies SQL queries that access AWS services |

### Snapshots and Lifecycle

| Feature | Why Included |
|---------|-------------|
| Automated snapshot retention | Core backup strategy |
| Final snapshot control | Critical for production data protection |

### Maintenance

| Feature | Why Included |
|---------|-------------|
| Maintenance window | Production teams need predictable maintenance |
| Version upgrade control | Prevents unexpected engine changes |
| Maintenance track (current/trailing) | Common pattern for staging vs. production |
| Apply immediately toggle | Needed for urgent changes |

### Audit Logging

| Feature | Why Included |
|---------|-------------|
| S3 log destination | Traditional logging path, well-understood |
| CloudWatch log destination | Modern logging path, integrates with alarms |
| Log type selection | Granular control over audit scope |

### Parameter Group

| Feature | Why Included |
|---------|-------------|
| Inline parameter creation | Covers SSL enforcement, activity logging, WLM |
| Existing parameter group reference | Supports shared parameter management |

## What's Excluded from v1

| Feature | Why Excluded | Adoption |
|---------|-------------|----------|
| Elastic resize scheduling | Newer API, complex scheduling logic | <10% |
| Concurrency scaling configuration | WLM-specific, advanced tuning | <15% |
| Snapshot copy to another region | Cross-region DR, complex setup | <10% |
| Snapshot schedule | Cron-like expressions, niche requirement | <15% |
| Redshift Serverless | Different resource model entirely | Separate component |
| Classic resize | Deprecated in favor of elastic resize | Legacy |
| HSM encryption | Hardware Security Module, very niche | <5% |
| Aqua configuration | Auto-enabled on RA3, no user config needed | N/A |
| Deferred maintenance window | Rarely used, maintenance is brief | <5% |
| Endpoint access (VPC endpoint) | Redshift-managed VPC endpoints | <10% |
| Partner integrations | Third-party data sharing, ETL connectors | Varies |
| Usage limits | Spectrum/concurrency scaling cost controls | <10% |
| Workload Management (WLM) JSON | Complex nested JSON parameter | <20% |

These features may be added in v2 based on community demand.

## Production Best Practices

### Node Type Selection

| Node Family | Best For | Storage Model |
|-------------|----------|--------------|
| `dc2.large` | Development, small datasets (<500 GB) | Local SSD only |
| `ra3.xlplus` | Small-medium production (1-4 nodes, up to ~32 TB managed) | Managed storage (SSD + S3) |
| `ra3.4xlarge` | Medium production (2-32 nodes) | Managed storage |
| `ra3.16xlarge` | Large production (2-128 nodes, petabyte-scale) | Managed storage |

**Recommendation**: Always use RA3 for production. DC2 is suitable only for
development clusters where cost is the primary concern and dataset sizes are small.

RA3 advantages over DC2:
- Independent compute and storage scaling
- Automatic data tiering (hot data on SSD, warm data on S3)
- Data sharing between clusters
- AQUA acceleration (hardware-level caching)
- Multi-AZ support

### Encryption

Always enable encryption. The `encrypted` field defaults to `true`, which uses
the AWS-managed Redshift service key (no additional cost).

For regulated workloads (PCI-DSS, HIPAA, SOC2):
- Use a customer-managed KMS key via `kmsKeyId`
- Enable KMS encryption for the Secrets Manager secret via `masterPasswordSecretKmsKeyId`
- Use `require_ssl` in the parameter group
- Enable `enhancedVpcRouting` for network-level audit trails

### VPC and Networking

- **Always deploy in private subnets** unless external BI tools require public access.
- **Use at least two subnets** in different Availability Zones for the subnet group.
- **Enable enhanced VPC routing** for compliance environments — it forces all
  COPY/UNLOAD traffic through the VPC, making it visible to VPC flow logs.
- **Use managed security groups** (via `securityGroupIds` or `allowedCidrBlocks`)
  rather than wide-open security group rules.

### Multi-AZ Deployment

Multi-AZ provides automatic failover to a standby cluster in a different
Availability Zone. Requirements:
- RA3 node type (not supported on DC2)
- No additional configuration — AWS manages the standby automatically
- RPO: 0 (synchronous replication); RTO: typically under 60 seconds

Enable Multi-AZ for any production cluster where downtime impacts business
operations.

### Audit Logging

Enable logging for production clusters. CloudWatch is recommended over S3 for
modern deployments:

- **CloudWatch** — Real-time log streaming, integrates with CloudWatch Alarms,
  Insights queries, and cross-account log aggregation.
- **S3** — Lower cost for long-term retention, integrates with Athena for
  historical analysis.

Log types:
- `connectionlog` — All authentication attempts (successful and failed)
- `useractivitylog` — All SQL statements executed (before execution)
- `userlog` — User creation, deletion, and privilege changes

### Password Management

Always use `manageMasterPassword: true` for production. This:
- Generates a strong random password
- Stores it in Secrets Manager
- Rotates it automatically
- Eliminates static credentials in configuration files
- Integrates with IAM authentication for additional security layers

## Cost Optimization

### Right-Sizing

| Cluster Size | Recommended Starting Point |
|-------------|---------------------------|
| < 500 GB, development | 1 × `dc2.large` |
| 500 GB - 2 TB, light queries | 2 × `ra3.xlplus` |
| 2 - 16 TB, moderate queries | 2-4 × `ra3.4xlarge` |
| 16+ TB, heavy concurrent queries | 2+ × `ra3.16xlarge` |

### Reserved Instances

Redshift offers 1-year and 3-year reserved instance pricing:
- **1-year no upfront**: ~25% savings
- **1-year partial upfront**: ~35% savings
- **3-year partial upfront**: ~60% savings

Reserved instances make sense for production clusters running 24/7. Use on-demand
for development clusters that can be paused.

### Pause and Resume

Development and staging clusters can be paused when not in use. A paused cluster:
- Stops compute charges (you still pay for storage on DC2; RA3 managed storage
  charges continue at a lower rate)
- Retains all data and configuration
- Can be resumed in minutes

**Note**: Pause/resume is not exposed in the v1 spec because it's an operational
action, not a deployment configuration. Use the AWS Console or CLI for
pause/resume operations.

### Concurrency Scaling

Redshift automatically adds transient capacity to handle query bursts. Each
cluster gets 1 hour of free concurrency scaling credits per day per main cluster
node. Beyond that, charges apply at the on-demand rate.

## Monitoring Considerations

### Key CloudWatch Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `CPUUtilization` | Cluster-wide CPU usage | > 80% sustained |
| `PercentageDiskSpaceUsed` | Local disk usage (DC2) | > 80% |
| `ReadIOPS` / `WriteIOPS` | Disk I/O operations | Baseline + 2 std dev |
| `DatabaseConnections` | Active connections | > 80% of limit (500) |
| `QueryDuration` | P99 query latency | Business SLA dependent |
| `MaintenanceMode` | Whether maintenance is active | Any value > 0 |
| `HealthStatus` | Cluster health (1 = healthy) | < 1 |

### System Tables for Query Performance

Redshift provides system tables for deep performance analysis:

- `stl_query` — Completed queries with execution time
- `stl_wlm_query` — Workload management queue wait times
- `svv_table_info` — Table storage, sort key effectiveness, compression
- `stl_alert_event_log` — Query optimizer alerts (missing stats, nested loops)
- `svl_query_report` — Step-level execution plan details

### Query Performance Tuning

1. **Distribution keys** — Choose keys that co-locate joined tables
2. **Sort keys** — Align with common WHERE clause filters
3. **Compression encoding** — Use `ANALYZE COMPRESSION` to find optimal encodings
4. **VACUUM and ANALYZE** — Run after bulk loads to reclaim space and update stats
5. **Result caching** — Enabled by default; repeated queries return cached results

## Architecture: Cluster Topology

### Single-Node (numberOfNodes = 1)

```
┌─────────────────────────────────┐
│          Single Node            │
│  ┌──────────┐  ┌─────────────┐ │
│  │  Leader   │  │   Compute   │ │
│  │ Functions │  │   Slices    │ │
│  └──────────┘  └─────────────┘ │
│        Combined on one node     │
└─────────────────────────────────┘
```

The leader and compute functions share a single node. Suitable for development
and small datasets.

### Multi-Node (numberOfNodes > 1)

```
                    ┌────────────────┐
SQL Clients ──────▶ │  Leader Node   │
                    │ (query plan,   │
                    │  aggregation)  │
                    └───────┬────────┘
                            │
              ┌─────────────┼─────────────┐
              ▼             ▼             ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐
        │ Compute  │  │ Compute  │  │ Compute  │
        │ Node 1   │  │ Node 2   │  │ Node N   │
        │ (slices) │  │ (slices) │  │ (slices) │
        └──────────┘  └──────────┘  └──────────┘
```

The leader node parses queries, generates execution plans, and coordinates
compute nodes. Compute nodes store data slices and execute parallel query
fragments. The leader aggregates results.

## Architecture: RA3 Managed Storage

```
┌─────────────────────────────────────────┐
│             RA3 Compute Node            │
│  ┌──────────────────────────────────┐   │
│  │    Local NVMe SSD Cache          │   │
│  │    (hot data, recently accessed) │   │
│  └──────────────┬───────────────────┘   │
│                 │ automatic tiering      │
│  ┌──────────────▼───────────────────┐   │
│  │    Redshift Managed Storage (S3) │   │
│  │    (warm/cold data, virtually    │   │
│  │     unlimited capacity)          │   │
│  └──────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

RA3 nodes automatically move frequently accessed data to local NVMe SSDs and
less-frequently accessed data to S3-backed managed storage. This provides
virtually unlimited storage capacity while maintaining SSD performance for
active queries.

## Service Limits

| Limit | Value |
|-------|-------|
| Maximum nodes per cluster | 128 (ra3.16xlarge) |
| Maximum databases per cluster | 60 |
| Maximum schemas per database | 9,900 |
| Maximum tables per cluster | 100,000 (with 100+ nodes) |
| Maximum columns per table | 1,600 |
| Maximum concurrent connections | 500 |
| Maximum concurrent queries (WLM) | 50 |
| Maximum COPY file size | 5 GB (recommended) |
| Snapshot retention | 1-35 days (automated) |
| Maximum IAM roles per cluster | 10 |
| Cluster identifier length | 1-63 characters |

## References

- [Amazon Redshift Documentation](https://docs.aws.amazon.com/redshift/latest/mgmt/welcome.html)
- [Redshift Best Practices](https://docs.aws.amazon.com/redshift/latest/dg/c_designing-tables-best-practices.html)
- [RA3 Node Types](https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-clusters.html#rs-ra3-node-types)
- [Redshift Pricing](https://aws.amazon.com/redshift/pricing/)
- [Database Encryption](https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-db-encryption.html)
- [Audit Logging](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html)
- [Multi-AZ Deployments](https://docs.aws.amazon.com/redshift/latest/mgmt/managing-cluster-multi-az.html)
- [Enhanced VPC Routing](https://docs.aws.amazon.com/redshift/latest/mgmt/enhanced-vpc-routing.html)
- [Workload Management](https://docs.aws.amazon.com/redshift/latest/dg/c_workload_mngmt_classification.html)
- [System Tables Reference](https://docs.aws.amazon.com/redshift/latest/dg/cm_chap_system-tables.html)
