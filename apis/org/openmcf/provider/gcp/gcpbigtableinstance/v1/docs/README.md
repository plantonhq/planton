# GcpBigtableInstance — Research & Design Documentation

Comprehensive research document covering Cloud Bigtable deployment, architecture, data model, replication, scaling, CMEK encryption, storage types, best practices, and design decisions.

---

## 1. Cloud Bigtable Deployment Landscape

### 1.1 Comparison with Other GCP NoSQL Services

| Aspect | Cloud Bigtable | Cloud Firestore | Cloud Spanner |
|--------|----------------|-----------------|---------------|
| **Data model** | Wide-column (row key + column families) | Document (collections/documents) | Relational (SQL) |
| **Consistency** | Strong per-row, eventual across replication | Strong | Strong (distributed) |
| **Scale** | Horizontal (add clusters/nodes) | Automatic | Horizontal (sharding) |
| **Use case** | Time-series, IoT, analytics, ML features | Mobile/web apps, real-time sync | Global OLTP, financial ledgers |
| **Latency** | Sub-10ms single-row reads | Sub-10ms document reads | Single-digit ms |
| **Storage** | Colossus (column-oriented) | Colossus (document-oriented) | Distributed |
| **Pricing** | Per node-hour + storage | Per operation + storage | Per processing unit + storage |
| **Max row size** | 256 MB (recommended < 10 MB) | 1 MB per document | 10 MB per row |

### 1.2 When to Choose Cloud Bigtable

- High-throughput, low-latency reads and writes at scale (millions of QPS)
- Time-series data: metrics, IoT sensor data, financial tick data
- Analytics workloads: user behavior, click streams, recommendation engines
- Machine learning feature stores requiring fast key-value lookups
- Data volumes exceeding terabytes with consistent sub-10ms latency requirements
- Workloads that benefit from HBase-compatible API (migration from Apache HBase)

### 1.3 When to Choose Firestore

- Mobile and web applications with real-time synchronization
- Document-oriented data with nested structures
- Moderate scale (hundreds of thousands of operations per second)
- Offline support for mobile clients

### 1.4 When to Choose Spanner

- Global distribution with strong consistency
- Relational data model requiring SQL and transactions
- Financial systems requiring serializable isolation
- Multi-region writes with automatic sharding

### 1.5 When to Choose Datastore / Memorystore

- **Datastore**: Legacy document store (Firestore in Datastore mode); good for App Engine workloads
- **Memorystore (Redis/Memcached)**: In-memory caching layer; sub-millisecond latency for hot data

---

## 2. Architecture

### 2.1 Instance, Cluster, and Node Model

```
Bigtable Instance (logical container)
├── Cluster 1 (zone: us-central1-a)
│   ├── Node 1
│   ├── Node 2
│   └── Node 3
├── Cluster 2 (zone: us-east1-b)
│   ├── Node 1
│   └── Node 2
└── Cluster 3 (zone: europe-west1-b)
    ├── Node 1
    └── Node 2
```

- **Instance** — Logical container for tables and data. Bigtable client libraries connect to an instance by project ID and instance name.
- **Cluster** — Physical deployment in a specific zone. Each cluster is an independent replica of the instance's data. Clusters handle reads and writes independently.
- **Node** — Compute unit within a cluster. Each node handles a portion of the cluster's read/write traffic. More nodes = higher throughput.

### 2.2 Storage Engine: Colossus

Bigtable stores data on Google's Colossus distributed file system:

- **Compute-storage separation** — Nodes handle query processing; Colossus handles durable storage
- **SSTable format** — Data is stored in sorted string tables (SSTables), similar to HBase/LevelDB
- **Automatic compaction** — Background process merges and compacts SSTables for optimal read performance
- **No storage provisioning** — Storage scales automatically with data volume

### 2.3 Data Model

Bigtable uses a wide-column data model:

- **Row key** — Unique identifier for each row (up to 4 KB). Row key design is critical for performance.
- **Column families** — Logical grouping of columns. Defined at table creation time. Each family has its own garbage collection policy.
- **Column qualifiers** — Columns within a family. Can be created dynamically (no schema required).
- **Cells** — Intersection of row, column family, and column qualifier. Each cell can store multiple versions with timestamps.
- **Timestamps** — Each cell version is identified by a timestamp. Used for garbage collection and time-range queries.

### 2.4 Connection Architecture

Applications connect to Bigtable via:

1. **Bigtable client libraries** — Official SDKs for Java, Go, Python, Node.js, C++, C#. Handle connection pooling, retries, and load balancing.
2. **HBase compatibility layer** — Drop-in replacement for Apache HBase clients via the HBase Bigtable client.
3. **REST/gRPC API** — Direct API access for custom integrations.

Client libraries connect to the Bigtable Data API endpoint (`bigtable.googleapis.com`) and are routed to the appropriate cluster based on app profiles.

---

## 3. Replication

### 3.1 Multi-Cluster Replication

When an instance has multiple clusters, Bigtable automatically replicates data across all clusters:

- **Eventually consistent** — Replication is asynchronous. Writes to one cluster are eventually visible in other clusters (typically within seconds).
- **Single-cluster routing** — App profiles can route all traffic to a specific cluster for strong consistency guarantees.
- **Multi-cluster routing** — Default behavior. Client libraries route requests to the nearest available cluster for lowest latency.
- **Automatic failover** — If a cluster becomes unavailable, clients automatically route to the next available cluster.

### 3.2 App Profiles and Routing

App profiles control how Bigtable routes traffic:

| Routing Policy | Consistency | Use Case |
|----------------|-------------|----------|
| Multi-cluster | Eventual | Lowest latency, automatic failover |
| Single-cluster | Strong | Read-your-writes, transactions |

App profiles are not included in this component (they are application-level configuration managed separately).

### 3.3 Replication Lag

- Typical replication lag: **seconds** (under normal conditions)
- Lag can increase during cluster scaling, compaction, or high write rates
- Monitor via `bigtable.googleapis.com/server/latencies` and replication delay metrics

### 3.4 Cross-Region Replication

Clusters can span multiple regions for disaster recovery:

```
Instance: global-bigtable
├── Cluster: us-central1-a    (Americas)
├── Cluster: europe-west1-b   (EMEA)
└── Cluster: asia-east1-a     (APAC)
```

Cross-region replication provides geographic redundancy but increases replication lag compared to same-region multi-zone deployments.

---

## 4. Scaling

### 4.1 Scaling Modes

| Mode | Configuration | Behavior |
|------|---------------|----------|
| Fixed | `num_nodes: N` | Static N nodes; manual scaling |
| Autoscaling | `autoscaling_config` | Dynamic scaling based on CPU and storage targets |
| Automatic | Neither set | Bigtable auto-allocates based on data footprint |

### 4.2 Fixed Node Count

- Set `numNodes` to the desired number of nodes per cluster
- Each node provides approximately:
  - **SSD**: ~10,000 reads/s or ~10,000 writes/s at 1 KB row size
  - **HDD**: ~500 reads/s or ~10,000 writes/s at 1 KB row size
- Manual scaling: update `numNodes` and apply the change

### 4.3 Autoscaling

Autoscaling dynamically adjusts node count based on utilization:

| Parameter | Description | Range |
|-----------|-------------|-------|
| `minNodes` | Minimum node count | >= 1 |
| `maxNodes` | Maximum node count | >= `minNodes` |
| `cpuTarget` | Target CPU utilization (%) | 10–80 |
| `storageTarget` | Target storage per node (GB) | SSD: 2560–5120; HDD: 8192–16384 |

Autoscaling evaluates every few minutes:

- **Scale up** — When average CPU exceeds `cpuTarget` or per-node storage exceeds `storageTarget`
- **Scale down** — When utilization drops sufficiently below targets (with cooldown)

### 4.4 Node Scaling Factor

