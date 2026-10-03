# RabbitMQ

Deploys a RabbitMQ cluster — the queue-messaging broker behind task queues, work distribution, RPC, and per-message routing (AMQP 0-9-1/1.0, with MQTT and STOMP via plugins) — declared as a `RabbitmqCluster` custom resource and reconciled by the RabbitMQ Cluster Operator. The operator renders the cluster as one StatefulSet plus two Services and generates the administrator credentials itself — they never pass through the spec.

**Queue, not log**: RabbitMQ delivers each message and forgets it on acknowledgement. For append-only event streaming and replayable history at scale, use **Apache Kafka** instead — that boundary is the first decision, not a tuning knob.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **RabbitmqCluster custom resource** (`rabbitmq.com/v1beta1`, named `metadata.name`) — topology, storage, TLS, configuration, and placement, reconciled by the operator
- **Kubernetes Namespace** — created only when `createNamespace` is true; otherwise the namespace must already exist

The operator reconciles these into the running cluster, following its naming contract: the StatefulSet `<name>-server` with one data PVC per pod (`persistence-<name>-server-<i>`), the client Service `<name>` (ClusterIP by default), the headless inter-node Service `<name>-nodes`, and the generated admin credentials Secret `<name>-default-user` (keys: username, password, host, port, connection_string, ...).

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.
- **RabbitMQ Cluster Operator** — a **RabbitMQ Cluster Operator** resource must be running and **watching the target namespace**. It watches ALL namespaces by default; if it was installed with a fenced `watch_namespaces` list, the target namespace must be covered — clusters elsewhere are silently ignored. Deploy the operator first.

### Kubernetes Cluster

- **A StorageClass** for the data volumes — most managed clusters provide a default; reference a **Kubernetes StorageClass** for explicit (SSD) placement. Not needed when the cluster is `ephemeral`.
- **A TLS certificate Secret** — only if you declare `tls`: an existing kubernetes.io/tls Secret covering the client Service DNS names (`<name>.<namespace>.svc` and its cluster-domain form). A **Cert Manager Certificate** reference plugs in directly.

## Deploy

### Console

Open the deployment store, find **RabbitMQ**, and click **Deploy**. The creation wizard walks you through namespace placement, the node count and quorum posture, storage, the server image, per-node resources, the client Service face, plugins and configuration, TLS, credential delivery, behavior switches, and scheduling. Start from the **Dev single node preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRabbitMq
metadata:
  name: orders-mq
  org: acme-corp
  env: dev
spec:
  namespace:
    value: messaging
  createNamespace: true
