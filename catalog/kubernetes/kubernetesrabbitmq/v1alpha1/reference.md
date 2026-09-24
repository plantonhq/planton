# KubernetesRabbitMq

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesRabbitMqSpec** declares one RabbitMQ cluster — the
queue-messaging broker (AMQP 0-9-1/1.0, MQTT, STOMP) — as a
`RabbitmqCluster` custom resource reconciled by the RabbitMQ Cluster
Operator (declare the operator with KubernetesRabbitMqOperator; it
is a registry prerequisite of this kind). RabbitMQ serves the
task-queue / work-distribution / RPC role — for append-only event
STREAMING at scale, use KubernetesKafka instead.

The operator renders the cluster as one StatefulSet plus two
Services and generates the administrator credentials itself — they
never pass through this spec (see the default-user Secret in the
outputs).

NAMING CONTRACT (operator source, verified at the pinned release):
the client Service is `<name>` (ClusterIP by default), the headless
inter-node Service is `<name>-nodes`, the generated admin
credentials live in Secret `<name>-default-user` (keys: username,
password, host, port, connection_string, ...), and each pod's data
volume claim is `persistence-<name>-server-<i>`.

REPLICAS AND QUORUM: production clusters use an ODD replica count
(3, 5, 7) so quorum queues and the Raft-based metadata store
survive node loss; 2-replica clusters lose availability when either
node fails. SCALING DOWN is not supported by the operator (removed
brokers strand their queue replicas) — size down by migrating to a
new cluster.

