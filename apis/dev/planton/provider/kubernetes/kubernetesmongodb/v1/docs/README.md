# MongoDB on Kubernetes: Choosing the Right Path

## Introduction

For years, the conventional wisdom was clear: don't run stateful databases on Kubernetes. The logic seemed sound—Kubernetes was designed for stateless workloads, and databases need stable identities, persistent storage, and complex orchestration that goes far beyond what standard Kubernetes controllers provide.

That wisdom is now outdated.

Modern Kubernetes has evolved to provide the primitives needed for stateful workloads (StatefulSets, persistent volumes, stable network identities), and more importantly, the **Operator pattern** has emerged to encode database-specific operational knowledge into self-healing controllers. Today, running MongoDB on Kubernetes isn't just viable—when done correctly, it can provide better reliability, easier scaling, and more consistent operations than traditional deployment methods.

But here's the critical insight: not all deployment methods are created equal. The path from "Hello World" to production-grade MongoDB on Kubernetes passes through several maturity levels, each solving different problems. This document explains what those levels are, which approaches are production-ready, and why Planton chose the Percona Operator as its default backend.

## The Evolution: From Anti-Patterns to Production

Understanding how to deploy MongoDB on Kubernetes requires understanding the maturity spectrum—what works, what doesn't, and why.

### Level 0: The Anti-Pattern (Deployments and ReplicaSets)

**What it is:** Using standard Kubernetes `Deployment` or `ReplicaSet` controllers to run MongoDB pods.

**Why it fails:** Kubernetes excels at managing **stateless** applications where pods are fungible—if one fails, the controller simply creates a new one with a fresh identity and empty storage. This is fundamentally incompatible with a stateful database:

1. **Loss of Identity:** A new pod gets a new hostname and IP address. For a MongoDB replica set, this new pod is an unknown entity, not the member that disappeared. The cluster loses quorum and requires manual intervention.

2. **Loss of State:** A new pod gets fresh, empty storage. The data from the failed pod is gone.

This is an **impedance mismatch**: Kubernetes abstracts away state and identity, while a database's entire value *is* its state and identity.

**Verdict:** Never use this approach for any database workload.

### Level 1: The Foundation (StatefulSets)

**What it is:** Using Kubernetes `StatefulSet` controllers, which provide stable network identities (predictable hostnames like `mongo-0`, `mongo-1`) and stable, persistent storage that survives pod restarts.

**What it solves:** StatefulSets fix the anti-pattern by ensuring that when `mongo-0` fails and restarts, it reconnects to its exact same persistent volume with its same hostname. The database cluster recognizes it as the same member returning, not a new entity.

**What it doesn't solve:** StatefulSets are just *primitives*. They don't understand MongoDB's replica set topology, election logic, or operational requirements. This creates a "Day 2" operations gap:

- **No application logic:** StatefulSets don't know how to initialize a replica set, manage elections, or perform graceful primary step-downs.
- **Manual scaling:** Scaling from 3 to 4 pods doesn't automatically run `rs.add()` to add the new member to the MongoDB cluster.
- **Orphaned storage:** StatefulSets don't delete PersistentVolumeClaims when scaled down, leaving expensive volumes stranded.

**Verdict:** Necessary foundation, but insufficient alone. You need automation on top of this.

### Level 2: The "Helm Trap" (Helm Charts)

**What it is:** Using Helm charts (like the popular Bitnami MongoDB chart) to package and deploy MongoDB on Kubernetes.

**What it solves:** Helm simplifies "Day 1" installation by templating all the required Kubernetes YAML (StatefulSets, Services, Secrets, ConfigMaps) into a single, configurable package. A simple `helm install` gives you a running MongoDB cluster in minutes.

**The trap:** This simplicity creates a false sense of production-readiness. Teams achieve a trivial installation and only hit the wall later, when they need "Day 2" operations. Helm is a **client-side, stateless templating tool**—it has no runtime controller and thus cannot automate:

- **Backups:** The Bitnami chart requires manual `mongodump` commands or separate tools like Velero.
- **Failover:** Helm plays no role in runtime operations; MongoDB's native replica set handles this, but without advanced automation.
- **Safe upgrades:** `helm upgrade` just re-templates and applies YAML. It has no application-specific knowledge to perform a safe, rolling upgrade (e.g., upgrade secondaries first, then step down and upgrade the primary).
- **Resource cleanup:** Helm doesn't track PVCs created by StatefulSets, leaving orphaned volumes after `helm uninstall`.

**Verdict:** Acceptable for development, testing, and non-critical applications. Not a production-grade solution for mission-critical workloads. The high operational cost and manual risk make this a technical debt time bomb.

### Level 3: The Production Solution (Kubernetes Operators)

**What it is:** A Kubernetes Operator is a custom controller that extends Kubernetes with application-specific operational knowledge. For MongoDB, this means a controller that watches custom resources (like `kind: MongoDB` or `kind: PerconaServerMongoDB`) and continuously reconciles the actual state of the cluster with the desired state you declare.

**What it solves:** Everything. An Operator encodes the domain expertise of a human MongoDB administrator into software:

- **Automated scaling:** Change `spec.members: 3` to `spec.members: 5`, and the Operator scales the StatefulSet, waits for new pods to be ready, then automatically runs `rs.add()` commands to join them to the cluster.
- **Automated backups:** Schedule point-in-time recovery (PITR) backups with oplog streaming to S3, GCS, or Azure Blob Storage.
- **Safe upgrades:** The Operator orchestrates rolling upgrades, upgrading secondaries first, then stepping down and upgrading the primary.
- **Self-healing:** If a pod fails, the Operator doesn't just restart it—it ensures the MongoDB cluster recognizes it and maintains quorum.

**The catch:** Not all operators are created equal. Licensing, feature completeness, and openness vary dramatically.

**Verdict:** This is the only production-ready approach for running MongoDB on Kubernetes. The question isn't "Operator or not?"—it's "*Which* Operator?"

## Operator Comparison: The Licensing Minefield

For an open-source infrastructure framework like Planton, choosing the right operator isn't just about features—it's about legal soundness, long-term viability, and avoiding "open core" traps where essential production features are paywalled.

### The Contenders

1. **MongoDB Community Operator** — Official, but incomplete
2. **MongoDB Enterprise Operator** — Official, but commercial
3. **Percona Operator for MongoDB** — 100% open source, feature-complete
4. **KubeDB MongoDB Operator** — Open core with restrictive paywalls

### Feature and Licensing Matrix

| Feature             | MongoDB Community | MongoDB Enterprise | Percona Operator | KubeDB |
|:--------------------|:------------------|:-------------------|:-----------------|:-------|
| **Code License**    | Apache 2.0        | Apache 2.0         | **Apache 2.0**   | Open Core |
| **DB License**      | SSPL ⚠️           | Commercial ⚠️      | **SSPL-Free** ✅ | SSPL ⚠️ |
| **Replica Sets**    | ✅ Free           | ✅ Free            | **✅ Free**      | ✅ Free |
| **Sharding**        | ❌ No             | 💰 Paid            | **✅ Free**      | ✅ Free |
| **Auto Backups**    | ❌ No             | 💰 Paid            | **✅ Free**      | 💰 Paid |
| **PITR**            | ❌ No             | 💰 Paid            | **✅ Free**      | 💰 Paid |
| **Safe Upgrades**   | Manual            | 💰 Paid            | **✅ Free**      | Manual |
| **Storage Scaling** | ❌ No             | 💰 Paid            | **✅ Free**      | 💰 Paid |
| **TLS/SSL**         | ✅ Free           | ✅ Free            | **✅ Free**      | 💰 Paid |
| **GUI**             | No                | Ops Mgr (Paid)     | **PMM (Free)**   | Paid |

### The MongoDB "Open Core Trap"

