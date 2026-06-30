# MySQL on Kubernetes: From StatefulSet Chaos to Production-Ready Operators

## Introduction

For years, the conventional wisdom was clear: **"Don't run databases on Kubernetes."** The reasoning seemed sound—Kubernetes was designed for stateless workloads, and databases are the ultimate stateful application. Yet today, running production MySQL on Kubernetes isn't just viable; it's becoming the preferred approach for organizations embracing cloud-native infrastructure.

What changed? The answer lies in understanding the difference between **Day-1 operations** (initial deployment) and **Day-2 operations** (ongoing lifecycle management). Early attempts at running MySQL on Kubernetes solved Day-1 with StatefulSets but left teams struggling with Day-2 challenges: automated failover, storage expansion, backup orchestration, and version upgrades. The **Kubernetes Operator pattern** changed everything by encoding these complex operational tasks into software.

This document explores the evolution of MySQL deployment methods on Kubernetes, from the dangerous anti-patterns that gave databases on Kubernetes a bad reputation, through the intermediate solutions that solved some problems but not all, to the production-ready operator-based approaches that power MySQL workloads at scale today.

## The MySQL on Kubernetes Maturity Spectrum

Understanding deployment methods as a progression from basic to production-ready helps clarify why operators represent the only viable path forward for serious MySQL workloads.

### Level 0: The Anti-Pattern (StatefulSets Without Operators)

**What it is:** Deploying MySQL using raw StatefulSets—Kubernetes' primitive for stateful workloads—which provide stable pod identities and persistent storage but no database-aware logic.

**The promise:** StatefulSets offer stable network identities (`mysql-0`, `mysql-1`, `mysql-2`) and guarantee that each pod gets its own persistent volume. On paper, this sounds perfect for a database.

**The reality:** The official Kubernetes documentation for running replicated MySQL explicitly warns: **"This is not a production configuration."** Here's why:

- **No automated failover:** If the primary pod (`mysql-0`) crashes, the StatefulSet controller will restart it, but it has no concept of promoting `mysql-1` to become the new primary. Your database is down until the restart completes.

- **Storage expansion is a nightmare:** Expanding storage requires a manual, multi-step process that's error-prone under pressure: delete the StatefulSet while orphaning pods (`--cascade=orphan`), manually edit each PVC to request more space, update the StatefulSet YAML, recreate it, and hope the CSI driver cooperates. One mistake means data loss.

- **Initialization fragility:** Distinguishing between "first boot on empty storage" (run initialization scripts) and "restart with existing data" (skip initialization) is complex logic that requires custom scripts. Getting this wrong destroys data.

- **Split-brain risk:** In a replicated topology, a network partition can create multiple pods that each believe they're the primary, causing datasets to diverge. StatefulSets have no quorum logic to prevent this catastrophic scenario.

**Verdict:** StatefulSets are building blocks, not solutions. They're analogous to providing someone a hammer and nails and calling it a house. For production databases, this approach is unacceptably risky.

### Level 1: The Illusion (Helm Charts Without Operators)

**What it is:** Using Helm charts like Bitnami MySQL to deploy MySQL on Kubernetes, which templates the YAML for StatefulSets, Services, and ConfigMaps.

**What it solves:** Helm excels at packaging complex Kubernetes manifests and managing the initial installation (Day-1). A single `helm install` command can deploy a complete MySQL setup with sensible defaults.

**What it doesn't solve:** Helm is a **package manager**, not a runtime controller. After `helm install` completes, Helm's active involvement ends. If the primary database pod fails tomorrow, Helm will not orchestrate failover. If you need to resize storage next month, Helm will not help. If automated backups fail silently, Helm won't notice or alert you.

**The modern pattern:** Today's best practice is to use Helm to **install the operator** itself, then use the operator's custom resources to **manage the database lifecycle**. This combines Helm's templating strengths with the operator's continuous reconciliation loop.

**Verdict:** Helm charts that deploy StatefulSets are simply automating the deployment of the anti-pattern. They're useful for development environments but inadequate for production Day-2 needs.

### Level 2: The Solution (Kubernetes Operators)

**What it is:** A Kubernetes Operator is an application-specific controller that runs inside your cluster and continuously manages the complete lifecycle of a complex application. Operators extend the Kubernetes API with **Custom Resource Definitions (CRDs)**, allowing you to declare your intent in simple YAML: "I want a 3-node, production-grade MySQL cluster with automated daily backups."

**How it works:** The operator runs a continuous **reconciliation loop**:
1. Read the desired state from your CRD (e.g., `PerconaXtraDBCluster`)
2. Observe the actual state of the cluster (pod health, replication status, backup completion)
3. Take automated actions to make actual state match desired state
4. Repeat continuously

