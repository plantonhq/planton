# Deploying ClickHouse on Kubernetes: From Anti-Patterns to Production-Ready Solutions

## Introduction

For years, conventional wisdom suggested that stateful databases and Kubernetes didn't mix. The early narratives warned of unpredictable pod scheduling, ephemeral storage, and operational complexity. Yet here we are: distributed databases like ClickHouse are not just running on Kubernetes—they're thriving there, powering real-time analytics at organizations like **Vimeo, eBay, Twilio, Lyft, PostHog, and Pulumi**.

What changed? The answer lies not in Kubernetes itself, but in how we deploy databases on it. The journey from "StatefulSet + manual configuration" to "production-grade operator" represents a fundamental shift in abstraction level—from managing infrastructure primitives to encoding operational knowledge as software.

This document explains the deployment methods available for running ClickHouse on Kubernetes, evaluates them against production requirements, and details why the Altinity ClickHouse Operator is the industry standard for self-hosted ClickHouse.

## The Deployment Maturity Spectrum

Running a distributed, stateful system like ClickHouse on Kubernetes presents a spectrum of approaches, each representing a different level of operational maturity.

### Level 0: The Anti-Pattern (Manual StatefulSets)

**What it is:** Deploying ClickHouse using raw Kubernetes StatefulSets, manually configuring replication, sharding, and cluster membership.

**The appeal:** StatefulSets are Kubernetes' native primitive for stateful workloads, providing stable pod identities and persistent storage. New users are naturally drawn to this "pure Kubernetes" approach.

**The fatal flaw:** StatefulSets are **infrastructure-level** primitives. They understand Pods and PersistentVolumeClaims, but have zero domain knowledge of ClickHouse's **application-level** logic—replication, sharding, schema propagation, distributed DDL, or user management.

**The evidence:** The ClickHouse Cloud team learned this lesson the hard way. They initially "leveraged StatefulSets to control server pods," which worked fine for basic operations. However, their decision to manage all replicas with a *single StatefulSet* created a critical bottleneck: vertically scaling the cluster (changing instance sizes) required a slow, one-at-a-time rolling restart of the *entire* StatefulSet. This architecture proved "operationally crippling at scale."

A mature operator avoids this pitfall by composing *multiple* StatefulSet resources (one per shard), enabling granular and parallelized operations.

**Verdict:** Using *only* StatefulSets forces human operators to manually implement complex distributed database logic. This approach is provably insufficient for "Day 2" operations like safe upgrades, dynamic scaling, schema propagation, and user management.

### Level 1: The Day 0 Stop-Gap (Helm Charts)

**What it is:** Using general-purpose Helm charts (like the Bitnami ClickHouse chart) to bootstrap a ClickHouse cluster.

**What it solves:** Helm charts excel at "Day 0" tasks—templating manifests, packaging dependencies, and providing a one-command installation experience. They're effective for getting started quickly.

**What it doesn't solve:** The "Day 2" operational challenge. A database is not a fire-and-forget deployment; it's a persistent, evolving system requiring ongoing lifecycle management.

As community discussions note, Helm "can't handle the dynamic management required to run databases properly." Helm is not a reconciliation loop. It cannot:
- Manage high-availability failover automatically
- Declaratively update user configurations on live clusters
- Safely orchestrate version upgrades with zero downtime
- Propagate schema changes to newly scaled nodes
- React to drift between desired and actual state

**The tacit admission:** Even the Bitnami chart's own documentation *discourages* using its `resourcesPreset` values for production, stating they "may not fully adapt to your specific needs." This effectively abandons users at the most critical juncture—configuring the database for production reliability.

**Verdict:** Helm charts are useful bootstrapping tools, but they're fundamentally Day 0 solutions. For a stateful, production database requiring continuous operational intelligence, they're necessary but insufficient.

### Level 2: The Production Standard (The Kubernetes Operator Pattern)

**What it is:** A purpose-built Kubernetes Operator that encodes operational expertise as software, providing a high-level Custom Resource API and running a continuous reconciliation loop.

