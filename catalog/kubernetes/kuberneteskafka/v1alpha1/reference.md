# KubernetesKafka

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesKafkaSpec** declares an Apache Kafka cluster on the
Strimzi `Kafka` + `KafkaNodePool` custom resources (KRaft mode —
Kafka's built-in Raft metadata quorum; ZooKeeper does not exist in
this architecture). The Strimzi cluster operator
(KubernetesStrimziKafkaOperator) reconciles the rendered resources
into brokers, controllers, listeners, certificates, and the
per-cluster entity operators.

SHAPE: `node_pools` declares the machines (roles, count, storage,
sizing) — the cluster cannot exist without at least one pool
carrying the `controller` role and one carrying the `broker` role
(one dual-role pool satisfies both; the production norm is separate
controller and broker pools). `listeners` declares how clients
reach the brokers; `config` carries broker configuration the
operator does not own.

TOPICS AND USERS are first-class resources — declare them with
KubernetesKafkaTopic / KubernetesKafkaUser (reconciled by this
cluster's entity operators, which `entity_operator` keeps enabled
by default). Exposure beyond listeners (Ingress controllers,
Gateways) composes from the exported service handles — never
embedded here.

DURABILITY defaults worth knowing: topic durability is governed by
per-topic `replicas` and the cluster's `min.insync.replicas` /
`default.replication.factor` entries in `config` — a 3-broker
cluster with the recommended entries survives a broker loss without
losing acknowledged writes (producers using acks=all).

## Example

```yaml
# Full-surface development manifest: exercises every typed field so the
# offline plan proofs cover arms the live lanes exclude (external
# listeners with cloud annotations, jbod storage, Cruise Control,
# authorization, rack awareness, CA policy). Not a runnable-on-kind shape —
# the loadbalancer listener and zone-keyed rack need a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafka
metadata:
  name: kafka-hack
spec:
  namespace:
    value: kafka-hack
  createNamespace: true
  kafkaVersion: 4.3.0
  metadataVersion: 4.3-IV0
  nodePools:
    - name: controller
      roles:
        - controller
      replicas: 3
      storage:
        size: 10Gi
        storageClass:
          value: gp3
      resources:
        requests:
          cpu: 250m
          memory: 1Gi
        limits:
          cpu: "1"
          memory: 2Gi
    - name: broker
      roles:
        - broker
      replicas: 3
      storage:
        type: jbod
        volumes:
          - id: 0
            size: 100Gi
            storageClass:
              value: gp3
            deleteClaim: false
            kraftMetadata: true
          - id: 1
            size: 100Gi
            storageClass:
              value: gp3
      resources:
        requests:
          cpu: 500m
          memory: 4Gi
        limits:
          cpu: "2"
          memory: 8Gi
      nodeSelector:
        workload: kafka
      tolerations:
        - key: dedicated
          operator: Equal
          value: kafka
          effect: NoSchedule
  listeners:
    - name: plain
      port: 9092
      tls: false
    - name: scram
      port: 9094
      tls: false
      authentication:
        type: scram-sha-512
      configuration:
        maxConnections: 1000
        maxConnectionCreationRate: 50
    - name: external
      port: 9095
      type: loadbalancer
      tls: true
      authentication:
        type: tls
      configuration:
        externalTrafficPolicy: Local
        loadBalancerSourceRanges:
          - 10.0.0.0/8
        allocateLoadBalancerNodePorts: false
        createBootstrapService: true
        publishNotReadyAddresses: false
        brokerCertChainAndKey:
          secretName:
            value: kafka-listener-cert
          certificate: tls.crt
          key: tls.key
        bootstrap:
          host: kafka.example.com
          annotations:
            service.beta.kubernetes.io/aws-load-balancer-type: external
            external-dns.alpha.kubernetes.io/hostname: kafka.example.com
          labels:
            exposure: external
          alternativeNames:
            - kafka-alt.example.com
        brokers:
          - broker: 0
            advertisedHost: broker-0.kafka.example.com
            advertisedPort: 9095
            annotations:
              external-dns.alpha.kubernetes.io/hostname: broker-0.kafka.example.com
          - broker: 1
            advertisedHost: broker-1.kafka.example.com
          - broker: 2
            advertisedHost: broker-2.kafka.example.com
  config:
    default.replication.factor: "3"
    min.insync.replicas: "2"
    offsets.topic.replication.factor: "3"
    transaction.state.log.replication.factor: "3"
    transaction.state.log.min.isr: "2"
    auto.create.topics.enable: "false"
  authorization:
    type: simple
    superUsers:
      - User:CN=platform-admin
  entityOperator:
    topicOperatorEnabled: true
    userOperatorEnabled: true
  cruiseControl:
    enabled: true
    config:
      hard.goals: com.linkedin.kafka.cruisecontrol.analyzer.goals.RackAwareGoal
    resources:
      requests:
        cpu: 100m
        memory: 512Mi
      limits:
        cpu: "1"
        memory: 1Gi
    autoRebalanceModes:
      - add-brokers
      - remove-brokers
  kafkaExporter:
    enabled: true
    groupRegex: ".*"
    topicRegex: ".*"
    resources:
      requests:
        cpu: 50m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 256Mi
  metrics:
    enabled: true
  clusterCa:
    validityDays: 730
    renewalDays: 60
  clientsCa:
    validityDays: 365
    renewalDays: 30
  rack:
    topologyKey: topology.kubernetes.io/zone
  jvm:
    xms: 2g
    xmx: 2g
  maintenanceTimeWindows:
    - "* * 0-3 ? * SUN"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.kafkaVersion` | `string` |  |  |  |