**What it solves:** Everything that StatefulSets and Helm don't:
- **Automated failover:** The operator detects when a database node fails and orchestrates the promotion of a new primary, updating routes and replication configurations automatically.
- **Backup orchestration:** Schedule automated backups to S3, GCS, or Azure Blob Storage. Restore with a single CRD object creation.
- **Online storage expansion:** Change a single field in the CRD, and the operator performs the complex dance of expanding volumes without downtime.
- **Database-aware upgrades:** The operator understands how to perform rolling upgrades of database nodes while maintaining quorum and minimizing risk.
- **Monitoring integration:** Automatically deploy and configure monitoring sidecars and exporters.

**Verdict:** For production MySQL on Kubernetes, operators are not optional—they're the only architecture that reliably solves Day-2 operational complexity.

## Production-Ready MySQL Operators: A Comparative Analysis

Not all operators are created equal. Licensing, replication architecture, and production maturity vary significantly. This section compares the major options.

### The Operator Landscape

| Operator | License (Operator) | License (Database) | Replication | Production Ready? | Key Differentiator |
|----------|-------------------|-------------------|------------|-------------------|-------------------|
| **Percona PXC** | Apache 2.0 | GPLv2 | **Galera (Synchronous)** | **✅ Yes (Recommended)** | Zero data loss with synchronous replication |
| **Percona PS** | Apache 2.0 | GPLv2 | Async, Group Rep | ❌ No (Tech Preview) | Not recommended for production |
| **Oracle MySQL** | GPLv2 | GPLv2 / Commercial | Group Replication | ⚠️ Yes (High Risk) | Container licensing trap (see below) |
| **Vitess** | Apache 2.0 | GPLv2 | Async (Sharded) | ✅ Yes (Hyperscale) | Horizontal sharding for YouTube-scale workloads |
| **Moco** | Apache 2.0 | GPLv2 | Semi-synchronous | ✅ Yes | GTID-based replication from Cybozu |
| **Bitpoke (Presslabs)** | Apache 2.0 | GPLv2 | Async | ✅ Yes | Orchestrator-managed failover |
| **KubeDB** | Apache 2.0 | GPLv2 | Various | ⚠️ Open-Core | Critical features behind paywall |

### Critical Analysis: Why Licensing Matters

#### The Percona Advantage: 100% Open Source, Zero Traps

The **Percona XtraDB Cluster (PXC) Operator** represents a complete open-source stack:
- **Operator code:** Apache 2.0 (permissive, business-friendly)
- **Database engine:** GPLv2 (same as MySQL Community Edition)
- **Container images:** Freely distributable from Docker Hub with no commercial restrictions
- **Monitoring (PMM):** 100% open source, not "open-core"
- **Backup tool (XtraBackup):** GPLv2

This stack is **free of licensing traps**. There are no "call-home" mechanisms, no features locked behind enterprise paywalls, and no ambiguous policies about what triggers a commercial license.

#### The Oracle Trap: Container Licensing Risk

While the Oracle MySQL Operator's code is GPLv2, Oracle's **container licensing policy** creates existential risk. Oracle's "Partitioning Policy" effectively states that pulling an Oracle container image (particularly MySQL Enterprise Edition) onto a Kubernetes node may expose **the entire physical host to commercial licensing requirements**.

The policy's exact scope is ambiguous, but Oracle's documentation and sales practices have historically interpreted container usage aggressively. Because Kubernetes' scheduler places pods dynamically on any available node, a single deployment mistake could theoretically trigger licensing obligations for an entire cluster—potentially millions of dollars in per-CPU-core license fees.

For organizations building on open-source principles or those without existing Oracle Enterprise Agreements, this unquantifiable risk makes the Oracle operator a non-starter.

#### When Vitess Makes Sense (And When It Doesn't)

**Vitess** is not a simple HA solution—it's a horizontal sharding platform designed for YouTube-scale workloads (petabytes, thousands of QPS). It introduces significant complexity (VTGate, VTTablet, Topology Service) that's only justified when you've definitively outgrown vertical scaling.

**Use Vitess when:** Your application requires sharding across dozens or hundreds of MySQL instances, and you have the engineering team to operate it.

**Don't use Vitess when:** A single large MySQL primary (even replicated) can handle your workload. For 99% of applications, Vitess is overkill.

#### Moco and Bitpoke: Valid but Less Integrated

