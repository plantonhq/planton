# Deploying NATS on Kubernetes: A Production-Ready Guide

## Introduction: The Messaging System That Embraces Simplicity

For years, the conventional wisdom in distributed systems was that powerful messaging requires complexity. Kafka demanded Zookeeper clusters and partition management. RabbitMQ introduced elaborate AMQP routing topologies. Redis Streams bolted persistence onto a cache.

Then came NATS: a single binary, a simple subject-based model, and performance that embarrasses the competition—6 million messages per second compared to RabbitMQ's 60,000. The message was clear: simplicity is not a compromise; it's a strategic advantage.

NATS is an open-source messaging system (Apache 2.0, CNCF project) designed as "connective technology" for distributed applications. It excels at microservices communication, request-reply patterns, and event streaming—all delivered through an intentionally straightforward API that respects both developer time and operational sanity.

This document explores how to deploy production NATS clusters on Kubernetes, examining the evolution from anti-patterns to modern best practices, and explaining why Planton's approach honors NATS's philosophy of simplicity.

## The Evolution: From Core NATS to JetStream

Understanding NATS deployment begins with understanding its two operational modes, both embedded in the same `nats-server` binary:

### Core NATS: The Fire-and-Forget Transport

Core NATS is an in-memory, at-most-once messaging system. It's exceptionally fast and low-latency, perfect for service discovery, RPC-style request-reply, and scenarios where TCP-level reliability is sufficient. If a subscriber is offline when a message arrives, that message is lost—by design.

This is NATS at its purest: a lightweight pub/sub fabric with no persistence overhead.

### NATS JetStream: The Persistence Layer

JetStream is not a separate server; it's an optionally enabled subsystem within the same NATS server. When enabled, it fundamentally transforms NATS by adding:

- **At-least-once and exactly-once delivery guarantees**
- **Historical message replay** and durable subscribers
- **Persistence** via disk-based file storage
- **Higher-level abstractions**: built-in Key-Value stores and Object Storage

JetStream, introduced in NATS 2.2, explicitly replaces the older NATS Streaming (STAN) system. Any modern Kubernetes deployment must focus exclusively on JetStream for persistence.

| Feature | Core NATS | JetStream |
|---------|-----------|-----------|
| **Quality of Service** | At-most-once | At-most-once, At-least-once, Exactly-once |
| **Persistence** | In-memory only | In-memory and/or disk-based file storage |
| **Primary Pattern** | Pub/Sub, Request-Reply | Durable streaming, event sourcing |
| **Data Replay** | Not supported | Historical replay from any point |
| **High-Level Services** | None | Key-Value store, Object store |

## The Deployment Landscape: From Anti-Patterns to Production

Deploying NATS on Kubernetes requires understanding which primitives match its stateful, clustered nature. Choosing wrongly isn't just inefficient—it's catastrophic.

### Level 0: The Anti-Pattern (Deployments and DaemonSets)

Using a standard Kubernetes Deployment or DaemonSet for a NATS cluster is an anti-pattern that guarantees failure.

**Why Deployments Fail**: Kubernetes Deployments are designed for stateless applications with interchangeable pods. When a Deployment restarts a pod, it creates a new pod with a random hostname (e.g., `nats-deployment-6b8f...`).

NATS JetStream clustering uses the RAFT consensus algorithm, which requires a quorum (majority of members) to elect a leader and operate. RAFT relies on **stable identities** for its members.

Here's the failure sequence:

1. A 3-replica cluster starts with pods `nats-abc`, `nats-def`, and `nats-ghi`. These identities are recorded in RAFT's persistent log.
2. Kubernetes reschedules `nats-abc`, terminating it and creating `nats-xyz`.
3. The cluster sees `nats-xyz` as a *fourth member* joining. The RAFT log now lists four members, with `nats-abc` marked "offline."
4. After several pod-churn events, the RAFT member list might contain 5, 6, or 7 "phantom-offline" servers.
5. The 3 active pods no longer constitute a quorum of the *total* registered members (3 active / 7 total = no quorum).
6. **The cluster fails to elect a leader, JetStream stops accepting messages, and the cluster is irrecoverably broken.**

**Why DaemonSets Fail**: DaemonSets run one pod per Kubernetes node, which doesn't match the desired 3- or 5-replica NATS cluster topology. This is a tool mismatch, not a solution.

### Level 1: The Required Primitive (StatefulSets)

For any application requiring stable identity and storage, Kubernetes provides the **StatefulSet** primitive. A clustered NATS server with JetStream is definitionally a stateful application.

