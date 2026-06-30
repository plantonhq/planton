# Deploying ClickHouse on Kubernetes: From StatefulSet Chaos to Operator Excellence

## Introduction: The Stateful Database Paradox

For years, conventional wisdom declared that stateful databases don't belong on Kubernetes. The argument was simple: Kubernetes excels at orchestrating stateless workloads—containers that can be created, destroyed, and rescheduled without consequence. Databases, by definition, are the opposite. They hold persistent state, require stable network identities, and demand careful coordination during upgrades and failures.

ClickHouse, a distributed columnar OLAP database, amplifies this challenge. A production ClickHouse deployment isn't just "a database with state." It's a complex, multi-dimensional topology of **N shards × M replicas**, coordinated by an external service (ZooKeeper or ClickHouse Keeper), with sophisticated rules for data distribution, replication, and DDL execution. The idea of managing this manually with raw Kubernetes manifests is daunting.

Yet today, ClickHouse on Kubernetes is not just viable—it's the **industry standard**. What changed?

The answer lies in the **Kubernetes Operator pattern**. An operator is an application-specific controller that encodes operational expertise into software. Instead of manually managing dozens of interconnected resources (StatefulSets, ConfigMaps, Services, PersistentVolumeClaims), you define a single, high-level custom resource like `ClickHouseInstallation`. The operator—acting as a "robotic SRE"—continuously reconciles your desired state with reality, automating everything from initial cluster provisioning to zero-downtime rolling upgrades and topology-aware scaling.

This guide explores the evolution of ClickHouse deployment methods on Kubernetes, from brittle anti-patterns to production-ready solutions. We'll examine why the **Altinity ClickHouse Operator** emerged as the industry standard, validate its production readiness, and explain why Planton chose it as the foundation for our `ClickHouseKubernetes` API.

---

## The Deployment Evolution: From Manual Chaos to Automated Excellence

### Level 0: The Anti-Pattern — Raw StatefulSets

Let's start with what **not** to do. The Kubernetes `StatefulSet` controller provides the essential primitives for stateful workloads: stable pod identities (pod-0, pod-1, pod-2) and persistent storage via PersistentVolumeClaims. This makes it theoretically possible to deploy ClickHouse manually.

**Why it fails in practice:**

1. **Topological Ignorance**: A StatefulSet is a flat construct—a single controller managing N identical pods. ClickHouse's production topology is multi-dimensional: **N shards × M replicas**. Forcing this complex structure into a single StatefulSet was what the ClickHouse Cloud engineering team described as a "fateful" decision that created long-term operational nightmares around scaling and upgrades.

2. **Operational Rigidity**: The most critical flaw is the **immutable volumeClaimTemplate**. Need to resize a disk? Change a StorageClass? You must **delete and recreate** the entire StatefulSet, forcing a full cluster restart. For a production database with terabytes of data, this is unacceptable downtime.

3. **Coordination Blindness**: A StatefulSet has no understanding that ClickHouse requires ZooKeeper or ClickHouse Keeper for replication. It can't orchestrate safe upgrades that respect coordination requirements, nor can it manage the lifecycle of the coordination service itself.

4. **No Schema Awareness**: When you scale up by adding shards, ClickHouse needs to propagate schema changes and reconfigure cluster topology. A StatefulSet knows nothing about this—it just creates pods and walks away.

**Verdict**: The StatefulSet is a necessary **building block**, not the final solution. An operator succeeds by using StatefulSets (or other primitives) as an implementation detail. For example, a mature operator might create **one StatefulSet per shard**—a level of orchestration complexity that's impossible to manage manually but essential for safe operations.

### Level 1: The Stop-Gap — Helm Charts

Helm charts, such as the popular Bitnami chart, represent a "Day 1" stop-gap. They excel at the initial **installation** of ClickHouse, providing parameterized templates to bootstrap a cluster quickly.

**Why this approach hits a wall:**

The failure becomes evident during **Day 2 operations**—the ongoing management of a running cluster. Community consensus is unambiguous: *"Don't use helm. The ClickHouse Kubernetes Operator is the way to go."*

Here's the fundamental problem: **Helm is a templating and release-management tool, not a runtime controller.** A Helm chart can't:

- Perform zero-downtime rolling upgrades of ClickHouse versions
- Automate the complex orchestration of adding a new shard (creating new StatefulSets, Services, ConfigMaps, and propagating schema)
- React to cluster failures or configuration drift
- Manage the lifecycle of the coordination service

A more sophisticated pattern has emerged: using Helm to install **the operator**, which then manages ClickHouse clusters via custom resources. This correctly delegates runtime logic to an active controller while preserving Helm's package management benefits.

**Verdict**: Helm alone is inadequate for production. Use it to install the operator, not the database.

### Level 2: The Production Standard — The Kubernetes Operator Pattern

The Kubernetes Operator pattern is the **industry standard** for running complex stateful applications on Kubernetes. An operator is an application-specific controller that extends the Kubernetes API with Custom Resource Definitions (CRDs).

**How it transforms operations:**

Instead of managing low-level resources, you interact with a single, declarative custom resource:

```yaml
apiVersion: "clickhouse.altinity.com/v1"
kind: "ClickHouseInstallation"
metadata:
  name: production-cluster
spec:
  configuration:
    clusters:
      - name: main
        layout:
          shardsCount: 3
          replicasCount: 2
```

The operator acts as a **robotic SRE**, continuously reconciling this desired state with actual cluster state. This automates:

- Cluster provisioning and configuration
- Lifecycle management of ClickHouse Keeper (the coordination service)
- Safe, zero-downtime rolling upgrades
- Topology-aware scaling (adding shards or replicas)
- Integration with monitoring, backup, and observability tools

For any production-grade ClickHouse deployment on Kubernetes, the operator pattern is **non-negotiable**. It's the only methodology that provides the required automation, safety, and operational efficiency.

---

## The Operator Ecosystem: Why Altinity Won

The ClickHouse operator landscape is dominated by one mature, production-ready solution. Here's why the **Altinity ClickHouse Operator** became the de facto standard.

### The Industry Leader: Altinity ClickHouse Operator

