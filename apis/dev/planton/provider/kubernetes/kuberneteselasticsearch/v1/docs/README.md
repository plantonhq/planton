# Elasticsearch on Kubernetes: The Operator Revolution

## The Paradigm Shift

For years, running stateful distributed systems like Elasticsearch on Kubernetes was considered an anti-pattern. The conventional wisdom was simple: "Kubernetes is for stateless applications. Run your databases somewhere else."

That wisdom is now obsolete.

The maturity of the **Operator pattern** has fundamentally changed what's possible. Today, deploying production-grade Elasticsearch clusters on Kubernetes is not just viable—it's often the superior choice. But this transformation didn't happen overnight, and understanding the evolution from basic primitives to autonomous operators is essential for making informed architectural decisions.

This document explains the deployment landscape, why operators are the only production-ready solution, and how Planton abstracts the complexity while giving you the power to choose between Elasticsearch and its open-source fork, OpenSearch.

## The Maturity Spectrum: From Anti-Patterns to Autonomy

### Level 0: The Anti-Pattern (Kubernetes Deployments)

A Kubernetes Deployment is designed for stateless workloads—applications where every pod is identical and interchangeable. Attempting to run Elasticsearch data nodes with a Deployment is a guaranteed failure for two fundamental reasons:

**1. Ephemeral Identity**: Pods in a Deployment have random, disposable identities. When a pod dies, it's replaced by a *new* pod with a *new* name. Elasticsearch nodes require stable, unique network identities to participate in cluster formation, discovery, and quorum elections.

**2. Shared Storage Model**: Deployments are designed for pods to either share a single PersistentVolumeClaim (PVC) or have no persistent storage at all. Elasticsearch requires that *each node* have its *own dedicated* persistent data directory.

The practical failure mode is immediate and obvious: if multiple Elasticsearch pods in a Deployment share a PVC, the first pod to start acquires the `node.lock` file lock in the data path. Every subsequent pod that attempts to start will fail with a "failed to obtain node locks" error. This demonstrates the fundamental incompatibility between Deployment's "shared-everything" model and Elasticsearch's "shared-nothing" storage architecture.

**Verdict**: Never use Deployments for Elasticsearch. This is a non-starter.

### Level 1: The Foundation (StatefulSets)

StatefulSets were created specifically to address the shortcomings of Deployments for stateful applications. They provide three crucial features:

**1. Stable, Unique Network Identity**: Pods are created with predictable, stable hostnames (e.g., `es-cluster-0`, `es-cluster-1`). This identity is maintained even if the pod is rescheduled to a different Kubernetes node, enabling reliable configuration of `discovery.seed_hosts` and `cluster.initial_master_nodes`.

**2. Stable, Persistent Storage**: StatefulSets use `volumeClaimTemplates` to create a *unique* PersistentVolumeClaim for *each* pod replica. When a pod fails, Kubernetes creates a new pod with the *exact same identity* and re-attaches it to its *original* PVC, ensuring data and state are preserved.

**3. Ordered Deployment and Scaling**: Pods are created, updated, and deleted in strict sequential order, preventing a "thundering herd" of nodes all attempting to join the cluster simultaneously.

**What StatefulSets Don't Solve**:

While StatefulSets are *necessary*, they're not *sufficient*. A StatefulSet is application-agnostic—it knows how to manage pods with stable state, but has zero knowledge of how to manage an *Elasticsearch cluster*. This creates a significant "orchestration gap" for Day 2 operations:

- **Complex Bootstrapping**: You must manually configure `elasticsearch.yml` (via ConfigMap) with correct discovery settings. Error-prone and manual.
- **Unsafe Rolling Upgrades**: StatefulSets perform "blind" updates without checking Elasticsearch cluster health or performing shard draining. This can cause service outages for indices with zero replicas.
- **Configuration Changes**: StatefulSets don't automatically roll pods when ConfigMaps are updated.
- **Security Management**: No automatic generation of TLS certificates or credential bootstrapping.
- **Complex Topology Management**: Scaling is "dumb" and doesn't safely decommission nodes at the application level.

**Verdict**: StatefulSets solve the "pod and storage" problem but leave the much harder "distributed database lifecycle" problem to humans.

### Level 2: The "Day 1" Solution (Helm Charts)

Helm charts act as package managers for Kubernetes, templating the complex collection of YAML manifests (StatefulSets, ConfigMaps, Services) required to deploy an application.

