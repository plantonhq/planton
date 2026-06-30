# Deploying MongoDB on Kubernetes: Understanding the Landscape

## The "Day 2" Challenge

For years, conventional wisdom held that stateful applications—especially databases—didn't belong on Kubernetes. The concerns were legitimate: Kubernetes' primitives were designed for stateless workloads that could be killed and restarted without ceremony. Databases, on the other hand, require careful orchestration for failover, coordinated upgrades, consistent backups, and continuous monitoring.

But something changed. Not Kubernetes' fundamental architecture—it still has no native understanding of what it means to run a MongoDB replica set. What changed was the **Operator pattern**: a way to encode expert operational knowledge into an automated controller that runs alongside your workloads. Today, running production-grade MongoDB on Kubernetes isn't just possible—with the right operator, it's often superior to traditional deployment methods.

This document explains the spectrum of deployment approaches for MongoDB on Kubernetes, from basic (and dangerous) anti-patterns to production-ready operator-based solutions. Most importantly, it explains why Planton has chosen the **Percona Server for MongoDB Operator** as the foundation for its MongoDB deployment abstraction.

## Part 1: Understanding MongoDB Distributions

Before discussing *how* to deploy MongoDB, we must first understand *what* we're deploying. The MongoDB ecosystem is fragmented across three distinct distributions, each with different capabilities and licensing models.

### MongoDB Community Edition: The Foundation

MongoDB Community Edition is the free, source-available distribution that serves as most developers' entry point. It provides core database functionality: replica sets for high availability, sharding for horizontal scalability, and the WiredTiger storage engine.

**Critical Limitations for Production:**

- **No Hot Backups**: The only built-in backup method is `mongodump`, a logical backup tool. For large databases (multi-terabyte), this is operationally infeasible—slow, resource-intensive, and incapable of achieving low Recovery Time Objectives (RTOs).

- **No Data-at-Rest Encryption**: The server provides no native encryption. Security-conscious organizations must rely entirely on filesystem-level or storage-layer encryption, which doesn't protect data if files are compromised.

- **No Enterprise Authentication**: No support for centralized identity management via LDAP or Kerberos—only native username/password authentication.

- **No In-Memory Storage Engine**: High-performance in-memory workloads are reserved for the Enterprise edition.

For development and proof-of-concept work, Community Edition is excellent. For production, especially in regulated industries, these limitations become blockers.

### MongoDB Enterprise Advanced: The Commercial Option

MongoDB Enterprise Advanced is the official, commercially-licensed offering from MongoDB Inc. It addresses all the Community Edition's limitations with a comprehensive suite of proprietary features:

- **Advanced Security**: Native data-at-rest encryption with KMIP key server support, LDAP and Kerberos authentication, and detailed audit logging
- **Operational Tooling**: Ops Manager and Cloud Manager for centralized management, monitoring, and backup orchestration
- **In-Memory Storage Engine**: For high-throughput, low-latency workloads
- **Commercial Support**: Backed by MongoDB Inc.'s support organization

This is a paid product, typically part of an enterprise subscription that includes either self-hosted licenses or MongoDB Atlas (their managed cloud service).

### Percona Server for MongoDB (PSMDB): The "Enterprise Without the License"

Percona Server for MongoDB is a free, source-available, drop-in replacement for MongoDB Community Edition. It's fully compatible—you can migrate from Community to Percona seamlessly.

Percona's strategic value proposition is bold: it provides the **enterprise features of MongoDB Enterprise Advanced, for free**, as a completely open-source-equivalent solution. Percona engineers have implemented their own versions of these enterprise features without using any proprietary code from MongoDB Inc.

**Enterprise Features Available in PSMDB (Free):**

