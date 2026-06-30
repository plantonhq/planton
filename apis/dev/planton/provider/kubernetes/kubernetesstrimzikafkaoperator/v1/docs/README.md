# Deploying Kafka on Kubernetes: The Operator Approach

For years, the prevailing wisdom held that stateful workloads like Apache Kafka didn't belong on Kubernetes. "Kubernetes is for stateless apps," the argument went. "Running Kafka there is asking for trouble." This conventional wisdom has been thoroughly disproven. Not only *can* Kafka run successfully on Kubernetes—it's become the preferred deployment method for many production environments, thanks to the emergence of sophisticated **Kafka operators**.

The transformation happened not because Kubernetes suddenly became better at managing state (though it has improved), but because the Kafka community developed **purpose-built controllers** that encode years of operational expertise into automated reconciliation loops. These operators don't just deploy Kafka—they continuously manage its lifecycle, handling the complex orchestration that once required dedicated ops teams armed with runbooks.

## The Evolution: From Manual Management to Operator Automation

Understanding where Kafka operators fit requires seeing the progression of deployment approaches and what each solved (or failed to solve).

### Level 0: The Manual Approach (StatefulSets with Scripts)

In the early days of Kubernetes adoption, teams would deploy Kafka using **StatefulSets** directly—essentially treating Kubernetes primitives as building blocks for DIY orchestration. Helm charts like Bitnami's Kafka chart emerged to simplify this: run `helm install`, get a Kafka cluster with configurable replicas and persistence. Configure TLS, SASL authentication, storage, and resource limits through a `values.yaml` file.

This approach handles the **Day-1** problem elegantly: getting Kafka running. A 3-broker cluster with ZooKeeper can be up in minutes. The chart creates StatefulSets, Services, ConfigMaps, and PVCs in the right order with proper naming conventions.

**But Day-2 operations expose the limitations:**

- **Scaling requires manual partition reassignment**: Helm can add a new broker pod, but it won't redistribute existing topic partitions. You must run Kafka's reassignment tools manually or the new broker sits idle.
- **Configuration changes need careful orchestration**: Updating broker configs that require restarts means manually rolling brokers one at a time, ensuring no under-replicated partitions between steps.
- **No automated failure recovery**: If a broker pod fails with corrupted state, Kubernetes restarts it, but there's no intelligence to check cluster health first or rebalance if needed.
- **Topic and user management stays external**: Creating topics, managing ACLs, rotating credentials—all must be handled via Kafka CLI tools or separate automation. There's no declarative, Kubernetes-native way to say "I want topic X with these settings."

The 80/20 principle applies: Helm-based StatefulSet deployment covers 80% of getting Kafka running, but the remaining 20%—graceful operations at scale—demands significant manual effort. For dev/test environments or small deployments, this might be acceptable. For production clusters serving critical workloads, it's a maintenance burden.

**Verdict**: Viable for experimentation and learning, but error-prone and operationally expensive for production.

### Level 1: Enter the Operators (Automated Kafka Management)

The breakthrough came with **Kubernetes operators**—controllers that extend the Kubernetes API with custom resources (CRDs) representing Kafka clusters, topics, and users. Instead of managing StatefulSets and ConfigMaps directly, you declare your desired Kafka state in a `Kafka` custom resource. The operator continuously reconciles reality to match that desired state.

This shift transformed Kafka on Kubernetes from "possible but painful" to "production-ready with automation." Operators handle:

- **Automated rolling upgrades**: Change the Kafka version in the CR; the operator orchestrates broker restarts one at a time, checking for under-replicated partitions before proceeding to the next.
- **Intelligent scaling**: Adding brokers is as simple as increasing `replicas`. The operator integrates with tools like **Cruise Control** to redistribute partitions automatically when you scale up or gracefully drain partitions before removing a broker.
- **Declarative topic management**: Define topics as `KafkaTopic` CRs with partition count, replication factor, and configs. The operator ensures they exist in Kafka and stay synchronized.
- **User and ACL management**: Create `KafkaUser` CRs with authentication methods (SCRAM-SHA-512, mTLS) and ACL rules. The operator generates credentials (stored in Kubernetes Secrets) and applies ACLs to brokers.
- **TLS certificate automation**: Operators can generate self-signed CAs, broker certificates, and client certificates, handling rotation before expiration—no manual keystore management.
- **Monitoring integration**: Operators deploy JMX exporters as sidecars, exposing Kafka metrics in Prometheus format, often with pre-built Grafana dashboards.

Operators turn Kafka into a **first-class Kubernetes workload**. They bring the same declarative, GitOps-friendly management to Kafka that Deployments bring to stateless apps.

## Comparing Production-Ready Kafka Operators

Three operators stand out as mature, production-tested solutions: **Strimzi**, **Banzai Cloud Koperator**, and **Confluent for Kubernetes**. Each has different strengths, trade-offs, and licensing models.

### Strimzi Kafka Operator (CNCF Sandbox Project)

**What it is**: An open-source operator (Apache 2.0) focused exclusively on running Kafka on Kubernetes. Strimzi is a **CNCF Sandbox project** and powers Red Hat AMQ Streams and IBM Event Streams—strong indicators of production readiness.

**Architecture**: Strimzi includes multiple operators:
- **Cluster Operator**: Manages Kafka brokers and ZooKeeper (or KRaft controllers in newer Kafka versions)
- **Topic Operator**: Reconciles `KafkaTopic` CRs with actual Kafka topics
- **User Operator**: Manages `KafkaUser` CRs, creating SCRAM credentials or mTLS certificates and applying ACLs

**Key Strengths**:

- **Fully open source end-to-end**: Code, container images, and all features are Apache 2.0. No hidden enterprise tiers or licensing fees. If you need commercial support, Red Hat offers it via AMQ Streams (essentially Strimzi with support), but the software itself is free.
- **Comprehensive feature set**: TLS encryption (with automated cert management), multiple SASL auth mechanisms (SCRAM, mTLS, OAuth 2.0), ACL and OPA authorization, rack/zone awareness for high availability, and integration with Cruise Control for partition rebalancing.
- **Wide adoption**: Used by major enterprises and cloud providers. The community is active, documentation is extensive, and the project has graduated from early beta (v0.x) to stable APIs (`v1beta2` and `v1`).
- **High-availability operator**: Supports running 3 operator replicas with leader election, spreading across availability zones for resilience.
- **Ecosystem integration**: Manages not just Kafka, but also Kafka Connect, MirrorMaker 2 (for multi-cluster replication), Kafka Bridge (HTTP API), and Cruise Control—all via CRs.

**Considerations**:

- **StatefulSet-based by default**: Strimzi uses StatefulSets for broker pods, which imposes some constraints (e.g., removing a specific broker mid-cluster requires workarounds). Recent versions introduced **StrimziPodSets** for finer control, addressing some limitations.
- **Semi-automated rebalancing**: Scaling operations are safe, but rebalancing isn't fully automatic by default. You use a `KafkaRebalance` CR to request a rebalance plan from Cruise Control, review it, and approve—giving you control but requiring a manual step.

**When to choose Strimzi**: 
- You want a 100% open-source stack with no licensing concerns.
- You need a battle-tested operator backed by a large community and enterprise vendors.
- You value comprehensive documentation and integrations (Prometheus, Grafana, GitOps tools).
- You're comfortable with Strimzi's opinions (StatefulSets, semi-automated rebalancing) or willing to use newer features (StrimziPodSets, KafkaNodePools) for more flexibility.

**Licensing**: Apache 2.0 for everything—source code, container images, all features. No commercial restrictions.

---

### Banzai Cloud Koperator (Cisco-backed)

**What it is**: An open-source Kafka operator (Apache 2.0) originally from Banzai Cloud, now maintained by **Cisco Outshift**. Koperator takes a different architectural approach from Strimzi, avoiding StatefulSets entirely in favor of managing individual broker Pods, ConfigMaps, and PVCs directly.