**What it solves:** The Operator Pattern bridges the gap between infrastructure primitives and application semantics. An operator:
1. Watches for high-level Custom Resources (like `ClickHouseInstallation`)
2. Compares the *desired state* in those resources against *actual cluster state*
3. Takes application-specific actions to reconcile the difference
4. Does this continuously, automatically handling drift and failures

For ClickHouse, this automated operational logic includes:
- **Complex Topology Management:** Automatically configuring sharding and replication
- **Schema Propagation:** Applying `ON CLUSTER` DDL statements to all nodes, including newly scaled ones
- **Declarative Management:** Managing users, profiles, quotas, and configuration files as code
- **Zero-Downtime Lifecycle:** Orchestrating upgrades and scaling operations aware of ClickHouse's internal state
- **Intelligent Restarts:** Understanding which configuration changes require restarts versus hot-reloads

**The paradigm shift:** An operator transforms database operations from imperative commands ("run this script to add a replica") to declarative intent ("this cluster should have 3 replicas"). Kubernetes takes care of making reality match the declaration.

**Verdict:** For production ClickHouse on Kubernetes, the operator pattern is not just best practice—it's the **only viable approach** for teams that want to sleep at night.

## The Altinity ClickHouse Operator: Industry Standard

Among operator implementations for ClickHouse, the Altinity ClickHouse Operator stands out as the definitive, production-grade standard.

### Production Pedigree

**Maturity:** The operator celebrated its 5th birthday in 2024 and has been production-ready for years. It's estimated to manage "tens of thousands of ClickHouse servers worldwide."

**Battle-Tested at Scale:** Publicly disclosed production users include major technology organizations: **Vimeo, eBay, Twilio, Lyft, PostHog, and Pulumi**. These aren't pilot projects—these are high-stakes, mission-critical analytics workloads.

**Community Consensus:** Within the ClickHouse engineering community, the Altinity operator is consistently recommended as "the easiest way" and the default "go-to" solution for running ClickHouse on Kubernetes.

### 100% Open Source: The Anti-Lock-In Guarantee

The Altinity operator's licensing model is a **critical strategic advantage** for any organization concerned about vendor lock-in or project abandonment.

**The License:** The operator's entire Go codebase is licensed under the permissive **Apache 2.0** license. This includes the source code, container images, and all features—there are no "enterprise-only" capabilities hidden behind paywalls.

**Image Freedom:** The operator can manage *any* ClickHouse container image:
- Official `clickhouse/clickhouse-server` images (Apache 2.0)
- Altinity's own "Stable Builds" (`altinity/clickhouse-server`, also Apache 2.0, offering LTS)
- Third-party images like Chainguard's hardened builds

There is **zero** lock-in to Altinity's container images or build pipeline.

**The Virtuous Cycle:** Altinity's business model is not "open core." They don't withhold features to upsell enterprise licenses. Instead, their revenue comes from:
1. **Altinity.Cloud:** A managed SaaS/BYOC platform that *runs the 100% open-source operator* on behalf of customers
2. **Enterprise Support:** Commercial support contracts for organizations running the operator themselves

This is "dogfooding" at its purest. Altinity's commercial success is *directly dependent* on the open-source operator being excellent. Every feature they build for their managed service *must* be open-sourced because the service literally runs on the operator. This alignment creates a powerful guarantee: the operator will remain actively maintained, feature-rich, and production-grade for as long as Altinity is in business.

This stands in sharp contrast to vendors who develop critical features as cloud-only exclusives or move components into closed-source "enterprise editions."

### Core Architecture: CHI and CHK

The operator's power comes from its two key Custom Resource Definitions (CRDs).

#### ClickHouseInstallation (CHI)

The `ClickHouseInstallation` (CHI) is the primary CRD. A single CHI manifest declaratively defines an entire ClickHouse cluster—its topology, storage, configuration, users, and more.

The operator's reconciliation loop watches CHI resources and translates them into lower-level Kubernetes primitives:
- **StatefulSet(s):** To manage ClickHouse server pods (typically one per shard)
- **PersistentVolumeClaim(s):** For persistent storage, created from `volumeClaimTemplates` in the CHI
- **Service(s):** For stable network endpoints
- **ConfigMap(s):** To dynamically generate and inject ClickHouse configurations (`config.d`, `users.d`)