| `spec.metadataVersion` | `string` |  |  |  |
| `spec.nodePools` | `[]KubernetesKafkaNodePool` | yes |  |  |
| `spec.nodePools[].name` | `string` | yes |  |  |
| `spec.nodePools[].roles` | `[]string` | yes |  |  |
| `spec.nodePools[].replicas` | `int32` | yes |  |  |
| `spec.nodePools[].storage` | `KubernetesKafkaStorage` | yes |  |  |
| `spec.nodePools[].storage.type` | `string` |  | `persistent-claim` |  |
| `spec.nodePools[].storage.size` | `string` |  |  |  |
| `spec.nodePools[].storage.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.nodePools[].storage.deleteClaim` | `bool` |  |  |  |
| `spec.nodePools[].storage.volumes` | `[]KubernetesKafkaStorageVolume` |  |  |  |
| `spec.nodePools[].storage.volumes[].id` | `int32` |  |  |  |
| `spec.nodePools[].storage.volumes[].size` | `string` | yes |  |  |
| `spec.nodePools[].storage.volumes[].storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.nodePools[].storage.volumes[].deleteClaim` | `bool` |  |  |  |
| `spec.nodePools[].storage.volumes[].kraftMetadata` | `bool` |  |  |  |
| `spec.nodePools[].resources` | `ContainerResources` |  |  |  |
| `spec.nodePools[].resources.limits` | `CpuMemory` |  |  |  |
| `spec.nodePools[].resources.limits.cpu` | `string` |  |  |  |
| `spec.nodePools[].resources.limits.memory` | `string` |  |  |  |
| `spec.nodePools[].resources.requests` | `CpuMemory` |  |  |  |
| `spec.nodePools[].resources.requests.cpu` | `string` |  |  |  |
| `spec.nodePools[].resources.requests.memory` | `string` |  |  |  |
| `spec.nodePools[].nodeSelector` | `map<string, string>` |  |  |  |
| `spec.nodePools[].tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.nodePools[].tolerations[].key` | `string` |  |  |  |
| `spec.nodePools[].tolerations[].operator` | `string` |  |  |  |
| `spec.nodePools[].tolerations[].value` | `string` |  |  |  |
| `spec.nodePools[].tolerations[].effect` | `string` |  |  |  |
| `spec.nodePools[].tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.listeners` | `[]KubernetesKafkaListener` | yes |  |  |
| `spec.listeners[].name` | `string` | yes |  |  |
| `spec.listeners[].port` | `int32` | yes |  |  |
| `spec.listeners[].type` | `string` |  | `internal` |  |
| `spec.listeners[].tls` | `bool` |  |  |  |
| `spec.listeners[].authentication` | `KubernetesKafkaListenerAuthentication` |  |  |  |
| `spec.listeners[].authentication.type` | `string` | yes |  |  |
| `spec.listeners[].authentication.sasl` | `bool` |  |  |  |
| `spec.listeners[].authentication.listenerConfig` | `map<string, string>` |  |  |  |
| `spec.listeners[].configuration` | `KubernetesKafkaListenerConfiguration` |  |  |  |
| `spec.listeners[].configuration.class` | `string` |  |  |  |
| `spec.listeners[].configuration.externalTrafficPolicy` | `string` |  |  |  |
| `spec.listeners[].configuration.loadBalancerSourceRanges` | `[]string` |  |  |  |
| `spec.listeners[].configuration.allocateLoadBalancerNodePorts` | `bool` |  |  |  |
| `spec.listeners[].configuration.createBootstrapService` | `bool` |  |  |  |
| `spec.listeners[].configuration.useServiceDnsDomain` | `bool` |  |  |  |
| `spec.listeners[].configuration.maxConnections` | `int32` |  |  |  |
| `spec.listeners[].configuration.maxConnectionCreationRate` | `int32` |  |  |  |
| `spec.listeners[].configuration.preferredNodePortAddressType` | `string` |  |  |  |
| `spec.listeners[].configuration.publishNotReadyAddresses` | `bool` |  |  |  |
| `spec.listeners[].configuration.brokerCertChainAndKey` | `KubernetesKafkaListenerCertificate` |  |  |  |
| `spec.listeners[].configuration.brokerCertChainAndKey.secretName` | `string \| valueFrom` | yes |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.listeners[].configuration.brokerCertChainAndKey.certificate` | `string` |  | `tls.crt` |  |
| `spec.listeners[].configuration.brokerCertChainAndKey.key` | `string` |  | `tls.key` |  |
| `spec.listeners[].configuration.bootstrap` | `KubernetesKafkaListenerBootstrap` |  |  |  |
| `spec.listeners[].configuration.bootstrap.host` | `string` |  |  |  |
| `spec.listeners[].configuration.bootstrap.annotations` | `map<string, string>` |  |  |  |
| `spec.listeners[].configuration.bootstrap.labels` | `map<string, string>` |  |  |  |
| `spec.listeners[].configuration.bootstrap.loadBalancerIp` | `string` |  |  |  |
| `spec.listeners[].configuration.bootstrap.nodePort` | `int32` |  |  |  |
| `spec.listeners[].configuration.bootstrap.alternativeNames` | `[]string` |  |  |  |
| `spec.listeners[].configuration.brokers` | `[]KubernetesKafkaListenerBroker` |  |  |  |
| `spec.listeners[].configuration.brokers[].broker` | `int32` |  |  |  |
| `spec.listeners[].configuration.brokers[].host` | `string` |  |  |  |
| `spec.listeners[].configuration.brokers[].advertisedHost` | `string` |  |  |  |
| `spec.listeners[].configuration.brokers[].advertisedPort` | `int32` |  |  |  |
| `spec.listeners[].configuration.brokers[].annotations` | `map<string, string>` |  |  |  |
| `spec.listeners[].configuration.brokers[].labels` | `map<string, string>` |  |  |  |
| `spec.listeners[].configuration.brokers[].loadBalancerIp` | `string` |  |  |  |
| `spec.listeners[].configuration.brokers[].nodePort` | `int32` |  |  |  |
| `spec.config` | `map<string, string>` |  |  |  |
| `spec.authorization` | `KubernetesKafkaAuthorization` |  |  |  |
| `spec.authorization.type` | `string` | yes |  |  |
| `spec.authorization.superUsers` | `[]string` |  |  |  |
| `spec.authorization.authorizerClass` | `string` |  |  |  |
| `spec.authorization.supportsAdminApi` | `bool` |  |  |  |
| `spec.entityOperator` | `KubernetesKafkaEntityOperator` |  |  |  |
| `spec.entityOperator.topicOperatorEnabled` | `bool` |  | `true` |  |
| `spec.entityOperator.userOperatorEnabled` | `bool` |  | `true` |  |
| `spec.cruiseControl` | `KubernetesKafkaCruiseControl` |  |  |  |
| `spec.cruiseControl.enabled` | `bool` |  |  |  |
| `spec.cruiseControl.config` | `map<string, string>` |  |  |  |
| `spec.cruiseControl.resources` | `ContainerResources` |  |  |  |
| `spec.cruiseControl.resources.limits` | `CpuMemory` |  |  |  |
| `spec.cruiseControl.resources.limits.cpu` | `string` |  |  |  |
| `spec.cruiseControl.resources.limits.memory` | `string` |  |  |  |
| `spec.cruiseControl.resources.requests` | `CpuMemory` |  |  |  |
| `spec.cruiseControl.resources.requests.cpu` | `string` |  |  |  |
| `spec.cruiseControl.resources.requests.memory` | `string` |  |  |  |
| `spec.cruiseControl.autoRebalanceModes` | `[]string` |  |  |  |
| `spec.kafkaExporter` | `KubernetesKafkaExporter` |  |  |  |
| `spec.kafkaExporter.enabled` | `bool` |  |  |  |
| `spec.kafkaExporter.groupRegex` | `string` |  |  |  |
| `spec.kafkaExporter.topicRegex` | `string` |  |  |  |
| `spec.kafkaExporter.resources` | `ContainerResources` |  |  |  |
| `spec.kafkaExporter.resources.limits` | `CpuMemory` |  |  |  |
| `spec.kafkaExporter.resources.limits.cpu` | `string` |  |  |  |
| `spec.kafkaExporter.resources.limits.memory` | `string` |  |  |  |
| `spec.kafkaExporter.resources.requests` | `CpuMemory` |  |  |  |
| `spec.kafkaExporter.resources.requests.cpu` | `string` |  |  |  |
| `spec.kafkaExporter.resources.requests.memory` | `string` |  |  |  |
| `spec.metrics` | `KubernetesKafkaMetrics` |  |  |  |
| `spec.metrics.enabled` | `bool` |  |  |  |
| `spec.clusterCa` | `KubernetesKafkaCa` |  |  |  |
| `spec.clusterCa.validityDays` | `int32` |  |  |  |
| `spec.clusterCa.renewalDays` | `int32` |  |  |  |
| `spec.clientsCa` | `KubernetesKafkaCa` |  |  |  |
| `spec.clientsCa.validityDays` | `int32` |  |  |  |
| `spec.clientsCa.renewalDays` | `int32` |  |  |  |
| `spec.rack` | `KubernetesKafkaRack` |  |  |  |
| `spec.rack.topologyKey` | `string` | yes |  |  |
| `spec.jvm` | `KubernetesKafkaJvm` |  |  |  |
| `spec.jvm.xms` | `string` |  |  |  |
| `spec.jvm.xmx` | `string` |  |  |  |
| `spec.maintenanceTimeWindows` | `[]string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace for the Kafka cluster. Accepts a literal namespace name
or a reference to a KubernetesNamespace resource. The namespace
must be watched by a Strimzi operator installation, and
KafkaTopic/KafkaUser declarations for this cluster must live in
THIS namespace (their entity operators watch only here).

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the
resource. When false, the namespace must already exist.