Both **Moco** (from Cybozu) and **Bitpoke** (evolved from Presslabs) are valid Apache 2.0-licensed operators. However, they're less "batteries-included" than Percona:
- **Moco** focuses on GTID-based semi-synchronous replication but lacks tight integration with a comprehensive monitoring and backup ecosystem.
- **Bitpoke** uses Orchestrator for failover management (an external dependency) and doesn't provide the first-party backup and monitoring integration that Percona offers out of the box.

#### KubeDB: The Open-Core Trap

**KubeDB** is a "swiss-army knife" operator for multiple database types, but it operates on an **open-core model**. While the base operator may be free, critical production features (automated backups, monitoring, advanced HA) are often locked behind commercial licenses. This model is antithetical to a truly open-source project's goals.

### The Critical Distinction: Percona PS vs. PXC

The research prompt specified `KubernetesPerconaMysqlOperator`, which refers to the **Percona Server (PS) Operator**. However, according to Percona's official documentation and GitHub repository, **the PS Operator is a tech preview release and not recommended for production**.

The **production-ready, enterprise-grade solution** from Percona is the **Percona XtraDB Cluster (PXC) Operator**, which uses:
- **Galera library** for virtually synchronous replication (zero RPO)
- **HAProxy or ProxySQL** for intelligent connection routing
- **Percona XtraBackup** for hot, non-locking physical backups
- **Percona Monitoring and Management (PMM)** for comprehensive observability

The PS Operator, by contrast, uses traditional asynchronous replication or MySQL Group Replication and requires the external Orchestrator tool for failover—a significantly less mature and integrated architecture.

**Planton's decision:** We deploy the operator infrastructure that **supports both PS and PXC**. However, users building production workloads should create `PerconaXtraDBCluster` custom resources, not `PerconaServerMySQL` resources. The naming of our API resource acknowledges the operator package but doesn't constrain users to the tech-preview PS implementation.

## Why Galera (PXC) Changes the Game

Understanding why Percona XtraDB Cluster's synchronous replication is architecturally superior clarifies why it's the recommended production approach.

### Traditional Asynchronous Replication

In MySQL's traditional replication:
1. Client sends `COMMIT` to the primary
2. Primary commits the transaction to its local binlog and storage
3. Primary returns "success" to the client immediately
4. Primary asynchronously sends the transaction to replicas (seconds or minutes later)

**The problem:** If the primary crashes after acknowledging the commit but before replicas receive it, that transaction is **permanently lost**. Your application believes the data was saved, but it's gone. This creates a non-zero **Recovery Point Objective (RPO)**.

### Galera Synchronous Replication

In Galera's virtually synchronous model:
1. Client sends `COMMIT` to a node
2. Node writes to its local binlog and broadcasts the transaction to all other cluster members
3. All nodes receive, certify, and prepare to commit the transaction
4. **Only after a quorum (majority) of nodes confirms** does the originating node return "success" to the client
5. All nodes then commit simultaneously

**The benefit:** It is **mathematically impossible** to lose a committed transaction. If the client received "success," a majority of nodes have that transaction. If a node fails, all surviving nodes are guaranteed to have all committed data. This achieves **zero RPO** (Recovery Point Objective).

**The trade-off:** Higher write latency due to the inter-node round trip. For this reason, Galera clusters are not ideal for geographically distributed deployments. However, for regional deployments (multi-AZ within a cloud region), the ~1-5ms inter-node latency is acceptable for most workloads, and the data safety guarantee is unmatched.

### Automated Failover Without Promotion Drama

Because all Galera nodes are active and have identical data, failover is trivial. When the proxy (HAProxy or ProxySQL) detects that the node it's routing to has failed a health check, it instantly routes traffic to another healthy node. There's no complex "promote replica to primary" logic, no need to reconfigure replication chains, and no risk of promoting a replica that's behind in replication.

### The Quorum Requirement (and Common Pitfall)

Galera's greatest strength is also its most misunderstood feature: **quorum**. A cluster requires a majority of nodes to be healthy and communicating to accept writes. In a 3-node cluster, 2 nodes must be healthy. This prevents split-brain scenarios where multiple nodes independently accept conflicting writes.

**Common pitfall:** If 2 of 3 nodes fail (or are network-partitioned), the remaining single node will **intentionally** go read-only and refuse writes. This is correct behavior—without a quorum, the node cannot safely determine if another partition exists that believes it's the primary. The solution is prevention: use pod anti-affinity to spread nodes across availability zones, ensuring a single zone failure doesn't break quorum.

## Production Deployment Best Practices

Running production MySQL on Kubernetes with the Percona Operator requires adherence to proven operational patterns.

### High Availability Architecture

