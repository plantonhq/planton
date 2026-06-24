# Zalando Postgres Operator: A Comprehensive Guide

## Introduction: Battle-Tested PostgreSQL at Scale

The Zalando Postgres Operator isn't a lab experiment. It's production infrastructure that has managed hundreds of PostgreSQL clusters at Zalando for years, powering one of Europe's largest e-commerce platforms. When you deploy PostgreSQL using our IaC modules, Zalando operator is working behind the scenes.

This guide explains how it works, how to operate it, and how to leverage its capabilities—from basic deployments to advanced disaster recovery scenarios.

---

## Architecture: The Three Pillars

Understanding Zalando operator requires understanding its three core components, each handling a distinct responsibility.

### 1. The Operator (The Orchestrator)

The operator itself is a Kubernetes controller—a pod running in your cluster that watches for `postgresql` custom resources. When you create or update a PostgreSQL cluster manifest, the operator detects the change and reconciles the actual state to match your desired state.

**What it does**:
- Creates and manages StatefulSets for PostgreSQL pods
- Provisions Services for client connections
- Configures Secrets for credentials
- Manages ConfigMaps for PostgreSQL configuration
- Handles rolling updates, scaling, and upgrades

**What it doesn't do**: The operator doesn't make database-specific decisions like "which replica should I promote?" That's Patroni's job.

### 2. Spilo (The All-in-One Container)

Spilo is Zalando's "fat container" image that bundles everything needed to run a PostgreSQL cluster node:

- **PostgreSQL** (the database itself)
- **Patroni** (high availability orchestration)
- **WAL-G** (backup and restore)
- **PgBouncer** (connection pooling, optional)

Think of Spilo as a self-contained PostgreSQL node that knows how to participate in a clustered environment. When a Spilo container starts, it reads environment variables injected by the operator and configures itself accordingly.

**The startup flow**:
1. Container starts, runs `/launch.sh`
2. Reads environment variables (connection strings, backup settings, etc.)
3. Runs `/scripts/configure_spilo.py` to generate Patroni configuration
4. Starts Patroni, which starts PostgreSQL

### 3. Patroni (The High Availability Brain)

Patroni is the component that makes your PostgreSQL cluster highly available. It's a Python daemon that runs inside each Spilo container alongside PostgreSQL.

**What Patroni does**:
- **Leader Election**: Ensures there's always one (and only one) primary
- **Health Monitoring**: Continuously checks if the primary is healthy
- **Automatic Failover**: Promotes a replica if the primary fails
- **Replication Management**: Configures PostgreSQL replication automatically

Patroni uses the **Kubernetes API as its Distributed Configuration Store (DCS)**. Instead of requiring etcd or Consul, it stores cluster state in Kubernetes ConfigMaps or Endpoints. This is elegant—one less moving part.

---

## How a Cluster Works: A Day in the Life

Let's walk through what happens when you deploy a PostgreSQL cluster with `replicas: 3`.

### Initial Bootstrap

1. **Operator creates a StatefulSet**: Three pods will be created: `postgres-cluster-0`, `postgres-cluster-1`, `postgres-cluster-2`.

2. **Race for Leadership**: All three Spilo containers start simultaneously. Each runs Patroni, which attempts to acquire a lock in the Kubernetes DCS (e.g., a ConfigMap named `postgres-cluster`).

3. **Primary Initialization**: The first Patroni to acquire the lock becomes the **leader** (primary). It initializes a new PostgreSQL database (runs `initdb`) and starts accepting writes.

4. **Replica Configuration**: The other two Patroni instances see that a leader exists. They configure themselves as **replicas**, using `pg_basebackup` to clone the primary's data, and start streaming replication.

5. **Services Created**: The operator creates two Services:
   - `postgres-cluster` → Points to the primary (for writes)
   - `postgres-cluster-repl` → Points to all replicas (for read-only queries)

### Steady State Operations

In normal operation, Patroni on each node continuously:
- **Checks leader status**: "Am I the leader? Is there a leader?"
- **Updates DCS**: Stores its state (lag, timeline, etc.) in the Kubernetes DCS
- **Monitors replication**: Ensures replicas are keeping up