### spec.kafkaVersion

`string`

Kafka version to run (e.g. "4.3.0"). Empty = the operator's
default version (the newest the pinned Strimzi release supports).
Upgrades are sequential per Kafka's own rules — set
`metadata_version` deliberately when upgrading (the KRaft metadata
format must never be newer than the oldest running broker
version).

### spec.metadataVersion

`string`

KRaft metadata version (e.g. "4.3-IV0"). Empty = the operator
pins it to the Kafka version's format. During a Kafka version
upgrade, leave this at the OLD version's format until every node
runs the new binaries — the metadata format is a one-way door
(downgrade becomes impossible once bumped).

### spec.nodePools

`[]KubernetesKafkaNodePool` · required

The node pools that make up the cluster. At least one pool must
carry the `controller` role and at least one the `broker` role —
a single dual-role pool (dev shape) satisfies both. Pools are
independently scalable groups of nodes sharing storage shape,
sizing and scheduling; production clusters typically run one
3-node controller pool and one or more broker pools.

- rule: no node pool carries the controller role — a KRaft cluster cannot form its metadata quorum; add 'controller' to a pool's roles (a single pool may carry both roles)
- rule: no node pool carries the broker role — the cluster would have no nodes to store or serve data; add 'broker' to a pool's roles (a single pool may carry both roles)
- rule: node pool names must be unique within the cluster
- rule: {"repeated":{"minItems":"1"}}

