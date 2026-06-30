# Deploying Apache Solr on Kubernetes: From Anti-Patterns to Production

## Introduction: The Journey to Production-Ready Solr

For years, the conventional wisdom around deploying stateful applications like Apache Solr on Kubernetes was simple: "Don't." The infrastructure-as-code world treated Kubernetes as a platform designed for ephemeral, stateless workloads—the kind that could be terminated and replaced without consequence. SolrCloud, with its requirements for stable network identities, persistent storage, and careful coordination through Zookeeper, seemed fundamentally at odds with Kubernetes's cattle-not-pets philosophy.

That wisdom is now obsolete.

The maturation of the Kubernetes Operator pattern has transformed how we deploy and manage stateful systems. For Solr specifically, the **Apache Solr Operator** represents a production-ready, application-aware control plane that not only makes Solr deployments _possible_ on Kubernetes but actually _superior_ to traditional approaches. The Operator understands Solr's coordination semantics, manages safe rolling updates based on shard replica health, and provides declarative APIs for backups and monitoring—capabilities that manual StatefulSet management could never achieve safely.

This document explains the evolution of Solr deployment methods on Kubernetes, from critical anti-patterns to production-ready operator-based solutions. It maps the landscape of deployment approaches, compares the major tools (Apache Solr Operator vs. Bitnami Helm charts), and explains why Planton builds upon the Operator pattern to provide both Day 1 simplicity and Day 2 operational safety.

## The Deployment Maturity Spectrum

Understanding where different deployment methods sit on the maturity spectrum is essential for making informed architectural decisions.

### Level 0: The Anti-Pattern—Deployments and Simple Pods

Using a standard Kubernetes Deployment object for Apache Solr is not just suboptimal—it's a guaranteed path to cluster instability and data loss.

Apache Solr is an inherently stateful application. It stores persistent, indexed data on disk. A Kubernetes Deployment is designed for stateless applications that can be terminated and replaced with a new, identical pod with a new name and no persistent state. This fundamental mismatch creates two critical failure modes:

1. **Lack of Stable Network Identity**: SolrCloud nodes must have stable, predictable network identities (e.g., `solr-0`, `solr-1`) to register with the Zookeeper ensemble, participate in leader election, and manage shard replicas. A Deployment-managed pod receives a new, random hostname on every restart. This breaks cluster coordination—Zookeeper loses track of nodes, leading to inconsistent cluster state and eventual service-wide outages.

2. **Lack of Stable, Unique Storage**: Each Solr node must have its own unique persistent volume (PV) to store its index data. A Deployment typically uses ephemeral storage or a shared `ReadWriteMany` (RWX) volume. The correct pattern requires a unique PersistentVolumeClaim (PVC) for each pod ordinal (e.g., `data-solr-0`, `data-solr-1`).

**Verdict**: Never use Deployments for Solr. This approach is a critical anti-pattern that guarantees cluster instability upon pod restarts or node failures.

### Level 1: The Foundation—Manual StatefulSet Configuration

The StatefulSet is the correct Kubernetes primitive for stateful applications. It provides the two guarantees that Deployments lack:

1. **Stable Network Identity**: Pods are named with a stable ordinal (e.g., `solr-0`, `solr-1`).
2. **Stable, Unique Storage**: Each pod gets its own unique PersistentVolumeClaim based on a template.

A StatefulSet is always paired with a Headless Service, which provides the stable DNS domain (e.g., `solr-0.solr-headless.default.svc.cluster.local`) that pods use for discovery.

While this is the correct _low-level_ approach, manually managing StatefulSet YAMLs for a complex, multi-component system like SolrCloud introduces severe operational pitfalls:

- **Zookeeper Bootstrap Complexity**: A Solr StatefulSet must not start its pods before the Zookeeper ensemble is fully ready. This requires complex, cross-StatefulSet coordination, typically managed with initContainers that poll the Zookeeper service.