The MongoDB Community Operator is intentionally crippled—it lacks sharding and automated backups, which are classified as "Enterprise" features. This isn't an accident; it's product segmentation designed to upsell users to the expensive MongoDB Enterprise Advanced subscription.

More concerning for Planton: MongoDB Community Server is licensed under the **Server Side Public License (SSPL)**, a viral license that's not OSI-approved. The SSPL requires any entity offering the software *as a service* to open-source all of its own management and service-delivery code. Planton's `MongodbKubernetes` API—which abstracts and manages MongoDB deployment—could legally be considered such a service, creating existential licensing risk for the entire framework.

### The KubeDB Dead End

KubeDB represents the worst of both worlds:

1. **Legal Risk:** It deploys standard SSPL-licensed MongoDB, inheriting the full legal exposure.
2. **Extreme Open Core:** The free version is a 30-day trial. Even basic features like TLS/SSL, automated backups, and scaling are paywalled.

This is a non-starter for any serious deployment.

### The Percona Advantage

The Percona Operator for MongoDB is the only solution that avoids all legal and technical traps:

**Legally Sound:**
- The operator code is 100% Apache 2.0 licensed
- It deploys **Percona Server for MongoDB**, a fully compatible, enhanced fork that is **SSPL-free**
- No commercial licenses, no viral restrictions, no legal risk

**Production-Ready and Free:**
- ✅ Automated backups with PITR via integrated Percona Backup for MongoDB
- ✅ Automated sharding support
- ✅ Safe, orchestrated rolling upgrades
- ✅ Automated storage and compute scaling (horizontal and vertical)
- ✅ Full TLS/SSL support with automated certificate management
- ✅ Integrated monitoring with Percona Monitoring and Management (PMM)

This is not a "community edition" with missing features. This is a complete, enterprise-grade operator with all production capabilities, provided entirely free and open.

## Planton's Choice: Percona Operator

The `MongodbKubernetes` API in Planton is built on top of the **Percona Operator for MongoDB**. This is the only choice that aligns with the philosophy of an open-source infrastructure framework:

1. **No Legal Risk:** Completely Apache 2.0 licensed with an SSPL-free database fork.
2. **No Paywalls:** All production features—PITR backups, sharding, scaling—are free.
3. **Production-Proven:** Used by organizations that require enterprise-grade reliability without enterprise-grade costs.
4. **Best-Practice Defaults:** Planton's API abstracts away complexity while enforcing best practices (like pod anti-affinity for HA) under the hood.

### How It Works

When you create a `MongodbKubernetes` resource in Planton:

1. The controller ensures the Percona Operator is installed in your cluster (as a prerequisite).
2. Your simple, high-level configuration is translated into a `PerconaServerMongoDB` custom resource.
3. Best-practice defaults are applied automatically (pod anti-affinity, resource requests/limits, TLS).
4. The Percona Operator takes over, managing the full lifecycle of your MongoDB cluster.

You get the simplicity of a "Day 1" Helm experience with the power and reliability of a "Day 2" Operator solution—without the licensing traps or operational complexity.

## Configuration Philosophy: The 80/20 Principle

Most infrastructure APIs suffer from the same problem: they expose *everything*, overwhelming users with hundreds of configuration options when 80% of users only need to configure 20% of the fields.

Planton's `MongodbKubernetes` API follows the **80/20 principle**: expose the essential fields as top-level, required properties, and group advanced/edge-case options into an optional `advanced` configuration block.

### Essential Fields (What You'll Actually Configure)

1. **Replicas:** How many MongoDB members in the replica set (1 for dev, 3+ for production)
2. **Version:** The MongoDB version (e.g., "8.0")
3. **Resources:** CPU and memory requests/limits (critical for stability)
4. **Persistence:** Enable persistent storage and set disk size
5. **Authentication:** Enable auth and provide credentials (via Kubernetes Secret)

### Advanced Fields (What You Probably Won't Touch)

