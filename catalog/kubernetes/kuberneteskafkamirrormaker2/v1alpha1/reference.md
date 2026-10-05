# KubernetesKafkaMirrorMaker2

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesKafkaMirrorMaker2Spec** declares a MirrorMaker 2
replication engine on the Strimzi `KafkaMirrorMaker2` custom
resource — continuous, offset-aware mirroring of topics and
consumer groups from one or more SOURCE Kafka clusters into one
TARGET cluster. This is the migration on-ramp: point a mirror at a
running Confluent / MSK / self-hosted cluster, let it replicate
topics, records and consumer positions into your Strimzi-managed
cluster, then cut consumers over with their offsets intact
(checkpointing translates source offsets to target offsets).

SHAPE (the Strimzi model on this line): ONE `target` cluster — the
Connect-style engine runs against it — and one `mirrors` entry per
source cluster, each carrying its own connection and what to
replicate. Under the hood each mirror runs a MirrorSourceConnector
(records + topic configuration) and a MirrorCheckpointConnector
(consumer-group offset translation).

TOPIC NAMING: mirrored topics are prefixed with the source's
`alias` by default ("prod-msk.orders"). To keep ORIGINAL names —
the usual migration posture — set the mirror's
`source_connector.config` entry "replication.policy.class" to
"org.apache.kafka.connect.mirror.IdentityReplicationPolicy" (set
the same on the checkpoint connector).

The GROUP IDENTITY fields on `target` (`group_id` and the three
storage topics) default from metadata.name and MUST be unique per
Connect-protocol workload (Connect clusters included) sharing the
target Kafka cluster.

## Example