**Architecture**: Koperator provides similar CRDs (`KafkaCluster`, `KafkaTopic`, `KafkaUser`) but manages resources at a lower level, giving fine-grained per-broker control.

**Key Strengths**:

- **Flexible per-broker configuration**: Each broker can have custom configs, resource allocations, or even different Kafka versions (rare, but possible). This is difficult with StatefulSets where all pods share a template.
- **Advanced storage management**: Supports **JBOD** (multiple disks per broker) natively. You can add a disk to a specific broker, and Koperator will trigger Cruise Control to redistribute partitions to the new volume—enabling capacity expansion without adding brokers.
- **Automated partition rebalancing**: Koperator integrates deeply with **Cruise Control**. When you scale up or down, it automatically creates `CruiseControlOperation` CRs to move partitions, making rebalancing more hands-off than Strimzi's approach.
- **Alert-driven automation**: Koperator can react to Prometheus alerts (e.g., high disk usage) by automatically adding brokers or storage—a powerful self-healing capability for dynamic workloads.
- **Smart external access**: Uses an Envoy-based load balancer to expose the entire cluster via a single LoadBalancer Service (instead of one per broker), reducing cloud load balancer costs.
- **Fully open source**: Like Strimzi, Koperator is Apache 2.0 with no proprietary components. All features (topic/user management, autoscaling, advanced rebalancing) are included.

**Considerations**:

- **Less widespread adoption**: While Koperator is used in production (notably by Adobe and within Cisco), it's not as universally known as Strimzi. The community is smaller, though active.
- **Complexity trade-off**: Koperator's flexibility and rich feature set (alert-driven autoscaling, per-broker management) come with more configuration surface area. Simple use cases might not need these capabilities.

**When to choose Koperator**:
- You need fine-grained operational control (per-broker configs, selective broker removal).
- You're managing large, dynamic Kafka clusters where automated rebalancing and alert-driven scaling justify the added complexity.
- You want to avoid StatefulSet constraints and prefer a design optimized for Kafka's specific needs.
- You value open-source but want a project with enterprise backing (Cisco).

**Licensing**: Apache 2.0 for code and images. No commercial features hidden behind paywalls.

---

### Confluent for Kubernetes (CFK)