- **Use 3 or 5 nodes, never 2 or 4:** A 3-node cluster tolerates one node failure. A 5-node cluster tolerates two node failures. A 2-node cluster has zero fault tolerance (loss of one node breaks quorum). A 4-node cluster has the same fault tolerance as a 3-node cluster (one failure tolerated) while costing more.

- **Pod anti-affinity is mandatory:** Configure `spec.pxc.affinity.topologyKey: "kubernetes.io/hostname"` to ensure pods run on different physical nodes. For cloud deployments, use `topology.kubernetes.io/zone` to spread pods across availability zones.

- **Proxy layer for HA:** Deploy the HAProxy or ProxySQL proxy layer with at least 2 replicas. This provides a stable, load-balanced endpoint for applications and handles connection pooling.

### Storage Best Practices

- **Use high-performance storage classes:** Database performance is I/O-bound. Use storage backed by NVMe SSDs (e.g., AWS `gp3`, GCP `pd-ssd`). Avoid magnetic-disk-backed storage.

- **Ensure `allowVolumeExpansion: true`:** The StorageClass must support online volume expansion. This allows the operator to resize volumes without downtime.

- **Size for growth:** Provision storage for 6-12 months of anticipated data growth. While online expansion works, it's not zero-risk, and you don't want to perform it under pressure.

### Backup and Disaster Recovery

- **Automate daily backups:** Use `spec.backup.schedule` to define a cron schedule (e.g., `0 1 * * *` for 1 AM daily). Store backups in S3, GCS, or Azure Blob Storage.

- **Enable Point-in-Time Recovery (PITR):** Set `spec.backup.pitr.enabled: true` to enable binlog streaming. This allows restoration to any specific second, not just to the time of the last full backup.

- **Test your restores:** A backup strategy you haven't tested is not a strategy. Quarterly, spin up a new cluster in a separate namespace and restore it from a production backup to validate your DR process.

### Security Essentials

- **Enable TLS everywhere:** Use `spec.tls.enabled: true` and integrate with cert-manager for automated certificate management. Encrypt both internal node-to-node communication and external client-to-proxy connections.

- **Network Policies:** By default, all pods in Kubernetes can talk to each other. Deploy NetworkPolicy resources to:
  - Deny all ingress to the database namespace by default
  - Allow ingress to the proxy only from trusted application namespaces
  - Allow internal communication between PXC pods and proxy pods

- **Resource requests and limits:** Always set `spec.pxc.resources` with both requests and limits. This prevents the kubelet from OOMKilling database pods due to memory pressure and ensures consistent performance.

### Monitoring with PMM

The Percona Operator's integration with **Percona Monitoring and Management (PMM)** is remarkably simple:

```yaml
spec:
  pmm:
    enabled: true
    serverHost: "pmm.example.com"
```

This single configuration block triggers the operator to automatically inject `pmm-client` sidecars into all PXC and proxy pods. These sidecars register themselves with the PMM server and stream:
- Detailed database metrics (queries/second, latency percentiles, lock waits)
- Proxy performance metrics (connection counts, query routing decisions)
- Query Analytics (slow query log analysis)
- Automated advisor checks (security, performance, configuration issues)

### Common Pitfalls to Avoid

**Pitfall 1: Misunderstanding quorum behavior**
- **Symptom:** Cluster stops accepting writes after losing 2 of 3 nodes.
- **Why:** This is correct behavior—without a majority, the cluster cannot safely determine if another partition exists.
- **Fix:** Prevention via multi-zone anti-affinity. For emergency recovery, manually bootstrap the remaining node (`SET GLOBAL wsrep_provider_options='pc.bootstrap=true';`).

**Pitfall 2: Using `latest` image tags**
- **Symptom:** Unintended, uncontrolled upgrades when pods restart.
- **Fix:** Always pin specific, immutable image tags (e.g., `percona/percona-xtradb-cluster:8.0.36-28.1`). Upgrades should be deliberate, tested changes, not surprises.

**Pitfall 3: Not setting resource limits**
- **Symptom:** Pods randomly OOMKilled or experiencing severe CPU throttling.
- **Fix:** Always define `spec.pxc.resources` with requests and limits. Use the operator's auto-tuning: `innodb_buffer_pool_size={{containerMemoryLimit * 3 / 4}}` to automatically set MySQL's critical memory parameter to 75% of the container limit.

## Planton's Approach

Planton's `KubernetesPerconaMysqlOperator` resource deploys the Percona Operator infrastructure to your Kubernetes cluster. This operator is the **control plane** that enables you to declaratively manage MySQL databases.

### What Gets Deployed

