# Apache Kafka

Deploys an Apache Kafka cluster on Kubernetes as Strimzi `Kafka` + `KafkaNodePool` custom resources in **KRaft mode** — Kafka's built-in Raft metadata quorum; ZooKeeper does not exist in this architecture. The cluster is declared as node pools (independently scalable groups of nodes carrying the `controller` and/or `broker` roles), listeners (how clients reach the brokers, from in-cluster plaintext to cloud LoadBalancers with TLS and SASL), and broker configuration. The Strimzi cluster operator (installed separately as Strimzi Kafka Operator) reconciles the declaration into brokers, controllers, certificates, and the per-cluster entity operators. Topics and users are first-class Infra Components — Kafka Topic and Kafka User — never embedded in this spec.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** — created only when `createNamespace` is `true`; otherwise deploys into an existing namespace. One namespace per cluster is the recommended posture: node pool resources carry the pool's own name, so two clusters sharing a namespace collide on same-named pools.
- **Strimzi `Kafka` resource** — the cluster declaration (listeners, broker configuration, authorization, entity operators, Cruise Control, metrics, CAs, rack awareness, JVM, maintenance windows), reconciled by the watching Strimzi operator.
- **Strimzi `KafkaNodePool` resources** — one per declared node pool, bound to the cluster via the `strimzi.io/cluster` label. Each pool carries its KRaft roles, replica count, storage shape (persistent-claim, ephemeral, or JBOD volumes), container resources, and scheduling.
- **Entity operators** — the per-cluster topic and user reconcilers (both enabled by default) that make KafkaTopic/KafkaUser declarations in this namespace real.
- **Per-listener Services** — ClusterIP services for internal listeners; NodePort or per-broker cloud LoadBalancer services (with your annotations) for external ones.
- **Strimzi-managed CAs and certificates** — the cluster CA (signs broker certificates; exported as the `<cluster>-cluster-ca-cert` Secret) and the clients CA (signs KafkaUser client certificates), renewed automatically.
- **Optional companions** — Cruise Control (the partition-rebalancing autopilot), the Kafka Exporter (consumer-lag metrics), and the canonical JMX Prometheus metrics ConfigMap, each only when enabled.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster. The connection decides WHICH cluster this deploys into — the spec deliberately carries no cluster reference; the namespace is the placement unit inside it.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Strimzi Kafka Operator watching the target namespace** — the declared prerequisite. Install it as the Strimzi Kafka Operator Infra Component; its chart default watches its OWN namespace only, so the simplest posture is the operator and this cluster in the same namespace. Without a watching operator the cluster is accepted by the API server and silently never reconciled.
- **A storage class** for persistent volumes — Kafka is a storage-bound system; persistent-claim storage is the default and the only production-sane choice.
- **Cloud LB / DNS controllers where the listeners need them** — a `loadbalancer` listener's annotations are answered by a specific controller (e.g. the AWS Load Balancer Controller for the `external` family); external-dns for automatic DNS records.

## Deploy

### Console

Open the deployment store, find **Apache Kafka**, and click **Deploy**. The creation wizard walks the operator's decision sequence — placement, version, the node pools (the machines), rack awareness, the listeners (client access), broker durability configuration, authorization, entity operators, observability, Cruise Control, JVM heap, and certificate lifecycle. Start from the **Dev single node preset** for development or the **Production three broker preset** for production in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafka
metadata:
  name: event-bus
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "kafka"
  createNamespace: true
  nodePools:
    - name: controller
      roles:
        - controller
      replicas: 3
      storage:
        size: 20Gi
    - name: broker
      roles:
        - broker
      replicas: 3
      storage:
        size: 100Gi
  listeners:
    - name: tls
      port: 9093
      tls: true
      authentication:
        type: scram-sha-512
  config:
    default.replication.factor: "3"
    min.insync.replicas: "2"
    offsets.topic.replication.factor: "3"
    transaction.state.log.replication.factor: "3"
    transaction.state.log.min.isr: "2"
