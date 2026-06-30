# Deploying Redis on Kubernetes: A Strategic Guide

## Introduction: The Redis Ecosystem at a Crossroads

For years, deploying Redis on Kubernetes seemed straightforward—spin up a StatefulSet, maybe use a popular Helm chart, and you're done. But in March 2024, the Redis landscape fundamentally changed. Redis Ltd. abandoned the permissive BSD license that made Redis a beloved open-source project and adopted a restrictive dual-license model (RSALv2/SSPLv1) designed to prevent cloud providers from offering managed Redis services.

This license change wasn't just a legal technicality. The Server-Side Public License (SSPLv1) contains what many call a "poison pill": if you provide Redis as a service, you must open-source your entire infrastructure stack under the same license. For any company building a platform or SaaS product, this creates existential legal risk.

The open-source community's response was swift and decisive. AWS, Google Cloud, Oracle, and the original Redis developers forked the last truly open-source version (Redis 7.2.4) and created **Valkey**—a drop-in replacement for Redis, permanently licensed under BSD-3-Clause and governed by the Linux Foundation.

This schism has profound implications for infrastructure decisions. **Planton defaults to Valkey**, not Redis. This document explains the deployment landscape, the evolution from anti-patterns to production-ready solutions, and why certain architectural choices matter for your applications running on Kubernetes.

## The Maturity Spectrum: From Anti-Patterns to Production

Deploying stateful applications like Redis on Kubernetes isn't trivial. The platform was designed for ephemeral, stateless workloads. Understanding the evolution of deployment methods—from what *doesn't* work to what does—is essential for making informed decisions.

### Level 0: The Anti-Pattern (Kubernetes Deployment)

Many tutorials demonstrate deploying Redis using a standard Kubernetes `Deployment` with a `ConfigMap`. This is a **critical anti-pattern** for any production use case.

**Why it fails:**

1. **No Stable Identity**: Deployments manage fungible pods. When a pod restarts, it gets a new name and IP address. Redis replication and clustering require stable, discoverable addresses. When your master pod fails and restarts with a new identity, replicas can't find it. The replication chain breaks.

2. **Storage Roulette**: If you use `emptyDir` volumes, all data vanishes on every restart. If you attach a PersistentVolumeClaim (PVC), Kubernetes provides no guarantee the restarted pod will reattach to the *correct* volume. You might end up with a new, empty volume while your data sits orphaned on another.

3. **The Sidecar Trap**: Running Redis as a sidecar in your application pod is even worse. Every application deployment terminates the entire pod—wiping Redis from memory. This defeats the entire purpose of having a cache or database.

**Verdict**: Never use a Deployment for Redis. It violates both Kubernetes scheduling assumptions and Redis's architectural requirements.

### Level 1: The Foundation (StatefulSet)

The `StatefulSet` controller is the *correct* Kubernetes primitive for stateful applications. It provides the foundational guarantees that Deployments lack:

- **Stable Network Identity**: Pods get predictable, ordinal names (redis-0, redis-1, redis-2) preserved across restarts
- **Stable Storage**: Each pod gets a dedicated PVC via `volumeClaimTemplate`, and Kubernetes ensures redis-0 always reattaches to its original volume
- **Ordered Lifecycle**: Pods are created and terminated in predictable sequence

**The Critical Gap**: StatefulSets are *application-agnostic*. A StatefulSet knows how to restart redis-0 with its original storage and network identity. It does *not* know that redis-0 is your master, or that when redis-0 fails, redis-1 should be promoted to master.

StatefulSets solve the *Kubernetes problem* (pod identity). They don't solve the *Redis problem* (application-level high availability). This is why production deployments need a higher-level abstraction.

### Level 2: The Day 1 Solution (Helm Charts)

Helm charts package templated YAML files for installation and initial configuration. They correctly provision StatefulSets, Services (both ClusterIP and Headless), ConfigMaps, and related resources.

**The Bitnami Redis Chart**: Historically the most popular option, supporting all deployment modes (Standalone, Sentinel, Cluster). However, Bitnami (owned by VMware) fundamentally changed its container image distribution:

- **Latest Tag Only**: Free public images now offer only a `latest` tag—a production anti-pattern that creates unpredictable, non-repeatable deployments
- **Unmaintained Versioned Images**: Existing versioned images (e.g., `bitnami/redis:7.2.3`) moved to "Bitnami Legacy" and no longer receive security updates
- **Commercial Upsell**: Secure, versioned images now require a commercial Tanzu subscription

**The Bitnami Valkey Chart**: Unlike its Redis counterpart, this chart *is* production-ready because it uses the community-maintained `valkey/valkey` image (BSD-licensed, actively maintained by the Linux Foundation).

**The Helm Limitation**: Helm is "fire and forget." It installs resources but has no running component to manage Day 2 operations. It can't perform automated failover, graceful version upgrades, or application-aware resizing.

Consider a version upgrade scenario: running `helm upgrade --set image.tag=...` triggers a rolling restart. Helm has no awareness of which pod is the master. It might terminate the master first, causing downtime or data loss. This is unacceptable in production.

### Level 3: The Production Solution (Kubernetes Operators)

A Kubernetes Operator is an active controller running in your cluster, extending the Kubernetes API with Custom Resource Definitions (CRDs). It continuously reconciles your declared desired state with actual state—effectively an automated SRE codified in software.

**The Operator Advantage** (version upgrade example):

When you update your `RedisCluster` CRD to specify `spec.image.tag: 7.2.11`, the Operator:

1. Upgrades replica pods one by one
2. After each upgrade, waits for the pod to become Ready *and verifies via Redis commands* that it's fully synced
3. Once all replicas are upgraded and synced, triggers a graceful Sentinel-managed failover to promote a replica
4. Finally upgrades the old master (now demoted to replica)

This is application-aware lifecycle management. It's the reason Operators are superior for complex, stateful, production deployments.

## The Operator and Helm Chart Landscape

The Redis/Valkey ecosystem has fragmented into "legacy Redis," "commercial Redis," and "future OSS Valkey." Here's how the major solutions compare:

### Production-Ready Open Source Solutions

**OT-Operator (by Opstree)**

- **License**: Apache 2.0 (code), uses standard Redis OSS images (BSD, versions 6/7)
- **Status**: Actively maintained, successor to the abandoned Spotahome operator
- **Architecture**: Modular design with separate CRDs per mode: `Redis` (standalone), `RedisReplication`/`RedisSentinel` (HA), `RedisCluster` (sharding)
- **Features**: Built-in monitoring (redis-exporter), TLS support, secure-by-default security contexts, pod anti-affinity
- **Valkey Support**: Roadmap item for official support; community confirms it works today by overriding the image in CRDs
- **Production Readiness**: **Yes** ✓ – Widely deployed, well-documented, stable

**Verdict**: This is the best-in-class operator for Redis/Valkey. Its CRD-per-mode design is excellent API design—simple for users, powerful for operators.

**Bitnami Valkey Helm Chart**

- **License**: Apache 2.0 (chart), BSD-3-Clause (valkey/valkey image)
- **Status**: Actively maintained
- **Modes**: Standalone, Sentinel, Cluster
- **Production Readiness**: **Yes** ✓ – Unlike the Redis chart, this uses truly open-source, maintained container images

**Verdict**: A solid Day 1 solution for teams that don't need Day 2 operator automation.

**Valkey-Native Operators (Emerging Ecosystem)**

Two operators built specifically for Valkey are production-ready:

- **Hyperspike Valkey Operator**: Focused on **Cluster mode** (sharding). Simple CRD exposing `nodes` (masters) and `replicas` per master. Apache 2.0 licensed.
- **SAP Valkey Operator**: Focused on **Sentinel mode** (HA). Apache 2.0 licensed.

**Verdict**: These represent the future of the open-source ecosystem. Their specialization (Cluster vs. Sentinel) allows for simpler, more focused controllers.

### Solutions to Avoid

**Spotahome Redis Operator**