**What it is**: A commercial operator from **Confluent** (the company founded by Kafka's original creators) that deploys the full **Confluent Platform**—not just Apache Kafka, but also Schema Registry, Kafka Connect, ksqlDB, Control Center (a management GUI), and enterprise features like RBAC and self-balancing clusters.

**Key Strengths**:

- **Enterprise polish**: CFK automates complex workflows with zero-downtime rolling upgrades, integrated monitoring via Control Center, and deep security integrations (Confluent RBAC, Audit Logs, OAuth).
- **Comprehensive platform management**: Manage the entire Confluent ecosystem via CRs—Kafka brokers, ZooKeeper/KRaft, Schema Registry, Connect, ksqlDB—all wired together and tested on major Kubernetes distributions.
- **Self-balancing clusters**: An enterprise feature that automatically rebalances partitions when brokers are added/removed or load shifts, eliminating manual intervention (similar to Cruise Control but proprietary).
- **Official Confluent support**: Backed by the team that created Kafka, with support plans and tight integration with Confluent Cloud for hybrid deployments.

**Considerations**:

- **Commercial licensing**: CFK uses the **Confluent Community License** (not OSI-approved open source). You can deploy and run CFK for a **30-day evaluation**, but sustained production use requires purchasing a license key. Without a key, enterprise features (RBAC, Audit Logs, Self-Balancing, Control Center) are unavailable or time-limited.
- **Not open source**: While the core Apache Kafka within Confluent Platform remains Apache 2.0, the operator itself and many platform components are source-available but not truly open source. You can't fork CFK or modify it freely as you can with Strimzi or Koperator.
- **Cost**: Confluent Platform subscriptions are the most expensive option, justified by enterprise features and support. For startups or organizations committed to open-source stacks, this might be prohibitive.
- **Vendor lock-in risk**: Using Confluent-specific features (RBAC, Schema Registry with governance, ksqlDB) creates tighter coupling to their ecosystem. Migrating away becomes harder.

**When to choose CFK**:
- You need enterprise features like RBAC, Audit Logs, or Self-Balancing and are willing to pay for them.
- You want official support from Kafka's creators and prioritize a polished, GUI-driven management experience (Control Center).
- Your organization already uses Confluent Cloud or Platform and wants consistency across environments.
- Licensing costs aren't a primary concern, and you value turnkey solutions over open-source flexibility.

**Licensing**: Confluent Community License (source-available, non-OSI). Requires paid subscription for production use with enterprise features. Container images also fall under Confluent's licensing.

---

## Operator Feature Comparison

| Feature | Strimzi | Koperator | Confluent (CFK) |
|---------|---------|-----------|-----------------|
| **License (Code)** | Apache 2.0 | Apache 2.0 | Confluent Community License |
| **License (Images)** | Apache 2.0 | Apache 2.0 | Confluent Community License |
| **Production Readiness** | ✅ High (CNCF, Red Hat, IBM) | ✅ High (Cisco, Adobe) | ✅ High (Confluent) |
| **Cost** | Free (support via Red Hat optional) | Free (support via Cisco optional) | Paid license required for production |
| **Topic Management (CRDs)** | ✅ KafkaTopic | ✅ KafkaTopic | ✅ KafkaTopic |
| **User/ACL Management** | ✅ KafkaUser (SCRAM, mTLS, ACLs) | ✅ KafkaUser (SCRAM, mTLS, ACLs) | ❌ No simple ACL CRD (RBAC via CRs requires license) |
| **TLS Automation** | ✅ Full (self-signed CA + certs) | ✅ Full (self-signed + cert-manager) | ✅ Full (with cert-manager integration) |
| **Cruise Control Integration** | ✅ Via KafkaRebalance CR (semi-auto) | ✅ Fully automated on scale events | ❌ Uses proprietary Self-Balancing (enterprise) |
| **Per-Broker Customization** | ⚠️ Limited (KafkaNodePools helps) | ✅ Full (no StatefulSet constraints) | ⚠️ Via broker groups |
| **Monitoring** | ✅ JMX Exporter, Grafana dashboards | ✅ JMX metrics, Prometheus alerts | ✅ Prometheus + Control Center GUI (licensed) |
| **ZooKeeper/KRaft** | ✅ Both supported | ✅ Both supported | ✅ Both supported |
| **Operator HA** | ✅ Multi-replica with leader election | ✅ Multi-replica with leader election | ✅ Multi-replica with leader election |
| **Community Size** | ⭐⭐⭐⭐⭐ Very large (CNCF) | ⭐⭐⭐ Growing (Cisco-backed) | ⭐⭐⭐⭐ Large (Confluent commercial) |
| **Deployment Methods** | Helm, YAML, OLM (OpenShift) | Helm, YAML, OLM | Helm (official), OLM (Red Hat Certified) |
| **Ecosystem Components** | Kafka, Connect, MirrorMaker, Bridge | Kafka, Cruise Control | Full Confluent Platform (Kafka, Schema Registry, Connect, ksqlDB, Control Center) |

---

## Planton's Choice: Strimzi as the Default

Planton defaults to **Strimzi** for Kafka operator deployments, aligning with our philosophy of open-source, production-ready infrastructure components. Here's why:

### Licensing Clarity and Openness

Strimzi is **Apache 2.0 end-to-end**—code, container images, and every feature. There are no hidden enterprise tiers, no 30-day trial periods, and no usage restrictions. For an open-source IaC framework like Planton, this matters. Users can deploy Kafka confidently knowing they won't hit licensing walls in production or need to negotiate commercial agreements.

While Koperator is also fully open source (and an excellent choice for advanced use cases), Strimzi's **CNCF affiliation** and backing by Red Hat and IBM provide additional assurance of long-term maintenance and community investment.

### Production Maturity and Adoption

Strimzi powers **Red Hat AMQ Streams** and **IBM Event Streams**—commercially supported products built directly on Strimzi. This isn't a side project; it's infrastructure trusted by enterprises running Kafka at scale in production. The community is large, documentation is comprehensive, and edge cases are well-tested.

When you deploy Kafka via Planton, you're standing on the shoulders of thousands of production deployments across industries.

### Comprehensive Feature Set with Sane Defaults

Strimzi supports everything most teams need:

- **Declarative cluster management**: Define brokers, ZooKeeper/KRaft, storage, and listeners in a single `Kafka` CR.
- **Security out of the box**: TLS encryption with automated cert management, SCRAM-SHA-512 and mTLS authentication, ACL authorization, and even OAuth 2.0 integration.
- **Topic and user management**: `KafkaTopic` and `KafkaUser` CRs make Kafka feel Kubernetes-native. GitOps teams can version-control topics and credentials alongside application configs.
- **Monitoring integration**: JMX Exporter sidecars expose metrics to Prometheus; Strimzi provides pre-built Grafana dashboards for cluster health, throughput, consumer lag, and more.
- **Ecosystem components**: Deploy Kafka Connect for streaming integrations, MirrorMaker 2 for multi-cluster replication, and Kafka Bridge for HTTP API access—all managed by the same operator.

Strimzi's defaults are **production-oriented**. For example, it uses pod anti-affinity to spread brokers across nodes (or zones) automatically, applies security contexts to run as non-root, and defaults to safe replication factors for internal topics.

### Operator as Infrastructure, Not Configuration

Planton's `KubernetesStrimziKafkaOperator` resource deploys **the operator itself**—not a Kafka cluster. This separation is intentional. The operator is infrastructure: install it once per cluster, configure it for high availability (multiple replicas with leader election), allocate resources, and let it run.

The Strimzi operator is stable, lightweight, and designed to watch multiple namespaces. One operator instance can manage many Kafka clusters across different namespaces or teams, enabling multi-tenancy without operator proliferation.

Once the operator is running, you define `Kafka` CRs (via Planton's `KafkaKubernetes` resource or directly) to create actual clusters. This clean separation simplifies operations: upgrading the operator is distinct from upgrading Kafka brokers, and different teams can manage Kafka clusters without touching the operator's deployment.

### GitOps-Friendly by Design

Every Strimzi resource—operators, Kafka clusters, topics, users—is defined declaratively via YAML. This fits naturally into GitOps workflows with ArgoCD, Flux, or Planton's own IaC approach.

Changes to Kafka clusters happen by committing CR updates to Git. The operator reconciles the changes safely, applying rolling restarts only when necessary and checking cluster health before proceeding. This is **infrastructure as code** at its best: auditable, version-controlled, and reproducible across environments.

### When You Might Choose Differently

Strimzi is the default, but Planton is flexible:

- **Choose Koperator** if you need per-broker customization, automated alert-driven scaling, or advanced storage management (JBOD with dynamic disk additions). Koperator's design shines in dynamic, large-scale deployments where operational flexibility justifies complexity.
- **Choose Confluent (CFK)** if you're using Confluent Platform commercially, need enterprise features like RBAC and Self-Balancing Clusters, and have budget for licensing. CFK is polished, GUI-driven, and backed by Kafka's creators.

For most teams building on open-source foundations, Strimzi hits the sweet spot: mature, feature-rich, community-driven, and free.

---

## Understanding Operator vs. Cluster Configuration

A common point of confusion: **What does the operator manage, and what does a Kafka cluster need?**

### Operator Configuration (Day-0 Setup)

The **operator deployment** (`KubernetesStrimziKafkaOperator` in Planton) is about installing the controller itself. Configuration includes:

- **Target namespace and watch scope**: Where the operator runs and which namespaces it monitors for Kafka CRs.
- **Operator image version**: Pinning to a stable Strimzi release (e.g., `0.34.0`) for reproducibility.
- **Replicas for HA**: Running 2-3 operator pods with leader election for resilience.
- **Resource allocation**: CPU and memory for the operator container (typically modest—e.g., 100m CPU, 256Mi memory—but tunable for clusters managing many resources).
- **RBAC setup**: ClusterRoles and ServiceAccounts granting the operator permissions to manage StatefulSets, Pods, PVCs, Services, and CRDs.

These settings are largely **set-and-forget**. Once the operator is deployed, you rarely change it except during upgrades.

### Kafka Cluster Configuration (Day-2 Operations)

The **Kafka cluster** itself is defined via a `Kafka` CR (or Planton's `KafkaKubernetes` resource). This is where most operational activity happens:

- **Broker count and resources**: Number of Kafka broker pods, CPU/memory allocations, JVM heap settings.
- **Storage**: Persistent volumes (size, storage class, JBOD if needed) or ephemeral storage for dev/test.
- **Kafka version**: Which Kafka version to run (e.g., `3.5.0`).
- **Listeners**: How clients connect—internal (ClusterIP), external (LoadBalancer, NodePort, Ingress), TLS encryption, authentication (SCRAM, mTLS).
- **Security**: TLS certificates (self-signed or from cert-manager), SASL authentication, ACL authorization.
- **ZooKeeper or KRaft**: Whether to deploy a ZooKeeper ensemble (for Kafka <3.x or legacy mode) or use KRaft controllers (Kafka 3.x+ metadata quorum).
- **Configuration tuning**: Kafka broker configs like retention periods, replication factors, min.insync.replicas, etc.

This is the **living configuration**—teams scale broker counts, update TLS settings, tune configs, and deploy new topics/users here. The operator watches these CRs and reconciles Kubernetes resources (Pods, StatefulSets, Services, Secrets) to match the desired state.

### The Boundary: Why It Matters

Separating operator config from cluster config enables:

- **Multi-tenancy**: One operator can manage Kafka clusters for multiple teams or projects, each in its own namespace with its own CRs. Teams manage their clusters without touching the operator.
- **Role-based access**: Cluster admins manage the operator (install, upgrade, RBAC). Application teams manage `Kafka`, `KafkaTopic`, and `KafkaUser` CRs for their deployments.
- **Independent upgrades**: Upgrade the operator to get bug fixes or new features without changing Kafka clusters. Upgrade Kafka clusters (change version in CR) without redeploying the operator.

In Planton terms:

- **`KubernetesStrimziKafkaOperator`** = "Ensure the Strimzi operator is running on this cluster with these settings."
- **`KafkaKubernetes`** (or direct `Kafka` CRs) = "Deploy a Kafka cluster with X brokers, Y storage, Z security settings."

This mirrors the distinction between "installing Docker" (enabling containerization) and "running containers" (deploying applications).

---

## Essential Configuration: The 80/20 for Production

Not every configuration knob matters equally. Here's what most teams **actually set** versus advanced options they rarely touch.

### Essential Operator Settings (80% of deployments)

When deploying `KubernetesStrimziKafkaOperator`:

- **Replicas**: `1` for dev/test, `2-3` for production (with leader election for HA).
- **Image version**: Pin to a stable Strimzi release (e.g., `quay.io/strimzi/operator:0.34.0`).
- **Resources**: Default `50m` CPU / `100Mi` memory is often fine; increase to `100m` / `256Mi` for clusters managing many CRs.
- **Watch namespaces**: Often left to watch all namespaces (default) or restricted to specific ones for isolation.

That's it. Most teams don't touch image pull secrets (unless air-gapped), custom CRDs, or advanced feature gates.

### Essential Kafka Cluster Settings (80% of deployments)

When defining a `Kafka` CR:

**Minimum for dev/test**:
- `replicas: 1` (single broker, non-HA)
- `version: 3.5.0` (or latest stable)
- `storage: ephemeral` (no persistence, fast teardown)
- `listeners: plaintext` (no TLS/auth for simplicity)
- `zookeeper.replicas: 1` (or use KRaft mode with one controller)

**Production essentials**:
- `replicas: 3` (minimum for HA and replication factor 3)
- `version: 3.5.0` (explicit, tested version)
- `storage: persistent-claim` with `size: 500Gi` and `class: fast-ssd`
- `listeners`:
  - Internal TLS listener for inter-broker and client connections
  - External LoadBalancer listener (if needed) with SCRAM-SHA-512 authentication
- `zookeeper.replicas: 3` with persistent storage (or 3 KRaft controllers)
- **Config tuning**:
  - `default.replication.factor: 3`
  - `min.insync.replicas: 2`
  - `unclean.leader.election.enable: false`
  - `auto.create.topics.enable: false` (explicit topic creation)
- **Pod anti-affinity** to spread brokers across availability zones (Strimzi defaults to this)
- **Resource requests/limits**: Allocate sufficient memory (e.g., 8Gi container with 4Gi heap for Kafka's JVM, leaving 4Gi for page cache)

### Advanced Settings (Rarely Needed)

- Custom JVM options (GC tuning, heap dumps)
- Rack/zone awareness labels (useful in multi-AZ, but Strimzi handles it via pod topology spread by default)
- Custom affinity/tolerations (for dedicated Kafka nodes)
- Feature gates (enabling experimental features)
- External CA certificates (if not using Strimzi's generated ones)
- Cruise Control tuning (goals, thresholds—most use defaults)
- JBOD with multiple disks per broker (advanced capacity management)

The philosophy: **Sensible defaults for 80% of use cases, escape hatches for the 20% who need them.**

---

## Production Best Practices

Operating Kafka on Kubernetes with an operator automates much, but these practices ensure stability and performance:

### High Availability (Operator and Kafka)

- **Operator HA**: Run 3 replicas of Strimzi's operator across availability zones. Use pod anti-affinity to spread them. If one operator pod fails, another takes over without Kafka clusters becoming unmanaged.
- **Kafka HA**: Deploy at least 3 brokers with `min.insync.replicas: 2` and replication factor 3. Spread brokers across zones using Strimzi's `topologyKey` in pod anti-affinity (e.g., `topology.kubernetes.io/zone`).
- **ZooKeeper HA** (if not using KRaft): 3 ZooKeeper nodes with persistent storage. ZooKeeper needs fast disks—use SSDs or local persistent volumes.

### Capacity Planning and Resource Allocation

- **Size brokers appropriately**: Kafka benefits from ample memory (for JVM heap and OS page cache). A typical config: 8Gi total memory with 4Gi heap, leaving 4Gi for page cache. Allocate enough CPU for compression and replication (e.g., 2-4 CPU cores).
- **Monitor disk usage**: Kafka doesn't autoscale storage. Set alerts for PVC usage >85%. If disks fill, brokers stall. Scale out (add brokers) rather than trying to expand PVCs (Kafka won't auto-rebalance data to new space).
- **Operator resources**: Usually modest (0.5 CPU, 512Mi memory), but increase if managing dozens of clusters or thousands of topics.

### Security Hardening

- **Enable TLS everywhere**: Encrypt inter-broker and client connections. Strimzi automates cert generation and rotation. There's no excuse not to use TLS in production.
- **Require authentication**: Use SCRAM-SHA-512 or mTLS for clients. Create `KafkaUser` CRs for each app; Strimzi generates credentials and stores them in Secrets.
- **Apply ACLs**: Grant least-privilege access per user. Use `KafkaUser` CRs to define topic-level permissions. Only the app that owns a topic should write to it.
- **Network policies**: Restrict pod-to-pod traffic. Only allow apps that need Kafka to connect to broker ports (9092/9093). Deny everything else.
- **Secure operator access**: Limit who can create/modify `Kafka` CRs (Kubernetes RBAC). The operator has broad permissions—restrict its namespace and CRD access to trusted admins.

### Observability (Monitoring and Logging)

- **Prometheus and Grafana**: Strimzi deploys JMX Exporter sidecars. Configure Prometheus to scrape Kafka metrics. Use Strimzi's pre-built Grafana dashboards for broker health, throughput, under-replicated partitions, consumer lag, and JVM metrics.
- **Key alerts**:
  - Under-replicated partitions >0 for >5 minutes (indicates broker issues or rebalancing)
  - Broker pod not ready (hardware failure or config problem)
  - Disk usage >85% (approaching capacity limits)
  - Consumer lag spiking (indicates slow consumers or broker overload)
- **Centralized logging**: Aggregate Kafka broker logs and operator logs (EFK stack, Datadog, Splunk, or cloud logging). Operator logs reveal reconciliation errors; broker logs show client connection issues, replication lag, etc.
- **Kubernetes events**: Check `kubectl get events` or describe `Kafka` CRs to see operator-reported issues (e.g., "Reconciliation failed: PVC not bound").

### Regular Maintenance

- **Upgrade operators carefully**: Follow Strimzi's upgrade guide (usually: upgrade CRDs first, then operator, then optionally Kafka clusters). Test in staging. Read release notes for breaking changes.
- **Upgrade Kafka brokers incrementally**: Change `spec.kafka.version` in the CR. Strimzi orchestrates rolling restarts, checking cluster health between brokers. Don't skip multiple Kafka versions at once.
- **Backup strategy**: Kafka isn't typically backed up via snapshots (it's a streaming system), but use **MirrorMaker 2** to replicate critical topics to a DR cluster. Or configure tiered storage (if supported) to offload old segments to object storage.
- **Topic lifecycle management**: Disable `auto.create.topics.enable`. Define topics explicitly via `KafkaTopic` CRs with appropriate retention and cleanup policies. Avoid infinite retention unless you have infinite storage.

### Common Pitfalls to Avoid

- **Advertised listeners misconfiguration**: Clients must be able to resolve and reach the advertised addresses Strimzi sets. Test external listeners from client networks before going live.
- **Ignoring ZooKeeper** (if used): Monitor ZooKeeper metrics and health. ZooKeeper issues cascade to Kafka. Ensure ZooKeeper has fast, stable storage.
- **Running out of disk**: No autoscaling for storage. Set alerts and add brokers proactively (or expand PVCs if your storage class supports it, then manually rebalance).
- **Not testing rolling restarts**: Ensure your workload can tolerate one broker restarting at a time. Test configuration changes in staging before applying to production.
- **Operator CRD drift**: After operator upgrades, update Kafka CRs to new apiVersions if needed. Strimzi logs deprecation warnings—don't ignore them.

---

## Conclusion: From Skepticism to Solid Foundation

The journey from "Kafka doesn't belong on Kubernetes" to "Kafka operators make Kubernetes the preferred platform" illustrates a broader shift in cloud-native infrastructure. Stateful workloads aren't inherently incompatible with Kubernetes—they require **purpose-built automation** that encodes operational expertise.

Kafka operators—especially Strimzi—provide that automation. They transform Kafka from a complex, script-dependent system into a **declarative, self-healing workload** that benefits from Kubernetes' strengths: declarative APIs, reconciliation loops, RBAC, GitOps workflows, and cloud portability.

Planton's `KubernetesStrimziKafkaOperator` resource makes deploying Strimzi straightforward: define the operator configuration, deploy it via Pulumi-backed IaC, and let the operator handle the rest. From there, Kafka clusters, topics, and users become Kubernetes-native resources—manageable via YAML, version-controlled in Git, and automatically reconciled to their desired state.

For teams building streaming platforms, event-driven architectures, or any workload that depends on Kafka, operators are no longer optional—they're the **production-ready path forward**.

