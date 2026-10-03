# Kubernetes Kafka

## When NOT to Use This

**The operator must already be on the cluster.** This component
declares a Kafka cluster; KubernetesStrimziKafkaOperator installs the
ENGINE that reconciles it. The default operator posture watches its
OWN namespace — install the operator in the cluster's namespace, or
widen its watch. Deploy the operator first, clusters after.

Also not the right kind when:

- **You want the operator itself** — installing and configuring the
  Strimzi cluster operator is KubernetesStrimziKafkaOperator; this
  component is one Kafka cluster it manages.
- **You want topics or users** — those are first-class resources:
  KubernetesKafkaTopic and KubernetesKafkaUser, reconciled by THIS
  cluster's entity operators (enabled by default). They must be
  declared in this cluster's namespace — the entity operators watch
  only here.
- **You expect ZooKeeper** — this is KRaft mode (Kafka's built-in Raft
  metadata quorum, run by nodes carrying the `controller` role).
  ZooKeeper does not exist in this architecture; Strimzi removed
  ZooKeeper support in 0.46.
- **You want a managed cloud Kafka** — use the host cloud provider's
  managed streaming kinds; this kind is for running Kafka ON the
  Kubernetes cluster itself.
- **You want HTTP-style exposure baked in** — Kafka is not HTTP.
  Listeners ARE the exposure surface here (internal, nodeport,
  loadbalancer, ingress, route, tlsroute, cluster-ip), and external
  listeners carry the cloud annotations. Anything beyond listeners
  (Gateways, custom proxies) composes from the exported service
  handles — never embedded here.
- **You need a Strimzi surface the spec deliberately leaves out** —
  tiered storage, the quotas plugin, per-listener networkPolicyPeers,
  Gateway API parentRefs for tlsroute listeners, custom pod templates
  beyond node_selector/tolerations, remote JMX (see the research doc).
  Those are reachable today by declaring the raw custom resource
  through KubernetesManifest.

## Overview

**KubernetesKafka** declares one KRaft-mode Apache Kafka cluster
reconciled by the Strimzi cluster operator. The spec renders a
`kafka.strimzi.io/v1` `Kafka` custom resource plus one
`KafkaNodePool` per `node_pools` entry (bound to the cluster by the
`strimzi.io/cluster` label) — so one resource carries the whole
cluster story: node pools (roles, storage, sizing), listeners with
authentication, broker configuration, authorization, the entity
operators, Cruise Control, metrics, CAs, rack awareness, and
maintenance windows.

**Cluster shape**: at least one pool must carry the `controller` role
and one the `broker` role (validated at the spec). One dual-role pool
is the dev shape; the production norm is a 3-node controller pool plus
one or more broker pools, so brokers scale without disturbing the
metadata quorum. Controller counts should be odd (3 tolerates one
loss).

**The naming contract**: every object the operator creates derives
from `metadata.name` — the internal bootstrap Service
(`<name>-kafka-bootstrap`), the cluster CA Secret
(`<name>-cluster-ca-cert`, key `ca.crt`), and the `strimzi.io/cluster`
label value that binds KafkaNodePool, KafkaTopic and KafkaUser
resources to this cluster.

**Key design points:**

- **Node pools are the only topology primitive** — independently
  scalable groups sharing roles, storage shape (persistent-claim,
  ephemeral, or JBOD — multiple disks per node for throughput),
  sizing and scheduling.
- **Listeners are the client surface** — seven exposure types
  (`internal`, `cluster-ip`, `nodeport`, `loadbalancer`, `ingress`,
  `route`, `tlsroute`), each with optional TLS and authentication
  (`tls` mutual-TLS, `scram-sha-512`, or `custom` — OAuth lost its
  first-class Strimzi type in 1.x and routes through custom).
  External listeners are where cloud annotations ride
  (`configuration.bootstrap.annotations` / `brokers[].annotations`).
  Ingress listeners REQUIRE `bootstrap.host`, a host per broker, and
  an ingress controller with SSL passthrough enabled.