- **Status**: **Abandoned** – No updates in years
- **Critical Bugs**: Sentinels from different clusters can accidentally join each other, causing catastrophic HA failures
- **Verdict**: Do not use

**Bitnami Redis Helm Chart**

- **Issue**: While the chart code is Apache 2.0, it now defaults to unmaintained "legacy" images or insecure `latest` tags
- **Verdict**: Production liability due to security and maintenance risks

**Redis Enterprise Operator**

- **License**: Proprietary (code and images)
- **Business Model**: Requires commercial license (30-day/4-shard trial available)
- **Verdict**: Excellent product, but not suitable for open-source projects

## Redis vs. Valkey: The Licensing Decision

For any open-source infrastructure framework like Planton, the licensing choice isn't optional—it's a fundamental responsibility to users.

**The Risk of Default Redis**:

- **Redis 7.4+**: Subject to RSALv2/SSPLv1. If a user unknowingly builds a SaaS product using Planton with Redis 7.4+, they may face severe license violations. The SSPL requires open-sourcing their entire infrastructure stack.
- **Bitnami redis:7.2.3**: Points to unmaintained, insecure images in the "Legacy" registry.

**The Valkey Path**:

- **Truly Open Source**: BSD-3-Clause, OSI-approved
- **Linux Foundation Governance**: Ensures long-term stability and community-driven development
- **Drop-In Replacement**: Binary-compatible with Redis 7.2.4
- **Active Development**: Security patches, features, and community support

**Planton's Choice**: Default to `valkey/valkey` images. The `RedisKubernetes` API name is retained for backward compatibility and familiarity (Redis remains the widely-recognized term), but the implementation uses Valkey to ensure legal safety and long-term viability.

## Deployment Modes: Standalone, Sentinel, and Cluster

The deployment mode determines your high availability, scaling strategy, and operational complexity. Choose carefully based on your actual requirements.

### Standalone Mode

**Architecture**: Single Redis/Valkey instance

**Use Case**: Development, testing, or non-critical ephemeral caching where you can tolerate complete data loss on restart

**Production Use**: **Never** for a database. Acceptable for cache-only workloads *if* your application gracefully handles cold cache scenarios

**Drawback**: Complete Single Point of Failure (SPOF). If the pod or node fails, service is 100% down.

### Sentinel Mode (High Availability)

**Architecture**: Master-replica setup monitored by a separate cluster of Sentinel processes (typically 3)

**How Automated Failover Works**:

1. **Monitor**: Sentinels continuously health-check the master and replica pods
2. **Quorum**: If the master fails, Sentinels vote. When a quorum (e.g., 2 of 3) agrees the master is down, failover begins
3. **Automatic Promotion**: Sentinels elect an in-sync replica, promote it to master, and reconfigure other replicas
4. **Service Discovery**: Application clients connect to Sentinels and ask "Who is the current master?" Sentinels provide the correct answer, even immediately after failover

**Limitations**: This provides **High Availability**, not scalability. All data must fit in a single master node's RAM, and all writes go to that single master.

**Use Case**: **This is the 80% production use case.** Choose Sentinel when you need resilience to pod/node failures and your dataset fits comfortably on a single large pod.

**Minimal Setup**: 1 master, 2 replicas, 3 Sentinels

**Production Recommendation**: Highly recommended. This is the robust, well-understood default for most production applications.

### Cluster Mode (Sharding + HA)

**Architecture**: Multi-master system where the dataset is sharded across multiple master nodes (typically 3+). Each master has its own replica(s) for HA.

**How Sharding Works**:

1. **Hash Slots**: The keyspace is divided into 16,384 hash slots. Each master owns a subset (e.g., Master-A: 0-5460, Master-B: 5461-10922, Master-C: 10923-16383)
2. **Internal Failover**: No Sentinel needed. Cluster nodes communicate over a "cluster bus" (data port + 10000). If a master fails, its replica is promoted via internal voting
3. **Client-Side Routing**: Cluster-aware clients connect to any node, fetch the slot map, compute each key's slot, and route commands directly to the correct master