#### ClickHouseKeeperInstallation (CHK)

The `ClickHouseKeeperInstallation` (CHK) CRD represents a **massive architectural evolution** and is one of the operator's most significant recent improvements.

**The Old Problem:** ClickHouse requires a coordination store (like ZooKeeper) for replication. Historically, this was a major operational burden. Users had to *separately* deploy, secure, and maintain a ZooKeeper ensemble, then configure their CHI to connect to it. This added significant complexity and multiple failure domains.

**The New Solution:** ClickHouse now offers **ClickHouse Keeper**—a built-in, more efficient, ZooKeeper-compatible coordination store. As of operator version 0.24.0, Altinity introduced the CHK CRD.

Now, users can:
1. Deploy a CHK resource to get a fully-managed Keeper ensemble
2. Deploy a CHI resource that simply points to the CHK's Kubernetes service name

The operator manages the **entire stack, end-to-end**. This eliminates the ZooKeeper operational burden and dramatically simplifies the deployment topology.

> **Important:** Any production-grade IaC tool must support deploying both CHI and CHK resources to be considered feature-complete.

### Key Features: Encoded Operational Intelligence

The operator's deep domain knowledge of ClickHouse is evident in its advanced lifecycle management capabilities:

**Intelligent Configuration Management:** The operator doesn't just apply configuration changes—it's *intelligent* about the process. It understands which ClickHouse settings support hot-reload versus which require a pod restart, thereby minimizing disruption.

**Zero-Downtime Upgrades:** Upgrading a cluster is as simple as changing the `image: tag` in the CHI manifest. The operator orchestrates a safe, rolling upgrade across all nodes, respecting replication and shard topology.

**Smart Scaling:** The operator manages both horizontal scaling (adding shards/replicas) and includes **automatic schema propagation**. When new nodes join, the operator ensures they receive the cluster's current schema via `ON CLUSTER` DDL statements.

**Non-Disruptive Storage Resizing:** For storage types that support it, the operator can resize PersistentVolumeClaims *without requiring pod restarts* (as of version 0.20+). This is a significant improvement over manual StatefulSet management.

**Declarative User Management:** ClickHouse users, profiles, and quotas are defined directly in the CHI manifest. Passwords are securely sourced from Kubernetes Secrets. The operator keeps user state synchronized automatically.

### Limitations and Operational Sharp Edges

The operator is powerful, but power comes with responsibility. There are several operational pitfalls users must understand:

**Storage Reclaim Policy:** The `volumeClaimTemplates` section in a CHI manifest includes a `reclaimPolicy` field. If not explicitly set to `Retain`, it defaults to the StorageClass setting—often `Delete`. This means a `kubectl delete chi my-cluster` command will *also delete all PersistentVolumeClaims and data*. **Always set `reclaimPolicy: Retain` for production clusters.**

**Deletion Order (Finalizer Issue):** CHI resources use Kubernetes finalizers to ensure clean shutdown. The *operator* removes these finalizers during deletion. If you delete the operator's Deployment *before* deleting CHIs, those CHI resources (and their namespace) will get stuck in `Terminating` state forever. **Correct deletion order:** (1) Delete CHI/CHK resources, (2) Wait for complete deletion, (3) Uninstall operator.

**Self-Hosted Responsibility:** The operator automates lifecycle management, but you still own the full operational stack: backup strategies, monitoring, alerting, security hardening, and underlying Kubernetes infrastructure.

## Alternative: KubeBlocks (The Generalist)

The primary alternative to the Altinity operator is **KubeBlocks**—a "multi-database operator" that provides a unified API for managing 35+ database types including ClickHouse, PostgreSQL, MySQL, and Kafka.

**The Trade-off:**
- **Altinity (Specialist):** Purpose-built for *only* ClickHouse. Deep, nuanced features like native CHK support and intelligent configuration reloads.
- **KubeBlocks (Generalist):** Unified API across many databases. Attractive for enterprises managing diverse database fleets from a single control plane.

