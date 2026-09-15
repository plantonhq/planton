# KubernetesNats

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**KubernetesNatsSpec** deploys NATS — the lightweight, high-speed
messaging system (pub/sub, request/reply, queue groups) with
JetStream persistence (streams, consumers, key-value and object
stores) — from the official `nats` Helm chart
(https://nats-io.github.io/k8s/helm/charts/).

WHAT GETS INSTALLED: the NATS server StatefulSet (with a config
hot-reload sidecar), its client Service, and by default the
nats-box utility pod (a shell with the `nats` CLI pre-configured
for this deployment).

JETSTREAM IS ON BY DEFAULT here (the chart leaves it off): each
server gets a persistent volume for stream data, so published
messages survive pod restarts. A single server is a complete
JetStream deployment for dev; REPLICATED streams (R3) need
`cluster` enabled with at least 3 servers.

SECURITY: with `auth` unset the server accepts unauthenticated
connections — fine inside a trusted cluster network, never for
anything reachable from outside. Declare users (flat, or grouped
into accounts for multi-tenant isolation) and the module GENERATES
their passwords, exports them in the `<name>-auth` Secret (one key
per username), and wires the server to read them from environment —
no password ever lands in the rendered config or Helm values.

EXPOSURE: the client Service is ClusterIP by default; clients in
the cluster connect through `client_endpoint`. For external
clients, set `service` to LoadBalancer with your cloud's
annotations, or use `websocket` behind first-class exposure kinds.

DECLARING STREAMS: this kind deploys the SERVER. Streams, consumers
and KV buckets are data-plane objects — create them from
applications (any NATS SDK), the nats CLI, or nats-box. A
declarative stream-as-resource surface (the NACK controller) is a
separate concern this kind deliberately does not bundle.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart values
beyond them (merged last, Helm `-f` semantics, identical on both
engines) — gateways (superclusters), the JWT/operator auth mode and
its resolver, per-listener TLS, raw nats.conf keys via the chart's
config merge — a safety valve, never the primary interface. Never
put secret material in `helm_values`: chart config values render
into a ConfigMap; every credential path in this spec rides Secrets
and environment expansion instead.

## Example

```yaml
# Full-surface shape: a 3-server JetStream cluster with flat authenticated
# users (module-generated passwords), websocket, metrics, LoadBalancer
# exposure and an escape-hatch entry — the offline plan/preview proof for
# the widest typed rendering.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesNats
metadata:
  name: nats-full
spec:
  namespace:
    value: nats-full
  createNamespace: true
  cluster:
    enabled: true
    replicas: 3
  jetStream:
    diskSize: 20Gi
    maxFileStore: 18Gi
    memoryStoreMaxSize: 1Gi
  auth:
    users:
      - username: orders-service
        permissions:
          publishAllow:
            - "orders.>"
          subscribeAllow:
            - "orders.>"
            - "_INBOX.>"
      - username: auditor
        permissions:
          subscribeAllow:
            - ">"
          publishDeny:
            - ">"
    noAuthUser: auditor
  websocket:
    enabled: true
    port: 8080
  metrics:
    exporterEnabled: true
    podMonitorEnabled: true
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      memory: 2Gi
  service:
    type: load_balancer
    annotations:
      external-dns.alpha.kubernetes.io/hostname: nats.example.com
  scheduling:
    nodeSelector:
      workload-tier: messaging
  helmValues: |
    podTemplate:
      patch:
        - op: add
          path: /metadata/annotations
          value:
            team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `2.14.2` |  |
| `spec.cluster` | `KubernetesNatsCluster` |  |  |  |
| `spec.cluster.enabled` | `bool` |  |  |  |
| `spec.cluster.replicas` | `int32` |  | `3` |  |
| `spec.jetStream` | `KubernetesNatsJetStream` |  |  |  |
| `spec.jetStream.enabled` | `bool` |  | `true` |  |
| `spec.jetStream.diskSize` | `string` |  | `10Gi` |  |
| `spec.jetStream.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.jetStream.maxFileStore` | `string` |  |  |  |
| `spec.jetStream.memoryStoreMaxSize` | `string` |  |  |  |
| `spec.auth` | `KubernetesNatsAuth` |  |  |  |
| `spec.auth.users` | `[]KubernetesNatsUser` |  |  |  |
| `spec.auth.users[].username` | `string` | yes |  |  |
| `spec.auth.users[].permissions` | `KubernetesNatsPermissions` |  |  |  |
| `spec.auth.users[].permissions.publishAllow` | `[]string` |  |  |  |
| `spec.auth.users[].permissions.publishDeny` | `[]string` |  |  |  |
| `spec.auth.users[].permissions.subscribeAllow` | `[]string` |  |  |  |
| `spec.auth.users[].permissions.subscribeDeny` | `[]string` |  |  |  |
| `spec.auth.accounts` | `[]KubernetesNatsAccount` |  |  |  |
| `spec.auth.accounts[].name` | `string` | yes |  |  |
| `spec.auth.accounts[].users` | `[]KubernetesNatsUser` | yes |  |  |
| `spec.auth.accounts[].users[].username` | `string` | yes |  |  |
| `spec.auth.accounts[].users[].permissions` | `KubernetesNatsPermissions` |  |  |  |
| `spec.auth.accounts[].users[].permissions.publishAllow` | `[]string` |  |  |  |
| `spec.auth.accounts[].users[].permissions.publishDeny` | `[]string` |  |  |  |
| `spec.auth.accounts[].users[].permissions.subscribeAllow` | `[]string` |  |  |  |
| `spec.auth.accounts[].users[].permissions.subscribeDeny` | `[]string` |  |  |  |
| `spec.auth.accounts[].jetStreamEnabled` | `bool` |  |  |  |
| `spec.auth.noAuthUser` | `string` |  |  |  |
| `spec.tls` | `KubernetesNatsTls` |  |  |  |
| `spec.tls.secretName` | `string \| valueFrom` | yes |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.tls.verifyClients` | `bool` |  |  |  |
| `spec.websocket` | `KubernetesNatsWebsocket` |  |  |  |
| `spec.websocket.enabled` | `bool` |  |  |  |
| `spec.websocket.port` | `int32` |  | `8080` |  |
| `spec.mqtt` | `KubernetesNatsMqtt` |  |  |  |
| `spec.mqtt.enabled` | `bool` |  |  |  |
| `spec.mqtt.port` | `int32` |  | `1883` |  |
| `spec.leafnodes` | `KubernetesNatsLeafnodes` |  |  |  |
| `spec.leafnodes.enabled` | `bool` |  |  |  |
| `spec.leafnodes.port` | `int32` |  | `7422` |  |
| `spec.metrics` | `KubernetesNatsMetrics` |  |  |  |
| `spec.metrics.exporterEnabled` | `bool` |  |  |  |
| `spec.metrics.podMonitorEnabled` | `bool` |  |  |  |
| `spec.natsBoxEnabled` | `bool` |  | `true` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.service` | `KubernetesNatsService` |  |  |  |
| `spec.service.type` | `enum` |  |  |  |
| `spec.service.annotations` | `map<string, string>` |  |  |  |
| `spec.scheduling` | `KubernetesNatsScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.images` | `KubernetesNatsImages` |  |  |  |
| `spec.images.nats` | `ContainerImage` |  |  |  |
| `spec.images.nats.repo` | `string` |  |  |  |
| `spec.images.nats.tag` | `string` |  |  |  |
| `spec.images.nats.pullSecretName` | `string` |  |  |  |
| `spec.images.reloader` | `ContainerImage` |  |  |  |
| `spec.images.reloader.repo` | `string` |  |  |  |
| `spec.images.reloader.tag` | `string` |  |  |  |
| `spec.images.reloader.pullSecretName` | `string` |  |  |  |
| `spec.images.exporter` | `ContainerImage` |  |  |  |
| `spec.images.exporter.repo` | `string` |  |  |  |
| `spec.images.exporter.tag` | `string` |  |  |  |
| `spec.images.exporter.pullSecretName` | `string` |  |  |  |
| `spec.images.natsBox` | `ContainerImage` |  |  |  |
| `spec.images.natsBox.repo` | `string` |  |  |  |
| `spec.images.natsBox.tag` | `string` |  |  |  |
| `spec.images.natsBox.pullSecretName` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the
resource. When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "2.14.2" — the chart and the
NATS server move in version lockstep, so chart 2.14.2 runs
nats-server 2.14.2). Versions must exist as SERVED charts in the
repository index (https://nats-io.github.io/k8s/helm/charts/).

- default: `2.14.2`

### spec.cluster

`KubernetesNatsCluster`

Multi-server clustering (full-mesh routes between servers). Empty
= a single server — a complete deployment for dev and modest
workloads. Enable for HA and for replicated JetStream streams
(stream replicas can never exceed the server count).

### spec.cluster.enabled

`bool`

Enable clustering (a full mesh of routes across the StatefulSet's
servers).

### spec.cluster.replicas

`int32` · optional (explicit presence)

Number of NATS servers. Use an odd count — JetStream placement
uses RAFT groups, and odd counts tolerate the most failures per
server added (3 tolerates 1 down, 5 tolerates 2). Minimum 2 when
clustering with JetStream (the chart's own floor); 3 is the
smallest count that keeps replicated streams available through a
pod loss. Empty = 3.

- default: `3`
- rule: {"int32":{"lte":9,"gte":2}}

### spec.jetStream

`KubernetesNatsJetStream`

JetStream — the persistence layer (streams, consumers, KV/object
stores). ON by default with a persistent file store; core NATS
pub/sub works either way.

### spec.jetStream.enabled

`bool` · optional (explicit presence)

Enable JetStream. Empty = true — this kind's default posture is
persistent messaging (the chart's raw default is off).

- default: `true`

### spec.jetStream.diskSize

`string` · optional (explicit presence)

Size of each server's persistent volume for the file store —
where stream data lives. Empty = "10Gi" (the chart default).

- default: `10Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.jetStream.storageClass

`string | valueFrom`

Storage class for the JetStream volumes. Empty = the cluster's
default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.jetStream.maxFileStore

`string` · optional (explicit presence)

Cap on total file-store usage per server (e.g. "8Gi"). Empty =
the volume size (the chart derives it from the PVC).

- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.jetStream.memoryStoreMaxSize

`string` · optional (explicit presence)

Enable the in-memory store tier and cap its size per server
(e.g. "1Gi"). Memory streams are fast and ephemeral — data is
gone on pod restart. Empty = memory store disabled. The server's
memory limit must comfortably exceed this.

- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.auth

`KubernetesNatsAuth`

Client authentication. Empty = UNAUTHENTICATED — any client that
can reach the Service connects with full access. Declare flat
`users` or multi-tenant `accounts` (never both); passwords are
module-generated and exported via the `<name>-auth` Secret.

- rule: declare flat `users` or multi-tenant `accounts`, never both — with accounts defined, every user must belong to an account (nats-server rejects a config mixing top-level authorization users with accounts)
- rule: auth is declared but empty — declare `users` or `accounts` (or omit auth entirely for an unauthenticated server)
- rule: no_auth_user must be one of the declared usernames (in users or in an account)
- rule: usernames must be unique across the whole auth block (including across accounts) — each becomes a key in the generated auth Secret

### spec.auth.users

`[]KubernetesNatsUser`

Flat users on the global account.

### spec.auth.users[].username

`string` · required

Username (also the key in the auth Secret).

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$"}}

### spec.auth.users[].permissions

`KubernetesNatsPermissions`

Subject-level permissions. Empty = full publish/subscribe on the
user's account.

### spec.auth.users[].permissions.publishAllow

`[]string`

Subjects the user MAY publish to. An allowlist fences the user
from EVERYTHING else — including JetStream, which is driven by
requests published to `$JS.API.>` and acknowledgements published
to `$JS.ACK.>`. A publish-allowlisted user who needs JetStream
must include both (responses arrive on `_INBOX.>` subscriptions);
the server silently drops denied publishes, so a fenced user's
stream operations hang until the client times out rather than
failing loudly (verified live: stream creation as an allowlisted
user times out with no server-side error surfaced to the client).

### spec.auth.users[].permissions.publishDeny

`[]string`

Subjects the user may NOT publish to.

### spec.auth.users[].permissions.subscribeAllow

`[]string`

Subjects the user MAY subscribe to.

### spec.auth.users[].permissions.subscribeDeny

`[]string`

Subjects the user may NOT subscribe to.

### spec.auth.accounts

`[]KubernetesNatsAccount`

Multi-tenant accounts, each with its own users and an isolated
subject namespace (messages never cross accounts unless you add
exports/imports via `helm_values`).

### spec.auth.accounts[].name

`string` · required

Account name.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$"}}

### spec.auth.accounts[].users

`[]KubernetesNatsUser` · required

Users belonging to this account. At least one — an account
nobody can connect to configures nothing.

- rule: {"repeated":{"minItems":"1"}}

### spec.auth.accounts[].users[].username

`string` · required

Username (also the key in the auth Secret).

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$"}}

### spec.auth.accounts[].users[].permissions

`KubernetesNatsPermissions`

Subject-level permissions. Empty = full publish/subscribe on the
user's account.

### spec.auth.accounts[].users[].permissions.publishAllow

`[]string`

Subjects the user MAY publish to. An allowlist fences the user
from EVERYTHING else — including JetStream, which is driven by
requests published to `$JS.API.>` and acknowledgements published
to `$JS.ACK.>`. A publish-allowlisted user who needs JetStream
must include both (responses arrive on `_INBOX.>` subscriptions);
the server silently drops denied publishes, so a fenced user's
stream operations hang until the client times out rather than
failing loudly (verified live: stream creation as an allowlisted
user times out with no server-side error surfaced to the client).

### spec.auth.accounts[].users[].permissions.publishDeny

`[]string`

Subjects the user may NOT publish to.

### spec.auth.accounts[].users[].permissions.subscribeAllow

`[]string`

Subjects the user MAY subscribe to.

### spec.auth.accounts[].users[].permissions.subscribeDeny

`[]string`

Subjects the user may NOT subscribe to.

### spec.auth.accounts[].jetStreamEnabled

`bool`

Let this account use JetStream (streams/consumers/KV are
per-account with `accounts` defined). Requires `jet_stream` on
the spec.

### spec.auth.noAuthUser

`string`

Username connections WITHOUT credentials are treated as (the
"guest" identity). Must be one of the declared usernames — scope
what anonymous clients may do through THAT user's permissions.
Empty = unauthenticated connections are rejected outright.

### spec.tls

`KubernetesNatsTls`

TLS on the client listener (port 4222). Websocket, cluster and
leafnode listener TLS ride `helm_values` (each listener has its
own tls block in the chart).

### spec.tls.secretName

`string | valueFrom` · required

Name of the existing TLS Secret in the install namespace.
Accepts a literal name or a KubernetesCertificate reference.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.tls.verifyClients

`bool`

Require clients to present certificates signed by the CA (mutual
TLS). The CA bundle must be present in the Secret (`ca.crt` —
cert-manager includes it) for verification to work.

### spec.websocket

`KubernetesNatsWebsocket`

The WebSocket listener — NATS over websockets for browser clients
and networks that only pass HTTP. Off by default.

### spec.websocket.enabled

`bool`

Enable NATS over WebSocket.

### spec.websocket.port

`int32` · optional (explicit presence)

WebSocket port. Empty = 8080.

- default: `8080`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.mqtt

`KubernetesNatsMqtt`

The MQTT listener — IoT devices speak MQTT 3.1.1 directly to
NATS, bridged into JetStream. Requires `jet_stream` (MQTT
sessions and retained messages live in JetStream). Off by
default.

### spec.mqtt.enabled

`bool`

Enable the MQTT bridge.

### spec.mqtt.port

`int32` · optional (explicit presence)

MQTT port. Empty = 1883.

- default: `1883`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.leafnodes

`KubernetesNatsLeafnodes`

The leafnode listener — edge/remote NATS servers extend this
deployment by connecting as leaf nodes (the hub side). Off by
default.

### spec.leafnodes.enabled

`bool`

Accept leafnode connections.

### spec.leafnodes.port

`int32` · optional (explicit presence)

Leafnode port. Empty = 7422.

- default: `7422`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.metrics

`KubernetesNatsMetrics`

Prometheus metrics. The server's monitoring endpoint (port 8222)
is always on in-pod; this block adds the prometheus-nats-exporter
sidecar and, optionally, a PodMonitor for operator-based scraping.

- rule: pod_monitor_enabled requires exporter_enabled — the PodMonitor scrapes the exporter sidecar's endpoint

### spec.metrics.exporterEnabled

`bool`

Run the prometheus-nats-exporter sidecar (scrape endpoint on port
7777 of each server pod).

### spec.metrics.podMonitorEnabled

`bool`

Create a PodMonitor so a Prometheus Operator scrapes the
exporter. Requires the Prometheus Operator CRDs on the cluster —
a KubernetesKubePrometheusStack composes naturally. Requires
`exporter_enabled`.

### spec.natsBoxEnabled

`bool` · optional (explicit presence)

Deploy the nats-box utility pod — a shell with the `nats` CLI
pre-wired to this deployment (contexts, credentials). Enabled by
default, matching the chart; the go-to surface for creating
streams and debugging.

- default: `true`

### spec.resources

`ContainerResources`

Container resources for each NATS server. Empty = the chart's
defaults (no requests — fine for dev). When `jet_stream` uses a
memory store, set a memory limit comfortably above
`memory_store_max_size`.

### spec.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.resources.limits.cpu

`string`

### spec.resources.limits.memory

`string`

### spec.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.resources.requests.cpu

`string`

### spec.resources.requests.memory

`string`

### spec.service

`KubernetesNatsService`

The client Service's exposure shape. Empty = ClusterIP (in-cluster
clients only).