- **`config` values are strings** — Kafka configuration entries are
  written as strings ("3", "false"); the operator serializes them
  into Java properties. Operator-owned prefixes (listeners,
  advertised., sasl., ssl., node.id, controller., and the rest) are
  IGNORED with an operator log warning — those concerns have typed
  fields.
- **Durability is configuration, not defaults** — a 3-broker cluster
  with `default.replication.factor: "3"` and
  `min.insync.replicas: "2"` survives one broker loss without losing
  acknowledged writes (producers using acks=all). The spec recommends
  exactly those entries.
- **Authorization is opt-in and cluster-wide** — `simple` enforces the
  ACLs declared on KubernetesKafkaUser resources; from that moment
  clients without matching ACLs are denied, so list real admin
  principals in `super_users` FIRST. Keycloak/OPA lost their
  first-class Strimzi types in 1.x and route through `custom`.
- **Certificates are operator-managed** — the cluster CA and clients
  CA are generated and renewed automatically (validity/renewal windows
  configurable); a per-listener `broker_cert_chain_and_key` is the
  cert-manager seam for custom server certificates.
- **Topics and users compose, never embed** — KubernetesKafkaTopic /
  KubernetesKafkaUser in this cluster's namespace, served by the
  entity operators this spec keeps enabled by default.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace for the cluster — literal or a
  KubernetesNamespace reference; the operator must watch it, and
  KafkaTopic/KafkaUser declarations must live here too
- **`spec.node_pools`**: at least one pool with the `controller` role
  and one with the `broker` role (one dual-role pool satisfies both);
  each pool requires `replicas` and `storage`
- **`spec.listeners`**: at least one — unique names (1–11 lowercase
  alphanumerics) and unique ports (9092+)

### Common

- **`spec.kafka_version` / `spec.metadata_version`**: empty = the
  operator's default (the newest the pinned Strimzi release
  supports). During upgrades, hold `metadata_version` at the OLD
  format until every node runs the new binaries — the KRaft metadata
  format is a one-way door
- **`spec.config`**: broker configuration strings — the recommended
  production set is `default.replication.factor: "3"`,
  `min.insync.replicas: "2"`, plus matching internal-topic entries