- **Data-at-Rest Encryption**: Open-source implementation supporting HashiCorp Vault and KMIP-compliant key servers
- **Hot Backups**: Physical backup capability for replica sets, enabling low-impact backups of running databases
- **Advanced Authentication**: LDAP (via SASL), LDAP authorization, and Kerberos authentication
- **Percona Memory Engine**: Open-source alternative to MongoDB's proprietary in-memory engine
- **Audit Logging & Log Redaction**: Security and compliance features typically paywalled in Enterprise
- **Native Monitoring**: Deep integration with Percona Monitoring and Management (PMM)

### The Strategic Interception

The relationship between these distributions reveals a clear market dynamic. Users begin with free MongoDB Community Edition. As they approach production, they hit a "production wall"—compliance requires encryption-at-rest, operations require automated backups, security requires LDAP integration.

At this moment, MongoDB Inc.'s intended path is clear: upsell to MongoDB Atlas (managed cloud) or MongoDB Enterprise Advanced (self-hosted with license fees).

Percona intercepts this exact moment. It positions PSMDB as a drop-in replacement that provides precisely those missing enterprise features—at zero license cost. The value proposition is "freedom from vendor lock-in," targeting users who have the technical capability to self-manage but are unwilling or unable to pay for Enterprise licenses.

For Planton, this means supporting PSMDB isn't just a technical choice—it's a strategic one. It empowers users with production-ready, secure, operationally robust database capabilities without forcing them into costly commercial relationships.

### Distribution Comparison

| Feature | MongoDB Community | Percona Server for MongoDB | MongoDB Enterprise |
|---------|-------------------|----------------------------|-------------------|
| **Base Compatibility** | MongoDB Wire Protocol | 100% Drop-in Replacement | 100% Drop-in Replacement |
| **Data-at-Rest Encryption** | ❌ No (filesystem-level only) | ✅ **Yes (Open Source)** | ✅ Yes (Commercial) |
| **Key Server Support** | ❌ N/A | ✅ **Yes** (Vault, KMIP) | ✅ Yes (KMIP) |
| **Hot Backups** | ❌ No (mongodump only) | ✅ **Yes (Open Source)** | ✅ Yes (via Ops Manager) |
| **LDAP/Kerberos Auth** | ❌ No | ✅ **Yes (Open Source)** | ✅ Yes (Commercial) |
| **Audit Logging** | ❌ No | ✅ **Yes (Open Source)** | ✅ Yes (Commercial) |
| **In-Memory Engine** | ❌ No | ✅ **Yes (Percona Memory Engine)** | ✅ Yes (Commercial) |
| **License** | SSPL (Free) | SSPL (Free) | Commercial Subscription |
| **Operator License** | Open Core (Limited) | **Apache 2.0 (Full-Featured)** | Open Core (Requires License) |

## Part 2: The Kubernetes Deployment Spectrum

Understanding what to deploy leads to the question of how. On Kubernetes, deployment methods have evolved from simple (and dangerous) approaches to sophisticated, production-ready solutions.

### Level 0: The Anti-Pattern (Basic StatefulSets)

The most common anti-pattern is deploying MongoDB using a basic StatefulSet. This is tempting because StatefulSets provide two critical features:

1. **Stable Network Identity**: Pods get predictable hostnames (mongo-0, mongo-1, mongo-2)
2. **Stable Persistent Storage**: Each pod gets its own PersistentVolumeClaim

However, this is where StatefulSet's utility ends. A StatefulSet is application-agnostic—it has zero knowledge of MongoDB's clustering logic or operational procedures.

**Why This Fails in Production:**

- **No Replica Set Management**: StatefulSets can't initialize MongoDB replica sets. They can't run `rs.initiate()` or coordinate adding members to the cluster.

- **No Automated Failover**: When a PRIMARY pod fails, StatefulSet just restarts it. It can't interact with MongoDB's election protocol, instruct a SECONDARY to become PRIMARY, or ensure the recovered pod rejoins correctly.

- **Unsafe Upgrades**: Database upgrades require careful orchestration (upgrade secondaries sequentially, step down primary, upgrade primary). StatefulSet's default RollingUpdate strategy is database-unaware and can cause downtime or data corruption.