The landscape here has been dramatically simplified by Elastic itself: the official `elastic/elasticsearch` Helm chart is **deprecated**. Elastic's documentation explicitly states that the vendor-recommended way to run the Elastic Stack on Kubernetes is with the **Elastic Cloud on Kubernetes (ECK) Operator**. The chart is being handed over to the community and will be archived.

This leaves `bitnami/elasticsearch` as the de facto community-maintained Helm chart. However, Helm charts suffer from a fundamental limitation: Helm is a "package manager," not an "operator." It provides a "Day 1" (install/template) solution but has no "Day 2" (active management) capabilities.

Because a Helm chart ultimately just templates Kubernetes YAML, any deployment it creates *inherits all the "Day 2" limitations of a manual StatefulSet*:

- `helm upgrade` is just a "dumb" resource patch—it cannot perform application-aware orchestration (check cluster health, drain shards) required for safe, zero-downtime upgrades.
- Users have historically reported that the Elastic chart was "very limited, lots of problems," which is precisely why Elastic now pushes users to the operator.
- Even Elastic's *new* Helm charts (`elastic/eck-operator` and `elastic/eck-stack`) don't deploy Elasticsearch directly—they exist *only* to install the ECK Operator and its Custom Resource Definitions.

This is a clear and unambiguous signal from the vendor: **the Operator pattern is the only fully supported solution for production deployments.**

**Verdict**: Helm charts are useful for Day 1 deployments but are fundamentally inadequate for production Day 2 operations.

### Level 3: The Production Solution (Kubernetes Operators)

The Operator pattern is the definitive solution for running stateful applications on Kubernetes. An Operator is a custom Kubernetes controller that embeds human operational knowledge into software. It extends the Kubernetes API by introducing Custom Resource Definitions (CRDs), such as `kind: Elasticsearch` or `kind: OpenSearchCluster`.

This fundamentally changes the paradigm. You no longer manage low-level StatefulSets or ConfigMaps. Instead, you manage a high-level, declarative Elasticsearch object. You declare your *intent* (e.g., "I want a 3-node HA cluster on version 8.19.6"), and the Operator's *active reconciliation loop* works continuously to make that intent a reality.

**The Operator model solves all the gaps left by StatefulSets and Helm**:

**1. Automated Bootstrapping**: The Operator automatically generates the underlying StatefulSets and injects the correct discovery configuration into each pod, ensuring a healthy cluster forms automatically.

**2. Safe, Application-Aware Upgrades**: When you change the `version` field in the CRD, the Operator initiates an application-aware rolling upgrade:
   1. Check cluster health
   2. Disable shard allocation
   3. Upgrade one pod
   4. Wait for the new node to rejoin and cluster health to return to green
   5. Re-enable shard allocation
   6. Repeat for the next pod

This procedure ensures zero-downtime upgrades for high-availability clusters.

**3. Security by Default**: Production-grade operators automatically generate all required TLS certificates for transport and HTTP layers and bootstrap the `elastic` superuser password, storing it securely in a Kubernetes Secret.

**4. Declarative Topology Management**: You can define complex hot-warm-cold topologies declaratively in the CRD using a `nodeSets` (ECK) or `nodePools` (OpenSearch) abstraction. The Operator intelligently manages creation and deletion of multiple underlying StatefulSets and automatically configures the necessary shard allocation awareness rules.

**Verdict**: Operators are the only production-ready solution. They provide both Day 1 automation and Day 2 lifecycle management.

## The Operator Showdown: ECK vs. OpenSearch

The market for production-grade Elasticsearch operators is dominated by two contenders:

1. **Elastic Cloud on Kubernetes (ECK)**: The official operator from Elastic
2. **OpenSearch Kubernetes Operator**: The official open-source operator for the OpenSearch project

The choice between them is not primarily technical—it's **political and legal**.

### Elastic Cloud on Kubernetes (ECK)

ECK is the official, vendor-supported operator from Elastic. It's mature, powerful, and battle-tested for managing Elasticsearch, Kibana, and other Elastic Stack components on Kubernetes.

#### Licensing: The Critical Issue

ECK's source code is publicly viewable but is **not open source**. It's licensed under the **Elastic License v2 (ELv2)**, a "source-available" license.

ECK operates on a freemium model:
- **Basic Tier (Free)**: Core orchestration features are "forever free"
- **Enterprise Tier (Paid)**: Advanced features like autoscaling, cross-cluster replication, and advanced security require a paid license

The core issue lies in the ELv2 license itself, which contains a "managed service" restriction:

> *"You may not: Provide the products to others as a managed service"*