- Custom `mongod.conf` parameters
- Sharding topologies (config servers, mongos routers, shard count)
- Advanced backup schedules and retention policies
- Pod affinity/anti-affinity rules (sane defaults applied automatically)
- Custom initialization scripts

One field that's technically "advanced"—**pod anti-affinity**—is so critical for high availability that Planton *enforces it by default* whenever `replicas > 1`. This protects you from the common mistake of placing all replica set members on the same node, which turns a simple node failure into a full cluster outage.

### Example Configurations

**Development (Ephemeral, 1 Replica):**

```yaml
version: "8.0"
replicas: 1
persistence:
  enabled: false
resources:
  requests:
    cpu: "500m"
    memory: "1Gi"
  limits:
    cpu: "1"
    memory: "1Gi"
auth:
  rootPasswordSecret: "mongo-dev-creds"
```

**Staging (HA, 3 Replicas, Persistent):**

```yaml
version: "8.0"
replicas: 3  # 3-member replica set for HA
persistence:
  enabled: true
  size: "50Gi"
resources:
  requests:
    cpu: "1"
    memory: "4Gi"
  limits:
    cpu: "2"
    memory: "4Gi"
auth:
  rootPasswordSecret: "mongo-staging-creds"
# Pod anti-affinity enforced automatically
```

**Production (HA, 3 Replicas, Backups + PITR):**

```yaml
version: "8.0"
replicas: 3
persistence:
  enabled: true
  size: "500Gi"  # Large, production-sized volume
resources:
  requests:
    cpu: "4"
    memory: "16Gi"
  limits:
    cpu: "8"
    memory: "16Gi"
auth:
  rootPasswordSecret: "mongo-prod-creds"
backup:
  enabled: true
  pitr: true  # Point-in-Time Recovery
  schedule: "0 0 * * *"  # Daily at midnight
  retention: 5  # Keep 5 daily backups
  storage:
    type: "s3"
    bucket: "my-prod-mongo-backups"
    region: "us-east-1"
    credentialsSecret: "s3-backup-creds"
```

## Production Best Practices

Running MongoDB on Kubernetes in production requires more than just correct configuration—it requires codifying best practices into your infrastructure.

### High Availability

**Replica Sets:** Always use at least 3 members in production to maintain quorum if one member fails.

**Pod Anti-Affinity (Critical):** Ensure MongoDB pods are scheduled on different nodes. Planton enforces this by default when `replicas > 1` using `podAntiAffinity` with `topologyKey: kubernetes.io/hostname`. This prevents a single node failure from taking down multiple replica set members.

**Multi-AZ Deployment:** For true resilience, spread pods across availability zones using `topologySpreadConstraints` to ensure the cluster survives zone-level failures.

### Backup and Disaster Recovery

**Never rely on manual backups.** Production systems require automated, application-consistent backups.

**Point-in-Time Recovery (PITR):** The gold standard is continuous oplog backup, which enables recovery to any specific moment (e.g., 5 minutes before a bad query was executed). The Percona Operator provides this for free via integrated Percona Backup for MongoDB (PBM).

**Test your restores.** Backups you've never tested are backups that don't work.

### Security Hardening

1. **Authentication:** Always enable authentication (SCRAM). Never run with open access.
2. **Authorization (RBAC):** Use MongoDB's internal role-based access control. Applications should connect with least-privilege users, not the root account.
3. **Encryption-in-Transit (TLS):** All traffic—between application and database, and between MongoDB replica set members—must be encrypted using TLS/SSL.
4. **Encryption-at-Rest:** Use a StorageClass that provisions encrypted volumes (e.g., dm-crypt, cloud provider encryption).
5. **Network Isolation:** Use Kubernetes `NetworkPolicy` to enforce zero-trust networking. MongoDB pods should only accept traffic on port 27017 from authorized application pods.

### Monitoring

Deploy the Prometheus `mongodb_exporter` to scrape metrics. Key metrics to track:

- **Health:** `mongodb_up`
- **Performance:** `mongodb_op_counters_total` (inserts, queries, updates, deletes), query latency
- **Resource Usage:** `mongodb_connections`, `mongodb_memory_usage_bytes`
- **Replication:** Oplog lag, replication health

The Percona Operator integrates natively with Percona Monitoring and Management (PMM) for comprehensive, pre-built dashboards.

### Resource Planning

**Storage:** Always use a StorageClass backed by SSDs. Don't rely on default provisioners that might use slow magnetic disks.

**CPU and Memory:** *Always* set resource requests (guaranteed minimum) and limits (hard cap). Failure to do so is the #1 cause of pod evictions and production instability.

### Common Production Pitfalls

1. **Skipping resource requests/limits:** Leads to "noisy neighbor" problems and OOMKilled pods.
2. **Ignoring pod anti-affinity:** All replica set members on the same node = single point of failure.
3. **Underprovisioning IOPS:** Choosing cheap, slow storage that can't handle database I/O demands.
4. **No backup strategy:** Discovering that backups weren't configured *after* data loss.
5. **The licensing trap:** Choosing MongoDB Community Operator and discovering in production that critical features are paywalled.

## Decision Framework: Which Path Should You Take?

### Choose Helm (Bitnami Chart) If:

- You're running development, testing, or CI/CD environments
- Your workload is non-critical and you're comfortable with manual operations
- You have a dedicated database team to handle backups, scaling, and upgrades

**Trade-off:** "Day 1" simplicity for "Day 2" operational burden and risk.

### Choose an Operator If:

- You're running any production-critical workload
- You want automated backups, safe upgrades, and self-healing
- You value long-term operational efficiency over short-term setup convenience

**Trade-off:** Slightly higher "Day 1" learning curve for massive "Day 2" automation gains.

### Choose Percona Operator If:

- You need a 100% open-source solution with no legal risk
- You need enterprise-grade features (PITR, sharding) without commercial licenses
- You're building infrastructure tooling (like Planton) that manages MongoDB on behalf of users

**Trade-off:** None. This is the complete solution.

### Avoid:

- **Standard Deployments/ReplicaSets:** Not viable for any database workload.
- **MongoDB Community Operator:** Missing critical production features (backups, sharding).
- **KubeDB:** Extreme open-core model with paywalls on basic features, plus SSPL legal risk.

## Migration Considerations

If you're currently running MongoDB with Helm (e.g., Bitnami chart) and want to migrate to an Operator-managed deployment, be aware: **there is no automated migration path.**

Migration requires:

1. Deploying a new, empty cluster using the Percona Operator
2. Running `mongodump` against the old cluster
3. Running `mongorestore` into the new cluster
4. Application cutover (update connection strings, restart services)

This is a high-downtime, high-risk operation. The difficulty and risk of this migration is the single strongest argument against *starting* with Helm. You trade short-term convenience for long-term technical debt.

Planton avoids this trap by starting you on an Operator-based deployment from day one, giving you the simplicity of a Helm-like API with the power of a production-grade Operator underneath.

## Conclusion: The Paradigm Shift

The database-on-Kubernetes paradigm shift is complete. The question is no longer *whether* you can run MongoDB on Kubernetes in production—it's *how you choose to do it*.

Helm charts provide a quick start but leave you with a mountain of manual operational work. "Official" operators from MongoDB Inc. trap you in an open-core licensing model that paywalls essential features and exposes you to legal risk via the SSPL.

The Percona Operator for MongoDB represents a third path: a 100% open-source, legally sound, enterprise-grade solution that automates the full lifecycle of MongoDB on Kubernetes—without paywalls, without commercial licenses, and without compromise.

Planton builds on this foundation, providing a simple, opinionated API that enforces best practices by default while giving you full access to advanced features when you need them. You get the best of both worlds: the ease of "Day 1" and the power of "Day 2."

Start on the right path from the beginning. Your future self will thank you.