**Limitations**:

- **Complexity**: More difficult to manage and debug
- **Multi-Key Operations**: Transactions, `MSET`, `SUNION` only work if all keys hash to the same slot (forceable with hash tags like `mykey{user123}`)

**Use Case**: **This is a scalability solution.** Use it only when your dataset is too large for a single node's RAM or write throughput exceeds a single master's CPU capacity.

**Minimal Setup**: 3 masters, 3 replicas (6 nodes total)

**Production Recommendation**: Don't use as a default. Start with Sentinel. Migrate to Cluster only when monitoring proves you've hit a scaling bottleneck.

### Mode Comparison Summary

| Feature | Sentinel (HA) | Cluster (Sharding + HA) |
|---------|---------------|-------------------------|
| **Primary Goal** | High Availability | Horizontal Scaling |
| **Dataset Constraint** | Must fit in one node's RAM | Distributed across multiple nodes |
| **Write Scalability** | Limited to single master | Scales linearly with masters |
| **Failover** | Automatic (external Sentinels) | Automatic (internal gossip) |
| **Complexity** | Moderate | High |
| **Minimal Production Setup** | 1 master + 2 replicas + 3 Sentinels | 3 masters + 3 replicas (6 nodes) |
| **Multi-Key Operations** | Fully supported | Limited (keys must be in same slot) |
| **Recommended For** | 80% of production use cases | Large datasets or extreme write load |

## Planton's Implementation Strategy

The `RedisKubernetes` API in Planton follows the 80/20 principle: expose the 20% of configuration that 80% of users need, and abstract away complexity.

### Current API Design

The API currently supports:

- **Replicas**: Number of Redis pods
- **Resources**: CPU and memory requests/limits per pod
- **Persistence**: Enable/disable persistence and set disk size
- **Ingress**: Optional external access via LoadBalancer

**Design Philosophy**: The API is intentionally simple, focusing on the most common use case (Standalone or basic replication) while leaving the door open for future enhancements (Sentinel and Cluster modes).

### Future Enhancements

Based on the research, future versions should consider:

**Mode Selection**: Add a `mode` enum (STANDALONE, SENTINEL, CLUSTER) with mode-specific configuration blocks:

- **Sentinel Config**: `redisReplicas` (total pods), `sentinelReplicas`, `quorum`
- **Cluster Config**: `masters` (shards), `replicasPerMaster`

**Image Flexibility**: Explicit `image` spec (repository, tag) to allow choosing between Valkey and legacy Redis

**Security Enhancements**: Native TLS support, authentication via Kubernetes Secrets

**Advanced HA**: Automatic pod anti-affinity, PodDisruptionBudgets, topology spread constraints

### Why Abstraction Matters

An infrastructure API should capture *intent*, not *implementation details*. Users should declare "I want a 50GB highly-available database" and the system should provision the StatefulSet, configure persistence, apply anti-affinity rules, create PodDisruptionBudgets, and set up monitoring—without requiring the user to understand every Kubernetes primitive.

This is the core value proposition of Planton: translate high-level intent into production-ready infrastructure.

## Production Hardening Best Practices

A production deployment requires more than just running pods. The implementation (whether Helm chart or Operator) should automatically apply these hardening measures:

### Data Persistence Strategy

**RDB (Redis Database)**: Point-in-time snapshots. Fast recovery, but potential data loss between snapshots.

**AOF (Append Only File)**: Logs every write operation. High durability (at most 1 second of data at risk with `fsync: everysec`), but larger file size and slower recovery.

**Hybrid (RDB + AOF)**: Modern default. Creates an RDB snapshot and appends AOF logs since that point. Best of both worlds.

**Recommendation**: Always use hybrid mode for production persistence.

### High Availability Scheduling

**Pod Anti-Affinity**: *Essential* for HA modes. Tells Kubernetes not to place redis-0 and redis-1 on the same physical node. Without this, a single node failure takes out both master and replica, defeating HA entirely.