```

```shell
planton apply -f kafka.yaml
```

This creates a KRaft cluster with a 3-node controller quorum, three brokers on 100Gi persistent volumes, one TLS listener with SCRAM authentication, and the recommended durability settings — one broker can be lost without losing acknowledged writes from `acks=all` producers. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the cluster to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: kafka-home
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace (and the Strimzi operator) first, then provisions the cluster into it.

## Key Configuration

These are the most important decisions when configuring an Apache Kafka deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Node pools — the machines.** At least one pool must carry the `controller` role (the KRaft metadata quorum — keep the count ODD: 3 for production, 1 for dev) and one the `broker` role (stores and serves data; 3 is the floor for surviving a broker loss with the recommended replication). One dual-role pool is the dev shape; production separates the pools so brokers scale without disturbing the quorum. Storage is per pool: `persistent-claim` with a `size` (the default), `ephemeral` (dev only — data dies with the pod), or `jbod` with multiple volumes per node (volume IDs are part of the on-disk layout — grow by adding IDs, never renumber).

**Listeners — how clients reach the brokers.** Each listener declares a name (1–11 lowercase alphanumerics), a unique port (9092+), an exposure type (`internal`, `cluster-ip`, `nodeport`, `loadbalancer`, `ingress` — deprecated upstream, prefer loadbalancer/nodeport — `route`, `tlsroute`), TLS, and client authentication (`tls`, `scram-sha-512`, or `custom`; credentials are minted by Kafka User resources, never typed here). Cloud LoadBalancers are shaped through Service annotations on the bootstrap/per-broker configuration — note which controller answers each annotation.

**Durability lives in `config`.** `default.replication.factor: "3"` + `min.insync.replicas: "2"` (plus the internal-topic RF entries) is the posture that survives a broker loss; a single-node dev cluster runs them at `"1"`. Values are Kafka configuration strings — write numbers and booleans as strings. The operator OWNS listener/security/quorum configuration: entries with prefixes like `listeners`, `ssl.`, `sasl.`, `node.id`, `controller.` are ignored with a log warning, never applied.

**Authorization is a deliberate switch.** Omitted = no authorizer: every client that can connect can do everything. Enabling `simple` enforces the ACLs declared on Kafka User resources from that moment — declare your operational tooling's principals as `superUsers` BEFORE enabling, or a single deploy can lock every client out.

**Upgrades and the metadata one-way door.** `kafkaVersion` empty runs the operator's default. When upgrading, hold `metadataVersion` at the old version's format until every node runs the new binaries — once bumped, the KRaft metadata format cannot be downgraded.

**Everything else** — rack awareness (`rack.topologyKey`, typically `topology.kubernetes.io/zone`, spreads replicas across zones), the entity-operator dials (disabling one makes KafkaTopic/KafkaUser declarations silently inert), Cruise Control with its add/remove-brokers auto-rebalance modes, JMX metrics + the consumer-lag exporter, JVM heap (set -Xms equal to -Xmx in production), and the CA validity/renewal windows with maintenance time windows for renewal-triggered rolling restarts.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesStorageClass** | `nodePools[].storage.storageClass` (and per JBOD volume) | `metadata.name` |
| **KubernetesCertificate** | `listeners[].configuration.brokerCertChainAndKey.secretName` | `status.outputs.secret_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that applications and downstream resources can consume:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the cluster runs in | Placing Kafka Topic / Kafka User resources beside the cluster (their entity operators watch only here) |
| `cluster_name` | Cluster name (`metadata.name`) — the `strimzi.io/cluster` label value | The `kafka_cluster` reference on Kafka Topic / Kafka User resources |
| `bootstrap_service_name` | Name of the internal bootstrap Service (`<cluster>-kafka-bootstrap`) | Custom Service/NetworkPolicy composition |
| `internal_bootstrap_endpoint` | In-cluster bootstrap address for the FIRST internal listener (`<cluster>-kafka-bootstrap.<namespace>.svc.cluster.local:<port>`); empty when no internal listener is declared | The value workloads put in `bootstrap.servers` |
| `cluster_ca_cert_secret_name` | Secret holding the cluster CA certificate (`<cluster>-cluster-ca-cert`, key `ca.crt`) | What TLS clients add to their truststore |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev single node** — the smallest declarable Kafka that actually serves: one dual-role pool (KRaft controller + broker in one pod), one plaintext internal listener, RF-1 settings. Real wire protocol, zero durability. Start from the **Dev single node preset**.

**Production three-broker** — the standard production posture: a 3-node controller pool, a 3-node broker pool on JBOD storage, RF-3 / min-ISR-2, SCRAM-over-TLS and mutual-TLS listeners, simple authorization with a real admin super user, zone-spread rack awareness, full observability, explicit CA policy with a Sunday-night maintenance window, and a fixed 4g heap. Start from the **Production three broker preset**.

**External access on AWS** — the production shape plus a `loadbalancer` listener: the AWS Load Balancer Controller provisions one NLB per broker plus bootstrap (Kafka is raw TCP — an L7 balancer cannot carry it), with external-dns annotations for DNS. Start from the **AWS external loadbalancer preset**.

**SCRAM application cluster** — a quorum-safe cluster with exactly one credential story: a single SCRAM-SHA-512-over-TLS listener plus simple ACL authorization. Start from the **SCRAM app cluster preset**.

## Works With

- [**Strimzi Kafka Operator**](/infra-catalog/kubernetes-strimzi-kafka-operator) — the declared prerequisite: it must watch this cluster's namespace
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — provides the namespace (one per cluster)
- [**Kafka Topic**](/infra-catalog/kubernetes-kafka-topic) — topics as first-class resources, reconciled by this cluster's topic operator
- [**Kafka User**](/infra-catalog/kubernetes-kafka-user) — authenticated principals + ACLs, reconciled by this cluster's user operator
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) — custom listener server certificates (the cert-manager seam)
- [**ExternalDNS**](/infra-catalog/kubernetes-external-dns) — DNS records for external listener addresses