- **Pod Configuration Pain**: Manually managing the Zookeeper StatefulSet itself is challenging. Each ZK pod must have a unique `ZOO_MY_ID` environment variable, which is non-trivial to inject correctly based on the pod's ordinal.

- **Application-Unaware Updates**: This is the most significant pitfall. A StatefulSet's built-in update strategies (`OnDelete` or `RollingUpdate`) are _Kubernetes-aware_, not _Solr-aware_. The default `RollingUpdate` strategy terminates pods from the highest ordinal to the lowest (e.g., `solr-N` down to `solr-0`). This "dumb" update can be catastrophic—it might kill a shard leader _before_ its replica is healthy, leading to brief but critical client-side write errors.

**Verdict**: StatefulSets are necessary but insufficient. They solve the fundamental stability problem but introduce complex operational challenges that are unsafe for production without sophisticated orchestration logic.

### Level 2: Packaged Deployments—Helm Charts

Helm charts, such as the popular Bitnami chart, represent the next evolutionary step. They "package" the complex set of StatefulSet, Service, ConfigMap, and PodDisruptionBudget YAMLs into a single, configurable `helm install` command.

This approach solves the _initial creation_ complexity (Day 1) but does not solve the _operational_ complexity (Day 2+). The chart stamps out the resources, but the cluster's lifecycle is still managed by Kubernetes's default StatefulSet controller. It still lacks the _application-aware_ logic to handle updates, failures, and backups safely.

#### The Bitnami Solr Chart: Standalone Simplicity

The Bitnami chart is a traditional, self-contained Helm chart. It directly creates StatefulSet, Service, and other Kubernetes primitives. It can optionally bundle the `bitnami/zookeeper` Helm chart as a dependency, allowing a single `helm install` to deploy a complete cluster including Zookeeper.

**Strengths**: 
- Simple Day 1 experience
- Single `values.yaml` file for all configuration
- Works "out of the box" for initial deployments

**Weaknesses**:
- No application-aware updates (uses standard StatefulSet rolling updates)
- Backup and restore are entirely manual API calls, not first-class features
- Scaling operations lack Solr-specific safety checks

**Verdict**: Good for getting started quickly, but lacks the operational safeguards required for production Day 2 operations.

### Level 3: The Production Solution—Apache Solr Operator

The Operator pattern is the cloud-native best practice for stateful applications. An operator is a custom controller that runs in the cluster and _extends_ the Kubernetes API with domain-specific knowledge of the application.

The **Apache Solr Operator** is the official, production-ready implementation of this pattern. It was started by Bloomberg and is now an official Apache Solr project. It is explicitly documented as "Production Ready" and has been "successfully used to manage production SolrClouds for some of the largest users of Solr."

#### How the Operator Changes Everything

Instead of managing a StatefulSet directly, users create a `SolrCloud` Custom Resource (CR). The Operator's controller watches for this CR and, in response, generates the necessary StatefulSet, Services, ConfigMaps, and other resources.

Crucially, the Operator's control loop continues to watch the cluster after creation. If a user changes the `spec.solrImage.tag` in the SolrCloud CR to trigger a version upgrade, the Operator begins a rolling update. However, instead of using the default StatefulSet logic, it intelligently queries the SolrCloud Collections API. It _only_ takes down a pod when it can verify that it is safe for that pod's shards and replicas, preventing data unavailability.

#### Key Operator Features

The Operator's power comes from its Custom Resource Definitions (CRDs), which provide declarative APIs for the entire Solr ecosystem:

- **SolrCloud CRD**: Manages the Solr cluster, including image, replicas, resources, storage, and Zookeeper connection configuration.

- **SolrBackup CRD**: Declaratively manages backups of collections. Supports Volume, S3, and GCS repositories, as well as scheduled, recurring backups. This transforms what would be manual `curl` commands into auditable, version-controlled YAML resources.