**PodDisruptionBudget (PDB)**: Limits voluntary disruptions (e.g., node draining). Prevents an administrator from accidentally draining all Sentinel pods simultaneously, which would break quorum.

**Topology Spread Constraints**: Advanced scheduling that distributes pods across nodes *and* availability zones for maximum availability.

**Recommendation**: Operators should automatically inject these for SENTINEL and CLUSTER modes.

### Resource Management

**Memory**: The most critical resource. Set requests and limits to the *same value* for "Guaranteed" QoS class (least likely to be evicted). Configure Redis `maxmemory` to 75-80% of container limit (Redis needs 20-25% overhead for replication buffers).

**CPU**: Redis is primarily single-threaded. Set reasonable request (e.g., 1 core) with higher limit (e.g., 2 cores) for bursting. Monitor throttling metrics—CPU throttling can hang the event loop and cause cascading timeouts.

**Disk**: Use SSD-backed storage. Provision at least 2-3x the `maxmemory` setting to accommodate RDB snapshots, AOF logs, and growth.

### Security Posture

**Authentication**: Always enforce `requirepass` in production (via Kubernetes Secret).

**securityContext**: Run as non-root user, set `readOnlyRootFilesystem: true`, configure `fsGroup` for volume ownership.

**NetworkPolicy**: Default-deny ingress, explicitly allow port 6379 (Redis), 26379 (Sentinel), and 16379 (Cluster bus) only from authorized pods.

**Recommendation**: Secure-by-default should be the implementation standard.

### Network Access Patterns

**For Standalone**: Use a standard ClusterIP Service. Applications connect to a single stable DNS name.

**For Sentinel and Cluster**: Use a Headless Service (`clusterIP: None`). This is *mandatory* because:

- **Sentinel clients** need to discover and communicate with *all* Sentinel pods to verify quorum
- **Cluster clients** receive `MOVED` redirects containing direct pod IPs. ClusterIP abstraction breaks this, causing infinite redirect loops

A ClusterIP service load-balances to a random pod, hiding individual pod identities. This breaks peer discovery and direct routing required by Sentinel and Cluster protocols.

**External Access**: Not recommended for v1. In-cluster access should be the only supported pattern. External access for Cluster mode is particularly complex (clients can't reach internal pod IPs from outside the cluster).

## Conclusion: Choosing the Right Foundation

The Redis ecosystem underwent a seismic shift in 2024. The license change by Redis Ltd. forced the community to choose: accept restrictive terms or fork and preserve open-source principles. The Linux Foundation's Valkey project represents the latter—a commitment to openness, community governance, and legal safety.

For infrastructure frameworks like Planton, this choice is clear. **Defaulting to Valkey** protects users from legal risk while maintaining full technical compatibility with the Redis they know.

The deployment landscape has matured from simple StatefulSets to sophisticated Operators that codify SRE best practices into software. Understanding the progression—from anti-patterns (Deployments) to foundational building blocks (StatefulSets) to production solutions (Operators)—informs better architectural decisions.

**Start simple**: For development, a Standalone instance is fine. For production, **default to Sentinel mode**—it provides automated failover without the operational burden of Cluster mode. Only move to Cluster when metrics prove you need horizontal scaling.

**Prioritize operational maturity**: Operators like OT-Operator automate the runbooks SREs would execute manually. In production, this automation isn't a luxury—it's essential for reliable, resilient infrastructure.

The `RedisKubernetes` API in Planton embodies these principles: sensible defaults, abstraction of complexity, and a foundation built on truly open-source components that will remain available and maintained for the long term.

---

**Further Reading**:

- [Valkey Project Homepage](https://valkey.io/)
- [OT-Operator Documentation](https://ot-container-kit.github.io/redis-operator/)
- [Bitnami Valkey Helm Chart](https://artifacthub.io/packages/helm/bitnami/valkey)
- [Redis Sentinel Documentation](https://redis.io/docs/latest/operate/oss_and_stack/management/sentinel/)
- [Redis Cluster Specification](https://redis.io/docs/latest/operate/oss_and_stack/reference/cluster-spec/)

