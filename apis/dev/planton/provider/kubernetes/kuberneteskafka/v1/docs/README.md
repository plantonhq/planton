# Apache Kafka on Kubernetes: Deployment Methods and Production Patterns

## Introduction: The Evolution from Anti-Pattern to Production-Ready

For years, conventional wisdom held that running Apache Kafka on Kubernetes was an anti-pattern. Kafka, with its complex cluster coordination, persistent storage requirements, and stateful broker identities, seemed fundamentally incompatible with Kubernetes' declarative, ephemeral nature. Yet today, Kafka on Kubernetes has become not just viable, but the preferred deployment method for modern infrastructure teams.

What changed? The answer lies in the maturation of the **Operator Pattern**—a Kubernetes extension that embeds domain-specific operational knowledge directly into the platform. A Kafka operator doesn't just deploy containers; it acts as a tireless cluster administrator, understanding broker rebalancing, graceful scaling, rolling upgrades, and failure recovery. This architectural evolution transformed Kafka on Kubernetes from a risky experiment into a production-standard deployment model.

This document explores the deployment landscape for Apache Kafka on Kubernetes, examining why simple approaches fail, how operators solve these challenges, and what production-ready solutions exist today. Whether you're deploying your first development cluster or architecting a multi-region production system, understanding this progression from anti-patterns to operators is essential.

## The Deployment Maturity Spectrum

### Level 0: The Anti-Pattern (Kubernetes Deployments)

**What it is:** Attempting to run Kafka using standard Kubernetes Deployment controllers.

**Why it fails immediately:** Kafka brokers must advertise their network addresses to clients and other brokers. In a Deployment, each pod receives a transient, randomly-generated hostname (e.g., `kafka-6799c65d58-f6tbt`). This hostname is not DNS-resolvable within the cluster. When a client connects to the cluster's bootstrap service, it receives broker addresses it cannot reach, causing universal connection failures.

**Verdict:** Not viable even for development. Kafka requires stable network identities that Deployments cannot provide.

### Level 1: The Foundation (StatefulSets Without Operators)

**What it is:** Deploying Kafka using StatefulSet controllers, either manually or via basic Helm charts.

**What it solves:** StatefulSets provide the two foundational requirements Kafka needs:

1. **Stable Network Identity:** Pods are created with predictable, ordinal hostnames (`kafka-0`, `kafka-1`, `kafka-2`) that are DNS-resolvable via a headless Service.
2. **Stable Persistent Storage:** Each pod receives its own persistent volume claim (PVC), ensuring topic data survives pod restarts.

**What it doesn't solve:** StatefulSets are generic tools unaware of Kafka's operational requirements. Critical "Day 2" operations fail:

- **Scaling:** When scaling down, StatefulSets always terminate the highest-ordinal pod. A human operator might need to decommission a specific failing broker, requiring careful partition reassignment before shutdown. StatefulSets simply kill the pod, triggering uncontrolled leadership elections and under-replicated partitions.

- **Configuration Management:** StatefulSets apply a uniform template to all pods. They cannot manage per-broker configurations or heterogeneous resource allocations.

- **Rolling Upgrades:** A basic StatefulSet rolling update doesn't check cluster health, partition replication status, or leadership distribution before proceeding. This creates high risk of downtime or data loss.

**Verdict:** Suitable only for throwaway development environments. Not production-ready due to lack of Kafka-aware lifecycle management.

### Level 2: The "Day 1" Trap (Helm Charts)