```yaml
# Full-surface development manifest: a two-source migration posture —
# an MSK-style SCRAM source and a Confluent-style PLAIN source mirroring
# into a TLS+SCRAM Strimzi target, both mirrors pinning
# IdentityReplicationPolicy (original topic names), explicit group
# identity, scope patterns, connector tuning with auto-restart, worker
# resources/JVM/rack/metrics, and scheduling knobs. Not a
# runnable-on-kind shape — the external bootstraps and zone-keyed rack
# need real clusters.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKafkaMirrorMaker2
metadata:
  name: mm2-hack
spec:
  namespace:
    value: kafka-hack
  createNamespace: true
  version: 4.3.0
  replicas: 3
  target:
    alias: target
    bootstrapServers:
      value: kafka-hack-kafka-bootstrap.kafka-hack.svc.cluster.local:9094
    tls:
      trustedCertificates:
        - secretName:
            value: kafka-hack-cluster-ca-cert
          certificate: ca.crt
    authentication:
      type: scram-sha-512
      username: mm2-writer
      passwordSecret:
        secretName:
          value: mm2-writer
        password: password
    groupId: mm2-hack-group
    configStorageTopic: mm2-hack-configs
    statusStorageTopic: mm2-hack-status
    offsetStorageTopic: mm2-hack-offsets
    config:
      config.storage.replication.factor: "3"
      offset.storage.replication.factor: "3"
      status.storage.replication.factor: "3"
  mirrors:
    - source:
        alias: prod-msk
        bootstrapServers:
          value: b-1.prod-msk.abc123.kafka.us-west-2.amazonaws.com:9096
        tls:
          trustedCertificates:
            - secretName:
                value: prod-msk-ca
              pattern: "*.crt"
        authentication:
          type: scram-sha-512
          username: mm2-reader
          passwordSecret:
            secretName:
              value: prod-msk-credentials
            password: password
        config:
          max.poll.records: "500"
      topicsPattern: "orders.*,payments.*"
      topicsExcludePattern: ".*\\.internal"
      groupsPattern: ".*"
      groupsExcludePattern: "console-consumer-.*"
      sourceConnector:
        tasksMax: 8
        config:
          replication.policy.class: org.apache.kafka.connect.mirror.IdentityReplicationPolicy
          refresh.topics.interval.seconds: "600"
          sync.topic.acls.enabled: "false"
        autoRestart:
          enabled: true
          maxRestarts: 10
      checkpointConnector:
        tasksMax: 2
        config:
          replication.policy.class: org.apache.kafka.connect.mirror.IdentityReplicationPolicy
          sync.group.offsets.enabled: "true"
          refresh.groups.interval.seconds: "600"
        autoRestart:
          enabled: true
    - source:
        alias: confluent
        bootstrapServers:
          value: pkc-xxxxx.us-west2.gcp.confluent.cloud:9092
        authentication:
          type: plain
          username: ABCDEFGH123456
          passwordSecret:
            secretName:
              value: confluent-api-secret
            password: password
        config:
          fetch.max.bytes: "52428800"
      topicsPattern: "clickstream.*"
      sourceConnector:
        tasksMax: 4
        config:
          replication.policy.class: org.apache.kafka.connect.mirror.IdentityReplicationPolicy
      checkpointConnector:
        config:
          replication.policy.class: org.apache.kafka.connect.mirror.IdentityReplicationPolicy
  resources:
    requests:
      cpu: 500m
      memory: 2Gi
    limits:
      cpu: "2"
      memory: 4Gi
  jvm:
    xms: 1g
    xmx: 1g
  rack:
    topologyKey: topology.kubernetes.io/zone
  metrics:
    enabled: true
  nodeSelector:
    workload: kafka
  tolerations:
    - key: dedicated
      operator: Equal
      value: kafka
      effect: NoSchedule
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.version` | `string` |  |  |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.target` | `KubernetesKafkaMirrorMaker2Target` | yes |  |  |
| `spec.target.alias` | `string` |  | `target` |  |
| `spec.target.bootstrapServers` | `string \| valueFrom` | yes |  | KubernetesKafka (`status.outputs.internal_bootstrap_endpoint`) |
| `spec.target.tls` | `StrimziKafkaClientTls` |  |  |  |
| `spec.target.tls.trustedCertificates` | `[]StrimziKafkaClientTrustedCertificate` | yes |  |  |
| `spec.target.tls.trustedCertificates[].secretName` | `string \| valueFrom` | yes |  | KubernetesKafka (`status.outputs.cluster_ca_cert_secret_name`) |
| `spec.target.tls.trustedCertificates[].certificate` | `string` |  |  |  |
| `spec.target.tls.trustedCertificates[].pattern` | `string` |  |  |  |
| `spec.target.authentication` | `StrimziKafkaClientAuthentication` |  |  |  |
| `spec.target.authentication.type` | `string` | yes |  |  |
| `spec.target.authentication.certificateAndKey` | `StrimziKafkaClientCertificateAndKey` |  |  |  |
| `spec.target.authentication.certificateAndKey.secretName` | `string \| valueFrom` | yes |  | KubernetesKafkaUser (`status.outputs.secret_name`) |
| `spec.target.authentication.certificateAndKey.certificate` | `string` |  | `user.crt` |  |
| `spec.target.authentication.certificateAndKey.key` | `string` |  | `user.key` |  |
| `spec.target.authentication.username` | `string` |  |  |  |
| `spec.target.authentication.passwordSecret` | `StrimziKafkaClientPasswordSecret` |  |  |  |
| `spec.target.authentication.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesKafkaUser (`status.outputs.secret_name`) |
| `spec.target.authentication.passwordSecret.password` | `string` |  | `password` |  |
| `spec.target.authentication.sasl` | `bool` |  |  |  |
| `spec.target.authentication.config` | `map<string, string>` |  |  |  |
| `spec.target.groupId` | `string` |  |  |  |
| `spec.target.configStorageTopic` | `string` |  |  |  |
| `spec.target.statusStorageTopic` | `string` |  |  |  |
| `spec.target.offsetStorageTopic` | `string` |  |  |  |
| `spec.target.config` | `map<string, string>` |  |  |  |
| `spec.mirrors` | `[]KubernetesKafkaMirrorMaker2Mirror` | yes |  |  |
| `spec.mirrors[].source` | `KubernetesKafkaMirrorMaker2Source` | yes |  |  |
| `spec.mirrors[].source.alias` | `string` | yes |  |  |
| `spec.mirrors[].source.bootstrapServers` | `string \| valueFrom` | yes |  | KubernetesKafka (`status.outputs.internal_bootstrap_endpoint`) |
| `spec.mirrors[].source.tls` | `StrimziKafkaClientTls` |  |  |  |
| `spec.mirrors[].source.tls.trustedCertificates` | `[]StrimziKafkaClientTrustedCertificate` | yes |  |  |
| `spec.mirrors[].source.tls.trustedCertificates[].secretName` | `string \| valueFrom` | yes |  | KubernetesKafka (`status.outputs.cluster_ca_cert_secret_name`) |
| `spec.mirrors[].source.tls.trustedCertificates[].certificate` | `string` |  |  |  |
| `spec.mirrors[].source.tls.trustedCertificates[].pattern` | `string` |  |  |  |
| `spec.mirrors[].source.authentication` | `StrimziKafkaClientAuthentication` |  |  |  |
| `spec.mirrors[].source.authentication.type` | `string` | yes |  |  |
| `spec.mirrors[].source.authentication.certificateAndKey` | `StrimziKafkaClientCertificateAndKey` |  |  |  |
| `spec.mirrors[].source.authentication.certificateAndKey.secretName` | `string \| valueFrom` | yes |  | KubernetesKafkaUser (`status.outputs.secret_name`) |
| `spec.mirrors[].source.authentication.certificateAndKey.certificate` | `string` |  | `user.crt` |  |
| `spec.mirrors[].source.authentication.certificateAndKey.key` | `string` |  | `user.key` |  |
| `spec.mirrors[].source.authentication.username` | `string` |  |  |  |
| `spec.mirrors[].source.authentication.passwordSecret` | `StrimziKafkaClientPasswordSecret` |  |  |  |
| `spec.mirrors[].source.authentication.passwordSecret.secretName` | `string \| valueFrom` | yes |  | KubernetesKafkaUser (`status.outputs.secret_name`) |
| `spec.mirrors[].source.authentication.passwordSecret.password` | `string` |  | `password` |  |
| `spec.mirrors[].source.authentication.sasl` | `bool` |  |  |  |
| `spec.mirrors[].source.authentication.config` | `map<string, string>` |  |  |  |
| `spec.mirrors[].source.config` | `map<string, string>` |  |  |  |
| `spec.mirrors[].topicsPattern` | `string` |  | `.*` |  |
| `spec.mirrors[].topicsExcludePattern` | `string` |  |  |  |
| `spec.mirrors[].groupsPattern` | `string` |  | `.*` |  |
| `spec.mirrors[].groupsExcludePattern` | `string` |  |  |  |
| `spec.mirrors[].sourceConnector` | `KubernetesKafkaMirrorMaker2Connector` |  |  |  |
| `spec.mirrors[].sourceConnector.tasksMax` | `int32` |  |  |  |
| `spec.mirrors[].sourceConnector.config` | `map<string, string>` |  |  |  |
| `spec.mirrors[].sourceConnector.autoRestart` | `KubernetesKafkaMirrorMaker2AutoRestart` |  |  |  |
| `spec.mirrors[].sourceConnector.autoRestart.enabled` | `bool` |  |  |  |
| `spec.mirrors[].sourceConnector.autoRestart.maxRestarts` | `int32` |  |  |  |
| `spec.mirrors[].checkpointConnector` | `KubernetesKafkaMirrorMaker2Connector` |  |  |  |
| `spec.mirrors[].checkpointConnector.tasksMax` | `int32` |  |  |  |
| `spec.mirrors[].checkpointConnector.config` | `map<string, string>` |  |  |  |
| `spec.mirrors[].checkpointConnector.autoRestart` | `KubernetesKafkaMirrorMaker2AutoRestart` |  |  |  |
| `spec.mirrors[].checkpointConnector.autoRestart.enabled` | `bool` |  |  |  |
| `spec.mirrors[].checkpointConnector.autoRestart.maxRestarts` | `int32` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.jvm` | `KubernetesKafkaMirrorMaker2Jvm` |  |  |  |
| `spec.jvm.xms` | `string` |  |  |  |
| `spec.jvm.xmx` | `string` |  |  |  |
| `spec.rack` | `KubernetesKafkaMirrorMaker2Rack` |  |  |  |
| `spec.rack.topologyKey` | `string` | yes |  |  |
| `spec.metrics` | `KubernetesKafkaMirrorMaker2Metrics` |  |  |  |
| `spec.metrics.enabled` | `bool` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace for the MirrorMaker 2 deployment. Accepts a literal
namespace name or a reference to a KubernetesNamespace resource.
The namespace must be watched by a Strimzi operator
installation.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the
resource. When false, the namespace must already exist.

### spec.version

`string`

Kafka version the MirrorMaker 2 workers run (e.g. "4.3.0").
Empty = the operator's default version.

### spec.replicas

`int32` · optional (explicit presence)

Number of worker pods. Mirroring tasks spread across workers —
scale with source partition counts.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.target

`KubernetesKafkaMirrorMaker2Target` · required

The TARGET cluster — where mirrored data lands and where the
engine keeps its state.

- rule: {"required":true}

### spec.target.alias

`string` · optional (explicit presence)

Alias identifying the target cluster in the replication flow
(e.g. "target", "onprem"). Empty = "target". Must differ from
every mirror's source alias.

- default: `target`
- rule: alias may use alphanumerics, '.', '_' and '-'

### spec.target.bootstrapServers

`string | valueFrom` · required

Bootstrap address of the target Kafka cluster, as host:port.
Accepts a literal address or a reference to a KubernetesKafka
resource, which resolves to its in-cluster bootstrap endpoint —
the migrate-INTO-Planton wiring.

containment_exempt: the mirror WRITES TO this cluster; it is not
deployed inside it.

- references: KubernetesKafka (`status.outputs.internal_bootstrap_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafka, name: <that resource's name>, fieldPath: status.outputs.internal_bootstrap_endpoint}} -- a bare string does not parse