```

```shell
planton apply -f rabbitmq.yaml
```

Every field has a working default: a single node on a 10Gi volume (the cluster's default StorageClass), the operator's default `-management` image, operator-default sizing. Workloads connect at `orders-mq.messaging.svc.cluster.local:5672` with the credentials from the `orders-mq-default-user` Secret (it even carries a ready-made `connection_string`); the management UI is on port 15672. An Infra Job tracks the provisioning in real time.

### InfraChart

Compose the cluster behind its namespace with a reference, and the InfraPipeline orders the deploys:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: messaging-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions the cluster into it.

## Key Configuration

These are the most important decisions when configuring a RabbitMQ cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Replicas are odd, and a one-way door** — production clusters use an ODD count (3, 5, 7) so quorum queues and the Raft-based metadata store survive node loss; a 2-node cluster loses availability when EITHER node fails. And the operator does NOT support scaling down — removed brokers strand their queue replicas, so sizing down means migrating to a new cluster, not editing a field. Grow in odd steps, deliberately.

**Memory requests must equal limits** — RabbitMQ derives its memory high watermark from the container memory LIMIT, so a gap between requests and limits triggers flow control at the wrong threshold. The operator's own defaults follow the rule (requests 1 CPU / 2Gi, limits 2 CPU / 2Gi — memory equal on both sides).

**`spreadAcrossNodes` is the production posture** — required pod anti-affinity puts every broker on a different Kubernetes node, so one node loss takes one broker instead of the quorum. Off by default so single-node dev clusters schedule; know that a cluster with more replicas than schedulable nodes sits Pending when it is on.

**Ephemeral is a dev-only trade** — `ephemeral: true` runs on emptyDir: every queue, message, and user vanishes with each pod restart. It is mutually exclusive with `diskSize` / `storageClass` (the spec rejects the combination). Everything else sizes `diskSize` per node — queues, quorum-queue Raft logs, and the metadata store live there, and PVCs cannot shrink.

**Credentials are operator-generated** — the spec never carries them; consumers read the `<name>-default-user` Secret from the outputs. The `secretBackend` block redirects delivery (exactly one arm): `vault` has the credential-updater sidecar read and rotate credentials at a Vault path — and the Secret output goes EMPTY — while `externalSecretName` points at a pre-existing Secret, the seam for **Kubernetes External Secret** flows that materialize credentials from a cloud secret manager. Avoid `default_user` / `default_pass` lines in `additionalConfig`: they override the generated credentials and put the password in plaintext on the CR.

**TLS is a mount, not a certificate authority** — `tls.secretName` mounts an existing certificate Secret (the cert-manager seam; a **Cert Manager Certificate** secret output plugs in directly). AMQPS moves to port 5671 and the management UI to 15671; `caSecretName` adds mutual TLS; `disableNonTlsListeners` closes every plain port, including the plain ports of enabled plugins. The exported endpoints follow the listener posture.

**Plugins and config roll the cluster** — `configuration.additionalPlugins` extends the always-on essentials (rabbitmq_management, rabbitmq_prometheus, rabbitmq_peer_discovery_k8s) with e.g. rabbitmq_mqtt, rabbitmq_stomp, rabbitmq_shovel, rabbitmq_federation, rabbitmq_stream. `additionalConfig` appends rabbitmq.conf lines to the operator-generated file; `advancedConfig`, `envConfig`, and `erlangInetConfig` cover the rest of RabbitMQ's own vocabulary. Changing any of them triggers a rolling restart — plan config changes like the restarts they are.

**Exposure is the Service, not an ingress** — no ingress resources are created. The client Service's `service.type` (cluster_ip / load_balancer / node_port) plus `service.annotations` (NLB annotations, external-dns hostname, ...) are the cloud-exposure surface; in-cluster consumers compose over the exported endpoints.

**Termination grace is 7 days on purpose** — the operator default (604800 seconds) gives a draining node however long it needs to rebalance and hand off cleanly rather than lose messages. Lower `terminationGracePeriodSeconds` for dev clusters where fast teardown matters more than clean handoff.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesStorageClass** | `storageClass` | `metadata.name` |
| **KubernetesCertificate** | `tls.secretName` | `status.outputs.secret_name` |
| **KubernetesSecret** | `tls.caSecretName` | `metadata.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the cluster runs in | Application deployment manifests |
| `cluster_name` | Name of the RabbitmqCluster resource (= metadata.name) | Operational tooling |
| `service_name` | The client Service (operator naming contract: `<name>`) | Ingress/Gateway composition |
| `headless_service_name` | The headless inter-node Service (`<name>-nodes`) | Diagnostics |
| `amqp_endpoint` | In-cluster AMQP endpoint (port 5672; 5671 when `tls.disableNonTlsListeners` closes the plain listener) | Application client configuration |
| `management_endpoint` | In-cluster management API / UI endpoint (http on 15672; https on 15671 when the plain listener is closed) | Admin tooling, HTTP API automation |
| `default_user_secret_name` | The operator-generated admin credentials Secret (`<name>-default-user` — keys: username, password, host, port, connection_string, ...); EMPTY when the Vault backend is used | Application pod env via `secretKeyRef` |
| `port_forward_command` | Copy-paste `kubectl port-forward` for the management UI | Local development access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev single node** — one node on ephemeral storage (emptyDir — no PVC involved), small memory with requests equal to limits, and a 60-second termination grace period in place of the 7-day production default: a real broker for developers and CI without production ceremony. Start from the **Dev single node preset**.

**Production quorum** — 3 nodes (the quorum posture), a 50Gi data volume per node, 4Gi of memory with requests equal to limits, required anti-affinity so a node loss takes one broker instead of the quorum, and automatic feature-flag enablement so upgrades never stall. Start from the **Production quorum preset**.

**MQTT IoT broker** — a device-fleet broker: 3 nodes, the rabbitmq_mqtt and rabbitmq_web_mqtt plugins on the shared broker core, and a LoadBalancer client Service so devices outside the cluster get a reachable address. Start from the **MQTT IoT broker preset**.

## Works With

- [**RabbitMQ Cluster Operator**](/infra-catalog/kubernetes-rabbit-mq-operator) — the engine that reconciles this cluster; deploy it first, keep the cluster inside its watched namespaces, and never destroy the operator while clusters exist (its CRD cascade-deletes them).
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — referenced placement; the InfraPipeline orders namespace-first.
- [**Kubernetes StorageClass**](/infra-catalog/kubernetes-storage-class) — SSD-backed classes for the per-node data volumes.
- [**Cert Manager**](/infra-catalog/kubernetes-cert-manager) and [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) — issue the TLS Secret the `tls` block references; the certificate must cover the client Service DNS names.
- [**External Secret**](/infra-catalog/kubernetes-external-secret) — materializes default-user credentials from a cloud secret manager; reference its target Secret in `secretBackend.externalSecretName`.
- [**Apache Kafka**](/infra-catalog/kubernetes-kafka) — the event-streaming sibling for append-only, replayable logs at scale; reach for it when consumers replay history instead of acknowledging work.
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) and [**Kubernetes CronJob**](/infra-catalog/kubernetes-cron-job) — the workloads that consume the exported `amqp_endpoint` and the default-user Secret for task queues, work distribution, and RPC.