The official NATS documentation is explicit: **"The recommended way to deploy NATS on Kubernetes is using Helm with the official NATS Helm Chart"**—and that chart deploys NATS as a StatefulSet.

**What StatefulSets Provide**:

1. **Stable, Unique Network IDs**: Each pod gets a predictable hostname based on an ordinal index: `nats-0`, `nats-1`, `nats-2`. This stable identity is precisely what RAFT requires across restarts and upgrades.

2. **Stable, Persistent Storage**: StatefulSet's `volumeClaimTemplates` ensures each pod gets its own unique PersistentVolumeClaim (PVC). `nats-0` is bound to `pvc-nats-0`, `nats-1` to `pvc-nats-1`, and so on. This is required for JetStream's file-based persistence.

3. **Sequenced, Graceful Rollouts**: StatefulSets perform ordered updates (e.g., updating `nats-2`, then `nats-1`, then `nats-0`). This controlled rollout maintains quorum during upgrades, ensuring a majority of the cluster is always available.

### Level 2: The Two-Service Model

A correct NATS deployment requires understanding that the cluster has two distinct types of network traffic:

1. **Client-to-Server Traffic (Port 4222)**: Applications need to connect to *any* healthy NATS pod. A standard **ClusterIP Service** is perfect for this—it provides a single, stable DNS name that load-balances client connections across the cluster.

2. **Server-to-Server Traffic (Port 6222)**: A NATS server pod (e.g., `nats-0`) needs to form a full mesh with *all* of its peers (`nats-1`, `nats-2`) for clustering and RAFT communication. It cannot use the ClusterIP service for this, as it might just get load-balanced back to itself.

The **Headless Service** (defined by setting `clusterIP: None`) solves this. When queried via DNS, it returns the *list of all individual pod IPs*, allowing `nats-0` to discover the actual IPs of `nats-1` and `nats-2`.

**Verdict**: A production NATS deployment requires a StatefulSet paired with *two* Service objects:
- A **ClusterIP Service** for client connections (port 4222)
- A **Headless Service** for server-to-server peering (port 6222), linked to the StatefulSet via `spec.serviceName`

## The Official NATS Tooling: Helm, Operators, and Nack

The NATS.io team provides several tools for Kubernetes, but their roles and recommendations have evolved significantly.

### The Standard: nats-io/k8s Helm Chart

The official NATS documentation is unambiguous: **"The recommended way to deploy NATS on Kubernetes is using Helm with the official NATS Helm Chart."**

This chart, hosted at `https://nats-io.github.io/k8s/helm/charts/`, is the community and maintainer-backed standard. It correctly provisions:

- The StatefulSet for NATS servers
- The Headless Service for cluster peering
- The ClusterIP service for client connections
- An optional `nats-box` Deployment for debugging

This chart serves as the best-practice model for configuration and is the foundation for any production deployment.

### The Deprecated: nats-io/nats-operator

The `nats-io/nats-operator` uses a Custom Resource Definition (CRD) called `NatsCluster` to manage NATS deployments.

However, the README file of the operator repository itself contains a prominent warning:

> **"⚠️ The recommended way of running NATS on Kubernetes is by using the Helm charts... The NATS Operator is not recommended to be used for new deployments."**

This is a definitive, first-party deprecation. Unlike systems like Kafka (where the Strimzi operator is recommended) or PostgreSQL, the NATS maintainers have abandoned the operator model in favor of the simpler Helm chart. This decision aligns with NATS's core philosophy: avoiding high complexity when a StatefulSet managed by Helm is sufficient.

**Verdict**: Do not use the nats-operator. It is obsolete.

### The Configuration Manager: nats-io/nack

Nack ("NATS Controllers for Kubernetes") is a separate controller that *does not* deploy the NATS cluster itself. Instead, it manages JetStream *resources* (Streams, Consumers, Key-Value Stores, Object Stores) using Kubernetes-native CRDs.

An administrator or application developer can define a NATS Stream in YAML:

```yaml
apiVersion: jetstream.nats.io/v1beta2
kind: Stream
metadata:
  name: orders-stream
spec:
  subjects:
    - orders.*
  storage: file
  replicas: 3
```

When this manifest is applied (via `kubectl apply` or GitOps), the Nack controller detects it and issues commands to the NATS cluster to create or update that stream.

This enables a powerful **"Infrastructure vs. Configuration"** pattern that separates concerns:

1. **Infrastructure (Day 1)**: The NATS cluster itself (StatefulSet, PVCs, Services) is a "slow-moving" resource deployed by a platform team.
2. **Configuration (Day 2)**: Streams and Consumers are "fast-moving" resources defined by application teams and managed declaratively via Git.

This avoids the disaster of letting application code create/manage streams ad-hoc, which leads to configuration drift and conflicts.

| Tool | Purpose | Production Readiness | Recommendation |
|------|---------|---------------------|----------------|
| **nats-io/k8s Helm Chart** | Deploys NATS infrastructure (StatefulSet, Services) | **Recommended** | Use as the model for deployment |
| **nats-io/nats-operator** | Deploys NATS infrastructure | **DEPRECATED** | Do not use. Obsolete. |
| **nats-io/nack** | Manages NATS resources (Streams, Consumers) | **Recommended** | Deploy *alongside* cluster for GitOps |

### Licensing: 100% Open Source

All official NATS tooling and container images are 100% open source under the Apache 2.0 license:

- `nats-io/k8s` (Helm Charts): Apache 2.0
- `nats-io/nack` (Controller): Apache 2.0
- `nats-io/nats-docker` (Container Images): Apache 2.0
- `nats-io/nats-box` (Utility Image): Apache 2.0

This confirms a clean bill of health for integration into any open-source or commercial platform.

## Planton's Approach: Simplicity by Design

Planton's `NatsKubernetes` API generates the same artifacts as the official `nats-io/k8s` Helm chart, following the maintainer-recommended pattern. The API focuses on the **80/20 configuration principle**: exposing the 20% of fields that 80% of users need for production deployments.

### Essential Configuration (The "80%")

**Server Configuration**:
- `replicas`: The most fundamental field. Development uses 1; production uses 3 or 5 (odd numbers for RAFT quorum).
- `resources`: CPU and memory requests/limits. Critical for production stability—NATS pods resyncing large JetStream streams without limits can be OOMKilled by Kubernetes, leading to crash-loops.

**JetStream Configuration**:
- `jetstream.enabled`: Top-level toggle for persistence.
- `jetstream.persistence.mode`: Choose `FILE` (disk, production) or `MEMORY` (ephemeral, staging).
- `jetstream.persistence.size`: PVC size (e.g., `10Gi`).

**Security Configuration**:
- `auth.enabled`: Default NATS has no authentication ("useful for development only"). Production must enable this.
- `auth.mode`: Support `BASIC` (username/password) or `TOKEN` for the 80% use case.
- `tls.enabled`: Enable TLS for client connections, referencing a Kubernetes Secret.

**Utility and Access**:
- `nats_box.enabled`: Deploy the `nats-box` utility pod for debugging.
- `external_access.type`: Expose via `LOADBALANCER` (L4 TCP) or `INGRESS` (L7 WebSocket).
- `nack.enabled`: Co-deploy the Nack controller to enable GitOps-based JetStream resource management.

**Monitoring**:
- `monitoring.enabled`: Deploy the `prometheus-nats-exporter` sidecar for Prometheus integration.

### Omitted Advanced Features (The "20%")

To honor NATS's philosophy of simplicity, the v1 API omits corner-case configurations:

- **Leafnodes**: For edge-to-cloud topologies
- **Gateway**: For "super-cluster" (multi-cluster) topologies
- **MQTT**: For MQTT protocol gateway
- **Advanced Auth**: Full NKey/JWT multi-tenancy
- **Advanced JetStream Tuning**: Per-server max_payload, etc.

These can be considered for v2 if user demand warrants the added complexity.

### Example Configurations

**Development (Minimal)**:
- `replicas: 1`
- `auth.enabled: false`
- `jetstream.enabled: false`

**Staging (Clustered, In-Memory)**:
- `replicas: 3`
- `auth.enabled: true` with `mode: BASIC`
- `jetstream.enabled: true` with `persistence.mode: MEMORY`
- `nats_box.enabled: true`

**Production (HA, Persistent, Secure)**:
- `replicas: 3`
- `resources`: CPU/memory requests and limits
- `auth.enabled: true` with `mode: TOKEN`
- `tls.enabled: true`
- `jetstream.enabled: true` with `persistence.mode: FILE, size: 50Gi`
- `external_access.type: LOADBALANCER`
- `nack.enabled: true`
- `monitoring.enabled: true`

## Production Best Practices

### Clustering and High Availability