| Factor | Behavior | Use Case |
|--------|----------|----------|
| `NodeScalingFactor1X` | Scale in increments of 1 node | Default; most workloads |
| `NodeScalingFactor2X` | Scale in increments of 2 nodes | Large workloads needing coarser scaling steps |

When using 2X, `numNodes`, `minNodes`, and `maxNodes` must all be even numbers. Node scaling factor is immutable after creation.

### 4.5 Capacity Planning Guidelines

| Metric | Recommendation |
|--------|---------------|
| CPU utilization | Keep below 70% for headroom |
| Storage per node (SSD) | Keep below 2.5 TB; add nodes if approaching limit |
| Storage per node (HDD) | Keep below 8 TB per node |
| Read latency (p99) | Target < 10ms; investigate if > 50ms |
| Write latency (p99) | Target < 10ms for SSD; higher for HDD |

---

## 5. Storage Types

### 5.1 SSD vs. HDD

| Aspect | SSD | HDD |
|--------|-----|-----|
| **Read latency** | Sub-10ms | Higher (tens of ms) |
| **Write latency** | Sub-10ms | Similar to SSD |
| **Cost** | ~$0.17/GB/month | ~$0.026/GB/month |
| **Throughput** | Higher reads/s per node | Lower reads/s per node |
| **Use case** | Real-time serving, latency-sensitive | Batch analytics, cold storage |
| **Storage per node** | Up to 5 TB | Up to 16 TB |

### 5.2 Storage Type Selection

- **Default to SSD** for most workloads (real-time serving, API backends, streaming)
- **Choose HDD** only for batch analytics workloads where:
  - Read latency is not critical (e.g., batch MapReduce/Dataflow jobs)
  - Data volume is very large and cost is a primary concern
  - Writes are the dominant operation (HDD write latency is comparable to SSD)

Storage type is immutable after cluster creation. Changing it requires deleting and recreating the cluster (and instance).

---

## 6. CMEK Encryption

### 6.1 Customer-Managed Encryption Keys

Bigtable supports CMEK at the cluster level:

| Level | Field | Purpose |
|-------|-------|---------|
| Cluster | `kms_key_name` | Encrypt all data stored in the cluster at rest |

Unlike AlloyDB (which has separate keys for cluster, backup, and PITR), Bigtable uses a single KMS key per cluster that covers all data in that cluster including backups.

### 6.2 Requirements

- KMS key must exist in the same region as the cluster's zone
- The Cloud Bigtable service account (`service-{PROJECT_NUMBER}@gcp-sa-bigtable.iam.gserviceaccount.com`) must have `cloudkms.cryptoKeyEncrypterDecrypter` on the key
- All clusters in an instance should use the same CMEK key (or keys in their respective regions) for consistent encryption
- CMEK is immutable after creation

### 6.3 Key Rotation

Cloud KMS supports automatic key rotation. When a key is rotated:

- New data is encrypted with the new key version
- Existing data remains encrypted with the original key version until re-encryption (handled automatically by Bigtable during compaction)

---

## 7. Best Practices for Production Deployments

1. **Use multi-cluster replication** — At least 2 clusters in different zones for automatic failover
2. **Enable deletion protection** — `deletionProtection: true` (default); keep it for production
3. **Use autoscaling** — Set `cpuTarget` to 60–70% for most workloads; avoids manual capacity management
4. **Choose SSD** — Default to SSD unless batch analytics is the primary use case
5. **CMEK for compliance** — Use `kmsKeyName` when HIPAA, PCI-DSS, or FedRAMP requires customer-managed keys
6. **Design row keys carefully** — Avoid monotonically increasing keys (e.g., timestamps alone) which cause hotspotting
7. **Monitor key metrics** — CPU utilization, storage utilization, read/write latencies, error rates
8. **Use app profiles** — Configure routing policies to match your consistency and latency requirements
9. **Set appropriate autoscaling bounds** — `minNodes` for cost floor, `maxNodes` for budget ceiling
10. **Plan for immutability** — Zone, storage type, KMS key, and node scaling factor cannot be changed after creation