This "source-available" license presents a significant legal "poison pill" for open-source projects. If a user of Planton builds a commercial SaaS product that includes an Elasticsearch cluster deployed via ECK, does that user (or Planton) constitute "providing a managed service"? The license language is notoriously ambiguous, creating unacceptable legal risk and FUD (Fear, Uncertainty, and Doubt) for both the project and its user community.

#### Features

**Basic (Free) Features**:
- Manages Elasticsearch and Kibana
- "Secure by default" with auto-generated TLS and password bootstrapping
- Safe rolling upgrades
- Advanced topology (hot-warm-cold) via `nodeSets`
- Snapshot scheduling

**Enterprise (Paid) Features**:
- Autoscaling
- Cross-cluster replication (CCR)
- Advanced security (SAML, LDAP, field/document-level security)
- Machine learning

ECK's adoption is extremely high, and it's widely regarded as robust and mature. Users report positive experiences, noting that the operator "took care of many complex upgrade issues."

### OpenSearch Kubernetes Operator

OpenSearch is an open-source fork of Elasticsearch 7.10.2, created by AWS and other partners in 2021 in response to Elastic's license change. The OpenSearch Kubernetes Operator is the official, community-driven operator for deploying and managing OpenSearch and OpenSearch Dashboards on Kubernetes.

#### Licensing: The Safe Harbor

The OpenSearch engine and the OpenSearch Kubernetes Operator are both licensed under the **Apache 2.0 license**.

Apache 2.0 is a permissive, OSI-approved open-source license. It places *no restrictions* on commercial use, modification, distribution, or the provision of managed services.

This makes the OpenSearch Operator the "safe harbor" for open-source projects. For a FOSS project like Planton, adopting the OpenSearch Operator is the path of *zero legal friction*. It fully protects the project and its users from the legal ambiguity of the Elastic License.

#### Features