- **Always use odd-numbered replicas** (3 or 5) for StatefulSets to satisfy RAFT quorum requirements.
- **Never scale from 1 to 3** in production. Start with 3 replicas. Scaling a live JetStream cluster from 1 to 3 is risky and can cause "group node missing" errors and inconsistent state.
- **Use pod anti-affinity** to ensure NATS server pods run on different physical nodes, preventing a single node failure from causing quorum loss.

### JetStream Persistence

- **Use file storage** (`fileStore`) with high-performance SSDs for production durability. Memory storage (`memStore`) does not survive pod restarts.
- **Provision adequate PVC size** for `fileStore`. Insufficient disk is a common and avoidable failure.

### Authentication and Authorization

- **Enforce authentication** (`auth.enabled: true`) in all production environments. No-auth is for development only.
- For most use cases, `TOKEN` or `BASIC` auth (via Kubernetes Secrets) covers the 80% need. Complex multi-tenancy with NKeys/JWTs is an advanced feature.

### TLS Configuration

- **Enable TLS** for all endpoints: client connections (port 4222) and internal cluster mesh (port 6222).
- **Integrate with cert-manager** to automatically provision and rotate TLS certificates, storing them in Kubernetes Secrets.

### Resource Sizing

- **Always set CPU and memory limits**. NATS pods that fall behind (e.g., due to network partitions) will attempt to catch up by replicating large amounts of data. Without limits, they can be OOMKilled by Kubernetes, entering a crash-loop.

### Monitoring

- **Enable the NATS monitoring endpoint** (port 8222).
- **Deploy the prometheus-nats-exporter** to expose metrics to Prometheus.
- **Import standard Grafana dashboards** for NATS Server and NATS JetStream to gain immediate operational visibility.

**Key Metrics to Monitor**:

| Metric | Description | Alert Condition |
|--------|-------------|-----------------|
| `nats_server_total_connections` | Total active client connections | > 90% of `max_connections` |
| `nats_server_slow_consumers` | Consumers not keeping up | > 0 (critical indicator) |
| `nats_jetstream_storage_bytes` | Total bytes used by JetStream | > 80% of disk size |
| `nats_jetstream_meta_cluster_leader` | Is this node the meta-cluster leader? | sum() != 1 (quorum loss!) |
| `nats_jetstream_stream_replicas_lag` | Replication lag between replicas | > threshold (replica falling behind) |

## Client Connectivity Patterns

### Internal Access: The Two-Service Model

Applications inside the cluster connect to NATS via the **ClusterIP Service** (e.g., `nats-client.default.svc.cluster.local`), which load-balances connections across all healthy NATS pods.

NATS server pods discover each other via the **Headless Service**, which returns individual pod IPs to enable the full-mesh peering required by RAFT.

### External Access: LoadBalancer vs. Ingress

**LoadBalancer**:
- **Pro**: Simplest way to expose the L4 TCP client port (4222) to the internet.
- **Con**: Provisions a dedicated, often expensive cloud load balancer for each NATS cluster.

**Ingress**:
- **Pro**: Cost-effective. Can share a single L7 load balancer among many services.
- **Con**: Standard Ingress controllers (e.g., NGINX) are HTTP-based, typically suitable only for NATS clients connecting via WebSockets, not raw TCP.

Choose **LoadBalancer** for general-purpose TCP clients. Choose **Ingress** for browser-based (WebSocket) clients.

### Multi-Cluster NATS (Advanced)

NATS natively supports "super-clusters" by federating multiple independent clusters (e.g., in different clouds or regions) using its `gateway` configuration. This is an advanced topology but a core strength of NATS's "connective technology" philosophy.

## Conclusion: The Paradigm Shift

The evolution of NATS deployment on Kubernetes mirrors the system's broader philosophy: **simplicity is not a limitation; it's a superpower**.

Where competitors demand elaborate operational overhead—Zookeeper clusters, partition management, complex routing topologies—NATS offers a single binary, a straightforward subject-based model, and production-proven performance. But simplicity doesn't mean naivety. Deploying NATS correctly on Kubernetes requires understanding stateful primitives (StatefulSets, not Deployments), consensus requirements (RAFT quorum), and the separation of infrastructure from configuration (Helm for clusters, Nack for streams).

Planton's `NatsKubernetes` API honors this philosophy by focusing on the 80% configuration that matters, generating maintainer-recommended artifacts, and enabling modern GitOps patterns. The result is a deployment model that feels as simple as NATS itself—and delivers the same production-grade reliability.

For deeper guides on specific implementation details, operator configuration, and advanced patterns, see the [NATS Kubernetes Deep Dive](./nats-kubernetes-deep-dive.md) (coming soon).

