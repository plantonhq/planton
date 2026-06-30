# SigNoz on Kubernetes: From Anti-Patterns to Production

## Introduction

"Just deploy it with a Deployment and mount a hostPath volume—it's just a database, right?"

If you've deployed stateful observability platforms on Kubernetes, you've probably encountered this sentiment. And if you've encountered it in production, you've probably also encountered the 3 AM page that follows when the node reboots and all your telemetry data vanishes into the void.

SigNoz is a modern, OpenTelemetry-native observability platform that unifies logs, metrics, and traces into a single application. But beneath its elegant UI lies a complex, stateful architecture: a high-performance columnar database (ClickHouse), a distributed coordination service (Zookeeper), a telemetry processing pipeline (OpenTelemetry Collector), and the SigNoz application itself. Getting this stack production-ready on Kubernetes is not a trivial exercise.

This document explores the landscape of SigNoz deployment methods, from fundamentally flawed approaches to battle-tested production patterns. More importantly, it explains **why** Planton made the architectural choices it did when designing the `SignozKubernetes` API.

## The Deployment Maturity Spectrum

### Level 0: The Anti-Pattern — Raw Kubernetes Primitives

**What it looks like**: Deploying SigNoz components using basic Kubernetes Deployments, manually configured StatefulSets, or worse—Pods with `hostPath` volumes.

**Why it fails**:

The fundamental problem is that SigNoz is a **stateful, multi-component system** where component relationships and startup ordering matter. The two most critical components—ClickHouse and Zookeeper—are inherently stateful and require:

- **Stable, persistent identities**: Each ClickHouse pod needs a stable network identity and persistent volume that survives pod rescheduling.
- **Ordered, coordinated startup**: ClickHouse replicas must start in a specific sequence and register with Zookeeper before accepting writes.
- **Persistent storage**: Data must survive node failures, pod rescheduling, and cluster maintenance.

**The Deployment primitive is designed for stateless applications**. Pods are ephemeral and interchangeable. If you deploy ClickHouse as a Deployment:
- A rescheduled pod gets a new name and identity, severing its connection to persistent data
- No mechanism ensures replicas start in order or maintain quorum
- Data written to `emptyDir` or `hostPath` is lost on pod termination or node failure

**Manual StatefulSets** are technically correct but practically untenable. Even the official SigNoz Helm chart—which is professionally maintained—experiences frequent user-reported failures during upgrades due to immutable StatefulSet fields. Helm upgrade commands commonly fail with errors like:

```
Forbidden: updates to statefulset spec for fields other than 'replicas', 'template', 
'updateStrategy' are forbidden
```

If the official chart struggles with this complexity, reimplementing it manually or through lightweight abstractions is a recipe for production outages.

**Verdict**: This is not a deployment method; it's a ticking time bomb. Don't do this.

---

### Level 1: The Official Helm Chart — The De Facto Standard

**What it is**: The `signoz/signoz` Helm chart is the community-supported, officially documented way to deploy SigNoz on Kubernetes. It's hosted at [https://github.com/SigNoz/charts](https://github.com/SigNoz/charts) and published to the Helm repository at `https://charts.signoz.io`.

**What it provides**:

This is not just a collection of templated YAML files. The Helm chart is a **sophisticated management layer** that encapsulates the conditional logic and operational complexity of SigNoz deployment. It:

1. **Provisions all required components**: SigNoz backend (UI, API, Alertmanager), OpenTelemetry Collector, ClickHouse, and Zookeeper.
2. **Manages critical dependencies**: Automatically enables Zookeeper when you configure a distributed ClickHouse cluster (replication > 1).
3. **Bundles the Altinity ClickHouse Operator**: This is crucial. The chart doesn't create ClickHouse StatefulSets directly. Instead, it:
   - Deploys the battle-tested Altinity ClickHouse Operator as a dependency
   - Creates a `ClickHouseInstallation` (CHI) Custom Resource
   - Lets the Altinity Operator handle all StatefulSet creation, reconciliation, and lifecycle management

**Why it matters**:

The Helm chart is not a convenience wrapper—it's the **API** for deploying SigNoz. The ClickHouse configuration, for example, is abstracted through the CHI Custom Resource, which the Altinity Operator translates into StatefulSets, Services, ConfigMaps, and PodDisruptionBudgets.

**Limitations**:

While the Helm chart is production-ready, it's also complex:
- **Configuration burden**: The `values.yaml` file contains hundreds of fields. Understanding which 20% of fields matter for 80% of use cases requires deep domain knowledge.
- **Dependency pitfalls**: Even simple deployments have hidden dependencies. For example, SigNoz hardcodes the ClickHouse cluster name as `"cluster"`, which forces ClickHouse into cluster mode. This **mandates Zookeeper** even for single-node deployments—an unintuitive requirement that confuses users and adds unnecessary resource overhead in dev/test environments.
- **Manual lifecycle management**: Users must manually calculate appropriate replica counts, resource allocations, and disk sizes. Misconfiguration is common.

**Verdict**: This is the **baseline** for any production deployment. Any higher-level abstraction—including Planton's `SignozKubernetes`—should be built as a wrapper **over this chart**, not as a replacement for it. The chart is the source of truth.

---

### Level 2: Declarative Orchestration — GitOps and IaC

**What it looks like**: Using tools like ArgoCD, Flux, Terraform, or Pulumi to manage the Helm chart declaratively.

**How it works**:

These tools don't change the deployment method—they wrap the official Helm chart in a declarative management layer:

- **ArgoCD**: Deploys an `Application` manifest that points to `https://charts.signoz.io` and applies your custom `values.yaml`.
- **Flux**: Uses a `HelmRelease` Custom Resource to manage the chart's lifecycle.
- **Terraform/Pulumi**: Use their respective Helm providers to declare the chart installation as infrastructure-as-code.
- **Kustomize**: Applies overlays and patches to the Helm chart's rendered manifests (can be integrated via Helm's `--post-renderer` flag).

**Why it's valuable**:

- **GitOps**: Configuration is version-controlled, auditable, and recoverable. Changes are peer-reviewed.
- **Repeatability**: Promotes consistent deployments across dev, staging, and production.
- **Drift detection**: ArgoCD and Flux detect configuration drift and can auto-heal resources.
- **IaC integration**: Terraform and Pulumi allow SigNoz deployment to be orchestrated alongside other infrastructure (VPCs, DNS, secrets).

**Limitations**:

These tools solve the **orchestration problem**, not the **configuration problem**. They don't tell you how to size ClickHouse disk volumes, calculate Zookeeper quorum requirements, or avoid the gRPC/HTTP ingress conflict. You still need deep knowledge of the Helm chart to configure it correctly.

**Verdict**: Essential for production environments, but not a substitute for understanding the underlying deployment architecture. Best used in combination with a higher-level abstraction that provides opinionated defaults.

---

### Level 3: The Production Solution — Opinionated, Structured Abstractions

**What it is**: A high-level API that encapsulates deployment best practices, provides intelligent defaults, and automatically handles cross-component dependencies.

**How Planton's SignozKubernetes API embodies this**:

The `SignozKubernetes` API is not a simple pass-through to the Helm chart. It's a **structured abstraction** that implements the 80/20 principle: expose the 20% of configuration that 80% of users need, with intelligent logic to handle the remaining 80% of complexity.

**Key design principles**:

1. **Logical, not flat**: Instead of exposing hundreds of Helm values as a flat map, the API is structured into logical components (`signoz_container`, `otel_collector_container`, `database`, `ingress`). This makes it clear what each setting controls.

2. **Dependency automation**: If a user configures a distributed ClickHouse cluster (replicas > 1), the controller knows this **requires Zookeeper** and automatically:
   - Enables Zookeeper in the Helm chart
   - Sets `zookeeper.replicaCount: 3` for production quorum (not 1, which is a single point of failure)
   - Configures ClickHouse to use the Zookeeper service

3. **Profile-based defaults**: Instead of requiring users to specify every resource allocation, the API can support deployment profiles:
   - `dev`: Minimal single-node ClickHouse, single Zookeeper (if required), minimal resources
   - `staging`: HA-ready with 1 shard, 2 replicas, 3 Zookeeper nodes
   - `production`: Multi-shard with proper anti-affinity, PodDisruptionBudgets, and production-scale resources

4. **Ingress intelligence**: The API handles the dual-protocol ingress requirement for the OpenTelemetry Collector by generating **two separate Ingress resources** automatically—one for gRPC (port 4317, with the correct annotations) and one for HTTP (port 4318).

5. **Validation and safety**: The protobuf schema enforces constraints:
   - External ClickHouse configuration is required when `is_external: true`
   - Disk sizes must match Kubernetes quantity formats (e.g., `"100Gi"`)
   - Zookeeper replicas should be odd numbers for quorum
   - Clustering can't be enabled without specifying shard and replica counts

**Example: Solving the "Single-Node Zookeeper" Pitfall**:

Users frequently encounter this frustrating scenario:
- They want a simple dev deployment with a single-node ClickHouse
- They don't enable clustering in the Helm chart
- The deployment fails because SigNoz hardcodes `cluster_name: "cluster"`, forcing ClickHouse into cluster mode
- Cluster mode **requires Zookeeper**, even with 1 replica

The SignozKubernetes API's controller can implement logic to handle this automatically:
```
IF (database.managed_database.container.replicas == 1 AND 
    (!database.managed_database.cluster.is_enabled OR 
     database.managed_database.cluster.replica_count == 1)) THEN
  // Single-node mode
  Enable Zookeeper with replicaCount: 1
  Configure ClickHouse with layout.replicasCount: 1, layout.shardsCount: 1
ELSE IF (database.managed_database.cluster.is_enabled AND 
         database.managed_database.cluster.replica_count > 1) THEN
  // Distributed mode
  Enable Zookeeper with replicaCount: 3 (production quorum)
  Configure ClickHouse with layout.replicasCount and layout.shardsCount from spec
  Apply podDistribution: ReplicaAntiAffinity for true HA
END IF
```

This logic is invisible to the user but prevents a common misconfiguration.

**What it doesn't do**:

The SignozKubernetes API doesn't try to replace the Helm chart or reimplement its logic. Instead, it **generates the appropriate Helm values** and uses the chart as its deployment engine. This ensures:
- Compatibility with upstream updates
- Leverage of the bundled Altinity ClickHouse Operator
- Access to advanced features for power users (via `helm_values` pass-through)

**Verdict**: This is the **strategic approach** for teams that want production-ready observability without becoming SigNoz deployment experts. It balances simplicity for common use cases with flexibility for advanced needs.

---

## Understanding ClickHouse Deployment Patterns

ClickHouse is the heart of SigNoz—a high-performance columnar database that stores and queries telemetry data. How you deploy ClickHouse determines SigNoz's reliability, scalability, and operational complexity.

### Single-Node ClickHouse (Dev/Test)

**Configuration**: `shardsCount: 1`, `replicasCount: 1`

**Use case**: Local development, quick evaluation, CI/CD test environments.

**Characteristics**:
- Simplest configuration
- No high availability—a pod or node failure causes downtime and potential data loss
- Vertical scaling only (increase CPU/memory of the single pod)
- Still requires 1 Zookeeper pod due to SigNoz's hardcoded cluster configuration

**Resources**: 
- ClickHouse: 500m CPU, 1Gi memory, 20Gi disk (default, increase for longer testing)
- Zookeeper: 100m CPU, 256Mi memory, 5Gi disk

**Verdict**: Acceptable for non-production environments. Do not use in staging or production.

---

### Distributed ClickHouse Cluster (Staging/Production)

**Configuration**: 
- Sharding: `shardsCount > 1` (horizontal partitioning of data)
- Replication: `replicasCount > 1` (data redundancy within each shard)
- Example: 2 shards, 2 replicas = 4 ClickHouse pods total

**Use case**: Staging environments, production deployments.

**Why it's essential**:

- **Sharding** scales write throughput and query parallelism. As telemetry volume grows, adding shards distributes the load.
- **Replication** provides fault tolerance. If a ClickHouse pod fails, its replica within the shard continues serving queries and accepting writes.
- **Requires Zookeeper quorum**: ClickHouse's `ReplicatedMergeTree` table engine uses Zookeeper for replica coordination, leader election, and schema migrations. A 3-node Zookeeper ensemble is the **minimum** for production (provides fault tolerance for 1 Zookeeper failure).

**Pod anti-affinity is critical**: Deploying multiple replicas is pointless if they all run on the same node. The Altinity Operator supports `podDistribution` rules:
- `ClickHouseAntiAffinity`: Basic spreading of all ClickHouse pods
- `ReplicaAntiAffinity` **(recommended)**: Ensures replicas of the same shard never run on the same node

**Resources** (baseline for small production):
- ClickHouse: 2 CPU, 4Gi memory, 100Gi+ disk per pod
- Zookeeper: 500m CPU, 1Gi memory, 20Gi disk per pod (3 pods)