---

## 8. 80/20 Analysis

### 8.1 What We Cover

| Feature | Included | Rationale |
|---------|----------|-----------|
| Instance + clusters | Yes | Core lifecycle; no instance without clusters |
| Multi-cluster replication | Yes | Up to 8 clusters for HA and geo-distribution |
| Fixed node count | Yes | Simple capacity management |
| Autoscaling (CPU/storage) | Yes | Dynamic scaling for variable workloads |
| SSD and HDD storage types | Yes | Cover both real-time and batch use cases |
| CMEK per cluster | Yes | Enterprise compliance |
| Node scaling factor (1X/2X) | Yes | Fine-grained scaling control |
| Deletion protection | Yes | Safety for production instances |
| Force destroy | Yes | Clean teardown when backups exist |
| Display name | Yes | Human-readable identification |
| GCP labels | Yes | Resource organization and cost allocation |

### 8.2 What We Exclude

| Feature | Excluded | Rationale |
|---------|----------|-----------|
| Tables and column families | Yes | Application-level schema; managed via client libraries or separate tooling |
| App profiles | Yes | Application-level routing configuration; independent lifecycle |
| IAM bindings | Yes | Managed via GCP IAM, not database config |
| Garbage collection policies | Yes | Table-level configuration, not instance-level |
| Bigtable backups | Yes | Table-level operation; managed separately |
| Instance type (DEV/PROD) | Yes | GCP is deprecating the distinction; all instances are effectively PRODUCTION |
| Monitoring and alerts | Yes | Managed via Cloud Monitoring, not instance config |

### 8.3 Deliberate Design Choices

**Bundled clusters:** An instance without clusters cannot store or serve data. Bundling ensures a usable deployment in one component. Tables and app profiles are excluded because they have independent lifecycles and are application-level concerns.

**Per-cluster CMEK:** Each cluster supports its own KMS key. This matches GCP's API and allows compliance scenarios where clusters in different regions use region-local KMS keys.

**No instance type field:** The GCP distinction between DEVELOPMENT and PRODUCTION instance types is deprecated and being removed. All instances are effectively PRODUCTION. For development workloads, use a single small cluster.

**No table management:** Tables, column families, and garbage collection policies are application-level concerns. They change frequently as applications evolve and are better managed via Bigtable client libraries, `cbt` CLI, or dedicated tooling.

---

## 9. Immutable Fields

The following fields cannot be changed after creation; changing them requires recreating the resource:

| Field | Scope | Notes |
|-------|-------|-------|
| `instance_name` | Instance | GCP resource identifier |
| `zone` | Cluster | Zone placement |
| `storage_type` | Cluster | SSD or HDD |
| `kms_key_name` | Cluster | CMEK encryption key |
| `node_scaling_factor` | Cluster | 1X or 2X scaling increments |

---

## 10. Row Key Design (Application Context)

While row key design is an application-level concern (not managed by this component), it significantly impacts Bigtable performance:

### 10.1 Good Row Key Patterns

- **Reverse domain** — `com.example.user#12345` distributes load across nodes
- **Salted timestamp** — `shard_id#timestamp` prevents hotspotting
- **Composite keys** — `tenant_id#entity_type#entity_id` for multi-tenant access patterns

### 10.2 Anti-Patterns

- **Sequential timestamps** — Causes write hotspots on a single node
- **Monotonically increasing IDs** — Same issue as timestamps
- **Hashed keys only** — Loses range-scan capability

This context is relevant because node count and autoscaling targets should account for the expected key distribution and access patterns.

---

## 11. Pricing Model

### 11.1 Cost Components

| Component | Unit | Approximate Cost |
|-----------|------|------------------|
| Nodes (SSD) | Per node-hour | ~$0.65/hr |
| Nodes (HDD) | Per node-hour | ~$0.325/hr |
| SSD storage | Per GB/month | ~$0.17 |
| HDD storage | Per GB/month | ~$0.026 |
| Network egress | Per GB (cross-region) | Standard GCP rates |
| Replication | No additional charge | Included (cross-cluster) |