- **SolrPrometheusExporter CRD**: Declaratively manages the monitoring exporter for each cluster, enabling seamless integration with Prometheus and Grafana dashboards.

#### Zookeeper Integration: The Decoupled Architecture

The Solr Operator follows a best-practice, decoupled architecture for Zookeeper management. Its Helm chart recommends and depends on the **Zookeeper Operator**. The user's SolrCloud CR creates a `ZookeeperCluster` CR, delegating ZK management to a dedicated, specialized controller. This separation of concerns means Zookeeper's lifecycle can be managed independently, which is essential for large, multi-tenant environments where multiple applications (Solr, Kafka, etc.) share a single ZK ensemble.

**Verdict**: This is the production-ready standard. The Operator provides application-aware rolling updates, declarative Day 2 operations (backup, monitoring), official Apache support, and a robust CRD-based API. It is the recommended approach for any production deployment.

## Comparative Analysis: Apache Solr Operator vs. Bitnami Helm Chart

The choice between deployment tools boils down to a "Managed" (Operator) versus "Packaged" (Helm) philosophy.

| Feature               | Apache Solr Operator (Recommended)                                                  | Bitnami Helm Chart (Alternative)                                    |
| :-------------------- | :---------------------------------------------------------------------------------- | :------------------------------------------------------------------ |
| **Primary Method**    | **Operator (Control Loop)**: Deploys a controller that manages Solr-aware logic   | **Helm Chart (Templating)**: Deploys StatefulSet YAMLs directly   |
| **Official Status**   | **Official** Apache project                                                       | **Community** chart by Bitnami (VMware/Broadcom)                   |
| **Zookeeper**         | **Decoupled**: Manages ZK via the Zookeeper Operator CRD                          | **Bundled**: Manages ZK as a Helm sub-chart dependency            |
| **Scaling & Updates** | **Application-Aware**: Safely drains and manages shard replicas                   | **Kubernetes-Aware**: Standard StatefulSet rolling update         |
| **Backup**            | **Declarative**: First-class SolrBackup CRD                                        | **Manual**: Requires manual `curl` calls to Solr's API             |
| **Monitoring**        | **Declarative**: First-class SolrPrometheusExporter CRD                            | **Enabled via Flag**: `metrics.enabled=true` in values.yaml       |
| **Production Ready**  | **Yes**, explicitly stated and battle-tested                                       | **Yes**, but lacks operational safeguards                          |
| **Licensing**         | Apache License 2.0                                                                | Apache License 2.0                                                 |

### The Day 1 vs. Day 2 Trade-off

Real-world experience highlights a critical distinction: the Bitnami chart is _easier_ to get started with (Day 1), while the Operator is _safer_ and _more powerful_ to operate long-term (Day 2+).

Users optimizing solely for Day 1 simplicity might choose Bitnami. A platform engineering team building production infrastructure must optimize for Day 2 operations—stability, safety, and automated lifecycle management. This is why Planton builds upon the Operator model.

## Production Zookeeper Integration

Zookeeper is the coordination service for SolrCloud. Its stability directly determines the stability of the entire cluster. Production deployments have strict, non-negotiable requirements.

### Embedded vs. External Zookeeper

Apache Solr ships with an _embedded_ Zookeeper server. This feature is provided _only_ for development, testing, or getting started.

Using embedded Zookeeper in production is a critical failure point. The embedded ZK instance runs inside a single Solr JVM. If that Solr pod or node fails, the _entire cluster's coordination service_ fails with it, resulting in a total service outage.

**Production Requirement**: All production SolrCloud deployments **must** use an external Zookeeper ensemble—a separate set of pods, managed by a separate StatefulSet, running only Zookeeper.

### Ensemble Sizing: 3 Nodes vs. 5 Nodes

Zookeeper's high availability is based on a "quorum" mechanism. For the ZK service to be active, a majority of its nodes must be running and able to communicate. The guiding principle is to deploy $2F+1$ nodes to tolerate $F$ failures.