### spec.nodePools[].name

`string` · required

Pool name (e.g. "controller", "broker", "dual-role"). Becomes
part of pod and PVC names — and the pool resource's OWN name, so
two Kafka clusters sharing a namespace must use DISTINCT pool
names (a same-named pool collides with the other cluster's,
verified live; prefer one namespace per cluster).

- rule: pool name must be a lowercase DNS-1123 label (lowercase alphanumerics and '-', starting and ending with an alphanumeric)
- rule: {"required":true,"string":{"maxLen":"63"}}

### spec.nodePools[].roles

`[]string` · required

KRaft roles this pool's nodes carry: "controller" (metadata
quorum member), "broker" (stores and serves data), or both
(dual-role — the dev/small-cluster shape; separate pools are the
production norm so brokers can scale without disturbing the
quorum).

- rule: each role must be 'controller' or 'broker'
- rule: roles must not repeat within a pool
- rule: {"repeated":{"minItems":"1"}}

### spec.nodePools[].replicas

`int32` · required

Number of nodes in the pool. Controller pools should be an ODD
count (3 for production — the quorum tolerates one controller
loss; 1 only for dev). Broker counts follow capacity; 3 is the
floor for a cluster that survives a broker loss with the
recommended replication settings.

- rule: {"required":true,"int32":{"gte":1}}

### spec.nodePools[].storage

`KubernetesKafkaStorage` · required

Node storage. Kafka is a storage-bound system — persistent
storage is the default and the only production-sane choice.

- rule: {"required":true}
- rule: persistent-claim storage requires a size (e.g. "100Gi")
- rule: jbod storage requires at least one entry in volumes
- rule: volumes are only used with the jbod storage type — for a single volume, set size/storage_class directly

### spec.nodePools[].storage.type

`string` · optional (explicit presence)

Storage type: "persistent-claim" (default), "ephemeral", or
"jbod".

- default: `persistent-claim`
- rule: storage type must be persistent-claim, ephemeral, or jbod

### spec.nodePools[].storage.size

`string`

Volume size for persistent-claim storage (e.g. "100Gi").
Required for persistent-claim; ignored for ephemeral and jbod
(jbod sizes live on each volume).

### spec.nodePools[].storage.storageClass

`string | valueFrom`

StorageClass for persistent-claim storage. Accepts a literal
class name or a reference to a KubernetesStorageClass resource.
Empty = the cluster default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.nodePools[].storage.deleteClaim

`bool`

Delete the PersistentVolumeClaims when the cluster (or pool) is
deleted. Default false — volumes outlive the cluster so data
survives accidental deletion; true hands the data's lifetime to
the resource's.

### spec.nodePools[].storage.volumes

`[]KubernetesKafkaStorageVolume`

JBOD volumes (jbod type only): multiple persistent volumes per
node, each a separate disk Kafka stripes log segments across.

### spec.nodePools[].storage.volumes[].id

`int32`

Volume ID, unique within the pool's volume list (0, 1, 2, ...).
IDs are part of the on-disk layout — never renumber existing
volumes.

- rule: {"int32":{"gte":0}}

### spec.nodePools[].storage.volumes[].size

`string` · required

Volume size (e.g. "500Gi").

- rule: {"required":true}

### spec.nodePools[].storage.volumes[].storageClass

`string | valueFrom`

StorageClass for this volume. Empty = the cluster default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.nodePools[].storage.volumes[].deleteClaim

`bool`

Delete this volume's PVCs with the cluster. Same semantics as
KubernetesKafkaStorage.delete_claim.

### spec.nodePools[].storage.volumes[].kraftMetadata

`bool`

Store the KRaft metadata log on this volume (exactly one volume
per pool may carry it; default: the volume with the lowest ID).

### spec.nodePools[].resources

`ContainerResources`

CPU/memory for each node in the pool. Empty = no requests/limits
(fine for kind/dev; always set for production — the JVM heap
default derives from the memory limit).

### spec.nodePools[].resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.nodePools[].resources.limits.cpu

`string`

### spec.nodePools[].resources.limits.memory

`string`

### spec.nodePools[].resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.nodePools[].resources.requests.cpu

`string`

### spec.nodePools[].resources.requests.memory

`string`

### spec.nodePools[].nodeSelector

`map<string, string>`

Node selector for this pool's pods.

### spec.nodePools[].tolerations

`[]WorkloadToleration`

Tolerations for this pool's pods.

### spec.nodePools[].tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.nodePools[].tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.nodePools[].tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.nodePools[].tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.nodePools[].tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.listeners

`[]KubernetesKafkaListener` · required