- **No Backup Coordination**: StatefulSets have no mechanism to trigger, schedule, or coordinate cluster-wide consistent backups.

**Verdict**: StatefulSets solve "Day 1" deployment but leave 100% of complex "Day 2" operations as high-risk manual work.

### Level 1: Helm Charts (Better Day 1, Still Manual Day 2)

Generic Helm charts (like Bitnami's MongoDB chart) are an improvement. They're pre-packaged, well-tested manifests that correctly configure the StatefulSet, Services, and Secrets required to launch a cluster.

However, these charts suffer from the same fundamental limitation: they're "fire and forget" installations. They provision the initial infrastructure but don't deploy an active controller that manages the cluster's ongoing lifecycle.

**What's Still Manual:**

- Backups and restores
- Database version upgrades
- Scaling replica set members
- Coordinated failover during node maintenance
- Security certificate rotation

**Verdict**: Excellent for initial deployment, but not a complete solution for production operations.

### Level 2: The Operator Pattern (The Production Solution)

The Kubernetes Operator pattern is the industry-standard solution for running stateful applications. An Operator is an application-specific controller that encodes expert operational knowledge into automated software.

**How Operators Work:**

1. **Custom Resource Definitions (CRDs)**: Operators extend the Kubernetes API with new resource types (e.g., `PerconaServerMongoDB`)

2. **Continuous Reconciliation**: The operator runs as a pod, watching for creation or modification of its CRDs

3. **Automated Operations**: When you create a custom resource defining your desired cluster state, the operator takes all necessary actions—creating StatefulSets, configuring replica sets, setting up backups, managing upgrades—to make reality match your specification

This is the only method that properly automates the full application lifecycle: initial deployment, automated failover, coordinated scaling, zero-downtime upgrades, and integrated backups.

## Part 3: The Operator Landscape

The MongoDB operator market is fragmented, with choices defined more by licensing strategy than technical capability. The primary production-ready options come from Percona, MongoDB Inc., and KubeDB.

### Comparative Analysis: MongoDB Operators

| Operator | Operator License | DB Deployed | Automated Upgrades | Scaling | Backups (Logical, Physical, PITR) | Key Limitation |
|----------|------------------|-------------|-------------------|---------|----------------------------------|----------------|
| **Percona Operator** | **Apache 2.0** | Percona Server for MongoDB | ✅ **Yes** | ✅ **Yes** (H/V, Storage) | ✅ **Yes** (All types) | **None—100% open source** |
| **MongoDB Community Operator** | Open Core | MongoDB Community | ❌ No | ⚠️ Limited | ❌ **NO BACKUPS** | **Not viable for production** |
| **MongoDB Enterprise Operator** | Open Core | MongoDB Enterprise | ✅ Yes | ✅ Yes | ✅ Yes (via Ops Manager) | **Requires paid license** |
| **KubeDB** | Open Core | MongoDB (various) | ❌ Manual | ⚠️ Enterprise Only | ⚠️ **Enterprise Only** | **Production features paywalled** |

*Note: MongoDB Atlas Operator is omitted—it doesn't manage databases within your cluster but provisions instances in MongoDB's cloud service.*

### The Licensing Divide

The comparison reveals a critical truth: this isn't a technical comparison between equal open-source projects. It's a choice of licensing models.

MongoDB Inc. has structured the market to make its "free" MongoDB Community Operator functionally useless for production. It lacks backup capability and can't perform automated upgrades or storage scaling. It's effectively a toy designed to push users toward paid solutions.

The other options—MongoDB Enterprise Operator and KubeDB—are "open-core." The operator is free to install, but the features that define production readiness (automated backups, restores, scaling) are paywalled behind commercial licenses.

**This leaves the Percona Operator in a unique position.** It is the only operator that is both:

1. **Permissively Licensed**: Apache 2.0, with no "open-core" restrictions
2. **Fully Feature-Complete**: Provides the entire suite of production-grade "Day 2" features for free

The choice isn't technical—it's strategic.

## Part 4: Why Planton Chose Percona

Planton's selection of the Percona Server for MongoDB Operator is based on four pillars:

### 1. True Open Source (No Hidden Costs)

The operator is licensed under Apache 2.0. This is not "open-core" with paywalled features. Every production capability—automated upgrades, full backup/restore (logical, physical, and point-in-time recovery), comprehensive scaling—is available without restriction.

For Planton's users, this means no surprise license fees, no feature gates, no vendor lock-in.

### 2. Enterprise-Grade Features Without Enterprise Licensing

Percona Server for MongoDB provides the security and operational features typically locked behind MongoDB Enterprise Advanced licenses:

- Native data-at-rest encryption with HashiCorp Vault integration
- Hot backups (physical backups) for large databases
- LDAP and Kerberos authentication
- Audit logging for compliance

Users get a production-ready, secure database stack at zero software cost.

### 3. Comprehensive Day 2 Automation

The operator automates the complex operational tasks that define production readiness:

- **High Availability**: Automated replica set management and failover
- **Sharding**: First-class support for sharded clusters (config servers, mongos routers, shards)
- **Backups**: Integrated Percona Backup for MongoDB (PBM) with scheduled backups, PITR, and S3-compatible storage
- **Monitoring**: Native integration with Percona Monitoring and Management (PMM)
- **Security**: TLS certificate management (with cert-manager integration), data-at-rest encryption
- **Upgrades**: Safe, zero-downtime rolling upgrades
- **Scaling**: Online storage expansion and compute scaling

### 4. Production-Proven and Supported

The operator is built on more than a decade of Percona's experience running databases in production. It's actively maintained, with a vibrant community forum and commercial support options available (24/7 with SLAs as low as 15 minutes for critical issues).

Percona's business model is transparent: the software is free, and they monetize through optional expert-level support contracts. This allows organizations to de-risk production deployments without paying for software licenses.

## Part 5: Licensing Clarity

A common point of confusion deserves explicit clarification: the Percona solution consists of two components with different licenses.

### The Operator: Apache 2.0 (Fully Permissive)

The operator itself—the Go controller, source code, and container images—is **100% open source under Apache 2.0**. This means:

- You can use it for any purpose, including commercial production
- You can modify and redistribute it
- There are no license fees or restrictions
- There is no "open-core" model—all features are free

### The Database Server: SSPL (Source-Available)

Percona Server for MongoDB (the `mongod` binary) is licensed under the Server Side Public License (SSPL) v1. This is the same license MongoDB Inc. uses, created to prevent cloud providers from offering MongoDB as a commercial DBaaS without contributing back.

**What SSPL Restricts**: Offering PSMDB as a commercial, managed Database-as-a-Service (DBaaS) to external customers.

**What SSPL Does NOT Restrict**: Using PSMDB to run your own applications, even in commercial, production environments.

**Conclusion**: For 99% of users—and all Planton users self-managing infrastructure—the entire stack (Apache 2.0 Operator + SSPL Server) is completely free for production use.

## Part 6: Total Cost of Ownership

The choice between MongoDB deployment options is fundamentally a "Build vs. Buy" decision:

### MongoDB Atlas (The "Buy" Model)

- **TCO**: 100% Opex. No software to manage, but costs can be high and unpredictable (especially data transfer fees)
- **Best for**: Teams with low operational capacity and high Opex budgets

### MongoDB Enterprise Advanced (The "Buy-and-Build" Model)

- **TCO**: High Capex (license fees) + High Opex (SRE team to manage it)
- **Best for**: Organizations with high operational capacity, high software budgets, and compliance mandates requiring official vendor support

### Percona Operator (The "Build" Model)

- **TCO**: $0 license cost + High Opex (SRE team to manage it)
- **Best for**: Organizations with high operational capacity and low/zero software licensing budgets

Percona creates a unique third path: **"Build-with-Enterprise-Features"**. You get all the enterprise capabilities (encryption, hot backups, LDAP, advanced monitoring) without the software license costs.

The strategic insight: Users who need enterprise features and have the engineering capability to self-manage no longer need to choose between "pay for licenses" or "live without critical features." Percona provides the third option.

## Part 7: Production Best Practices

While the operator automates complex operations, production success still requires following best practices:

### High Availability

- **Minimum topology**: 3-node replica set (`size: 3`)
- **Pod Anti-Affinity**: Set `antiAffinityTopologyKey: "kubernetes.io/hostname"` to ensure members aren't co-located on the same node
- **Result**: Near-zero downtime during node failures with automated failover

### Storage

- **Don't use default StorageClass**: Specify a production-grade class with high IOPS (e.g., `gp3`, `io1` on AWS)
- **Enable volume expansion**: Ensure `allowVolumeExpansion: true` in your StorageClass
- **Right-size initially**: Storage expansion is possible but plan capacity appropriately

### Backups

- **Remote storage**: Always store backups in S3-compatible object storage (AWS S3, MinIO, Backblaze)
- **Enable PITR**: Point-in-time recovery is critical for minimizing data loss (RPO)
- **Use physical backups**: For databases >100GB, physical backups provide significantly faster restore times (lower RTO)

### Security

- **Never use default passwords**: The operator's example secrets are for development only
- **Enable TLS**: Integrate with cert-manager for automated certificate management
- **Data-at-rest encryption**: Use HashiCorp Vault for master key management, not Kubernetes Secrets
- **Network Policies**: Restrict ingress to database pods using Kubernetes NetworkPolicy

### Monitoring

- **Enable PMM**: Set `spec.pmm.enabled: true` for MongoDB-specific dashboards and query analytics
- **Alternative**: PMM exposes Prometheus-compatible metrics if you have an existing observability stack

### What NOT to Do

- ❌ **Don't use `size: 1` or `size: 2` in production**: Minimum is 3 for proper quorum
- ❌ **Don't use 2 replicas + 1 arbiter**: This topology is unsafe and doesn't provide write guarantees
- ❌ **Don't skip resource requests/limits**: Without them, pods are OOMKilled under pressure
- ❌ **Don't hardcode connection strings**: Always use the `mongodb+srv://` SRV record for failover resilience
- ❌ **Don't manually edit StatefulSets**: All changes must go through the operator's Custom Resource

## Conclusion: A Strategic Choice for Production MongoDB

The decision to run MongoDB on Kubernetes is no longer a question of "if" but "how." The operator pattern has matured to the point where Kubernetes deployments can match or exceed traditional deployments in reliability, security, and operational efficiency.

Planton's choice of the Percona Server for MongoDB Operator reflects a clear strategic position: **production-grade database capabilities should not require expensive commercial licenses**.

For users who have—or are building—the engineering capability to self-manage infrastructure, Percona provides an exceptional value proposition:

- Enterprise security features (encryption, LDAP, audit logging)
- Enterprise operational features (hot backups, PITR, automated upgrades)
- Production-grade monitoring and observability
- Full automation of complex "Day 2" operations
- Zero software license costs
- No vendor lock-in

By building a thoughtful, opinionated abstraction layer over the Percona Operator, Planton shields users from operational complexity while preserving flexibility and avoiding costly commercial relationships.

The result: a best-in-class path for running production MongoDB on Kubernetes that's both technically excellent and economically sensible.

---

**Further Reading**:
- [Percona Operator for MongoDB Official Documentation](https://docs.percona.com/percona-operator-for-mongodb/)
- [Percona Server for MongoDB Feature Comparison](https://docs.percona.com/percona-server-for-mongodb/8.0/comparison.html)
- [Understanding SSPL Licensing](https://www.percona.com/blog/navigating-mongodb-licensing-challenges-why-percona-is-a-game-changer/)

