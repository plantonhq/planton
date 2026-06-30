# Deploying Neo4j on Kubernetes: From Anti-Patterns to Production

## The Landscape of Graph Database Deployment

"Just use a Deployment with a PersistentVolume" — a phrase that has doomed more than a few production Neo4j installations. While Kubernetes has matured the deployment patterns for stateless web applications, stateful workloads like graph databases demand a fundamentally different approach. The challenge isn't just keeping Neo4j running; it's ensuring data consistency, enabling fault tolerance, and providing the operational automation that production systems require.

This document explores the evolution of Neo4j deployment on Kubernetes, from common anti-patterns through intermediate solutions to production-ready approaches. It explains what Planton supports and why, grounded in the realities of running graph databases at scale.

## The Maturity Spectrum: Understanding Your Options

### Level 0: The Anti-Pattern (Kubernetes Deployments)

Using a standard Kubernetes `Deployment` object for Neo4j is the most common beginner mistake, and it fails in three catastrophic ways:

**Failure 1: Data Loss**  
Deployments treat pods as ephemeral and interchangeable. When a pod is rescheduled (which happens routinely in Kubernetes), it's terminated and replaced with a new pod that has a new name, new IP, and—critically—a new empty filesystem. Without careful volume management, your entire graph database vanishes on the first pod restart.

**Failure 2: Unstable Network Identity**  
Deployments give pods random-hash-suffixed names like `neo4j-deploy-5b8f7c4d-xk2p9`. Neo4j's Causal Clustering architecture relies on the Raft consensus protocol, which requires stable, predictable network identities (like `neo4j-core-0`, `neo4j-core-1`) for cluster members to discover each other, form a quorum, and manage leader election. Random pod names break this fundamental requirement.

**Failure 3: Misconfigured Health Probes**  
A liveness probe that's too aggressive (simple HTTP check with low timeout) will interpret a long garbage collection pause or disk checkpoint as a failure and kill the pod. For a database, this forced restart triggers expensive data flushes and cluster re-synchronization, potentially leading to cascading failures. A readiness probe that's too simple (checking if port 7474 is open) will route traffic to a node that hasn't finished warming its page cache or joining the cluster, causing widespread application errors.

**Verdict:** Never use Deployments for Neo4j. This approach is fundamentally incompatible with stateful database workloads.

### Level 1: The Correct Foundation (StatefulSets)

StatefulSets are the Kubernetes primitive purpose-built for stateful workloads. They solve the critical problems that Deployments ignore:

**What StatefulSets Provide:**
- **Stable Network Identity:** Pods are created with predictable, ordinal-based names (`neo4j-core-0`, `neo4j-core-1`) that persist across restarts
- **Stable Persistent Storage:** Each pod identity is permanently bound to a unique PersistentVolumeClaim. The pod `neo4j-core-0` will always re-attach to the PVC `data-neo4j-core-0`, ensuring data persistence across restarts and rescheduling
- **Ordered Operations:** Pods are created and updated in strict order. `neo4j-core-1` won't be created until `neo4j-core-0` is fully ready, which is essential for databases that must join clusters sequentially

**What StatefulSets Don't Provide:**  
StatefulSets are building blocks, not database administrators. They have no application-specific knowledge and cannot:
- Bootstrap the first node differently from subsequent cluster members
- Distinguish Core servers from Read Replicas
- Perform application-aware backups using `neo4j-admin`
- Orchestrate complex multi-stage version upgrades
- Automatically recover from split-brain scenarios

**Verdict:** StatefulSets are necessary but insufficient. They provide the stable foundation, but you need a higher-level abstraction to handle database-specific operations.

### Level 2: The Standard Approach (Official Helm Charts)

The Neo4j deployment landscape has significantly simplified over time. Early approaches involved multiple charts (`neo4j-standalone`, `neo4j-cluster`, various labs projects), but these are all now deprecated. Today, there is a single, unified, production-ready path: the official **neo4j/neo4j** Helm chart.

**What the Unified Helm Chart Provides:**
- **Templated Best Practices:** The chart packages StatefulSets, Services, ConfigMaps, and security configurations following Neo4j's official recommendations
- **Flexible Architecture:** The same chart supports both standalone single-node instances and multi-node clusters. The distinction is made through configuration parameters, not by using different charts
- **Smooth Growth Path:** A single-node deployment can be upgraded to a cluster by modifying its configuration values, without requiring a complete reinstallation
- **Vendor Support:** This is the officially maintained, tested, and documented deployment method from Neo4j

**What Helm Charts Don't Provide:**  
Helm is a "Day 1" provisioning tool—excellent for installation and initial configuration, but fundamentally "fire-and-forget." It has no runtime awareness and cannot:
- Monitor application health post-installation
- Automate failure recovery (like re-seeding a new pod from backup)
- Schedule and manage automated backups
- Coordinate complex "Day 2" operations

**Verdict:** The official Helm chart is the foundation of any production Neo4j deployment on Kubernetes. It's the only vendor-supported, actively maintained option.

### Level 3: The Operator Dream (and Reality)