### spec.service.type

`enum`

Service type. Default ClusterIP.

Allowed values (use exactly as shown):

- `cluster_ip` -- ClusterIP (the default) — in-cluster clients only; compose external exposure from first-class kinds or switch the type below.
- `load_balancer` -- LoadBalancer — the cloud provisions an external address; combine with `annotations` for the cloud's LB controller (internal LBs, NLB mode, external-dns hostnames).
- `node_port` -- NodePort — every cluster node forwards a high port.

### spec.service.annotations

`map<string, string>`

Annotations on the client Service — the cloud-integration surface
(LB class/scheme, external-dns hostnames, internal-LB flags).

### spec.scheduling

`KubernetesNatsScheduling`

Pod scheduling for the NATS server pods.

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the server pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the server pods.

### spec.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.images

`KubernetesNatsImages`

Container image overrides for air-gapped clusters and private
mirrors. Empty = the chart's pinned upstream images.

### spec.images.nats

`ContainerImage`

The NATS server image (upstream: nats, alpine tag).

### spec.images.nats.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.images.nats.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.images.nats.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.images.reloader

`ContainerImage`

The config-reloader sidecar image (upstream:
natsio/nats-server-config-reloader).

### spec.images.reloader.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.images.reloader.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.images.reloader.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.images.exporter