This "majority" rule is why an odd number of nodes is essential:

- **2 Nodes**: Quorum is 2. If 1 node fails, 1 remains. 1 is not a majority of 2. **Fault tolerance: 0**.
- **3 Nodes**: Quorum is 2. If 1 node fails, 2 remain. 2 is a majority of 3. **Fault tolerance: 1**.
- **4 Nodes**: Quorum is 3. If 1 node fails, 3 remain. 3 is a majority of 4. If 2 nodes fail, 2 remain. 2 is not a majority of 4. **Fault tolerance: 1**.
- **5 Nodes**: Quorum is 3. If 2 nodes fail, 3 remain. 3 is a majority of 5. **Fault tolerance: 2**.

A 4-node cluster provides no more fault tolerance than a 3-node cluster but adds coordination overhead.

**Recommendations**:
- **3 Nodes**: Standard for most production clusters. Tolerates a single node failure (e.g., a pod crash or node drain).
- **5 Nodes**: Recommended for very large or mission-critical clusters. Provides higher fault tolerance, allowing the cluster to survive one _planned_ maintenance event (node drain) and _simultaneously_ tolerate one _unplanned_ failure (pod crash).

Adding more ZK nodes _decreases_ write performance (as more nodes must coordinate) and only _slightly increases_ read performance. Since ZK handles low-volume, high-importance coordination (not high-throughput data), the choice is driven by high-availability requirements rather than performance.

### Deployment Mechanics

There are two primary methods to deploy the external ZK ensemble:

1. **Zookeeper Operator (Solr Operator's Preferred Method)**:
   - The `apache-solr/solr-operator` Helm chart lists the Zookeeper Operator as a dependency.
   - When deploying a SolrCloud CR, users configure the `spec.zookeeperRef.provided` block.
   - The Solr Operator creates a `ZookeeperCluster` CR, which the Zookeeper Operator then reconciles into the necessary StatefulSet, Services, and ConfigMaps.
   - **Benefit**: Fully declarative, ecosystem-native approach where ZK lifecycle is managed as part of the Solr deployment.

2. **Standalone Helm (Independent/Shared Method)**:
   - Deploy Zookeeper independently using a standalone Helm chart (e.g., `helm install my-zk bitnami/zookeeper --set replicaCount=3`).
   - Configure the SolrCloud CR to point to the existing ensemble using `spec.zookeeperRef.connectionInfo.externalConnectionString`.
   - **Benefit**: Correct pattern when ZK's lifecycle must be independent of any single Solr cluster—common for large, multi-tenant ZK ensembles shared by many applications (Solr, Kafka, etc.).

## The Planton Approach

Planton's `SolrKubernetes` API is designed to provide the Day 1 simplicity of a Helm chart with the Day 2 power and safety of the Apache Solr Operator.

### The Operator-as-a-Service Model

The `SolrKubernetes` resource is a high-level abstraction that generates and manages the underlying `SolrCloud` and `ZookeeperCluster` Custom Resources. This approach:

- **Hides Complexity**: Users interact with a simplified API surface (10-12 core fields) rather than the hundreds of knobs exposed by the underlying CRDs.
- **Preserves Power**: All the Operator's application-aware logic—safe rolling updates, declarative backups, monitoring—remains intact.
- **Ensures Safety**: The API prevents users from misconfiguring critical settings (e.g., disabling the Operator's managed update strategy) that could compromise cluster stability.

### The 80/20 Configuration Philosophy

Just as Planton's APIs focus on the 20% of configuration that 80% of users need, the `SolrKubernetes` API exposes only the essential fields:

**Essential Solr Configuration**:
- `replicas`: The number of Solr nodes (maps to `spec.replicas` in SolrCloud CR)
- `image`: Container image repository and tag
- `resources`: Standard Kubernetes CPU/memory requests and limits
- `disk_size`: Persistent storage size (production requires persistent storage)
- `java_mem`: Simplified JVM heap configuration (e.g., "8g")

**Essential Zookeeper Configuration**:
- `replicas`: Number of ZK nodes (default: 3, production minimum for HA)
- `resources`: CPU/memory allocations
- `disk_size`: Persistent storage for ZK data

**Ingress Configuration**:
- `ingress`: Standard Kubernetes ingress specification for external access

### What's Not Exposed (By Design)

The v1 API intentionally excludes "expert-only" fields that could lead to misconfigurations:

- Custom Solr modules and additional libraries (plugin loading)
- Update strategy overrides (the Operator's default "Managed" strategy is optimal)
- Advanced GC tuning parameters (basic heap sizing covers most use cases)
- Low-level Kubernetes customization escape hatches

These omissions are deliberate. The goal is to make it nearly impossible for users to create unsafe configurations while still supporting the vast majority of production use cases.

### Reference Configurations

These examples demonstrate how the simplified API supports different deployment scales:

**Dev Cluster (Minimal Cost HA)**:
```yaml
solr_container:
  replicas: 1
  resources:
    requests:
      cpu: "1"
      memory: "2Gi"
  disk_size: "20Gi"
  image:
    repo: "apache/solr"
    tag: "9.4.0"
config:
  java_mem: "-Xms2g -Xmx2g"
zookeeper_container:
  replicas: 3  # ZK HA is still required
  resources:
    requests:
      cpu: "100m"
      memory: "512Mi"
  disk_size: "2Gi"
```

**Staging Cluster (Scaled-Down Prod)**:
```yaml
solr_container:
  replicas: 3
  resources:
    requests:
      cpu: "2"
      memory: "10Gi"
    limits:
      memory: "10Gi"
  disk_size: "100Gi"
  image:
    repo: "apache/solr"
    tag: "9.4.0"
config:
  java_mem: "-Xms8g -Xmx8g"
zookeeper_container:
  replicas: 3
  resources:
    requests:
      cpu: "250m"
      memory: "1Gi"
  disk_size: "10Gi"
```

**Production Cluster (HA & Resilient)**:
```yaml
solr_container:
  replicas: 5
  resources:
    requests:
      cpu: "4"
      memory: "24Gi"  # Higher than heap for OS file system cache
    limits:
      memory: "24Gi"
  disk_size: "500Gi"
  image:
    repo: "apache/solr"
    tag: "9.4.0"
config:
  java_mem: "-Xms16g -Xmx16g"
zookeeper_container:
  replicas: 5  # Higher HA, tolerating F=2 failures
  resources:
    requests:
      cpu: "500m"
      memory: "2Gi"
  disk_size: "20Gi"
```

Note: In the production configuration, total pod memory is set higher than the JVM heap to leave room for the OS file system cache, which Solr relies on heavily for performance.

## Day 2 Operations

A production-ready platform must account for ongoing operational needs beyond initial deployment.

### High Availability and Scaling

High availability in this architecture is achieved through three layers:

1. **Coordination HA**: A 3-node or 5-node Zookeeper ensemble ensures the cluster's "brain" is fault-tolerant.
2. **Node Resilience**: StatefulSet + PVCs ensure that if a Solr pod fails, Kubernetes reschedules it and reconnects it to its existing persistent volume.
3. **Data Replication**: Within SolrCloud, collections must be configured with `replicationFactor=2` or higher, ensuring that if one Solr node (holding a replica) dies, another node still holds a copy of that shard.

The Solr Operator provides intelligent, application-aware scaling. When you increase `replicas`, a new node is added. When you _decrease_ replicas, the Operator first tries to safely _move shards off_ that node before decommissioning it. This Solr-specific logic is far superior to a blind `kubectl scale` command, which could lead to data loss or unavailability.

### Disaster Recovery: Backups and Restore

**Backup (Declarative)**: The Operator provides the `SolrBackup` CRD, making backups a declarative operation. Users define a repository (Volume, GCS, or S3) in the SolrCloud CR's `spec.backupRepositories`, then create a `SolrBackup` resource. The CRD also supports scheduled, recurring backups via the `spec.recurrence` field.

**Restore (Manual API Call)**: There is currently **no `SolrRestore` CRD**. The Operator manages backups but not restores. The restore operation is a standard Solr Collections API call (`action=RESTORE`) that must be triggered manually:

```bash
curl 'http://<solr-service>/solr/admin/collections?action=RESTORE&name=<backup-name>&collection=<new-collection-name>&repository=<repo-name>'
```

This is a known limitation of the current Operator ecosystem, and users must be aware of the manual restore process.

### Monitoring and Alerting

The Operator provides a declarative `SolrPrometheusExporter` CRD. Users create this resource pointing to their SolrCloud instance, and the Operator deploys the exporter. The cluster's Prometheus instance scrapes the `/metrics` endpoint, and users can import the official Solr Grafana dashboard for instant visibility.

**Key Metrics to Alert On**:
- `solr_ping`: Cluster liveness
- **Query Latency (p95, p99)**: Most important user-facing metric
- **Cache Hit Ratio**: (e.g., `filterCache`, `queryResultCache`) Low hit ratios indicate performance tuning issues
- **JVM Heap Usage & GC Pauses**: High heap usage or long GC pauses are precursors to OutOfMemoryErrors

### Security Hardening

Production clusters must not be open to the public. The Operator has first-class support for:

- **TLS (Encryption)**: Via `SolrCloud.spec.solrTLS`, with support for auto-generated self-signed certificates, cert-manager integration, or user-provided Kubernetes Secrets.
- **Authentication**: Basic Authentication can be bootstrapped via `SolrCloud.spec.solrSecurity.authenticationType: Basic`, storing users and hashed passwords in a `security.json` file.
- **Authorization**: Rule-based authorization restricts access to collections or APIs.

### Networking: Ingress and External Access

Solr clients (e.g., web applications) often run outside the Kubernetes cluster. The standard cloud-native way to expose HTTP services is through an Ingress resource—a "smart" L7 reverse proxy that allows multiple services to share a single load balancer using hostnames or URL paths.

The Solr Operator has native support for this via `spec.solrAddressability.external.method: Ingress`. This integrates cleanly with ingress controllers (e.g., ingress-nginx) and cert-manager for automated TLS certificate provisioning.

**TLS Termination Patterns**:
- **End-to-End Encryption**: Solr pods have TLS enabled, and the Ingress passes through encrypted traffic (most secure, more complex).
- **TLS Termination at Ingress**: The Ingress handles the public TLS certificate, decrypts traffic, and forwards it to Solr pods as unencrypted HTTP (common, practical pattern).

For automated TLS, the Operator supports passing cert-manager annotations through to the managed Ingress resource, enabling fully automated certificate provisioning.

## Conclusion: Production-Ready Solr on Kubernetes

The journey from anti-patterns to production-ready Solr deployments on Kubernetes reflects the broader maturation of the cloud-native ecosystem. What was once considered risky or infeasible is now not only possible but represents best-in-class infrastructure.

The Apache Solr Operator, as an official Apache project backed by production deployments at some of the largest Solr users, embodies this maturation. It provides the application-aware intelligence necessary to safely manage stateful systems at scale. By combining stable network identities, persistent storage, and sophisticated rolling update logic that understands Solr's coordination semantics, the Operator transforms Kubernetes from a challenging platform for Solr into an ideal one.

Planton's `SolrKubernetes` API builds upon this foundation, providing a simplified abstraction that makes the Operator's power accessible without exposing its complexity. This "Operator-as-a-Service" model delivers both the Day 1 simplicity that developers want and the Day 2 operational safety that production environments demand.

The result is a deployment approach that is simultaneously simple, powerful, and—most importantly—production-safe.