**The Signal:** Platforms like Cozystack *bundle the Altinity Operator* as their "Managed App" for ClickHouse rather than using a generalist tool. This is a strong third-party endorsement of Altinity as the best-in-breed component for production ClickHouse.

**Verdict:** For users focused on providing a production-grade ClickHouse service (rather than managing a zoo of different databases), the specialist Altinity operator delivers deeper domain expertise and better operational outcomes.

## Self-Hosted vs. Managed Services (DBaaS)

The Altinity operator represents the "Build" side of the classic "Build vs. Buy" decision.

**Managed Services (ClickHouse Cloud, Aiven, Altinity.Cloud):**
- **Pros:** Full operational abstraction. Vendor manages HA, backups, upgrades, security, observability. Lower Total Cost of Ownership when engineering time is factored in.
- **Cons:** Higher infrastructure costs. Potential vendor lock-in. Less control over data placement, network architecture, and feature timelines.

**Self-Hosted (Altinity Operator):**
- **Pros:** Full control over deployment, data sovereignty, and infrastructure. No vendor lock-in—complete data portability. Open-source foundation aligns with organizational policies.
- **Cons:** Higher operational responsibility. You own backup strategies, monitoring, security hardening, and infrastructure management.

**The Strategic Choice:** For organizations with:
- Specific security, compliance, or data sovereignty requirements
- Existing Kubernetes expertise and operations
- Strong preference for open-source, vendor-neutral infrastructure
- Need for complete data portability

...the Altinity operator is the **only solution** that delivers production-grade ClickHouse while preserving those values.

## Installation and Configuration

### The Two-Step Installation Process (Critical)

The operator's Helm chart *intentionally does not include CRD manifests*. This prevents `helm uninstall` from accidentally deleting CRDs, which would orphan all running ClickHouse clusters.

**Correct Installation/Upgrade Procedure:**
1. **Step 1:** Manually `kubectl apply` the CRD definitions from the new version's GitHub release
2. **Step 2:** Run `helm upgrade --install` to deploy the operator Deployment

```bash
# Step 1: Apply CRDs
kubectl apply -f https://raw.githubusercontent.com/Altinity/clickhouse-operator/0.25.5/deploy/operator/clickhouse-operator-install-crds.yaml

# Step 2: Install/Upgrade Operator
helm upgrade --install clickhouse-operator \
  altinity/altinity-clickhouse-operator \
  --version 0.25.5 \
  --namespace clickhouse \
  --create-namespace \
  -f values.yaml
```

> **IaC Requirement:** Any automation must implement this two-step process. A simple `helm install` wrapper is insufficient and will fail on future upgrades.

### Multi-Tenancy and Security

The most critical security decision is controlled by a single Helm value: `rbac.namespaceScoped`.

**Default (`false`):** The operator installs with cluster-wide permissions (ClusterRole/ClusterRoleBinding). It watches for and manages CHI resources in *all namespaces*.

**Secure Multi-Tenant (`true`):** Set `rbac.namespaceScoped: true` to create a namespaced Role/RoleBinding. The operator can *only* watch and manage CHI resources in its *own namespace*.

For secure, multi-tenant environments, **always set `rbac.namespaceScoped: true`**.

### Minimal Production Configuration

Here's a minimal, production-grade configuration:

```yaml
# values.yaml
rbac:
  namespaceScoped: false  # Set to 'true' for multi-tenant environments

serviceAccount:
  create: true

operator:
  image:
    tag: "0.25.5"  # Pin the version
  
  resources:  # Essential for production stability
    requests:
      cpu: 500m
      memory: 256Mi
    limits:
      cpu: 1
      memory: 512Mi
```

### GitOps Integration

The operator is explicitly designed for GitOps workflows. The standard pattern is a two-level "App of Apps":

1. **Application 1 (Operator):** An ArgoCD/Flux Application targeting the Altinity Helm repository
2. **Application 2 (Clusters):** A second Application targeting a Git repo containing your CHI/CHK manifests as YAML

This allows the entire stack—both the operator and the clusters it manages—to be declared in Git as the source of truth.