A Kubernetes Operator extends the Kubernetes API with Custom Resources and runs a controller that continuously watches and manages the application's lifecycle. Operators encode operational expertise to automate installation, backup, recovery, and upgrades—everything a human SRE would do.

**The Operator Landscape for Neo4j:**
- **No Official Self-Managed Operator:** Neo4j has explicitly stated they have "no immediate plans to make an operator available for users who want to run self-managed instances." While Neo4j uses a sophisticated internal operator to power their AuraDB managed service, this is proprietary and not available to customers
- **Community Operators Not Production-Ready:** The most visible community operator hasn't been actively maintained for years and is incompatible with modern Neo4j versions

**Planton's Role:**  
This is where Planton fits. The `Neo4jKubernetes` resource provides the operator-like experience users expect (declarative API, automated management, Day 2 operations), while building on the reliable foundation of the official Helm chart. Planton's controller translates the simple `Neo4jKubernetes` API into the complex Helm chart configuration, then manages the ongoing lifecycle.

**Verdict:** The ideal user experience requires operator-like automation built on the stable, vendor-supported Helm chart foundation. This is exactly what Planton delivers.

## Community vs. Enterprise: The Most Critical Decision

The official `neo4j/neo4j` Helm chart can deploy both Neo4j Community and Enterprise editions. The choice between them is the single most important architectural decision for your deployment, as it fundamentally determines what production features are available.

### Feature Comparison

| Feature | Community Edition | Enterprise Edition | Implication |
|---------|------------------|-------------------|-------------|
| **Deployment Model** | Single-node only | Causal Clustering (HA) | Community cannot provide fault tolerance |
| **High Availability** | ❌ None | ✅ Multi-node clusters with automatic failover | Production systems serving critical data require Enterprise |
| **Backups** | Offline only (`neo4j-admin dump`)<br/>Requires shutdown | Online "hot" backups (`neo4j-admin backup`)<br/>Zero downtime | Community requires downtime for backups |
| **Clustering** | ❌ Not supported | ✅ Core servers + Read Replicas | Read scaling requires Enterprise |
| **Security** | Basic authentication | Role-based access control (RBAC)<br/>Fine-grained ACLs<br/>LDAP integration | Advanced security is Enterprise-only |
| **License** | GPL v3 (open source) | Commercial license required | Legal and cost implications |
| **Production Use** | Development and testing | Production deployments | Community has high operational risk |

### When to Use Each Edition

**Community Edition:**  
Use for development environments, testing, proof-of-concepts, and non-critical applications where downtime for maintenance and backups is acceptable. The GPL v3 license means it's free to use, but the lack of high availability and online backups makes it unsuitable for production systems with uptime requirements.

**Enterprise Edition:**  
Required for any production system that:
- Cannot tolerate downtime (needs HA and automatic failover)
- Requires scheduled backups without service interruption
- Needs to scale read performance with Read Replicas
- Has compliance requirements for access control and audit logging

The Enterprise license represents a trade-off: direct license cost in exchange for reduced operational risk and downtime prevention.

## High Availability with Causal Clustering (Enterprise)

Causal Clustering is Neo4j's fault-tolerant, read-scalable architecture, available exclusively in Enterprise Edition.

### Architecture Overview

A Causal Cluster consists of two server roles:

**Core Servers:**  
These form the "fault-tolerant backbone" of the cluster. Core servers:
- Process all write transactions
- Replicate data using the Raft consensus protocol
- Participate in leader election
- Require an odd number of nodes (3, 5, or 7) for fault tolerance
- Must be able to survive losing `(n-1)/2` nodes (e.g., a 5-node cluster can survive 2 failures)

**Read Replicas:**  
These are fully-functional Neo4j instances optimized for read-heavy workloads. Read Replicas:
- Serve complex read-only Cypher queries
- Asynchronously replicate data from Core servers
- Do not participate in write consensus
- Scale horizontally to handle read load
- Can be added or removed without affecting cluster consensus

### Leader Election and the Raft Protocol

The Raft consensus protocol manages all write operations:

1. Core servers elect one of themselves as the **Leader**
2. All write queries are routed to the Leader
3. The Leader writes the transaction to its log and replicates it to Follower Core servers
4. The transaction commits only after a majority of Core servers acknowledge it
5. If the Leader fails, remaining Followers detect this and automatically elect a new Leader