EXPOSURE: no ingress resources are created here. The client Service
type and annotations are the cloud-exposure surface (set
service.type load_balancer plus the cloud's annotations); in-cluster
consumers compose over the exported endpoints.

CONFIGURATION LAYERS: fully typed fields cover topology, storage,
TLS, and placement. RabbitMQ's own configuration vocabulary flows
through `configuration.additional_config` (rabbitmq.conf lines,
appended to the operator-generated file) and its siblings — the
upstream CRD's own model, not an escape hatch bolted on top of one.

## Example

```yaml
# Full-surface offline-proof manifest: exercises a 3-node quorum cluster
# with an image override and pull secrets, a LoadBalancer client Service
# with cloud annotations and dual-stack policy, persistent storage on an
# explicit class, resources, every configuration layer (plugins,
# rabbitmq.conf, advanced.config, rabbitmq-env.conf, Erlang inet), mutual
# TLS with closed plain listeners, tolerations + node selector + the
# spread-across-nodes anti-affinity, the tuning knobs at non-default
# values, and the Vault secret backend with PKI TLS — so the offline tofu
# plan and pulumi preview proofs cover the full typed surface.
# Placeholder values; never applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRabbitMq
metadata:
  name: rabbitmq-hack
spec:
  namespace:
    value: rabbitmq-hack
  createNamespace: true
  replicas: 3
  image:
    repo: mirror.internal/rabbitmq
    tag: 4.2.6-management
    pullSecretName: mirror-pull
  imagePullSecrets:
    - mirror-pull
  service:
    type: load_balancer
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: nlb
      external-dns.alpha.kubernetes.io/hostname: mq.example.internal
    labels:
      team: platform
    ipFamilyPolicy: prefer_dual_stack
  diskSize: 50Gi
  storageClass:
    value: fast-ssd
  resources:
    requests:
      cpu: "1"
      memory: 4Gi
    limits:
      cpu: "2"
      memory: 4Gi
  configuration:
    additionalPlugins:
      - rabbitmq_shovel
      - rabbitmq_shovel_management
      - rabbitmq_mqtt
    additionalConfig: |
      vm_memory_high_watermark.relative = 0.8
      consumer_timeout = 1800000
    advancedConfig: |
      [
        {rabbit, [{channel_max, 1024}]}
      ].
    envConfig: |
      RABBITMQ_LOGS=-
    erlangInetConfig: |
      {lookup, [dns]}.
  tls:
    secretName:
      value: rabbitmq-hack-tls
    caSecretName:
      value: rabbitmq-hack-ca
    disableNonTlsListeners: true
  tolerations:
    - key: dedicated
      operator: Equal
      value: messaging
      effect: NoSchedule
  spreadAcrossNodes: true
  nodeSelector:
    kubernetes.io/os: linux
  terminationGracePeriodSeconds: 3600
  delayStartSeconds: 10
  skipPostDeploySteps: true
  autoEnableAllFeatureFlags: true
  secretBackend:
    vault:
      role: rabbitmq
      defaultUserPath: secret/data/rabbitmq/config
      annotations:
        vault.hashicorp.com/template-static-secret-render-interval: 15s
      pkiIssuerPath: pki/issue/cert-issuer
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.image` | `ContainerImage` |  |  |  |
| `spec.image.repo` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.image.pullSecretName` | `string` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.service` | `KubernetesRabbitMqService` |  |  |  |
| `spec.service.type` | `enum` |  |  |  |
| `spec.service.annotations` | `map<string, string>` |  |  |  |
| `spec.service.labels` | `map<string, string>` |  |  |  |
| `spec.service.ipFamilyPolicy` | `enum` |  |  |  |
| `spec.diskSize` | `string` |  | `10Gi` |  |
| `spec.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.ephemeral` | `bool` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.configuration` | `KubernetesRabbitMqConfiguration` |  |  |  |
| `spec.configuration.additionalPlugins` | `[]string` |  |  |  |
| `spec.configuration.additionalConfig` | `string` |  |  |  |
| `spec.configuration.advancedConfig` | `string` |  |  |  |
| `spec.configuration.envConfig` | `string` |  |  |  |
| `spec.configuration.erlangInetConfig` | `string` |  |  |  |
| `spec.tls` | `KubernetesRabbitMqTls` |  |  |  |
| `spec.tls.secretName` | `string \| valueFrom` | yes |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.tls.caSecretName` | `string \| valueFrom` |  |  | KubernetesSecret (`metadata.name`) |
| `spec.tls.disableNonTlsListeners` | `bool` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.spreadAcrossNodes` | `bool` |  |  |  |
| `spec.terminationGracePeriodSeconds` | `int64` |  | `604800` |  |
| `spec.delayStartSeconds` | `int32` |  | `30` |  |
| `spec.skipPostDeploySteps` | `bool` |  |  |  |
| `spec.autoEnableAllFeatureFlags` | `bool` |  |  |  |
| `spec.secretBackend` | `KubernetesRabbitMqSecretBackend` |  |  |  |
| `spec.secretBackend.vault` | `KubernetesRabbitMqVaultBackend` |  |  |  |
| `spec.secretBackend.vault.role` | `string` | yes |  |  |
| `spec.secretBackend.vault.defaultUserPath` | `string` | yes |  |  |
| `spec.secretBackend.vault.annotations` | `map<string, string>` |  |  |  |
| `spec.secretBackend.vault.pkiIssuerPath` | `string` |  |  |  |
| `spec.secretBackend.externalSecretName` | `string` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to deploy the cluster into. Accepts a literal namespace
name or a reference to a KubernetesNamespace resource. The
RabbitmqCluster is namespaced; the operator must be watching this
namespace (see the operator kind's `watch_namespaces` — the
operator watches ALL namespaces by default).

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before deploying and deleted with the resource.
When false, the namespace must already exist.

### spec.replicas

`int32` · optional (explicit presence)

Number of RabbitMQ nodes. Default 1 (dev only — no redundancy).
Production: an ODD count (3, 5, 7) — see the quorum note on the
spec. The operator does NOT support scaling down.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.image

`ContainerImage`

Override the RabbitMQ server image. Empty = the operator's
default (`rabbitmq:4.2.6-management` at the pinned operator
release, or the operator kind's `default_rabbitmq_image`). KNOW
THIS: the image must be a `-management` variant — the operator's
generated configuration expects the management plugin, and the
plain `rabbitmq:<version>` tags do not carry it.

### spec.image.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.image.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.image.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.imagePullSecrets

`[]string`

Names of image-pull secrets (in the deployment namespace) for
pulling the server image from a private mirror.

### spec.service

`KubernetesRabbitMqService`

The client Service (`<name>`) — type, annotations, dual-stack
policy. Empty = ClusterIP with no annotations.

### spec.service.type

`enum`

Service type. Default ClusterIP.

Allowed values (use exactly as shown):

- `cluster_ip` -- ClusterIP (the operator default) — in-cluster access only; compose external exposure from first-class kinds or switch the type below.
- `load_balancer` -- LoadBalancer — the cloud provisions an external address; combine with `annotations` for the cloud's LB controller (internal LBs, NLB mode, ...).
- `node_port` -- NodePort — every cluster node forwards a high port.

### spec.service.annotations

`map<string, string>`

Annotations for the client Service — the cloud-exposure surface
(e.g. AWS NLB or internal-LB annotations, external-dns hostname).

### spec.service.labels

`map<string, string>`

Extra labels for the client Service.

### spec.service.ipFamilyPolicy

`enum`

Dual-stack policy. Unspecified = the cluster default.

Allowed values (use exactly as shown):

- `ip_family_policy_unspecified` -- Cluster default (SingleStack on most clusters).
- `single_stack` -- Exactly one IP family.
- `prefer_dual_stack` -- Dual-stack where available, single otherwise.
- `require_dual_stack` -- Dual-stack or fail.

### spec.diskSize

`string` · optional (explicit presence)

Size of the persistent data volume for EACH node (e.g. "10Gi" —
the operator default). Queues, quorum-queue Raft logs and the
metadata store live here. Kubernetes cannot shrink PVCs — plan
for growth. Ignored when `ephemeral` is true.

- default: `10Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.storageClass

`string | valueFrom`

Storage class for the data volumes. Accepts a literal name or a
reference to a KubernetesStorageClass resource. Empty = the
cluster's default storage class. Ignored when `ephemeral` is
true.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.ephemeral

`bool`

Run WITHOUT persistent storage (the CR's storage-0 + emptyDir
posture): every queue, message and user vanishes with each pod
restart. Throwaway dev/test clusters only.

### spec.resources

`ContainerResources`

CPU and memory for each RabbitMQ node container. Empty = the
operator defaults (requests 1 CPU / 2Gi, limits 2 CPU / 2Gi).
KNOW THIS: RabbitMQ derives its memory high watermark from the
container memory LIMIT — always set requests and limits to the
SAME memory value, or the broker's flow control triggers at the
wrong threshold.

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

### spec.configuration

`KubernetesRabbitMqConfiguration`

RabbitMQ server configuration — plugins, rabbitmq.conf /
advanced.config / rabbitmq-env.conf additions.

### spec.configuration.additionalPlugins

`[]string`

Plugins to enable IN ADDITION to the always-on essentials
(rabbitmq_management, rabbitmq_prometheus,
rabbitmq_peer_discovery_k8s) — e.g. rabbitmq_shovel,
rabbitmq_federation, rabbitmq_mqtt, rabbitmq_stomp,
rabbitmq_stream. Plugin names only (letters, digits, underscore);
changing the list rolls the cluster.

- rule: {"repeated":{"maxItems":"100","items":{"string":{"maxLen":"100","pattern":"^\\w+$"}}}}

### spec.configuration.additionalConfig

`string`

rabbitmq.conf lines APPENDED to the operator-generated file (ini
format, one `key = value` per line) — memory watermarks, default
vhost, consumer timeouts, ... KNOW THIS: `default_user` /
`default_pass` lines here override the operator-GENERATED admin
credentials and put the password in plaintext on the CR — leave
credential management to the operator (see the default-user
Secret in the outputs) unless a migration demands otherwise.
Changing this field triggers a rolling restart.

- rule: {"string":{"maxLen":"100000"}}

### spec.configuration.advancedConfig

`string`

Full advanced.config content (Erlang terms) for the settings
rabbitmq.conf cannot express. Rare; most clusters never set it.

- rule: {"string":{"maxLen":"100000"}}

### spec.configuration.envConfig

`string`

rabbitmq-env.conf additions (environment for the server process).
Shell command substitution is rejected by the CRD itself
(`$(...)` and backticks — mirrored here). Changing this field
triggers a rolling restart.

- rule: env_config must not contain shell command substitution ('$(...)' or backticks) — the CRD rejects it
- rule: {"string":{"maxLen":"100000"}}

### spec.configuration.erlangInetConfig

`string`

Erlang Inet configuration for the VM running RabbitMQ (DNS/kernel
tuning; see the Erlang inet_cfg documentation). Rare.

- rule: {"string":{"maxLen":"2000"}}

### spec.tls

`KubernetesRabbitMqTls`

TLS for client connections, from an existing certificate Secret
(the cert-manager seam).

### spec.tls.secretName

`string | valueFrom` · required

Name of a kubernetes.io/tls Secret (keys tls.crt / tls.key) in
the cluster's namespace holding the server certificate. Accepts a
literal name or a reference to a KubernetesCertificate resource's
secret output — the cert-manager seam. The certificate must cover
the client Service DNS names (`<name>.<namespace>.svc` and its
cluster-domain form).

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.tls.caSecretName

`string | valueFrom`

Name of a Secret (key ca.crt) in the cluster's namespace holding
the certificate authority for MUTUAL TLS — set it to require
client certificates. Empty = server-side TLS only.

- references: KubernetesSecret (`metadata.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: metadata.name}} -- a bare string does not parse

### spec.tls.disableNonTlsListeners

`bool`

Close every non-TLS listener (AMQP 5672, management 15672, and
the plain ports of any enabled plugin: MQTT, STOMP, their
WebSocket forms). Only TLS-enabled clients can connect after
this.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for every RabbitMQ node pod.

### spec.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.spreadAcrossNodes

`bool`

Never schedule two RabbitMQ nodes on the same Kubernetes node
(rendered as REQUIRED pod anti-affinity on the cluster's own
pods). Off by default so single-node dev clusters schedule; turn
on in production — co-located brokers make quorum queues
pointless against node loss. A cluster with more replicas than
schedulable nodes will sit Pending when this is on.

### spec.terminationGracePeriodSeconds

`int64` · optional (explicit presence)

Seconds each node gets to finish rebalancing and shut down
cleanly on pod termination. The operator default is 604800 (7
DAYS — deliberately generous so draining nodes never lose
messages); lower it for dev clusters where fast teardown matters
more than clean handoff.

- default: `604800`
- rule: {"int64":{"gte":"0"}}

### spec.delayStartSeconds

`int32` · optional (explicit presence)

Seconds the readiness of the cluster is delayed at startup (the
CR's delayStartSeconds). Operator default 30.

- default: `30`
- rule: {"int32":{"gte":0}}

### spec.skipPostDeploySteps

`bool`

Skip the operator's post-deploy steps (queue rebalancing after
rolling upgrades). Leave off unless an external process manages
rebalancing.

### spec.autoEnableAllFeatureFlags

`bool`

Enable ALL stable feature flags automatically after upgrades (the
CR's autoEnableAllFeatureFlags). Keeps the cluster upgrade-ready
without manual flag management.

### spec.secretBackend

`KubernetesRabbitMqSecretBackend`

Deliver the default-user credentials through an external secret
store instead of the operator-generated Kubernetes Secret.

### spec.secretBackend.vault

`KubernetesRabbitMqVaultBackend`

HashiCorp Vault: the operator reads (and the updater sidecar
rotates) the default-user credentials at a Vault path instead
of generating a Kubernetes Secret.

### spec.secretBackend.vault.role

`string` · required

Vault Kubernetes-auth role the cluster pods authenticate as.

- rule: {"required":true}

### spec.secretBackend.vault.defaultUserPath

`string` · required

Vault path holding the default-user credentials (fields username
/ password), e.g. "secret/data/rabbitmq/config".

- rule: {"required":true}

### spec.secretBackend.vault.annotations

`map<string, string>`

Extra Vault agent annotations placed on the cluster pods (the
vault.hashicorp.com/* vocabulary).

### spec.secretBackend.vault.pkiIssuerPath

`string`

Vault PKI engine path that issues the cluster's server
certificate (e.g. "pki/issue/cert-issuer") — the Vault
alternative to `tls.secret_name`. Empty = no Vault-issued TLS.

### spec.secretBackend.externalSecretName

`string`

Name of a PRE-EXISTING Kubernetes Secret (in the cluster's
namespace) carrying the default-user credentials — for
external-secrets-operator flows that materialize credentials
from a cloud secret manager (compose with
KubernetesExternalSecret and reference its target Secret here).

### spec.nodeSelector

`map<string, string>`

Node selector for every RabbitMQ node pod. The CR has no
nodeSelector field — the modules render this as REQUIRED node
affinity with one In-match per label (behaviorally identical for
exact matches, documented in both engines).

## Validation Rules

- `spec.ephemeral.excludes_storage`: ephemeral: true runs on emptyDir — a non-default disk_size or a storage_class must not be set with it

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesRabbitMq, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the cluster runs in. |
| `status.outputs.cluster_name` | `string` | name of the RabbitmqCluster resource (= metadata.name). |
| `status.outputs.service_name` | `string` | name of the client Service (operator naming contract: <name>). |
| `status.outputs.headless_service_name` | `string` | name of the headless inter-node Service (operator naming contract: <name>-nodes). |
| `status.outputs.amqp_endpoint` | `string` | in-cluster AMQP endpoint for clients, e.g. orders-mq.messaging.svc.cluster.local:5672 (5671 when tls.disable_non_tls_listeners closes the plain listener). |
| `status.outputs.management_endpoint` | `string` | in-cluster management API / UI endpoint, e.g. http://orders-mq.messaging.svc.cluster.local:15672 (https on 15671 when tls.disable_non_tls_listeners closes the plain listener). |
| `status.outputs.default_user_secret_name` | `string` | name of the operator-generated Secret holding the administrator credentials (keys: username, password, host, port, connection_string, ...), e.g. <name>-default-user. When secret_backend.vault is set the credentials live in Vault instead and this is empty. |
| `status.outputs.port_forward_command` | `string` | command to port-forward the management UI to a developer laptop, e.g. kubectl port-forward svc/orders-mq -n messaging 15672:15672 |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.tls.secretName` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.tls.caSecretName` | KubernetesSecret | `metadata.name` |

## See Also

- [Overview](../README.md)
