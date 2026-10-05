# Kubernetes RabbitMQ

## When NOT to Use This

**RabbitMQ is the queue-messaging broker, not the event-streaming
log.** This component serves the task-queue / work-distribution / RPC
role — per-message acknowledgement and routing over AMQP 0-9-1/1.0,
with MQTT and STOMP via plugins. For append-only event STREAMING and
replayable logs at scale, use KubernetesKafka instead: consumers that
re-read history are a log's job, not a queue's.

Also not the right kind when:

- **You want the operator itself** — installing and configuring the
  RabbitMQ Cluster Operator is KubernetesRabbitMqOperator; this
  component is one cluster it manages. The operator is a registry
  prerequisite — declare it first (it watches ALL namespaces by
  default, so one operator serves every cluster).
- **You want queues, exchanges and users as Kubernetes resources** —
  that is the messaging-topology-operator's surface, deliberately not
  part of this catalog today. Topology stays a data operation,
  declared through RabbitMQ's own interfaces.
- **You want a single-container dev broker without an operator** —
  run the rabbitmq image as a KubernetesDeployment.
- **You expect a public endpoint out of the box** — no ingress
  resources are created here. The client Service's `service.type`
  (cluster_ip / load_balancer / node_port) plus `service.annotations`
  are the cloud-exposure surface; in-cluster consumers compose over
  the exported endpoints.

## Overview

**KubernetesRabbitMq** declares one RabbitMQ cluster — the
queue-messaging broker behind task queues, work distribution, RPC and
per-message routing — as a `RabbitmqCluster` custom resource
reconciled by the RabbitMQ Cluster Operator. The operator renders the
cluster as one StatefulSet plus two Services and manages rolling
restarts and upgrades around them.

**The naming contract** (operator source, verified at the pinned
release): the client Service is `<name>` (ClusterIP by default), the
headless inter-node Service is `<name>-nodes`, the generated admin
credentials Secret is `<name>-default-user` (keys: username,
password, host, port, connection_string, ...), the StatefulSet is
`<name>-server`, and each pod's data volume claim is
`persistence-<name>-server-<i>`. This is what makes the exported
endpoints and handles deterministic.

**The credentials truth**: the operator GENERATES the administrator
credentials — they never pass through this spec. Consumers read the
`<name>-default-user` Secret, exported as the
`default_user_secret_name` output. Setting `default_user` /
`default_pass` lines through `configuration.additional_config`
overrides the generated credentials but puts a plaintext password on
the CR — migration-only, discouraged.

**Key design points:**

- **Odd replica counts in production.** Quorum queues and the
  Raft-based metadata store survive node loss with 3, 5 or 7 nodes; a
  2-node cluster loses availability when EITHER node fails. The
  operator does NOT support scaling down (removed brokers strand
  their queue replicas) — size down by migrating to a new cluster.
- **Memory requests = limits.** RabbitMQ derives its memory high
  watermark from the container memory LIMIT — always set requests and
  limits to the SAME memory value, or the broker's flow control
  triggers at the wrong threshold. Operator defaults: requests
  1 CPU / 2Gi, limits 2 CPU / 2Gi.
- **The image must be a `-management` variant** (default
  `rabbitmq:4.2.6-management` at the pinned operator release) — the
  operator's generated configuration expects the management plugin,
  and the plain `rabbitmq:<version>` tags do not carry it. Always-on
  essential plugins: rabbitmq_management, rabbitmq_prometheus,
  rabbitmq_peer_discovery_k8s; everything further (rabbitmq_shovel,
  rabbitmq_federation, rabbitmq_mqtt, rabbitmq_stomp,
  rabbitmq_stream) rides `configuration.additional_plugins`.
- **Storage is per node** — `disk_size` (default 10Gi) carries
  queues, quorum-queue Raft logs and the metadata store; PVCs cannot
  shrink, plan for growth. `storage_class` accepts a
  KubernetesStorageClass reference. `ephemeral: true` is the CR's
  storage-0 + emptyDir posture — everything vanishes with each pod
  restart, throwaway dev only — and is mutually exclusive with
  `disk_size` / `storage_class` (CEL-enforced).