## Ecosystem Integration

### Monitoring (Prometheus/Grafana)

The operator provides "batteries-included" monitoring. With `metrics.enabled: true` (the default), the operator automatically deploys with a built-in metrics-exporter sidecar that queries ClickHouse system tables and exposes Prometheus-compatible metrics.

**For Prometheus Operator:**
```yaml
serviceMonitor:
  enabled: true
```

**For Grafana:**
```yaml
dashboards:
  enabled: true
```

The operator will create ServiceMonitor CRs and ConfigMaps with pre-built dashboard JSON.

### Backup and Restore

The operator *facilitates* backups but doesn't perform them directly. The community-standard tool is **altinity/clickhouse-backup** (also Apache 2.0).

**Production Pattern:**
1. **Sidecar:** Add the `altinity/clickhouse-backup` container to your CHI's `podTemplates`, mounting the same data volume as ClickHouse
2. **CronJob:** Deploy a separate Kubernetes CronJob that executes backup commands via `kubectl exec` targeting the sidecar

This pattern enables scheduled, automated backups to object storage (S3, GCS, etc.).

### Multi-Region Deployment

Cross-region replication is supported but requires careful architecture. The coordination store (ClickHouse Keeper) is *extremely* sensitive to network latency and should **not** be stretched across regions.

**Recommended Architecture:**
1. Deploy CHK in a single region (e.g., `us-east-1`)
2. Deploy primary CHI in `us-east-1`, pointing to the local Keeper
3. Deploy a read-replica CHI in `us-west-2`, configured to connect back to the `us-east-1` Keeper

This provides low-latency reads in the secondary region but creates a single point of failure for writes if the primary region fails.

## Planton's Choice

For Planton, the Altinity ClickHouse Operator is the **only viable choice** for providing ClickHouse on Kubernetes. This decision is based on:

**1. Production Maturity:** Five years of battle-testing at scale by organizations like eBay, Twilio, and Lyft.

**2. Open-Source Alignment:** 100% Apache 2.0 licensed, with a business model ("dogfooding") that guarantees long-term project health without vendor lock-in.

**3. Complete Lifecycle Management:** Native support for both ClickHouse (CHI) and ClickHouse Keeper (CHK), eliminating the ZooKeeper operational burden.

**4. Operational Intelligence:** Features like smart restarts, schema propagation, and non-disruptive storage resizing demonstrate deep, encoded domain expertise that manual approaches simply cannot match.

**5. Community Standard:** When other platforms (like Cozystack) choose to bundle the Altinity operator rather than build their own or use a generalist tool, that's a powerful market signal.

### The Abstraction Layer

Planton's API abstracts the operator's Helm chart complexity into approximately **six essential fields**:

| Field | Purpose |
|-------|---------|
| `namespace` | Where to install the operator |
| `version` | Operator version (maps to `image.tag`) |
| `namespace_scoped` | Multi-tenancy control (maps to `rbac.namespaceScoped`) |
| `resources` | CPU/Memory for operator pod |
| `service_account` | ServiceAccount configuration |
| `install_crds` | **[Value-Add]** Automates the two-step CRD installation |

The `install_crds` field is particularly valuable—it automates the complex "CRD-first, then-Helm" installation process that trips up many users.

## Conclusion: From Infrastructure to Intelligence

The evolution from manual StatefulSets to the Altinity Operator represents more than a technical upgrade—it's a fundamental shift in how we think about database operations.

StatefulSets give us infrastructure. The operator gives us **intelligence**.

Where StatefulSets say "here's a pod with stable storage," the operator says "here's a sharded, replicated ClickHouse cluster that automatically heals, scales, and upgrades itself while preserving data consistency and zero downtime."

That's not automation—that's operational expertise encoded as software. And for production ClickHouse on Kubernetes, that's exactly what you need.

With five years of production validation, tens of thousands of managed servers worldwide, and a virtuous-cycle open-source business model, the Altinity ClickHouse Operator isn't just the best choice for self-hosted ClickHouse on Kubernetes.

It's the **only choice** for teams that take production seriously.