**What it is:** Using Helm charts (like Bitnami's popular Kafka chart) to template and install Kafka StatefulSets.

**Value added:** Helm provides packaging, versioning, and parameterized installation. It simplifies the initial deployment with reasonable defaults.

**Critical limitation:** Helm is an installation tool, not a lifecycle management system. It doesn't track the state of resources after deployment.

**The PVC gotcha:** When you run `helm uninstall`, the StatefulSet is deleted, but persistent volume claims created by `volumeClaimTemplates` are left orphaned. This creates:
- Resource leaks (storage continues accruing costs)
- Data loss risks (reinstalling the chart may not reattach to old volumes)
- Manual cleanup burden (users must discover and delete PVCs separately)

**How mature charts evolved:** Modern Kafka Helm charts (from Strimzi, Confluent) no longer deploy Kafka directly. Instead, they install a Kubernetes operator. The actual Kafka cluster is then defined in a separate Custom Resource that the operator manages. This hybrid approach uses Helm for operator installation, but delegates cluster lifecycle to the operator pattern.

**Verdict:** Acceptable for installing operators, but insufficient for managing production Kafka clusters directly.

### Level 3: The Production Solution (Kubernetes Operators)

**What it is:** A Kubernetes Operator extends the Kubernetes API with custom controllers that watch Custom Resource Definitions (CRDs) like `kind: Kafka`. The operator embeds all the domain-specific logic of a human Kafka administrator.

**How it works:** When you declare a `Kafka` custom resource specifying 5 replicas, the operator:

1. Provisions persistent volumes for each broker
2. Generates unique broker configurations with sequential broker IDs
3. Launches broker pods with proper dependencies (e.g., waiting for storage to attach)
4. Monitors cluster health and partition replication status
5. When scaling up, adds new brokers and optionally triggers partition rebalancing
6. During upgrades, performs rolling restarts one broker at a time, waiting for each to fully sync before proceeding
7. Handles failure recovery by detecting pod failures and orchestrating safe restarts

**What this enables:**

- **Declarative Operations:** Scaling from 3 to 5 brokers is changing a number in YAML and applying it. The operator handles the complexity.
- **Kafka-Aware Upgrades:** Version upgrades are safe because the operator verifies cluster health at each step.
- **Graceful Scaling:** The operator can decommission specific brokers (not just the highest ordinal) after reassigning their partitions.
- **GitOps Compatibility:** Your entire Kafka infrastructure becomes version-controlled YAML that operators reconcile into reality.

**Verdict:** The only production-ready approach. Operators are mandatory for managing Kafka's "Day 2" operational complexity.

## Operator Comparison: Choosing Your Production Foundation

The operator pattern is essential, but which operator should you use? The ecosystem has consolidated around a few mature options, each with different licensing models, feature sets, and target use cases.

### Licensing: The Critical Decision Factor

For open-source infrastructure frameworks, licensing determines long-term viability. Not all "open-source" operators are created equal.

#### 100% Free and Open Source (FOSS)

**Strimzi (CNCF Sandbox Project)**

- **License:** Apache 2.0 for all code and container images
- **Governance:** Cloud Native Computing Foundation (vendor-neutral)
- **Container Images:** Freely available, no usage restrictions, no trial periods
- **Production Use:** Unlimited, including commercial SaaS offerings

Strimzi represents a genuine open-source project where the entire stack—operator code, broker images, and all supporting components—is freely usable without legal constraints.

#### "Open Core" Models (Use with Caution)

**Confluent Operator (Confluent for Kubernetes / CFK)**

- **Operator Code:** Apache 2.0
- **Runtime Components:** Licensed under Confluent Community License or Confluent Enterprise License
- **Key Constraint:** The Confluent Community License is **not an OSI-approved open-source license**. It explicitly prohibits creating a "SaaS offering" of the code, creating legal ambiguity for infrastructure platforms.
- **Enterprise Features:** Many production features and even the operator itself require a paid enterprise license key after a 30-day trial.

This "open core" model means the operator may be open source, but the software it deploys is not. This creates vendor lock-in and potential licensing complications.

#### Archived/Unmaintained Projects

**Banzai Cloud Koperator**

- **Status:** GitHub repository archived
- **Historical Note:** Featured an innovative architecture avoiding StatefulSets in favor of direct Pod management, but is no longer maintained
- **Recommendation:** Do not use for new deployments

**IBM Event Streams**

- **Nature:** Fully commercial, licensed product
- **Context:** Part of IBM Cloud Pak for Integration
- **Licensing:** Tied to IBM's specific licensing model
- **Use Case:** Only suitable if already in the IBM ecosystem

### Feature Comparison Matrix

| Feature | Strimzi (CNCF) | Confluent Operator | Banzai Koperator | IBM Event Streams |
|---------|----------------|-------------------|------------------|-------------------|
| **Code License** | **Apache 2.0** | Apache 2.0 | Apache 2.0 | Proprietary |
| **Runtime License** | **Apache 2.0** | **Confluent Enterprise/Community** | Apache 2.0 | Proprietary |
| **Status** | **Active, CNCF** | Active, Enterprise | **Archived** | Active, Commercial |
| **Metadata Management** | **KRaft-Only** (modern) | KRaft & ZooKeeper | KRaft & ZooKeeper | KRaft Support |
| **Declarative Topics** | **Yes** (`KafkaTopic` CRD) | Yes | Yes | Yes |
| **Declarative Users** | **Yes** (`KafkaUser` CRD) | Yes | Yes | Yes |
| **Schema Registry** | **No** (by design; use Apicurio) | **Yes** (bundled Confluent SR) | Yes | Yes (Apicurio) |
| **Monitoring** | **Excellent** (native Prometheus) | Yes (Confluent metrics) | Yes | Yes |
| **Security** | **Excellent** (TLS, SASL, ACLs) | Excellent | Good | Excellent |
| **Key Strength** | 100% FOSS, CNCF-backed, production-proven | Unified Confluent Platform integration | Innovative Pod architecture | IBM ecosystem integration |
| **Key Weakness** | No bundled schema registry | **Proprietary licensing** | **Archived/unmaintained** | Vendor lock-in |

### The Recommendation: Strimzi

For an open-source infrastructure-as-code framework like Planton, **Strimzi is the clear choice**:

1. **Licensing Alignment:** 100% Apache 2.0, no restrictions, fully FOSS
2. **Community Governance:** CNCF sandbox project with vendor-neutral oversight
3. **Production Maturity:** Widely deployed in production environments across industries
4. **Modern Architecture:** KRaft-native (ZooKeeper support removed), reflecting Apache Kafka's evolution
5. **Ecosystem Integration:** Works seamlessly with other open-source tools (Apicurio Registry for schemas, Prometheus/Grafana for monitoring)

Strimzi's one "gap"—the lack of a bundled schema registry—is a deliberate design choice to keep the project focused. This gap is easily filled by Apicurio Registry, another Apache 2.0-licensed project with its own Kubernetes operator.

## Architectural Decisions for Production Deployments

### ZooKeeper vs. KRaft: The New Reality

**This is no longer a choice.** Apache Kafka deprecated ZooKeeper as of version 3.5 and **completely removed it in Kafka 4.0**. All new production deployments must use **KRaft** (Kafka Raft Metadata mode).

#### What is KRaft?

KRaft is a fundamental re-architecture that embeds a Raft-based consensus quorum directly inside Kafka, eliminating the dependency on external ZooKeeper ensembles. Instead of brokers + ZooKeeper, you now have:

- **Controller Nodes:** Form a Raft quorum to manage cluster metadata
- **Broker Nodes:** Handle data streaming (producer/consumer traffic)
- **Combined Nodes:** (Dev/testing) Nodes that serve both roles

#### Why KRaft Won

1. **Operational Simplicity:** One system to deploy, monitor, secure, and upgrade instead of two
2. **Massive Scalability:** ZooKeeper was a bottleneck limiting clusters to tens of thousands of partitions. KRaft supports millions.
3. **Faster Recovery:** Controller failover is nearly instantaneous (sub-second) compared to ZooKeeper-based failovers that could take tens of seconds.
4. **Reduced Complexity:** No separate ZooKeeper connection strings, ACLs, or TLS configurations

#### Strimzi's KRaft Implementation: KafkaNodePool

Modern Strimzi is **KRaft-only**. The core abstraction is the `KafkaNodePool` Custom Resource, which allows you to define "pools" of nodes with specific roles:

**Production Pattern (Separated Roles):**

```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaNodePool
metadata:
  name: controller
spec:
  replicas: 3
  roles: [controller]  # Pure metadata management
  storage:
    type: persistent-claim
    size: 100Gi
  resources:
    requests: { cpu: "2", memory: "4Gi" }
---
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaNodePool
metadata:
  name: broker
spec:
  replicas: 5
  roles: [broker]  # Pure data streaming
  storage:
    type: persistent-claim
    size: 1000Gi
  resources:
    requests: { cpu: "4", memory: "8Gi" }
```

**Development Pattern (Combined Roles):**

```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaNodePool
metadata:
  name: combined
spec:
  replicas: 1
  roles: [controller, broker]  # Dual-purpose for minimal footprint
  storage:
    type: persistent-claim
    size: 10Gi
  resources:
    requests: { cpu: "1", memory: "2Gi" }
```

### Storage Architecture

#### Persistent Volume Requirements

Kafka is storage-intensive. All topic data must be backed by high-performance, low-latency block storage. Your Kubernetes cluster must have a `StorageClass` configured to provision:

- **AWS:** gp3 volumes (balanced performance/cost)
- **Azure:** Premium SSD
- **GCP:** SSD persistent disks
- **On-premises:** Local SSDs or high-performance SAN

#### Storage Types in Strimzi

1. **ephemeral:** Uses `emptyDir`. All data lost on pod restart. Development only.
2. **persistent-claim:** Standard production option. One PVC per broker.
3. **jbod (Just a Bunch of Disks):** Production best practice. Multiple PVCs per broker, striping data across them for increased throughput.

Example JBOD configuration:

```yaml
storage:
  type: jbod
  volumes:
    - id: 0
      type: persistent-claim
      size: 500Gi
      deleteClaim: false
    - id: 1
      type: persistent-claim
      size: 500Gi
      deleteClaim: false
```

This provides 1TB total capacity per broker with parallelized I/O.

#### Live Volume Expansion

Strimzi automates disk expansion. If your `StorageClass` has `allowVolumeExpansion: true`:

1. Edit the `KafkaNodePool` resource, increasing `storage.size` (e.g., 500Gi → 1000Gi)
2. Strimzi requests Kubernetes resize the live PVC
3. Once the storage provider completes the expansion, Strimzi performs a graceful rolling restart of each broker to allow filesystem expansion
4. Brokers resume with increased capacity, no data loss

### Networking Architecture

#### Internal Cluster Communication

Strimzi automatically configures internal listeners for cluster operations (KRaft consensus, broker-to-broker replication). This communication is always secured with mutual TLS (mTLS) and abstracted from users.

#### External Access Patterns

External access is declared via the `spec.kafka.listeners` array in the `Kafka` CRD:

**Option 1: Internal Only (ClusterIP)**

```yaml
listeners:
  - name: plain
    port: 9092
    type: internal
    tls: false
```

Kafka is only accessible from within the Kubernetes cluster. Suitable for microservices architectures where all consumers/producers run in-cluster.

**Option 2: NodePort (Development)**

```yaml
listeners:
  - name: external
    port: 9094
    type: nodeport
    tls: true
```

Exposes Kafka on high-numbered ports (30000-32767) on every Kubernetes node. Clients connect to any node IP. Useful for development but not recommended for production (requires firewall rules, exposes node IPs).

**Option 3: LoadBalancer (Production)**

```yaml
listeners:
  - name: external
    port: 9094
    type: loadbalancer
    tls: true
```

Strimzi provisions:
- One LoadBalancer service for the bootstrap endpoint
- One LoadBalancer per broker (for direct broker connections)

This provides stable, DNS-based external access at the cost of multiple cloud load balancers.

**Option 4: Ingress (HTTP Only)**

The `type: ingress` listener is **only for HTTP-based components** like the Kafka REST Proxy or Kafka Bridge. It does **not work** for the native Kafka protocol (port 9092), which is TCP-based. Standard L7 Ingress controllers expect HTTP traffic and will not route Kafka broker connections correctly.

#### TLS: End-to-End Encryption

The production best practice is **end-to-end encryption**. Set `tls: true` on external listeners:

```yaml
listeners:
  - name: external
    port: 9094
    type: loadbalancer
    tls: true
```

Strimzi's Cluster Operator acts as a Certificate Authority (CA), automatically:
- Issuing TLS certificates for every broker
- Rotating certificates before expiration
- Making the CA certificate available to clients (via a Secret)

When using LoadBalancer type, configure **TLS pass-through** (not termination). This ensures clients establish a TLS session directly with the broker, verifying its identity. Terminating TLS at the load balancer decrypts traffic in-flight within the cluster, weakening the security model.

## Production Best Practices

### High Availability (HA)

**Controller Quorum (KRaft):** Controller nodes form a Raft consensus quorum requiring 2n+1 nodes to tolerate n failures:

- **3 controllers:** Tolerates 1 failure (minimum HA)
- **5 controllers:** Tolerates 2 failures (high resilience)

**Broker Replication:** Configure topic `replicationFactor` (usually 3) to ensure partition data is replicated across multiple brokers. If one broker fails, replicas on other brokers serve traffic.

**Multi-Zone (Rack Awareness):** The most critical HA feature for surviving Availability Zone failures:

1. Ensure Kubernetes nodes are labeled by zone (e.g., `topology.kubernetes.io/zone: us-east-1a`)
2. Configure the Kafka CR:

```yaml
spec:
  kafka:
    rack:
      topologyKey: "topology.kubernetes.io/zone"
```

3. Strimzi uses an init container to read the zone label and inject it into each broker's `broker.rack` configuration
4. Kafka's partition assignment algorithm then spreads replicas across racks (zones), ensuring zone failure doesn't cause data loss

### Disaster Recovery (DR)

**Strategy:** Asynchronous replication to a geographically separate cluster using **MirrorMaker 2**.

**Pattern (Active/Passive):**

- **Active Cluster (Region A):** Serves all producer/consumer traffic
- **Passive Cluster (Region B):** MirrorMaker 2 continuously replicates topics from Active to Passive
- **DR Event:** Fail over all clients (by changing their bootstrap servers) to the Passive cluster, which becomes the new Active

Strimzi provides the `KafkaMirrorMaker2` CRD for declarative MirrorMaker management:

```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaMirrorMaker2
metadata:
  name: region-a-to-region-b
spec:
  replicas: 3
  connectCluster: "region-b"
  clusters:
    - alias: "region-a"
      bootstrapServers: region-a-kafka-bootstrap:9092
    - alias: "region-b"
      bootstrapServers: region-b-kafka-bootstrap:9092
  mirrors:
    - sourceCluster: "region-a"
      targetCluster: "region-b"
      sourceConnector: {}
```

### Security

**Encryption:** All internal communication is mTLS-secured automatically. External listeners should have `tls: true`. Strimzi's Cluster Operator manages the entire certificate lifecycle.

**Authentication:** Options include:

- **mTLS** (`authentication.type: tls`): Client certificates for authentication
- **SCRAM-SHA-512** (`authentication.type: scram-sha-512`): Username/password-based authentication

Managed via the `KafkaUser` CRD:

```yaml
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaUser
metadata:
  name: my-app-user
spec:
  authentication:
    type: scram-sha-512
  authorization:
    type: simple
    acls:
      - resource:
          type: topic
          name: my-topic
        operation: Read
```

**Authorization (ACLs):** Set `authorization.type: simple` on the Kafka CR. The User Operator then translates `KafkaUser` ACL specs into Kafka ACLs, enabling GitOps-friendly, declarative access control.

### Common Production Pitfalls

**Resource Starvation (Memory):** Kafka relies heavily on the operating system's page cache for performance. A common mistake is setting Kubernetes `memory.limits` equal to the JVM heap size (e.g., `-Xmx4g` and `limits.memory: 4Gi`). This leaves zero memory for the page cache, severely degrading performance.

**Rule of thumb:** `memory.limits` should be 2x the JVM heap size (e.g., `-Xmx4g` with `limits.memory: 8Gi`).

**Disk I/O Bottlenecks:** Using slow, network-backed storage will cripple Kafka. Always use high-performance SSDs (gp3, Premium SSD, SSD persistent disks).

**Over-partitioning:** Don't create thousands of partitions "just in case." Each partition is a file handle, log segment, and metadata object. Excessive partitions add overhead to controllers and increase recovery times.

**Ignoring Consumer Lag:** Consumer lag (`kafka_consumergroup_lag` metric) is the most critical alert. If lag consistently grows, consumers can't keep up, backlog is building, and retention policies may start dropping unprocessed data.

## Ecosystem: Schema Registry, Management UIs, and Monitoring

### Schema Registry: The 100% FOSS Choice

Strimzi deliberately does not bundle a schema registry, choosing to focus on core Kafka. For Planton, **Apicurio Registry** is the ideal companion:

- **License:** Apache 2.0 (100% FOSS)
- **Operator:** Dedicated Kubernetes operator available on OperatorHub
- **Strimzi Integration:** Can use a Strimzi-managed Kafka cluster as its storage backend (`persistence: kafkasql`)

This creates a fully integrated, 100% open-source stack.

**Apicurio Registry Configuration:**

```yaml
apiVersion: registry.apicur.io/v1
kind: ApicurioRegistry
metadata:
  name: my-registry
spec:
  configuration:
    persistence: "kafkasql"
    kafkasql:
      bootstrapServers: "my-kafka-cluster-kafka-bootstrap:9092"
      security:
        tls:
          truststoreSecretName: "my-kafka-cluster-cluster-ca-cert"
```

**Alternative (Not Recommended):** Confluent Schema Registry is widely used but licensed under the Confluent Community License (not OSI-approved open source, prohibits SaaS offerings). Advanced features require a paid enterprise license.

### Kafka Management UIs

**Open Source Options (Apache 2.0):**

- **Kafka UI (Kafbat):** Modern, feature-rich UI for managing topics, consumer groups, and browsing messages
- **AKHQ:** Mature, powerful UI often cited as the FOSS alternative to Confluent Control Center

**Commercial Options:**

- **Confluent Control Center:** The gold standard for Kafka management, but requires Confluent Enterprise License (not free)

### Monitoring Stack (Prometheus + Grafana)

Strimzi is built for Prometheus and Grafana:

**Prometheus Integration:** Enable metrics in the Kafka CR:

```yaml
spec:
  kafka:
    metricsConfig:
      type: jmxPrometheusExporter
      valueFrom:
        configMapKeyRef:
          name: kafka-metrics-config
          key: kafka-metrics-config.yml
```

Strimzi deploys a JMX Prometheus exporter sidecar that scrapes Kafka's internal JMX MBeans.

**Kafka Exporter:** For critical metrics not easily available via JMX (like consumer lag):

```yaml
spec:
  kafkaExporter:
    groupRegex: ".*"
    topicRegex: ".*"
```

**Grafana Dashboards:** Strimzi provides official, pre-built Grafana dashboards in its GitHub repository (`/examples/metrics/grafana-dashboards`). These dashboards visualize:

- Broker health (under-replicated partitions, active controller count)
- Consumer lag (the most important metric)
- Throughput (bytes in/out per second)
- Resource usage (CPU, memory, JVM heap)

Simply import these dashboards into your Grafana instance configured with your Prometheus data source.

### Key Metrics to Monitor

| Metric | What It Measures | Why It Matters | Alert Threshold |
|--------|------------------|----------------|-----------------|
| `UnderReplicatedPartitions` | Partitions lacking full replication | Indicates broker failure or replication lag | > 0 |
| `ActiveControllerCount` | Number of active controllers | Should always be 1 in stable cluster | != 1 |
| `kafka_consumergroup_lag` | Consumer group processing delay | Core business metric; backlog indicator | > threshold for your SLA |
| `BytesInPerSec` / `BytesOutPerSec` | Data throughput | Capacity planning, anomaly detection | Sudden drops or spikes |
| `request-latency-avg` | Request processing time | Client-perceived performance | > SLA latency |
| JVM Memory Usage | Heap utilization | Prevents out-of-memory errors | > 80% sustained |

## Conclusion: The Operator-Native Future

The journey from anti-patterns to production-ready Kafka on Kubernetes is a story of architectural maturity. Simple Deployments fail immediately due to network identity issues. StatefulSets provide a foundation but lack Kafka-aware lifecycle management. Helm charts simplify installation but don't solve operational complexity. Only Kubernetes Operators—with their embedded domain expertise—make Kafka on Kubernetes production-viable.

For open-source infrastructure frameworks like Planton, the path forward is clear:

1. **Adopt the Operator Pattern:** It's the only production-ready approach for Kafka's Day 2 operations
2. **Choose Strimzi:** 100% Apache 2.0, CNCF-backed, modern KRaft-native architecture, production-proven
3. **Embrace the FOSS Ecosystem:** Pair Strimzi with Apicurio Registry (schemas), Prometheus/Grafana (monitoring), and open-source management UIs

This stack provides not just ease of deployment, but ease of operation—the ability to scale, upgrade, recover, and monitor Kafka clusters with confidence. The ecosystem has consolidated around Strimzi as the definitive open-source solution, and its KRaft-native approach positions it perfectly for the ZooKeeper-free future Apache Kafka has already entered.

By understanding this evolution—from the why-it-fails foundation to the operator-driven present—you can deploy Kafka on Kubernetes not as an experiment, but as a core production service.