- **`spec.authorization`**: `simple` (Kafka's built-in ACL authorizer)
  or `custom` (bring-your-own authorizer class); `super_users` in
  Kafka principal form (`User:CN=...` for TLS users, `User:name` for
  SCRAM users)
- **`spec.entity_operator`**: topic/user operators — both default
  true; disabling one makes the corresponding declarations silently
  inert
- **`spec.cruise_control`**: partition-rebalancing autopilot;
  `auto_rebalance_modes` (add-brokers/remove-brokers) rebalances on
  pool scale events
- **`spec.kafka_exporter`**: consumer-lag and topic/partition metrics
  (a separate exporter pod)
- **`spec.metrics`**: the canonical Strimzi JMX Prometheus rules,
  rendered as a module-owned ConfigMap and wired as `metricsConfig`
  (port 9404 in the pods)
- **`spec.cluster_ca` / `spec.clients_ca`**: validity/renewal windows
  for the operator-managed CAs
- **`spec.rack`**: `topology_key` (e.g. `topology.kubernetes.io/zone`)
  spreads replicas across zones and enables closest-replica fetching
- **`spec.jvm`**: broker heap (`xms`/`xmx`) — set both to the SAME
  value in production; empty = Strimzi's dynamic default from the
  container memory limit
- **`spec.maintenance_time_windows`**: cron expressions fencing WHEN
  certificate-renewal rolling updates may run

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the cluster runs in |
| `cluster_name` | Name of the Kafka resource (equals `metadata.name`) — the `strimzi.io/cluster` binding value for KafkaNodePool/KafkaTopic/KafkaUser resources |
| `bootstrap_service_name` | Internal bootstrap Service (`<name>-kafka-bootstrap`) |
| `internal_bootstrap_endpoint` | In-cluster bootstrap address for the FIRST internal listener (`<service>.<namespace>.svc.cluster.local:<port>`) — what workloads put in `bootstrap.servers`; empty when no internal listener is declared |
| `cluster_ca_cert_secret_name` | Secret holding the cluster CA certificate (`<name>-cluster-ca-cert`, key `ca.crt`) — what TLS clients add to their truststore |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`); pool and JBOD
  **`storage_class`** fields reference a KubernetesStorageClass;
  a listener's **`broker_cert_chain_and_key.secret_name`** references
  a KubernetesCertificate's output secret — the cert-manager seam.
- **Applications consume the outputs**: `internal_bootstrap_endpoint`
  as `bootstrap.servers`, `cluster_ca_cert_secret_name` for the TLS
  truststore, and a KubernetesKafkaUser's credential Secret for
  authentication — credentials ride operator-managed Secrets, never
  the manifest.
- **Topics and users compose against `cluster_name`**: declare
  KubernetesKafkaTopic / KubernetesKafkaUser in this cluster's
  namespace; their kinds render the `strimzi.io/cluster` label from
  it.
- **External DNS composes with external listeners**: put
  `external-dns.alpha.kubernetes.io/hostname` annotations on
  `bootstrap`/`brokers` and run KubernetesExternalDns — the listener
  provisions the LoadBalancers, external-dns points names at them.
- **The operator is a cluster prerequisite**, not a reference: deploy
  KubernetesStrimziKafkaOperator first — in the cluster's namespace,
  or with its watch widened to cover it.

## Examples

### Development (single dual-role node)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafka
metadata:
  name: dev-kafka
spec:
  namespace:
    value: dev-kafka
  create_namespace: true
  node_pools:
    - name: dual
      roles: [controller, broker]
      replicas: 1
      storage:
        size: 10Gi
  listeners:
    - name: plain
      port: 9092
      tls: false
  config:
    default.replication.factor: "1"
    min.insync.replicas: "1"
    offsets.topic.replication.factor: "1"
    transaction.state.log.replication.factor: "1"
    transaction.state.log.min.isr: "1"
```

### Production (3 controllers + 3 brokers, SCRAM + mTLS, RF 3)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafka
metadata:
  name: events
spec:
  namespace:
    value: kafka
  node_pools:
    - name: controller
      roles: [controller]
      replicas: 3
      storage:
        size: 20Gi
      resources:
        requests: { cpu: 500m, memory: 2Gi }
        limits: { cpu: "1", memory: 2Gi }
    - name: broker
      roles: [broker]
      replicas: 3
      storage:
        size: 500Gi
      resources:
        requests: { cpu: "2", memory: 8Gi }
        limits: { cpu: "4", memory: 8Gi }
  listeners:
    - name: scram
      port: 9092
      tls: true
      authentication:
        type: scram-sha-512
    - name: mtls
      port: 9093
      tls: true
      authentication:
        type: tls
  config:
    default.replication.factor: "3"
    min.insync.replicas: "2"
    offsets.topic.replication.factor: "3"
    transaction.state.log.replication.factor: "3"
    transaction.state.log.min.isr: "2"
  authorization:
    type: simple
    super_users:
      - User:CN=platform-admin
  rack:
    topology_key: topology.kubernetes.io/zone
  jvm:
    xms: 4g
    xmx: 4g
```

### External access (AWS NLB loadbalancer listener)

```yaml
listeners:
  - name: external
    port: 9095
    type: loadbalancer
    tls: true
    authentication:
      type: tls
    configuration:
      bootstrap:
        annotations:
          service.beta.kubernetes.io/aws-load-balancer-type: external
          external-dns.alpha.kubernetes.io/hostname: kafka.example.com
      brokers:
        - broker: 0
          advertised_host: broker-0.kafka.example.com
          annotations:
            external-dns.alpha.kubernetes.io/hostname: broker-0.kafka.example.com
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