`ContainerImage`

The metrics exporter image (upstream:
natsio/prometheus-nats-exporter).

### spec.images.exporter.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.images.exporter.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.images.exporter.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.images.natsBox

`ContainerImage`

The nats-box image (upstream: natsio/nats-box).

### spec.images.natsBox.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.images.natsBox.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.images.natsBox.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.helmValues

`string`

Additional Helm values merged LAST (Helm `-f` semantics, identical
on both engines) — the escape hatch for chart values the typed
fields do not model: gateways, the JWT resolver, per-listener TLS,
raw nats.conf keys through `config.merge`, pod-template patches.
YAML document as a string. Never put secret material here (config
values render into a ConfigMap); credentials belong in the typed
auth model, which keeps them in Secrets.

## Validation Rules

- `spec.mqtt.requires_jetstream`: mqtt requires jet_stream to be enabled — MQTT sessions and retained messages are stored in JetStream, and the server refuses to start MQTT without it

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesNats, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the servers run in. |
| `status.outputs.service_name` | `string` | Name of the client Service (equals metadata.name). |
| `status.outputs.headless_service_name` | `string` | Name of the headless Service (`<name>-headless`) — per-server DNS for clients that need direct server addressing. |
| `status.outputs.client_endpoint` | `string` | In-cluster client endpoint, `nats://<name>.<namespace>.svc.cluster.local:4222` — what NATS clients set as their server URL. |
| `status.outputs.websocket_endpoint` | `string` | In-cluster WebSocket endpoint (`ws://<name>.<namespace>.svc.cluster.local:<port>`); empty when the websocket listener is off. |
| `status.outputs.auth_secret_name` | `string` | Name of the module-generated auth Secret (`<name>-auth`, one key per declared username holding that user's password); empty when auth is not declared. |
| `status.outputs.port_forward_command` | `string` | Port-forward command for reaching the client port from a workstation when no exposure is composed (`kubectl port-forward svc/<name> -n <namespace> 4222:4222`). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.jetStream.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.tls.secretName` | KubernetesCertificate | `status.outputs.secret_name` |

## See Also

- [Overview](../README.md)