**Operational considerations**:
- **PodDisruptionBudgets**: Prevent voluntary disruptions (node drains) from taking down all replicas simultaneously
- **Volume expansion**: The StorageClass **must** support `allowVolumeExpansion: true` (many cloud defaults don't enable this). Otherwise, expanding disk size requires complex and risky data migrations.

**Verdict**: This is the **production standard**. The added complexity is justified by the resilience and scalability it provides.

---

### External ClickHouse

**Configuration**: Set `database.is_external: true` and provide connection details.

**Use cases**:
1. **Managed services**: Connect to Altinity.Cloud (BYOC managed ClickHouse)
2. **Centralized database**: Multiple SigNoz instances sharing a single ClickHouse cluster
3. **Organizational boundaries**: ClickHouse managed by a dedicated DBA team

**Critical incompatibility—ClickHouse Cloud**:

SigNoz is **fundamentally incompatible** with the official multi-tenant ClickHouse Cloud service. Here's why:

- SigNoz uses User-Defined Functions (UDFs) to calculate histogram quantiles for Prometheus-style metrics
- These are not SQL functions—they're **executable scripts** (located in `deploy/docker/clickhouse-setup/user_scripts/`)
- ClickHouse Cloud, for valid security and multi-tenancy reasons, **does not allow** users to install arbitrary executable UDFs
- Result: Basic tracing might work, but **metrics ingestion and querying will fail**

This is a hard blocker. If you want managed ClickHouse, use Altinity.Cloud (built by the authors of the Altinity ClickHouse Operator) or self-host externally.

**Operational complexity**:

External ClickHouse is "expert mode":
- You must manually create a distributed cluster named `"cluster"` (SigNoz hardcodes this name)
- You must manage schema migrations, user provisioning, and UDF installation
- No built-in operator management—you handle all ClickHouse lifecycle operations

**Verdict**: High complexity, high maintenance. Only viable for organizations with dedicated ClickHouse expertise or using Altinity.Cloud.

---

## The 80/20 Configuration Principle

The official SigNoz Helm chart's `values.yaml` contains hundreds of fields. But for 80% of deployments, users only need to configure about 20% of them.

### The Essential 20%

These are the fields that genuinely matter for deploying a functional, production-ready SigNoz cluster:

**Global**:
- `global.storageClass`: The only universally mandatory field. Defines the StorageClass for all PersistentVolumeClaims.

**ClickHouse**:
- `clickhouse.layout.shardsCount`: Scale write performance (default: 1, production: 2+)
- `clickhouse.layout.replicasCount`: Enable high availability (default: 1, production: 2)
- `clickhouse.persistence.size`: Disk volume size (default 20Gi is unusable in production; set 100Gi+ for real workloads)
- `clickhouse.resources`: CPU/memory allocations

**Zookeeper**:
- `zookeeper.enabled`: Auto-required when replication > 1
- `zookeeper.replicaCount`: Quorum size (production: 3, never 1)
- `zookeeper.persistence.size`: Disk for transaction logs (20Gi recommended)

**OpenTelemetry Collector**:
- `otelCollector.replicaCount`: Scale ingestion layer (2+ for production)
- `otelCollector.resources`: Prevent OOMKills during high ingestion load

**SigNoz Application**:
- `queryService.replicaCount`: Scale query/API layer (2+ for production)
- `frontend.replicaCount`: Scale UI (2 for redundancy)

**Ingress**:
- `frontend.ingress.enabled`, `frontend.ingress.hosts`, `frontend.ingress.tls`: Expose the UI
- `otelCollector.ingress.enabled`, `otelCollector.ingress.hosts`: Expose data ingestion endpoints
  - **Critical**: Must configure as two separate host entries (one for gRPC with `backend-protocol: GRPC` annotation, one for HTTP without it)

### The Advanced 80% (Skip These)

These fields exist for niche use cases and power users:
- `otelCollector.config`: Raw OpenTelemetry Collector YAML (use the OTel Operator separately if you need this)
- `image.repository` / `image.tag` overrides: Stick with chart-bundled versions
- `externalClickhouse.*`: High-maintenance expert mode
- `clickhouse.podDistribution`: Critical for production but complex to model—better abstracted by an opinionated controller

---

## Why Planton Chose This Approach

The `SignozKubernetes` API design was informed by these key insights from the deployment landscape research:

### 1. The Helm Chart is the API

We don't reimplement ClickHouse deployment logic. We wrap the official Helm chart, which means:
- We benefit from the bundled Altinity ClickHouse Operator
- We stay compatible with upstream improvements
- Power users can access advanced features via `helm_values` pass-through

### 2. Intelligent Dependency Management

The controller enforces correct dependency relationships:
- Distributed ClickHouse **requires** Zookeeper (and automatically sets quorum to 3)
- Single-node deployments still enable Zookeeper (required due to hardcoded cluster name) but use 1 replica
- External ClickHouse skips in-cluster database deployment entirely

### 3. The 80/20 Principle

The protobuf spec exposes only the essential fields that most users actually configure:
- Component replicas and resources
- ClickHouse clustering (shards, replicas, disk size)
- Zookeeper quorum size
- Ingress endpoints

Everything else has smart defaults or is hidden behind the optional `helm_values` map for advanced users.

### 4. Dual-Ingress Pattern for OpenTelemetry Collector

We handle the gRPC/HTTP protocol conflict automatically. When a user enables ingress for the OTel Collector, the controller generates **two Ingress resources**:
- One for gRPC (port 4317, `backend-protocol: GRPC` annotation)
- One for HTTP (port 4318, no special annotation)

This is transparent to the user but prevents a common misconfiguration that breaks telemetry ingestion.

### 5. Validation and Safety

The protobuf schema prevents common mistakes:
- Can't enable external ClickHouse without providing connection details
- Disk sizes must match Kubernetes quantity formats
- Clustering requires explicit shard/replica counts
- Ingress requires hostnames when enabled

### 6. Production-Ready Defaults

The API's default values are designed for production readiness:
- OpenTelemetry Collector: 2 replicas (load balancing and redundancy)
- ClickHouse: Reasonable resource allocations (1 CPU, 2Gi memory)
- Zookeeper: Sufficient disk for transaction logs (8Gi)

---

## Common Pitfalls and How Planton Avoids Them

Based on community pain points and production outages, here are the most common SigNoz deployment failures:

### 1. Resource Under-provisioning
**Problem**: Default ClickHouse resources (500m CPU, 1Gi memory) cause OOMKills and query timeouts under real load.

**Solution**: SignozKubernetes API defaults to production-appropriate resources (2 CPU, 4Gi memory limits for ClickHouse).

### 2. Disk Size Miscalculation
**Problem**: Default 20Gi ClickHouse disk fills in hours or days with production telemetry volume.

**Solution**: API documentation clearly states 100Gi+ for production, and validation warns if values seem too small.

### 3. Single-Node Zookeeper in Production
**Problem**: Using 1 Zookeeper replica creates a single point of failure that negates ClickHouse HA.

**Solution**: Controller logic automatically sets `zookeeper.replicaCount: 3` when ClickHouse replication is enabled.

### 4. The gRPC/HTTP Ingress Conflict
**Problem**: Using a single Ingress resource with `backend-protocol: GRPC` annotation breaks HTTP ingestion.

**Solution**: Controller generates two separate Ingress resources automatically.

### 5. Non-Expandable Storage
**Problem**: Choosing a StorageClass without `allowVolumeExpansion: true` forces complex data migrations when disk fills.

**Solution**: API documentation emphasizes this requirement; future controller versions could validate StorageClass capabilities.

### 6. ClickHouse Cloud Incompatibility
**Problem**: Users discover too late that SigNoz metrics don't work on ClickHouse Cloud (UDF restriction).

**Solution**: Documentation clearly warns about this incompatibility; external ClickHouse setup guides recommend Altinity.Cloud.

### 7. Immutable StatefulSet Field Errors
**Problem**: Helm upgrades fail with "forbidden field" errors when users manually edit chart resources.

**Solution**: SignozKubernetes API is declarative—controller generates correct Helm values, reducing manual chart manipulation.

---

## Conclusion

Deploying SigNoz on Kubernetes is a spectrum from "catastrophically wrong" to "battle-tested and production-ready." The key insights:

- **Never use basic Deployments** for stateful components—it's a guaranteed failure.
- **The official Helm chart is the standard**—any abstraction should wrap it, not replace it.
- **ClickHouse architecture determines everything**—single-node for dev, distributed with replication for production.
- **Dependencies matter**—distributed ClickHouse requires Zookeeper, which requires quorum (3+ nodes).
- **The 80/20 principle is real**—most users only need to configure 20% of available fields.

Planton's `SignozKubernetes` API embodies these lessons. It's not a lowest-common-denominator abstraction—it's an opinionated, intelligent layer that handles the complex 80% so you can focus on the essential 20%. It enforces best practices, automates dependency management, and provides production-ready defaults while still allowing power users to customize advanced features.

The result: You can deploy a production-grade, OpenTelemetry-native observability platform without becoming a ClickHouse operator expert. And when you inevitably need to scale or troubleshoot, the underlying Helm chart and Altinity ClickHouse Operator are still there, well-understood and well-documented.

That's the kind of abstraction worth building.

