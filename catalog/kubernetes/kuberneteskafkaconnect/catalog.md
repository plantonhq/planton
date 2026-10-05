# Kafka Connect

Deploys a Kafka Connect cluster on Kubernetes as a Strimzi `KafkaConnect` custom resource — the worker pool that runs connector plugins streaming data between Kafka and external systems (Debezium CDC, object stores, search indexes). Connector management is declarative-only on this platform: the module pins Strimzi's `strimzi.io/use-connector-resources` annotation, so pipes are declared as Kafka Connector resources and anything created through the Connect REST API is reverted by the operator. Plugin delivery is the star decision — the stock image carries ONLY the MirrorMaker 2 connector classes, so every real integration needs a prebuilt `image`, OCI image-volume `plugins`, or an operator `build`.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Strimzi `KafkaConnect` resource** — the cluster declaration (Kafka connection, plugin delivery, group identity, worker config, sizing), reconciled by the watching Strimzi operator into the worker Deployment and the Connect REST API Service (`<name>-connect-api`, port 8083).
- **Kubernetes Namespace** — created only when `createNamespace` is `true`; otherwise deploys into an existing namespace. A Strimzi operator must watch that namespace or the resource is accepted by the API server and silently never reconciled.
- **JMX Prometheus metrics ConfigMap** — created only when `metrics.enabled` is `true`; the canonical Strimzi Connect rule set, wired as the cluster's `metricsConfig` (port 9404 inside the worker pods).
- **Internal Connect storage topics** — derived on the target Kafka cluster by the Connect workers themselves (config / status / offset topics); default names derive from `metadata.name` and must stay unique among Connect-protocol workloads sharing the cluster.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster. The connection decides WHICH cluster this deploys into; the namespace is the placement unit inside it.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Strimzi Kafka Operator watching the target namespace** — the declared prerequisite. Kafka Connector declarations for this cluster must live in the same namespace.
- **A reachable Kafka cluster** — the `bootstrapServers` address, plus TLS trust material and credentials matching the target listener's authentication type. For a Strimzi-managed cluster, both resolve from the Apache Kafka resource's outputs.
- **ImageVolume feature (only for the `plugins` arm)** — mounting plugins from OCI artifacts requires the cluster and its container runtime to support Kubernetes image volumes; workers fail to schedule with an admission error on clusters without it.
- **A pushable registry (only for the `build` arm)** — the operator's Kaniko/Buildah build pushes to `build.output.image`, using the docker-registry Secret named in `pushSecret` when the registry needs credentials.

## Deploy

### Console

