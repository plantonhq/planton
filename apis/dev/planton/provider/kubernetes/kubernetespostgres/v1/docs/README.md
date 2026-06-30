# PostgreSQL on Kubernetes: From Controversy to Convention

## Introduction: The Great Reversal

For years, the database community held a firm position: **stateful workloads and Kubernetes don't mix**. The reasoning was sound—Kubernetes was architected for ephemeral, stateless applications. Databases are the antithesis of ephemeral.

That wisdom is now outdated.

The industry has witnessed a fundamental paradigm shift. In 2018, Kelsey Hightower publicly stated he did not support running stateful workloads on Kubernetes. By 2023, his position had evolved: "You can run databases on Kubernetes." This isn't a change of opinion—it's recognition of a change in technological maturity.

**What changed?** Three critical advancements converged:

1. **Storage Primitives**: Kubernetes' storage subsystem matured. LocalPersistentVolumes, CSI drivers, and volume snapshots reached General Availability, providing the low-level constructs databases require.

2. **The Operator Pattern**: The Kubernetes community developed a framework for encoding complex, application-specific operational knowledge into automated controllers that extend the Kubernetes API itself.

3. **Production Validation**: Organizations like Zalando, GitLab, and thousands of others have run PostgreSQL on Kubernetes at scale for years, proving it's not just possible but preferable.

**Why this matters strategically**:

- **Economics**: Managed database services like Amazon RDS can cost 3-5x more than self-managed equivalents at scale, especially when factoring in egress costs and vendor lock-in.
- **Automation**: Kubernetes' declarative model enables database provisioning, scaling, and disaster recovery to integrate seamlessly with GitOps workflows and infrastructure-as-code.
- **Portability**: A Kubernetes-based data platform is inherently multi-cloud, giving you leverage to avoid vendor lock-in at every layer of the stack.

This document explains the deployment landscape, compares production-ready operators with rigorous attention to licensing, and provides the strategic reasoning behind our API design choices.

---

## The Maturity Spectrum: From Anti-Pattern to Production-Ready

Running PostgreSQL on Kubernetes requires solving three non-negotiable problems:

1. **Data Persistence**: Pods are ephemeral. Their filesystems vanish when they die. Your database must outlive pod restarts.
2. **Stable Identity**: Database nodes aren't cattle—they're individuals. A primary must remain identifiable as the primary. Replicas must maintain their replication streams.
3. **Day 2 Operations**: Installation is trivial compared to operations. Production databases require automated failover, continuous backups, point-in-time recovery, major version upgrades, and connection pooling.

Different deployment approaches solve these problems with varying degrees of success. Think of it as an evolutionary ladder:

### Level 0: The Anti-Pattern (Don't Do This)

**Deployments and Simple Pods** will cause data loss or corruption. This isn't hyperbole—it's a certainty.

**Scenario 1: No Persistent Storage**

If you run PostgreSQL in a pod without a PersistentVolume, the database writes to the container's ephemeral filesystem. When the pod dies (node failure, eviction, upgrade), your data is gone. Forever. No recovery possible.

**Scenario 2: Shared Volume Corruption**

If you create a Deployment with `replicas: 3` and a single PersistentVolumeClaim, Kubernetes will schedule three PostgreSQL processes that all try to mount and write to the **same** data directory. PostgreSQL's PGDATA directory expects exclusive access. Multiple processes fighting for control will corrupt the database instantly and irreversibly.

**Verdict**: This is not a deployment strategy. It's a disaster you're consciously choosing.

### Level 1: The Foundation (StatefulSet)

**StatefulSet** is the correct Kubernetes primitive for stateful applications. It solves the first two problems elegantly:

- **Stable Identity**: Pods receive predictable, stable names (`postgres-0`, `postgres-1`, `postgres-2`) that persist across restarts and rescheduling.
- **Stable Storage**: Each pod gets its own PersistentVolumeClaim via `volumeClaimTemplates`. When `postgres-1` crashes and restarts, it reattaches to **its** original volume, not a new empty one.

**What's Still Missing**: StatefulSet has zero PostgreSQL-specific knowledge. It's a generic primitive. It can restart a pod, but it doesn't understand that `postgres-0` was the primary. It can't:

- Promote a replica to become the new primary during a failover
- Reconfigure replication streams to point to the new primary
- Schedule continuous backups or perform point-in-time recovery
- Perform safe major version upgrades

**Verdict**: StatefulSet is necessary but not sufficient. You've solved the Kubernetes problem, but you're left with enormous operational toil. You're now responsible for being a PostgreSQL Database Reliability Engineer—without the automation.

### Level 2: The Packaging Trap (Helm Charts)

**Helm** is a package manager that bundles complex YAML manifests into a deployable chart. Popular charts like Bitnami's `postgresql-ha` add high-availability tooling:

- **repmgr** or **Patroni**: Manages replication setup and automates failover
- **pgpool-II**: Routes connections to the current primary
- **PgBouncer**: Connection pooling

**Why This Is a Trap**:

Helm is a "fire-and-forget" installer. It has no running component that actively manages your cluster after deployment. It's a templating engine, not an operator. If the replication manager fails, if pgpool loses track of the primary, or if a replica falls too far behind, Helm doesn't know and can't react.

**The Brittleness**: Users report instability with Helm-deployed PostgreSQL clusters—pods in CrashLoopBackoff, obscure configuration requirements (like security context tweaks for OpenShift), and a lack of meaningful logs when things fail. These charts bundle complexity rather than abstract it.

**Verdict**: Better than manual StatefulSets for Day 1, but Day 2 operations—failovers, backups, disaster recovery, upgrades—remain fragile and manual. This is declarative installation, not declarative management.

### Level 3: The Production Standard (Kubernetes Operators)

**Kubernetes Operators** are the state-of-the-art solution. An operator is software that runs inside your cluster, continuously watching your database's state and autonomously managing its entire lifecycle.

Think of it as encoding the knowledge of an expert PostgreSQL SRE into a control loop that never sleeps, never gets paged at 3 AM, and never makes a typo in a failover script.

**How Operators Work**:

1. **Custom Resource Definition (CRD)**: You define a high-level resource like `kind: Postgresql` with fields like `replicas: 3`, `storage: 500Gi`, and `backups.enabled: true`.
2. **Controller**: The operator is a pod running a control loop that watches for changes to your `Postgresql` resources.
3. **Reconciliation Loop**: The operator continuously compares your **desired state** (what you declared) with **actual state** (what's running) and takes autonomous action to reconcile any drift.

**What This Enables**:

- **Automated Failover**: Primary pod dies? The operator detects it, runs application-level health checks, selects the most up-to-date replica, promotes it to primary, reconfigures all other replicas to stream from the new primary, and updates Kubernetes Service endpoints—all in seconds, without human intervention.

- **Automated Backups**: Declare a backup schedule in your manifest. The operator configures continuous WAL archiving to S3-compatible object storage, schedules periodic full base backups, and enables point-in-time recovery—the ability to restore your database to any second in the past.

- **Automated Upgrades**: Change `postgresVersion: 15` to `postgresVersion: 16` in your YAML. The operator performs a safe, rolling major version upgrade following PostgreSQL best practices, testing the upgrade path with a temporary replica before committing.

**Verdict**: This is the only approach that provides production-grade, fully automated lifecycle management. It's the difference between having a junior DBA on-call and having an expert SRE team encoded in software.

---

## The Operator Landscape: A Licensing-First Comparison

The PostgreSQL-on-Kubernetes ecosystem has consolidated around six major operators. However, a critical question often goes unasked: **Which are truly 100% open source with no production restrictions?**

Many operators are "open-core"—the code is open source, but the container images, essential features, or production use requires a commercial license. For an open-source infrastructure-as-code project, this distinction is critical.

### Operator Comparison Matrix

| Operator | Operator License | Container Image License | HA Technology | Backup Tool | 100% Open Source? |
|:---------|:-----------------|:------------------------|:--------------|:------------|:------------------|
| **CloudNativePG** | Apache 2.0 | Apache 2.0 (no restrictions) | Kubernetes-native (no Patroni) | barman-cloud | ✅ Yes |
| **Crunchy PGO** | Apache 2.0 | Developer Program Terms | Patroni | pgBackRest | ❌ No (image restrictions) |
| **Zalando Operator** | MIT | MIT (Spilo image) | Patroni | WAL-G | ✅ Yes |
| **Percona Operator** | Apache 2.0 | Apache 2.0 (on Docker Hub) | Patroni | pgBackRest | ✅ Yes |
| **StackGres** | **AGPLv3** | AGPLv3 (commercial available) | Patroni | Custom | ❌ No (viral copyleft) |
| **KubeDB** | **Commercial** | Commercial (license required) | Proprietary | Proprietary | ❌ No (fully commercial) |

### Deep Dive: The Finalists

#### CloudNativePG (by EDB)

**Philosophy**: Pure Kubernetes-native. No external dependencies.

**Architecture**: CloudNativePG is architecturally distinct. Unlike every other operator, it does **not** use Patroni for high availability, and it does **not** use StatefulSets. Instead, it manages Pods and PersistentVolumeClaims directly and uses the Kubernetes API server itself as the Distributed Configuration Store (DCS) for leader election and cluster state.

**Strengths**:

- **100% Open Source**: Both code and container images are Apache 2.0 licensed with no production restrictions. Images are published to GitHub Container Registry (ghcr.io) with no commercial terms.
- **Modern Architecture**: The "Kubernetes-native" design is philosophically aligned with cloud-native principles—fewer moving parts, fewer external dependencies.
- **Excellent Backup System**: Uses barman-cloud for continuous WAL archiving and base backups. Native support for S3, Azure Blob, Google Cloud Storage, and S3-compatible services like Cloudflare R2 and MinIO.
- **CNCF Affiliation**: Recently became a CNCF-affiliated project, signaling strong community momentum and governance.
- **Prometheus Integration**: First-class monitoring with built-in Prometheus exporters and example Grafana dashboards.
- **Automated Upgrades**: Supports in-place major version upgrades using PostgreSQL's `pg_upgrade` tool.

**Weaknesses**:

- **Newer Architecture**: While elegant, the non-Patroni approach is less battle-tested in extreme edge cases compared to Patroni's decade-long production history.
- **Backup Format Lock-In**: Uses barman-cloud. If you have existing pgBackRest backups from another operator, you cannot restore them with CloudNativePG—backup formats are incompatible.

**Production Readiness**: High. Backed by EDB (EnterpriseDB), rapidly growing community (5,000+ GitHub stars), and a public list of production adopters.

**Best For**: Teams that prioritize pure open-source licensing, modern Kubernetes-native architecture, and want to avoid external orchestration dependencies.

---

#### Zalando Postgres Operator

**Philosophy**: Production-proven at scale. Batteries-included with integrated connection pooling.

**Architecture**: Uses the battle-tested Patroni for high availability, packaged in the "Spilo" container image (PostgreSQL + Patroni + WAL-G bundled). Uses the Kubernetes API as Patroni's Distributed Configuration Store. Manages StatefulSets and Services for the database cluster.

**Strengths**:

- **100% Open Source**: Both operator and Spilo container images are MIT licensed with zero production restrictions.
- **Extreme Battle-Testing**: Zalando has run this operator in production for over five years, managing hundreds of PostgreSQL clusters at scale. This is not a lab project—it's real-world, production-hardened infrastructure.
- **Integrated PgBouncer**: The operator can automatically deploy and configure PgBouncer connection poolers alongside your database. This is a significant differentiator—most applications need connection pooling, and having it managed declaratively is a major operational win.
- **Comprehensive Backup & DR**: Uses WAL-G for continuous backup to object storage (S3, GCS, Azure Blob, Cloudflare R2). Supports sophisticated disaster recovery patterns, including standby-then-promote workflows for cross-cluster failover.
- **Fast Major Upgrades**: Supports both in-place upgrades and clone-then-promote upgrade strategies.

**Weaknesses**:

- **Restore Complexity**: Some users report that the disaster recovery workflow is less straightforward than ideal, requiring careful orchestration for cross-cluster restores.
- **Less "Pure" K8s-Native**: The architecture (Operator → Patroni → StatefulSet) has more layers than CloudNativePG's direct pod management.

**Production Readiness**: Extremely high. Five years of continuous production use at Zalando and widespread adoption by other organizations.

**Best For**: Teams that want a 100% open-source solution with integrated connection pooling and need a production-proven operator with real-world scale validation.

---

#### Percona Operator for PostgreSQL

**Philosophy**: 100% open-source alternative to Crunchy PGO with integrated monitoring.

**Architecture**: Forked from Crunchy PGO's earlier architecture. Uses Patroni for high availability and pgBackRest for backups—the same "best-of-breed" stack as Crunchy, but without licensing restrictions.

**Strengths**:

- **100% Open Source**: Both code and container images are Apache 2.0 licensed and publicly available on Docker Hub with no restrictions.
- **Powerful Backup System**: Inherits pgBackRest, which is exceptionally feature-rich—parallel backups, incremental backups, extremely fast restores, and first-class S3 support.
- **Percona Monitoring & Management (PMM)**: Out-of-the-box integration with PMM provides enterprise-grade monitoring, query analytics, and performance dashboards with zero configuration.
- **Automated Major Upgrades**: Includes a declarative `PerconaPGUpgrade` custom resource for triggering and managing major version upgrades—a unique feature that automates one of the riskiest Day 2 operations.

**Weaknesses**:

- **Ecosystem Lock-In**: The operator's value is maximized within the Percona ecosystem (PMM, Percona Distribution for PostgreSQL). If you're not using that stack, some value is lost.

**Production Readiness**: High. Percona is a major, well-regarded player in the open-source database world.

**Best For**: Teams that want the architecture and maturity of Crunchy PGO without commercial restrictions, especially if integrated monitoring is important.

---

### Disqualified: Licensing Non-Starters

#### Crunchy PostgreSQL Operator (PGO)

**Why Disqualified**: While the operator source code is Apache 2.0, the pre-built container images required to run the operator and PostgreSQL are distributed from `registry.developers.crunchydata.com` and are subject to "Crunchy Data Developer Program" terms that restrict production use. This is a "freemium" model—code is open, but production deployment requires a commercial relationship.

**Verdict**: Unsuitable for a 100% open-source framework.

---

#### StackGres

**Why Disqualified**: The operator and container images are licensed under **AGPLv3**—the GNU Affero General Public License. This is a viral copyleft license that imposes requirements on any software that interacts with it. Most corporations and open-source projects avoid AGPL because of the legal ambiguity around what constitutes "interacting" with AGPL software in a network-service context.

The vendor acknowledges this by offering a "GPL-free" commercial license, confirming this is a "copyleft-as-a-business-model" strategy.

**Verdict**: Legal non-starter for most open-source projects and enterprises.

---

#### KubeDB (by AppsCode)

**Why Disqualified**: KubeDB is a **commercial product**. Installation explicitly requires a license file: `--set-file global.license=/path/to/license.txt`. The formerly available "Community License" has been discontinued, with the vendor stating they are "going forward... intend to focus on our commercial offerings."

**Verdict**: Not open source. Disqualified.

---

## The Critical Decision: Which Operator to Model?

After eliminating commercial and legally problematic operators, three strong, 100% open-source contenders remain:

1. **CloudNativePG**: The future-proof, pure Kubernetes-native choice
2. **Zalando Operator**: The most battle-tested, production-proven choice
3. **Percona Operator**: The best-of-breed backup and monitoring choice

### Our Recommendation: CloudNativePG

We recommend modeling the `PostgresKubernetes` API and implementation primarily on **CloudNativePG** for the following strategic reasons:

#### 1. Licensing Clarity

CloudNativePG is Apache 2.0 licensed for both code **and** container images, with no production restrictions, no commercial terms, and no vendor lock-in. For an open-source IaC project, this alignment is paramount.

#### 2. Architectural Elegance

The Kubernetes-native design—managing pods directly, using the Kubernetes API as the DCS, avoiding StatefulSets—is philosophically aligned with modern cloud-native principles. It's a forward-looking architecture that minimizes dependencies.

#### 3. Production Readiness

Backed by EDB (a major PostgreSQL vendor), affiliated with the CNCF, rapidly growing community (5,000+ stars), and a public list of production adopters demonstrate real-world validation.

#### 4. Comprehensive Backup & DR

barman-cloud provides excellent continuous backup with PITR to any S3-compatible storage (including Cloudflare R2), which is critical for multi-cloud disaster recovery.

#### 5. Clean, Minimal API

CloudNativePG's CRD is minimal and elegant, requiring only a handful of fields for a production-ready cluster. This aligns perfectly with our 80/20 API design philosophy.

### The Zalando Option

That said, **Zalando Operator** is an equally valid choice and may be preferable for specific use cases:

- If integrated PgBouncer management is a hard requirement
- If you prioritize the longest possible production track record (5+ years at scale)
- If you prefer Patroni's battle-tested HA model over a newer Kubernetes-native approach

Our IaC modules can support multiple operator backends. The `PostgresKubernetes` API abstracts the implementation details, allowing flexibility.

---

## API Design: The 80/20 Principle

The most important output of this research is understanding **which configuration fields actually matter**. A common anti-pattern in infrastructure APIs is exposing every possible knob, creating a 500-field API that's overwhelming and error-prone.

Our philosophy: **The 20% of configuration that covers 80% of use cases should be simple and obvious. The other 80% of configuration should be hidden or defaulted.**

### The Essential 20%: Fields for v1

These fields cover the vast majority of production PostgreSQL deployments:

#### Required (No Safe Default)

- **`postgresVersion`** (string, e.g., `"16"`, `"15"`): The user must consciously choose the major version. No safe default.
- **`storage.size`** (string, e.g., `"10Gi"`, `"1Ti"`): The most critical user-defined value. No safe default—this impacts cost and capability directly.

#### Required with Sensible Defaults (User Can Override)

- **`replicas`** (int, default: `1`): The primary knob for high availability. 1 is safe for development; 3 is standard for production.
- **`resources.requests.memory`** (string, default: `"512Mi"`)
- **`resources.requests.cpu`** (string, default: `"250m"`)
- **`resources.limits.memory`** (string, default: `"1Gi"`)
- **`resources.limits.cpu`** (string, default: `"500m"`)

**Why These Defaults Matter**: Pods without resource requests are assigned the `BestEffort` Quality of Service class, making them the first to be evicted during node pressure. Setting any request moves the pod to the `Burstable` class, which is far more stable. These defaults prevent a critical anti-pattern.

#### Optional (Enable Features by Presence)

- **`storage.storageClass`** (string, default: cluster default): Critical for production. Allows selecting high-IOPS storage (e.g., `"io-optimized-ssd"`) instead of slow network-based storage like NFS.
- **`backup.s3`** (struct): The presence of this block enables automated backups.
  - `backup.s3.bucket` (string, required if s3 present)
  - `backup.s3.endpoint` (string, required for R2/MinIO)
  - `backup.s3.secretName` (string, required, name of K8s Secret with credentials)
- **`backup.schedule`** (string, default: `"0 1 * * *"`): Cron schedule for base backups. Defaults to 1 AM daily.
- **`backup.retention`** (string, default: `"7d"`): How long to retain backups.
- **`disasterRecovery.restoreFrom`** (struct): Triggers a restore workflow instead of initializing a new cluster.
  - `restoreFrom.s3` (same struct as `backup.s3`)
  - `restoreFrom.timestamp` (string, optional): Target for point-in-time recovery. If omitted, restores to the latest available point.

### What to Exclude from v1

A successful v1 API is defined by what it **omits**:

- **Custom `postgresql.conf`**: This is a power-user feature that opens the door to misconfiguration. Operator-provided defaults are robust. Defer to v2 or advanced mode.
- **Pod Anti-Affinity**: Do not expose this as a user-facing option. This is a best practice that should be **automatically applied** when `replicas > 1`. All major operators support this—it's non-negotiable for real HA.
- **Security Contexts**: Should be managed by the operator/framework to ensure security and compatibility, not exposed as user configuration.
- **Connection Pooling**: While important, connection pooling (PgBouncer) is a network proxy, not part of the database itself. Consider a separate `PostgresPooler` resource or defer to v2.
- **External Access**: Highly environment-specific. The framework should create a ClusterIP service for internal access. Users can manually create LoadBalancer or Ingress resources as needed.

### Example Configurations

#### Example 1: Development (Minimal)

Single replica, no backups, small storage. Uses all defaults.

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: dev-db
spec:
  postgresVersion: "16"
  storage:
    size: "1Gi"
```

#### Example 2: Staging (HA with Backups)

Two replicas for high availability, daily backups to Cloudflare R2.

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: staging-db
spec:
  postgresVersion: "16"
  replicas: 2
  storage:
    size: "50Gi"
    storageClass: "general-purpose-ssd"
  resources:
    requests:
      memory: "2Gi"
      cpu: "1"
    limits:
      memory: "4Gi"
      cpu: "2"
  backup:
    schedule: "0 2 * * *"  # 2 AM daily
    retention: "14d"
    s3:
      bucket: "planton-staging-backups"
      endpoint: "https://your-account.r2.cloudflarestorage.com"
      secretName: "staging-s3-creds"
```

#### Example 3: Production (Disaster Recovery)

Three replicas, production-grade resources, restoring from an existing backup to a specific point in time.

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: prod-db-restored
spec:
  postgresVersion: "16"
  replicas: 3
  storage:
    size: "500Gi"
    storageClass: "io-optimized-nvme"
  resources:
    requests:
      memory: "16Gi"
      cpu: "8"
    limits:
      memory: "16Gi"
      cpu: "8"
  # Bootstrap this cluster by restoring from backup, not creating empty
  disasterRecovery:
    restoreFrom:
      s3:
        bucket: "planton-prod-backups"
        endpoint: "https://s3.us-east-1.amazonaws.com"
        secretName: "prod-s3-creds"
      # Restore to a specific point in time
      timestamp: "2025-11-10T05:30:00Z"
```

---

## Production Best Practices: What You Must Know

### High Availability

**Replication Model**: The industry standard is **asynchronous streaming replication**. The primary streams its Write-Ahead Log (WAL) to replicas in real time. This provides excellent data protection with minimal write-performance impact.

**Synchronous replication** (where writes block until replicas confirm receipt) is possible for workloads with an RPO (Recovery Point Objective) of absolute zero, but it imposes significant write latency. It's rarely the default.

**Pod Anti-Affinity**: A 3-replica cluster where all three pods are scheduled on the same Kubernetes node provides **zero** high availability. A single node failure kills the entire cluster.

**Critical Best Practice**: The framework must **automatically and non-negotiably** enforce `podAntiAffinity` rules when `replicas > 1`. This ensures Kubernetes schedules replicas on different nodes (using `topologyKey: "kubernetes.io/hostname"`), providing true resilience to node-level failures. This is not a user-facing option—it's a requirement that should be encoded in the implementation.

### Backup & Disaster Recovery

**The Technology**: Production-grade backup is built on **continuous WAL archiving**, not `pg_dump`. The operator configures PostgreSQL's `archive_command` to continuously ship WAL files (which contain every data change) to durable object storage (S3, GCS, Azure Blob, R2).

This continuous archiving, combined with periodic full "base backups," enables **Point-in-Time Recovery (PITR)**—the ability to restore the database to any given second in the past (e.g., "5 minutes before the accidental `DROP TABLE`").

**Backup Format Lock-In (Critical)**: Different operators use incompatible backup tools:

- CloudNativePG: barman-cloud
- Crunchy PGO / Percona: pgBackRest
- Zalando: WAL-G

These formats are **not interchangeable**. A barman-cloud restore command cannot read a pgBackRest backup repository. If you have 900 PostgreSQL servers with pgBackRest backups, switching to CloudNativePG means you cannot restore from those existing backups—you'd need to take new base backups with barman-cloud.

This is an acceptable trade-off for new deployments, but it's a critical consideration for migrations.

**Cross-Cluster DR**: The ultimate test of a disaster recovery plan is restoring to a **different cluster**, potentially in a different region or cloud provider, after the source cluster is destroyed.

The workflow: Create a new `PostgresKubernetes` resource, potentially in a different environment, and configure its `disasterRecovery.restoreFrom` block to point to the production backup bucket. The operator bootstraps this new cluster by **restoring** from that backup, not by initializing an empty database.

### Storage

**The Anti-Pattern**: Using slow, file-based storage like NFS or Amazon EFS for a primary database is a common mistake that causes extreme performance bottlenecks and unexpected downtime.

**Best Practice**: Always use a high-performance, block-storage `StorageClass`—AWS `gp3` or `io1`, GCP `pd-ssd`, Azure `premium-lrs`, or Ceph-RBD. This is why the `storage.storageClass` field is critical in the v1 API.

**Volume Expansion**: When a `StorageClass` has `allowVolumeExpansion: true`, you can perform a safe, often zero-downtime storage-size increase by simply changing `storage.size` from `"100Gi"` to `"150Gi"` in your manifest. The operator detects the change and triggers the PVC expansion automatically.

### Security

**Secrets**: All sensitive information—database passwords, replication credentials, S3 backup keys—**must** be managed via Kubernetes Secrets. Never hard-code credentials in manifests or ConfigMaps.

**TLS**: All client-to-database and replica-to-primary communication should be encrypted with TLS by default. The operator should handle certificate generation, rotation, and distribution transparently.

### Resource Management

**The Critical Anti-Pattern**: The single most common cause of database instability in Kubernetes is **not setting CPU and memory requests**. Pods without requests are assigned the `BestEffort` Quality of Service class, making them the **first to be evicted** by the kubelet during node pressure (memory exhaustion, CPU saturation).

**Best Practice**: Always set `resources.requests` and `resources.limits`. This assigns the pod the `Guaranteed` (if requests == limits) or `Burstable` (if requests < limits) QoS class. These pods are far more stable and protected from eviction.

This is why our API **must** provide sensible defaults for resources, protecting users from this critical, often invisible mistake.

---

## Conclusion: The Paradigm Has Shifted

Running PostgreSQL on Kubernetes is no longer controversial—it's a mature, well-supported pattern with robust, production-proven tooling. The key insight is understanding the maturity spectrum and choosing the right level of abstraction.

**For production workloads**, only Level 3 (Kubernetes Operators) provides the autonomous, continuous lifecycle management required. Among those operators, licensing and architectural considerations narrow the field significantly.

**CloudNativePG** represents the best balance of:

- 100% open-source licensing with no gotchas
- Modern, forward-looking Kubernetes-native architecture
- Production readiness backed by EDB and the CNCF
- Comprehensive backup and disaster recovery capabilities

Our `PostgresKubernetes` API abstracts the complexity, giving you production-grade PostgreSQL clusters through simple, declarative manifests. The 80/20 principle ensures the API is minimal yet powerful, with sensible defaults that prevent common anti-patterns.

Under the hood, the operator handles the hard parts—high availability, automated failover, continuous backups, disaster recovery, and major version upgrades—so you can focus on building applications, not becoming a PostgreSQL DBA.

Welcome to the modern era of database operations. Kubernetes and PostgreSQL aren't just compatible—they're a powerful combination for building cost-effective, automated, multi-cloud data platforms. 🐘☸️

---

## Further Reading

- **[Zalando Operator Deep Dive](./zalando-operator.md)** - Comprehensive guide to Zalando's architecture and operations (if you're using Zalando-based IaC modules)
- [CloudNativePG Official Documentation](https://cloudnative-pg.io/documentation/) - Comprehensive operator documentation
- [Zalando Postgres Operator Documentation](https://opensource.zalando.com/postgres-operator/) - Official upstream docs
- [CNCF: Recommended Architectures for PostgreSQL in Kubernetes](https://www.cncf.io/blog/2023/09/29/recommended-architectures-for-postgresql-in-kubernetes/) - Industry perspective on the paradigm shift
- [Data on Kubernetes (DoK) Community](https://dok.community/) - Community dedicated to running stateful workloads on Kubernetes