Listeners — how clients reach the brokers. At least one is
required. Names must be unique and ports must be unique across
the list.

- rule: listener names must be unique within the cluster
- rule: listener ports must be unique within the cluster
- rule: {"repeated":{"minItems":"1"}}
- rule: ingress, route and tlsroute listeners are TLS-passthrough and require tls: true (verified: the operator rejects them without TLS)
- rule: mutual-TLS client authentication requires tls: true on the listener
- rule: ingress listeners require configuration.bootstrap.host and a broker host per broker ID (verified: the operator rejects ingress listeners without them)

### spec.listeners[].name

`string` · required

Listener name — lowercase alphanumerics, at most 11 characters
(it becomes part of service names and certificate SANs).

- rule: listener name must be 1-11 lowercase alphanumeric characters
- rule: {"required":true}

### spec.listeners[].port

`int32` · required

Listener port, 9092 or above (ports below 9092 are reserved for
Kafka's internal listeners). Must be unique across listeners.

- rule: {"required":true,"int32":{"lte":65535,"gte":9092}}

### spec.listeners[].type

`string` · optional (explicit presence)

Listener type — how the listener is exposed:
"internal" (in-cluster ClusterIP services, the default),
"cluster-ip" (per-broker ClusterIP services — for custom proxies),
"nodeport" (NodePort services — reachable on node IPs),
"loadbalancer" (one cloud LoadBalancer per broker + bootstrap —
cloud annotations ride configuration.bootstrap/brokers),
"ingress" (nginx-Ingress-based TLS passthrough — REQUIRES
per-broker hosts and an ingress controller with SSL passthrough
enabled; DEPRECATED upstream in this Strimzi line after the
Ingress NGINX Controller's archiving — prefer loadbalancer or
nodeport for new clusters), or
"route"/"tlsroute" (OpenShift Routes / Gateway API TLSRoutes).

- default: `internal`
- rule: listener type must be one of internal, cluster-ip, nodeport, loadbalancer, ingress, route, tlsroute

### spec.listeners[].tls

`bool`

Enable TLS on this listener. ingress, route and tlsroute
listeners are TLS-passthrough by construction and require true.
Client traffic on tls=false listeners is PLAINTEXT — keep those
internal.

### spec.listeners[].authentication

`KubernetesKafkaListenerAuthentication`

Client authentication on this listener. Omitted = unauthenticated
(acceptable only for fenced internal listeners).

### spec.listeners[].authentication.type

`string` · required

Authentication type:
"tls" (mutual TLS — clients present certificates issued via
KubernetesKafkaUser tls authentication),
"scram-sha-512" (username/password — credentials generated by
KubernetesKafkaUser scram-sha-512 authentication), or
"custom" (bring-your-own SASL mechanism/OAuth — configured
through listener_config; OAuth lost its first-class type in
Strimzi 1.x and routes through custom).

- rule: authentication type must be tls, scram-sha-512, or custom
- rule: {"required":true}

### spec.listeners[].authentication.sasl

`bool`

custom type only: enable SASL on the listener.

### spec.listeners[].authentication.listenerConfig

`map<string, string>`

custom type only: the listener's Kafka configuration entries
(rendered under the custom authentication's listenerConfig).
Values are Kafka configuration strings — write numbers and
booleans as strings; the operator serializes them into listener
properties.

### spec.listeners[].configuration

`KubernetesKafkaListenerConfiguration`

Type-specific exposure configuration (hosts, cloud annotations,
traffic policy, connection limits).

### spec.listeners[].configuration.class

`string`

Controller class for exposure: the IngressClass name for ingress
listeners, the LoadBalancerClass for loadbalancer listeners.

### spec.listeners[].configuration.externalTrafficPolicy

`string` · optional (explicit presence)

externalTrafficPolicy for loadbalancer/nodeport listeners:
"Local" (preserves client IPs, no extra hop) or "Cluster" (the
Kubernetes default).

- rule: external_traffic_policy must be Local or Cluster

### spec.listeners[].configuration.loadBalancerSourceRanges

`[]string`

CIDR ranges allowed to reach loadbalancer listeners (rendered as
loadBalancerSourceRanges).

### spec.listeners[].configuration.allocateLoadBalancerNodePorts

`bool` · optional (explicit presence)

Allocate node ports for loadbalancer listeners. Unset = the
Kubernetes default (true). Set false on clouds routing LB
traffic directly to pods (the ingress-nginx annotation recipes'
companion knob).

### spec.listeners[].configuration.createBootstrapService

`bool` · optional (explicit presence)

Create the bootstrap service. Unset = true. Disabling is an
advanced shape where clients bootstrap through per-broker
addresses only.

### spec.listeners[].configuration.useServiceDnsDomain

`bool`

Use the fully-qualified *.svc.<cluster-domain> DNS names in
advertised addresses of internal listeners.

### spec.listeners[].configuration.maxConnections

`int32` · optional (explicit presence)

Per-listener client connection cap per broker (Kafka's
max.connections).

- rule: {"int32":{"gte":1}}

### spec.listeners[].configuration.maxConnectionCreationRate

`int32` · optional (explicit presence)

Per-listener new-connection rate cap per broker (Kafka's
max.connection.creation.rate).

- rule: {"int32":{"gte":1}}

### spec.listeners[].configuration.preferredNodePortAddressType

`string` · optional (explicit presence)

Preferred address type advertised by nodeport listeners:
ExternalDNS, ExternalIP, InternalDNS, InternalIP, or Hostname.

- rule: preferred_node_port_address_type must be one of ExternalDNS, ExternalIP, InternalDNS, InternalIP, Hostname

### spec.listeners[].configuration.publishNotReadyAddresses

`bool`

Publish broker addresses before readiness (nodeport/loadbalancer
listeners on slow-provisioning clouds).

### spec.listeners[].configuration.brokerCertChainAndKey

`KubernetesKafkaListenerCertificate`

Serve this listener with a custom server certificate instead of
the operator-generated one (the cert-manager seam: reference a
KubernetesCertificate's secret). The certificate's SANs must
cover the listener's advertised names — the operator does NOT
validate this; clients fail TLS verification at connect time if
they don't.

### spec.listeners[].configuration.brokerCertChainAndKey.secretName

`string | valueFrom` · required

Name of the Secret holding the certificate. Accepts a literal
name or a reference to a KubernetesCertificate resource's output
secret.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.listeners[].configuration.brokerCertChainAndKey.certificate

`string` · optional (explicit presence)

Key of the certificate chain within the Secret. cert-manager
writes "tls.crt".

- default: `tls.crt`

### spec.listeners[].configuration.brokerCertChainAndKey.key

`string` · optional (explicit presence)

Key of the private key within the Secret. cert-manager writes
"tls.key".

- default: `tls.key`

### spec.listeners[].configuration.bootstrap

`KubernetesKafkaListenerBootstrap`

Bootstrap service/host configuration (ingress/route/tlsroute
hosts, loadbalancer annotations and static IPs, nodeport
overrides).

### spec.listeners[].configuration.bootstrap.host

`string`

Bootstrap host (REQUIRED for ingress and tlsroute listeners;
used for route listeners' host override). This is the DNS name
clients put in bootstrap.servers — point it (and the per-broker
hosts) at the ingress controller's address, e.g. via
KubernetesExternalDns.

### spec.listeners[].configuration.bootstrap.annotations

`map<string, string>`

Additional annotations on the bootstrap service — the cloud
LoadBalancer configuration surface (e.g.
service.beta.kubernetes.io/aws-load-balancer-type: external for
an AWS NLB via the AWS Load Balancer Controller, or
external-dns.alpha.kubernetes.io/hostname for DNS automation).

### spec.listeners[].configuration.bootstrap.labels

`map<string, string>`

Additional labels on the bootstrap service.

### spec.listeners[].configuration.bootstrap.loadBalancerIp

`string`

Static loadBalancerIP for the bootstrap LoadBalancer (clouds that
support IP reservation; most modern controllers prefer
annotations).

### spec.listeners[].configuration.bootstrap.nodePort

`int32` · optional (explicit presence)

Static node port for the bootstrap NodePort service (30000-32767
on default clusters). Unset = allocated by Kubernetes.

- rule: {"int32":{"gte":1}}

### spec.listeners[].configuration.bootstrap.alternativeNames

`[]string`

Additional SANs to put on the bootstrap certificate (extra DNS
names clients may use).

### spec.listeners[].configuration.brokers

`[]KubernetesKafkaListenerBroker`

Per-broker configuration entries — REQUIRED for ingress and
tlsroute listeners (each broker needs its own host).

- rule: each broker ID may appear at most once in the brokers list

### spec.listeners[].configuration.brokers[].broker

`int32`

Broker (node) ID this entry configures.

- rule: {"int32":{"gte":0}}

### spec.listeners[].configuration.brokers[].host

`string`

Broker host (REQUIRED per broker for ingress and tlsroute
listeners — each broker must be individually addressable).

### spec.listeners[].configuration.brokers[].advertisedHost

`string`

Override the advertised hostname clients are told to reach THIS
broker on (NAT/proxy topologies).

### spec.listeners[].configuration.brokers[].advertisedPort

`int32` · optional (explicit presence)

Override the advertised port.

- rule: {"int32":{"gte":1}}

### spec.listeners[].configuration.brokers[].annotations

`map<string, string>`

Additional annotations on this broker's service (same cloud
surface as bootstrap.annotations).

### spec.listeners[].configuration.brokers[].labels

`map<string, string>`

Additional labels on this broker's service.

### spec.listeners[].configuration.brokers[].loadBalancerIp

`string`

Static loadBalancerIP for this broker's LoadBalancer.

### spec.listeners[].configuration.brokers[].nodePort

`int32` · optional (explicit presence)

Static node port for this broker's NodePort service.

- rule: {"int32":{"gte":1}}

### spec.config

`map<string, string>`

Kafka broker configuration (server.properties entries), e.g.
"default.replication.factor", "min.insync.replicas",
"auto.create.topics.enable". Values are Kafka configuration
strings — write numbers and booleans as strings ("3", "false");
the operator serializes every value into Java properties form.

The operator OWNS listener, node identity, security and quorum
configuration: entries with the prefixes listeners, advertised.,
broker., listener., host.name, port, inter.broker.listener.name,
sasl., ssl., security., password., log.dir, authorizer.,
super.user, node.id, process.roles, controller. (and the
cruise-control metrics reporter keys) are IGNORED with an
operator log warning, not applied — configure those concerns
through their typed fields instead.

### spec.authorization

`KubernetesKafkaAuthorization`

Authorization for the whole cluster. Omitted = no authorizer:
every authenticated (or anonymous, on a no-auth listener) client
can do everything. Enable `simple` to enforce the ACLs declared
on KubernetesKafkaUser resources — from that moment, clients
WITHOUT matching ACLs are denied.

- rule: custom authorization requires authorizer_class (the fully-qualified class name of the authorizer)
- rule: authorizer_class is only used with custom authorization — simple uses Kafka's built-in ACL authorizer

### spec.authorization.type

`string` · required

Authorizer type: "simple" (Kafka's built-in ACL authorizer —
ACLs come from KubernetesKafkaUser authorization blocks) or
"custom" (bring-your-own authorizer class; Keycloak/OPA lost
their first-class types in Strimzi 1.x and route through
custom).

- rule: authorization type must be simple or custom
- rule: {"required":true}

### spec.authorization.superUsers

`[]string`

Principals with unrestricted access regardless of ACLs, in
Kafka principal form (e.g. "User:CN=my-user" for TLS users,
"User:my-user" for SCRAM users). Include your operational
tooling's principals BEFORE enabling authorization — turning on
an authorizer without super users can lock every client out at
once.

### spec.authorization.authorizerClass

`string`

custom type only: fully-qualified authorizer class name (must be
on the broker classpath via a custom image).

### spec.authorization.supportsAdminApi

`bool`

custom type only: whether the custom authorizer supports the
Kafka Admin API for ACL management — required for the user
operator to manage ACLs through it.

### spec.entityOperator

`KubernetesKafkaEntityOperator`

The per-cluster entity operators that reconcile KafkaTopic and
KafkaUser resources. Both enabled by default — required for
KubernetesKafkaTopic / KubernetesKafkaUser declarations against
this cluster to have any effect.

### spec.entityOperator.topicOperatorEnabled

`bool` · optional (explicit presence)

Topic operator — reconciles KafkaTopic resources in THIS
cluster's namespace into real topics. Enabled by default;
disabling it makes KubernetesKafkaTopic declarations against
this cluster silently inert.

- default: `true`

### spec.entityOperator.userOperatorEnabled

`bool` · optional (explicit presence)

User operator — reconciles KafkaUser resources in THIS cluster's
namespace into authenticated principals + credential Secrets.
Enabled by default; disabling it makes KubernetesKafkaUser
declarations against this cluster silently inert.

- default: `true`

### spec.cruiseControl

`KubernetesKafkaCruiseControl`

Cruise Control — Kafka's partition-rebalancing autopilot. When
enabled, the operator deploys Cruise Control beside the cluster
and `KafkaRebalance` operations (an operational verb, applied
with kubectl or automation — deliberately not a declared
resource) can optimize partition placement. `auto_rebalance`
additionally rebalances automatically when pools scale.

### spec.cruiseControl.enabled

`bool`

Deploy Cruise Control beside the cluster.

### spec.cruiseControl.config

`map<string, string>`

Cruise Control configuration entries (goals, thresholds). Values
are configuration strings — write numbers and booleans as
strings; the operator serializes them. The metrics topic and
bootstrap-server entries are operator-owned and ignored here.

### spec.cruiseControl.resources

`ContainerResources`

Cruise Control container resources.

### spec.cruiseControl.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.cruiseControl.resources.limits.cpu

`string`

### spec.cruiseControl.resources.limits.memory

`string`

### spec.cruiseControl.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.cruiseControl.resources.requests.cpu

`string`

### spec.cruiseControl.resources.requests.memory

`string`

### spec.cruiseControl.autoRebalanceModes

`[]string`

Automatic rebalancing on pool scale events: "add-brokers"
rebalances onto newly added brokers, "remove-brokers" drains
brokers before removal. Empty = rebalances only when a
KafkaRebalance is applied manually.

- rule: each auto-rebalance mode must be add-brokers or remove-brokers

### spec.kafkaExporter

`KubernetesKafkaExporter`

Kafka Exporter — consumer-lag and topic/partition Prometheus
metrics derived from the cluster's own protocol (the metrics the
JMX exporter cannot see). Deploys a separate exporter pod.

### spec.kafkaExporter.enabled

`bool`

Deploy the exporter.

### spec.kafkaExporter.groupRegex

`string`

Regex of consumer groups to export. Empty = the operator default
(".*").

### spec.kafkaExporter.topicRegex

`string`

Regex of topics to export. Empty = the operator default (".*").

### spec.kafkaExporter.resources

`ContainerResources`

Exporter container resources.

### spec.kafkaExporter.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.kafkaExporter.resources.limits.cpu

`string`

### spec.kafkaExporter.resources.limits.memory

`string`

### spec.kafkaExporter.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.kafkaExporter.resources.requests.cpu

`string`

### spec.kafkaExporter.resources.requests.memory

`string`

### spec.metrics

`KubernetesKafkaMetrics`

JMX Prometheus metrics for brokers (and controllers). When
enabled, the module renders the canonical Strimzi
kafka-metrics ConfigMap (the upstream example rules: broker,
network, log and controller metric families) and wires it as the
cluster's metricsConfig — pair with KubernetesKafkaExporter for
consumer-lag visibility.

### spec.metrics.enabled

`bool`

Render the canonical Strimzi JMX exporter rules ConfigMap and
enable the metrics endpoint on every broker/controller (port
9404 inside the pods; scrape with a PodMonitor or annotations
per your Prometheus setup).

### spec.clusterCa

`KubernetesKafkaCa`

Cluster CA (signs broker certificates). Omitted = the operator
generates and renews the CA (the recommended posture). Configure
validity/renewal windows here; bringing your OWN CA is a
helm-values-class edge deliberately not modeled — the operator
renews self-generated CAs automatically.

### spec.clusterCa.validityDays

`int32` · optional (explicit presence)

CA certificate validity in days. Operator default: 365.

- rule: {"int32":{"gte":1}}

### spec.clusterCa.renewalDays

`int32` · optional (explicit presence)

Days before expiry at which the operator renews the CA (and
rolls certificates during maintenance windows when configured).
Operator default: 30.

- rule: {"int32":{"gte":1}}

### spec.clientsCa

`KubernetesKafkaCa`

Clients CA (signs KafkaUser TLS client certificates). Same
posture as cluster_ca.

### spec.clientsCa.validityDays

`int32` · optional (explicit presence)

CA certificate validity in days. Operator default: 365.

- rule: {"int32":{"gte":1}}

### spec.clientsCa.renewalDays

`int32` · optional (explicit presence)

Days before expiry at which the operator renews the CA (and
rolls certificates during maintenance windows when configured).
Operator default: 30.

- rule: {"int32":{"gte":1}}

### spec.rack

`KubernetesKafkaRack`

Rack awareness: spreads partition replicas across the topology
domains named by this key (e.g. "topology.kubernetes.io/zone")
and enables follower-fetching from the closest replica. Requires
nodes labeled with the key; the operator injects each broker's
rack from its node's label value.

### spec.rack.topologyKey

`string` · required

Node label whose value identifies a node's failure domain,
e.g. "topology.kubernetes.io/zone".

- rule: {"required":true}

### spec.jvm

`KubernetesKafkaJvm`

JVM heap for broker/controller nodes (rendered as -Xms/-Xmx).
Empty = Strimzi's dynamic default (a fraction of the container
memory limit). Set both to the SAME value for production —
heap growth causes long GC pauses that look like broker
failures to the controller quorum.

### spec.jvm.xms

`string`

Initial heap (-Xms), e.g. "4g". Set equal to xmx in production.

### spec.jvm.xmx

`string`

Maximum heap (-Xmx), e.g. "4g". Kafka relies on the OS page
cache — heap beyond ~6g rarely helps; give the rest of the
container memory to the page cache instead.

### spec.maintenanceTimeWindows

`[]string`

Maintenance time windows (cron expressions, e.g.
"* * 0-3 ? * SUN") during which the operator performs rolling
updates triggered by certificate renewals. Empty = renewals roll
pods whenever they come due.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesKafka, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the cluster runs in. |
| `status.outputs.cluster_name` | `string` | Cluster name (`metadata.name`) — the value of the strimzi.io/cluster label that binds KafkaNodePool, KafkaTopic and KafkaUser resources to this cluster. |
| `status.outputs.bootstrap_service_name` | `string` | Name of the bootstrap Service for internal listeners (`<cluster>-kafka-bootstrap`). |
| `status.outputs.internal_bootstrap_endpoint` | `string` | In-cluster bootstrap address for the FIRST internal-type listener (`<cluster>-kafka-bootstrap.<namespace>.svc.cluster.local:<port>`) — the value workloads put in bootstrap.servers. Empty when the cluster declares no internal listener. |
| `status.outputs.cluster_ca_cert_secret_name` | `string` | Name of the Secret holding the cluster CA certificate (`<cluster>-cluster-ca-cert`, key `ca.crt`) — what TLS clients add to their truststore. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.nodePools[].storage.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.nodePools[].storage.volumes[].storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.listeners[].configuration.brokerCertChainAndKey.secretName` | KubernetesCertificate | `status.outputs.secret_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesKafkaConnect | `spec.bootstrapServers` | `status.outputs.internal_bootstrap_endpoint` |
| KubernetesKafkaConnect | `spec.tls.trustedCertificates[].secretName` | `status.outputs.cluster_ca_cert_secret_name` |
| KubernetesKafkaMirrorMaker2 | `spec.target.bootstrapServers` | `status.outputs.internal_bootstrap_endpoint` |
| KubernetesKafkaMirrorMaker2 | `spec.target.tls.trustedCertificates[].secretName` | `status.outputs.cluster_ca_cert_secret_name` |
| KubernetesKafkaMirrorMaker2 | `spec.mirrors[].source.bootstrapServers` | `status.outputs.internal_bootstrap_endpoint` |
| KubernetesKafkaMirrorMaker2 | `spec.mirrors[].source.tls.trustedCertificates[].secretName` | `status.outputs.cluster_ca_cert_secret_name` |
| KubernetesKafkaTopic | `spec.kafkaCluster` | `status.outputs.cluster_name` |
| KubernetesKafkaUi | `spec.clusters[].bootstrapServers` | `status.outputs.internal_bootstrap_endpoint` |
| KubernetesKafkaUi | `spec.clusters[].tls.caSecretName` | `status.outputs.cluster_ca_cert_secret_name` |
| KubernetesKafkaUser | `spec.kafkaCluster` | `status.outputs.cluster_name` |
| KubernetesKarapace | `spec.kafka.bootstrapServers` | `status.outputs.internal_bootstrap_endpoint` |
| KubernetesKarapace | `spec.kafka.tls.caSecretName` | `status.outputs.cluster_ca_cert_secret_name` |

## See Also

- [Overview](../README.md)