The operator monitors the StatefulSet and Services, ensuring they match the desired configuration.

### Automated Failover

**Scenario**: The node running `postgres-cluster-0` (the primary) crashes.

1. **Detection** (< 10 seconds): The other Patroni instances notice the leader's heartbeat has stopped in the DCS.

2. **Election** (< 5 seconds): Patroni on `postgres-cluster-1` and `postgres-cluster-2` race to acquire the leader lock. The replica with the most up-to-date data (lowest replication lag) wins.

3. **Promotion** (< 5 seconds): The winning replica promotes itself to primary using `pg_ctl promote`.

4. **Reconfiguration** (< 5 seconds): The other replica reconfigures itself to stream from the new primary.

5. **Service Update** (< 2 seconds): The `postgres-cluster` Service automatically points to the new primary (Kubernetes watches the `role: master` label).

**Total downtime**: ~15-30 seconds. Your application retries failed connections and reconnects to the new primary.

---

## Backups and Disaster Recovery: The Complete Story

Backups are where Zalando operator truly shines. It supports **continuous backups with point-in-time recovery (PITR)** using WAL-G.

### How Continuous Backups Work

When backups are enabled, WAL-G runs inside each Spilo container and performs two types of operations:

1. **WAL Archiving (Continuous)**:
   - PostgreSQL is configured with `archive_mode = on`
   - Every completed WAL segment is immediately pushed to object storage (S3, GCS, Azure, Cloudflare R2)
   - This happens constantly as transactions occur
   - **Recovery Point Objective (RPO)**: < 1 minute (you can lose at most one WAL segment's worth of data)

2. **Base Backups (Scheduled)**:
   - A full backup of the PostgreSQL data directory is taken on a schedule (default: daily at 2 AM UTC)
   - WAL-G uses efficient delta compression to minimize storage
   - These base backups serve as restore points

**Storage Structure** in object storage:

```
s3://your-bucket/backups/postgres-cluster/15/
├── basebackups_005/
│   ├── base_000000010000000000000042
│   ├── base_000000010000000000000076
│   └── base_000000010000000000000098
└── wal_005/
    ├── 000000010000000000000042
    ├── 000000010000000000000043
    ├── 000000010000000000000044
    └── ... (continuous WAL segments)
```

### Same-Cluster Cloning

**Use Case**: You want a copy of your production database for staging or testing.

**How it works**: Create a new PostgreSQL cluster manifest with a `clone` section pointing to the source cluster:

```yaml
apiVersion: acid.zalan.do/v1
kind: postgresql
metadata:
  name: postgres-staging
spec:
  clone:
    cluster: "postgres-production"  # Source cluster in same Kubernetes cluster
```

The operator will:
1. Fetch the latest base backup from object storage
2. Replay WAL segments to bring it up to date
3. Start the new cluster as an independent primary

**Limitation**: The source cluster must exist in the same Kubernetes cluster. This is for cloning, not disaster recovery.

### Cross-Cluster Disaster Recovery: The Standby-Then-Promote Pattern

**Use Case**: Your entire Kubernetes cluster is destroyed (region outage, accidental deletion). You need to restore PostgreSQL on a completely new cluster using only the backups in object storage.

This is the **real disaster recovery** scenario. The `clone` mechanism doesn't work here because the source cluster no longer exists.

**The Solution**: Use the operator's **standby cluster** feature.

#### Stage 1: Bootstrap as Standby (Read-Only)

Deploy a new cluster with the `standby` configuration:

```yaml
apiVersion: acid.zalan.do/v1
kind: postgresql
metadata:
  name: postgres-restored
spec:
  numberOfInstances: 1
  postgresql:
    version: "15"
  
  # THIS IS THE KEY: standby configuration
  standby:
    s3_wal_path: "s3://your-bucket/backups/postgres-production/15"
  
  # CRITICAL: Use STANDBY_* prefixed environment variables
  env:
    - name: STANDBY_AWS_ENDPOINT
      value: "https://s3.amazonaws.com"
    - name: STANDBY_AWS_ACCESS_KEY_ID
      valueFrom:
        secretKeyRef:
          name: backup-credentials
          key: access_key_id
    - name: STANDBY_AWS_SECRET_ACCESS_KEY
      valueFrom:
        secretKeyRef:
          name: backup-credentials
          key: secret_access_key
```

**What happens**:
1. Spilo detects the `standby` configuration via `STANDBY_*` environment variables
2. Patroni bootstraps as a "standby leader"
3. WAL-G fetches the latest base backup from object storage
4. PostgreSQL starts in **recovery mode**, continuously replaying WAL segments
5. The cluster is **read-only**—you can query data but not write

**Why the standby stage matters**:
- **Validate data**: Connect and verify your data is correct before committing to failover
- **Run checks**: Test application compatibility, run integrity checks
- **No write risk**: Can't accidentally modify data while validating

#### Stage 2: Promote to Primary (Read-Write)

Once you've validated the data, promote the standby to a full primary:

**Remove the `standby` block** from your manifest:

```yaml
apiVersion: acid.zalan.do/v1
kind: postgresql
metadata:
  name: postgres-restored
spec:
  numberOfInstances: 1
  postgresql:
    version: "15"
  
  # standby: REMOVED - this triggers promotion
  
  env:
    # env vars can remain (optional)
```

Apply the updated manifest. The operator detects the change:

1. **No pod restart needed**: The Spilo container reconfigures itself
2. Patroni removes the standby configuration
3. PostgreSQL exits recovery mode (promoted to primary)
4. The database is now **read-write** and fully operational

**Timeline advanced**: PostgreSQL's timeline ID increments (e.g., 1 → 2), signaling a promotion event.

#### Point-in-Time Recovery (PITR)

To restore to a specific timestamp instead of the latest data:

```yaml
standby:
  s3_wal_path: "s3://your-bucket/backups/postgres-production/15"
  standby_mode: "on"
  recovery_target_time: "2024-11-07 10:30:00 UTC"
```

PostgreSQL will replay WAL segments up to that exact timestamp and stop. Perfect for recovering from accidental data deletion.

---

## Connection Pooling: Integrated PgBouncer

PostgreSQL connections are heavyweight—each connection spawns a backend process consuming memory. For web applications with bursty traffic, connection pooling is essential.

### Why You Need Pooling

Without pooling, an application with 100 web servers, each maintaining 20 database connections, creates **2,000 PostgreSQL backend processes**. This is inefficient and can exhaust PostgreSQL's `max_connections` limit.

With pooling, those 2,000 application connections are multiplexed over a much smaller pool of ~50 actual PostgreSQL connections.

### Zalando's Integrated PgBouncer

Zalando operator can deploy and manage PgBouncer automatically:

```yaml
spec:
  enableConnectionPooler: true
  connectionPooler:
    numberOfInstances: 2
    mode: transaction  # or session
    schema: pooler
    user: pooler
```

The operator will:
1. Deploy a separate PgBouncer Deployment (2 replicas for HA)
2. Configure PgBouncer with your database credentials
3. Create a Service (`postgres-cluster-pooler`) for pooled connections
4. Automatically update PgBouncer config when you scale or modify the database

**Application changes**: Point your app to the pooler service instead of the direct database service:

```
# Without pooling:
postgres-cluster.namespace.svc.cluster.local:5432

# With pooling:
postgres-cluster-pooler.namespace.svc.cluster.local:5432
```

### Pooling Modes

- **Session Mode**: One PostgreSQL connection per PgBouncer client session. Safe for all queries, including prepared statements. Less efficient pooling.
- **Transaction Mode** (default): PostgreSQL connections are released after each transaction. More efficient, but **breaks prepared statements**. Most ORMs are compatible.

**Gotcha**: If your application uses prepared statements extensively, you'll need `session` mode or to disable prepared statements in your ORM configuration.

---

## Upgrades and Scaling: Operational Tasks

### Scaling Replicas

**Horizontal scaling** (adding more replicas) is a simple manifest change:

```yaml
spec:
  numberOfInstances: 3  # was 2
```

Apply the change. The operator will:
1. Update the StatefulSet to `replicas: 3`
2. Kubernetes creates `postgres-cluster-2`
3. Patroni on the new pod automatically configures itself as a replica
4. The replica clones data from the primary and starts streaming replication

**Downtime**: Zero. The primary and existing replicas continue serving traffic.

### Vertical Scaling

**Vertical scaling** (changing CPU/memory) requires pod restarts:

```yaml
spec:
  resources:
    requests:
      cpu: 500m      # was 250m
      memory: 1Gi    # was 512Mi
    limits:
      cpu: 2000m
      memory: 4Gi
```

The operator performs a rolling restart:
1. Replica pods restart one at a time with new resources
2. Finally, the primary fails over to a replica (brief downtime: ~15-30 seconds)
3. The old primary restarts and joins as a replica

### Major Version Upgrades

Upgrading from PostgreSQL 14 to 15 is automated:

```yaml
spec:
  postgresql:
    version: "15"  # was "14"
```

The operator's "fast in-place upgrade" process:
1. Backs up the current data to object storage
2. Performs a rolling restart with the new version
3. Runs `pg_upgrade` inside each pod
4. Validates the upgrade succeeded

**Downtime**: Usually < 2 minutes for small databases, scales with size.

**Recommendation**: Test upgrades in a staging environment first. Always ensure backups are recent.

---

## Monitoring and Troubleshooting

### Health Checks

Check cluster health with `kubectl`:

```bash
# Get cluster status
kubectl get postgresql postgres-cluster

# Check pod status
kubectl get pods -l application=spilo,cluster-name=postgres-cluster

# Check Patroni cluster state (from inside a pod)
kubectl exec postgres-cluster-0 -- patronictl list
```

### Logs

**Operator logs** (for provisioning issues):
```bash
kubectl logs -n postgres-operator -l name=postgres-operator
```

**Patroni logs** (for failover/replication issues):
```bash
kubectl logs postgres-cluster-0 | grep -i patroni
```

**PostgreSQL logs** (for query issues):
```bash
kubectl logs postgres-cluster-0 | grep -i postgres
```

### Common Issues

**Pod stuck in `CrashLoopBackOff`**:
- Check logs: `kubectl logs postgres-cluster-0`
- Common causes: Corrupted data, insufficient resources, misconfigured secrets

**Replication lag**:
- Check Patroni status: `kubectl exec postgres-cluster-0 -- patronictl list`
- Look for "Lag in MB" column
- Causes: Network issues, insufficient replica resources, large transactions on primary

**Failover not happening**:
- Patroni may be unable to reach the Kubernetes API (DCS)
- Check network policies
- Verify ServiceAccount permissions

**Backup failures**:
- Check environment variables: `kubectl exec postgres-cluster-0 -- env | grep -E "WALG|AWS"`
- Verify object storage credentials and connectivity
- Check WAL-G logs in pod

---

## How This Maps to Our API

Our `PostgresKubernetes` API (defined in `api.proto`) abstracts Zalando operator's complexity into a clean, declarative interface.

### API Structure

```protobuf
message PostgresKubernetesSpec {
  // Container resources (CPU, memory, replicas, disk)
  PostgresKubernetesContainer container = 1;
  
  // External access configuration
  PostgresKubernetesIngress ingress = 2;
  
  // Backup and disaster recovery settings
  PostgresKubernetesBackupConfig backup_config = 3;
}
```

### The Translation Layer

Our IaC modules (Pulumi/Terraform) translate your high-level API to Zalando's `postgresql` CRD:

| API Field | Zalando CRD Field | What It Does |
|-----------|-------------------|--------------|
| `container.replicas` | `spec.numberOfInstances` | Number of PostgreSQL pods |
| `container.resources` | `spec.resources` | CPU/memory allocation |
| `container.disk_size` | `spec.volume.size` | PVC size per pod |
| `backup_config.bucket` + `backup_config.object_prefix` | `env: WALG_S3_PREFIX` | Backup storage location; the module composes `s3://<bucket>/<object_prefix>/$(SCOPE)/$(PGVERSION)` |
| `backup_config.schedule` | `env: BACKUP_SCHEDULE` | Cron schedule for base backups |
| `backup_config.restore` | `spec.standby.s3_wal_path` | Disaster recovery restore source |
| `ingress.enabled` | `spec.service.type: LoadBalancer` | External access |

### Disaster Recovery in the API

When you set `backup_config.restore.enabled` (with its `bucket`, `object_prefix`, and `credentials`), the IaC module:

1. Adds `spec.standby` to the generated manifest, pointing at `s3://<bucket>/<object_prefix>`
2. Injects `STANDBY_*` environment variables
3. Creates the cluster as a read-only standby

To promote to primary, you set `backup_config.restore.enabled` to `false`. The IaC regenerates the manifest without `spec.standby`, triggering promotion.

This abstraction makes disaster recovery as simple as:

```yaml
# Stage 1: Restore as standby
backup_config:
  restore:
    enabled: true
    bucket:
      value: prod-db-backups
    object_prefix: production
    credentials:
      cloudflare_account_id: a1b2c3d4
      access_key_id: r2-access-key-id
      secret_access_key: $secret/postgres-r2-secret-access-key

# Stage 2: Promote (disable restore)
backup_config:
  restore:
    enabled: false
```

---

## Best Practices

### 1. Always Enable Backups

Backups should be the default, not an afterthought. Configure object storage credentials at the operator level so all databases inherit backup settings:

```yaml
# Operator-level configuration
backup_config:
  use_walg_backup: true
  backup_schedule: "0 2 * * *"  # Daily at 2 AM UTC
```

### 2. Test Your Disaster Recovery

Don't wait for a real disaster to discover your backups don't work. Quarterly DR drills:

1. Deploy a standby cluster from production backups
2. Verify data integrity
3. Promote to primary and test write operations
4. Document the recovery time

### 3. Use Connection Pooling

Unless you have a very small number of clients, enable PgBouncer. The overhead is minimal, and the benefits are significant.

### 4. Monitor Replication Lag

Set up alerts for replication lag > 1 GB or > 5 minutes. High lag means your replicas are falling behind, which increases failover recovery time.

### 5. Version Pinning

Pin your PostgreSQL version explicitly in your manifests. Don't use `latest`. Upgrades should be deliberate, tested changes.

### 6. Resource Right-Sizing

Start with conservative resource requests and adjust based on actual usage. PostgreSQL is memory-hungry—allocate enough RAM for caching (`shared_buffers` + OS cache).

### 7. Use Separate Namespaces

Don't run production and staging databases in the same namespace. Isolate them for security and operational clarity.

---

## Advanced: Customizing PostgreSQL Configuration

Zalando operator allows custom `postgresql.conf` settings via the `postgres_parameters` field:

```yaml
spec:
  postgresql:
    parameters:
      max_connections: "200"
      shared_buffers: "2GB"
      effective_cache_size: "6GB"
      work_mem: "16MB"
      maintenance_work_mem: "512MB"
```

**Be careful**: Some settings (like `max_connections`) require pod restarts and can trigger failovers. Test in staging first.

---

## Conclusion: Production-Grade PostgreSQL, Automated

Zalando Postgres Operator transforms PostgreSQL on Kubernetes from "possible but painful" to "automated and reliable." By encoding expert SRE knowledge into a reconciliation loop, it handles the hard parts:

- **High Availability**: Automatic failover in seconds
- **Disaster Recovery**: Cross-cluster restoration from object storage
- **Backups**: Continuous WAL archiving with PITR
- **Scaling**: Horizontal and vertical scaling with minimal downtime
- **Upgrades**: Automated major version upgrades
- **Connection Pooling**: Integrated PgBouncer management

When you deploy PostgreSQL using our IaC modules, you're leveraging battle-tested infrastructure that has proven itself at scale in one of Europe's largest e-commerce platforms.

Focus on your applications. Let Zalando operator handle your databases.

---

## Further Reading

- [Zalando Operator Official Documentation](https://opensource.zalando.com/postgres-operator/)
- [Patroni Documentation](https://patroni.readthedocs.io/)
- [WAL-G Documentation](https://wal-g.readthedocs.io/)
- [PostgreSQL High Availability Best Practices](https://www.postgresql.org/docs/current/high-availability.html)
- [Disaster Recovery Testing Guide](https://www.crunchydata.com/blog/postgres-disaster-recovery) - Applicable concepts