**All-in-One (Free)**:
- Manages both OpenSearch and OpenSearch Dashboards
- No "paid" tier for orchestration features (FOSS-native)
- Core orchestration: declarative deployment, scaling, rolling upgrades
- `nodePools` abstraction (conceptually identical to ECK's `nodeSets`)
- Integrated with OpenSearch Security plugin (standard and *free* component)

The OpenSearch Operator is newer than ECK but is production-ready, under active development, and is the standard deployment method for any team building on a pure-FOSS stack.

### Operator Comparison Summary

| Feature | ECK (Elastic Cloud on Kubernetes) | OpenSearch Kubernetes Operator |
|:--------|:----------------------------------|:-------------------------------|
| **Engine** | Elasticsearch | OpenSearch |
| **Engine License** | Elastic License v2 & SSPL | **Apache 2.0** |
| **Operator License** | Elastic License v2 (Source Available, *not* OSI-approved) | **Apache 2.0** (100% FOSS) |
| **"Managed Service" Restriction?** | **Yes** (Legally ambiguous, high risk for FOSS projects) | **No** (Unrestricted commercial use) |
| **Governance Model** | Elastic Inc. (Single-vendor) | OpenSearch Project (Community/Consortium) |
| **Production Readiness** | Very High (Mature, vendor-backed, battle-tested) | High (Newer, but production-grade and standard for FOSS) |
| **Core Abstraction** | `nodeSets` | `nodePools` |
| **Key Free Features** | ES/Kibana Mgmt, Rolling Upgrades, Auto-TLS, Snapshot Mgmt | OS/Dashboards Mgmt, Rolling Upgrades, Security Plugin Mgmt |
| **Key Paid/Advanced Features** | **Autoscaling**, Cross-Cluster Replication (CCR), ML, Advanced Security | N/A (Features like Security and Anomaly Detection are free in OpenSearch) |

**Key Finding**: The minor feature gaps are a trade-off. OpenSearch provides rich *application* features (like security and anomaly detection) for free, which are part of Elastic's paid "Platinum" subscription. Conversely, ECK provides more advanced *orchestration* features in its paid "Enterprise" tier, most notably autoscaling and cross-cluster replication, which are not yet maturely implemented in the OpenSearch Operator.

For the vast majority of log analytics and search use cases, the performance of both engines is "good enough," and this should not be the primary decision driver.

## The Planton Approach: FOSS-First with Choice

Based on this analysis, Planton implements a clear, principled strategy:

### 1. Default to OpenSearch

`engine: OPENSEARCH` is the **default and recommended** choice for the `ElasticsearchKubernetes` resource. This aligns with the project's open-source values and completely shields all users from the legal risks of ELv2/SSPL.

### 2. Support Elasticsearch as an Opt-In

Planton also supports `engine: ELASTICSEARCH` as a secondary, non-default option. This is a pragmatic concession to Elastic's significant market share and allows users who have their own commercial licenses or have completed their own legal review to use it.

### 3. Forcible Licensing Warning

When a user explicitly sets `engine: ELASTICSEARCH`, the Planton controller (or CLI) prints a clear, non-dismissible warning:

```
WARNING: You are deploying Elasticsearch, which is licensed under the Elastic 
License v2. This license is NOT open source and contains restrictions on providing 
it as a managed service. Planton is not a legal advisor. You MUST conduct 
your own legal review if you are using this component in a commercial or hosted 
product. To use a 100% FOSS alternative, set engine: OPENSEARCH.
```

### 4. Unified API Abstraction

The `ElasticsearchKubernetes` API is designed as a "façade" that abstracts both operators behind a common interface:

- **Top-level discriminator**: `engine` field (ELASTICSEARCH | OPENSEARCH)
- **Common abstraction**: `nodePools` structure that maps to ECK's `nodeSets` and OpenSearch's `nodePools`
- **80/20 configuration**: Exposes only the essential fields that serve 80% of use cases

## Production Best Practices

The Planton API is designed to enforce production best practices and prevent common pitfalls.

### Cluster Topology and Split-Brain Prevention

In production, node roles should be separated for stability and performance. A common HA pattern involves 3 dedicated master nodes, with separate pools for data and ingest nodes.

**The most critical rule**: To prevent "split-brain" (where network partitions cause two parts of a cluster to elect separate masters), you must have an **odd number of master-eligible nodes, with a minimum of 3**.

Modern Elasticsearch and OpenSearch use quorum-based decision making. To form a quorum and elect a master, a majority of master-eligible nodes must be available. 

**Why 2-node master clusters are an anti-pattern**:
- A 1-node cluster has no fault tolerance. If the node fails, the cluster is down.
- A 2-node master cluster requires a quorum of (2/2) + 1 = 2 nodes to be healthy.
- If *either* node fails, the remaining single node *cannot* form a quorum. The cluster is down.
- Therefore, a 2-node cluster has *zero* fault tolerance, just like a 1-node cluster, but with twice the cost and complexity.

The Planton controller validates this: if a `nodePool` contains the "master" role, the `replicas` field must be an odd number, and configurations with exactly 2 master-eligible nodes are rejected.

### Resource Allocation: Guaranteed QoS & JVM Heap

This is the most common area of configuration failure.

**Guaranteed QoS**: Kubernetes pods should always be configured with `resources.requests` set equal to `resources.limits`. This places the pod in the **"Guaranteed" QoS class**, making it the last to be killed by the Kubernetes scheduler during resource contention.

**JVM Heap Sizing**: The JVM heap (`-Xms` and `-Xmx`) must be set to **50% of the pod's memory limit**. The other 50% is required by Lucene (which powers Elasticsearch) for the operating system's file system cache. This off-heap cache is critical for performance. If the JVM heap is set too high (e.g., 80% of pod RAM), the pod will be unstable and likely OOMKilled by Kubernetes.

**Planton's Opinionated Design**: The API intentionally *omits* a `jvmHeap` field. This creates a "pit of success" for users. Modern versions of ECK (>= 7.11) automatically derive the heap size directly from the pod's `resources.limits.memory` setting. By not exposing a `jvmHeap` field, the API forces users to follow this best practice. To increase the heap, they must increase the pod's memory limit. This prevents the most common failure mode.

### Storage and Persistence

**Storage Class**: Production workloads must use a `StorageClass` that provisions high-performance, SSD-backed PersistentVolumes. This is configured via the `storage.storageClassName` field.

**Volume Binding Mode**: The `StorageClass` should be configured by Kubernetes administrators with `volumeBindingMode: WaitForFirstConsumer`. This delays provisioning the PV until a pod is scheduled, allowing Kubernetes to provision the disk in the *same* availability zone as the node.

**Disaster Recovery**: The primary DR strategy is application-level **snapshots** to external repositories like S3, GCS, or Azure Blob. This functionality is planned for a future API version.

### High Availability

Cluster HA is achieved through:
1. A 3-node (or 5-node) master quorum
2. Ensuring all production indices have `index.number_of_replicas: 1` (or more)
3. **Pod Anti-Affinity**: Spreading pods across different physical Kubernetes nodes and availability zones

The Operator pattern provides a significant "free" win: ECK, by default, automatically injects a `podAntiAffinity` rule to prevent scheduling multiple nodes from the same cluster on the same Kubernetes host. This prevents a single Kubernetes node failure from taking down multiple Elasticsearch nodes—a critical 80% best practice enabled by default.

## Kibana and OpenSearch Dashboards Integration

Kibana (and its fork, OpenSearch Dashboards) is the web UI and visualization tool for the stack. Its deployment is integral to the user experience.

### Deployment Pattern

Kibana is a **stateless** application. Its configuration and saved objects are stored in an index *within* the Elasticsearch cluster. Because it's stateless, operators model it using a Kubernetes Deployment and Service.

### Authentication and Security

The operators automate the connection between Kibana and Elasticsearch. The Kibana CRD includes an `elasticsearchRef` field that tells the Kibana instance which Elasticsearch cluster to connect to. The operator automatically generates the correct service names and injects necessary credentials (like the `elastic` superuser password) into the Kibana pod's environment.

### Ingress and External Access

By default, operators create a `ClusterIP` service for Kibana, which is only accessible from *inside* the Kubernetes cluster. A user cannot access the UI from their browser.

The manual solution requires understanding and creating a Kubernetes Ingress resource that targets the Kibana `ClusterIP` service.

**Planton's Abstraction**: The `kibana.externalAccess` field is a high-level abstraction that solves this. When `enabled: true` and a `host` is provided, the Planton controller automatically generates the required Ingress resource. This is a significant usability improvement that abstracts away complex Kubernetes networking.

## Example Configurations

### Single-Node Dev (OpenSearch)

```yaml
engine: OPENSEARCH
version: "2.19.2"
nodePools:
  - name: "default"
    replicas: 1
    roles: ["master", "data", "ingest"]
    resources:
      limits: { memory: "4Gi", cpu: "1" }
      requests: { memory: "4Gi", cpu: "1" }
    storage: { size: "20Gi" }
kibana:
  replicas: 1
  resources:
    limits: { memory: "1Gi", cpu: "500m" }
    requests: { memory: "1Gi", cpu: "500m" }
```

### Small Production Cluster (HA, OpenSearch)

```yaml
engine: OPENSEARCH
version: "2.19.2"
nodePools:
  - name: "master-data-ingest"
    replicas: 3  # Enforces 3-node master quorum for HA
    roles: ["master", "data", "ingest"]
    resources:
      limits: { memory: "16Gi", cpu: "4" }
      requests: { memory: "16Gi", cpu: "4" }
    storage: { size: "200Gi", storageClassName: "ssd-fast" }
kibana:
  replicas: 2  # HA Kibana
  resources:
    limits: { memory: "2Gi", cpu: "1" }
    requests: { memory: "2Gi", cpu: "1" }
  externalAccess: 
    enabled: true
    host: "kibana.prod.company.com"
```

### HA Multi-Node Production (Role Separation, Elasticsearch)

```yaml
engine: ELASTICSEARCH
version: "8.19.6"
nodePools:
  - name: "master"
    replicas: 3  # 3 dedicated masters for HA
    roles: ["master"]
    resources:
      limits: { memory: "8Gi", cpu: "2" }
      requests: { memory: "8Gi", cpu: "2" }
    storage: { size: "50Gi" }
  - name: "data-ingest"
    replicas: 3  # 3 dedicated data/ingest nodes
    roles: ["data", "ingest"]
    resources:
      limits: { memory: "32Gi", cpu: "8" }
      requests: { memory: "32Gi", cpu: "8" }
    storage: { size: "1Ti", storageClassName: "io1" }
kibana:
  replicas: 2
  resources:
    limits: { memory: "2Gi", cpu: "1" }
    requests: { memory: "2Gi", cpu: "1" }
  externalAccess:
    enabled: true
    host: "kibana.prod.company.com"
```

## Conclusion: The Operator Paradigm Shift

The evolution from manual StatefulSets to autonomous Operators represents a fundamental paradigm shift in how we run stateful applications on Kubernetes. What was once considered an anti-pattern is now not just viable but often superior to traditional VM-based deployments.

The choice between Elasticsearch and OpenSearch is not primarily technical—both are mature, performant search engines. The decisive factor is licensing and the legal clarity that comes with Apache 2.0 versus the ambiguity of ELv2.

Planton's FOSS-first approach—defaulting to OpenSearch while supporting Elasticsearch as an opt-in—reflects a principled stance: open-source infrastructure should build on open-source foundations, with clear warnings when stepping outside that safe harbor.

By abstracting the complexity of both operators behind a unified, 80/20 API, Planton makes it easy to deploy production-grade search clusters on any Kubernetes platform—while letting you choose the engine and license model that aligns with your needs.