### 11.2 Cost Optimization

- Use autoscaling to scale down during low-traffic periods
- Choose HDD for cold/batch data to reduce storage costs by ~85%
- Minimize cross-region replication if not needed (reduces network costs)
- Use single-cluster for development/testing workloads

---

## 12. Monitoring and Observability

### 12.1 Key Metrics

| Metric | Source | Alert Threshold |
|--------|--------|-----------------|
| CPU utilization | Cloud Monitoring | > 70% sustained |
| Storage utilization per node | Cloud Monitoring | > 70% of limit |
| Read latency (p99) | Cloud Monitoring | > 50ms |
| Write latency (p99) | Cloud Monitoring | > 50ms |
| Error rate | Cloud Monitoring | > 1% |
| Replication lag | Cloud Monitoring | > 60s |

### 12.2 Dashboards

GCP provides built-in Bigtable monitoring dashboards in Cloud Console. For custom dashboards, use Cloud Monitoring with the `bigtable.googleapis.com` metric prefix.

---

## 13. Infra Chart Composition Patterns

### Pattern 1: Single Cluster (Dev)

```
GcpBigtableInstance (single cluster, auto-allocate nodes)
```

### Pattern 2: Multi-Cluster HA (Production)

```
GcpBigtableInstance (3 clusters, autoscaling)
├── Cluster: us-central1-a
├── Cluster: us-east1-b
└── Cluster: europe-west1-b
```

### Pattern 3: CMEK Encrypted (Enterprise)

```
GcpKmsKeyRing (per region)
├── GcpKmsKey (us-central1)
├── GcpKmsKey (us-east1)
└── GcpKmsKey (europe-west1)
GcpBigtableInstance (clusters reference KMS keys)
├── Cluster: us-central1-a (kmsKeyName → us-central1 key)
├── Cluster: us-east1-b (kmsKeyName → us-east1 key)
└── Cluster: europe-west1-b (kmsKeyName → europe-west1 key)
```

### Pattern 4: Shared Instance (Multi-Tenant)

```
GcpBigtableInstance "shared"
├── Cluster: us-central1-a
└── Cluster: us-east1-b
(Tables per tenant managed at application level)
```

---

## 14. Migration from Apache HBase

Cloud Bigtable provides an HBase-compatible API, making migration straightforward:

1. **Client library swap** — Replace HBase client with Bigtable HBase client (same API surface)
2. **Schema compatibility** — Bigtable supports HBase column families, qualifiers, and timestamps natively
3. **Data migration** — Use Cloud Dataflow or HBase snapshots to migrate data
4. **Performance tuning** — Bigtable may need different node counts than HBase due to different storage architecture

The HBase compatibility layer is not a separate component; it is built into the Bigtable Data API.

---

## 15. Integration with GCP Data Services

| Service | Integration | Use Case |
|---------|-------------|----------|
| Cloud Dataflow | Native Bigtable I/O connector | ETL pipelines, streaming ingestion |
| BigQuery | External data source (federated queries) | Analytics on Bigtable data |
| Cloud Dataproc | HBase connector | Spark/Hadoop jobs on Bigtable |
| Cloud Functions | Client library | Event-driven reads/writes |
| Pub/Sub | Via Dataflow | Streaming ingestion pipeline |

---

## 16. References

- [Cloud Bigtable Documentation](https://cloud.google.com/bigtable/docs)
- [Cloud Bigtable REST API](https://cloud.google.com/bigtable/docs/reference/admin/rest)
- [Bigtable Pricing](https://cloud.google.com/bigtable/pricing)
- [Schema Design Best Practices](https://cloud.google.com/bigtable/docs/schema-design)
- [Key Visualizer](https://cloud.google.com/bigtable/docs/keyvis-overview)
- [Bigtable Monitoring](https://cloud.google.com/bigtable/docs/monitoring-instance)
