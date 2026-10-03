# Kubernetes Kafka Connect

## When NOT to Use This

**The operator must already be on the cluster.** This component
declares a Kafka Connect cluster; KubernetesStrimziKafkaOperator
installs the ENGINE that reconciles it. The operator must watch this
cluster's namespace. Deploy the operator first, Connect clusters
after.

Also not the right kind when:

- **You want a connector (a data pipe)** — connectors are first-class
  resources: KubernetesKafkaConnector, declared in THIS cluster's
  namespace and bound to it by name. This kind is the worker
  fleet the connectors run on.
- **You plan to manage connectors through the Connect REST API** —
  the module always stamps the
  `strimzi.io/use-connector-resources: "true"` annotation, so the
  operator owns connectors on this cluster and reverts REST-API-made
  changes. Declare pipes as KubernetesKafkaConnector resources; the
  exported REST endpoint is for read-only inspection.
- **You want Kafka itself** — the Kafka cluster this Connect cluster
  reads from and writes to is KubernetesKafka (or an external
  cluster's literal bootstrap address).
- **You want MirrorMaker 2** — cluster-to-cluster replication runs
  the same Connect machinery but has its own kind:
  KubernetesKafkaMirrorMaker2.
- **You need a Strimzi Connect surface the spec deliberately leaves
  out** — tracing, remote JMX, custom logging categories, custom pod
  templates beyond node_selector/tolerations (see the research doc).
  Those remain reachable by declaring the raw custom resource through
  KubernetesManifest.

## Overview

**KubernetesKafkaConnect** declares one Kafka Connect cluster on the
Strimzi `kafka.strimzi.io/v1` `KafkaConnect` custom resource — the
pluggable integration engine that streams data between Kafka and
external systems (databases via Debezium CDC, object stores, search
indexes, SaaS APIs). The Strimzi cluster operator reconciles it into
Connect worker pods and the REST API Service; individual pipes are
declared as KubernetesKafkaConnector resources against this cluster.

**Connector plugins reach the workers four ways**, from simplest to
most self-contained:

1. **The stock image** — carries ONLY the MirrorMaker 2 connector
   classes (MirrorSource/MirrorCheckpoint/MirrorHeartbeat; Kafka's
   FileStream examples are NOT on the distribution's classpath).
   Zero plugin machinery — enough for a first self-mirror pipe, and
   every real integration graduates to one of the arms below.
2. **`image`** — run a prebuilt Connect image that already carries
   your plugins (the fastest path when a vendor publishes one, e.g. a
   Debezium Connect image).
3. **`plugins`** — mount plugins from OCI artifacts as Kubernetes
   image volumes: no image build, no registry push. REQUIRES the
   cluster's ImageVolume feature AND container-runtime support — a
   cluster-level capability; workers fail to schedule with an
   image-volume admission error on clusters without it.
4. **`build`** — have the OPERATOR build a custom image from declared
   artifacts (jar/tgz/zip URLs or Maven coordinates, via Kaniko or
   Buildah on Kubernetes) and push it to your registry; the workers
   then run the built image.

`image` and `build` are mutually exclusive (spec-validated): when
`build` is configured the operator deploys the image IT builds and a
declared `image` would be silently overridden.

**The group identity contract**: `group_id` and the three internal
storage topics default from `metadata.name` (`<name>`,
`<name>-connect-configs`, `<name>-connect-status`,
`<name>-connect-offsets`) and MUST be unique among Connect-protocol
workloads (MirrorMaker 2 instances included) sharing a Kafka cluster —
two clusters sharing a group.id or a storage topic corrupt each
other's state.

**Key design points:**

- **The Kafka connection is typed and composable** —
  `bootstrap_servers` accepts a literal address (external clusters,
  Confluent, MSK) or a KubernetesKafka reference; `tls` trusts the
  referenced cluster's CA Secret; `authentication` supports mutual
  TLS, SCRAM, PLAIN and custom SASL, wired to KubernetesKafkaUser
  credential Secrets by reference — credentials ride Secrets, never
  the manifest.
- **`config` values are strings** — worker configuration entries
  (converters, storage replication factors) are written as strings
  ("3", "false"); the operator serializes them into Java properties.
  Connection, identity and listener prefixes (`bootstrap.servers`,
  `group.id`, `ssl.`, `sasl.`, `rest.`, `plugin.path`, ...) are
  operator-owned and IGNORED with a log warning — those concerns have
  typed fields.
- **Storage replication factors must fit the cluster** — Connect's
  own default of 3 cannot be satisfied on a single-broker dev
  cluster and the workers wedge creating their internal topics; set
  the three `*.storage.replication.factor` entries to `"-1"` (broker
  default) there, and `"3"` in production.
- **Metrics are module-owned** — `metrics.enabled` renders the
  canonical Strimzi JMX Prometheus rules ConfigMap
  (`<name>-connect-metrics`) and wires it as the cluster's
  `metricsConfig` (port 9404 in the pods).
- **Sizing is explicit** — `replicas` scales task capacity; `jvm`
  (set xms = xmx in production) and `resources` size each worker.
  `rack.topology_key` enables closest-replica fetching.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace for the Connect cluster — literal
  or a KubernetesNamespace reference; the operator must watch it, and
  KubernetesKafkaConnector declarations for this cluster must live
  here too
- **`spec.bootstrap_servers`**: the Kafka cluster to read from and
  write to — literal `host:port` or a KubernetesKafka reference
  (resolves to its in-cluster bootstrap endpoint)

### Common

- **`spec.replicas`**: worker count (default 1) — workers share
  connector tasks through the Connect group protocol
- **`spec.version`**: Kafka version the workers run; empty = the
  operator's default — keep aligned with the target cluster during
  upgrades
- **`spec.tls` / `spec.authentication`**: TLS trust and client
  authentication matching the target listener (reference the
  KubernetesKafka cluster CA and a KubernetesKafkaUser credential
  Secret for Strimzi-managed targets)
- **`spec.config`**: worker configuration strings — converters
  (`key.converter`, `value.converter`) and the three
  `*.storage.replication.factor` entries
- **`spec.image` XOR `spec.build`**: prebuilt plugin image OR
  operator-driven image build (registry destination, push Secret
  name, plugin artifacts)
- **`spec.plugins`**: OCI image-volume plugins (cluster ImageVolume
  feature required)
- **`spec.group_id` / storage topics**: override the
  metadata.name-derived group identity when naming matters
- **`spec.resources` / `spec.jvm` / `spec.rack` / `spec.metrics` /
  `spec.node_selector` / `spec.tolerations`**

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the Connect cluster runs in |
| `connect_name` | The Connect cluster's name (`metadata.name`) — what KubernetesKafkaConnector resources bind to via `connect_cluster` |
| `rest_api_service_name` | The Connect REST API Service (`<name>-connect-api`) |
| `rest_api_endpoint` | In-cluster REST endpoint (`http://<name>-connect-api.<namespace>.svc.cluster.local:8083`) — read-only inspection; also what a KubernetesKafkaUi console wires as a Connect cluster address |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace); **`bootstrap_servers`** references a
  KubernetesKafka (`status.outputs.internal_bootstrap_endpoint`);
  **`tls.trusted_certificates[].secret_name`** references a
  KubernetesKafka (`status.outputs.cluster_ca_cert_secret_name`);
  **`authentication`** Secret names reference a KubernetesKafkaUser
  (`status.outputs.secret_name` — `user.crt`/`user.key` for TLS,
  `password` for SCRAM).
- **Connectors compose against `connect_name`**: declare
  KubernetesKafkaConnector resources in THIS namespace; their kind
  renders the `strimzi.io/cluster` label from it.
- **A KubernetesKafkaUi console** wires `rest_api_endpoint` as a
  `kafka_connect` address to monitor the pipes.
- **The operator is a cluster prerequisite**, not a reference: deploy
  KubernetesStrimziKafkaOperator first, watching this namespace.

## Examples

### Minimal (stock image, dev cluster)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaConnect
metadata:
  name: dev-connect
spec:
  namespace:
    value: dev-kafka
  bootstrap_servers:
    value: dev-kafka-kafka-bootstrap.dev-kafka.svc.cluster.local:9092
  replicas: 1
  config:
    config.storage.replication.factor: "-1"
    offset.storage.replication.factor: "-1"
    status.storage.replication.factor: "-1"
```

### Prebuilt Debezium image against a TLS+SCRAM cluster

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaConnect
metadata:
  name: cdc-connect
spec:
  namespace:
    value: kafka
  bootstrap_servers:
    value: events-kafka-bootstrap.kafka.svc.cluster.local:9093
  replicas: 3
  image: quay.io/example/debezium-connect:3.1 # a prebuilt image carrying your plugins
  tls:
    trusted_certificates:
      - secret_name:
          value: events-cluster-ca-cert
        certificate: ca.crt
  authentication:
    type: scram-sha-512
    username: connect
    password_secret:
      secret_name:
        value: connect
      password: password
  config:
    config.storage.replication.factor: "3"
    offset.storage.replication.factor: "3"
    status.storage.replication.factor: "3"
```

### Operator-built image (Maven Debezium artifact)

```yaml
spec:
  build:
    output:
      type: docker
      image: registry.example.com/team/cdc-connect:latest
      push_secret: registry-push-creds
    plugins:
      - name: debezium-connector-postgres
        artifacts:
          - type: maven
            group: io.debezium
            artifact: debezium-connector-postgres
            version: 3.1.0.Final
  # no `image` field — the operator runs the image it builds
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