The Altinity ClickHouse Operator ([github.com/Altinity/clickhouse-operator](https://github.com/Altinity/clickhouse-operator)) is not a new or experimental project. It's been **battle-tested in production for over five years**, first introduced in 2019.

**Production credibility:**

- **Powers Altinity.Cloud**: The operator is the foundational technology behind Altinity's managed ClickHouse platform, meaning it receives continuous professional development and testing
- **Estimated to manage tens of thousands** of ClickHouse servers worldwide
- **Publicly used by major companies**: eBay, Cisco, Twilio (Segment.io)
- **Over 2,000 GitHub stars** and an active community on Slack and GitHub

**Feature completeness:**

- Zero-downtime, rolling upgrades of ClickHouse versions
- Advanced management of sharding and replication topologies
- Automatic schema propagation when scaling
- Declarative management of ClickHouse configurations, users, and profiles
- Native Prometheus metrics integration via automatic `ServiceMonitor` creation
- **Robust, native support for ClickHouse Keeper** via a dedicated CRD (more on this below)
- Designed to work seamlessly with `clickhouse-backup` for disaster recovery
- GitOps-ready: integrates smoothly with Argo CD and Flux

### The Licensing Advantage: The "Clean Stack"

For an open-source framework like Planton, licensing is a critical, non-negotiable requirement. The analysis must cover both the **operator code** and the **container images** it deploys.

**The Altinity "Clean Stack":**

1. **Operator Code**: Apache 2.0 license—fully permissive, no restrictions
2. **Container Images**: This is the crucial differentiator. Altinity maintains its own `altinity/clickhouse-server` images, which it **explicitly guarantees are "100% Open Source"** and also Apache 2.0 licensed. This is a deliberate strategic decision to ensure "No vendor lock in. Ever."

**Contrast with the alternatives:**

- **Bitnami**: Bitnami's Helm chart is popular but comes with a significant warning. Bitnami is migrating to a "Bitnami Secure Images" model, deprecating non-hardened images and transitioning older tags to a "Legacy" repository. This signals a commercial shift that introduces licensing friction and long-term uncertainty.

- **Official Docker Images**: The official `clickhouse/clickhouse-server` images are usable with the operator, but their licensing page is complex and may contain software under various licenses. They're not co-tested with Altinity's operator in the same way `altinity/clickhouse-server` images are.

The combination of `Altinity/clickhouse-operator` (Apache 2.0) and `altinity/clickhouse-server` (Apache 2.0) creates a **"Clean Stack"**—a fully permissive, open-source, production-tested foundation from top to bottom.

### The Alternative: RadonDB (Effectively Deprecated)

An open-source alternative exists: the **RadonDB ClickHouse Operator**. On paper, its features appear similar—CRD-based cluster creation, custom storage templates, scaling, and upgrades.

**Why it's not a viable alternative:**

1. **Stale Maintenance**: The `radondb/radondb-clickhouse-kubernetes` repository shows its last significant commit as March 2023. In the rapidly evolving Kubernetes and ClickHouse ecosystem, this is a red flag.

2. **Lacks Critical Modern Features**: RadonDB's documentation **only references ZooKeeper**. It has no apparent support for **ClickHouse Keeper**, which is the modern, production standard for all new deployments. The Altinity operator, in contrast, introduced robust native support for Keeper in version 0.24.0.

Given its inactive maintenance and missing support for essential modern features, RadonDB is effectively deprecated and unsuitable for production use.

### Other Solutions: Vendor-Specific and Proprietary

The ecosystem includes several proprietary or platform-specific solutions:

- **ClickHouse Cloud**: Runs on a proprietary, in-house operator (also using StatefulSets as a primitive)
- **KubeDB**: Part of AppsCode's broader multi-database commercial platform. It lists ClickHouse as one of many databases it manages, but it's not a dedicated, community-focused operator with the same depth as Altinity's.

**Verdict**: The Altinity operator is the **only mature, production-grade, community-centric open-source solution** available.

### Operator Comparison Table

| Feature                     | Altinity ClickHouse Operator                                 | RadonDB ClickHouse Operator                    |
|:---------------------------|:------------------------------------------------------------|:----------------------------------------------|
| **Code License**            | Apache 2.0                                                  | Apache 2.0                                    |
| **Recommended Image**       | `altinity/clickhouse-server` (Apache 2.0)                   | Various                                       |
| **Production Readiness**    | Proven (powers Altinity.Cloud; eBay, Cisco, Twilio)        | Limited public evidence                       |
| **Maintenance Activity**    | Very Active (v0.24.0, v0.25.5 in 2024+)                     | Stale (last update March 2023)                |
| **ClickHouse Keeper Support** | **Yes** (Native `ClickHouseKeeperInstallation` CRD)        | **No** (Only ZooKeeper documented)            |
| **Clustering/Sharding**     | Advanced, topology-aware                                    | Basic                                         |
| **Automated Upgrades**      | Yes (Zero-downtime rolling upgrades)                        | Yes (Basic)                                   |
| **Backup Integration**      | Yes (via `clickhouse-backup` sidecar pattern)               | Not specified                                 |
| **Community Health**        | Active (Slack, GitHub, enterprise support available)        | Inactive                                      |

---

## The Coordination Revolution: ClickHouse Keeper

ClickHouse replication and distributed DDL execution require a coordination service. The ecosystem has decisively shifted from the legacy **ZooKeeper** to the modern, native **ClickHouse Keeper**.

### The End of ZooKeeper

ZooKeeper was the original coordination engine for ClickHouse. But it comes with well-known "pain points":

1. **zxid Overflow**: A 32-bit transaction ID that overflows after ~2 billion transactions, forcing a disruptive restart of the entire ZooKeeper ensemble
2. **1MB Data Limit**: A hard limit on data stored in a single znode, which can be problematic
3. **Operational Overhead**: As a Java application, it requires JVM heap tuning and is susceptible to "Full GC" pauses that can freeze the service and destabilize the entire ClickHouse cluster

### Why ClickHouse Keeper is Superior

**ClickHouse Keeper** is the official "drop-in replacement" for ZooKeeper, written in **C++** and built directly into the ClickHouse binary.

**Key advantages:**

- **Protocol Compatibility**: It implements the same client-server protocol as ZooKeeper. ClickHouse and client libraries can connect to Keeper without modification.
- **Superior Performance**: A production case study by Bonree reported a **>75% reduction in CPU and memory usage** and **performance improved nearly 8 times** after migrating from ZooKeeper to Keeper.
- **No JVM Issues**: Being native C++, it has no garbage collection pauses, no zxid overflow, and no 1MB data limit.
- **Production-Ready**: Battle-tested in ClickHouse Cloud since May 2022 and declared "ready for prod use" by Altinity as of ClickHouse version 22.3.

**Migration path**: For existing clusters, the `clickhouse-keeper-converter` tool can convert ZooKeeper logs and snapshots to Keeper format, enabling offline migration.

### Operator-Managed Keeper: The Best Practice

The Altinity operator (version 0.24.0+) solves the "auto-managed vs. external" coordination dilemma with a second Custom Resource Definition: **ClickHouseKeeperInstallation** (CHK).

**The pattern:**

1. Define a simple `ClickHouseKeeperInstallation` resource (e.g., 3 replicas for HA)
2. The **same operator** that manages ClickHouse also deploys and manages the Keeper cluster
3. Point your `ClickHouseInstallation` at the Keeper Service name

This "auto-managed" pattern is the **clear best practice**. It simplifies deployment, reduces operational burden, and keeps the entire ClickHouse stack (database + coordinator) within a unified management domain.

**Example:**

```yaml
apiVersion: "clickhouse-keeper.altinity.com/v1"
kind: "ClickHouseKeeperInstallation"
metadata:
  name: keeper-prod
spec:
  configuration:
    clusters:
      - name: keeper-cluster
        layout:
          replicasCount: 3  # 3-node HA ensemble
```

---

## The 80/20 Configuration Rule

The Altinity operator's API is powerful and comprehensive. But production deployments consistently configure only a small subset of fields. Here's the essential configuration surface.

### Essential Fields (Always Configure)

#### 1. Cluster Topology

The most fundamental decision: how many shards and replicas?

```yaml
spec:
  configuration:
    clusters:
      - layout:
          shardsCount: 3      # Horizontal data partitioning
          replicasCount: 2    # Replicas per shard for HA
```

**Why this matters**: Shards scale **write throughput** and **total data volume**. Replicas scale **read concurrency** and **resiliency**. A production standard is **3 replicas per shard**.

#### 2. ClickHouse Version

```yaml
spec:
  templates:
    podTemplates:
      - name: default
        spec:
          containers:
            - name: clickhouse
              image: "altinity/clickhouse-server:24.3.5.46.altinitystable"
```

**Why this matters**: Version pinning ensures reproducible deployments. Using Altinity's stable images guarantees the "Clean Stack."

#### 3. Compute Resources

```yaml
spec:
  templates:
    podTemplates:
      - name: default
        spec:
          containers:
            - name: clickhouse
              resources:
                requests:
                  cpu: "4"
                  memory: "16Gi"
                limits:
                  cpu: "8"
                  memory: "32Gi"
```

**Why this matters**: ClickHouse is resource-intensive. Proper resource allocation is non-negotiable for production stability and performance.

#### 4. Persistence and Storage

```yaml
spec:
  templates:
    volumeClaimTemplates:
      - name: data
        reclaimPolicy: Retain  # CRITICAL: Prevents data loss
        spec:
          accessModes:
            - ReadWriteOnce
          storageClassName: gp3
          resources:
            requests:
              storage: 500Gi
```

**Why this matters**: 
- **Size**: ClickHouse needs substantial storage for columnar data
- **StorageClass**: Use production-grade storage (e.g., `gp3` on AWS, `premium-rwo` on GCP)
- **reclaimPolicy: Retain**: **This is mandatory**. Without it, deleting the `ClickHouseInstallation` resource will delete all data. This is the #1 data loss pitfall.

#### 5. Coordination Link

```yaml
spec:
  configuration:
    zookeeper:
      nodes:
        - host: keeper-keeper-prod.default.svc.cluster.local
          port: 2181
```

**Why this matters**: Links the ClickHouse cluster to its coordination service. With the "auto-managed" pattern, this points to the operator-managed Keeper Service.

### Configuration You Can Skip (Advanced Use Only)

For a minimal, production-ready API, these fields are rarely customized:

- **Custom Configs**: `spec.configuration.settings`, `spec.configuration.users`, `spec.configuration.profiles` (powerful "escape hatches" for injecting raw XML into `config.xml` and `users.xml`)
- **Pod Scheduling**: `affinity`, `tolerations`, `nodeSelector` (essential for multi-zone HA, but sane defaults often suffice)
- **Networking/Ingress**: `serviceTemplates` (the default ClusterIP is suitable for most in-cluster access)

---

## Configuration Examples

### Example 1: Development (Single-Node, Minimal Resources)

**Use case**: Local development, CI/CD testing

```yaml
apiVersion: "clickhouse-keeper.altinity.com/v1"
kind: "ClickHouseKeeperInstallation"
metadata:
  name: keeper-dev
spec:
  configuration:
    clusters:
      - layout:
          replicasCount: 1  # Single node, no HA
---
apiVersion: "clickhouse.altinity.com/v1"
kind: "ClickHouseInstallation"
metadata:
  name: clickhouse-dev
spec:
  configuration:
    zookeeper:
      nodes:
        - host: keeper-keeper-dev.default.svc.cluster.local
          port: 2181
    clusters:
      - layout:
          shardsCount: 1
          replicasCount: 1
  templates:
    podTemplates:
      - name: default
        spec:
          containers:
            - name: clickhouse
              image: "altinity/clickhouse-server:24.3.5.46.altinitystable"
              resources:
                requests:
                  cpu: "500m"
                  memory: "1Gi"
```

**Verdict**: Fast startup, minimal footprint. Not for production.

### Example 2: Production (Simple, Replicated, HA Keeper)

**Use case**: Single-shard, replicated cluster for moderate-scale production workloads

```yaml
apiVersion: "clickhouse-keeper.altinity.com/v1"
kind: "ClickHouseKeeperInstallation"
metadata:
  name: keeper-prod
spec:
  defaults:
    templates:
      podTemplate: keeper-pod
      volumeClaimTemplate: keeper-pvc
  configuration:
    clusters:
      - layout:
          replicasCount: 3  # 3-node HA ensemble
  templates:
    podTemplates:
      - name: keeper-pod
        spec:
          containers:
            - name: clickhouse-keeper
              image: "clickhouse/clickhouse-keeper:24.3.5.46"
              resources:
                requests:
                  cpu: "500m"
                  memory: "1Gi"
    volumeClaimTemplates:
      - name: keeper-pvc
        reclaimPolicy: Retain
        spec:
          accessModes:
            - ReadWriteOnce
          storageClassName: gp3
          resources:
            requests:
              storage: 10Gi
---
apiVersion: "clickhouse.altinity.com/v1"
kind: "ClickHouseInstallation"
metadata:
  name: clickhouse-prod-simple
spec:
  defaults:
    templates:
      podTemplate: ch-pod
      dataVolumeClaimTemplate: ch-pvc
  configuration:
    zookeeper:
      nodes:
        - host: keeper-keeper-prod.default.svc.cluster.local
          port: 2181
    clusters:
      - layout:
          shardsCount: 1
          replicasCount: 3  # 3 replicas for HA
  templates:
    podTemplates:
      - name: ch-pod
        spec:
          containers:
            - name: clickhouse
              image: "altinity/clickhouse-server:24.3.5.46.altinitystable"
              resources:
                requests:
                  cpu: "4"
                  memory: "16Gi"
                limits:
                  cpu: "8"
                  memory: "32Gi"
    volumeClaimTemplates:
      - name: ch-pvc
        reclaimPolicy: Retain  # CRITICAL
        spec:
          accessModes:
            - ReadWriteOnce
          storageClassName: gp3
          resources:
            requests:
              storage: 500Gi
```

**Verdict**: Robust, production-ready for many use cases.

### Example 3: Production (Clustered, Sharded, Multi-Replica)

**Use case**: High-performance, large-scale production with horizontal data partitioning

The **only significant change** from Example 2 is `shardsCount: 3`:

```yaml
spec:
  configuration:
    clusters:
      - layout:
          shardsCount: 3  # 3 shards
          replicasCount: 2  # 2 replicas per shard
```

**Critical note**: Storage size is now **per shard replica**. Total cluster storage = `shardsCount × replicasCount × storage`.

**Verdict**: Maximum performance and scalability for large-scale analytics workloads.

---

## Production Best Practices

### High Availability and Resiliency

1. **Replica Count**: The production standard is **3 replicas per shard** for strong HA and read concurrency
2. **Keeper Ensemble**: Deploy a **3-node or 5-node** ClickHouse Keeper cluster (must be odd for quorum)
3. **Zone-Awareness**: Use `affinity` and `topologySpreadConstraints` to spread replicas across availability zones

```yaml
spec:
  templates:
    podTemplates:
      - name: ch-pod
        spec:
          affinity:
            podAntiAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
                - labelSelector:
                    matchLabels:
                      clickhouse.altinity.com/cluster: main
                  topologyKey: topology.kubernetes.io/zone
```

### Backup and Disaster Recovery

**The tool**: Altinity's open-source `clickhouse-backup` tool supports full and incremental backups to S3, GCS, or Azure Blob.

**The Kubernetes pattern** (sidecar injection):

1. Add `altinity/clickhouse-backup` as a sidecar container to the `ClickHouseInstallation` pod template
2. Mount the same data volume (`/var/lib/clickhouse`) to both the ClickHouse container and the sidecar
3. Run the sidecar with `command: ["server"]` to expose a REST API on port 7171
4. Create a Kubernetes `CronJob` that curls the sidecar's API to trigger backups

**Critical note**: `clickhouse-backup` provides full and incremental **snapshots**, not continuous Point-in-Time Recovery (PITR). True PITR is currently a ClickHouse Cloud feature.

### Monitoring and Observability

The Altinity operator automates Prometheus integration:

- Automatically configures the `<prometheus>` endpoint in ClickHouse's config
- Can create `ServiceMonitor` resources for Prometheus to scrape
- Helm chart can provision default Grafana dashboards

**Best practice**: Enable monitoring at installation time via Helm values:

```yaml
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
```

### Security Hardening

1. **NetworkPolicy**: Restrict access to ClickHouse ports (8123, 9000) to trusted application pods only
2. **Authentication**: Manage users declaratively via `spec.configuration.users` in the CHI manifest or use ClickHouse's SQL-driven workflow (`CREATE USER...`)
3. **Encryption at-Rest**: Use encrypted storage classes (e.g., encrypted EBS on AWS)
4. **Encryption in-Transit (TLS)**: Set `spec.configuration.clusters[].secure: "yes"` in the operator. This automatically enables TLS ports (9440, 8443) and routes distributed queries securely.

### Scaling Patterns

1. **Horizontal Scaling (Adding Replicas)**: Increase `replicasCount`. Scales **read concurrency** and **resiliency**. The operator performs a rolling update.
2. **Horizontal Scaling (Adding Shards)**: Increase `shardsCount`. Scales **write throughput** and **total data volume**. **Critical warning**: The operator does **not** rebalance existing data. New data is sharded across the new topology, but old data remains on original shards.
3. **Vertical Scaling**: Modify CPU/memory resources. The operator performs a safe, rolling restart.
4. **Storage Expansion**: Modify `storage` in `volumeClaimTemplates`. If the StorageClass supports online expansion, Kubernetes handles it automatically.

### Common Pitfalls and Troubleshooting

1. **Data Loss on Deletion**: The #1 pitfall. Always set `reclaimPolicy: Retain` on all `volumeClaimTemplates`.
2. **Shard Removal Data Loss**: Decreasing `shardsCount` **does not** migrate data. Data on removed shards will be lost if not manually migrated first.
3. **CrashLoopBackOff on Startup**: Almost always a configuration error in `spec.configuration.settings` or corrupted metadata after a failed upgrade.
4. **"Too Many Parts" Error**: This is an application-level anti-pattern, not infrastructure failure. Caused by inserting data in many small batches instead of fewer large batches.
5. **Coordinator Under-Provisioning**: Using a single-node Keeper for production or providing insufficient resources to the Keeper ensemble will cripple the entire cluster.

---

## Why Planton Chose Altinity

After evaluating the entire landscape, we integrated the **Altinity ClickHouse Operator** into Planton's IaC framework for clear, decisive reasons:

### 1. Industry Standard and Production-Proven

With over 5 years of production use, adoption by companies like eBay and Cisco, and its role powering Altinity.Cloud, this operator is battle-tested at scale. It's not an experimental project—it's the de facto industry standard.

### 2. The "Clean Stack" Licensing Advantage

The combination of Apache 2.0 operator code and Apache 2.0 container images (`altinity/clickhouse-server`) provides a **fully permissive, open-source foundation** with no production restrictions or vendor lock-in. This aligns perfectly with Planton's open-source philosophy.

### 3. Modern ClickHouse Keeper Support

Native support for ClickHouse Keeper via the `ClickHouseKeeperInstallation` CRD is essential for modern deployments. The operator's unified management of both the database and coordination layers is a significant operational simplification.

### 4. Feature Completeness

Zero-downtime upgrades, topology-aware scaling, declarative configuration management, native Prometheus integration, and seamless `clickhouse-backup` integration—the operator provides everything required for production-grade operations.

### 5. Active Maintenance and Community Health

The operator receives continuous development, with recent major releases (v0.24.0, v0.25.5) adding critical features. The active community on Slack and GitHub provides strong support.

---

## The Planton API

Our `ClickHouseKubernetes` API (see `spec.proto`) abstracts the operator's complexity into a minimal, production-ready Protobuf specification. The API exposes only the essential 20% of configuration that 80% of users need:

- **Cluster topology**: Shard count, replica count per shard
- **Version**: ClickHouse version (translated to `altinity/clickhouse-server` image)
- **Resources**: CPU/memory requests and limits
- **Persistence**: Data volume size and storage class
- **Coordination**: Auto-managed Keeper configuration (replica count, storage) or external coordinator service
- **Access**: ClusterIP vs. LoadBalancer

**Critical safety feature**: The API **hardcodes `reclaimPolicy: Retain`** internally. This is not exposed as a configurable field—it's a mandatory safeguard against accidental data loss.

**Escape hatches**: For the 20% of power users who need advanced customization, the API provides:
- `map<string, string> advanced_settings` (for `spec.configuration.settings`)
- `map<string, string> advanced_users` (for `spec.configuration.users`)

This allows injecting raw XML configuration without making the Planton API a brittle, high-maintenance proxy for the operator's full CRD.

---

## Conclusion: The OLAP Database Cloud Native

The journey from "databases don't belong on Kubernetes" to "ClickHouse on Kubernetes is the industry standard" represents a fundamental shift in how we think about stateful workloads in cloud-native environments.

The **Altinity ClickHouse Operator** succeeded where raw manifests and Helm charts failed because it encoded operational expertise into software. It transformed a complex, multi-dimensional orchestration problem into a simple, declarative API. You define your desired cluster topology, and the operator handles the rest—provisioning, upgrades, scaling, and coordination.

What makes Altinity the clear choice isn't just technical excellence—it's **licensing clarity**. The "Clean Stack" of Apache 2.0 code and images provides the legal foundation essential for open-source frameworks like Planton. There are no hidden costs, no vendor lock-in, no production restrictions. Just a mature, production-proven operator that you can deploy with confidence.

Our IaC modules abstract the operator's complexity further, providing a minimal yet powerful API that enforces best practices by default. Safety features like mandatory `reclaimPolicy: Retain` prevent common pitfalls. Auto-managed ClickHouse Keeper simplifies coordination. And structured escape hatches preserve power-user flexibility without overwhelming the 80% use case.

Deploy once, scale confidently, and let the operator handle the operational complexity. That's the promise of cloud-native ClickHouse. 📊🚀☸️

---

## Further Reading

- [Altinity ClickHouse Operator Documentation](https://docs.altinity.com/altinitykubernetesoperator/) - Comprehensive operator documentation
- [Altinity ClickHouse Operator GitHub](https://github.com/Altinity/clickhouse-operator) - Source code, examples, and issues
- [ClickHouse Keeper Guide](https://clickhouse.com/docs/guides/sre/keeper/clickhouse-keeper) - Official ClickHouse Keeper documentation
- [Altinity Stable Builds](https://altinity.com/altinity-stable/) - Information about `altinity/clickhouse-server` images
- [clickhouse-backup GitHub](https://github.com/Altinity/clickhouse-backup) - Backup and restore tool
- [ClickHouse on Kubernetes Best Practices](https://altinity.com/blog/) - Altinity blog with production case studies