- **TLS is the cert-manager seam** — `tls.secret_name` references a
  kubernetes.io/tls Secret (a KubernetesCertificate's secret output);
  `ca_secret_name` (key ca.crt) adds MUTUAL TLS;
  `disable_non_tls_listeners: true` closes every plain port — AMQPS
  moves to 5671, management to 15671, and the plain ports of any
  enabled plugin (MQTT, STOMP, their WebSocket forms) close with
  them. The certificate must cover the `<name>.<namespace>.svc` DNS
  forms.
- **Configuration is layered, not escape-hatched** — typed fields
  cover topology, storage, TLS and placement; RabbitMQ's own
  configuration vocabulary flows through
  `configuration.additional_config` (rabbitmq.conf lines appended to
  the operator-generated file) and its siblings (`advanced_config`,
  `env_config`, `erlang_inet_config`) — the upstream CRD's own model,
  not an escape hatch bolted on top of one.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to deploy into — literal or a
  KubernetesNamespace reference (`create_namespace` to own it); the
  operator watches all namespaces by default

### Common

- **`spec.replicas`**: node count (default 1 — dev only); production
  runs an ODD count (3, 5, 7); scaling down is not supported
- **`spec.disk_size` / `spec.storage_class`**: the per-node data
  volume (default 10Gi) and its StorageClass (literal or a
  KubernetesStorageClass reference; empty = the cluster default)
- **`spec.ephemeral`**: run on emptyDir — throwaway dev only;
  excludes `disk_size` / `storage_class`
- **`spec.resources`**: CPU/memory per node — same memory value for
  requests and limits (the watermark rule); empty = operator defaults
- **`spec.configuration`**: `additional_plugins`,
  `additional_config` (rabbitmq.conf lines), `advanced_config`,
  `env_config`, `erlang_inet_config` — changing plugins or config
  rolls the cluster
- **`spec.tls`**: existing certificate Secret (`secret_name`),
  optional mutual-TLS CA (`ca_secret_name`), and
  `disable_non_tls_listeners` to close the plain ports
- **`spec.service`**: the client Service's type (default ClusterIP),
  annotations (the cloud-exposure surface — NLB annotations,
  external-dns hostname), labels, dual-stack policy
- **`spec.spread_across_nodes`**: REQUIRED pod anti-affinity over
  hostnames — off by default so single-node dev clusters schedule; on
  in production (a cluster with more replicas than schedulable nodes
  sits Pending)
- **`spec.node_selector` / `spec.tolerations`**: placement —
  `node_selector` renders as REQUIRED node affinity (the CR has no
  nodeSelector field; behaviorally identical for exact matches, same
  shape on both engines)
- **`spec.termination_grace_period_seconds`**: operator default
  604800 (7 DAYS — deliberately generous so draining nodes never lose
  messages); lower for dev clusters where fast teardown matters more
- **`spec.delay_start_seconds`**: startup readiness delay (operator
  default 30 — the DNS-propagation guard)
- **`spec.skip_post_deploy_steps` /
  `spec.auto_enable_all_feature_flags`**: skip post-upgrade queue
  rebalancing / enable all stable feature flags after upgrades
  (upgrade-readiness)
- **`spec.secret_backend`**: `vault` (role, default_user_path, agent
  annotations, optional pki_issuer_path for Vault-issued TLS; the
  credential-updater sidecar rotates) or `external_secret_name` (a
  pre-existing Secret, e.g. materialized by KubernetesExternalSecret)
- **`spec.image` / `spec.image_pull_secrets`**: the air-gap /
  private-mirror path — the image must stay a `-management` variant

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the cluster runs in |
| `cluster_name` | Name of the RabbitmqCluster resource (= `metadata.name`) |
| `service_name` | The client Service (operator contract: `<name>`) |
| `headless_service_name` | The inter-node Service (`<name>-nodes`) |
| `amqp_endpoint` | In-cluster AMQP endpoint, port 5672 (5671 when the plain listeners are closed) |
| `management_endpoint` | In-cluster management API / UI endpoint (http 15672 / https 15671) |
| `default_user_secret_name` | The operator-generated `<name>-default-user` Secret; EMPTY when the Vault backend owns credentials |
| `port_forward_command` | Port-forward command for the management UI from a workstation |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace); **`storage_class`** references a
  KubernetesStorageClass; **`tls.secret_name`** accepts a
  KubernetesCertificate reference (its issued Secret already carries
  `tls.crt`/`tls.key` — the shape the CR consumes).
- **Applications consume the outputs**: `amqp_endpoint` as the broker
  address, credentials from `default_user_secret_name` as env-from
  references (the Secret even carries a ready-made
  `connection_string`) — credentials ride the Secret, never the
  manifest.
- **Exposure composes, never embeds**: `service.type: load_balancer`
  plus the cloud's annotations for external clients; everything else
  stays in-cluster over the exported endpoints.
- **The operator is a registry prerequisite**, not a reference:
  declare KubernetesRabbitMqOperator first — it watches all
  namespaces by default.

## Examples

### Development (single node, ephemeral)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRabbitMq
metadata:
  name: dev-rabbitmq
spec:
  namespace:
    value: dev-rabbitmq
  create_namespace: true
  ephemeral: true
  resources:
    requests: { cpu: 500m, memory: 1Gi }
    limits: { cpu: "1", memory: 1Gi }
  termination_grace_period_seconds: 60
```

One node on emptyDir: every queue, message and user vanishes with
each pod restart — that trade-off is the point of a throwaway dev
broker. The grace period drops from the operator's 7-day default to
60 seconds so teardown is fast.

### Production (3-node quorum posture)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRabbitMq
metadata:
  name: prod-rabbitmq
spec:
  namespace:
    value: messaging
  create_namespace: true
  replicas: 3
  disk_size: 50Gi
  resources:
    requests: { cpu: "1", memory: 4Gi }
    limits: { cpu: "2", memory: 4Gi }
  spread_across_nodes: true
  auto_enable_all_feature_flags: true
```

Three nodes — an odd count, so quorum queues and the metadata store
survive one node loss — forced onto different Kubernetes nodes, with
the same memory value on requests and limits so the memory high
watermark lands where the limit says. Remember the one-way door:
scale UP is a `replicas` change, scale DOWN is a migration.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