When you deploy a `KubernetesPerconaMysqlOperator` resource, Planton installs:
- The Percona Operator deployment (the controller application)
- Custom Resource Definitions (CRDs) for `PerconaXtraDBCluster`, `PerconaServerMySQL`, `PerconaXtraDBClusterBackup`, `PerconaXtraDBClusterRestore`
- RBAC (ServiceAccounts, Roles, RoleBindings) required for the operator to function

### After the Operator is Installed

Once the operator is running, you create database clusters by deploying the operator's CRDs:

**For production workloads (recommended):**
```yaml
apiVersion: pxc.percona.com/v1
kind: PerconaXtraDBCluster
metadata:
  name: prod-mysql
spec:
  pxc:
    size: 3
    image: percona/percona-xtradb-cluster:8.0.36-28.1
    volumeSpec:
      persistentVolumeClaim:
        storageClassName: gp3-fast
        resources:
          requests:
            storage: 100Gi
  haproxy:
    enabled: true
    size: 3
  backup:
    schedule:
      - name: daily
        schedule: "0 1 * * *"
        keep: 7
    storages:
      s3:
        bucket: mysql-backups
        region: us-east-1
    pitr:
      enabled: true
  pmm:
    enabled: true
    serverHost: pmm.example.com
```

**For development/testing:**
```yaml
apiVersion: pxc.percona.com/v1
kind: PerconaXtraDBCluster
metadata:
  name: dev-mysql
spec:
  allowUnsafeConfigurations: true  # Required for single-node
  pxc:
    size: 1
    image: percona/percona-xtradb-cluster:8.0.36-28.1
    volumeSpec:
      persistentVolumeClaim:
        resources:
          requests:
            storage: 10Gi
  haproxy:
    enabled: true
    size: 1
```

### Why This Architecture

Planton's philosophy is **separation of concerns**:
1. The `KubernetesPerconaMysqlOperator` resource manages the **operator lifecycle** (installation, upgrades, resource allocation for the operator itself)
2. The operator's native CRDs manage the **database lifecycle** (clusters, backups, restores)

This separation allows:
- **Operator upgrades** without touching database configurations
- **Multiple database clusters** managed by a single operator instance
- **Flexibility** to use the full power of the operator's CRDs without Planton abstracting away critical features

### Abstraction Philosophy

Planton **does not** create a simplified API wrapper around database cluster creation. The operator's CRDs (`PerconaXtraDBCluster`, etc.) are already well-designed, declarative, and stable. Adding an abstraction layer would:
- Hide powerful features users need for production
- Create a maintenance burden (keeping the abstraction in sync with operator updates)
- Force users to learn two APIs instead of one

Instead, Planton focuses on the **platform layer**: deploying and managing the operator itself, ensuring it runs reliably with appropriate resource allocation and RBAC configurations.

## Conclusion: The Operator Revolution

The evolution of MySQL on Kubernetes mirrors a broader shift in how we think about infrastructure. Early attempts to force stateful workloads into stateless patterns (bare StatefulSets) failed because they ignored the complexity of database operations. Kubernetes didn't need to reinvent databases; it needed a way to **encode operational expertise into software**. The Operator pattern achieved this.

Today, running production MySQL on Kubernetes isn't a compromise or a workaround—it's a first-class deployment model that delivers on the cloud-native promise: declarative configuration, automated operations, and infrastructure as code.

The Percona XtraDB Cluster Operator stands out in this landscape not only for its technical excellence (synchronous replication, automated backups, integrated monitoring) but for its **licensing integrity**. It's a 100% open-source stack with no traps, no paywalls, and no ambiguous policies that could expose you to unexpected commercial obligations.

For production MySQL workloads on Kubernetes, the choice is clear: deploy the Percona Operator, create `PerconaXtraDBCluster` resources, and embrace the Day-2 automation that makes database operations reliable, repeatable, and remarkably boring—which is exactly what you want production operations to be.

## Further Reading

- **[Percona XtraDB Cluster Operator Official Documentation](https://docs.percona.com/percona-operator-for-mysql/pxc/)** - Comprehensive guide to all features and configurations
- **[Percona Monitoring and Management (PMM) Setup](https://docs.percona.com/percona-monitoring-and-management/)** - Deploy and configure the monitoring stack
- **[Backup and Restore Deep Dive](https://docs.percona.com/percona-operator-for-mysql/pxc/backups.html)** - Advanced backup strategies and PITR implementation
- **[Galera Replication Architecture](https://docs.percona.com/percona-xtradb-cluster/latest/intro.html)** - Understanding how synchronous replication works
- **[Kubernetes StatefulSets Limitations](https://www.plural.sh/blog/kubernetes-statefulsets-are-broken/)** - Why StatefulSets alone aren't enough