### spec.target.tls

`StrimziKafkaClientTls`

TLS trust for the target connection. For a Strimzi-managed
target, reference the KubernetesKafka resource to trust its
cluster CA. Omitted = plaintext.

### spec.target.tls.trustedCertificates

`[]StrimziKafkaClientTrustedCertificate` · required

Certificates to trust. For a Strimzi-managed cluster, reference
the cluster's CA certificate Secret (the KubernetesKafka
resource's `cluster_ca_cert_secret_name` output — the default
wiring below); for external clusters, name any Secret in the
consumer's namespace holding the PEM certificate(s).

- rule: {"repeated":{"minItems":"1"}}
- rule: set exactly one of certificate (a single file name in the Secret, e.g. "ca.crt") or pattern (a glob over the Secret's files, e.g. "*.crt")

### spec.target.tls.trustedCertificates[].secretName

`string | valueFrom` · required

Name of the Secret (in the consuming resource's namespace)
holding the certificate. Accepts a literal Secret name or a
reference to a KubernetesKafka resource, which resolves to that
cluster's CA certificate Secret — the common wiring when the
target is a Strimzi-managed cluster in the same namespace.

containment_exempt: trust material fetched FROM the cluster —
access, never placement.

- references: KubernetesKafka (`status.outputs.cluster_ca_cert_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafka, name: <that resource's name>, fieldPath: status.outputs.cluster_ca_cert_secret_name}} -- a bare string does not parse

### spec.target.tls.trustedCertificates[].certificate

`string`

The certificate file name within the Secret. Strimzi cluster CA
Secrets carry "ca.crt". Exactly one of certificate or pattern
must be set.

### spec.target.tls.trustedCertificates[].pattern

`string`

Glob pattern selecting certificate files within the Secret (e.g.
"*.crt") — the multi-certificate alternative to naming one file.
Exactly one of certificate or pattern must be set.

### spec.target.authentication

`StrimziKafkaClientAuthentication`

How the workers authenticate to the target. Must match the
target listener's authentication type.

- rule: tls authentication requires certificate_and_key (the client certificate the workload presents — reference a KubernetesKafkaUser with tls authentication)
- rule: scram-sha-512, scram-sha-256 and plain authentication require username and password_secret (reference a KubernetesKafkaUser with scram-sha-512 authentication for Strimzi-managed clusters)
- rule: certificate_and_key is only used with tls authentication

### spec.target.authentication.type

`string` · required

Authentication type:
"tls" (mutual TLS — the client presents the certificate in
`certificate_and_key`; pairs with tls-auth listeners),
"scram-sha-512" / "scram-sha-256" (SASL username/password from
`username` + `password_secret`; Strimzi-managed clusters use
scram-sha-512),
"plain" (SASL PLAIN — username/password in the clear inside the
TLS session; for external clusters that only offer PLAIN), or
"custom" (bring-your-own SASL mechanism via `sasl` + `config`).

- rule: authentication type must be one of tls, scram-sha-512, scram-sha-256, plain, custom
- rule: {"required":true}

### spec.target.authentication.certificateAndKey

`StrimziKafkaClientCertificateAndKey`

tls type only: the client certificate and key the workload
presents. Reference a KubernetesKafkaUser resource (tls
authentication) to use its operator-generated credential Secret
— the default wiring below.

### spec.target.authentication.certificateAndKey.secretName

`string | valueFrom` · required

Name of the Secret (in the consuming resource's namespace)
holding the client certificate and key. Accepts a literal Secret
name or a reference to a KubernetesKafkaUser resource, which
resolves to that user's operator-generated credential Secret.

- references: KubernetesKafkaUser (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafkaUser, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.target.authentication.certificateAndKey.certificate

`string` · optional (explicit presence)

Certificate file name within the Secret. KubernetesKafkaUser
credential Secrets carry "user.crt"; cert-manager Secrets carry
"tls.crt".

- default: `user.crt`

### spec.target.authentication.certificateAndKey.key

`string` · optional (explicit presence)

Private-key file name within the Secret. KubernetesKafkaUser
credential Secrets carry "user.key"; cert-manager Secrets carry
"tls.key".

- default: `user.key`

### spec.target.authentication.username

`string`

SASL username (scram-sha-512, scram-sha-256, plain).

### spec.target.authentication.passwordSecret

`StrimziKafkaClientPasswordSecret`

SASL password source (scram-sha-512, scram-sha-256, plain) — a
key within a Secret. Reference a KubernetesKafkaUser resource
(scram-sha-512 authentication) to use its operator-generated
Secret, whose password lives under the "password" key.

### spec.target.authentication.passwordSecret.secretName

`string | valueFrom` · required

Name of the Secret (in the consuming resource's namespace)
holding the password. Accepts a literal Secret name or a
reference to a KubernetesKafkaUser resource, which resolves to
that user's operator-generated credential Secret.

- references: KubernetesKafkaUser (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafkaUser, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.target.authentication.passwordSecret.password

`string` · optional (explicit presence)

The key within the Secret whose value is the password.
KubernetesKafkaUser credential Secrets carry it under
"password".

- default: `password`

### spec.target.authentication.sasl

`bool`

custom type only: enable SASL for the custom mechanism.

### spec.target.authentication.config

`map<string, string>`

custom type only: the mechanism's client configuration entries
(sasl.mechanism, sasl.jaas.config references, callback handlers).
Values are Kafka configuration strings — write numbers and
booleans as strings.

### spec.target.groupId

`string`

Engine group ID on the target. Empty = the resource's
metadata.name. MUST be unique among Connect-protocol workloads
sharing the target cluster.

### spec.target.configStorageTopic

`string`

Topic storing the engine's connector configurations on the
target. Empty = "<metadata.name>-mirrormaker2-configs".

### spec.target.statusStorageTopic

`string`

Topic storing connector/task status. Empty =
"<metadata.name>-mirrormaker2-status".

### spec.target.offsetStorageTopic

`string`

Topic storing mirroring offsets. Empty =
"<metadata.name>-mirrormaker2-offsets".

### spec.target.config

`map<string, string>`

Additional target-cluster client configuration (producer/admin
tuning). Values are configuration strings — write numbers and
booleans as strings. Connection, identity and security entries
are operator-owned and configured through their typed fields.

### spec.mirrors

`[]KubernetesKafkaMirrorMaker2Mirror` · required

One entry per SOURCE cluster to mirror from. Source aliases must
be unique (they prefix mirrored topic names under the default
replication policy) and must differ from the target's alias.

- rule: each mirror's source alias must be unique — aliases identify clusters in the replication flow and prefix mirrored topic names
- rule: {"repeated":{"minItems":"1"}}

### spec.mirrors[].source

`KubernetesKafkaMirrorMaker2Source` · required

The source cluster connection.

- rule: {"required":true}

### spec.mirrors[].source.alias

`string` · required

Alias identifying this source in the replication flow (e.g.
"prod-msk", "confluent"). Prefixes mirrored topic names under
the default replication policy — pick a name you can live with
in topic lists, or switch to IdentityReplicationPolicy.

- rule: alias may use alphanumerics, '.', '_' and '-'
- rule: {"required":true}

### spec.mirrors[].source.bootstrapServers

`string | valueFrom` · required

Bootstrap address of the source cluster, as host:port — for
migrations this is usually an EXTERNAL address (Confluent / MSK
bootstrap, a datacenter cluster). Accepts a literal address or a
reference to a KubernetesKafka resource.

containment_exempt: the mirror READS FROM this cluster; it is not
deployed inside it.

- references: KubernetesKafka (`status.outputs.internal_bootstrap_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafka, name: <that resource's name>, fieldPath: status.outputs.internal_bootstrap_endpoint}} -- a bare string does not parse

### spec.mirrors[].source.tls

`StrimziKafkaClientTls`

TLS trust for the source connection. Omitted = plaintext.

### spec.mirrors[].source.tls.trustedCertificates

`[]StrimziKafkaClientTrustedCertificate` · required

Certificates to trust. For a Strimzi-managed cluster, reference
the cluster's CA certificate Secret (the KubernetesKafka
resource's `cluster_ca_cert_secret_name` output — the default
wiring below); for external clusters, name any Secret in the
consumer's namespace holding the PEM certificate(s).

- rule: {"repeated":{"minItems":"1"}}
- rule: set exactly one of certificate (a single file name in the Secret, e.g. "ca.crt") or pattern (a glob over the Secret's files, e.g. "*.crt")

### spec.mirrors[].source.tls.trustedCertificates[].secretName

`string | valueFrom` · required

Name of the Secret (in the consuming resource's namespace)
holding the certificate. Accepts a literal Secret name or a
reference to a KubernetesKafka resource, which resolves to that
cluster's CA certificate Secret — the common wiring when the
target is a Strimzi-managed cluster in the same namespace.

containment_exempt: trust material fetched FROM the cluster —
access, never placement.

- references: KubernetesKafka (`status.outputs.cluster_ca_cert_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafka, name: <that resource's name>, fieldPath: status.outputs.cluster_ca_cert_secret_name}} -- a bare string does not parse

### spec.mirrors[].source.tls.trustedCertificates[].certificate

`string`

The certificate file name within the Secret. Strimzi cluster CA
Secrets carry "ca.crt". Exactly one of certificate or pattern
must be set.

### spec.mirrors[].source.tls.trustedCertificates[].pattern

`string`

Glob pattern selecting certificate files within the Secret (e.g.
"*.crt") — the multi-certificate alternative to naming one file.
Exactly one of certificate or pattern must be set.

### spec.mirrors[].source.authentication

`StrimziKafkaClientAuthentication`

How the workers authenticate to the source (for Confluent Cloud:
plain with the API key/secret; for MSK SCRAM: scram-sha-512).

- rule: tls authentication requires certificate_and_key (the client certificate the workload presents — reference a KubernetesKafkaUser with tls authentication)
- rule: scram-sha-512, scram-sha-256 and plain authentication require username and password_secret (reference a KubernetesKafkaUser with scram-sha-512 authentication for Strimzi-managed clusters)
- rule: certificate_and_key is only used with tls authentication

### spec.mirrors[].source.authentication.type

`string` · required

Authentication type:
"tls" (mutual TLS — the client presents the certificate in
`certificate_and_key`; pairs with tls-auth listeners),
"scram-sha-512" / "scram-sha-256" (SASL username/password from
`username` + `password_secret`; Strimzi-managed clusters use
scram-sha-512),
"plain" (SASL PLAIN — username/password in the clear inside the
TLS session; for external clusters that only offer PLAIN), or
"custom" (bring-your-own SASL mechanism via `sasl` + `config`).

- rule: authentication type must be one of tls, scram-sha-512, scram-sha-256, plain, custom
- rule: {"required":true}

### spec.mirrors[].source.authentication.certificateAndKey

`StrimziKafkaClientCertificateAndKey`

tls type only: the client certificate and key the workload
presents. Reference a KubernetesKafkaUser resource (tls
authentication) to use its operator-generated credential Secret
— the default wiring below.

### spec.mirrors[].source.authentication.certificateAndKey.secretName

`string | valueFrom` · required

Name of the Secret (in the consuming resource's namespace)
holding the client certificate and key. Accepts a literal Secret
name or a reference to a KubernetesKafkaUser resource, which
resolves to that user's operator-generated credential Secret.

- references: KubernetesKafkaUser (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafkaUser, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.mirrors[].source.authentication.certificateAndKey.certificate

`string` · optional (explicit presence)

Certificate file name within the Secret. KubernetesKafkaUser
credential Secrets carry "user.crt"; cert-manager Secrets carry
"tls.crt".

- default: `user.crt`

### spec.mirrors[].source.authentication.certificateAndKey.key

`string` · optional (explicit presence)

Private-key file name within the Secret. KubernetesKafkaUser
credential Secrets carry "user.key"; cert-manager Secrets carry
"tls.key".

- default: `user.key`

### spec.mirrors[].source.authentication.username

`string`

SASL username (scram-sha-512, scram-sha-256, plain).

### spec.mirrors[].source.authentication.passwordSecret

`StrimziKafkaClientPasswordSecret`

SASL password source (scram-sha-512, scram-sha-256, plain) — a
key within a Secret. Reference a KubernetesKafkaUser resource
(scram-sha-512 authentication) to use its operator-generated
Secret, whose password lives under the "password" key.

### spec.mirrors[].source.authentication.passwordSecret.secretName

`string | valueFrom` · required

Name of the Secret (in the consuming resource's namespace)
holding the password. Accepts a literal Secret name or a
reference to a KubernetesKafkaUser resource, which resolves to
that user's operator-generated credential Secret.

- references: KubernetesKafkaUser (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKafkaUser, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.mirrors[].source.authentication.passwordSecret.password

`string` · optional (explicit presence)

The key within the Secret whose value is the password.
KubernetesKafkaUser credential Secrets carry it under
"password".

- default: `password`

### spec.mirrors[].source.authentication.sasl

`bool`

custom type only: enable SASL for the custom mechanism.

### spec.mirrors[].source.authentication.config

`map<string, string>`

custom type only: the mechanism's client configuration entries
(sasl.mechanism, sasl.jaas.config references, callback handlers).
Values are Kafka configuration strings — write numbers and
booleans as strings.

### spec.mirrors[].source.config

`map<string, string>`

Additional source-cluster client configuration (consumer
tuning). Values are configuration strings.

### spec.mirrors[].topicsPattern

`string` · optional (explicit presence)

Regex of topics to mirror (Kafka regex form; comma-separated
lists also work, e.g. "orders,payments.*"). The recommended
starting point mirrors everything except internal topics.

- default: `.*`

### spec.mirrors[].topicsExcludePattern

`string`

Regex of topics to EXCLUDE (applied after topics_pattern). Empty
= the engine's built-in exclusions (internal/replica topics).

### spec.mirrors[].groupsPattern

`string` · optional (explicit presence)

Regex of consumer groups whose offsets to checkpoint into the
target — what makes consumer cutover seamless.

- default: `.*`

### spec.mirrors[].groupsExcludePattern

`string`

Regex of consumer groups to EXCLUDE.

### spec.mirrors[].sourceConnector

`KubernetesKafkaMirrorMaker2Connector`

The MirrorSourceConnector — replicates records and topic
configuration. Set replication.policy.class here (and on
checkpoint_connector) to IdentityReplicationPolicy to keep
original topic names — the usual migration posture.

### spec.mirrors[].sourceConnector.tasksMax

`int32` · optional (explicit presence)

Maximum parallel tasks. For the source connector, match the
source's partition volume; empty = the Connect default (1).

- rule: {"int32":{"gte":1}}

### spec.mirrors[].sourceConnector.config

`map<string, string>`

Connector configuration entries (replication.policy.class,
refresh intervals, sync.group.offsets.enabled, ...). Values are
configuration strings — write numbers and booleans as strings.

- rule: connector.plugin.version is not accepted in mirror connector config on this Strimzi line — the engine pins its own connector versions

### spec.mirrors[].sourceConnector.autoRestart

`KubernetesKafkaMirrorMaker2AutoRestart`

Automatic restart of failed connectors/tasks with incremental
back-off — strongly recommended for long-running migrations
(transient source outages otherwise leave the mirror FAILED).

### spec.mirrors[].sourceConnector.autoRestart.enabled

`bool`

Enable automatic restarts.

### spec.mirrors[].sourceConnector.autoRestart.maxRestarts

`int32` · optional (explicit presence)

Give up after this many consecutive restarts. Empty = the
operator default (7).

- rule: {"int32":{"gte":1}}

### spec.mirrors[].checkpointConnector

`KubernetesKafkaMirrorMaker2Connector`

The MirrorCheckpointConnector — translates committed consumer
offsets from source to target so consumers cut over without
reprocessing or data loss. Keep its replication.policy.class in
lockstep with source_connector's.

### spec.mirrors[].checkpointConnector.tasksMax

`int32` · optional (explicit presence)

Maximum parallel tasks. For the source connector, match the
source's partition volume; empty = the Connect default (1).

- rule: {"int32":{"gte":1}}

### spec.mirrors[].checkpointConnector.config

`map<string, string>`

Connector configuration entries (replication.policy.class,
refresh intervals, sync.group.offsets.enabled, ...). Values are
configuration strings — write numbers and booleans as strings.

- rule: connector.plugin.version is not accepted in mirror connector config on this Strimzi line — the engine pins its own connector versions

### spec.mirrors[].checkpointConnector.autoRestart

`KubernetesKafkaMirrorMaker2AutoRestart`

Automatic restart of failed connectors/tasks with incremental
back-off — strongly recommended for long-running migrations
(transient source outages otherwise leave the mirror FAILED).

### spec.mirrors[].checkpointConnector.autoRestart.enabled

`bool`

Enable automatic restarts.

### spec.mirrors[].checkpointConnector.autoRestart.maxRestarts

`int32` · optional (explicit presence)

Give up after this many consecutive restarts. Empty = the
operator default (7).

- rule: {"int32":{"gte":1}}

### spec.resources

`ContainerResources`

CPU/memory for each worker pod. Empty = no requests/limits
(fine for kind/dev; always set for production).

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

### spec.jvm

`KubernetesKafkaMirrorMaker2Jvm`

JVM heap for the workers (rendered as -Xms/-Xmx). Empty =
Strimzi's dynamic default. Set both to the SAME value for
production.

### spec.jvm.xms

`string`

Initial heap (-Xms), e.g. "1g". Set equal to xmx in production.

### spec.jvm.xmx

`string`

Maximum heap (-Xmx), e.g. "1g".

### spec.rack

`KubernetesKafkaMirrorMaker2Rack`

Rack awareness for the workers (e.g.
"topology.kubernetes.io/zone"). Requires nodes labeled with the
key.

### spec.rack.topologyKey

`string` · required

Node label whose value identifies a node's failure domain.

- rule: {"required":true}

### spec.metrics

`KubernetesKafkaMirrorMaker2Metrics`

JMX Prometheus metrics. When enabled, the module renders the
canonical Strimzi connect-metrics ConfigMap and wires it as the
deployment's metricsConfig — mirroring lag rides the standard
Connect/MirrorMaker metric families.

### spec.metrics.enabled

`bool`

Render the canonical Strimzi JMX exporter rules ConfigMap and
enable the metrics endpoint on every worker.

### spec.nodeSelector

`map<string, string>`

Node selector for the worker pods.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for the worker pods.

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

## Validation Rules

- `spec.target_alias_distinct_from_sources`: the target alias must differ from every mirror's source alias — aliases identify clusters in the replication flow, and a duplicate collapses source and target into one identity (remember the target alias defaults to "target")

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesKafkaMirrorMaker2, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the MirrorMaker 2 deployment runs in. |
| `status.outputs.mirrormaker_name` | `string` | The deployment's name (metadata.name). |
| `status.outputs.rest_api_endpoint` | `string` | In-cluster Connect REST API endpoint of the underlying engine (`http://<name>-mirrormaker2-api.<namespace>.svc.cluster.local:8083`) — read-only inspection of mirror connector status. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.target.bootstrapServers` | KubernetesKafka | `status.outputs.internal_bootstrap_endpoint` |
| `spec.target.tls.trustedCertificates[].secretName` | KubernetesKafka | `status.outputs.cluster_ca_cert_secret_name` |
| `spec.target.authentication.certificateAndKey.secretName` | KubernetesKafkaUser | `status.outputs.secret_name` |
| `spec.target.authentication.passwordSecret.secretName` | KubernetesKafkaUser | `status.outputs.secret_name` |
| `spec.mirrors[].source.bootstrapServers` | KubernetesKafka | `status.outputs.internal_bootstrap_endpoint` |
| `spec.mirrors[].source.tls.trustedCertificates[].secretName` | KubernetesKafka | `status.outputs.cluster_ca_cert_secret_name` |
| `spec.mirrors[].source.authentication.certificateAndKey.secretName` | KubernetesKafkaUser | `status.outputs.secret_name` |
| `spec.mirrors[].source.authentication.passwordSecret.secretName` | KubernetesKafkaUser | `status.outputs.secret_name` |

## See Also

- [Overview](../README.md)