Planton doesn't configure Raft directly—the Neo4j software handles this automatically. Planton's responsibility is ensuring the prerequisites for Raft are met: stable network identities (via StatefulSets) and reliable cluster discovery (via the Helm chart's Kubernetes API integration).

### Connection Routing with Smart Clients

Neo4j uses a "smart client" routing model. When an application connects using `bolt+routing://` or `neo4j://`:

1. The driver connects to any cluster member
2. It requests a routing table
3. The server responds with all cluster members and their current roles
4. The driver intelligently routes subsequent queries:
   - Write queries → sent only to the Leader
   - Read queries → load-balanced across Read Replicas and Followers

**External Access Challenge:**  
The routing table contains internal Kubernetes DNS names by default (like `neo4j-core-0.neo4j-headless.default.svc.cluster.local`). External clients cannot resolve these names. The simplest solution for 80% of use cases is to run your application inside the same Kubernetes cluster as Neo4j.

### Production Anti-Affinity Requirements

For true fault tolerance, Core server pods must be deployed with **Pod Anti-Affinity** rules. Without these rules, Kubernetes might schedule all Core pods on the same physical worker node. If that node fails, your entire "HA" cluster goes down simultaneously.

Anti-affinity rules instruct Kubernetes to place each Core pod on a different worker node (ideally in different availability zones). Planton should automatically configure these rules for any deployment with more than one Core server.

## Backup and Disaster Recovery

High Availability protects against infrastructure failure (a node dies). Backups protect against data corruption or human error (a developer runs `MATCH (n) DETACH DELETE n`). HA will not save you from a bad query—it will faithfully replicate that destruction across the cluster. A separate backup strategy is mandatory.

### Backup Methods

**Online Backups (Enterprise Edition):**  
The `neo4j-admin database backup` command performs "hot" backups on a live database with zero downtime. It supports both full and differential backups. This is the production standard.

**Offline Backups (Community/Enterprise):**  
The `neo4j-admin database dump` command requires shutting down the database. This creates consistent backups but incurs downtime—generally unacceptable for production.

### Recommended Production Strategy

The best-practice approach uses Kubernetes-native automation:

1. Deploy a Kubernetes `CronJob` that runs on a schedule (e.g., daily at 2 AM)
2. The CronJob mounts a volume and runs a container with the `neo4j-admin` tool
3. The container executes an online backup against the live cluster
4. A second container in the job pod copies the backup artifact to off-site object storage (AWS S3, GCS, Azure Blob)

This provides fully automated, application-consistent backups with no downtime. A significant value-add for Planton would be to provide a simple `backup.cron` API field that automatically creates and manages this entire workflow.

## Planton's Approach: The 80/20 Abstraction

Planton's `Neo4jKubernetes` resource is designed around the principle that 80% of users need only 20% of the configuration options, while the remaining 20% of advanced users need access to 100% of features.

### Current API Philosophy

The `Neo4jKubernetesSpec` currently focuses on the minimal viable configuration for single-node Community Edition deployments:

- **Container Resources:** CPU and memory allocation
- **Persistence:** Toggle for persistent storage and disk size
- **Memory Tuning:** Optional heap and page cache configuration
- **Ingress:** Simple external access via LoadBalancer with external-dns

### The Critical Memory Abstraction

The most common operational failure for Neo4j on Kubernetes is memory misconfiguration. Neo4j requires careful tuning of two memory pools:

- **JVM Heap:** Used for query execution and transaction state
- **Page Cache:** Neo4j's off-heap memory for graph data from disk

The critical constraint is: **Heap + Page Cache + OS Headroom < Container Memory Limit**

The problem: Neo4j doesn't automatically size these pools to fit the container limit. If users don't explicitly configure them, they'll either use small defaults (poor performance) or try to allocate more memory than the container allows (instant OOMKill).

**Planton's Solution:**  
Users should only specify the total container memory (e.g., `resources.memory: "16Gi"`). The Planton controller should automatically calculate and configure the heap and page cache settings, reserving ~1-2Gi for OS overhead and splitting the remaining memory appropriately between heap and cache. This automation transforms the #1 operational pitfall into a zero-configuration best practice.

### Growth Path: Enterprise and Advanced Features

The current minimal API provides a solid foundation for development and testing workloads. A future iteration could expand to support:

**Enterprise Edition Support:**
- `edition` field (community | enterprise)
- `license.accept` boolean (required for Enterprise)
- `replicas.core` for cluster sizing
- `replicas.readReplica` for read scaling
- Automatic validation (e.g., `replicas.core: 1` when `edition: community`)

**Advanced Features:**
- Structured plugin configuration (APOC, GDS, Bloom)
- LDAP integration for enterprise authentication
- Automated backup scheduling with S3/GCS integration
- Custom Neo4j configuration overrides

**Production Optimizations:**
- Automatic Pod Anti-Affinity for HA clusters
- Storage class recommendations for IOPS optimization
- Prometheus metrics exposure
- SSL/TLS certificate management

## Conclusion: The Evolution of Deployment Maturity

The Neo4j on Kubernetes deployment landscape has evolved significantly. What once required managing multiple deprecated charts, wrestling with StatefulSet configurations, and manually calculating memory parameters has consolidated into a single, vendor-supported path: the official `neo4j/neo4j` Helm chart.

Planton builds on this stable foundation, providing the operator-like automation and user-friendly API that developers expect, while maintaining compatibility with Neo4j's official deployment method. The result is a deployment experience that's simultaneously simple for common cases and flexible for advanced needs.

Whether you're deploying a development instance or building a production-grade, fault-tolerant cluster, understanding this evolution—from anti-patterns through StatefulSets to Helm charts and operator-style automation—provides the foundation for making informed architectural decisions.

Start with the fundamentals (StatefulSets and the official Helm chart), understand the distinction between Community and Enterprise editions, and leverage Planton's abstractions to eliminate common operational pitfalls. That's the path to production-ready Neo4j on Kubernetes.