Open the deployment store, find **Kafka Connect**, and click **Deploy**. The creation wizard walks you through placement, the Kafka connection, plugin delivery (the star step), worker count, group identity, worker config, sizing, scheduling, and metrics. Start from the **Debezium prebuilt image preset** in the [Presets](#presets) tab for a real integration posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaConnect
metadata:
  name: cdc-connect
  org: acme-corp
  env: prod
spec:
  namespace:
    value: kafka
  createNamespace: false
  bootstrapServers:
    valueFrom:
      kind: KubernetesKafka
      name: event-bus
      fieldPath: status.outputs.internal_bootstrap_endpoint
  image: quay.io/debezium/connect:2.5
  replicas: 2
```

```shell
planton apply -f cdc-connect.yaml
```

This creates a two-worker Connect cluster running the Debezium Connect image against the `event-bus` cluster's internal bootstrap endpoint; pipes are then declared as Kafka Connector resources against this cluster. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire bootstrap and TLS trust from the sibling Kafka cluster:

```yaml
spec:
  namespace:
    value: kafka
  bootstrapServers:
    valueFrom:
      kind: KubernetesKafka
      name: event-bus
      fieldPath: status.outputs.internal_bootstrap_endpoint
  tls:
    trustedCertificates:
      - secretName:
          valueFrom:
            kind: KubernetesKafka
            name: event-bus
            fieldPath: status.outputs.cluster_ca_cert_secret_name
        certificate: ca.crt
  image: quay.io/debezium/connect:2.5
```

The InfraPipeline deploys the Kafka cluster first, then provisions the Connect cluster with its resolved endpoint and trust material.

## Key Configuration

These are the most important decisions when configuring a Kafka Connect cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Plugin delivery — the architecture decision.** The stock image carries only the MirrorMaker 2 connector classes (verified against the workers' own plugin listing — Kafka's FileStream examples are not on the distribution's classpath), so a connector declaring any real class fails with "class not found" until you pick an arm: `image` (a prebuilt Connect image, the fastest path when a vendor publishes one), `plugins` (OCI artifacts mounted as image volumes — no build, but requires the cluster's ImageVolume feature), or `build` (the operator builds and pushes a custom image from Maven/URL artifacts). `image` and `build` are mutually exclusive — when `build` is configured the operator runs the image IT builds and a declared `image` is silently overridden, so the spec rejects setting both.

**Connectors are resources, not REST calls.** The module always sets `strimzi.io/use-connector-resources: "true"` — declare every pipe as a Kafka Connector resource; connectors created or modified through the REST API are reverted by the operator. The REST endpoint output exists for read-only inspection (connector status, plugin listing) only.

**Group identity must be unique per shared Kafka cluster.** `groupId` and the three storage topics (`configStorageTopic`, `statusStorageTopic`, `offsetStorageTopic`) default from `metadata.name`. Two Connect clusters — including MirrorMaker 2 instances, which speak the same protocol — sharing a group ID or a storage topic corrupt each other's state. Distinct `metadata.name` values keep the defaults safe; override deliberately when names collide.

**Match the connection to the target listener.** `tls.trustedCertificates` decides trust (reference the Apache Kafka resource to trust its cluster CA; for external clusters, name any Secret holding the PEM), and `authentication.type` must match what the listener accepts — `scram-sha-512` credentials against a mutual-TLS listener fail at connect time. Reference a Kafka User resource to consume its operator-generated credential Secret instead of hand-managing one.

**Worker config has operator-owned zones.** Converters, replication factors for the storage topics, and connector-client policies belong in `config` — but entries with prefixes like `group.id`, `bootstrap.servers`, `ssl.`, `sasl.`, `rest.`, `plugin.path` are IGNORED with an operator log warning, never applied; those concerns are typed fields. One hard rejection: `connector.plugin.version` is not accepted in worker config on this Strimzi line — declare plugin versions on each Kafka Connector's `version` field.

**Size for the JVM, not just the pods.** Empty `resources` means no requests/limits (fine for dev); in production always set them, because Strimzi derives the default JVM heap from the memory limit. Set `jvm.xms` equal to `jvm.xmx` to avoid growth-triggered GC pauses, and scale `replicas` for task throughput — workers share connector tasks through the Connect group protocol.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesKafka** | `bootstrapServers` | `status.outputs.internal_bootstrap_endpoint` |
| **KubernetesKafka** | `tls.trustedCertificates[].secretName` | `status.outputs.cluster_ca_cert_secret_name` |
| **KubernetesKafkaUser** | `authentication.certificateAndKey.secretName` / `authentication.passwordSecret.secretName` | `status.outputs.secret_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the Connect workers run in | Placing Kafka Connector resources beside the cluster (they must live here) |
| `connect_name` | The Connect cluster's name (`metadata.name`) | The `connect_cluster` reference on Kafka Connector resources (rendered as the `strimzi.io/cluster` label) |
| `rest_api_service_name` | Name of the Connect REST API Service (`<name>-connect-api`) | Custom Service/NetworkPolicy composition |
| `rest_api_endpoint` | In-cluster REST endpoint (`http://<name>-connect-api.<namespace>.svc.cluster.local:8083`) | Read-only inspection — connector status, plugin listing (management stays declarative) |

## Common Patterns

**Exploring the Connect surface** — the smallest declarable cluster: stock image, one worker, plaintext bootstrap. Only the MirrorMaker 2 connectors can run, which makes this shape a protocol sandbox, not an integration platform. Start from the **Minimal stock Connect preset**.

**Real integrations from a vendor image** — the standard production entry: a prebuilt Connect image that already carries the connector classes (Debezium's published images are the canonical example), TLS trust and SCRAM credentials resolved from the sibling Kafka cluster and a Kafka User. Start from the **Debezium prebuilt image preset**.

**Operator-built plugin image** — when no vendor image carries your exact plugin set: declare Maven coordinates or artifact URLs (with `sha512sum` checksums so a tampered download fails the build instead of running in the workers) and let the operator build and push the image. Trades registry setup for exact plugin control. Start from the **Operator-built image preset**.

## Works With

- [**Apache Kafka**](/infra-catalog/kubernetes-kafka) — the cluster the workers read from and write to; its outputs resolve bootstrap and CA trust
- [**Kafka Connector**](/infra-catalog/kubernetes-kafka-connector) — the individual pipes, declared one per integration against this cluster's `connect_name`
- [**Kafka User**](/infra-catalog/kubernetes-kafka-user) — the authenticated principal whose credential Secret the workers present
- [**Strimzi Kafka Operator**](/infra-catalog/kubernetes-strimzi-kafka-operator) — the declared prerequisite: it must watch this cluster's namespace
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — provides the namespace shared with the Kafka cluster and its connectors
- [**Kafka MirrorMaker 2**](/infra-catalog/kubernetes-kafka-mirror-maker2) — purpose-built mirroring on the same Connect protocol; its group identity must not collide with this cluster's
- [**Kafka UI**](/infra-catalog/kubernetes-kafka-ui) — observe Connect cluster and connector status alongside topics
